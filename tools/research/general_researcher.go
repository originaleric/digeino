package research

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
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
	// 实际应用中会接入 github.com/ledongthuc/pdf 或类似库
	// 暂时返回 Mock 消息说明功能意图
	return &DocToMarkdownResponse{
		Markdown: fmt.Sprintf("# 文档预览: %s\n\n[此处将来会是转换后的 Markdown 内容]", req.Path),
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
