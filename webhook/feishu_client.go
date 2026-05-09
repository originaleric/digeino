package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/originaleric/digeino/config"
)

type FeishuClient struct {
	cfg    config.FeishuAPIConfig
	client *http.Client

	mu          sync.Mutex
	token       string
	tokenExpire time.Time
}

func NewFeishuClient(cfg config.FeishuAPIConfig) *FeishuClient {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "https://open.feishu.cn"
	}
	cfg.BaseURL = strings.TrimRight(baseURL, "/")
	if cfg.RetryCount < 0 {
		cfg.RetryCount = 0
	}
	if cfg.RetryDelayMs <= 0 {
		cfg.RetryDelayMs = 500
	}
	if cfg.ReceiveIDType == "" {
		cfg.ReceiveIDType = "chat_id"
	}
	return &FeishuClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *FeishuClient) SendText(ctx context.Context, receiveIDType string, receiveIDs []string, content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is required")
	}
	if receiveIDType == "" {
		receiveIDType = c.cfg.ReceiveIDType
	}
	if receiveIDType == "" {
		receiveIDType = "chat_id"
	}
	if len(receiveIDs) == 0 {
		receiveIDs = c.cfg.ReceiveIDs
	}
	if len(receiveIDs) == 0 {
		return fmt.Errorf("receive_ids is empty and no default receive_ids configured")
	}

	var lastErr error
	for i := 0; i <= c.cfg.RetryCount; i++ {
		if i > 0 {
			time.Sleep(time.Duration(c.cfg.RetryDelayMs) * time.Millisecond)
		}
		if err := c.sendTextOnce(ctx, receiveIDType, receiveIDs, content); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return fmt.Errorf("send feishu message failed after %d retries: %w", c.cfg.RetryCount, lastErr)
}

func (c *FeishuClient) sendTextOnce(ctx context.Context, receiveIDType string, receiveIDs []string, content string) error {
	token, err := c.getTenantAccessToken(ctx)
	if err != nil {
		return err
	}
	for _, id := range receiveIDs {
		if strings.TrimSpace(id) == "" {
			continue
		}
		if err := c.sendTextToOne(ctx, token, receiveIDType, id, content); err != nil {
			return err
		}
	}
	return nil
}

func (c *FeishuClient) sendTextToOne(ctx context.Context, token, receiveIDType, receiveID, content string) error {
	endpoint := fmt.Sprintf("%s/open-apis/im/v1/messages", c.cfg.BaseURL)
	q := url.Values{}
	q.Set("receive_id_type", receiveIDType)
	reqURL := endpoint + "?" + q.Encode()

	payload := map[string]string{
		"receive_id": receiveID,
		"msg_type":   "text",
		"content":    fmt.Sprintf("{\"text\":%q}", content),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal message payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var out struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	_ = json.Unmarshal(body, &out)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && out.Code == 0 {
		return nil
	}
	return fmt.Errorf("feishu send failed status=%d code=%d msg=%s", resp.StatusCode, out.Code, out.Msg)
}

func (c *FeishuClient) getTenantAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.token != "" && time.Now().Before(c.tokenExpire) {
		token := c.token
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	if strings.TrimSpace(c.cfg.AppID) == "" || strings.TrimSpace(c.cfg.AppSecret) == "" {
		return "", fmt.Errorf("feishu app_id/app_secret not configured")
	}

	reqBody, _ := json.Marshal(map[string]string{
		"app_id":     c.cfg.AppID,
		"app_secret": c.cfg.AppSecret,
	})
	reqURL := fmt.Sprintf("%s/open-apis/auth/v3/tenant_access_token/internal", c.cfg.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var out struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("parse token response failed: %w", err)
	}
	if out.Code != 0 || out.TenantAccessToken == "" {
		return "", fmt.Errorf("get tenant_access_token failed code=%d msg=%s", out.Code, out.Msg)
	}

	expireSec := out.Expire
	if expireSec <= 300 {
		expireSec = 7200
	}
	c.mu.Lock()
	c.token = out.TenantAccessToken
	c.tokenExpire = time.Now().Add(time.Duration(expireSec-300) * time.Second)
	token := c.token
	c.mu.Unlock()
	return token, nil
}

