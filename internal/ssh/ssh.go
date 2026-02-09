// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/chenyang-zz/boxify/internal/logger"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
)

// ViaSSHDialer为MySQL注册一个通过SSH代理的自定义网络
type ViaSSHDialer struct {
	sshClient *ssh.Client
}

// Dial SSH连接拨号函数
func (d *ViaSSHDialer) Dial(ctx context.Context, addr string) (net.Conn, error) {
	return dialContext(ctx, d.sshClient, "tcp", addr)
}

// RegisterSSHNetwork为指定的SSH隧道注册一个唯一的网络名
// 返回在DSN中使用的网络名
func RegisterSSHNetwork(sshConfig *connection.SSHConfig) (string, error) {
	client, err := connectSSH(sshConfig)
	if err != nil {
		return "", err
	}

	// 生产唯一的网络名
	netName := fmt.Sprintf("ssh_%s_%d", sshConfig.Host, time.Now().UnixNano())
	logger.Infof("注册 SSH 网络：%s（地址=%s:%d 用户=%s）", netName, sshConfig.Host, sshConfig.Port, sshConfig.User)

	mysql.RegisterDialContext(netName, func(ctx context.Context, addr string) (net.Conn, error) {
		return dialContext(ctx, client, "tcp", addr)
	})

	return netName, nil
}

// connectSSH建立一个SSH连接并返回一个Dialer
func connectSSH(config *connection.SSHConfig) (*ssh.Client, error) {
	logger.Infof("开始建立ssh连接，地址=%s:%d 用户=%s", config.Host, config.Port, config.User)
	authMethods := []ssh.AuthMethod{}

	if config.KeyPath != "" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			logger.Warnf("读取 SSH 私钥失败：路径=%s，原因：%v", config.KeyPath, err)
		} else {
			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				logger.Warnf("解析 SSH 私钥失败：路径=%s，原因：%v", config.KeyPath, err)
			} else {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		logger.Warnf("SSH 未配置认证方式（密码或私钥）")
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 在生产中使用严格的检查！
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		logger.ErrorfWithTrace(err, "SSH 连接建立失败：地址=%s 用户=%s", addr, config.User)
		return nil, err
	}
	logger.Infof("SSH 连接建立成功：地址=%s 用户=%s", addr, config.User)
	return client, nil
}

// dialContext 是一个辅助函数，用于在SSH连接上拨号，并支持上下文取消
func dialContext(ctx context.Context, client *ssh.Client, network, addr string) (net.Conn, error) {
	if client == nil {
		return nil, fmt.Errorf("SSH 客户端为 nil")
	}

	type result struct {
		conn net.Conn
		err  error
	}

	ch := make(chan result, 1)
	go func() {
		// 添加恢复机制以防止 panic
		defer func() {
			if r := recover(); r != nil {
				ch <- result{conn: nil, err: fmt.Errorf("连接 panic: %v", r)}
			}
		}()
		c, err := client.Dial(network, addr)
		ch <- result{conn: c, err: err}
	}()

	select {
	case <-ctx.Done():
		go func() {
			r := <-ch
			if r.conn != nil {
				_ = r.conn.Close()
			}
		}()
		return nil, ctx.Err()
	case r := <-ch:
		return r.conn, r.err
	}
}
