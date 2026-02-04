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

// GenerateDesignSystemRequest 设计系统生成请求
type GenerateDesignSystemRequest struct {
	Query       string `json:"query" jsonschema:"description=用户需求描述，如 'beauty spa wellness service'"`
	ProjectName string `json:"project_name,omitempty" jsonschema:"description=项目名称，如 'Serenity Spa'"`
}

// GenerateDesignSystemResponse 设计系统生成响应
type GenerateDesignSystemResponse struct {
	DesignSystem string `json:"design_system"` // 设计系统的格式化输出
}

// NewGenerateDesignSystemTool 创建设计系统生成工具
func NewGenerateDesignSystemTool(ctx context.Context) (tool.BaseTool, error) {
	service := NewUIUXService()

	return utils.InferTool("generate_design_system",
		"生成完整的 UI/UX 设计系统，包括样式、配色、字体、布局模式等。基于产品类型和需求，自动匹配最佳的设计方案。",
		func(ctx context.Context, req *GenerateDesignSystemRequest) (*GenerateDesignSystemResponse, error) {
			if req.ProjectName == "" {
				req.ProjectName = "Project"
			}

			ds, err := service.GenerateDesignSystem(req.Query, req.ProjectName)
			if err != nil {
				return nil, err
			}

			// 格式化输出为可读字符串
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("设计系统: %s\n", ds.ProjectName))
			sb.WriteString(fmt.Sprintf("产品类型: %s\n\n", ds.Category))

			sb.WriteString("=== 布局模式 ===\n")
			sb.WriteString(fmt.Sprintf("模式名称: %s\n", ds.Pattern.Name))
			if ds.Pattern.Sections != "" {
				sb.WriteString(fmt.Sprintf("区块顺序: %s\n", ds.Pattern.Sections))
			}
			if ds.Pattern.CTAPlacement != "" {
				sb.WriteString(fmt.Sprintf("CTA 位置: %s\n", ds.Pattern.CTAPlacement))
			}
			if ds.Pattern.Conversion != "" {
				sb.WriteString(fmt.Sprintf("转化策略: %s\n", ds.Pattern.Conversion))
			}
			sb.WriteString("\n")

			sb.WriteString("=== 样式 ===\n")
			sb.WriteString(fmt.Sprintf("样式名称: %s\n", ds.Style.Name))
			if ds.Style.Keywords != "" {
				sb.WriteString(fmt.Sprintf("关键词: %s\n", ds.Style.Keywords))
			}
			if ds.Style.BestFor != "" {
				sb.WriteString(fmt.Sprintf("适用场景: %s\n", ds.Style.BestFor))
			}
			if ds.Style.Performance != "" {
				sb.WriteString(fmt.Sprintf("性能: %s\n", ds.Style.Performance))
			}
			if ds.Style.Accessibility != "" {
				sb.WriteString(fmt.Sprintf("可访问性: %s\n", ds.Style.Accessibility))
			}
			sb.WriteString("\n")

			sb.WriteString("=== 配色方案 ===\n")
			sb.WriteString(fmt.Sprintf("主色: %s\n", ds.Colors.Primary))
			sb.WriteString(fmt.Sprintf("次色: %s\n", ds.Colors.Secondary))
			sb.WriteString(fmt.Sprintf("CTA: %s\n", ds.Colors.CTA))
			sb.WriteString(fmt.Sprintf("背景: %s\n", ds.Colors.Background))
			sb.WriteString(fmt.Sprintf("文字: %s\n", ds.Colors.Text))
			if ds.Colors.Notes != "" {
				sb.WriteString(fmt.Sprintf("备注: %s\n", ds.Colors.Notes))
			}
			sb.WriteString("\n")

			sb.WriteString("=== 字体排版 ===\n")
			sb.WriteString(fmt.Sprintf("标题字体: %s\n", ds.Typography.Heading))
			sb.WriteString(fmt.Sprintf("正文字体: %s\n", ds.Typography.Body))
			if ds.Typography.Mood != "" {
				sb.WriteString(fmt.Sprintf("风格: %s\n", ds.Typography.Mood))
			}
			if ds.Typography.BestFor != "" {
				sb.WriteString(fmt.Sprintf("适用场景: %s\n", ds.Typography.BestFor))
			}
			if ds.Typography.GoogleFontsURL != "" {
				sb.WriteString(fmt.Sprintf("Google Fonts: %s\n", ds.Typography.GoogleFontsURL))
			}
			sb.WriteString("\n")

			if ds.KeyEffects != "" {
				sb.WriteString("=== 关键效果 ===\n")
				sb.WriteString(fmt.Sprintf("%s\n\n", ds.KeyEffects))
			}

			if ds.AntiPatterns != "" {
				sb.WriteString("=== 避免的反模式 ===\n")
				antiList := strings.Split(ds.AntiPatterns, "+")
				for _, anti := range antiList {
					if trimmed := strings.TrimSpace(anti); trimmed != "" {
						sb.WriteString(fmt.Sprintf("- %s\n", trimmed))
					}
				}
			}

			return &GenerateDesignSystemResponse{DesignSystem: sb.String()}, nil
		})
}

// PersistDesignSystemRequest 持久化设计系统请求
type PersistDesignSystemRequest struct {
	Query       string `json:"query" jsonschema:"description=用户需求描述，用于生成设计系统"`
	ProjectName string `json:"project_name" jsonschema:"description=项目名称，如 'Serenity Spa'"`
	PageName    string `json:"page_name,omitempty" jsonschema:"description=页面名称，用于生成页面特定覆盖文件"`
	BaseDir     string `json:"base_dir,omitempty" jsonschema:"description=基础目录，如果为空则从配置读取（config.UIUX.Storage.BaseDir）"`
	AppName     string `json:"app_name,omitempty" jsonschema:"description=应用/agent 名称，用于隔离不同应用的存储（可选）。如果指定，存储路径为 {BaseDir}/{app_name}/design-system/{project}/MASTER.md"`
}

// PersistDesignSystemResponse 持久化设计系统响应
type PersistDesignSystemResponse struct {
	Success      bool     `json:"success"`
	MasterFile   string   `json:"master_file,omitempty"`
	PageFile     string   `json:"page_file,omitempty"`
	ProjectSlug  string   `json:"project_slug"`
	Message      string   `json:"message"`
}

// NewPersistDesignSystemTool 创建持久化设计系统工具
func NewPersistDesignSystemTool(ctx context.Context) (tool.BaseTool, error) {
	service := NewUIUXService()

	return utils.InferTool("persist_design_system",
		"将设计系统持久化到文件系统，使用 Master + Overrides 模式。会创建 design-system/{project}/MASTER.md 和可选的 pages/{page}.md 文件。",
		func(ctx context.Context, req *PersistDesignSystemRequest) (*PersistDesignSystemResponse, error) {
			if req.ProjectName == "" {
				return nil, fmt.Errorf("project_name is required")
			}

			// 生成设计系统
			ds, err := service.GenerateDesignSystem(req.Query, req.ProjectName)
			if err != nil {
				return nil, fmt.Errorf("failed to generate design system: %w", err)
			}

			// 持久化
			manager := NewPersistenceManager(req.BaseDir, req.AppName)
			projectSlug := strings.ToLower(strings.ReplaceAll(req.ProjectName, " ", "-"))
			
			if err := manager.PersistDesignSystem(ds, projectSlug, req.PageName); err != nil {
				return nil, fmt.Errorf("failed to persist design system: %w", err)
			}

			// 构建响应路径（用于显示）
			var masterPath, pagePath string
			baseDir := manager.GetBaseDir()
			appName := manager.GetAppName()
			
			if appName != "" {
				masterPath = fmt.Sprintf("%s/%s/design-system/%s/MASTER.md", baseDir, appName, projectSlug)
				if req.PageName != "" {
					pageSlug := strings.ToLower(strings.ReplaceAll(req.PageName, " ", "-"))
					pagePath = fmt.Sprintf("%s/%s/design-system/%s/pages/%s.md", baseDir, appName, projectSlug, pageSlug)
				}
			} else {
				masterPath = fmt.Sprintf("%s/design-system/%s/MASTER.md", baseDir, projectSlug)
				if req.PageName != "" {
					pageSlug := strings.ToLower(strings.ReplaceAll(req.PageName, " ", "-"))
					pagePath = fmt.Sprintf("%s/design-system/%s/pages/%s.md", baseDir, projectSlug, pageSlug)
				}
			}

			response := &PersistDesignSystemResponse{
				Success:     true,
				ProjectSlug: projectSlug,
				MasterFile:  masterPath,
				PageFile:    pagePath,
				Message:     fmt.Sprintf("Design system persisted successfully for project '%s'", req.ProjectName),
			}

			if req.PageName != "" {
				response.Message += fmt.Sprintf(" with page override '%s'", req.PageName)
			}
			if req.AppName != "" {
				response.Message += fmt.Sprintf(" (isolated for app '%s')", req.AppName)
			}

			return response, nil
		})
}
