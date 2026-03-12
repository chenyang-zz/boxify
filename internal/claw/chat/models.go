package chat

import "time"

// ChatEventType 描述聊天链路的事件类型。
type ChatEventType string

const (
	// ChatEventTypeAssistantDelta 表示 Boxify 前端应将该文本片段追加到助手草稿中。
	ChatEventTypeAssistantDelta ChatEventType = "assistant_delta"
	// ChatEventTypeAssistantDone 表示本次助手回复已完成，前端可结束流式态并刷新正式消息。
	ChatEventTypeAssistantDone ChatEventType = "assistant_done"
	// ChatEventTypeAssistantError 表示本次助手回复失败，前端应展示错误状态。
	ChatEventTypeAssistantError ChatEventType = "assistant_error"
	// ChatEventTypeStreamStart 表示插件 SSE 流已建立，OpenClaw 开始生成回复。
	ChatEventTypeStreamStart ChatEventType = "start"
	// ChatEventTypeStreamDelta 表示插件 SSE 返回了一段新的文本增量。
	ChatEventTypeStreamDelta ChatEventType = "delta"
	// ChatEventTypeStreamDone 表示插件 SSE 已完成本次回复输出。
	ChatEventTypeStreamDone ChatEventType = "done"
	// ChatEventTypeStreamError 表示插件 SSE 在回复过程中发生错误。
	ChatEventTypeStreamError ChatEventType = "error"
)

// Conversation 描述 Boxify 侧维护的聊天会话。
type Conversation struct {
	ID                string    `json:"id"`                          // Boxify 会话 ID。
	Title             string    `json:"title"`                       // 会话标题。
	AgentID           string    `json:"agentId"`                     // 绑定的 Agent ID。
	OpenClawSessionID string    `json:"openClawSessionId,omitempty"` // 映射到 OpenClaw 的会话 ID。
	CreatedAt         time.Time `json:"createdAt"`                   // 创建时间。
	UpdatedAt         time.Time `json:"updatedAt"`                   // 最近更新时间。
}

// Message 描述会话中的单条消息。
type Message struct {
	ID             string    `json:"id"`               // Boxify 消息 ID。
	ConversationID string    `json:"conversationId"`   // 所属会话 ID。
	RunID          string    `json:"runId,omitempty"`  // 所属执行 ID。
	Role           string    `json:"role"`             // user/assistant/system。
	Content        string    `json:"content"`          // 消息文本。
	Status         string    `json:"status,omitempty"` // streaming/done/error。
	CreatedAt      time.Time `json:"createdAt"`        // 创建时间。
}

// ChannelInboxRequest 描述 Boxify 发往原生 channel inbox 的入站消息。
type ChannelInboxRequest struct {
	ConversationID string                 `json:"conversationId"`     // Boxify 会话 ID。
	MessageID      string                 `json:"messageId"`          // Boxify 消息 ID。
	RunID          string                 `json:"runId,omitempty"`    // 本次执行 ID。
	AgentID        string                 `json:"agentId"`            // 目标 Agent ID。
	Text           string                 `json:"text"`               // 用户输入文本。
	Metadata       map[string]interface{} `json:"metadata,omitempty"` // 附加元数据。
}

// ChannelInboxResponse 描述插件同步返回给 Boxify 的执行结果。
type ChannelInboxResponse struct {
	OK             bool   `json:"ok"`                       // 请求是否成功。
	ConversationID string `json:"conversationId,omitempty"` // Boxify 会话 ID。
	SessionID      string `json:"sessionId,omitempty"`      // OpenClaw 会话 ID。
	Text           string `json:"text,omitempty"`           // 助手最终回复文本。
	Error          string `json:"error,omitempty"`          // 错误信息。
}

// ChatReplyEvent 描述聊天链路回推给 Boxify 的事件。
type ChatReplyEvent struct {
	ConversationID string                 `json:"conversationId"` // Boxify 会话 ID。
	SessionID      string                 `json:"sessionId"`      // OpenClaw 会话 ID。
	RunID          string                 `json:"runId"`          // OpenClaw 执行 ID。
	EventType      ChatEventType          `json:"eventType"`      // 事件类型。
	Payload        map[string]interface{} `json:"payload"`        // 事件负载。
	Timestamp      time.Time              `json:"timestamp"`      // 事件时间。
}

// ChatStreamEvent 描述插件 SSE 返回的流式事件。
type ChatStreamEvent struct {
	EventType      ChatEventType          `json:"eventType"`                // start/delta/done/error。
	ConversationID string                 `json:"conversationId,omitempty"` // Boxify 会话 ID。
	SessionID      string                 `json:"sessionId,omitempty"`      // OpenClaw 会话 ID。
	RunID          string                 `json:"runId,omitempty"`          // 本次执行 ID。
	Text           string                 `json:"text,omitempty"`           // 文本增量或最终文本。
	Error          string                 `json:"error,omitempty"`          // 错误信息。
	Payload        map[string]interface{} `json:"payload,omitempty"`        // 扩展负载。
}

// ChatEvent 描述发往前端的聊天事件载荷。
type ChatEvent struct {
	ConversationID string                 `json:"conversationId"` // Boxify 会话 ID。
	SessionID      string                 `json:"sessionId"`      // OpenClaw 会话 ID。
	RunID          string                 `json:"runId"`          // OpenClaw 执行 ID。
	EventType      ChatEventType          `json:"eventType"`      // 事件类型。
	Payload        map[string]interface{} `json:"payload"`        // 事件负载。
	Timestamp      int64                  `json:"timestamp"`      // 事件时间（Unix 毫秒）。
}
