package websearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TavilyProvider 使用 Tavily Search API 作为搜索引擎
type TavilyProvider struct {
	apiKey      string
	baseURL     string
	searchDepth string
	topic       string
	httpClient  *http.Client
}

// NewTavilyProvider 创建 Tavily 搜索提供者
// 期望从 config 中读取:
// - TavilyApiKey: string
// - TavilyBaseUrl: string (可选，默认为 https://api.tavily.com)
// - TavilySearchDepth: string (可选，basic/fast/advanced/ultra-fast，默认 basic)
// - TavilyTopic: string (可选，general/news/finance，默认 general)
func NewTavilyProvider(config map[string]interface{}) (*TavilyProvider, error) {
	rawKey, ok := config["TavilyApiKey"]
	if !ok {
		return nil, fmt.Errorf("TavilyApiKey is required for Tavily provider")
	}
	apiKey, ok := rawKey.(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("TavilyApiKey must be a non-empty string")
	}

	baseURL := "https://api.tavily.com"
	if v, ok := config["TavilyBaseUrl"].(string); ok && v != "" {
		baseURL = v
	}

	searchDepth := "basic"
	if v, ok := config["TavilySearchDepth"].(string); ok && v != "" {
		searchDepth = v
	}

	topic := "general"
	if v, ok := config["TavilyTopic"].(string); ok && v != "" {
		topic = v
	}

	return &TavilyProvider{
		apiKey:      apiKey,
		baseURL:     baseURL,
		searchDepth: searchDepth,
		topic:       topic,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (p *TavilyProvider) Name() string {
	return "tavily"
}

// Search 调用 Tavily /search 接口执行搜索
func (p *TavilyProvider) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	searchURL := fmt.Sprintf("%s/search", p.baseURL)

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 20 {
		maxResults = 20
	}

	payload := map[string]interface{}{
		"query":         req.Query,
		"max_results":   maxResults,
		"search_depth":  p.searchDepth,
		"topic":         p.topic,
		"include_answer": false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tavily payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", searchURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create tavily request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tavily request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tavily response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		preview := string(respBody)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("tavily returned status %d: %s", resp.StatusCode, preview)
	}

	// Tavily 响应结构:
	// {
	//   "query": "...",
	//   "results": [
	//     { "url": "...", "title": "...", "content": "...", "score": 0.81 }
	//   ]
	// }
	var tvResp struct {
		Query   string `json:"query"`
		Results []struct {
			URL     string  `json:"url"`
			Title   string  `json:"title"`
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}

	if err := json.Unmarshal(respBody, &tvResp); err != nil {
		preview := string(respBody)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("failed to parse tavily response: %w, preview: %s", err, preview)
	}

	results := make([]*SearchResult, 0, len(tvResp.Results))
	for _, item := range tvResp.Results {
		results = append(results, &SearchResult{
			Title:       item.Title,
			URL:         item.URL,
			Description: item.Content,
			Source:      "Tavily",
		})
	}

	return &SearchResponse{
		Results: results,
	}, nil
}
