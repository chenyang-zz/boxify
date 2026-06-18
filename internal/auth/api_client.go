package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultAuthAPIBaseURL = "http://localhost:8000"
	authAPIBaseURLEnv     = "BOXIFY_AUTH_API_BASE_URL"
)

// AuthUser 描述认证接口返回的用户基础信息。
type AuthUser struct {
	ID       string `json:"id"`        // 用户 ID
	Username string `json:"username"`  // 用户名
	IsActive bool   `json:"is_active"` // 是否启用
	IsAdmin  bool   `json:"is_admin"`  // 是否管理员
}

// OAuthAuthorizeResponse 描述 OAuth 授权地址响应。
type OAuthAuthorizeResponse struct {
	AuthorizationURL string `json:"authorization_url"` // 第三方授权地址
}

// LoginResponse 描述 OAuth 回调换取的登录响应。
type LoginResponse struct {
	AccessToken string   `json:"access_token"` // Bearer token
	TokenType   string   `json:"token_type"`   // token 类型
	ExpiresIn   int      `json:"expires_in"`   // 过期秒数
	User        AuthUser `json:"user"`         // 登录用户
}

type apiResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    *T     `json:"data"`
}

// AuthAPIClient 负责调用远端认证接口。
type AuthAPIClient struct {
	baseURL    string       // 认证服务基础地址
	httpClient *http.Client // HTTP 客户端
}

// DefaultAuthAPIBaseURL 返回默认认证服务地址，允许通过环境变量覆盖。
func DefaultAuthAPIBaseURL() string {
	if value := strings.TrimSpace(os.Getenv(authAPIBaseURLEnv)); value != "" {
		return value
	}
	return defaultAuthAPIBaseURL
}

// NewAuthAPIClient 创建认证 API 客户端。
func NewAuthAPIClient(baseURL string, httpClient *http.Client) *AuthAPIClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = DefaultAuthAPIBaseURL()
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &AuthAPIClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// CreateOAuthAuthorization 创建 OAuth 授权地址。
func (c *AuthAPIClient) CreateOAuthAuthorization(ctx context.Context, provider string, frontendCallbackURL string) (OAuthAuthorizeResponse, error) {
	var result OAuthAuthorizeResponse
	endpoint, err := c.endpoint(fmt.Sprintf("/api/auth/oauth/%s/authorize", url.PathEscape(provider)))
	if err != nil {
		return result, err
	}
	query := endpoint.Query()
	if strings.TrimSpace(frontendCallbackURL) != "" {
		query.Set("frontend_redirect_uri", frontendCallbackURL)
	}
	endpoint.RawQuery = query.Encode()

	if err := c.get(ctx, endpoint, &result); err != nil {
		return result, err
	}
	return result, nil
}

// FinalizeOAuthLogin 使用后端一次性 ticket 换取登录 token。
func (c *AuthAPIClient) FinalizeOAuthLogin(ctx context.Context, provider string, ticket string) (LoginResponse, error) {
	var result LoginResponse
	endpoint, err := c.endpoint(fmt.Sprintf("/api/auth/oauth/%s/finalize", url.PathEscape(provider)))
	if err != nil {
		return result, err
	}
	query := endpoint.Query()
	query.Set("ticket", ticket)
	endpoint.RawQuery = query.Encode()

	if err := c.get(ctx, endpoint, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (c *AuthAPIClient) endpoint(path string) (*url.URL, error) {
	parsed, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("认证服务地址无效: %w", err)
	}
	return parsed, nil
}

func (c *AuthAPIClient) get(ctx context.Context, endpoint *url.URL, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("创建认证请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求认证服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("认证服务返回异常状态: %d", resp.StatusCode)
	}

	var wrapper apiResponse[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return fmt.Errorf("解析认证响应失败: %w", err)
	}
	if wrapper.Code != 0 && wrapper.Code != http.StatusOK {
		return fmt.Errorf("认证服务返回失败: %s", wrapper.Message)
	}
	if wrapper.Data == nil {
		return fmt.Errorf("认证服务响应缺少 data")
	}
	if err := json.Unmarshal(*wrapper.Data, result); err != nil {
		return fmt.Errorf("解析认证 data 失败: %w", err)
	}
	return nil
}
