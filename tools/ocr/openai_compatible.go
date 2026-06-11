package ocr

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/originaleric/digeino/config"
)

const defaultOpenAICompatibleVisionProvider = "openai-compatible-vision"

type openAICompatibleVisionProvider struct {
	cfg    config.OpenAICompatibleVisionOCRConfig
	ocrCfg config.OCRConfig
	client *http.Client
}

func newOpenAICompatibleVisionProvider(cfg config.OpenAICompatibleVisionOCRConfig, ocrCfg config.OCRConfig) (*openAICompatibleVisionProvider, error) {
	if openAICompatibleVisionAPIKey(cfg) == "" {
		return nil, newOCRError(CodeConfigMissing, "OpenAI compatible vision ApiKey not configured (Tools.OCR.OpenAICompatible.ApiKey or QWEN_API_KEY)")
	}
	if err := validateOpenAICompatibleVisionBaseURL(cfg); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, newOCRError(CodeConfigMissing, "OpenAI compatible vision Model not configured (Tools.OCR.OpenAICompatible.Model)")
	}
	timeout := time.Duration(ocrCfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &openAICompatibleVisionProvider{
		cfg:    cfg,
		ocrCfg: ocrCfg,
		client: &http.Client{Timeout: timeout},
	}, nil
}

func (p *openAICompatibleVisionProvider) Name() string {
	return defaultOpenAICompatibleVisionProvider
}

func (p *openAICompatibleVisionProvider) Recognize(ctx context.Context, req *OCRRequest, img *OCRImage) (*OCRResponse, error) {
	endpoint := p.baseURL() + "/chat/completions"
	body := map[string]any{
		"model": p.model(),
		"messages": []chatMessage{
			{
				Role: "user",
				Content: []contentPart{
					{Type: "text", Text: buildPrompt(req)},
					{Type: "image_url", ImageURL: &imageURLPart{URL: img.DataURL}},
				},
			},
		},
		"stream": false,
	}
	for k, v := range req.Options {
		if _, exists := body[k]; !exists {
			body[k] = v
		}
	}

	raw, usage, err := p.doWithRetry(ctx, endpoint, body)
	if err != nil {
		return nil, err
	}
	text, blocks, tables, conf, parsedMetadata := parseModelOutput(raw, req)
	return &OCRResponse{
		Text:       text,
		Blocks:     blocks,
		Tables:     tables,
		Confidence: conf,
		Provider:   p.Name(),
		Model:      p.model(),
		Metadata:   mergeMetadata(parsedMetadata, p.metadata(req, img, endpoint)),
		Usage:      usage,
	}, nil
}

func (p *openAICompatibleVisionProvider) model() string {
	return strings.TrimSpace(p.cfg.Model)
}

func (p *openAICompatibleVisionProvider) baseURL() string {
	return strings.TrimRight(strings.TrimSpace(p.cfg.BaseUrl), "/")
}

func (p *openAICompatibleVisionProvider) doWithRetry(ctx context.Context, url string, body any) (content string, usage *OCRUsage, err error) {
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
				return "", nil, newOCRError(CodeProviderTimeout, ctx.Err().Error())
			case <-time.After(backoff):
			}
		}
		content, usage, lastErr = postOpenAICompatibleJSON(ctx, p.client, openAICompatibleVisionAPIKey(p.cfg), url, body)
		if lastErr == nil {
			return content, usage, nil
		}
		if !isRetryable(lastErr) {
			return "", nil, lastErr
		}
	}
	return "", nil, lastErr
}

func (p *openAICompatibleVisionProvider) metadata(req *OCRRequest, img *OCRImage, endpoint string) map[string]any {
	metadata := map[string]any{
		"mode":     "openai_compatible_vision",
		"endpoint": endpoint,
	}
	if img != nil {
		metadata["source"] = img.Source
		metadata["mime_type"] = img.MimeType
	}
	if req != nil {
		task := req.Task
		if task == "" {
			task = "plain_text"
		}
		metadata["task"] = task
		if languages := requestLanguages(req); len(languages) > 0 {
			metadata["languages"] = languages
		}
	}
	return metadata
}

func openAICompatibleVisionAPIKey(cfg config.OpenAICompatibleVisionOCRConfig) string {
	if key := resolvedSecret(cfg.ApiKey); key != "" {
		return key
	}
	for _, env := range []string{"OPENAI_COMPATIBLE_VISION_API_KEY", "QWEN_API_KEY", "DASHSCOPE_API_KEY", "OPENAI_API_KEY"} {
		if v := strings.TrimSpace(os.Getenv(env)); v != "" {
			return v
		}
	}
	return ""
}

func openAICompatibleVisionProviderAliases(name string) bool {
	switch normalizeProviderName(name) {
	case "openai-compatible-vision", "openai-compatible", "openai_compatible_vision", "qwen-vl", "qwen3-vl":
		return true
	default:
		return false
	}
}

func canonicalOpenAICompatibleVisionProviderName(name string) string {
	if openAICompatibleVisionProviderAliases(name) {
		return defaultOpenAICompatibleVisionProvider
	}
	return normalizeProviderName(name)
}

func validateOpenAICompatibleVisionBaseURL(cfg config.OpenAICompatibleVisionOCRConfig) error {
	if strings.TrimSpace(cfg.BaseUrl) == "" {
		return newOCRError(CodeConfigMissing, "OpenAI compatible vision BaseUrl not configured (Tools.OCR.OpenAICompatible.BaseUrl)")
	}
	if strings.Contains(strings.TrimSpace(cfg.BaseUrl), "/chat/completions") {
		return newOCRError(CodeConfigMissing, fmt.Sprintf("OpenAI compatible vision BaseUrl should not include /chat/completions: %q", cfg.BaseUrl))
	}
	return nil
}
