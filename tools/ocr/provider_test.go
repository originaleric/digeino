package ocr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestNewProviderFromConfig_usesOpenAICompatibleVision(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	cfg := config.Default()
	enabled := true
	cfg.Tools.OCR.Enabled = &enabled
	cfg.Tools.OCR.Provider = "openai-compatible-vision"
	cfg.Tools.OCR.OpenAICompatible.ApiKey = "test-key"
	cfg.Tools.OCR.OpenAICompatible.BaseUrl = "https://example.com/v1"
	cfg.Tools.OCR.OpenAICompatible.Model = "qwen3-vl-plus"
	config.Set(cfg)

	provider, err := newProviderFromConfig()
	if err != nil {
		t.Fatalf("newProviderFromConfig: %v", err)
	}
	if provider.Name() != "openai-compatible-vision" {
		t.Fatalf("unexpected provider: %s", provider.Name())
	}
}

func TestNewProviderFromConfig_emptyProviderKeepsDeepSeekCompatibility(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	cfg := config.Default()
	enabled := true
	cfg.Tools.OCR.Enabled = &enabled
	cfg.Tools.OCR.Provider = ""
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

func TestNewProviderFromConfig_acceptsOpenAICompatibleVisionAlias(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	cfg := config.Default()
	enabled := true
	cfg.Tools.OCR.Enabled = &enabled
	cfg.Tools.OCR.Provider = "qwen-vl"
	cfg.Tools.OCR.OpenAICompatible.ApiKey = "test-key"
	cfg.Tools.OCR.OpenAICompatible.BaseUrl = "https://example.com/v1"
	cfg.Tools.OCR.OpenAICompatible.Model = "qwen3-vl-plus"
	config.Set(cfg)

	provider, err := newProviderFromConfig()
	if err != nil {
		t.Fatalf("newProviderFromConfig: %v", err)
	}
	if provider.Name() != "openai-compatible-vision" {
		t.Fatalf("unexpected provider: %s", provider.Name())
	}
}

func TestNewProviderFromConfig_usesMultipartOCRProvider(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	cfg := config.Default()
	enabled := true
	cfg.Tools.OCR.Enabled = &enabled
	cfg.Tools.OCR.Provider = "multipart-ocr-http"
	cfg.Tools.OCR.MultipartOCR.BaseUrl = "https://example.com"
	cfg.Tools.OCR.MultipartOCR.OCREndpoint = "/v1/ocr"
	cfg.Tools.OCR.MultipartOCR.ResponseTextPath = "data.text"
	config.Set(cfg)

	provider, err := newProviderFromConfig()
	if err != nil {
		t.Fatalf("newProviderFromConfig: %v", err)
	}
	if provider.Name() != "multipart-ocr-http" {
		t.Fatalf("unexpected provider: %s", provider.Name())
	}
}

func TestMultipartOCRRecognize_postsMultipartAndReadsResponsePath(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotContentType string
	var gotPrompt string
	var gotLanguage string
	var gotFileBytes string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		gotPrompt = r.FormValue("prompt")
		gotLanguage = r.FormValue("language")
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile: %v", err)
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		gotFileBytes = string(data)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"text":"hello multipart"}}`))
	}))
	defer server.Close()

	provider, err := newMultipartOCRProvider(config.MultipartOCRConfig{
		ApiKey:           "service-key",
		BaseUrl:          server.URL,
		OCREndpoint:      "/v1/ocr",
		ResponseTextPath: "data.text",
		Model:            "internal-ocr",
	}, config.OCRConfig{})
	if err != nil {
		t.Fatalf("newMultipartOCRProvider: %v", err)
	}

	resp, err := provider.Recognize(context.Background(), &OCRRequest{
		Task:     "plain_text",
		Language: "zh",
	}, &OCRImage{
		DataURL:  "data:image/png;base64,aW1hZ2UtYnl0ZXM=",
		Data:     []byte("image-bytes"),
		MimeType: "image/png",
		Source:   "base64",
	})
	if err != nil {
		t.Fatalf("Recognize: %v", err)
	}
	if gotPath != "/v1/ocr" {
		t.Fatalf("expected /v1/ocr, got %q", gotPath)
	}
	if gotAuth != "Bearer service-key" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Fatalf("unexpected content type: %q", gotContentType)
	}
	if gotLanguage != "zh" {
		t.Fatalf("unexpected language: %q", gotLanguage)
	}
	if gotPrompt == "" {
		t.Fatal("expected prompt field")
	}
	if gotFileBytes != "image-bytes" {
		t.Fatalf("unexpected file bytes: %q", gotFileBytes)
	}
	if resp.Provider != "multipart-ocr-http" || resp.Model != "internal-ocr" || resp.Text != "hello multipart" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestMultipartOCRRecognize_canDisablePromptAndLanguageFields(t *testing.T) {
	var sawPrompt bool
	var sawLanguage bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		_, sawPrompt = r.MultipartForm.Value["prompt"]
		_, sawLanguage = r.MultipartForm.Value["language"]
		if _, _, err := r.FormFile("file"); err != nil {
			t.Fatalf("FormFile: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"file only"}`))
	}))
	defer server.Close()

	provider, err := newMultipartOCRProvider(config.MultipartOCRConfig{
		BaseUrl:       server.URL,
		PromptField:   "-",
		LanguageField: "-",
	}, config.OCRConfig{})
	if err != nil {
		t.Fatalf("newMultipartOCRProvider: %v", err)
	}

	resp, err := provider.Recognize(context.Background(), &OCRRequest{
		Task:     "plain_text",
		Language: "zh",
	}, &OCRImage{
		Data:     []byte("image-bytes"),
		MimeType: "image/png",
		Source:   "base64",
	})
	if err != nil {
		t.Fatalf("Recognize: %v", err)
	}
	if sawPrompt || sawLanguage {
		t.Fatalf("expected prompt/language fields disabled, sawPrompt=%v sawLanguage=%v", sawPrompt, sawLanguage)
	}
	if resp.Text != "file only" {
		t.Fatalf("unexpected text: %q", resp.Text)
	}
}

func TestOpenAICompatibleVisionRecognize_usesChatCompletionsEndpoint(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		var payload struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		gotModel = payload.Model
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello"}}],"usage":{"prompt_tokens":2,"completion_tokens":1}}`))
	}))
	defer server.Close()

	provider, err := newOpenAICompatibleVisionProvider(config.OpenAICompatibleVisionOCRConfig{
		ApiKey:  "test-key",
		BaseUrl: server.URL,
		Model:   "qwen3-vl-plus",
	}, config.OCRConfig{})
	if err != nil {
		t.Fatalf("newOpenAICompatibleVisionProvider: %v", err)
	}

	resp, err := provider.Recognize(context.Background(), &OCRRequest{Task: "plain_text"}, &OCRImage{
		DataURL:  "data:image/png;base64,abc",
		MimeType: "image/png",
		Source:   "base64",
	})
	if err != nil {
		t.Fatalf("Recognize: %v", err)
	}
	if gotPath != "/chat/completions" {
		t.Fatalf("expected /chat/completions, got %q", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotModel != "qwen3-vl-plus" {
		t.Fatalf("unexpected model: %q", gotModel)
	}
	if resp.Provider != "openai-compatible-vision" || resp.Model != "qwen3-vl-plus" || resp.Text != "hello" {
		t.Fatalf("unexpected response: %+v", resp)
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
