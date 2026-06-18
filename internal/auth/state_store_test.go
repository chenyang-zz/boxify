package auth

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAuthStateStoreMissingFileIsLoggedOut(t *testing.T) {
	store := NewAuthStateStore(filepath.Join(t.TempDir(), "auth-state.json"), newTestLogger())

	loggedIn, err := store.IsLoggedIn()
	if err != nil {
		t.Fatalf("IsLoggedIn() error = %v", err)
	}
	if loggedIn {
		t.Fatal("IsLoggedIn() = true, want false for missing state file")
	}
}

func TestAuthStateStoreMarkLoggedInPersistsState(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth-state.json")
	store := NewAuthStateStore(statePath, newTestLogger())

	if err := store.MarkLoggedIn("github"); err != nil {
		t.Fatalf("MarkLoggedIn() error = %v", err)
	}

	reloaded := NewAuthStateStore(statePath, newTestLogger())
	loggedIn, err := reloaded.IsLoggedIn()
	if err != nil {
		t.Fatalf("IsLoggedIn() error = %v", err)
	}
	if !loggedIn {
		t.Fatal("IsLoggedIn() = false, want true after MarkLoggedIn")
	}
}

func TestAuthStateStoreSaveLoginPersistsTokenSession(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth-state.json")
	store := NewAuthStateStore(statePath, newTestLogger())

	session := LoginResponse{
		AccessToken: "token-123",
		TokenType:   "bearer",
		ExpiresIn:   3600,
		User: AuthUser{
			ID:       "user-1",
			Username: "octocat",
			IsActive: true,
			IsAdmin:  false,
		},
	}
	if err := store.SaveLogin("github", session); err != nil {
		t.Fatalf("SaveLogin() error = %v", err)
	}

	reloaded := NewAuthStateStore(statePath, newTestLogger())
	state, err := reloaded.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if !state.LoggedIn || state.AccessToken != "token-123" || state.User.Username != "octocat" {
		t.Fatalf("state = %#v, want persisted GitHub session", state)
	}
}

func TestAuthStateStoreExpiredTokenIsLoggedOut(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth-state.json")
	store := NewAuthStateStore(statePath, newTestLogger())

	if err := store.writeState(AuthState{
		LoggedIn:    true,
		Provider:    "github",
		AccessToken: "token-123",
		TokenType:   "bearer",
		ExpiresAt:   time.Now().Add(-time.Minute),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}); err != nil {
		t.Fatalf("writeState() error = %v", err)
	}

	loggedIn, err := store.IsLoggedIn()
	if err != nil {
		t.Fatalf("IsLoggedIn() error = %v", err)
	}
	if loggedIn {
		t.Fatal("IsLoggedIn() = true, want false for expired token")
	}
}

func TestAuthStateStoreMissingTokenIsLoggedOut(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth-state.json")
	store := NewAuthStateStore(statePath, newTestLogger())

	if err := store.writeState(AuthState{
		LoggedIn:  true,
		Provider:  "github",
		ExpiresAt: time.Now().Add(time.Hour),
		UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("writeState() error = %v", err)
	}

	loggedIn, err := store.IsLoggedIn()
	if err != nil {
		t.Fatalf("IsLoggedIn() error = %v", err)
	}
	if loggedIn {
		t.Fatal("IsLoggedIn() = true, want false when token is missing")
	}
}

func TestAuthStateStoreLogoutClearsState(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth-state.json")
	store := NewAuthStateStore(statePath, newTestLogger())

	if err := store.MarkLoggedIn("google"); err != nil {
		t.Fatalf("MarkLoggedIn() error = %v", err)
	}
	if err := store.Logout(); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	loggedIn, err := store.IsLoggedIn()
	if err != nil {
		t.Fatalf("IsLoggedIn() error = %v", err)
	}
	if loggedIn {
		t.Fatal("IsLoggedIn() = true, want false after Logout")
	}
}

func TestAuthStateStoreCorruptJSONIsLoggedOutWithError(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "auth-state.json")
	if err := os.WriteFile(statePath, []byte("{broken"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	store := NewAuthStateStore(statePath, newTestLogger())

	loggedIn, err := store.IsLoggedIn()
	if err == nil {
		t.Fatal("IsLoggedIn() error = nil, want error for corrupt state file")
	}
	if loggedIn {
		t.Fatal("IsLoggedIn() = true, want false for corrupt state file")
	}
}
