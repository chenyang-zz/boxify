package auth

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AuthState 描述本地登录状态文件内容。
type AuthState struct {
	LoggedIn    bool      `json:"loggedIn"`              // 是否已登录
	Provider    string    `json:"provider,omitempty"`    // 登录来源
	AccessToken string    `json:"accessToken,omitempty"` // 访问 token
	TokenType   string    `json:"tokenType,omitempty"`   // token 类型
	ExpiresAt   time.Time `json:"expiresAt,omitempty"`   // token 过期时间
	User        AuthUser  `json:"user,omitempty"`        // 登录用户
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`   // 状态更新时间
}

// AuthStateStore 负责读写本地登录状态。
type AuthStateStore struct {
	path   string       // 状态文件路径
	logger *slog.Logger // 日志记录器
}

// DefaultStatePath 返回默认登录状态文件路径。
func DefaultStatePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		return filepath.Join(".", "auth-state.json")
	}
	return filepath.Join(configDir, "Boxify", "auth-state.json")
}

// NewAuthStateStore 创建本地登录状态存储。
func NewAuthStateStore(path string, logger *slog.Logger) *AuthStateStore {
	if strings.TrimSpace(path) == "" {
		path = DefaultStatePath()
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthStateStore{
		path:   path,
		logger: logger,
	}
}

// Path 返回当前状态文件路径，便于测试和诊断。
func (s *AuthStateStore) Path() string {
	return s.path
}

// IsLoggedIn 判断当前本地状态是否已登录。
func (s *AuthStateStore) IsLoggedIn() (bool, error) {
	state, err := s.GetState()
	if err != nil {
		return false, err
	}
	return state.LoggedIn, nil
}

// GetState 读取并校验当前登录状态。
func (s *AuthStateStore) GetState() (AuthState, error) {
	state, err := s.readState()
	if err != nil {
		return AuthState{}, err
	}
	return s.normalizeState(state), nil
}

// MarkLoggedIn 写入已登录状态。
func (s *AuthStateStore) MarkLoggedIn(provider string) error {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "local"
	}

	state := AuthState{
		LoggedIn:    true,
		Provider:    provider,
		AccessToken: "local",
		TokenType:   "local",
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
		UpdatedAt:   time.Now(),
	}
	return s.writeState(state)
}

// SaveLogin 保存真实登录会话。
func (s *AuthStateStore) SaveLogin(provider string, session LoginResponse) error {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "github"
	}
	tokenType := strings.TrimSpace(session.TokenType)
	if tokenType == "" {
		tokenType = "bearer"
	}
	if strings.TrimSpace(session.AccessToken) == "" {
		return fmt.Errorf("登录 token 为空")
	}
	if session.ExpiresIn <= 0 {
		return fmt.Errorf("登录 token 过期时间无效")
	}

	state := AuthState{
		LoggedIn:    true,
		Provider:    provider,
		AccessToken: session.AccessToken,
		TokenType:   tokenType,
		ExpiresAt:   time.Now().Add(time.Duration(session.ExpiresIn) * time.Second),
		User:        session.User,
		UpdatedAt:   time.Now(),
	}
	return s.writeState(state)
}

// Logout 清除本地登录状态。
func (s *AuthStateStore) Logout() error {
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("清除登录状态失败: %w", err)
	}
	return nil
}

// readState 读取登录状态，缺失或损坏时按未登录处理。
func (s *AuthStateStore) readState() (AuthState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("登录状态文件不存在，按未登录处理", "path", s.path)
			return AuthState{}, nil
		}
		s.logger.Warn("读取登录状态文件失败，按未登录处理", "path", s.path, "error", err)
		return AuthState{}, fmt.Errorf("读取登录状态失败: %w", err)
	}

	var state AuthState
	if err := json.Unmarshal(data, &state); err != nil {
		s.logger.Warn("解析登录状态文件失败，按未登录处理", "path", s.path, "error", err)
		return AuthState{}, fmt.Errorf("解析登录状态失败: %w", err)
	}
	return state, nil
}

// writeState 写入登录状态文件。
func (s *AuthStateStore) writeState(state AuthState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return fmt.Errorf("创建登录状态目录失败: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化登录状态失败: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("写入登录状态失败: %w", err)
	}
	return nil
}

func (s *AuthStateStore) normalizeState(state AuthState) AuthState {
	if !state.LoggedIn {
		return state
	}
	if strings.TrimSpace(state.AccessToken) == "" {
		s.logger.Warn("登录状态缺少 token，按未登录处理", "path", s.path)
		state.LoggedIn = false
		return state
	}
	if !state.ExpiresAt.IsZero() && time.Now().After(state.ExpiresAt) {
		s.logger.Warn("登录 token 已过期，按未登录处理", "path", s.path, "expiresAt", state.ExpiresAt)
		state.LoggedIn = false
		return state
	}
	return state
}
