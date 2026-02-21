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

// BingProvider Bing 搜索提供者
type BingProvider struct {
	config     map[string]interface{}
	apiKey     string
	httpClient *http.Client
}

// NewBingProvider 创建 Bing 搜索提供者
func NewBingProvider(config map[string]interface{}) (*BingProvider, error) {
	apiKey, ok := config["BingApiKey"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("BingApiKey is required for Bing provider")
	}

	timeout := 30 * time.Second
	if t, ok := config["BingTimeout"].(int); ok && t > 0 {
		timeout = time.Duration(t) * time.Second
	}

	return &BingProvider{
		config: config,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Name 返回提供者名称
func (p *BingProvider) Name() string {
	return "bing"
}

// Search 执行搜索
func (p *BingProvider) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// 构建搜索 URL
	baseURL := "https://api.bing.microsoft.com/v7.0/search"
	params := url.Values{}
	params.Add("q", req.Query)
	params.Add("count", fmt.Sprintf("%d", p.getMaxResults(req)))

	if req.Region != "" {
		params.Add("mkt", req.Region)
	}

	// 添加 SafeSearch 配置
	if safeSearch, ok := p.config["BingSafeSearch"].(string); ok && safeSearch != "" {
		params.Add("safeSearch", safeSearch)
	}

	searchURL := baseURL + "?" + params.Encode()

	// 创建请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Ocp-Apim-Subscription-Key", p.apiKey)

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
	var bingResp struct {
		WebPages struct {
			Value []struct {
				Name    string `json:"name"`
				URL     string `json:"url"`
				Snippet string `json:"snippet"`
			} `json:"value"`
		} `json:"webPages"`
	}

	if err := json.Unmarshal(body, &bingResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 转换为统一格式
	results := make([]*SearchResult, 0, len(bingResp.WebPages.Value))
	for _, item := range bingResp.WebPages.Value {
		results = append(results, &SearchResult{
			Title:       item.Name,
			URL:         item.URL,
			Description: item.Snippet,
			Source:      "Bing",
		})
	}

	return &SearchResponse{
		Results: results,
	}, nil
}

// getMaxResults 获取最大结果数
func (p *BingProvider) getMaxResults(req *SearchRequest) int {
	if req.MaxResults > 0 {
		// Bing API 限制最多 50 个结果
		if req.MaxResults > 50 {
			return 50
		}
		return req.MaxResults
	}
	if maxResults, ok := p.config["MaxResults"].(int); ok && maxResults > 0 {
		if maxResults > 50 {
			return 50
		}
		return maxResults
	}
	return 10 // 默认值
}
