package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SerpAPIProvider SerpAPI 搜索提供者
// SerpAPI 是 Google API 的替代方案，提供对多个搜索引擎的访问
type SerpAPIProvider struct {
	config     map[string]interface{}
	apiKey     string
	engine     string // google, bing, baidu 等
	httpClient *http.Client
}

// NewSerpAPIProvider 创建 SerpAPI 搜索提供者
func NewSerpAPIProvider(config map[string]interface{}) (*SerpAPIProvider, error) {
	apiKey, ok := config["SerpAPIKey"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("SerpAPIKey is required for SerpAPI provider")
	}

	// 默认使用 Google 引擎
	engine := "google"
	if eng, ok := config["SerpAPIEngine"].(string); ok && eng != "" {
		engine = eng
	}

	return &SerpAPIProvider{
		config: config,
		apiKey: apiKey,
		engine: engine,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Name 返回提供者名称
func (p *SerpAPIProvider) Name() string {
	return "serpapi"
}

// Search 执行搜索
func (p *SerpAPIProvider) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// 构建搜索 URL
	baseURL := "https://serpapi.com/search"
	params := url.Values{}
	params.Add("api_key", p.apiKey)
	params.Add("engine", p.engine)
	params.Add("q", req.Query)
	params.Add("num", fmt.Sprintf("%d", p.getMaxResults(req)))

	// 添加地区参数
	if req.Region != "" {
		params.Add("gl", req.Region) // gl = Google Location
		params.Add("hl", req.Region) // hl = Host Language
	}

	searchURL := baseURL + "?" + params.Encode()

	// 创建请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 执行请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析响应
	var serpResp struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body, &serpResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查 API 错误
	if serpResp.Error != "" {
		return nil, fmt.Errorf("SerpAPI error: %s", serpResp.Error)
	}

	// 转换为统一格式
	results := make([]*SearchResult, 0, len(serpResp.OrganicResults))
	for _, item := range serpResp.OrganicResults {
		results = append(results, &SearchResult{
			Title:       item.Title,
			URL:         item.Link,
			Description: item.Snippet,
			Source:      fmt.Sprintf("SerpAPI (%s)", p.engine),
		})
	}

	return &SearchResponse{
		Results: results,
	}, nil
}

// getMaxResults 获取最大结果数
func (p *SerpAPIProvider) getMaxResults(req *SearchRequest) int {
	if req.MaxResults > 0 {
		// SerpAPI 限制最多 100 个结果
		if req.MaxResults > 100 {
			return 100
		}
		return req.MaxResults
	}
	if maxResults, ok := p.config["MaxResults"].(int); ok && maxResults > 0 {
		if maxResults > 100 {
			return 100
		}
		return maxResults
	}
	return 10 // 默认值
}
