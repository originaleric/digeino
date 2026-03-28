package tempstorage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/originaleric/digeino/config"
)

// Context 键约定（与 DigFlow 兼容）
const (
	ContextKeyWorkspacePath   = "workspace_path"
	ContextKeyAgentSessionID  = "agent_session_id"
)

// SaveForReview 将内容写入临时文件，返回绝对路径供 audit/critique 使用。
// ctx 需包含以下之一：
//   - workspace_path (string): 写入 workspace_path/filename
//   - agent_session_id (string): 写入 {BaseDir}/agent/{session_id}/{filename}
func SaveForReview(ctx context.Context, filename, content string) (path string, err error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}
	if err := validateFilename(filename); err != nil {
		return "", err
	}

	baseDir, err := resolveBaseDir(ctx)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(baseDir, filename)
	dir = filepath.Dir(dir)
	fullPath := filepath.Join(baseDir, filename)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}
	return absPathOf(fullPath)
}

// SaveBytesForReview 将二进制内容写入临时/工作空间文件（如 zip），返回绝对路径。
func SaveBytesForReview(ctx context.Context, filename string, content []byte) (path string, err error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}
	if err := validateFilename(filename); err != nil {
		return "", err
	}

	baseDir, err := resolveBaseDir(ctx)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(baseDir, filename)
	dir = filepath.Dir(dir)
	fullPath := filepath.Join(baseDir, filename)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}
	return absPathOf(fullPath)
}

func absPathOf(fullPath string) (string, error) {
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fullPath, nil
	}
	return absPath, nil
}

// ValidateRelativePath 校验相对路径（与 SaveForReview 规则一致），供其他包复用。
func ValidateRelativePath(filename string) error {
	return validateFilename(filename)
}

// GetBasePath 从 context 解析基础目录绝对路径。
// 若有 workspace_path 则返回该路径；若有 agent_session_id 则返回 {BaseDir}/agent/{session_id}。
func GetBasePath(ctx context.Context) (string, error) {
	return resolveBaseDir(ctx)
}

func resolveBaseDir(ctx context.Context) (string, error) {
	if wp, ok := ctx.Value(ContextKeyWorkspacePath).(string); ok && wp != "" {
		abs, err := filepath.Abs(wp)
		if err != nil {
			return "", fmt.Errorf("resolve workspace_path: %w", err)
		}
		return abs, nil
	}
	if sid, ok := ctx.Value(ContextKeyAgentSessionID).(string); ok && sid != "" {
		baseDir := getTempBaseDir()
		dir := filepath.Join(baseDir, "agent", sid)
		abs, err := filepath.Abs(dir)
		if err != nil {
			return "", fmt.Errorf("resolve session dir: %w", err)
		}
		return abs, nil
	}
	return "", fmt.Errorf("context must contain %q or %q", ContextKeyWorkspacePath, ContextKeyAgentSessionID)
}

func getTempBaseDir() string {
	baseDir := "storage/temp"
	if cfg := config.Get(); cfg != nil && cfg.UIUX.TempStorage.BaseDir != "" {
		baseDir = cfg.UIUX.TempStorage.BaseDir
	}
	return baseDir
}

func validateFilename(filename string) error {
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename cannot contain '..'")
	}
	if filepath.IsAbs(filename) || strings.HasPrefix(filename, "/") {
		return fmt.Errorf("filename must be relative")
	}
	if strings.Contains(filename, string(os.PathSeparator)) && os.PathSeparator != '/' {
		// Allow forward slash for subdirs
	}
	// Reject path traversal via cleaned path
	cleaned := filepath.Clean(filename)
	if strings.HasPrefix(cleaned, "..") {
		return fmt.Errorf("filename cannot escape base directory")
	}
	return nil
}
