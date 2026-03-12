package chat

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	clawprocess "github.com/chenyang-zz/boxify/internal/claw/process"
)

// Service 封装 Boxify 与 OpenClaw 原生 channel 之间的交互。
type Service struct {
	store     ConversationStore
	client    ChannelClient
	publisher EventPublisher
	manager   *clawprocess.Manager
	logger    *slog.Logger
}

// NewService 创建聊天服务。
func NewService(store ConversationStore, client ChannelClient, publisher EventPublisher, manager *clawprocess.Manager, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if store == nil {
		store = NewMemoryConversationStore()
	}
	return &Service{
		store:     store,
		client:    client,
		publisher: publisher,
		manager:   manager,
		logger:    logger,
	}
}

// CreateConversation 创建一个 Boxify 聊天会话。
func (s *Service) CreateConversation(agentID string) (*Conversation, error) {
	s.logger.Info("开始创建聊天会话", "agent_id", agentID)
	conv, err := s.store.CreateConversation(strings.TrimSpace(agentID))
	if err != nil {
		s.logger.Error("创建聊天会话失败", "agent_id", agentID, "error", err)
		return nil, err
	}
	s.logger.Info("创建聊天会话完成", "conversation_id", conv.ID, "agent_id", conv.AgentID)
	return conv, nil
}

// ListConversations 返回全部聊天会话。
func (s *Service) ListConversations() ([]Conversation, error) {
	return s.store.ListConversations()
}

// ListMessages 返回指定会话的消息列表。
func (s *Service) ListMessages(conversationID string) ([]Message, error) {
	return s.store.ListMessages(strings.TrimSpace(conversationID))
}

// SendMessage 将用户消息发送到插件 inbox。
func (s *Service) SendMessage(ctx context.Context, conversationID, text string) (string, error) {
	conversationID = strings.TrimSpace(conversationID)
	text = strings.TrimSpace(text)
	if conversationID == "" {
		return "", fmt.Errorf("会话 ID 不能为空")
	}
	if text == "" {
		return "", fmt.Errorf("消息内容不能为空")
	}

	conv, err := s.store.GetConversation(conversationID)
	if err != nil {
		return "", err
	}

	if s.manager != nil {
		if err := s.manager.Start(); err != nil {
			s.logger.Warn("OpenClaw 启动失败，继续尝试请求插件 inbox", "error", err)
		}
	}

	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())
	msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	if err := s.store.AppendMessage(Message{
		ID:             msgID,
		ConversationID: conversationID,
		RunID:          runID,
		Role:           "user",
		Content:        text,
		Status:         "done",
		CreatedAt:      time.Now(),
	}); err != nil {
		return "", err
	}

	if s.client == nil {
		s.logger.Warn("未配置插件 channel client，当前仅完成本地消息入库", "conversation_id", conversationID, "run_id", runID)
		return runID, nil
	}

	result, err := s.client.SendMessageStream(ctx, ChannelInboxRequest{
		ConversationID: conversationID,
		MessageID:      msgID,
		RunID:          runID,
		AgentID:        conv.AgentID,
		Text:           text,
		Metadata: map[string]interface{}{
			"source": "boxify",
		},
	}, func(event ChatStreamEvent) error {
		return s.handleStreamEvent(conversationID, runID, event)
	})
	if err != nil {
		s.logger.Error("请求插件 inbox 失败", "conversation_id", conversationID, "run_id", runID, "error", err)
		return "", err
	}
	if result != nil && strings.TrimSpace(result.SessionID) != "" {
		if updateErr := s.store.UpdateOpenClawSessionID(conversationID, result.SessionID); updateErr != nil {
			s.logger.Warn("更新 OpenClaw 会话映射失败", "conversation_id", conversationID, "session_id", result.SessionID, "error", updateErr)
		}
	}

	s.logger.Info("插件 inbox 请求完成", "conversation_id", conversationID, "run_id", runID, "agent_id", conv.AgentID)
	return runID, nil
}

// handleStreamEvent 处理插件 SSE 回传的流式聊天事件。
func (s *Service) handleStreamEvent(conversationID, fallbackRunID string, event ChatStreamEvent) error {
	runID := strings.TrimSpace(event.RunID)
	if runID == "" {
		runID = fallbackRunID
	}

	if strings.TrimSpace(event.SessionID) != "" {
		if err := s.store.UpdateOpenClawSessionID(conversationID, event.SessionID); err != nil {
			s.logger.Warn("流式阶段更新 OpenClaw 会话映射失败", "conversation_id", conversationID, "session_id", event.SessionID, "error", err)
		}
	}

	switch event.EventType {
	case ChatEventTypeStreamDelta:
		if strings.TrimSpace(event.Text) == "" {
			return nil
		}
		if err := s.store.UpdateAssistantDraft(conversationID, runID, event.Text); err != nil {
			s.logger.Error("写入助手流式草稿失败", "conversation_id", conversationID, "run_id", runID, "error", err)
			return err
		}
		s.publishEvent(conversationID, event.SessionID, runID, ChatEventTypeAssistantDelta, map[string]interface{}{
			"text": event.Text,
		})
	case ChatEventTypeStreamDone:
		if err := s.store.FinalizeAssistantMessage(conversationID, runID); err != nil {
			s.logger.Error("收敛助手消息失败", "conversation_id", conversationID, "run_id", runID, "error", err)
			return err
		}
		s.publishEvent(conversationID, event.SessionID, runID, ChatEventTypeAssistantDone, map[string]interface{}{
			"text": event.Text,
		})
	case ChatEventTypeStreamError:
		s.publishEvent(conversationID, event.SessionID, runID, ChatEventTypeAssistantError, map[string]interface{}{
			"error": event.Error,
		})
	}

	return nil
}

// publishEvent 向前端广播统一的聊天事件。
func (s *Service) publishEvent(conversationID, sessionID, runID string, eventType ChatEventType, payload map[string]interface{}) {
	if s.publisher == nil {
		return
	}
	s.publisher.PublishConversationEvent(conversationID, ChatReplyEvent{
		ConversationID: conversationID,
		SessionID:      sessionID,
		RunID:          runID,
		EventType:      eventType,
		Payload:        payload,
		Timestamp:      time.Now(),
	})
}
