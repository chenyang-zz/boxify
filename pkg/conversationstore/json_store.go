package conversationstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	clawchat "github.com/chenyang-zz/boxify/internal/claw/chat"
)

const defaultStoreFileName = "conversations.json"

// JSONConversationStore 提供基于本地 JSON 文件的会话持久化实现。
type JSONConversationStore struct {
	mu            sync.RWMutex
	path          string
	logger        *slog.Logger
	conversations map[string]*clawchat.Conversation
	messages      map[string][]clawchat.Message
}

var _ clawchat.ConversationStore = (*JSONConversationStore)(nil)

// persistedState 描述落盘的完整状态。
type persistedState struct {
	Conversations map[string]*clawchat.Conversation `json:"conversations"` // 全部会话。
	Messages      map[string][]clawchat.Message     `json:"messages"`      // 按会话分组的消息。
}

// NewJSONConversationStore 创建基于本地 JSON 文件的会话存储。
func NewJSONConversationStore(path string, logger *slog.Logger) clawchat.ConversationStore {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("module", "conversationstore")
	storePath, err := resolveStoreFilePath(path)
	if err != nil {
		logger.Error("解析会话存储路径失败", "error", err, "dir", path)
		return nil
	}
	if path == "" {
		logger.Info("会话存储目录为空，使用默认路径", "path", storePath)
	} else {
		logger.Info("使用指定目录的会话存储路径", "dir", path, "path", storePath)
	}

	store := &JSONConversationStore{
		path:          storePath,
		logger:        logger,
		conversations: make(map[string]*clawchat.Conversation),
		messages:      make(map[string][]clawchat.Message),
	}
	if err := store.load(); err != nil {
		logger.Error("加载会话存储失败，使用空状态", "error", err)
		return nil
	}
	return store
}

// resolveStoreFilePath 将目录参数解析为固定文件路径。
func resolveStoreFilePath(dir string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return resolveDefaultStorePath()
	}
	return filepath.Join(dir, defaultStoreFileName), nil
}

// resolveDefaultStorePath 返回默认会话存储文件路径。
func resolveDefaultStorePath() (string, error) {
	if customDir := strings.TrimSpace(os.Getenv("BOXIFY_CONVERSATION_STORE_DIR")); customDir != "" {
		return filepath.Join(customDir, defaultStoreFileName), nil
	}
	dataDir, err := resolveBoxifyDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, defaultStoreFileName), nil
}

// resolveBoxifyDataDir 返回 Boxify 数据目录。
func resolveBoxifyDataDir() (string, error) {
	if dataDir := strings.TrimSpace(os.Getenv("BOXIFY_DATA_DIR")); dataDir != "" {
		return dataDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}
	return filepath.Join(home, ".boxify"), nil
}

// CreateConversation 创建会话。
func (s *JSONConversationStore) CreateConversation(agentID string) (*clawchat.Conversation, error) {
	now := time.Now()
	id := fmt.Sprintf("conv_%d", now.UnixNano())
	if agentID == "" {
		agentID = "main"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.cloneStateLocked()
	next.Conversations[id] = &clawchat.Conversation{
		ID:        id,
		Title:     "新会话",
		AgentID:   agentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.commitLocked(next); err != nil {
		return nil, err
	}
	return cloneConversation(next.Conversations[id]), nil
}

// ListConversations 返回全部会话。
func (s *JSONConversationStore) ListConversations() ([]clawchat.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]clawchat.Conversation, 0, len(s.conversations))
	for _, item := range s.conversations {
		items = append(items, *cloneConversation(item))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items, nil
}

// GetConversation 读取单个会话。
func (s *JSONConversationStore) GetConversation(id string) (*clawchat.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.conversations[id]
	if !ok {
		return nil, clawchat.ErrConversationNotFound
	}
	return cloneConversation(item), nil
}

// UpdateOpenClawSessionID 更新映射的 OpenClaw 会话 ID。
func (s *JSONConversationStore) UpdateOpenClawSessionID(conversationID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.cloneStateLocked()
	item, ok := next.Conversations[conversationID]
	if !ok {
		return clawchat.ErrConversationNotFound
	}
	item.OpenClawSessionID = sessionID
	item.UpdatedAt = time.Now()
	return s.commitLocked(next)
}

// AppendMessage 追加消息。
func (s *JSONConversationStore) AppendMessage(msg clawchat.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.cloneStateLocked()
	item, ok := next.Conversations[msg.ConversationID]
	if !ok {
		return clawchat.ErrConversationNotFound
	}
	next.Messages[msg.ConversationID] = append(next.Messages[msg.ConversationID], msg)
	item.UpdatedAt = time.Now()
	return s.commitLocked(next)
}

// ListMessages 返回会话全部消息。
func (s *JSONConversationStore) ListMessages(conversationID string) ([]clawchat.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.conversations[conversationID]; !ok {
		return nil, clawchat.ErrConversationNotFound
	}
	src := s.messages[conversationID]
	dst := make([]clawchat.Message, len(src))
	copy(dst, src)
	return dst, nil
}

// UpdateAssistantDraft 将 chunk 合并到同一 run 的 assistant 草稿消息。
func (s *JSONConversationStore) UpdateAssistantDraft(conversationID, runID, chunk string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.cloneStateLocked()
	item, ok := next.Conversations[conversationID]
	if !ok {
		return clawchat.ErrConversationNotFound
	}

	list := next.Messages[conversationID]
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].Role == "assistant" && list[i].RunID == runID && list[i].Status == "streaming" {
			list[i].Content += chunk
			next.Messages[conversationID] = list
			item.UpdatedAt = time.Now()
			return s.commitLocked(next)
		}
	}

	list = append(list, clawchat.Message{
		ID:             fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		ConversationID: conversationID,
		RunID:          runID,
		Role:           "assistant",
		Content:        chunk,
		Status:         "streaming",
		CreatedAt:      time.Now(),
	})
	next.Messages[conversationID] = list
	item.UpdatedAt = time.Now()
	return s.commitLocked(next)
}

// FinalizeAssistantMessage 将 run 对应的 assistant 草稿收敛为完成态。
func (s *JSONConversationStore) FinalizeAssistantMessage(conversationID, runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.cloneStateLocked()
	item, ok := next.Conversations[conversationID]
	if !ok {
		return clawchat.ErrConversationNotFound
	}

	list := next.Messages[conversationID]
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].Role == "assistant" && list[i].RunID == runID {
			list[i].Status = "done"
			next.Messages[conversationID] = list
			item.UpdatedAt = time.Now()
			return s.commitLocked(next)
		}
	}
	return nil
}

// load 从本地文件恢复状态。
func (s *JSONConversationStore) load() error {
	s.logger.Info("开始加载会话存储", "path", s.path)

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		s.logger.Error("创建会话存储目录失败", "path", s.path, "error", err)
		return fmt.Errorf("创建会话存储目录失败: %w", err)
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.logger.Info("会话存储文件不存在，创建空文件", "path", s.path)
			if err := s.persistState(persistedState{
				Conversations: make(map[string]*clawchat.Conversation),
				Messages:      make(map[string][]clawchat.Message),
			}); err != nil {
				return err
			}
			s.logger.Info("完成创建空会话存储文件", "path", s.path)
			return nil
		}
		s.logger.Error("读取会话存储文件失败", "path", s.path, "error", err)
		return fmt.Errorf("读取会话存储文件失败: %w", err)
	}
	if len(data) == 0 {
		s.logger.Warn("会话存储文件为空，使用空状态", "path", s.path)
		return nil
	}

	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		s.logger.Error("解析会话存储文件失败", "path", s.path, "error", err)
		return fmt.Errorf("解析会话存储文件失败: %w", err)
	}

	if state.Conversations == nil {
		state.Conversations = make(map[string]*clawchat.Conversation)
	}
	if state.Messages == nil {
		state.Messages = make(map[string][]clawchat.Message)
	}

	s.conversations = cloneConversationMap(state.Conversations)
	s.messages = cloneMessageMap(state.Messages)
	s.logger.Info("完成加载会话存储", "path", s.path, "conversation_count", len(s.conversations))
	return nil
}

// commitLocked 将新状态写入内存并持久化到磁盘。
func (s *JSONConversationStore) commitLocked(next persistedState) error {
	if err := s.persistState(next); err != nil {
		return err
	}
	s.conversations = next.Conversations
	s.messages = next.Messages
	return nil
}

// cloneStateLocked 复制当前内存状态，供写路径修改。
func (s *JSONConversationStore) cloneStateLocked() persistedState {
	return persistedState{
		Conversations: cloneConversationMap(s.conversations),
		Messages:      cloneMessageMap(s.messages),
	}
}

// persistState 将状态安全写入本地 JSON 文件。
func (s *JSONConversationStore) persistState(state persistedState) error {
	s.logger.Debug("开始持久化会话存储", "path", s.path, "conversation_count", len(state.Conversations))

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		s.logger.Error("序列化会话存储失败", "path", s.path, "error", err)
		return fmt.Errorf("序列化会话存储失败: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		s.logger.Error("写入临时会话存储文件失败", "path", tmpPath, "error", err)
		return fmt.Errorf("写入临时会话存储文件失败: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		s.logger.Error("替换会话存储文件失败", "path", s.path, "error", err)
		return fmt.Errorf("替换会话存储文件失败: %w", err)
	}

	s.logger.Debug("完成持久化会话存储", "path", s.path)
	return nil
}

// cloneConversationMap 深拷贝会话 map，避免读写共享底层状态。
func cloneConversationMap(src map[string]*clawchat.Conversation) map[string]*clawchat.Conversation {
	dst := make(map[string]*clawchat.Conversation, len(src))
	for id, item := range src {
		dst[id] = cloneConversation(item)
	}
	return dst
}

// cloneMessageMap 深拷贝消息 map，避免读写共享底层切片。
func cloneMessageMap(src map[string][]clawchat.Message) map[string][]clawchat.Message {
	dst := make(map[string][]clawchat.Message, len(src))
	for id, list := range src {
		cp := make([]clawchat.Message, len(list))
		copy(cp, list)
		dst[id] = cp
	}
	return dst
}

// cloneConversation 复制单个会话对象。
func cloneConversation(src *clawchat.Conversation) *clawchat.Conversation {
	if src == nil {
		return nil
	}
	cp := *src
	return &cp
}
