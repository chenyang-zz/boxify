package chat

import (
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

// ChannelClient 抽象 Boxify 到原生 channel inbox 的投递能力。
type ChannelClient interface {
	SendMessage(ctx context.Context, req BridgeInboxRequest) (*BridgeInboxResponse, error)
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
func (c *HTTPChannelClient) SendMessage(ctx context.Context, req BridgeInboxRequest) (*BridgeInboxResponse, error) {
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
		var result BridgeInboxResponse
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
