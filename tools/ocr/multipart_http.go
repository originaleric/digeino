package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/originaleric/digeino/config"
)

const defaultMultipartOCRProvider = "multipart-ocr-http"

type multipartOCRProvider struct {
	cfg    config.MultipartOCRConfig
	ocrCfg config.OCRConfig
	client *http.Client
}

func newMultipartOCRProvider(cfg config.MultipartOCRConfig, ocrCfg config.OCRConfig) (*multipartOCRProvider, error) {
	if strings.TrimSpace(cfg.BaseUrl) == "" {
		return nil, newOCRError(CodeConfigMissing, "Multipart OCR BaseUrl not configured (Tools.OCR.MultipartOCR.BaseUrl)")
	}
	timeout := time.Duration(ocrCfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &multipartOCRProvider{
		cfg:    cfg,
		ocrCfg: ocrCfg,
		client: &http.Client{Timeout: timeout},
	}, nil
}

func (p *multipartOCRProvider) Name() string {
	return defaultMultipartOCRProvider
}

func (p *multipartOCRProvider) Recognize(ctx context.Context, req *OCRRequest, img *OCRImage) (*OCRResponse, error) {
	if img == nil || len(img.Data) == 0 {
		return nil, newOCRError(CodeInvalidInput, "multipart OCR requires resolved image bytes")
	}
	raw, err := p.doWithRetry(ctx, p.endpoint(), req, img)
	if err != nil {
		return nil, err
	}
	text, blocks, tables, conf, parsedMetadata := parseModelOutput(raw, req)
	if text == "" {
		text = raw
	}
	return &OCRResponse{
		Text:       text,
		Blocks:     blocks,
		Tables:     tables,
		Confidence: conf,
		Provider:   p.Name(),
		Model:      p.model(),
		Metadata:   mergeMetadata(parsedMetadata, p.metadata(req, img)),
	}, nil
}

func (p *multipartOCRProvider) endpoint() string {
	base := strings.TrimRight(strings.TrimSpace(p.cfg.BaseUrl), "/")
	endpoint := strings.TrimSpace(p.cfg.OCREndpoint)
	if endpoint == "" {
		endpoint = "/v1/ocr"
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	return base + endpoint
}

func (p *multipartOCRProvider) model() string {
	if model := strings.TrimSpace(p.cfg.Model); model != "" {
		return model
	}
	return defaultMultipartOCRProvider
}

func (p *multipartOCRProvider) doWithRetry(ctx context.Context, url string, req *OCRRequest, img *OCRImage) (string, error) {
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
				return "", newOCRError(CodeProviderTimeout, ctx.Err().Error())
			case <-time.After(backoff):
			}
		}
		var content string
		content, lastErr = p.postMultipart(ctx, url, req, img)
		if lastErr == nil {
			return content, nil
		}
		if !isRetryable(lastErr) {
			return "", lastErr
		}
	}
	return "", lastErr
}

func (p *multipartOCRProvider) postMultipart(ctx context.Context, url string, req *OCRRequest, img *OCRImage) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	fileField := defaultString(p.cfg.FileField, "file")
	filePart, err := writer.CreateFormFile(fileField, multipartOCRFilename(img))
	if err != nil {
		return "", newOCRError(CodeInvalidInput, "create multipart file field: "+err.Error())
	}
	if _, err := filePart.Write(img.Data); err != nil {
		return "", newOCRError(CodeInvalidInput, "write multipart file field: "+err.Error())
	}

	if promptField, ok := optionalMultipartField(p.cfg.PromptField, "prompt"); ok {
		if err := writer.WriteField(promptField, buildPrompt(req)); err != nil {
			return "", newOCRError(CodeInvalidInput, "write multipart prompt field: "+err.Error())
		}
	}
	if languageField, ok := optionalMultipartField(p.cfg.LanguageField, "language"); ok {
		if languages := requestLanguages(req); len(languages) > 0 {
			if err := writer.WriteField(languageField, strings.Join(languages, ",")); err != nil {
				return "", newOCRError(CodeInvalidInput, "write multipart language field: "+err.Error())
			}
		}
	}
	if err := writer.Close(); err != nil {
		return "", newOCRError(CodeInvalidInput, "close multipart payload: "+err.Error())
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return "", newOCRError(CodeConfigMissing, "invalid multipart OCR provider URL: "+err.Error())
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	if apiKey := multipartOCRAPIKey(p.cfg); apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return "", newOCRError(CodeProviderTimeout, err.Error())
		}
		return "", newOCRError(CodeProviderError, err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	if err != nil {
		return "", newOCRError(CodeProviderError, err.Error())
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", newOCRError(CodeProviderError, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 512)))
	}
	return extractMultipartOCRText(respBody, p.responseTextPath())
}

func (p *multipartOCRProvider) responseTextPath() string {
	return defaultString(p.cfg.ResponseTextPath, "text")
}

func (p *multipartOCRProvider) metadata(req *OCRRequest, img *OCRImage) map[string]any {
	metadata := map[string]any{
		"mode":               "multipart_ocr_endpoint",
		"endpoint":           p.endpoint(),
		"response_text_path": p.responseTextPath(),
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

func multipartOCRAPIKey(cfg config.MultipartOCRConfig) string {
	if key := resolvedSecret(cfg.ApiKey); key != "" {
		return key
	}
	for _, env := range []string{"MULTIPART_OCR_API_KEY", "OCR_HTTP_API_KEY"} {
		if v := strings.TrimSpace(os.Getenv(env)); v != "" {
			return v
		}
	}
	return ""
}

func multipartOCRProviderAliases(name string) bool {
	switch normalizeProviderName(name) {
	case "multipart-ocr-http", "multipart-ocr", "http-ocr", "ocr-http":
		return true
	default:
		return false
	}
}

func canonicalMultipartOCRProviderName(name string) string {
	if multipartOCRProviderAliases(name) {
		return defaultMultipartOCRProvider
	}
	return normalizeProviderName(name)
}

func multipartOCRFilename(img *OCRImage) string {
	ext := ".img"
	if img != nil {
		switch normalizeMIME(img.MimeType) {
		case "image/png":
			ext = ".png"
		case "image/jpeg":
			ext = ".jpg"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		case "image/bmp":
			ext = ".bmp"
		case "image/tiff":
			ext = ".tiff"
		}
	}
	return "image" + ext
}

func extractMultipartOCRText(data []byte, path string) (string, error) {
	path = defaultString(path, "text")
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return strings.TrimSpace(string(data)), nil
	}
	value, ok := lookupJSONPath(parsed, path)
	if !ok {
		return "", newOCRError(CodeProviderError, fmt.Sprintf("multipart OCR response missing text path %q", path))
	}
	text, ok := value.(string)
	if !ok {
		return "", newOCRError(CodeProviderError, fmt.Sprintf("multipart OCR response text path %q is not a string", path))
	}
	return strings.TrimSpace(text), nil
}

func lookupJSONPath(value any, path string) (any, bool) {
	current := value
	for _, part := range strings.Split(path, ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = obj[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func optionalMultipartField(value, fallback string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "-" {
		return "", false
	}
	return defaultString(value, fallback), true
}
