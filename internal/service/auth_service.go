package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/auth"
	"github.com/chenyang-zz/boxify/internal/connection"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	defaultOAuthTimeout             = 2 * time.Minute
	defaultOAuthFrontendCallbackURL = "boxify://auth/callback"
	authOAuthCompletedEventName     = "auth:oauth-completed"
)

type oauthClient interface {
	CreateOAuthAuthorization(ctx context.Context, provider string, frontendCallbackURL string) (auth.OAuthAuthorizeResponse, error)
	FinalizeOAuthLogin(ctx context.Context, provider string, ticket string) (auth.LoginResponse, error)
}

type browserOpener func(url string) error

// AuthOAuthCompletedEvent 描述 OAuth 登录完成事件。
type AuthOAuthCompletedEvent struct {
	Success   bool          `json:"success"`             // 是否登录成功
	Message   string        `json:"message"`             // 中文结果消息
	Provider  string        `json:"provider,omitempty"`  // 登录来源
	User      auth.AuthUser `json:"user,omitempty"`      // 登录用户
	ExpiresIn int           `json:"expiresIn,omitempty"` // token 过期秒数
}

// AuthAccessTokenResponse 描述前端直连 API 所需的访问 token。
type AuthAccessTokenResponse struct {
	AccessToken string `json:"accessToken"` // 访问 token
	TokenType   string `json:"tokenType"`   // token 类型
}

type oauthEventEmitter func(event AuthOAuthCompletedEvent)

// AuthService 提供登录状态读写能力。
type AuthService struct {
	BaseService
	store                  *auth.AuthStateStore // 本地登录状态存储
	oauthClient            oauthClient          // OAuth 接口客户端
	openBrowser            browserOpener        // 浏览器打开函数
	emitOAuthEvent         oauthEventEmitter    // OAuth 完成事件发送函数
	oauthTimeout           time.Duration        // OAuth 请求超时时间
	cancelDeepLinkListener func()               // 系统 Deep Link 监听取消函数
	mu                     sync.RWMutex         // OAuth pending 状态锁
	pendingStates          map[string]string    // provider -> state
}

// NewAuthService 创建 AuthService。
func NewAuthService(deps *ServiceDeps) *AuthService {
	base := NewBaseService(deps)
	return NewAuthServiceWithStore(deps, auth.NewAuthStateStore("", base.Logger()))
}

// NewAuthServiceWithStore 使用指定状态存储创建 AuthService，便于测试注入。
func NewAuthServiceWithStore(deps *ServiceDeps, store *auth.AuthStateStore) *AuthService {
	base := NewBaseService(deps)
	if store == nil {
		store = auth.NewAuthStateStore("", base.Logger())
	}
	return &AuthService{
		BaseService: base,
		store:       store,
		oauthClient: auth.NewAuthAPIClient("", nil),
		openBrowser: func(rawURL string) error { return deps.App().Browser.OpenURL(rawURL) },
		emitOAuthEvent: func(event AuthOAuthCompletedEvent) {
			deps.App().Event.Emit(authOAuthCompletedEventName, event)
		},
		oauthTimeout:  defaultOAuthTimeout,
		pendingStates: map[string]string{},
	}
}

// ServiceStartup 是在应用程序启动时调用的函数。
func (as *AuthService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	as.SetContext(ctx)
	as.Logger().Info("服务启动", "service", "AuthService")
	as.cancelDeepLinkListener = as.App().Event.OnApplicationEvent(events.Common.ApplicationLaunchedWithUrl, func(event *application.ApplicationEvent) {
		rawURL := strings.TrimSpace(event.Context().URL())
		if rawURL == "" {
			return
		}
		go as.handleOAuthDeepLink(rawURL)
	})
	return nil
}

// ServiceShutdown 服务关闭。
func (as *AuthService) ServiceShutdown() error {
	if as.cancelDeepLinkListener != nil {
		as.cancelDeepLinkListener()
		as.cancelDeepLinkListener = nil
	}
	as.Logger().Info("服务关闭", "service", "AuthService")
	return nil
}

// GetLoginState 获取当前登录状态。
func (as *AuthService) GetLoginState() *connection.QueryResult {
	loggedIn, err := as.store.IsLoggedIn()
	if err != nil {
		return &connection.QueryResult{
			Success: true,
			Message: fmt.Sprintf("读取登录状态失败，已按未登录处理: %s", err.Error()),
			Data:    false,
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "登录状态读取成功",
		Data:    loggedIn,
	}
}

// GetAccessToken 获取当前有效访问 token，供前端直连远端 API 使用。
func (as *AuthService) GetAccessToken() *connection.QueryResult {
	state, err := as.store.GetState()
	if err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("读取登录状态失败: %s", err.Error()),
		}
	}
	if !state.LoggedIn || strings.TrimSpace(state.AccessToken) == "" {
		return &connection.QueryResult{
			Success: false,
			Message: "登录已过期，请重新登录",
		}
	}

	tokenType := strings.TrimSpace(state.TokenType)
	if tokenType == "" {
		tokenType = "bearer"
	}
	return &connection.QueryResult{
		Success: true,
		Message: "访问 token 读取成功",
		Data: AuthAccessTokenResponse{
			AccessToken: state.AccessToken,
			TokenType:   tokenType,
		},
	}
}

// MarkLoggedIn 标记当前用户已登录。
func (as *AuthService) MarkLoggedIn(provider string) *connection.QueryResult {
	if err := as.store.MarkLoggedIn(provider); err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("写入登录状态失败: %s", err.Error()),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "登录状态已保存",
		Data:    true,
	}
}

// StartOAuthLogin 发起 OAuth 登录流程，打开浏览器后等待 Deep Link 回调。
func (as *AuthService) StartOAuthLogin(provider string) *connection.QueryResult {
	provider = normalizeOAuthProvider(provider)
	if provider != "github" {
		return &connection.QueryResult{
			Success: false,
			Message: "暂未支持该登录方式",
		}
	}

	as.Logger().Info("开始 OAuth 登录", "provider", provider)
	ctx, cancel := context.WithTimeout(as.Context(), as.oauthTimeout)
	defer cancel()

	authorize, err := as.oauthClient.CreateOAuthAuthorization(ctx, provider, defaultOAuthFrontendCallbackURL)
	if err != nil {
		as.Logger().Warn("创建 OAuth 授权地址失败", "provider", provider, "error", err)
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("创建 GitHub 授权地址失败: %s", err.Error()),
		}
	}

	state, err := extractAuthorizationState(authorize.AuthorizationURL)
	if err != nil {
		as.Logger().Warn("OAuth 授权地址无效", "provider", provider, "error", err)
		return &connection.QueryResult{
			Success: false,
			Message: err.Error(),
		}
	}
	as.setPendingState(provider, state)

	if as.openBrowser == nil {
		as.clearPendingState(provider)
		return &connection.QueryResult{
			Success: false,
			Message: "浏览器打开函数未配置",
		}
	}
	if err := as.openBrowser(authorize.AuthorizationURL); err != nil {
		as.clearPendingState(provider)
		as.Logger().Warn("打开 GitHub 授权页失败", "provider", provider, "error", err)
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("打开 GitHub 授权页失败: %s", err.Error()),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "已打开 GitHub 授权页",
		Data: map[string]string{
			"provider":            provider,
			"frontendCallbackURL": defaultOAuthFrontendCallbackURL,
		},
	}
}

// LoginWithOAuth 兼容旧绑定，当前行为等同于 StartOAuthLogin。
func (as *AuthService) LoginWithOAuth(provider string) *connection.QueryResult {
	return as.StartOAuthLogin(provider)
}

// Logout 清除当前用户登录状态。
func (as *AuthService) Logout() *connection.QueryResult {
	if err := as.store.Logout(); err != nil {
		return &connection.QueryResult{
			Success: false,
			Message: fmt.Sprintf("退出登录失败: %s", err.Error()),
		}
	}

	return &connection.QueryResult{
		Success: true,
		Message: "已退出登录",
		Data:    false,
	}
}

func (as *AuthService) handleOAuthDeepLink(rawURL string) *connection.QueryResult {
	session, provider, err := as.completeOAuthDeepLink(rawURL)
	if err != nil {
		as.Logger().Warn("OAuth Deep Link 登录失败", "url", rawURL, "error", err)
		event := AuthOAuthCompletedEvent{
			Success:  false,
			Message:  fmt.Sprintf("GitHub 登录失败: %s", err.Error()),
			Provider: provider,
		}
		as.emitOAuthCompleted(event)
		return &connection.QueryResult{
			Success: false,
			Message: event.Message,
		}
	}

	event := AuthOAuthCompletedEvent{
		Success:   true,
		Message:   "GitHub 登录成功",
		Provider:  provider,
		User:      session.User,
		ExpiresIn: session.ExpiresIn,
	}
	as.emitOAuthCompleted(event)
	as.Logger().Info("OAuth Deep Link 登录完成", "provider", provider, "user", session.User.Username)
	return &connection.QueryResult{
		Success: true,
		Message: event.Message,
		Data:    session,
	}
}

func (as *AuthService) completeOAuthDeepLink(rawURL string) (auth.LoginResponse, string, error) {
	var result auth.LoginResponse
	parsed, err := parseOAuthDeepLink(rawURL)
	if err != nil {
		return result, "", err
	}
	if parsed.Provider != "github" {
		return result, parsed.Provider, fmt.Errorf("暂未支持该登录方式")
	}
	if expected := as.pendingState(parsed.Provider); expected != "" && parsed.State != "" && parsed.State != expected {
		return result, parsed.Provider, fmt.Errorf("GitHub 回调 state 校验失败")
	}

	if parsed.Session != nil {
		result = *parsed.Session
		if err := as.store.SaveLogin(parsed.Provider, result); err != nil {
			return result, parsed.Provider, err
		}
		as.clearPendingState(parsed.Provider)
		return result, parsed.Provider, nil
	}

	ctx, cancel := context.WithTimeout(as.Context(), as.oauthTimeout)
	defer cancel()
	result, err = as.oauthClient.FinalizeOAuthLogin(ctx, parsed.Provider, parsed.Ticket)
	if err != nil {
		return result, parsed.Provider, err
	}
	if err := as.store.SaveLogin(parsed.Provider, result); err != nil {
		return result, parsed.Provider, err
	}
	as.clearPendingState(parsed.Provider)
	return result, parsed.Provider, nil
}

func (as *AuthService) emitOAuthCompleted(event AuthOAuthCompletedEvent) {
	if as.emitOAuthEvent != nil {
		as.emitOAuthEvent(event)
	}
}

func (as *AuthService) setPendingState(provider string, state string) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.pendingStates[provider] = state
}

func (as *AuthService) pendingState(provider string) string {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.pendingStates[provider]
}

func (as *AuthService) clearPendingState(provider string) {
	as.mu.Lock()
	defer as.mu.Unlock()
	delete(as.pendingStates, provider)
}

func normalizeOAuthProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func extractAuthorizationState(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("授权地址无效: %w", err)
	}
	state := strings.TrimSpace(parsed.Query().Get("state"))
	if state == "" {
		return "", fmt.Errorf("授权地址缺少 state")
	}
	return state, nil
}

type oauthDeepLink struct {
	Provider string              // 登录来源
	Ticket   string              // 后端一次性 ticket
	State    string              // OAuth state
	Session  *auth.LoginResponse // fragment 携带的登录结果
}

func parseOAuthDeepLink(rawURL string) (oauthDeepLink, error) {
	var result oauthDeepLink
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return result, fmt.Errorf("登录回调地址无效: %w", err)
	}
	if parsed.Scheme != "boxify" || parsed.Host != "auth" || parsed.Path != "/callback" {
		return result, fmt.Errorf("登录回调地址不受支持")
	}

	query := parsed.Query()
	fragment, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		return result, fmt.Errorf("解析登录回调 fragment 失败: %w", err)
	}
	if errText := firstQueryValue(query, fragment, "error"); errText != "" {
		return result, fmt.Errorf("GitHub 授权被拒绝: %s", errText)
	}

	result.Provider = normalizeOAuthProvider(firstQueryValue(query, fragment, "provider"))
	result.Ticket = strings.TrimSpace(firstQueryValue(query, fragment, "ticket"))
	result.State = strings.TrimSpace(firstQueryValue(query, fragment, "state"))
	if result.Provider == "" {
		result.Provider = "github"
	}

	if strings.TrimSpace(fragment.Get("access_token")) != "" {
		session, err := loginResponseFromFragment(fragment)
		if err != nil {
			return result, err
		}
		result.Session = &session
		return result, nil
	}

	if result.Ticket == "" {
		return result, fmt.Errorf("登录回调缺少 ticket 或 access_token")
	}
	return result, nil
}

func loginResponseFromFragment(fragment url.Values) (auth.LoginResponse, error) {
	var result auth.LoginResponse
	result.AccessToken = strings.TrimSpace(fragment.Get("access_token"))
	result.TokenType = strings.TrimSpace(fragment.Get("token_type"))
	if result.TokenType == "" {
		result.TokenType = "bearer"
	}
	expiresIn, err := strconv.Atoi(strings.TrimSpace(fragment.Get("expires_in")))
	if err != nil || expiresIn <= 0 {
		return result, fmt.Errorf("登录回调 expires_in 无效")
	}
	result.ExpiresIn = expiresIn
	result.User = auth.AuthUser{
		ID:       strings.TrimSpace(fragment.Get("user_id")),
		Username: strings.TrimSpace(fragment.Get("username")),
		IsActive: parseBoolFragment(fragment.Get("is_active")),
		IsAdmin:  parseBoolFragment(fragment.Get("is_admin")),
	}
	if result.AccessToken == "" {
		return result, fmt.Errorf("登录回调缺少 access_token")
	}
	if result.User.ID == "" || result.User.Username == "" {
		return result, fmt.Errorf("登录回调缺少用户信息")
	}
	return result, nil
}

func firstQueryValue(primary url.Values, fallback url.Values, key string) string {
	if value := strings.TrimSpace(primary.Get(key)); value != "" {
		return value
	}
	return strings.TrimSpace(fallback.Get(key))
}

func parseBoolFragment(value string) bool {
	parsed, _ := strconv.ParseBool(strings.TrimSpace(value))
	return parsed
}
