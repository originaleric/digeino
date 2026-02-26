package research

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
)

// FirecrawlRequest 深度爬取请求
type FirecrawlRequest struct {
	URL string `json:"url" jsonschema:"required,description=要爬取的网页URL"`
}

// Link 从 Markdown 中提取的链接
type Link struct {
	URL  string `json:"url" jsonschema:"description=链接地址"`
	Text string `json:"text" jsonschema:"description=链接描述文本"`
}

// FirecrawlResponse 深度爬取响应
type FirecrawlResponse struct {
	Markdown string `json:"markdown" jsonschema:"description=爬取并转换后的 Markdown 内容"`
	Links    []Link `json:"links" jsonschema:"description=从 Markdown 中提取的相关链接，便于 Agent 进行 URL 跟进爬取"`
	Success  bool   `json:"success"`
}

// markdownLinkRegex 匹配 [text](url) 格式
var markdownLinkRegex = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

// extractLinksFromMarkdown 从 Markdown 中提取 [text](url) 格式的链接，过滤图片、锚点等
func extractLinksFromMarkdown(md string) []Link {
	seen := make(map[string]bool)
	var links []Link

	matches := markdownLinkRegex.FindAllStringSubmatch(md, -1)
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		text, linkURL := strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
		if linkURL == "" {
			continue
		}
		// 过滤：锚点、data URL、常见图片格式
		if strings.HasPrefix(linkURL, "#") ||
			strings.HasPrefix(linkURL, "data:") ||
			strings.HasPrefix(linkURL, "mailto:") ||
			strings.HasPrefix(linkURL, "javascript:") {
			continue
		}
		lower := strings.ToLower(linkURL)
		if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") ||
			strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".gif") ||
			strings.HasSuffix(lower, ".webp") || strings.HasSuffix(lower, ".svg") ||
			strings.Contains(lower, ".png?") || strings.Contains(lower, ".jpg?") {
			continue
		}
		// 校验为有效 HTTP(S) URL
		if parsed, err := url.Parse(linkURL); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			continue
		}
		if seen[linkURL] {
			continue
		}
		seen[linkURL] = true
		links = append(links, Link{URL: linkURL, Text: text})
	}
	return links
}

// FirecrawlScrape 调用 Firecrawl API 将网页转换为 Markdown
func FirecrawlScrape(ctx context.Context, req *FirecrawlRequest) (*FirecrawlResponse, error) {
	cfg := config.Get()
	apiKey := cfg.Tools.Firecrawl.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("FIRECRAWL_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("Firecrawl API Key 未配置，可通过 eino.yml 或 FIRECRAWL_API_KEY 环境变量设置")
	}

	apiUrl := "https://api.firecrawl.dev/v0/scrape"
	payload := map[string]interface{}{
		"url": req.URL,
		"pageOptions": map[string]interface{}{
			"onlyMainContent": true,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("firecrawl 接口返回错误: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Markdown string `json:"markdown"`
			Content  string `json:"content"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	markdown := result.Data.Markdown
	if markdown == "" {
		markdown = result.Data.Content
	}
	links := extractLinksFromMarkdown(markdown)

	return &FirecrawlResponse{
		Markdown: markdown,
		Links:    links,
		Success:  result.Success,
	}, nil
}

func NewFirecrawlTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("firecrawl_scrape", "深度爬取指定的网页 URL，返回 Markdown 内容及从中提取的 links 列表（便于 Agent 进行 URL 跟进爬取）", FirecrawlScrape)
}
