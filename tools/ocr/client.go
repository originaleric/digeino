package ocr

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/originaleric/digeino/config"
)

// Client OCR 客户端，供固定流程节点或业务服务直接调用。
type Client struct {
	provider OCRProvider
}

// NewClient 根据配置创建 OCR 客户端。
func NewClient() (*Client, error) {
	p, err := newProviderFromConfig()
	if err != nil {
		return nil, err
	}
	return &Client{provider: p}, nil
}

// Recognize 执行 OCR。
func (c *Client) Recognize(ctx context.Context, req *OCRRequest) (*OCRResponse, error) {
	if c == nil || c.provider == nil {
		return nil, newOCRError(CodeConfigMissing, "OCR client is not initialized")
	}
	img, err := resolveImage(ctx, req)
	if err != nil {
		return nil, err
	}
	// URL 在 ocr_endpoint 模式下由 provider 自行下载；chat 模式可直接传 URL
	return c.provider.Recognize(ctx, req, img)
}

func newProviderFromConfig() (OCRProvider, error) {
	cfg := config.Get().Tools.OCR
	if cfg.Enabled == nil || !*cfg.Enabled {
		return nil, newOCRError(CodeConfigMissing, "Tools.OCR is not enabled")
	}
	providerName := normalizeProviderName(cfg.Provider)
	if providerName == "" {
		providerName = "deepseek-ocr"
	}

	if providerName == "deepseek" {
		providerName = "deepseek-ocr"
	}
	providerName = canonicalOpenAICompatibleVisionProviderName(providerName)
	providerName = canonicalMultipartOCRProviderName(providerName)

	registry := configuredProviderRegistry(providerName)
	if providerName == "deepseek-ocr" {
		deepSeek, err := newDeepSeekProvider(cfg.DeepSeek, cfg)
		if err != nil {
			return nil, err
		}
		registry.Register(deepSeek)
	}
	if providerName == defaultOpenAICompatibleVisionProvider {
		compatibleVision, err := newOpenAICompatibleVisionProvider(cfg.OpenAICompatible, cfg)
		if err != nil {
			return nil, err
		}
		registry.Register(compatibleVision)
	}
	if providerName == defaultMultipartOCRProvider {
		multipartOCR, err := newMultipartOCRProvider(cfg.MultipartOCR, cfg)
		if err != nil {
			return nil, err
		}
		registry.Register(multipartOCR)
	}

	if p, ok := registry.Get(providerName); ok {
		return p, nil
	}
	return nil, newOCRError(CodeConfigMissing, fmt.Sprintf("unsupported OCR provider %q", cfg.Provider))
}

func deepSeekAPIKey(ds config.DeepSeekOCRConfig) string {
	if key := resolvedSecret(ds.ApiKey); key != "" {
		return key
	}
	for _, env := range []string{"DEEPSEEK_OCR_API_KEY", "DEEPSEEK_API_KEY"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return ""
}

func resolvedSecret(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	expanded := strings.TrimSpace(os.ExpandEnv(raw))
	if expanded == "" || strings.Contains(expanded, "${") {
		return ""
	}
	return expanded
}
