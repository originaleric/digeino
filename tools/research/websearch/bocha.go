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

// BochaProvider 博查搜索提供者
type BochaProvider struct {
	config     map[string]interface{}
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewBochaProvider 创建博查搜索提供者
func NewBochaProvider(config map[string]interface{}) (*BochaProvider, error) {
	apiKey, ok := config["BochaApiKey"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("BochaApiKey is required for Bocha provider")
	}

	baseURL := "https://api.bochaai.com"
	if url, ok := config["BochaBaseUrl"].(string); ok && url != "" {
		baseURL = url
	}

	return &BochaProvider{
		config:  config,
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Name 返回提供者名称
func (p *BochaProvider) Name() string {
	return "bocha"
}

// Search 执行搜索
func (p *BochaProvider) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// 构建请求 URL
	// 注意：正确的端点是 /v1/web-search（连字符），不是 /v1/web/search（斜杠）
	searchURL := fmt.Sprintf("%s/v1/web-search", p.baseURL)

	// 构建请求体
	// 根据 bocha API 文档：https://open.bochaai.com/
	// 请求体应使用 "query" 而不是 "q"
	// 官方示例包含 summary 和 freshness 参数
	requestBody := map[string]interface{}{
		"query":   req.Query,
		"count":   p.getMaxResults(req),
		"summary": true, // 请求返回详细摘要
	}

	if req.Region != "" {
		requestBody["region"] = req.Region
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", searchURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

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

	// 读取响应字符串（用于错误日志）
	bodyStr := string(body)

	// 解析响应
	// 实际响应格式为：
	// {
	//   "code": 200,
	//   "log_id": "...",
	//   "msg": null,
	//   "data": {
	//     "_type": "SearchResponse",
	//     "webPages": {
	//       "value": [
	//         {
	//           "name": "标题",
	//           "url": "URL",
	//           "snippet": "摘要",
	//           "summary": "详细摘要"
	//         }
	//       ]
	//     }
	//   }
	// }
	var bochaResp struct {
		Code  int    `json:"code"`
		LogID string `json:"log_id"`
		Msg   string `json:"msg"`
		Data  struct {
			Type     string `json:"_type"`
			WebPages struct {
				Value []struct {
					Name    string `json:"name"`
					URL     string `json:"url"`
					Snippet string `json:"snippet"`
					Summary string `json:"summary"`
				} `json:"value"`
			} `json:"webPages"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &bochaResp); err != nil {
		errorPreview := bodyStr
		if len(bodyStr) > 200 {
			errorPreview = bodyStr[:200]
		}
		fmt.Printf("解析 Bocha API 响应失败 %v, response_preview: %s\n", err, errorPreview)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("Bocha API 响应解析成功, code: %d, webPages_count: %d\n", bochaResp.Code, len(bochaResp.Data.WebPages.Value))

	// 检查响应状态码
	if bochaResp.Code != 200 {
		return nil, fmt.Errorf("Bocha API returned error code %d: %s", bochaResp.Code, bochaResp.Msg)
	}

	// 转换为统一格式
	results := make([]*SearchResult, 0, len(bochaResp.Data.WebPages.Value))
	for _, item := range bochaResp.Data.WebPages.Value {
		// 优先使用 summary，如果没有则使用 snippet
		description := item.Summary
		if description == "" {
			description = item.Snippet
		}
		results = append(results, &SearchResult{
			Title:       item.Name,
			URL:         item.URL,
			Description: description,
			Source:      "Bocha",
		})
	}

	return &SearchResponse{
		Results: results,
	}, nil
}

// getMaxResults 获取最大结果数
func (p *BochaProvider) getMaxResults(req *SearchRequest) int {
	if req.MaxResults > 0 {
		return req.MaxResults
	}
	if maxResults, ok := p.config["MaxResults"].(int); ok && maxResults > 0 {
		return maxResults
	}
	return 10 // 默认值
}
