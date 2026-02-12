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
	searchPath := req.Path
	if searchPath == "" {
		// 默认搜索当前目录
		searchPath = "."
	}

	var matches []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
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

		content, err := os.ReadFile(path)
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
	content, err := os.ReadFile(req.Path)
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
	file, err := os.Open(req.Path)
	if err != nil {
		return nil, fmt.Errorf("无法打开文档: %w", err)
	}
	defer file.Close()

	// 2. 准备 multipart 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filepath.Base(req.Path))
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
	return utils.InferTool("research_grep", "在源码或选定目录中全文检索关键词", GrepSearch)
}

func NewReadFileTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("research_read", "精确阅读指定文件的内容", ReadFile)
}

func NewDocToMarkdownTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("research_doc_to_md", "将 PDF 或 Word 文档解析为 Markdown 文本", DocToMarkdown)
}
