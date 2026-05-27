package ocr

import (
	"context"
	"net"
	"testing"

	"github.com/originaleric/digeino/config"
)

// minimal PNG header
var pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}

func TestValidateImageURL_blocksPrivate(t *testing.T) {
	cfg := config.Default()
	enabled := true
	cfg.Tools.OCR.BlockPrivateNetworks = &enabled
	config.Set(cfg)

	cases := []string{
		"http://127.0.0.1/a.png",
		"http://localhost/a.png",
		"http://10.0.0.1/a.png",
	}
	for _, u := range cases {
		if err := validateImageURL(u); err == nil {
			t.Fatalf("expected block for %s", u)
		}
	}
}

func TestValidateRequest_singleSource(t *testing.T) {
	if err := validateRequest(&OCRRequest{}); err == nil {
		t.Fatal("expected error for empty sources")
	}
	if err := validateRequest(&OCRRequest{
		ImageURL:    "https://example.com/a.png",
		ImageBase64: "abc",
	}); err == nil {
		t.Fatal("expected error for multiple sources")
	}
}

func TestNormalizeMIME_jpgAlias(t *testing.T) {
	if got := normalizeMIME("image/jpg"); got != "image/jpeg" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateImageBytes_sniffAndDeclared(t *testing.T) {
	cfg := config.Default()
	cfg.Tools.OCR.MaxImageBytes = 1024
	config.Set(cfg)

	mime, err := validateImageBytes(pngMagic, "image/png")
	if err != nil || mime != "image/png" {
		t.Fatalf("png: mime=%q err=%v", mime, err)
	}
	if _, err := validateImageBytes(pngMagic, "image/jpeg"); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestValidateImageBytes_maxSize(t *testing.T) {
	cfg := config.Default()
	cfg.Tools.OCR.MaxImageBytes = 8
	config.Set(cfg)
	big := append(pngMagic, make([]byte, 16)...)
	if _, err := validateImageBytes(big, ""); err == nil {
		t.Fatal("expected too large")
	}
}

func TestCheckIPNotPrivate(t *testing.T) {
	if err := checkIPNotPrivate(net.ParseIP("127.0.0.1")); err == nil {
		t.Fatal("expected block loopback")
	}
	if err := checkIPNotPrivate(net.ParseIP("8.8.8.8")); err != nil {
		t.Fatalf("public IP should pass: %v", err)
	}
}

func TestResolveValidatedHostIPs_literalIP(t *testing.T) {
	cfg := config.Default()
	enabled := true
	cfg.Tools.OCR.BlockPrivateNetworks = &enabled
	config.Set(cfg)

	ips, err := resolveValidatedHostIPs(context.Background(), "8.8.8.8")
	if err != nil || len(ips) != 1 || ips[0].String() != "8.8.8.8" {
		t.Fatalf("public literal: ips=%v err=%v", ips, err)
	}
	if _, err := resolveValidatedHostIPs(context.Background(), "127.0.0.1"); err == nil {
		t.Fatal("expected block for loopback literal")
	}
}

func TestParseModelOutput_json(t *testing.T) {
	raw := `{"text":"hello","blocks":[{"type":"paragraph","text":"hello","confidence":0.9}],"confidence":0.9}`
	text, blocks, _, conf := parseModelOutput(raw, &OCRRequest{ReturnLayout: true})
	if text != "hello" || len(blocks) != 1 || conf != 0.9 {
		t.Fatalf("unexpected parse: text=%q blocks=%d conf=%v", text, len(blocks), conf)
	}
}
