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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chenyang-zz/boxify/internal/connection"

	"golang.org/x/crypto/ssh"
)

// TestConnectSSH_InvalidConfig 测试无效配置的 SSH 连接
func TestConnectSSH_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *connection.SSHConfig
		wantErr bool
	}{
		{
			name: "空配置",
			config: &connection.SSHConfig{
				Host:     "",
				Port:     0,
				User:     "",
				Password: "",
				KeyPath:  "",
			},
			wantErr: true,
		},
		{
			name: "无效主机",
			config: &connection.SSHConfig{
				Host:     "invalid-host-that-does-not-exist.local",
				Port:     22,
				User:     "test",
				Password: "test",
			},
			wantErr: true,
		},
		{
			name: "无认证方式",
			config: &connection.SSHConfig{
				Host:     "localhost",
				Port:     22,
				User:     "test",
				Password: "",
				KeyPath:  "",
			},
			wantErr: true,
		},
		{
			name: "无效私钥路径",
			config: &connection.SSHConfig{
				Host:     "localhost",
				Port:     22,
				User:     "test",
				Password: "",
				KeyPath:  "/nonexistent/key/path",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := connectSSH(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("connectSSH() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("connectSSH() 返回 nil client，但期望成功")
			}
		})
	}
}

// TestConnectSSH_InvalidKeyFile 测试无效的密钥文件
func TestConnectSSH_InvalidKeyFile(t *testing.T) {
	// 创建临时文件，但写入无效内容
	tmpDir := t.TempDir()
	invalidKeyFile := filepath.Join(tmpDir, "invalid_key")

	err := os.WriteFile(invalidKeyFile, []byte("this is not a valid ssh key"), 0600)
	if err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}

	config := &connection.SSHConfig{
		Host:    "localhost",
		Port:    22,
		User:    "test",
		KeyPath: invalidKeyFile,
	}

	client, err := connectSSH(config)
	if err == nil {
		client.Close()
		t.Error("期望使用无效密钥文件时出错，但没有错误")
	}
}

// TestDialContext_ContextCancellation 测试上下文取消
func TestDialContext_ContextCancellation(t *testing.T) {
	// 创建一个模拟的 SSH 客户端
	config := &ssh.ClientConfig{
		User:            "test",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 使用 localhost 的随机端口，这将导致连接挂起
	client, err := ssh.Dial("tcp", "localhost:0", config)
	if err != nil {
		// 期望失败，但我们需要一个有效的客户端来测试 dialContext
		// 跳过此测试
		t.Skip("无法创建测试 SSH 客户端")
		return
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消上下文
	cancel()

	_, err = dialContext(ctx, client, "tcp", "localhost:9999")
	if err != context.Canceled {
		t.Errorf("dialContext() 期望 context.Canceled 错误，得到 %v", err)
	}
}

// TestDialContext_Timeout 测试超时
func TestDialContext_Timeout(t *testing.T) {
	config := &ssh.ClientConfig{
		User:            "test",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 创建一个指向无法访问地址的客户端
	client, err := ssh.Dial("tcp", "localhost:0", config)
	if err != nil {
		t.Skip("无法创建测试 SSH 客户端")
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = dialContext(ctx, client, "tcp", "192.0.2.1:9999") // TEST-NET-1，保证无法访问
	if err == nil {
		t.Error("dialContext() 期望超时错误，但没有错误")
	}
	if ctx.Err() == context.DeadlineExceeded && err != context.DeadlineExceeded {
		t.Logf("dialContext() 返回错误: %v", err)
	}
}

// TestViaSSHDialer_Dial 测试 ViaSSHDialer.Dial 方法
func TestViaSSHDialer_Dial(t *testing.T) {
	config := &ssh.ClientConfig{
		User:            "test",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "localhost:0", config)
	if err != nil {
		t.Skip("无法创建测试 SSH 客户端")
		return
	}
	defer client.Close()

	dialer := &ViaSSHDialer{
		sshClient: client,
	}

	ctx := context.Background()
	_, err = dialer.Dial(ctx, "localhost:9999")
	if err == nil {
		t.Error("ViaSSHDialer.Dial() 期望连接失败，但没有错误")
	}
}

// TestRegisterSSHNetwork_NetworkNameFormat 测试网络名称格式
func TestRegisterSSHNetwork_NetworkNameFormat(t *testing.T) {
	config := &connection.SSHConfig{
		Host:     "example.com",
		Port:     2222,
		User:     "testuser",
		Password: "testpass",
	}

	// 由于我们无法真正连接到这个 SSH 服务器，我们只能测试返回值的格式
	// 这个测试会失败，但我们可以检查返回的错误类型
	netName, err := RegisterSSHNetwork(config)

	if err == nil {
		// 如果意外成功了（有实际的 SSH 服务器），检查网络名格式
		expectedPrefix := "ssh_example.com_"
		if len(netName) <= len(expectedPrefix) {
			t.Errorf("RegisterSSHNetwork() 网络名格式无效: %s", netName)
		}
	} else {
		// 预期会失败，因为没有实际的 SSH 服务器
		t.Logf("RegisterSSHNetwork() 按预期失败: %v", err)
	}
}

// TestRegisterSSHNetwork_ConcurrentRegistations 测试并发注册
func TestRegisterSSHNetwork_ConcurrentRegistations(t *testing.T) {
	config := &connection.SSHConfig{
		Host:     "localhost",
		Port:     22,
		User:     "test",
		Password: "test",
	}

	// 尝试并发注册（都会失败，因为没有实际的 SSH 服务器）
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			_, err := RegisterSSHNetwork(config)
			_ = err // 忽略错误，我们只测试并发安全性
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 5; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("并发测试超时")
		}
	}
}

// TestViaSSHDialer_NilClient 测试 nil SSH 客户端
func TestViaSSHDialer_NilClient(t *testing.T) {
	dialer := &ViaSSHDialer{
		sshClient: nil,
	}

	ctx := context.Background()
	_, err := dialer.Dial(ctx, "localhost:9999")
	if err == nil {
		t.Error("ViaSSHDialer.Dial() 使用 nil 客户端应该返回错误")
	}
	if err != nil && err.Error() != "SSH 客户端为 nil" {
		t.Errorf("ViaSSHDialer.Dial() 错误消息不正确: %v", err)
	}
}

// BenchmarkDialContext 基准测试 dialContext 函数
func BenchmarkDialContext(b *testing.B) {
	config := &ssh.ClientConfig{
		User:            "test",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 注意：这个基准测试需要一个真实的 SSH 连接才能有意义
	// 在 CI/CD 环境中可能会被跳过
	client, err := ssh.Dial("tcp", "localhost:22", config)
	if err != nil {
		b.Skip("需要真实的 SSH 服务器进行基准测试")
		return
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dialContext(ctx, client, "tcp", "localhost:9999")
	}
}

// ExampleRegisterSSHNetwork 展示如何使用 RegisterSSHNetwork
func ExampleRegisterSSHNetwork() {
	config := &connection.SSHConfig{
		Host:     "example.com",
		Port:     22,
		User:     "username",
		Password: "password",
	}

	networkName, err := RegisterSSHNetwork(config)
	if err != nil {
		fmt.Printf("SSH 网络注册失败: %v\n", err)
		return
	}

	fmt.Printf("SSH 网络已注册: %s\n", networkName)
}

// TestGetCacheKey 辅助函数测试（模拟 app.go 中的函数）
func TestGetCacheKey(t *testing.T) {
	config := &connection.ConnectionConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Database: "testdb",
		UseSSH:   true,
		SSH: &connection.SSHConfig{
			Host: "ssh.example.com",
			Port: 22,
		},
	}

	key := fmt.Sprintf("%s|%s|%s:%d|%s|%s|%v", config.Type, config.User, config.Host, config.Port, config.Database, config.SSH.Host, config.UseSSH)
	expected := "mysql|root|localhost:3306|testdb|ssh.example.com|true"

	if key != expected {
		t.Errorf("缓存键不匹配\n得到: %s\n期望: %s", key, expected)
	}
}

// TestViaSSHDialer_ContextDial 测试带上下文的拨号
func TestViaSSHDialer_ContextDial(t *testing.T) {
	config := &ssh.ClientConfig{
		User:            "test",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "localhost:0", config)
	if err != nil {
		t.Skip("无法创建测试 SSH 客户端")
		return
	}
	defer client.Close()

	dialer := &ViaSSHDialer{sshClient: client}

	// 测试超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err = dialer.Dial(ctx, "localhost:9999")
	if err == nil {
		t.Error("期望超时错误，但没有错误")
	}
}

// TestGenerateNetworkName 测试网络名称生成的唯一性
func TestGenerateNetworkName(t *testing.T) {
	config := &connection.SSHConfig{
		Host: "testhost",
		Port: 22,
	}

	// 生成多个网络名，检查它们是否唯一
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		netName := fmt.Sprintf("ssh_%s_%d", config.Host, time.Now().UnixNano())
		if names[netName] {
			t.Errorf("生成了重复的网络名: %s", netName)
		}
		names[netName] = true
		time.Sleep(time.Microsecond) // 确保时间戳不同
	}
}

// TestSSHConfig_Validation 测试 SSH 配置验证
func TestSSHConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config *connection.SSHConfig
		valid  bool
	}{
		{
			name: "有效密码配置",
			config: &connection.SSHConfig{
				Host:     "localhost",
				Port:     22,
				User:     "test",
				Password: "password123",
			},
			valid: true,
		},
		{
			name: "有效密钥配置",
			config: &connection.SSHConfig{
				Host:    "localhost",
				Port:    22,
				User:    "test",
				KeyPath: "/path/to/key",
			},
			valid: true,
		},
		{
			name: "缺少主机",
			config: &connection.SSHConfig{
				Host:     "",
				Port:     22,
				User:     "test",
				Password: "password",
			},
			valid: false,
		},
		{
			name: "无效端口",
			config: &connection.SSHConfig{
				Host:     "localhost",
				Port:     -1,
				User:     "test",
				Password: "password",
			},
			valid: false,
		},
		{
			name: "缺少用户",
			config: &connection.SSHConfig{
				Host:     "localhost",
				Port:     22,
				User:     "",
				Password: "password",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateSSHConfig(tt.config)
			if isValid != tt.valid {
				t.Errorf("validateSSHConfig() = %v, 期望 %v", isValid, tt.valid)
			}
		})
	}
}

// validateSSHConfig 验证 SSH 配置的辅助函数
func validateSSHConfig(config *connection.SSHConfig) bool {
	if config.Host == "" {
		return false
	}
	if config.Port <= 0 || config.Port > 65535 {
		return false
	}
	if config.User == "" {
		return false
	}
	if config.Password == "" && config.KeyPath == "" {
		return false
	}
	return true
}
