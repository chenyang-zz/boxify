package service

import (
	"context"
	"log/slog"
	"os"
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

func TestWriteCommandWithBlock_ReusePreferredBlockID(t *testing.T) {
	app := application.New(application.Options{
		LogLevel: slog.LevelInfo,
	})
	service := NewTerminalService(NewServiceDeps(app, nil))
	if err := service.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("service startup failed: %v", err)
	}
	t.Cleanup(func() {
		_ = service.ServiceShutdown()
	})

	ptyFile, err := os.CreateTemp("", "terminal-pty-*")
	if err != nil {
		t.Fatalf("create temp pty file failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(ptyFile.Name())
		_ = ptyFile.Close()
	})

	session := terminal.NewSession(
		service.Context(),
		"session-1",
		ptyFile,
		nil,
		terminal.ShellTypeBash,
		false,
		app.Logger,
	)
	service.sessionManager.Add(session)

	blockID, err := service.WriteCommandWithBlock("session-1", "block-custom", "echo hello")
	if err != nil {
		t.Fatalf("write command with block failed: %v", err)
	}
	if blockID != "block-custom" {
		t.Fatalf("expected block id block-custom, got %s", blockID)
	}
	if session.CurrentBlock() != "block-custom" {
		t.Fatalf("expected session current block block-custom, got %s", session.CurrentBlock())
	}
}

func TestWriteCommandWithBlock_GenerateWhenPreferredBlockIDEmpty(t *testing.T) {
	app := application.New(application.Options{
		LogLevel: slog.LevelInfo,
	})
	service := NewTerminalService(NewServiceDeps(app, nil))
	if err := service.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("service startup failed: %v", err)
	}
	t.Cleanup(func() {
		_ = service.ServiceShutdown()
	})

	ptyFile, err := os.CreateTemp("", "terminal-pty-*")
	if err != nil {
		t.Fatalf("create temp pty file failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(ptyFile.Name())
		_ = ptyFile.Close()
	})

	session := terminal.NewSession(
		service.Context(),
		"session-2",
		ptyFile,
		nil,
		terminal.ShellTypeBash,
		false,
		app.Logger,
	)
	service.sessionManager.Add(session)

	blockID, err := service.WriteCommandWithBlock("session-2", "  ", "echo world")
	if err != nil {
		t.Fatalf("write command with empty preferred block failed: %v", err)
	}
	if blockID == "" {
		t.Fatal("expected generated block id, got empty")
	}
	if session.CurrentBlock() != blockID {
		t.Fatalf("expected session current block %s, got %s", blockID, session.CurrentBlock())
	}
}
