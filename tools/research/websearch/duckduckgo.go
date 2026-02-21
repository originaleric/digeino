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

// DuckDuckGoProvider DuckDuckGo 搜索提供者
type DuckDuckGoProvider struct {
	config     map[string]interface{}
	httpClient *http.Client
}

// NewDuckDuckGoProvider 创建 DuckDuckGo 搜索提供者
func NewDuckDuckGoProvider(config map[string]interface{}) *DuckDuckGoProvider {
	return &DuckDuckGoProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name 返回提供者名称
func (p *DuckDuckGoProvider) Name() string {
	return "duckduckgo"
}

// Search 执行搜索
func (p *DuckDuckGoProvider) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// 构建搜索 URL
	searchURL := "https://api.duckduckgo.com/"
	params := url.Values{}
	params.Add("q", req.Query)
	params.Add("format", "json")
	params.Add("no_html", "1")
	params.Add("skip_disambig", "1")

	fullURL := searchURL + "?" + params.Encode()

	// 创建请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置 User-Agent
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; DigFlow/1.0)")

	// 执行请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析响应
	var ddgResp struct {
		AbstractText   string `json:"AbstractText"`
		AbstractSource string `json:"AbstractSource"`
		AbstractURL    string `json:"AbstractURL"`
		RelatedTopics  []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.Unmarshal(body, &ddgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 转换为统一格式
	results := make([]*SearchResult, 0)

	// 添加主要结果
	if ddgResp.AbstractText != "" {
		results = append(results, &SearchResult{
			Title:       ddgResp.AbstractSource,
			URL:         ddgResp.AbstractURL,
			Description: ddgResp.AbstractText,
			Source:      "DuckDuckGo",
		})
	}

	// 添加相关主题
	maxResults := p.getMaxResults(req)
	for i, topic := range ddgResp.RelatedTopics {
		if len(results) >= maxResults {
			break
		}
		if topic.Text != "" && topic.FirstURL != "" {
			results = append(results, &SearchResult{
				Title:       topic.Text,
				URL:         topic.FirstURL,
				Description: topic.Text,
				Source:      "DuckDuckGo",
			})
		}
		if i >= maxResults-1 {
			break
		}
	}

	return &SearchResponse{
		Results: results,
	}, nil
}

// getMaxResults 获取最大结果数
func (p *DuckDuckGoProvider) getMaxResults(req *SearchRequest) int {
	if req.MaxResults > 0 {
		return req.MaxResults
	}
	if maxResults, ok := p.config["MaxResults"].(int); ok && maxResults > 0 {
		return maxResults
	}
	return 10 // 默认值
}
