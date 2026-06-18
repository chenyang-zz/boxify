package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthAPIClientCreateOAuthAuthorizationIncludesFrontendRedirectURI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/oauth/github/authorize" {
			t.Fatalf("path = %q, want authorize endpoint", r.URL.Path)
		}
		if got := r.URL.Query().Get("frontend_redirect_uri"); got != "boxify://auth/callback" {
			t.Fatalf("frontend_redirect_uri = %q", got)
		}

		writeJSON(t, w, apiResponse[OAuthAuthorizeResponse]{
			Code:    200,
			Message: "success",
			Data: &OAuthAuthorizeResponse{
				AuthorizationURL: "https://github.com/login/oauth/authorize?state=state-123",
			},
		})
	}))
	defer server.Close()

	client := NewAuthAPIClient(server.URL, server.Client())
	result, err := client.CreateOAuthAuthorization(context.Background(), "github", "boxify://auth/callback")
	if err != nil {
		t.Fatalf("CreateOAuthAuthorization() error = %v", err)
	}
	if result.AuthorizationURL != "https://github.com/login/oauth/authorize?state=state-123" {
		t.Fatalf("AuthorizationURL = %q", result.AuthorizationURL)
	}
}

func TestAuthAPIClientFinalizeOAuthLoginParsesLoginResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/oauth/github/finalize" {
			t.Fatalf("path = %q, want finalize endpoint", r.URL.Path)
		}
		if got := r.URL.Query().Get("ticket"); got != "ticket-123" {
			t.Fatalf("ticket = %q", got)
		}

		writeJSON(t, w, apiResponse[LoginResponse]{
			Code:    200,
			Message: "success",
			Data: &LoginResponse{
				AccessToken: "token-123",
				TokenType:   "bearer",
				ExpiresIn:   3600,
				User: AuthUser{
					ID:       "user-1",
					Username: "octocat",
					IsActive: true,
					IsAdmin:  false,
				},
			},
		})
	}))
	defer server.Close()

	client := NewAuthAPIClient(server.URL, server.Client())
	result, err := client.FinalizeOAuthLogin(context.Background(), "github", "ticket-123")
	if err != nil {
		t.Fatalf("FinalizeOAuthLogin() error = %v", err)
	}
	if result.AccessToken != "token-123" || result.User.Username != "octocat" {
		t.Fatalf("login result = %#v", result)
	}
}

func TestAuthAPIClientReturnsAPIErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, apiResponse[OAuthAuthorizeResponse]{
			Code:    400,
			Message: "provider 不支持",
		})
	}))
	defer server.Close()

	client := NewAuthAPIClient(server.URL, server.Client())
	_, err := client.CreateOAuthAuthorization(context.Background(), "github", "boxify://auth/callback")
	if err == nil {
		t.Fatal("CreateOAuthAuthorization() error = nil, want API error")
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
}
