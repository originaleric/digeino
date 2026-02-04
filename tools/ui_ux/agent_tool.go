package ui_ux

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// UIDesignSystemAgentToolRequest Agent 工具请求
type UIDesignSystemAgentToolRequest struct {
	Query       string `json:"query" jsonschema:"description=用户需求描述，如 'beauty spa wellness service'"`
	ProjectName string `json:"project_name,omitempty" jsonschema:"description=项目名称，如 'Serenity Spa'"`
	Stack       string `json:"stack,omitempty" jsonschema:"description=技术栈，如 'react', 'tailwind' 等"`
	Persist     bool   `json:"persist,omitempty" jsonschema:"description=是否持久化设计系统"`
	PageName    string `json:"page_name,omitempty" jsonschema:"description=页面名称，用于生成页面特定覆盖"`
}

// UIDesignSystemAgentToolResponse Agent 工具响应
type UIDesignSystemAgentToolResponse struct {
	DesignSystem string `json:"design_system"` // 设计系统的完整输出
	Message      string `json:"message"`       // 附加消息
}

// NewUIDesignSystemAgentTool 将 Agent 包装成工具，供其他 Agent 使用
// 注意：需要传入 ChatModel，因为 Agent 需要 LLM 来决策工作流
// additionalTools: 可选的额外工具，用于扩展 Agent 能力
func NewUIDesignSystemAgentTool(ctx context.Context, chatModel model.ChatModel, additionalTools ...tool.BaseTool) (tool.BaseTool, error) {
	// 创建 Agent 实例
	agentInstance, err := NewUIDesignSystemAgent(ctx, chatModel, additionalTools...)
	if err != nil {
		return nil, fmt.Errorf("failed to create UI design system agent: %w", err)
	}

	// 将 Agent 包装成工具
	return utils.InferTool(
		"generate_ui_design_system",
		"生成完整的 UI/UX 设计系统，包括样式、配色、字体、布局模式等。这是一个智能编排的 Agent，会自动分析需求、生成设计系统、补充搜索并获取技术栈指南。如果需要持久化，可以设置 persist=true。",
		func(ctx context.Context, req *UIDesignSystemAgentToolRequest) (*UIDesignSystemAgentToolResponse, error) {
			// 构建查询字符串
			queryParts := []string{req.Query}
			if req.ProjectName != "" {
				queryParts = append(queryParts, fmt.Sprintf("项目名称: %s", req.ProjectName))
			}
			if req.Stack != "" {
				queryParts = append(queryParts, fmt.Sprintf("技术栈: %s", req.Stack))
			}
			if req.Persist {
				queryParts = append(queryParts, "请持久化设计系统")
				if req.PageName != "" {
					queryParts = append(queryParts, fmt.Sprintf("页面名称: %s", req.PageName))
				}
			}

			query := fmt.Sprintf("生成设计系统：%s", strings.Join(queryParts, "，"))

			// 调用 Agent 执行完整工作流
			result, err := agentInstance.Invoke(ctx, query)
			if err != nil {
				return nil, fmt.Errorf("agent execution failed: %w", err)
			}

			// 解析 Agent 返回的结果
			response := &UIDesignSystemAgentToolResponse{
				DesignSystem: result.Content,
				Message:      "Design system generated successfully",
			}

			if req.Persist {
				response.Message += " and persisted to design-system/ directory"
			}

			return response, nil
		},
	)
}

// NewUIDesignSystemAgentToolFromConfig 从配置创建 Agent 工具（便捷方法）
// 自动从 config.yaml 加载 ChatModel 配置
func NewUIDesignSystemAgentToolFromConfig(ctx context.Context, additionalTools ...tool.BaseTool) (tool.BaseTool, error) {
	chatModel, err := NewChatModelFromConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model from config: %w", err)
	}

	return NewUIDesignSystemAgentTool(ctx, chatModel, additionalTools...)
}
