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

package main

import (
	"Boxify/internal/connection"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
)

// ViaSSHDialer为MySQL注册一个通过SSH代理的自定义网络
type ViaSSHDialer struct {
	sshClient *ssh.Client
}

func (d *ViaSSHDialer) Dial(ctx context.Context, addr string) (net.Conn, error) {
	return d.sshClient.Dial("tcp", addr)
}

// connectSSH建立一个SSH连接并返回一个Dialer
func connectSSH(config *connection.SSHConfig) (*ssh.Client, error) {
	authMethods := []ssh.AuthMethod{}

	if config.KeyPath != "" {
		key, err := os.ReadFile(config.KeyPath)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 在生产中使用严格的检查！
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return ssh.Dial("tcp", addr, sshConfig)
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

	mysql.RegisterDialContext(netName, func(ctx context.Context, addr string) (net.Conn, error) {
		return client.Dial("tcp", addr)
	})

	return netName, nil
}
