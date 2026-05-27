package ocr

import "context"

// OCRProvider OCR 模型提供者接口。
type OCRProvider interface {
	Recognize(ctx context.Context, req *OCRRequest, img *resolvedImage) (*OCRResponse, error)
	Name() string
}
