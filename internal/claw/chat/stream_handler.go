package chat

import (
	"log/slog"
	"strings"
	"time"
)

// StreamEventHandler 负责把 channel 流式事件翻译为本地状态与前端事件。
type StreamEventHandler struct {
	store     ConversationStore
	publisher EventPublisher
	logger    *slog.Logger
}

// NewStreamEventHandler 创建流式事件处理器。
func NewStreamEventHandler(store ConversationStore, publisher EventPublisher, logger *slog.Logger) *StreamEventHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &StreamEventHandler{
		store:     store,
		publisher: publisher,
		logger:    logger,
	}
}

// Handle 处理插件 SSE 回传的流式聊天事件。
func (h *StreamEventHandler) Handle(conversationID, fallbackRunID string, event ChatStreamEvent) error {
	runID := strings.TrimSpace(event.RunID)
	if runID == "" {
		runID = fallbackRunID
	}

	if strings.TrimSpace(event.SessionID) != "" {
		if err := h.store.UpdateOpenClawSessionID(conversationID, event.SessionID); err != nil {
			h.logger.Warn("流式阶段更新 OpenClaw 会话映射失败", "conversation_id", conversationID, "session_id", event.SessionID, "error", err)
		}
	}

	switch event.EventType {
	case ChatEventTypeStreamDelta:
		if strings.TrimSpace(event.Text) == "" {
			return nil
		}
		if err := h.store.UpdateAssistantDraft(conversationID, runID, event.Text); err != nil {
			h.logger.Error("写入助手流式草稿失败", "conversation_id", conversationID, "run_id", runID, "error", err)
			return err
		}
		h.publish(conversationID, event.SessionID, runID, ChatEventTypeAssistantDelta, map[string]interface{}{
			"text": event.Text,
		})
	case ChatEventTypeStreamDone:
		if err := h.store.FinalizeAssistantMessage(conversationID, runID); err != nil {
			h.logger.Error("收敛助手消息失败", "conversation_id", conversationID, "run_id", runID, "error", err)
			return err
		}
		h.publish(conversationID, event.SessionID, runID, ChatEventTypeAssistantDone, map[string]interface{}{
			"text": event.Text,
		})
	case ChatEventTypeStreamError:
		h.publish(conversationID, event.SessionID, runID, ChatEventTypeAssistantError, map[string]interface{}{
			"error": event.Error,
		})
	}

	return nil
}

// publish 向前端广播统一的聊天事件。
func (h *StreamEventHandler) publish(conversationID, sessionID, runID string, eventType ChatEventType, payload map[string]interface{}) {
	if h.publisher == nil {
		return
	}
	h.publisher.PublishConversationEvent(conversationID, ChatReplyEvent{
		ConversationID: conversationID,
		SessionID:      sessionID,
		RunID:          runID,
		EventType:      eventType,
		Payload:        payload,
		Timestamp:      time.Now(),
	})
}
