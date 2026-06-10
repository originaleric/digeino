package ocr

import (
	"context"
	"testing"

	"github.com/originaleric/digeino/config"
)

type fakeProvider struct {
	name string
}

func (p fakeProvider) Name() string {
	return p.name
}

func (p fakeProvider) Recognize(context.Context, *OCRRequest, *OCRImage) (*OCRResponse, error) {
	return &OCRResponse{Text: "ok", Provider: p.name}, nil
}

func TestOCRRegistry(t *testing.T) {
	registry := NewOCRRegistry("fake")
	registry.Register(fakeProvider{name: "fake"})

	provider, ok := registry.Get(" fake ")
	if !ok || provider.Name() != "fake" {
		t.Fatalf("expected provider from registry, got ok=%v provider=%v", ok, provider)
	}
	if registry.Default().Name() != "fake" {
		t.Fatalf("unexpected default provider: %v", registry.Default())
	}
}

func TestDeepSeekAPIKey_envAndPlaceholder(t *testing.T) {
	t.Setenv("DEEPSEEK_OCR_API_KEY", "ocr-env-key")
	if got := deepSeekAPIKey(config.DeepSeekOCRConfig{ApiKey: "${DEEPSEEK_OCR_API_KEY}"}); got != "ocr-env-key" {
		t.Fatalf("expected expanded env key, got %q", got)
	}

	t.Setenv("DEEPSEEK_OCR_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "fallback-key")
	if got := deepSeekAPIKey(config.DeepSeekOCRConfig{ApiKey: "${DEEPSEEK_OCR_API_KEY}"}); got != "fallback-key" {
		t.Fatalf("expected fallback env key, got %q", got)
	}
}

func TestNewProviderFromConfig_usesRegistryDefault(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	cfg := config.Default()
	enabled := true
	cfg.Tools.OCR.Enabled = &enabled
	cfg.Tools.OCR.Provider = "deepseek"
	cfg.Tools.OCR.DeepSeek.ApiKey = "test-key"
	config.Set(cfg)

	provider, err := newProviderFromConfig()
	if err != nil {
		t.Fatalf("newProviderFromConfig: %v", err)
	}
	if provider.Name() != "deepseek-ocr" {
		t.Fatalf("unexpected provider: %s", provider.Name())
	}
}

func TestNewProviderFromConfig_usesRegisteredProviderWithoutDeepSeekKey(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	const providerName = "unit-external-ocr"
	RegisterOCRProvider(fakeProvider{name: providerName})

	cfg := config.Default()
	enabled := true
	cfg.Tools.OCR.Enabled = &enabled
	cfg.Tools.OCR.Provider = providerName
	cfg.Tools.OCR.DeepSeek.ApiKey = ""
	config.Set(cfg)

	provider, err := newProviderFromConfig()
	if err != nil {
		t.Fatalf("newProviderFromConfig: %v", err)
	}
	if provider.Name() != providerName {
		t.Fatalf("unexpected provider: %s", provider.Name())
	}
}

func TestGetOCRProvider_returnsRegisteredProvider(t *testing.T) {
	const providerName = "unit-get-ocr"
	RegisterOCRProvider(fakeProvider{name: providerName})

	provider, ok := GetOCRProvider(providerName)
	if !ok || provider.Name() != providerName {
		t.Fatalf("expected registered provider, got ok=%v provider=%v", ok, provider)
	}
}

func TestDeepSeekPostJSON_classifiesPayloadMarshalError(t *testing.T) {
	provider := &deepSeekProvider{}

	_, _, err := provider.postJSON(context.Background(), "https://example.com/v1/ocr", map[string]any{
		"bad": make(chan int),
	})
	oe, ok := asOCRError(err)
	if !ok || oe.Code != CodeInvalidInput {
		t.Fatalf("expected %s, got ok=%v err=%v", CodeInvalidInput, ok, err)
	}
}

func TestDeepSeekPostJSON_classifiesInvalidProviderURL(t *testing.T) {
	provider := &deepSeekProvider{}

	_, _, err := provider.postJSON(context.Background(), "://bad-url", map[string]any{"ok": true})
	oe, ok := asOCRError(err)
	if !ok || oe.Code != CodeConfigMissing {
		t.Fatalf("expected %s, got ok=%v err=%v", CodeConfigMissing, ok, err)
	}
}
