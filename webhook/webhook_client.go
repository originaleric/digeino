package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/originaleric/digeino/config"
)

// WebhookClient Webhook 客户端
type WebhookClient struct {
	config *config.WebhookConfig
	client *http.Client
}

var (
	// 共享 Transport 以复用连接池
	defaultTransport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100, // 显著提高单主机并发连接数，特别适用于本地 Webhook
		IdleConnTimeout:     90 * time.Second,
	}
)

// NewWebhookClient 创建 Webhook 客户端
func NewWebhookClient(config *config.WebhookConfig) *WebhookClient {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &WebhookClient{
		config: config,
		client: &http.Client{
			Timeout:   timeout,
			Transport: defaultTransport,
		},
	}
}

// SendStatus 发送状态更新
func (c *WebhookClient) SendStatus(ctx context.Context, status ExecutionStatus) error {
	// 检查事件是否被订阅
	if len(c.config.Events) > 0 {
		subscribed := false
		for _, e := range c.config.Events {
			if e == status.Type {
				subscribed = true
				break
			}
		}
		if !subscribed {
			return nil
		}
	}

	payload := WebhookPayload{
		Event:  status.Type,
		Status: status,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload failed: %w", err)
	}

	// 如果配置了密钥，生成签名
	if c.config.Secret != "" {
		payload.Signature = c.generateSignature(data, c.config.Secret)
		// 重新序列化包含签名的 payload
		data, _ = json.Marshal(payload)
	}

	var lastErr error
	retryCount := c.config.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}
	retryDelay := time.Duration(c.config.RetryDelay) * time.Millisecond
	if retryDelay == 0 {
		retryDelay = 1000 * time.Millisecond
	}

	for i := 0; i <= retryCount; i++ {
		if i > 0 {
			time.Sleep(retryDelay)
		}

		err = c.doPost(ctx, data)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("send webhook failed after %d retries: %w", retryCount, lastErr)
}

func (c *WebhookClient) doPost(ctx context.Context, data []byte) error {
	method := c.config.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequestWithContext(ctx, method, c.config.URL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("http status error: %d", resp.StatusCode)
}

func (c *WebhookClient) generateSignature(data []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
