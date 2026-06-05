//go:build !unix

package research

import (
	"context"
	"testing"

	"github.com/originaleric/digeino/config"
)

func TestSecureFileAccessUnsupportedPlatform(t *testing.T) {
	if f, err := openAllowedReadFile("test.txt"); err == nil {
		_ = f.Close()
		t.Fatal("expected secure read access to fail on unsupported platform")
	}

	if f, _, err := openAllowedWriteFile("test.txt", "overwrite"); err == nil {
		_ = f.Close()
		t.Fatal("expected secure write access to fail on unsupported platform")
	}
}

func TestFilesystemToolsNotRegisteredOnUnsupportedPlatform(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	cfg := config.Default()
	cfg.Gateway.AllowedReadPaths = []string{"."}
	cfg.Gateway.AllowedWritePaths = []string{"."}
	config.Set(cfg)

	if _, err := NewGrepSearchTool(context.Background()); err == nil {
		t.Fatal("expected grep tool registration to fail on unsupported platform")
	}
	if _, err := NewReadFileTool(context.Background()); err == nil {
		t.Fatal("expected read tool registration to fail on unsupported platform")
	}
	if _, err := NewDocToMarkdownTool(context.Background()); err == nil {
		t.Fatal("expected doc tool registration to fail on unsupported platform")
	}
	if _, err := NewWriteFileTool(context.Background()); err == nil {
		t.Fatal("expected write tool registration to fail on unsupported platform")
	}
}
