package chat

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

var (
	// ErrConversationNotFound 表示会话不存在。
	ErrConversationNotFound = errors.New("会话不存在")
)

// ConversationStore 抽象会话与消息持久化。
type ConversationStore interface {
	CreateConversation(agentID string) (*Conversation, error)
	ListConversations() ([]Conversation, error)
	GetConversation(id string) (*Conversation, error)
	UpdateOpenClawSessionID(conversationID, sessionID string) error
	AppendMessage(msg Message) error
	ListMessages(conversationID string) ([]Message, error)
	UpdateAssistantDraft(conversationID, runID, chunk string) error
	FinalizeAssistantMessage(conversationID, runID string) error
}

// MemoryConversationStore 提供可编译、可替换的内存版原型实现。
type MemoryConversationStore struct {
	mu            sync.RWMutex
	conversations map[string]*Conversation
	messages      map[string][]Message
}

// NewMemoryConversationStore 创建内存会话存储。
func NewMemoryConversationStore() *MemoryConversationStore {
	return &MemoryConversationStore{
		conversations: make(map[string]*Conversation),
		messages:      make(map[string][]Message),
	}
}

// CreateConversation 创建会话。
func (s *MemoryConversationStore) CreateConversation(agentID string) (*Conversation, error) {
	now := time.Now()
	id := fmt.Sprintf("conv_%d", now.UnixNano())
	if agentID == "" {
		agentID = "main"
	}
	conv := &Conversation{
		ID:        id,
		Title:     "新会话",
		AgentID:   agentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.conversations[id] = conv
	return cloneConversation(conv), nil
}

// ListConversations 返回全部会话。
func (s *MemoryConversationStore) ListConversations() ([]Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Conversation, 0, len(s.conversations))
	for _, item := range s.conversations {
		items = append(items, *cloneConversation(item))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items, nil
}

// GetConversation 读取单个会话。
func (s *MemoryConversationStore) GetConversation(id string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.conversations[id]
	if !ok {
		return nil, ErrConversationNotFound
	}
	return cloneConversation(item), nil
}

// UpdateOpenClawSessionID 更新映射的 OpenClaw 会话 ID。
func (s *MemoryConversationStore) UpdateOpenClawSessionID(conversationID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.conversations[conversationID]
	if !ok {
		return ErrConversationNotFound
	}
	item.OpenClawSessionID = sessionID
	item.UpdatedAt = time.Now()
	return nil
}

// AppendMessage 追加消息。
func (s *MemoryConversationStore) AppendMessage(msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.conversations[msg.ConversationID]
	if !ok {
		return ErrConversationNotFound
	}
	s.messages[msg.ConversationID] = append(s.messages[msg.ConversationID], msg)
	item.UpdatedAt = time.Now()
	return nil
}

// ListMessages 返回会话全部消息。
func (s *MemoryConversationStore) ListMessages(conversationID string) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.conversations[conversationID]; !ok {
		return nil, ErrConversationNotFound
	}
	src := s.messages[conversationID]
	dst := make([]Message, len(src))
	copy(dst, src)
	return dst, nil
}

// UpdateAssistantDraft 将 chunk 合并到同一 run 的 assistant 草稿消息。
func (s *MemoryConversationStore) UpdateAssistantDraft(conversationID, runID, chunk string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.conversations[conversationID]
	if !ok {
		return ErrConversationNotFound
	}
	list := s.messages[conversationID]
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].Role == "assistant" && list[i].RunID == runID && list[i].Status == "streaming" {
			list[i].Content += chunk
			list[i].CreatedAt = list[i].CreatedAt
			s.messages[conversationID] = list
			item.UpdatedAt = time.Now()
			return nil
		}
	}

	list = append(list, Message{
		ID:             fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		ConversationID: conversationID,
		RunID:          runID,
		Role:           "assistant",
		Content:        chunk,
		Status:         "streaming",
		CreatedAt:      time.Now(),
	})
	s.messages[conversationID] = list
	item.UpdatedAt = time.Now()
	return nil
}

// FinalizeAssistantMessage 将 run 对应的 assistant 草稿收敛为完成态。
func (s *MemoryConversationStore) FinalizeAssistantMessage(conversationID, runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.conversations[conversationID]
	if !ok {
		return ErrConversationNotFound
	}
	list := s.messages[conversationID]
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].Role == "assistant" && list[i].RunID == runID {
			list[i].Status = "done"
			s.messages[conversationID] = list
			item.UpdatedAt = time.Now()
			return nil
		}
	}
	return nil
}

func cloneConversation(src *Conversation) *Conversation {
	if src == nil {
		return nil
	}
	cp := *src
	return &cp
}
