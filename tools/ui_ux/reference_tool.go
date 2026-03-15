package ui_ux

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// UIUXReferenceRequest 参考文档检索请求
type UIUXReferenceRequest struct {
	Domain     string `json:"domain" jsonschema:"description=参考文档领域，可选值: typography, color, spatial, motion, interaction, responsive, ux_writing"`
	Query      string `json:"query,omitempty" jsonschema:"description=检索关键词（可选），如 'vertical rhythm' 或 'focus states'"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"description=最大返回结果数，默认 5,default=5"`
}

// UIUXReferenceResponse 参考文档检索响应
type UIUXReferenceResponse struct {
	Result string `json:"result"`
}

// NewUIUXReferenceTool 创建参考文档检索工具
func NewUIUXReferenceTool(ctx context.Context) (tool.BaseTool, error) {
	service := NewUIUXService()

	return utils.InferTool("ui_ux_reference",
		"检索 UI/UX 设计参考文档，包括排版(typography)、颜色与对比度(color)、空间设计(spatial)、动效设计(motion)、交互设计(interaction)、响应式设计(responsive)、UX 文案(ux_writing)等领域的最佳实践和反模式。",
		func(ctx context.Context, req *UIUXReferenceRequest) (*UIUXReferenceResponse, error) {
			if req.Domain == "" {
				return nil, fmt.Errorf("domain is required")
			}

			if req.MaxResults <= 0 {
				req.MaxResults = 5
			}

			// 映射领域名称到配置键
			domainMap := map[string]string{
				"typography":  "reference_typography",
				"color":       "reference_color",
				"spatial":     "reference_spatial",
				"motion":      "reference_motion",
				"interaction": "reference_interaction",
				"responsive":  "reference_responsive",
				"ux_writing":  "reference_ux_writing",
			}

			configKey, ok := domainMap[req.Domain]
			if !ok {
				return nil, fmt.Errorf("unknown domain: %s. Valid domains: typography, color, spatial, motion, interaction, responsive, ux_writing", req.Domain)
			}

			// 如果没有查询词，返回该领域的所有内容
			query := req.Query
			if query == "" {
				query = req.Domain // 使用领域名称作为默认查询
			}

			// 使用现有的 Search 方法，但指定参考文档配置
			config, ok := CSVConfigs[configKey]
			if !ok {
				return nil, fmt.Errorf("reference domain not configured: %s", req.Domain)
			}

			// 加载 CSV 数据
			data, err := service.LoadCSV(config.File)
			if err != nil {
				return nil, fmt.Errorf("failed to load reference csv %s: %w", config.File, err)
			}

			// 构建文档用于 BM25 搜索
			documents := make([]string, len(data))
			for i, row := range data {
				var builder strings.Builder
				for _, col := range config.SearchCols {
					builder.WriteString(row[col])
					builder.WriteString(" ")
				}
				documents[i] = builder.String()
			}

			// 执行 BM25 搜索
			bm25 := NewBM25(1.5, 0.75)
			bm25.Fit(documents)
			ranked := bm25.Score(query)

			// 按分数排序
			type scoredDoc struct {
				index int
				score float64
			}
			scoredDocs := make([]scoredDoc, len(ranked))
			for i, r := range ranked {
				scoredDocs[i] = scoredDoc{index: r.Index, score: r.Score}
			}
			sort.Slice(scoredDocs, func(i, j int) bool {
				return scoredDocs[i].score > scoredDocs[j].score
			})

			// 构建结果
			var finalResults []map[string]string
			for i := 0; i < len(scoredDocs) && len(finalResults) < req.MaxResults; i++ {
				if scoredDocs[i].score > 0 {
					origRow := data[scoredDocs[i].index]
					filteredRow := make(map[string]string)
					for _, col := range config.OutputCols {
						if val, ok := origRow[col]; ok {
							filteredRow[col] = val
						}
					}
					finalResults = append(finalResults, filteredRow)
				}
			}

			// 如果没有匹配结果，返回前几个结果
			if len(finalResults) == 0 && len(data) > 0 {
				for i := 0; i < len(data) && i < req.MaxResults; i++ {
					origRow := data[i]
					filteredRow := make(map[string]string)
					for _, col := range config.OutputCols {
						if val, ok := origRow[col]; ok {
							filteredRow[col] = val
						}
					}
					finalResults = append(finalResults, filteredRow)
				}
			}

			// 格式化为可读字符串
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("参考文档检索结果 (领域: %s, 查询: %s):\n\n", req.Domain, query))
			for i, r := range finalResults {
				sb.WriteString(fmt.Sprintf("--- 参考 %d ---\n", i+1))
				for k, v := range r {
					if v != "" {
						sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
					}
				}
				sb.WriteString("\n")
			}

			return &UIUXReferenceResponse{Result: sb.String()}, nil
		})
}
