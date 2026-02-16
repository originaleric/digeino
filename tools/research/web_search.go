package research

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/bingsearch"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino-ext/components/tool/googlesearch"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
)

// SearchRequest 搜索请求
type SearchRequest struct {
	Query      string `json:"query" jsonschema:"required,description=搜索关键词"`
	MaxResults int    `json:"max_results" jsonschema:"description=最大结果数，默认为10"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// WebSearch 执行网页搜索
func WebSearch(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	cfg := config.Get()
	toolCfg := cfg.Tools.WebSearch

	engine := toolCfg.Engine
	if engine == "" {
		engine = "duckduckgo" // 默认为 duckduckgo 占位
	}

	switch engine {
	case "bocha":
		return searchBocha(ctx, toolCfg.Bocha, req)
	case "serpapi":
		return searchSerpApi(ctx, toolCfg.SerpApi, req)
	case "google":
		return searchGoogle(ctx, toolCfg.Google, req)
	case "bing":
		return searchBing(ctx, toolCfg.Bing, req)
	case "duckduckgo":
		return searchDuckDuckGo(ctx, toolCfg.DuckDuckGo, req)
	default:
		return nil, fmt.Errorf("暂不支持的搜索引擎: %s，请配置 bocha, serpapi, google, bing 或 duckduckgo", engine)
	}
}

func searchBocha(ctx context.Context, cfg config.BochaConfig, req *SearchRequest) (*SearchResponse, error) {
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("bocha 搜索需要配置 ApiKey")
	}

	baseUrl := "https://api.bochaai.com"
	if cfg.BaseUrl != "" {
		baseUrl = cfg.BaseUrl
	}

	apiUrl := fmt.Sprintf("%s/v1/web/search", baseUrl)
	count := req.MaxResults
	if count <= 0 {
		count = 10
	}

	payload := map[string]interface{}{
		"q":     req.Query,
		"count": count,
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.ApiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bocha API 错误: %d, %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			WebPages struct {
				Value []struct {
					Name    string `json:"name"`
					Url     string `json:"url"`
					Snippet string `json:"snippet"`
				} `json:"value"`
			} `json:"webPages"`
		} `json:"data"`
	}

	// 兼容不同版本的 Bocha API 或结构
	// 如果直接是 results 数组
	var resultAlt struct {
		Results []struct {
			Title   string `json:"title"`
			Url     string `json:"url"`
			Snippet string `json:"snippet"`
		} `json:"results"`
	}

	allBody, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(allBody, &resultAlt); err == nil && len(resultAlt.Results) > 0 {
		var finalResults []SearchResult
		for _, r := range resultAlt.Results {
			finalResults = append(finalResults, SearchResult{
				Title:       r.Title,
				URL:         r.Url,
				Description: r.Snippet,
			})
		}
		return &SearchResponse{Results: finalResults}, nil
	}

	json.Unmarshal(allBody, &result)
	var finalResults []SearchResult
	for _, r := range result.Data.WebPages.Value {
		finalResults = append(finalResults, SearchResult{
			Title:       r.Name,
			URL:         r.Url,
			Description: r.Snippet,
		})
	}

	return &SearchResponse{Results: finalResults}, nil
}

func searchSerpApi(ctx context.Context, cfg config.SerpApiConfig, req *SearchRequest) (*SearchResponse, error) {
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("serpapi 搜索需要配置 ApiKey")
	}

	apiUrl := "https://serpapi.com/search"
	if cfg.BaseUrl != "" {
		apiUrl = cfg.BaseUrl
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	q := httpReq.URL.Query()
	q.Add("q", req.Query)
	q.Add("api_key", cfg.ApiKey)
	q.Add("engine", "google")
	httpReq.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var finalResults []SearchResult
	for _, r := range result.OrganicResults {
		finalResults = append(finalResults, SearchResult{
			Title:       r.Title,
			URL:         r.Link,
			Description: r.Snippet,
		})
	}

	return &SearchResponse{Results: finalResults}, nil
}

func searchGoogle(ctx context.Context, cfg config.GoogleConfig, req *SearchRequest) (*SearchResponse, error) {
	if cfg.ApiKey == "" || cfg.Cx == "" {
		return nil, fmt.Errorf("google 搜索需要配置 ApiKey 和 Cx")
	}

	officialConfig := &googlesearch.Config{
		APIKey:         cfg.ApiKey,
		SearchEngineID: cfg.Cx,
		BaseURL:        cfg.BaseUrl,
	}

	tl, err := googlesearch.NewTool(ctx, officialConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create official google search tool: %w", err)
	}

	officialReq := &googlesearch.SearchRequest{
		Query: req.Query,
		Num:   req.MaxResults,
	}
	reqData, _ := json.Marshal(officialReq)

	// 调用官方工具
	respJson, err := tl.InvokableRun(ctx, string(reqData))
	if err != nil {
		return nil, fmt.Errorf("google search invoke failed: %w", err)
	}

	// 解析官方输出 (SearchResult 结构在 googlesearch 包中定义为 SearchResult{Query, Items: []*SimplifiedSearchItem})
	var officialResp struct {
		Items []struct {
			Link    string `json:"link"`
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
			Desc    string `json:"desc"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(respJson), &officialResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal official google response: %w", err)
	}

	results := make([]SearchResult, 0, len(officialResp.Items))
	for _, item := range officialResp.Items {
		desc := item.Snippet
		if item.Desc != "" {
			desc = item.Desc
		}
		results = append(results, SearchResult{
			Title:       item.Title,
			URL:         item.Link,
			Description: desc,
		})
	}

	return &SearchResponse{Results: results}, nil
}

func searchBing(ctx context.Context, cfg config.BingConfig, req *SearchRequest) (*SearchResponse, error) {
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("bing 搜索需要配置 ApiKey")
	}

	officialConfig := &bingsearch.Config{
		APIKey:     cfg.ApiKey,
		MaxResults: cfg.MaxResults,
		Region:     bingsearch.Region(cfg.Region),
		SafeSearch: bingsearch.SafeSearch(cfg.SafeSearch),
	}

	tl, err := bingsearch.NewTool(ctx, officialConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create official bing search tool: %w", err)
	}

	officialReq := &bingsearch.SearchRequest{
		Query: req.Query,
	}
	reqData, _ := json.Marshal(officialReq)

	respJson, err := tl.InvokableRun(ctx, string(reqData))
	if err != nil {
		return nil, fmt.Errorf("bing search invoke failed: %w", err)
	}

	var officialResp struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(respJson), &officialResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal official bing response: %w", err)
	}

	results := make([]SearchResult, 0, len(officialResp.Results))
	for _, item := range officialResp.Results {
		results = append(results, SearchResult{
			Title:       item.Title,
			URL:         item.URL,
			Description: item.Description,
		})
	}

	return &SearchResponse{Results: results}, nil
}

func searchDuckDuckGo(ctx context.Context, cfg config.DuckDuckGoConfig, req *SearchRequest) (*SearchResponse, error) {
	officialConfig := &duckduckgo.Config{
		MaxResults: cfg.MaxResults,
		Region:     ddgsearch.Region(cfg.Region),
		SafeSearch: ddgsearch.SafeSearch(cfg.SafeSearch),
	}

	tl, err := duckduckgo.NewTool(ctx, officialConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create official ddg search tool: %w", err)
	}

	officialReq := &duckduckgo.SearchRequest{
		Query: req.Query,
	}
	reqData, _ := json.Marshal(officialReq)

	respJson, err := tl.InvokableRun(ctx, string(reqData))
	if err != nil {
		return nil, fmt.Errorf("ddg search invoke failed: %w", err)
	}

	var officialResp struct {
		Results []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Link        string `json:"link"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(respJson), &officialResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal official ddg response: %w", err)
	}

	results := make([]SearchResult, 0, len(officialResp.Results))
	for _, item := range officialResp.Results {
		results = append(results, SearchResult{
			Title:       item.Title,
			URL:         item.Link,
			Description: item.Description,
		})
	}

	return &SearchResponse{Results: results}, nil
}

func NewWebSearchTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("web_search", "执行网页搜索以获取互联网信息", WebSearch)
}
