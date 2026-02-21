package websearch

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
)

// WebSearch 执行网页搜索
func WebSearch(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	cfg := config.Get()
	toolCfg := cfg.Tools.WebSearch

	engine := toolCfg.Engine
	if engine == "" {
		engine = "duckduckgo" // 默认为 duckduckgo 占位
	}

	fmt.Printf("执行搜索, engine: %s, query: %s, max_results: %d\n", engine, req.Query, req.MaxResults)

	var provider SearchProvider
	var err error

	// 转换为 map[string]interface{} 以兼容各 Provider 接参
	configMap := make(map[string]interface{})

	switch engine {
	case "bocha":
		configMap["BochaApiKey"] = toolCfg.Bocha.ApiKey
		configMap["BochaBaseUrl"] = toolCfg.Bocha.BaseUrl
		provider, err = NewBochaProvider(configMap)
	case "serpapi":
		configMap["SerpAPIKey"] = toolCfg.SerpApi.ApiKey
		configMap["SerpAPIEngine"] = "google"
		if toolCfg.SerpApi.BaseUrl != "" {
			// 如果有自定义URL需求，此处可以扩展，目前原生没这个参数
		}
		provider, err = NewSerpAPIProvider(configMap)
	case "google":
		configMap["GoogleApiKey"] = toolCfg.Google.ApiKey
		configMap["GoogleSearchEngineId"] = toolCfg.Google.Cx
		provider, err = NewGoogleProvider(configMap)
	case "bing":
		configMap["BingApiKey"] = toolCfg.Bing.ApiKey
		// configMap["BingSafeSearch"] = toolCfg.Bing.SafeSearch
		provider, err = NewBingProvider(configMap)
	case "duckduckgo":
		provider = NewDuckDuckGoProvider(configMap)
	default:
		return nil, fmt.Errorf("暂不支持的搜索引擎: %s，请配置 bocha, serpapi, google, bing 或 duckduckgo", engine)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w", engine, err)
	}

	// 执行搜索
	response, err := provider.Search(ctx, req)
	if err != nil {
		fmt.Printf("搜索失败, engine: %s, query: %s, error: %v\n", engine, req.Query, err)
		return nil, fmt.Errorf("search failed with %s: %w", provider.Name(), err)
	}

	fmt.Printf("搜索成功, engine: %s, query: %s, results_count: %d\n", engine, req.Query, len(response.Results))
	return response, nil
}

// NewWebSearchTool 创建搜索工具
func NewWebSearchTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("web_search", "执行网页搜索以获取互联网信息", WebSearch)
}
