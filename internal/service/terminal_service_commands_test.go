package service

import (
	"log/slog"
	"testing"

	"github.com/chenyang-zz/boxify/internal/terminal"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestListExecutableCommandsWithShellType(t *testing.T) {
	app := application.New(application.Options{
		LogLevel: slog.LevelInfo,
	})
	service := NewTerminalService(NewServiceDeps(app, nil))

	result := service.ListExecutableCommands(terminal.ShellTypeBash)
	if !result.Success {
		t.Fatalf("expected success, got message: %s", result.Message)
	}
	if result.Data == nil {
		t.Fatal("expected data not nil")
	}
	if result.Data.ResolvedShell != string(terminal.ShellTypeBash) {
		t.Fatalf("expected resolved shell bash, got %s", result.Data.ResolvedShell)
	}
	if len(result.Data.DefaultCommands) == 0 {
		t.Fatal("expected non-empty default commands")
	}
}

func TestListExecutableCommandsWithInvalidShellType(t *testing.T) {
	app := application.New(application.Options{
		LogLevel: slog.LevelInfo,
	})
	service := NewTerminalService(NewServiceDeps(app, nil))

	result := service.ListExecutableCommands(terminal.ShellType("fish"))
	if result.Success {
		t.Fatal("expected failure for unsupported shell type")
	}
}
