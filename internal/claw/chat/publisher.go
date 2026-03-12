package chat

import (
	"log/slog"

	"github.com/chenyang-zz/boxify/internal/events"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// EventPublisher 抽象对前端的事件发布。
type EventPublisher interface {
	PublishConversationEvent(conversationID string, event BridgeEvent)
}

// WailsEventPublisher 负责将聊天事件广播给前端。
type WailsEventPublisher struct {
	app    *application.App
	logger *slog.Logger
}

// NewWailsEventPublisher 创建 Wails 事件发布器。
func NewWailsEventPublisher(app *application.App, logger *slog.Logger) *WailsEventPublisher {
	if logger == nil {
		logger = slog.Default()
	}
	return &WailsEventPublisher{app: app, logger: logger}
}

// PublishConversationEvent 广播聊天事件。
func (p *WailsEventPublisher) PublishConversationEvent(conversationID string, event BridgeEvent) {
	if p == nil || p.app == nil {
		return
	}
	p.app.Event.Emit(string(events.EventTypeClawChatEvent), ChatEvent{
		ConversationID: conversationID,
		SessionID:      event.SessionID,
		RunID:          event.RunID,
		EventType:      event.EventType,
		Payload:        event.Payload,
		Timestamp:      event.Timestamp.UnixMilli(),
	})
}
