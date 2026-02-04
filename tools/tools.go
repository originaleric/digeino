package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/originaleric/digeino/tools/ui_ux"
	"github.com/originaleric/digeino/tools/wx"
)

// safeTool 安全的工具包装器，用于处理错误
type safeTool struct {
	tool.InvokableTool
}

func (s safeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return s.InvokableTool.Info(ctx)
}

func (s safeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	out, e := s.InvokableTool.InvokableRun(ctx, argumentsInJSON, opts...)
	if e != nil {
		return e.Error(), nil
	}

	return out, nil
}

// SafeInferTool 安全地创建工具
func SafeInferTool[T, D any](toolName, toolDesc string, i utils.InvokeFunc[T, D]) (tool.InvokableTool, error) {
	t, err := utils.InferTool(toolName, toolDesc, i)
	if err != nil {
		return nil, err
	}

	return safeTool{
		InvokableTool: t,
	}, nil
}

// BaseTools 获取 digeino 提供的通用基础工具集
func BaseTools(ctx context.Context) ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	// 1. UI/UX 设计智能工具
	uiTool, err := ui_ux.NewUIUXSearchTool(ctx)
	if err == nil {
		tools = append(tools, uiTool)
	}

	// 1.1. UI/UX 设计系统生成工具
	designSystemTool, err := ui_ux.NewGenerateDesignSystemTool(ctx)
	if err == nil {
		tools = append(tools, designSystemTool)
	}

	// 1.2. UI/UX 设计系统持久化工具
	persistTool, err := ui_ux.NewPersistDesignSystemTool(ctx)
	if err == nil {
		tools = append(tools, persistTool)
	}

	// 2. 微信推送工具（如果已启用）
	wxTool, err := wx.NewWeChatMessageTool(ctx)
	if err == nil {
		tools = append(tools, wxTool)
	}

	// 这里未来可以放入通用的 Google 搜索、计算器等跨项目工具
	return tools, nil
}
