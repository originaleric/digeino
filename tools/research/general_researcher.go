package research

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
)

type pathAccessMode string

const (
	pathAccessRead  pathAccessMode = "read"
	pathAccessWrite pathAccessMode = "write"
)

func allowedPathsFor(mode pathAccessMode) []string {
	cfg := config.Get()
	switch mode {
	case pathAccessWrite:
		return cfg.Gateway.AllowedWritePaths
	default:
		return cfg.Gateway.AllowedReadPaths
	}
}

func validateConfiguredPath(path string, mode pathAccessMode) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("文件路径不能为空")
	}
	allowed := allowedPathsFor(mode)
	if len(allowed) == 0 {
		return "", fmt.Errorf("%s path access is disabled without configured allowed paths", mode)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("无法解析文件路径: %w", err)
	}
	clean, err := resolvedPathForAccess(abs, mode)
	if err != nil {
		return "", err
	}
	for _, prefix := range allowed {
		base, ok := resolvedAllowedBase(prefix)
		if !ok {
			continue
		}
		if clean == base || strings.HasPrefix(clean, base+string(os.PathSeparator)) {
			return clean, nil
		}
	}
	return "", fmt.Errorf("path %q is not under configured allowed %s paths", path, mode)
}

func resolvedAllowedBase(prefix string) (string, bool) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "", false
	}
	base, err := filepath.Abs(prefix)
	if err != nil {
		return "", false
	}
	base, err = filepath.EvalSymlinks(base)
	if err != nil {
		return "", false
	}
	return filepath.Clean(base), true
}

func resolvedPathForAccess(absPath string, mode pathAccessMode) (string, error) {
	clean := filepath.Clean(absPath)
	resolved, err := filepath.EvalSymlinks(clean)
	if err == nil {
		return filepath.Clean(resolved), nil
	}
	if mode != pathAccessWrite || !os.IsNotExist(err) {
		return "", fmt.Errorf("无法解析真实文件路径: %w", err)
	}
	parent := filepath.Dir(clean)
	existing, missing, parentErr := nearestExistingAncestor(parent)
	if parentErr != nil {
		return "", parentErr
	}
	resolvedParent, parentErr := filepath.EvalSymlinks(existing)
	if parentErr != nil {
		return "", fmt.Errorf("无法解析真实父目录: %w", parentErr)
	}
	parts := append(missing, filepath.Base(clean))
	resolvedPath := filepath.Clean(resolvedParent)
	for _, part := range parts {
		resolvedPath = filepath.Join(resolvedPath, part)
	}
	return resolvedPath, nil
}

func nearestExistingAncestor(path string) (existing string, missing []string, err error) {
	clean := filepath.Clean(path)
	for {
		if _, statErr := os.Stat(clean); statErr == nil {
			return clean, missing, nil
		} else if !os.IsNotExist(statErr) {
			return "", nil, fmt.Errorf("无法检查真实父目录: %w", statErr)
		}
		parent := filepath.Dir(clean)
		if parent == clean {
			return "", nil, fmt.Errorf("无法解析真实父目录: no existing ancestor for %q", path)
		}
		missing = append([]string{filepath.Base(clean)}, missing...)
		clean = parent
	}
}

func readAllowedFile(path string) ([]byte, error) {
	file, err := openAllowedReadFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func allowedTarget(path string, mode pathAccessMode) (base string, rel string, err error) {
	resolved, err := validateConfiguredPath(path, mode)
	if err != nil {
		return "", "", err
	}
	for _, prefix := range allowedPathsFor(mode) {
		candidate, ok := resolvedAllowedBase(prefix)
		if !ok {
			continue
		}
		if resolved == candidate || strings.HasPrefix(resolved, candidate+string(os.PathSeparator)) {
			rel, err := filepath.Rel(candidate, resolved)
			if err != nil {
				return "", "", err
			}
			if rel == "." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." {
				continue
			}
			return candidate, rel, nil
		}
	}
	return "", "", fmt.Errorf("path %q is not under configured allowed %s paths", path, mode)
}

// --- Grep Search Tool ---

type GrepRequest struct {
	Query string `json:"query" jsonschema:"required,description=搜索关键词"`
	Path  string `json:"path" jsonschema:"description=搜索路径，默认为项目根目录"`
}

type GrepResponse struct {
	Matches []string `json:"matches"`
	Count   int      `json:"count"`
}

// GrepSearch 模拟代码搜索逻辑 (考虑到云端环境，这里使用简单的文件系统遍历，未来可接入 ripgrep)
func GrepSearch(ctx context.Context, req *GrepRequest) (*GrepResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	searchPath := req.Path
	if searchPath == "" {
		// 默认搜索当前目录
		searchPath = "."
	}
	searchPath, err := validateConfiguredPath(searchPath, pathAccessRead)
	if err != nil {
		return nil, err
	}

	var matches []string
	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误文件夹
		}
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// 简单起见，只检查文本类的文件后缀
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".md" && ext != ".yml" && ext != ".json" && ext != ".txt" {
			return nil
		}

		checkedPath, err := validateConfiguredPath(path, pathAccessRead)
		if err != nil {
			return nil
		}
		content, err := readAllowedFile(checkedPath)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), req.Query) {
			matches = append(matches, path)
		}

		if len(matches) > 20 {
			return filepath.SkipAll // 限制前20个匹配，防止 Token 爆炸
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &GrepResponse{Matches: matches, Count: len(matches)}, nil
}

// --- Read File Tool ---

type ReadFileRequest struct {
	Path string `json:"path" jsonschema:"required,description=要阅读的文件路径"`
}

type ReadFileResponse struct {
	Content string `json:"content"`
}

func ReadFile(ctx context.Context, req *ReadFileRequest) (*ReadFileResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	path, err := validateConfiguredPath(req.Path, pathAccessRead)
	if err != nil {
		return nil, err
	}
	content, err := readAllowedFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	return &ReadFileResponse{Content: string(content)}, nil
}

// --- Doc to Markdown Tool (Placeholder) ---

type DocToMarkdownRequest struct {
	Path string `json:"path" jsonschema:"required,description=PDF或文档路径"`
}

type DocToMarkdownResponse struct {
	Markdown string `json:"markdown"`
}

func DocToMarkdown(ctx context.Context, req *DocToMarkdownRequest) (*DocToMarkdownResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	path, err := validateConfiguredPath(req.Path, pathAccessRead)
	if err != nil {
		return nil, err
	}
	cfg := config.Get()
	toolCfg := cfg.Tools.Unstructured

	apiKey := toolCfg.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("UNSTRUCTURED_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("Unstructured API Key 未配置")
	}

	baseUrl := toolCfg.BaseUrl
	if baseUrl == "" {
		baseUrl = "https://api.unstructured.io/general/v0/general"
	}

	// 1. 打开文件
	file, err := openAllowedReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("无法打开文档: %w", err)
	}
	defer file.Close()

	// 2. 准备 multipart 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	// 添加额外参数：策略设置为 'fast' 或 'hi_res'，输出格式为 'text/markdown'
	_ = writer.WriteField("output_format", "text/markdown")
	_ = writer.WriteField("strategy", "fast")
	writer.Close()

	// 3. 执行请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseUrl, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("unstructured-api-key", apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Unstructured API 错误: %d, %s", resp.StatusCode, string(respBody))
	}

	// 4. 解析结果
	markdown, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &DocToMarkdownResponse{
		Markdown: string(markdown),
	}, nil
}

// --- Factory Functions ---

func NewGrepSearchTool(ctx context.Context) (tool.BaseTool, error) {
	if !secureFileAccessSupported() {
		return nil, fmt.Errorf("research_grep tool is disabled because secure file access is not supported on this platform")
	}
	if len(allowedPathsFor(pathAccessRead)) == 0 {
		return nil, fmt.Errorf("research_grep tool is disabled without Gateway.AllowedReadPaths")
	}
	return utils.InferTool("research_grep", "在源码或选定目录中全文检索关键词", GrepSearch)
}

func NewReadFileTool(ctx context.Context) (tool.BaseTool, error) {
	if !secureFileAccessSupported() {
		return nil, fmt.Errorf("research_read tool is disabled because secure file access is not supported on this platform")
	}
	if len(allowedPathsFor(pathAccessRead)) == 0 {
		return nil, fmt.Errorf("research_read tool is disabled without Gateway.AllowedReadPaths")
	}
	return utils.InferTool("research_read", "精确阅读指定文件的内容", ReadFile)
}

func NewDocToMarkdownTool(ctx context.Context) (tool.BaseTool, error) {
	if !secureFileAccessSupported() {
		return nil, fmt.Errorf("research_doc_to_md tool is disabled because secure file access is not supported on this platform")
	}
	if len(allowedPathsFor(pathAccessRead)) == 0 {
		return nil, fmt.Errorf("research_doc_to_md tool is disabled without Gateway.AllowedReadPaths")
	}
	return utils.InferTool("research_doc_to_md", "将 PDF 或 Word 文档解析为 Markdown 文本", DocToMarkdown)
}
