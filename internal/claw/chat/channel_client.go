package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrChannelRequestRejected 表示插件端拒绝本次请求。
	ErrChannelRequestRejected = errors.New("插件端拒绝请求")
)

// BoxifyInboxPath 是原生 channel inbox 的固定 HTTP 路径。
const BoxifyInboxPath = "/channels/boxify/inbox"

// BoxifyInboxStreamPath 是原生 channel inbox 的 SSE 流式路径。
const BoxifyInboxStreamPath = "/channels/boxify/inbox/stream"

// ChannelClient 抽象 Boxify 到原生 channel inbox 的投递能力。
type ChannelClient interface {
	SendMessage(ctx context.Context, req ChannelInboxRequest) (*ChannelInboxResponse, error)
	SendMessageStream(ctx context.Context, req ChannelInboxRequest, onEvent func(ChatStreamEvent) error) (*ChannelInboxResponse, error)
}

// HTTPChannelClient 通过本地 HTTP 将消息发送给插件 inbox。
type HTTPChannelClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewHTTPChannelClient 创建 HTTP channel 客户端。
func NewHTTPChannelClient(baseURL, token string) *HTTPChannelClient {
	baseURL = strings.TrimSpace(strings.TrimRight(baseURL, "/"))
	if baseURL == "" {
		baseURL = "http://127.0.0.1:32124"
	}
	return &HTTPChannelClient{
		baseURL: baseURL,
		token:   strings.TrimSpace(token),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SendMessage 将用户消息发送到插件 inbox。
func (c *HTTPChannelClient) SendMessage(ctx context.Context, req ChannelInboxRequest) (*ChannelInboxResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+BoxifyInboxPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("X-Boxify-Token", c.token)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求插件 inbox 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		var result ChannelInboxResponse
		if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
			return nil, fmt.Errorf("解析插件响应失败: %w", err)
		}
		if !result.OK {
			if strings.TrimSpace(result.Error) == "" {
				return &result, ErrChannelRequestRejected
			}
			return &result, fmt.Errorf("%w: %s", ErrChannelRequestRejected, strings.TrimSpace(result.Error))
		}
		return &result, nil
	}

	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if len(payload) == 0 {
		return nil, ErrChannelRequestRejected
	}
	return nil, fmt.Errorf("%w: %s", ErrChannelRequestRejected, strings.TrimSpace(string(payload)))
}

// SendMessageStream 通过 SSE 消费插件返回的增量事件。
func (c *HTTPChannelClient) SendMessageStream(ctx context.Context, req ChannelInboxRequest, onEvent func(ChatStreamEvent) error) (*ChannelInboxResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化流式请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+BoxifyInboxStreamPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建流式请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if c.token != "" {
		httpReq.Header.Set("X-Boxify-Token", c.token)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求插件流式 inbox 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if len(payload) == 0 {
			return nil, ErrChannelRequestRejected
		}
		return nil, fmt.Errorf("%w: %s", ErrChannelRequestRejected, strings.TrimSpace(string(payload)))
	}

	reader := bufio.NewReader(io.LimitReader(resp.Body, 8<<20))
	result := &ChannelInboxResponse{OK: true}
	var eventName string
	dataLines := make([]string, 0, 4)

	flushEvent := func() error {
		if len(dataLines) == 0 {
			eventName = ""
			return nil
		}

		var event ChatStreamEvent
		if err := json.Unmarshal([]byte(strings.Join(dataLines, "\n")), &event); err != nil {
			return fmt.Errorf("解析 SSE 事件失败: %w", err)
		}
		dataLines = dataLines[:0]
		if strings.TrimSpace(string(event.EventType)) == "" {
			event.EventType = ChatEventType(eventName)
		}
		eventName = ""

		if strings.TrimSpace(event.ConversationID) != "" {
			result.ConversationID = event.ConversationID
		}
		if strings.TrimSpace(event.SessionID) != "" {
			result.SessionID = event.SessionID
		}
		switch event.EventType {
		case ChatEventTypeStreamDone:
			result.Text = event.Text
		case ChatEventTypeStreamError:
			result.OK = false
			result.Error = strings.TrimSpace(event.Error)
		}

		if onEvent != nil {
			if err := onEvent(event); err != nil {
				return err
			}
		}
		return nil
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("读取 SSE 响应失败: %w", err)
		}

		line = strings.TrimRight(line, "\r\n")
		switch {
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		case line == "":
			if err := flushEvent(); err != nil {
				return nil, err
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	if err := flushEvent(); err != nil {
		return nil, err
	}
	if !result.OK {
		if result.Error == "" {
			return result, ErrChannelRequestRejected
		}
		return result, fmt.Errorf("%w: %s", ErrChannelRequestRejected, result.Error)
	}
	return result, nil
}
