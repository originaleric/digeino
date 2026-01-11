package ui_ux

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// UIUXSearchRequest defines the input parameters for the UI/UX search tool.
type UIUXSearchRequest struct {
	Query      string `json:"query" jsonschema:"description=检索关键词，如 'minimalism color palette' 或 'dashboard layout'"`
	Domain     string `json:"domain,omitempty" jsonschema:"description=检索领域，可选值: style, color, chart, landing, product, prompt, ux, typography. 不填则自动检测"`
	Stack      string `json:"stack,omitempty" jsonschema:"description=特定技术栈，如 react, tailwind, swiftui 等"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"description=最大返回结果数，默认 5,default=5"`
}

// UIUXSearchResponse defines the output format for the UI/UX search tool.
type UIUXSearchResponse struct {
	Result string `json:"result"`
}

// NewUIUXSearchTool creates a new Eino tool for UI/UX design intelligence.
func NewUIUXSearchTool(ctx context.Context) (tool.BaseTool, error) {
	service := NewUIUXService()

	return utils.InferTool("ui_ux_search", "检索 UI/UX 设计知识库，包括样式指南(style)、配色方案(color)、字体排版(typography)、交互原则(ux)、落地页模式(landing)等。支持按技术栈(stack)过滤，如 react, tailwind 等。",
		func(ctx context.Context, req *UIUXSearchRequest) (*UIUXSearchResponse, error) {
			if req.MaxResults <= 0 {
				req.MaxResults = 5
			}

			var res *SearchResult
			var err error

			if req.Stack != "" {
				res, err = service.SearchStack(req.Query, req.Stack, req.MaxResults)
			} else {
				res, err = service.Search(req.Query, req.Domain, req.MaxResults)
			}

			if err != nil {
				return nil, err
			}

			// Format results as a readable string for the LLM
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("检索结果 (领域: %s, 原始查询: %s):\n\n", res.Domain, res.Query))
			for i, r := range res.Results {
				sb.WriteString(fmt.Sprintf("--- 建议 %d ---\n", i+1))
				for k, v := range r {
					if v != "" {
						sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
					}
				}
				sb.WriteString("\n")
			}

			return &UIUXSearchResponse{Result: sb.String()}, nil
		})
}
