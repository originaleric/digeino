package ocr

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func resolveImage(ctx context.Context, req *OCRRequest) (*OCRImage, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	switch {
	case strings.TrimSpace(req.ImageURL) != "":
		dl := time.Duration(ocrCfg().URLDownloadTimeoutSec) * time.Second
		return fetchImageFromURL(ctx, strings.TrimSpace(req.ImageURL), dl)

	case strings.TrimSpace(req.ImageBase64) != "":
		dataURL, mime, err := normalizeBase64Input(req.ImageBase64, req.MimeType)
		if err != nil {
			return nil, err
		}
		raw, err := decodeDataURL(dataURL)
		if err != nil {
			if _, ok := asOCRError(err); ok {
				return nil, err
			}
			return nil, newOCRError(CodeInvalidInput, err.Error())
		}
		mime, err = validateImageBytes(raw, mime)
		if err != nil {
			return nil, err
		}
		dataURL = fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(raw))
		return &OCRImage{DataURL: dataURL, Data: raw, MimeType: mime, Source: "base64"}, nil

	case strings.TrimSpace(req.FilePath) != "":
		if err := validateFilePath(req.FilePath); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(req.FilePath)
		if err != nil {
			return nil, newOCRError(CodeInvalidInput, fmt.Sprintf("read file: %v", err))
		}
		declared := req.MimeType
		if declared == "" {
			declared = mimeFromPath(req.FilePath)
		}
		mime, err := validateImageBytes(data, declared)
		if err != nil {
			return nil, err
		}
		b64 := base64.StdEncoding.EncodeToString(data)
		dataURL := fmt.Sprintf("data:%s;base64,%s", mime, b64)
		return &OCRImage{DataURL: dataURL, Data: data, MimeType: mime, Source: "file"}, nil
	}
	return nil, newOCRError(CodeInvalidInput, errNoImageSource.Error())
}

func normalizeBase64Input(raw, mimeHint string) (dataURL string, mime string, err error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "data:") {
		parts := strings.SplitN(raw, ",", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid data URL")
		}
		header := parts[0]
		if i := strings.Index(header, ";base64"); i > 5 {
			mime = header[5:i]
		}
		dataURL = raw
		if mime == "" {
			mime = mimeHint
		}
		return dataURL, normalizeMIME(mime), nil
	}
	mime = mimeHint
	if mime == "" {
		mime = "image/png"
	}
	dataURL = fmt.Sprintf("data:%s;base64,%s", normalizeMIME(mime), raw)
	return dataURL, normalizeMIME(mime), nil
}

func decodeDataURL(dataURL string) ([]byte, error) {
	idx := strings.Index(dataURL, ",")
	if idx < 0 {
		return nil, fmt.Errorf("invalid data URL")
	}
	encoded := strings.TrimSpace(dataURL[idx+1:])
	if len(encoded) > maxEncodedImageBase64Length() {
		return nil, newOCRError(CodeImageTooLarge, fmt.Sprintf("image_base64 exceeds encoded limit %d", maxEncodedImageBase64Length()))
	}
	return base64.StdEncoding.DecodeString(encoded)
}

func mimeFromPath(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".bmp"):
		return "image/bmp"
	case strings.HasSuffix(lower, ".tiff"), strings.HasSuffix(lower, ".tif"):
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}

// fetchImageFromURL 下载 URL 图片，经 SSRF 防护与大小/MIME 校验后转为 data URL。
func fetchImageFromURL(ctx context.Context, imageURL string, downloadTimeout time.Duration) (*OCRImage, error) {
	if err := validateImageURL(imageURL); err != nil {
		return nil, err
	}
	if downloadTimeout <= 0 {
		downloadTimeout = 30 * time.Second
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, newOCRError(CodeInvalidInput, err.Error())
	}
	client := newSecureImageHTTPClient(downloadTimeout)
	resp, err := client.Do(req)
	if err != nil {
		if oe, ok := asOCRError(err); ok {
			return nil, oe
		}
		return nil, newOCRError(CodeDownloadError, "download image: "+err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, newOCRError(CodeDownloadError, fmt.Sprintf("download image: HTTP %d", resp.StatusCode))
	}

	maxBytes := maxImageBytesLimit()
	data, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxBytes)+1))
	if err != nil {
		return nil, newOCRError(CodeDownloadError, err.Error())
	}
	if len(data) > maxBytes {
		return nil, newOCRError(CodeImageTooLarge, fmt.Sprintf("image size exceeds limit %d", maxBytes))
	}

	headerMIME := resp.Header.Get("Content-Type")
	mime, err := validateImageBytes(data, headerMIME)
	if err != nil {
		return nil, err
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, b64)
	return &OCRImage{DataURL: dataURL, Data: data, MimeType: mime, Source: "url"}, nil
}
