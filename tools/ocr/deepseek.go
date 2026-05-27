package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/originaleric/digeino/config"
)

const (
	defaultDeepSeekBaseURL = "https://api.deepseek.com"
	defaultDeepSeekModel   = "deepseek-ocr"
)

type deepSeekProvider struct {
	cfg    config.DeepSeekOCRConfig
	ocrCfg config.OCRConfig
	client *http.Client
}

func newDeepSeekProvider(ds config.DeepSeekOCRConfig, ocrCfg config.OCRConfig) (*deepSeekProvider, error) {
	if deepSeekAPIKey(ds) == "" {
		return nil, newOCRError(CodeConfigMissing, "DeepSeek OCR ApiKey not configured (Tools.OCR.DeepSeek.ApiKey or DEEPSEEK_OCR_API_KEY)")
	}
	timeout := time.Duration(ocrCfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &deepSeekProvider{
		cfg:    ds,
		ocrCfg: ocrCfg,
		client: &http.Client{Timeout: timeout},
	}, nil
}

func (p *deepSeekProvider) Name() string {
	return "deepseek-ocr"
}

func (p *deepSeekProvider) Recognize(ctx context.Context, req *OCRRequest, img *resolvedImage) (*OCRResponse, error) {
	mode := strings.ToLower(strings.TrimSpace(p.cfg.Mode))
	if mode == "" {
		mode = "chat"
	}
	var resp *OCRResponse
	var err error
	switch mode {
	case "ocr", "ocr_endpoint":
		resp, err = p.recognizeOCREndpoint(ctx, req, img)
	default:
		resp, err = p.recognizeChat(ctx, req, img)
	}
	if err != nil {
		return nil, err
	}
	if resp.Provider == "" {
		resp.Provider = p.Name()
	}
	if resp.Model == "" {
		resp.Model = p.model()
	}
	return resp, nil
}

func (p *deepSeekProvider) model() string {
	if m := strings.TrimSpace(p.cfg.Model); m != "" {
		return m
	}
	return defaultDeepSeekModel
}

func (p *deepSeekProvider) baseURL() string {
	if u := strings.TrimSpace(p.cfg.BaseUrl); u != "" {
		return strings.TrimRight(u, "/")
	}
	return defaultDeepSeekBaseURL
}

func (p *deepSeekProvider) recognizeChat(ctx context.Context, req *OCRRequest, img *resolvedImage) (*OCRResponse, error) {
	body := chatCompletionRequest{
		Model: p.model(),
		Messages: []chatMessage{
			{
				Role: "user",
				Content: []contentPart{
					{Type: "text", Text: buildPrompt(req)},
					{Type: "image_url", ImageURL: &imageURLPart{URL: img.DataURL}},
				},
			},
		},
		Stream: false,
	}
	raw, usage, err := p.doWithRetry(ctx, p.baseURL()+"/v1/chat/completions", body)
	if err != nil {
		return nil, err
	}
	text, blocks, tables, conf := parseModelOutput(raw, req)
	return &OCRResponse{
		Text:       text,
		Blocks:     blocks,
		Tables:     tables,
		Confidence: conf,
		Provider:   p.Name(),
		Model:      p.model(),
		Usage:      usage,
	}, nil
}

func (p *deepSeekProvider) recognizeOCREndpoint(ctx context.Context, req *OCRRequest, img *resolvedImage) (*OCRResponse, error) {
	payload := map[string]any{
		"image":      img.DataURL,
		"image_type": "base64",
		"prompt":     buildPrompt(req),
	}
	if len(req.Languages) > 0 {
		payload["language"] = strings.Join(req.Languages, ",")
	}
	if req.Task != "" {
		payload["task"] = req.Task
	}
	endpoint := strings.TrimSpace(p.cfg.OCREndpoint)
	if endpoint == "" {
		endpoint = "/v1/ocr"
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	raw, _, err := p.doWithRetry(ctx, p.baseURL()+endpoint, payload)
	if err != nil {
		return nil, err
	}
	text, blocks, tables, conf := parseModelOutput(raw, req)
	return &OCRResponse{
		Text:       text,
		Blocks:     blocks,
		Tables:     tables,
		Confidence: conf,
		Provider:   p.Name(),
		Model:      p.model(),
	}, nil
}

func (p *deepSeekProvider) doWithRetry(ctx context.Context, url string, body any) (content string, usage *OCRUsage, err error) {
	attempts := p.ocrCfg.RetryCount + 1
	if attempts < 1 {
		attempts = 1
	}
	backoff := time.Duration(p.ocrCfg.RetryDelayMs) * time.Millisecond
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return "", nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
		content, usage, lastErr = p.postJSON(ctx, url, body)
		if lastErr == nil {
			return content, usage, nil
		}
		if !isRetryable(lastErr) {
			return "", nil, lastErr
		}
	}
	return "", nil, lastErr
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if oe, ok := asOCRError(err); ok {
		return oe.Code == CodeProviderTimeout || oe.Code == CodeProviderError
	}
	return false
}

func (p *deepSeekProvider) postJSON(ctx context.Context, url string, body any) (string, *OCRUsage, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return "", nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+deepSeekAPIKey(p.cfg))

	resp, err := p.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", nil, newOCRError(CodeProviderTimeout, err.Error())
		}
		return "", nil, newOCRError(CodeProviderError, err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	if err != nil {
		return "", nil, newOCRError(CodeProviderError, err.Error())
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, newOCRError(CodeProviderError, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 512)))
	}

	// chat completions
	var chatResp chatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err == nil && len(chatResp.Choices) > 0 {
		text := strings.TrimSpace(chatResp.Choices[0].Message.Content)
		var usage *OCRUsage
		if chatResp.Usage.PromptTokens > 0 || chatResp.Usage.CompletionTokens > 0 {
			usage = &OCRUsage{
				InputTokens:  chatResp.Usage.PromptTokens,
				OutputTokens: chatResp.Usage.CompletionTokens,
			}
		}
		return text, usage, nil
	}

	// dedicated OCR endpoint: {"text":"..."} or {"content":"..."}
	var ocrResp struct {
		Text    string `json:"text"`
		Content string `json:"content"`
		Result  string `json:"result"`
	}
	if err := json.Unmarshal(respBody, &ocrResp); err == nil {
		t := ocrResp.Text
		if t == "" {
			t = ocrResp.Content
		}
		if t == "" {
			t = ocrResp.Result
		}
		if t != "" {
			return t, nil, nil
		}
	}
	return string(respBody), nil, nil
}

func buildPrompt(req *OCRRequest) string {
	task := req.Task
	if task == "" {
		task = "plain_text"
	}
	langHint := ""
	if len(req.Languages) > 0 {
		langHint = " Focus on languages: " + strings.Join(req.Languages, ", ") + "."
	}
	structured := req.ReturnLayout || req.ReturnBBox || task == "layout" || task == "table" || task == "form" || task == "invoice"

	switch task {
	case "table":
		base := "Extract all tables from this image. Preserve row and column structure."
		if structured {
			return base + langHint + ` Return JSON only: {"text":"...","blocks":[],"tables":[{"rows":[["h1","h2"],["v1","v2"]]}],"confidence":0.0}`
		}
		return base + langHint + " Return markdown tables."
	case "form":
		base := "Extract form fields and values from this image."
		if structured {
			return base + langHint + ` Return JSON only: {"text":"...","blocks":[{"type":"field","text":"label: value","bbox":[x1,y1,x2,y2],"confidence":0.0}],"tables":[],"confidence":0.0}`
		}
		return base + langHint
	case "invoice":
		base := "Extract invoice/receipt fields (merchant, date, items, amounts, tax)."
		if structured {
			return base + langHint + ` Return JSON only: {"text":"...","blocks":[],"tables":[],"confidence":0.0}`
		}
		return base + langHint
	case "layout":
		base := "Perform layout-aware OCR on this image."
		if structured {
			return base + langHint + ` Return JSON only: {"text":"full text","blocks":[{"type":"paragraph|title|table","text":"...","bbox":[x1,y1,x2,y2],"confidence":0.0}],"tables":[],"confidence":0.0}`
		}
		return base + langHint
	default:
		if structured {
			return "Extract all visible text from this image." + langHint + ` Return JSON only: {"text":"...","blocks":[],"tables":[],"confidence":0.0}`
		}
		return "Extract all visible text from this image. Return plain text only." + langHint
	}
}

func parseModelOutput(raw string, req *OCRRequest) (text string, blocks []OCRBlock, tables []OCRTable, confidence float64) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil, nil, 0
	}
	// 剥离 markdown 代码块
	if strings.HasPrefix(raw, "```") {
		raw = stripCodeFence(raw)
	}
	var parsed OCRResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil && (parsed.Text != "" || len(parsed.Blocks) > 0 || len(parsed.Tables) > 0) {
		return parsed.Text, parsed.Blocks, parsed.Tables, parsed.Confidence
	}
	return raw, nil, nil, 0
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// OpenAI-compatible chat API types

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type contentPart struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *imageURLPart `json:"image_url,omitempty"`
}

type imageURLPart struct {
	URL string `json:"url"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}
