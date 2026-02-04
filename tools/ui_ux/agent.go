package ui_ux

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// UIDesignSystemAgent UI/UX 设计系统生成 Agent
// 这是一个编排的 Agent，内部使用工具来完成设计系统生成工作流
type UIDesignSystemAgent struct {
	agent compose.Runnable[[]*schema.Message, *schema.Message]
}

// NewUIDesignSystemAgent 创建 UI/UX 设计系统生成 Agent
// 需要传入 ChatModel，Agent 内部会使用工具进行工作流编排
func NewUIDesignSystemAgent(ctx context.Context, chatModel model.ChatModel, additionalTools ...tool.BaseTool) (*UIDesignSystemAgent, error) {
	// 准备工具集
	allTools := []tool.BaseTool{}

	// 1. UI/UX 搜索工具
	searchTool, err := NewUIUXSearchTool(ctx)
	if err == nil {
		allTools = append(allTools, searchTool)
	}

	// 2. 设计系统生成工具
	designSystemTool, err := NewGenerateDesignSystemTool(ctx)
	if err == nil {
		allTools = append(allTools, designSystemTool)
	}

	// 3. 持久化工具
	persistTool, err := NewPersistDesignSystemTool(ctx)
	if err == nil {
		allTools = append(allTools, persistTool)
	}

	// 4. 添加外部传入的额外工具（如其他项目工具）
	allTools = append(allTools, additionalTools...)

	// 绑定工具到模型
	toolInfos := make([]*schema.ToolInfo, 0, len(allTools))
	for _, t := range allTools {
		if info, err := t.Info(ctx); err == nil {
			toolInfos = append(toolInfos, info)
		}
	}
	if err := chatModel.BindTools(toolInfos); err != nil {
		return nil, fmt.Errorf("failed to bind tools: %w", err)
	}

	// 创建工具执行节点
	toolNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: allTools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tool node: %w", err)
	}

	// 创建 Agent Lambda 节点（手动调用模型和工具）
	agentLambda := func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
		// 构建消息列表（包含系统提示词）
		messages := []*schema.Message{
			schema.SystemMessage(getSystemPrompt()),
		}
		messages = append(messages, input...)

		// 调用模型生成
		output, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("model generate failed: %w", err)
		}

		// 如果有工具调用，执行工具
		if len(output.ToolCalls) > 0 {
			// 执行工具（工具节点接受单个消息，返回消息数组）
			toolResults, err := toolNode.Invoke(ctx, output)
			if err != nil {
				return nil, fmt.Errorf("tool execution failed: %w", err)
			}

			// 将工具结果添加到消息列表，再次调用模型
			messages = append(messages, output)
			messages = append(messages, toolResults...)

			// 再次调用模型处理工具结果（可能需要多轮工具调用）
			for {
				finalOutput, err := chatModel.Generate(ctx, messages)
				if err != nil {
					return nil, fmt.Errorf("model generate after tools failed: %w", err)
				}

				// 如果还有工具调用，继续执行
				if len(finalOutput.ToolCalls) > 0 {
					toolResults, err := toolNode.Invoke(ctx, finalOutput)
					if err != nil {
						return nil, fmt.Errorf("tool execution failed: %w", err)
					}
					messages = append(messages, finalOutput)
					messages = append(messages, toolResults...)
					continue
				}

				return finalOutput, nil
			}
		}

		return output, nil
	}

	// 创建 Graph 工作流
	graph := compose.NewGraph[[]*schema.Message, *schema.Message]()
	_ = graph.AddLambdaNode("ui_design_system_agent", compose.InvokableLambda(agentLambda))

	// 编译 Graph 为 Runnable
	compiledGraph, err := graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	return &UIDesignSystemAgent{agent: compiledGraph}, nil
}

// NewUIDesignSystemAgentFromConfig 从配置创建 UI/UX 设计系统生成 Agent
// 自动从 config.yaml 加载 ChatModel 配置
func NewUIDesignSystemAgentFromConfig(ctx context.Context, additionalTools ...tool.BaseTool) (*UIDesignSystemAgent, error) {
	chatModel, err := NewChatModelFromConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model from config: %w", err)
	}

	return NewUIDesignSystemAgent(ctx, chatModel, additionalTools...)
}

// Invoke 执行 Agent 工作流
func (a *UIDesignSystemAgent) Invoke(ctx context.Context, query string) (*schema.Message, error) {
	messages := []*schema.Message{
		schema.UserMessage(query),
	}

	return a.agent.Invoke(ctx, messages)
}

// getSystemPrompt 获取系统提示词
func getSystemPrompt() string {
	return `You are a UI/UX Design System Generator Agent. Your task is to help users generate complete design systems for their projects.

Workflow:
1. **Analyze Requirements**: Extract key information from user request:
   - Product type (SaaS, e-commerce, portfolio, dashboard, landing page, etc.)
   - Style keywords (minimal, playful, professional, elegant, dark mode, etc.)
   - Industry (healthcare, fintech, gaming, education, etc.)
   - Stack (React, Vue, Next.js, or default to html-tailwind)

2. **Generate Design System** (REQUIRED): Always start by calling generate_design_system tool with:
   - query: User's product description and requirements
   - project_name: Project name if provided

3. **Supplement with Detailed Searches** (if needed): After getting the design system, use ui_ux_search tool to get additional details:
   - More style options: domain="style"
   - Chart recommendations: domain="chart"
   - UX best practices: domain="ux"
   - Alternative fonts: domain="typography"
   - Landing structure: domain="landing"

4. **Stack Guidelines** (if stack specified): Use ui_ux_search with stack parameter to get implementation-specific best practices.

5. **Persist Design System** (if requested): If user wants to save the design system, call persist_design_system tool.

Always provide comprehensive, actionable design system recommendations based on the generated results.`
}
