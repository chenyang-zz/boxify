package chat

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	clawprocess "github.com/chenyang-zz/boxify/internal/claw/process"
)

// ChannelCoordinator 协调 Boxify 会话、channel client 与流式事件处理。
type ChannelCoordinator struct {
	store         ConversationStore    // 会话与消息存储。
	client        ChannelClient        // Boxify 到 OpenClaw channel 的传输客户端。
	publisher     EventPublisher       // 向前端广播聊天事件。
	streamHandler *StreamEventHandler  // 负责流式事件落库与事件分发。
	manager       *clawprocess.Manager // OpenClaw 进程管理器。
	logger        *slog.Logger         // 模块日志。
}

// NewChannelCoordinator 创建 channel 协调器。
func NewChannelCoordinator(store ConversationStore, client ChannelClient, publisher EventPublisher, manager *clawprocess.Manager, logger *slog.Logger) *ChannelCoordinator {
	if logger == nil {
		logger = slog.Default()
	}
	if store == nil {
		store = NewMemoryConversationStore()
	}
	return &ChannelCoordinator{
		store:         store,
		client:        client,
		publisher:     publisher,
		streamHandler: NewStreamEventHandler(store, publisher, logger),
		manager:       manager,
		logger:        logger,
	}
}

// CreateConversation 创建一个 Boxify 聊天会话。
func (c *ChannelCoordinator) CreateConversation(agentID string) (*Conversation, error) {
	c.logger.Info("开始创建聊天会话", "agent_id", agentID)
	conv, err := c.store.CreateConversation(strings.TrimSpace(agentID))
	if err != nil {
		c.logger.Error("创建聊天会话失败", "agent_id", agentID, "error", err)
		return nil, err
	}
	c.logger.Info("创建聊天会话完成", "conversation_id", conv.ID, "agent_id", conv.AgentID)
	return conv, nil
}

// ListConversations 返回全部聊天会话。
func (c *ChannelCoordinator) ListConversations() ([]Conversation, error) {
	return c.store.ListConversations()
}

// ListMessages 返回指定会话的消息列表。
func (c *ChannelCoordinator) ListMessages(conversationID string) ([]Message, error) {
	return c.store.ListMessages(strings.TrimSpace(conversationID))
}

// SendMessage 将用户消息发送到插件 inbox。
func (c *ChannelCoordinator) SendMessage(ctx context.Context, conversationID, text string) (string, error) {
	conversationID = strings.TrimSpace(conversationID)
	text = strings.TrimSpace(text)
	if conversationID == "" {
		return "", fmt.Errorf("会话 ID 不能为空")
	}
	if text == "" {
		return "", fmt.Errorf("消息内容不能为空")
	}

	conv, err := c.store.GetConversation(conversationID)
	if err != nil {
		return "", err
	}

	if c.manager != nil {
		if err := c.manager.Start(); err != nil {
			c.logger.Warn("OpenClaw 启动失败，继续尝试请求插件 inbox", "error", err)
		}
	}

	envelope := BuildSendMessageEnvelope(SendMessageCommand{
		ConversationID: conversationID,
		AgentID:        conv.AgentID,
		Text:           text,
	})
	if err := c.store.AppendMessage(envelope.UserMessage); err != nil {
		return "", err
	}

	if c.client == nil {
		c.logger.Warn("未配置插件 channel client，当前仅完成本地消息入库", "conversation_id", conversationID, "run_id", envelope.RunID)
		return envelope.RunID, nil
	}

	result, err := c.client.SendMessageStream(ctx, envelope.InboxRequest, func(event ChatStreamEvent) error {
		return c.streamHandler.Handle(conversationID, envelope.RunID, event)
	})
	if err != nil {
		c.logger.Error("请求插件 inbox 失败", "conversation_id", conversationID, "run_id", envelope.RunID, "error", err)
		return "", err
	}
	if result != nil && strings.TrimSpace(result.SessionID) != "" {
		if updateErr := c.store.UpdateOpenClawSessionID(conversationID, result.SessionID); updateErr != nil {
			c.logger.Warn("更新 OpenClaw 会话映射失败", "conversation_id", conversationID, "session_id", result.SessionID, "error", updateErr)
		}
	}

	c.logger.Info("插件 inbox 请求完成", "conversation_id", conversationID, "run_id", envelope.RunID, "agent_id", conv.AgentID)
	return envelope.RunID, nil
}
