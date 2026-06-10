package ocr_test

import (
	"context"
	"testing"

	"github.com/originaleric/digeino/tools/ocr"
)

type externalProvider struct{}

func (externalProvider) Name() string {
	return "external-compile-ocr"
}

func (externalProvider) Recognize(context.Context, *ocr.OCRRequest, *ocr.OCRImage) (*ocr.OCRResponse, error) {
	return &ocr.OCRResponse{Text: "ok", Provider: "external-compile-ocr"}, nil
}

func TestExternalPackageCanImplementOCRProvider(t *testing.T) {
	var provider ocr.OCRProvider = externalProvider{}
	ocr.RegisterOCRProvider(provider)

	got, ok := ocr.GetOCRProvider(provider.Name())
	if !ok || got.Name() != provider.Name() {
		t.Fatalf("expected external provider registration, got ok=%v provider=%v", ok, got)
	}
}
