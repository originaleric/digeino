package ocr

import "errors"

// 归一化错误码（供上层治理与审计）。
const (
	CodeInvalidInput    = "OCR_INVALID_INPUT"
	CodeConfigMissing   = "OCR_CONFIG_MISSING"
	CodeImageTooLarge   = "OCR_IMAGE_TOO_LARGE"
	CodeMimeNotAllowed  = "OCR_MIME_NOT_ALLOWED"
	CodePathNotAllowed  = "OCR_PATH_NOT_ALLOWED"
	CodeURLNotAllowed   = "OCR_URL_NOT_ALLOWED"
	CodeDownloadError   = "OCR_DOWNLOAD_ERROR"
	CodeProviderError   = "OCR_PROVIDER_ERROR"
	CodeProviderTimeout = "OCR_PROVIDER_TIMEOUT"
)

// OCRError 带码错误。
type OCRError struct {
	Code    string
	Message string
}

func (e *OCRError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}

func newOCRError(code, msg string) error {
	return &OCRError{Code: code, Message: msg}
}

var (
	errNoImageSource = errors.New("one of image_url, image_base64, or file_path is required")
)

// asOCRError 从被包装的错误（如 *url.Error）中提取 OCRError。
func asOCRError(err error) (*OCRError, bool) {
	var oe *OCRError
	if errors.As(err, &oe) {
		return oe, true
	}
	return nil, false
}
