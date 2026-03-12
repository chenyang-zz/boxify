package chat

import (
	"fmt"
	"time"
)

// SendMessageCommand 描述一次用户发送动作的输入参数。
type SendMessageCommand struct {
	ConversationID string
	AgentID        string
	Text           string
}

// SendMessageEnvelope 描述发送前需要落库与投递的完整上下文。
type SendMessageEnvelope struct {
	RunID        string
	UserMessage  Message
	InboxRequest ChannelInboxRequest
}

// BuildSendMessageEnvelope 构造一次发送所需的运行时载荷。
func BuildSendMessageEnvelope(cmd SendMessageCommand) SendMessageEnvelope {
	now := time.Now()
	runID := fmt.Sprintf("run_%d", now.UnixNano())
	msgID := fmt.Sprintf("msg_%d", now.UnixNano())

	return SendMessageEnvelope{
		RunID: runID,
		UserMessage: Message{
			ID:             msgID,
			ConversationID: cmd.ConversationID,
			RunID:          runID,
			Role:           "user",
			Content:        cmd.Text,
			Status:         "done",
			CreatedAt:      now,
		},
		InboxRequest: ChannelInboxRequest{
			ConversationID: cmd.ConversationID,
			MessageID:      msgID,
			RunID:          runID,
			AgentID:        cmd.AgentID,
			Text:           cmd.Text,
			Metadata: map[string]interface{}{
				"source": "boxify",
			},
		},
	}
}
