package tools

import (
	"context"
	"testing"

	"github.com/originaleric/digeino/config"
)

func TestBaseTools_registersImageOCRWhenEnabled(t *testing.T) {
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

	baseTools, err := BaseTools(context.Background())
	if err != nil {
		t.Fatalf("BaseTools: %v", err)
	}
	for _, baseTool := range baseTools {
		info, err := baseTool.Info(context.Background())
		if err != nil {
			t.Fatalf("tool info: %v", err)
		}
		if info.Name == "image_ocr" {
			return
		}
	}
	t.Fatal("expected image_ocr in BaseTools")
}
