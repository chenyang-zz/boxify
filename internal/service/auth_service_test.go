package service

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/chenyang-zz/boxify/internal/auth"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type fakeOAuthClient struct {
	t                   *testing.T
	frontendCallbackURL string
	finalizeTicket      string
	finalizeCalled      bool
}

func (c *fakeOAuthClient) CreateOAuthAuthorization(ctx context.Context, provider string, frontendCallbackURL string) (auth.OAuthAuthorizeResponse, error) {
	c.frontendCallbackURL = frontendCallbackURL
	return auth.OAuthAuthorizeResponse{
		AuthorizationURL: "https://github.com/login/oauth/authorize?state=state-123&frontend_redirect_uri=" + url.QueryEscape(frontendCallbackURL),
	}, nil
}

func (c *fakeOAuthClient) FinalizeOAuthLogin(ctx context.Context, provider string, ticket string) (auth.LoginResponse, error) {
	c.finalizeCalled = true
	if ticket != "ticket-123" {
		c.t.Fatalf("finalize ticket = %q", ticket)
	}
	c.finalizeTicket = ticket
	return auth.LoginResponse{
		AccessToken: "token-123",
		TokenType:   "bearer",
		ExpiresIn:   3600,
		User: auth.AuthUser{
			ID:       "user-1",
			Username: "octocat",
			IsActive: true,
			IsAdmin:  false,
		},
	}, nil
}

func newAuthServiceForTest(t *testing.T) *AuthService {
	t.Helper()

	app := application.New(application.Options{
		LogLevel: slog.LevelInfo,
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := auth.NewAuthStateStore(filepath.Join(t.TempDir(), "auth-state.json"), logger)
	return NewAuthServiceWithStore(NewServiceDeps(app, nil), store)
}

func TestAuthServiceGetLoginStateDefaultsToLoggedOut(t *testing.T) {
	service := newAuthServiceForTest(t)

	result := service.GetLoginState()
	if !result.Success {
		t.Fatalf("GetLoginState().Success = false, message = %q", result.Message)
	}
	if loggedIn, ok := result.Data.(bool); !ok || loggedIn {
		t.Fatalf("GetLoginState().Data = %#v, want false", result.Data)
	}
}

func TestAuthServiceMarkLoggedInAndLogout(t *testing.T) {
	service := newAuthServiceForTest(t)

	loginResult := service.MarkLoggedIn("github")
	if !loginResult.Success {
		t.Fatalf("MarkLoggedIn().Success = false, message = %q", loginResult.Message)
	}

	stateResult := service.GetLoginState()
	if loggedIn, ok := stateResult.Data.(bool); !ok || !loggedIn {
		t.Fatalf("GetLoginState().Data = %#v, want true", stateResult.Data)
	}

	logoutResult := service.Logout()
	if !logoutResult.Success {
		t.Fatalf("Logout().Success = false, message = %q", logoutResult.Message)
	}

	stateResult = service.GetLoginState()
	if loggedIn, ok := stateResult.Data.(bool); !ok || loggedIn {
		t.Fatalf("GetLoginState().Data = %#v, want false after Logout", stateResult.Data)
	}
}

func TestAuthServiceStartOAuthLoginOpensBrowserWithFrontendCallback(t *testing.T) {
	service := newAuthServiceForTest(t)
	fakeClient := &fakeOAuthClient{t: t}
	service.oauthClient = fakeClient
	service.oauthTimeout = 3 * time.Second
	var openedURL string
	service.openBrowser = func(rawURL string) error {
		openedURL = rawURL
		return nil
	}

	result := service.StartOAuthLogin("github")
	if !result.Success {
		t.Fatalf("StartOAuthLogin().Success = false, message = %q", result.Message)
	}
	if fakeClient.frontendCallbackURL != defaultOAuthFrontendCallbackURL {
		t.Fatalf("frontendCallbackURL = %q", fakeClient.frontendCallbackURL)
	}
	if openedURL == "" {
		t.Fatal("browser was not opened")
	}
	authURL, err := url.Parse(openedURL)
	if err != nil {
		t.Fatalf("Parse authorization URL error = %v", err)
	}
	if authURL.Query().Get("state") != "state-123" {
		t.Fatalf("authorization state = %q", authURL.Query().Get("state"))
	}
}

func TestAuthServiceHandleOAuthDeepLinkFinalizesAndSavesLogin(t *testing.T) {
	service := newAuthServiceForTest(t)
	fakeClient := &fakeOAuthClient{t: t}
	service.oauthClient = fakeClient
	service.openBrowser = func(rawURL string) error { return nil }
	result := service.StartOAuthLogin("github")
	if !result.Success {
		t.Fatalf("StartOAuthLogin().Success = false, message = %q", result.Message)
	}

	callbackResult := service.handleOAuthDeepLink("boxify://auth/callback?provider=github&ticket=ticket-123&state=state-123")
	if !callbackResult.Success {
		t.Fatalf("handleOAuthDeepLink().Success = false, message = %q", callbackResult.Message)
	}
	if fakeClient.finalizeTicket != "ticket-123" {
		t.Fatalf("finalizeTicket = %q", fakeClient.finalizeTicket)
	}

	stateResult := service.GetLoginState()
	if loggedIn, ok := stateResult.Data.(bool); !ok || !loggedIn {
		t.Fatalf("GetLoginState().Data = %#v, want true", stateResult.Data)
	}
}

func TestAuthServiceHandleOAuthDeepLinkSavesFragmentLoginResponse(t *testing.T) {
	service := newAuthServiceForTest(t)
	fakeClient := &fakeOAuthClient{t: t}
	service.oauthClient = fakeClient
	service.openBrowser = func(rawURL string) error { return nil }
	result := service.StartOAuthLogin("github")
	if !result.Success {
		t.Fatalf("StartOAuthLogin().Success = false, message = %q", result.Message)
	}

	callbackURL := "boxify://auth/callback#access_token=token-fragment&token_type=bearer&expires_in=3600&user_id=user-1&username=octocat&is_active=true&is_admin=false"
	callbackResult := service.handleOAuthDeepLink(callbackURL)
	if !callbackResult.Success {
		t.Fatalf("handleOAuthDeepLink().Success = false, message = %q", callbackResult.Message)
	}
	if fakeClient.finalizeCalled {
		t.Fatal("FinalizeOAuthLogin was called for fragment login response")
	}

	state, err := service.store.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if !state.LoggedIn || state.AccessToken != "token-fragment" || state.User.Username != "octocat" {
		t.Fatalf("saved state = %#v", state)
	}
}

func TestAuthServiceHandleOAuthDeepLinkRejectsStateMismatch(t *testing.T) {
	service := newAuthServiceForTest(t)
	service.oauthClient = &fakeOAuthClient{t: t}
	service.openBrowser = func(rawURL string) error { return nil }
	result := service.StartOAuthLogin("github")
	if !result.Success {
		t.Fatalf("StartOAuthLogin().Success = false, message = %q", result.Message)
	}

	callbackResult := service.handleOAuthDeepLink("boxify://auth/callback?provider=github&ticket=ticket-123&state=wrong-state")
	if callbackResult.Success {
		t.Fatal("handleOAuthDeepLink().Success = true, want false for state mismatch")
	}
}

func TestAuthServiceHandleOAuthDeepLinkRejectsInvalidCallbackURL(t *testing.T) {
	service := newAuthServiceForTest(t)

	tests := []struct {
		name   string
		rawURL string
	}{
		{name: "错误 scheme", rawURL: "https://auth/callback?provider=github&ticket=ticket-123&state=state-123"},
		{name: "缺少 ticket 或 fragment token", rawURL: "boxify://auth/callback?provider=github&state=state-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.handleOAuthDeepLink(tt.rawURL)
			if result.Success {
				t.Fatalf("handleOAuthDeepLink(%q).Success = true, want false", tt.rawURL)
			}
		})
	}
}
