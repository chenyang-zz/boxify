package chat

import "time"

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

// BridgeInboxRequest 描述 Boxify 发往插件桥的入站消息。
type BridgeInboxRequest struct {
	ConversationID string                 `json:"conversationId"`     // Boxify 会话 ID。
	MessageID      string                 `json:"messageId"`          // Boxify 消息 ID。
	AgentID        string                 `json:"agentId"`            // 目标 Agent ID。
	Text           string                 `json:"text"`               // 用户输入文本。
	Metadata       map[string]interface{} `json:"metadata,omitempty"` // 附加元数据。
}

// BridgeInboxResponse 描述插件同步返回给 Boxify 的执行结果。
type BridgeInboxResponse struct {
	OK             bool   `json:"ok"`                       // 请求是否成功。
	ConversationID string `json:"conversationId,omitempty"` // Boxify 会话 ID。
	SessionID      string `json:"sessionId,omitempty"`      // OpenClaw 会话 ID。
	Text           string `json:"text,omitempty"`           // 助手最终回复文本。
	Error          string `json:"error,omitempty"`          // 错误信息。
}

// BridgeEvent 描述插件桥回推给 Boxify 的事件。
type BridgeEvent struct {
	ConversationID string                 `json:"conversationId"` // Boxify 会话 ID。
	SessionID      string                 `json:"sessionId"`      // OpenClaw 会话 ID。
	RunID          string                 `json:"runId"`          // OpenClaw 执行 ID。
	EventType      string                 `json:"eventType"`      // 事件类型。
	Payload        map[string]interface{} `json:"payload"`        // 事件负载。
	Timestamp      time.Time              `json:"timestamp"`      // 事件时间。
}

// ChatEvent 描述发往前端的聊天事件载荷。
type ChatEvent struct {
	ConversationID string                 `json:"conversationId"` // Boxify 会话 ID。
	SessionID      string                 `json:"sessionId"`      // OpenClaw 会话 ID。
	RunID          string                 `json:"runId"`          // OpenClaw 执行 ID。
	EventType      string                 `json:"eventType"`      // 事件类型。
	Payload        map[string]interface{} `json:"payload"`        // 事件负载。
	Timestamp      int64                  `json:"timestamp"`      // 事件时间（Unix 毫秒）。
}
