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

// FirecrawlProvider 使用 Firecrawl /search 作为搜索引擎
type FirecrawlProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewFirecrawlProvider 创建 Firecrawl 搜索提供者
// 期望从 config 中读取:
// - FirecrawlApiKey: string
// - FirecrawlBaseUrl: string (可选，默认为 https://api.firecrawl.dev/v2)
func NewFirecrawlProvider(config map[string]interface{}) (*FirecrawlProvider, error) {
	rawKey, ok := config["FirecrawlApiKey"]
	if !ok {
		return nil, fmt.Errorf("FirecrawlApiKey is required for Firecrawl provider")
	}
	apiKey, ok := rawKey.(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("FirecrawlApiKey must be a non-empty string")
	}

	baseURL := "https://api.firecrawl.dev/v2"
	if v, ok := config["FirecrawlBaseUrl"].(string); ok && v != "" {
		baseURL = v
	}

	return &FirecrawlProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (p *FirecrawlProvider) Name() string {
	return "firecrawl"
}

// Search 调用 Firecrawl /search 接口执行搜索
// 仅使用 web 源，不附带 scrapeOptions（只取 url/title/description）
func (p *FirecrawlProvider) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	searchURL := fmt.Sprintf("%s/search", p.baseURL)

	limit := req.MaxResults
	if limit <= 0 {
		limit = 10
	}

	payload := map[string]interface{}{
		"query": req.Query,
		"limit": limit,
		"sources": []map[string]string{
			{"type": "web"},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal firecrawl payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", searchURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create firecrawl request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("firecrawl request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read firecrawl response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		preview := string(respBody)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("firecrawl returned status %d: %s", resp.StatusCode, preview)
	}

	// 参考 Firecrawl v2 /search 响应结构:
	// {
	//   "success": true,
	//   "data": {
	//     "web": [
	//       { "url": "...", "title": "...", "description": "...", "position": 1 }
	//     ]
	//   }
	// }
	var fcResp struct {
		Success bool `json:"success"`
		Data    struct {
			Web []struct {
				URL         string `json:"url"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Position    int    `json:"position"`
			} `json:"web"`
		} `json:"data"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal(respBody, &fcResp); err != nil {
		preview := string(respBody)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("failed to parse firecrawl response: %w, preview: %s", err, preview)
	}

	if !fcResp.Success {
		if fcResp.Error != "" {
			return nil, fmt.Errorf("firecrawl search failed: %s", fcResp.Error)
		}
		return nil, fmt.Errorf("firecrawl search failed with unknown error")
	}

	results := make([]*SearchResult, 0, len(fcResp.Data.Web))
	for _, item := range fcResp.Data.Web {
		results = append(results, &SearchResult{
			Title:       item.Title,
			URL:         item.URL,
			Description: item.Description,
			Source:      "Firecrawl",
		})
	}

	return &SearchResponse{
		Results: results,
	}, nil
}

