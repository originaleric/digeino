package research

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
)

// FirecrawlRequest 深度爬取请求
type FirecrawlRequest struct {
	URL string `json:"url" jsonschema:"required,description=要爬取的网页URL"`
}

// FirecrawlResponse 深度爬取响应
type FirecrawlResponse struct {
	Markdown string `json:"markdown" jsonschema:"description=爬取并转换后的 Markdown 内容"`
	Success  bool   `json:"success"`
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
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &FirecrawlResponse{
		Markdown: result.Data.Markdown,
		Success:  result.Success,
	}, nil
}

func NewFirecrawlTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("firecrawl_scrape", "深度爬取指定的网页 URL，并将其内容转换为干净的 Markdown 格式", FirecrawlScrape)
}
