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

// GoogleProvider Google 搜索提供者
type GoogleProvider struct {
	config         map[string]interface{}
	apiKey         string
	searchEngineID string
	httpClient     *http.Client
}

// NewGoogleProvider 创建 Google 搜索提供者
func NewGoogleProvider(config map[string]interface{}) (*GoogleProvider, error) {
	apiKey, ok := config["GoogleApiKey"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("GoogleApiKey is required for Google provider")
	}

	searchEngineID, ok := config["GoogleSearchEngineId"].(string)
	if !ok || searchEngineID == "" {
		return nil, fmt.Errorf("GoogleSearchEngineId is required for Google provider")
	}

	return &GoogleProvider{
		config:         config,
		apiKey:         apiKey,
		searchEngineID: searchEngineID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Name 返回提供者名称
func (p *GoogleProvider) Name() string {
	return "google"
}

// Search 执行搜索
func (p *GoogleProvider) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// 构建搜索 URL
	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Add("key", p.apiKey)
	params.Add("cx", p.searchEngineID)
	params.Add("q", req.Query)
	params.Add("num", fmt.Sprintf("%d", p.getMaxResults(req)))

	if req.Region != "" {
		params.Add("gl", req.Region)
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
	var googleResp struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
			Pagemap struct {
				Metatags []map[string]string `json:"metatags"`
			} `json:"pagemap"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 转换为统一格式
	results := make([]*SearchResult, 0, len(googleResp.Items))
	for _, item := range googleResp.Items {
		description := item.Snippet

		// 尝试从 Pagemap 获取更详细的描述
		if len(item.Pagemap.Metatags) > 0 {
			if desc, ok := item.Pagemap.Metatags[0]["description"]; ok && desc != "" {
				description = desc
			} else if ogDesc, ok := item.Pagemap.Metatags[0]["og:description"]; ok && ogDesc != "" {
				description = ogDesc
			}
		}

		results = append(results, &SearchResult{
			Title:       item.Title,
			URL:         item.Link,
			Description: description,
			Source:      "Google",
		})
	}

	return &SearchResponse{
		Results: results,
	}, nil
}

// getMaxResults 获取最大结果数
func (p *GoogleProvider) getMaxResults(req *SearchRequest) int {
	if req.MaxResults > 0 {
		// Google API 限制最多 10 个结果
		if req.MaxResults > 10 {
			return 10
		}
		return req.MaxResults
	}
	if maxResults, ok := p.config["MaxResults"].(int); ok && maxResults > 0 {
		if maxResults > 10 {
			return 10
		}
		return maxResults
	}
	return 10 // 默认值
}
