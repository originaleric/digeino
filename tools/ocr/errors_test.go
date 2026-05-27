package ocr

import (
	"net/url"
	"testing"
)

func TestAsOCRError_wrappedURLError(t *testing.T) {
	inner := newOCRError(CodeURLNotAllowed, "private or loopback address")
	wrapped := &url.Error{Op: "Get", URL: "https://example.com/x.png", Err: inner}
	oe, ok := asOCRError(wrapped)
	if !ok || oe.Code != CodeURLNotAllowed {
		t.Fatalf("got ok=%v code=%q", ok, oe.Code)
	}
}
