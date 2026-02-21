package websearch

import (
	"context"
)

// SearchRequest 搜索请求
type SearchRequest struct {
	Query      string `json:"query" jsonschema_description:"搜索关键词"`
	MaxResults int    `json:"max_results,omitempty" jsonschema_description:"最大返回结果数，默认10"`
	Region     string `json:"region,omitempty" jsonschema_description:"搜索地区，如 zh-CN, en-US"`
}

// SearchResult 搜索结果项
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results []*SearchResult `json:"results"`
}

// SearchProvider 搜索提供者接口
type SearchProvider interface {
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
	Name() string
}
