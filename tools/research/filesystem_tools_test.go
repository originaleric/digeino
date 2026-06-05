//go:build unix

package research

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/originaleric/digeino/config"
)

func TestFilesystemToolsRequireAllowedPaths(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	config.Set(config.Default())

	if _, err := NewReadFileTool(context.Background()); err == nil {
		t.Fatal("expected read tool to be disabled without read paths")
	}
	if _, err := NewWriteFileTool(context.Background()); err == nil {
		t.Fatal("expected write tool to be disabled without write paths")
	}
}

func TestReadFileRequiresAllowedReadPath(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	dir := t.TempDir()
	allowedFile := filepath.Join(dir, "allowed.txt")
	if err := os.WriteFile(allowedFile, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	blockedFile := filepath.Join(t.TempDir(), "blocked.txt")
	if err := os.WriteFile(blockedFile, []byte("blocked"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Gateway.AllowedReadPaths = []string{dir}
	config.Set(cfg)

	if _, err := NewReadFileTool(context.Background()); err != nil {
		t.Fatalf("expected read tool registration: %v", err)
	}
	resp, err := ReadFile(context.Background(), &ReadFileRequest{Path: allowedFile})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "ok" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if _, err := ReadFile(context.Background(), &ReadFileRequest{Path: blockedFile}); err == nil {
		t.Fatal("expected blocked path to fail")
	}
}

func TestReadFileRejectsSymlinkEscape(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	cfg := config.Default()
	cfg.Gateway.AllowedReadPaths = []string{dir}
	config.Set(cfg)

	if _, err := ReadFile(context.Background(), &ReadFileRequest{Path: link}); err == nil {
		t.Fatal("expected symlink escape read to fail")
	}
}

func TestReadFileRejectsSymlinkDirectoryEscape(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	dir := t.TempDir()
	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(dir, "link-dir")
	if err := os.Symlink(outsideDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	cfg := config.Default()
	cfg.Gateway.AllowedReadPaths = []string{dir}
	config.Set(cfg)

	if _, err := ReadFile(context.Background(), &ReadFileRequest{Path: filepath.Join(linkDir, "secret.txt")}); err == nil {
		t.Fatal("expected symlink directory escape read to fail")
	}
}

func TestGrepSearchRejectsSymlinkEscape(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("needle"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	cfg := config.Default()
	cfg.Gateway.AllowedReadPaths = []string{dir}
	config.Set(cfg)

	resp, err := GrepSearch(context.Background(), &GrepRequest{Query: "needle", Path: dir})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count != 0 {
		t.Fatalf("expected symlink escape to be skipped, got %+v", resp)
	}
}

func TestWriteFileRequiresAllowedWritePath(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	dir := t.TempDir()
	cfg := config.Default()
	cfg.Gateway.AllowedWritePaths = []string{dir}
	config.Set(cfg)

	if _, err := NewWriteFileTool(context.Background()); err != nil {
		t.Fatalf("expected write tool registration: %v", err)
	}
	target := filepath.Join(dir, "out.txt")
	if _, err := WriteFile(context.Background(), &WriteFileRequest{Path: target, Content: "ok"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "ok" {
		t.Fatalf("unexpected content: %q", data)
	}
	if _, err := WriteFile(context.Background(), &WriteFileRequest{Path: filepath.Join(t.TempDir(), "out.txt"), Content: "blocked"}); err == nil {
		t.Fatal("expected blocked write path to fail")
	}
}

func TestWriteFileCreatesMissingSubdirectoriesUnderAllowedPath(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	dir := t.TempDir()
	cfg := config.Default()
	cfg.Gateway.AllowedWritePaths = []string{dir}
	config.Set(cfg)

	target := filepath.Join(dir, "new-dir", "nested", "out.txt")
	if _, err := WriteFile(context.Background(), &WriteFileRequest{Path: target, Content: "ok"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "ok" {
		t.Fatalf("unexpected content: %q", data)
	}
}

func TestWriteFileRejectsSymlinkEscape(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	cfg := config.Default()
	cfg.Gateway.AllowedWritePaths = []string{dir}
	config.Set(cfg)

	if _, err := WriteFile(context.Background(), &WriteFileRequest{Path: link, Content: "changed"}); err == nil {
		t.Fatal("expected symlink escape write to fail")
	}
	data, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("outside file was modified: %q", data)
	}
}

func TestWriteFileRejectsSymlinkDirectoryEscape(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	dir := t.TempDir()
	outsideDir := t.TempDir()
	linkDir := filepath.Join(dir, "link-dir")
	if err := os.Symlink(outsideDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	cfg := config.Default()
	cfg.Gateway.AllowedWritePaths = []string{dir}
	config.Set(cfg)

	if _, err := WriteFile(context.Background(), &WriteFileRequest{Path: filepath.Join(linkDir, "out.txt"), Content: "changed"}); err == nil {
		t.Fatal("expected symlink directory escape write to fail")
	}
	if _, err := os.Stat(filepath.Join(outsideDir, "out.txt")); !os.IsNotExist(err) {
		t.Fatalf("outside file should not exist, stat err=%v", err)
	}
}
