package ocr

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/originaleric/digeino/config"
)

func ocrCfg() config.OCRConfig {
	return config.Get().Tools.OCR
}

// validateRequest 校验输入与安全策略。
func validateRequest(req *OCRRequest) error {
	if req == nil {
		return newOCRError(CodeInvalidInput, "request is nil")
	}
	sources := 0
	if strings.TrimSpace(req.ImageURL) != "" {
		sources++
	}
	if strings.TrimSpace(req.ImageBase64) != "" {
		sources++
	}
	if strings.TrimSpace(req.FilePath) != "" {
		sources++
	}
	if sources == 0 {
		return newOCRError(CodeInvalidInput, errNoImageSource.Error())
	}
	if sources > 1 {
		return newOCRError(CodeInvalidInput, "only one of image_url, image_base64, or file_path may be set")
	}
	if req.Task != "" {
		switch req.Task {
		case "plain_text", "layout", "table", "form", "invoice":
		default:
			return newOCRError(CodeInvalidInput, fmt.Sprintf("unsupported task %q", req.Task))
		}
	}
	return nil
}

func validateImageURL(raw string) error {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return newOCRError(CodeInvalidInput, "invalid image_url: "+err.Error())
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return newOCRError(CodeInvalidInput, "image_url must use http or https")
	}
	if u.Hostname() == "" {
		return newOCRError(CodeInvalidInput, "image_url missing host")
	}
	cfg := ocrCfg()
	if blockPrivate := cfg.BlockPrivateNetworks == nil || *cfg.BlockPrivateNetworks; blockPrivate {
		if err := checkHostNotPrivate(u.Hostname()); err != nil {
			return err
		}
	}
	if len(cfg.AllowedImageDomains) > 0 {
		host := strings.ToLower(u.Hostname())
		allowed := false
		for _, d := range cfg.AllowedImageDomains {
			d = strings.ToLower(strings.TrimSpace(d))
			if d == "" {
				continue
			}
			if host == d || strings.HasSuffix(host, "."+d) {
				allowed = true
				break
			}
		}
		if !allowed {
			return newOCRError(CodeURLNotAllowed, fmt.Sprintf("domain %q is not in AllowedImageDomains", u.Hostname()))
		}
	}
	return nil
}

func checkHostNotPrivate(host string) error {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "localhost" || strings.HasSuffix(h, ".localhost") {
		return newOCRError(CodeURLNotAllowed, "localhost URLs are not allowed")
	}
	if ip := net.ParseIP(h); ip != nil {
		return checkIPNotPrivate(ip)
	}
	return nil
}

func maxImageBytesLimit() int {
	max := ocrCfg().MaxImageBytes
	if max <= 0 {
		return 10 * 1024 * 1024
	}
	return max
}

// normalizeMIME 规范化 MIME（去参数、jpg→jpeg）。
func normalizeMIME(mime string) string {
	mime = strings.ToLower(strings.TrimSpace(mime))
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}
	if mime == "image/jpg" {
		return "image/jpeg"
	}
	return mime
}

// detectImageMIME 基于文件魔数嗅探图片 MIME。
func detectImageMIME(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	sniff := normalizeMIME(http.DetectContentType(data))
	if strings.HasPrefix(sniff, "image/") {
		return sniff
	}
	return ""
}

func mimeCompatible(declared, sniffed string) bool {
	declared = normalizeMIME(declared)
	sniffed = normalizeMIME(sniffed)
	if declared == "" || declared == "application/octet-stream" {
		return true
	}
	return declared == sniffed
}

func validateFilePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return newOCRError(CodeInvalidInput, "file_path is empty")
	}
	cfg := ocrCfg()
	prefixes := cfg.AllowedFilePaths
	if len(prefixes) == 0 {
		return newOCRError(CodePathNotAllowed, "file_path is disabled without Tools.OCR.AllowedFilePaths")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return newOCRError(CodeInvalidInput, err.Error())
	}
	clean := filepath.Clean(abs)
	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		base, err := filepath.Abs(prefix)
		if err != nil {
			continue
		}
		base = filepath.Clean(base)
		if clean == base || strings.HasPrefix(clean, base+string(filepath.Separator)) {
			return nil
		}
	}
	return newOCRError(CodePathNotAllowed, fmt.Sprintf("path %q is not under AllowedFilePaths", path))
}

func validateImageBytes(data []byte, declaredMIME string) (effectiveMIME string, err error) {
	maxBytes := maxImageBytesLimit()
	if len(data) > maxBytes {
		return "", newOCRError(CodeImageTooLarge, fmt.Sprintf("image size %d exceeds limit %d", len(data), maxBytes))
	}
	sniffed := detectImageMIME(data)
	if sniffed == "" {
		return "", newOCRError(CodeMimeNotAllowed, "content is not a recognized image format")
	}
	declared := normalizeMIME(declaredMIME)
	if !mimeCompatible(declared, sniffed) {
		return "", newOCRError(CodeMimeNotAllowed, fmt.Sprintf("declared mime %q does not match content %q", declared, sniffed))
	}
	effective := sniffed
	cfg := ocrCfg()
	allowed := cfg.AllowedMimeTypes
	if len(allowed) == 0 {
		allowed = []string{
			"image/png", "image/jpeg", "image/jpg", "image/webp",
			"image/gif", "image/bmp", "image/tiff",
		}
	}
	for _, a := range allowed {
		if normalizeMIME(a) == effective {
			return effective, nil
		}
	}
	return "", newOCRError(CodeMimeNotAllowed, fmt.Sprintf("mime type %q is not allowed", effective))
}
