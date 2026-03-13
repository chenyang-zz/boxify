package conversationstore

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	clawchat "github.com/chenyang-zz/boxify/internal/claw/chat"
)

// TestJSONConversationStorePersistence 验证会话和消息可从 JSON 文件恢复。
func TestJSONConversationStorePersistence(t *testing.T) {
	t.Parallel()

	storeDir := filepath.Join(t.TempDir(), "chat")
	store := NewJSONConversationStore(storeDir, slog.Default())

	conv, err := store.CreateConversation("")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if conv.AgentID != "main" {
		t.Fatalf("CreateConversation() agentID = %q, want %q", conv.AgentID, "main")
	}

	msg := clawchat.Message{
		ID:             "msg_user_1",
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "你好",
		Status:         "done",
		CreatedAt:      time.Now(),
	}
	if err := store.AppendMessage(msg); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}
	if err := store.UpdateAssistantDraft(conv.ID, "run_1", "片段A"); err != nil {
		t.Fatalf("UpdateAssistantDraft() first error = %v", err)
	}
	if err := store.UpdateAssistantDraft(conv.ID, "run_1", "片段B"); err != nil {
		t.Fatalf("UpdateAssistantDraft() second error = %v", err)
	}
	if err := store.FinalizeAssistantMessage(conv.ID, "run_1"); err != nil {
		t.Fatalf("FinalizeAssistantMessage() error = %v", err)
	}
	if err := store.UpdateOpenClawSessionID(conv.ID, "session_1"); err != nil {
		t.Fatalf("UpdateOpenClawSessionID() error = %v", err)
	}

	reloaded := NewJSONConversationStore(storeDir, slog.Default())
	if reloaded == nil {
		t.Fatalf("NewJSONConversationStore() reload error")
	}

	gotConv, err := reloaded.GetConversation(conv.ID)
	if err != nil {
		t.Fatalf("GetConversation() error = %v", err)
	}
	if gotConv.OpenClawSessionID != "session_1" {
		t.Fatalf("GetConversation() sessionID = %q, want %q", gotConv.OpenClawSessionID, "session_1")
	}

	messages, err := reloaded.ListMessages(conv.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("ListMessages() len = %d, want %d", len(messages), 2)
	}
	if messages[1].Content != "片段A片段B" {
		t.Fatalf("assistant content = %q, want %q", messages[1].Content, "片段A片段B")
	}
	if messages[1].Status != "done" {
		t.Fatalf("assistant status = %q, want %q", messages[1].Status, "done")
	}
}

// TestJSONConversationStoreNotFound 验证不存在的会话返回统一错误。
func TestJSONConversationStoreNotFound(t *testing.T) {
	t.Parallel()

	storeDir := t.TempDir()
	store := NewJSONConversationStore(storeDir, slog.Default())
	if store == nil {
		t.Fatalf("NewJSONConversationStore() error")
	}

	if _, err := store.GetConversation("missing"); !errors.Is(err, clawchat.ErrConversationNotFound) {
		t.Fatalf("GetConversation() error = %v, want %v", err, clawchat.ErrConversationNotFound)
	}
	if _, err := store.ListMessages("missing"); !errors.Is(err, clawchat.ErrConversationNotFound) {
		t.Fatalf("ListMessages() error = %v, want %v", err, clawchat.ErrConversationNotFound)
	}
}

// TestNewJSONConversationStoreUsesDefaultPath 验证空路径时会回退到默认文件位置。
func TestNewJSONConversationStoreUsesDefaultPath(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("BOXIFY_CONVERSATION_STORE_DIR", filepath.Join(configDir, "custom"))

	store := NewJSONConversationStore("", slog.Default())
	if store == nil {
		t.Fatalf("NewJSONConversationStore() returned nil")
	}

	conv, err := store.CreateConversation("main")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if conv == nil {
		t.Fatalf("CreateConversation() returned nil conversation")
	}

	storeFile := filepath.Join(configDir, "custom", "conversations.json")
	if _, err := os.Stat(storeFile); err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
}

// TestResolveBoxifyDataDir 验证默认路径位于 Boxify 数据目录下。
func TestResolveBoxifyDataDir(t *testing.T) {
	customDataDir := filepath.Join(t.TempDir(), ".boxify-test")
	t.Setenv("BOXIFY_DATA_DIR", customDataDir)

	got, err := resolveBoxifyDataDir()
	if err != nil {
		t.Fatalf("resolveBoxifyDataDir() error = %v", err)
	}
	if got != customDataDir {
		t.Fatalf("resolveBoxifyDataDir() = %q, want %q", got, customDataDir)
	}

	gotPath, err := resolveDefaultStorePath()
	if err != nil {
		t.Fatalf("resolveDefaultStorePath() error = %v", err)
	}
	wantPath := filepath.Join(customDataDir, "conversations.json")
	if gotPath != wantPath {
		t.Fatalf("resolveDefaultStorePath() = %q, want %q", gotPath, wantPath)
	}
}

// TestResolveStoreFilePath 验证构造函数传入目录时文件名固定。
func TestResolveStoreFilePath(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("/tmp", "boxify-data")
	got, err := resolveStoreFilePath(dir)
	if err != nil {
		t.Fatalf("resolveStoreFilePath() error = %v", err)
	}
	want := filepath.Join(dir, "conversations.json")
	if got != want {
		t.Fatalf("resolveStoreFilePath() = %q, want %q", got, want)
	}
}

// TestNewJSONConversationStoreCreatesFileWhenMissing 验证初始化时不存在的文件会被创建。
func TestNewJSONConversationStoreCreatesFileWhenMissing(t *testing.T) {
	t.Parallel()

	storeDir := filepath.Join(t.TempDir(), "store")
	storeFile := filepath.Join(storeDir, "conversations.json")
	if _, err := os.Stat(storeFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("os.Stat() error = %v, want not exist", err)
	}

	store := NewJSONConversationStore(storeDir, slog.Default())
	if store == nil {
		t.Fatalf("NewJSONConversationStore() returned nil")
	}

	if _, err := os.Stat(storeFile); err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
}
