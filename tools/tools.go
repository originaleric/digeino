package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/originaleric/digeino/tools/research"
	"github.com/originaleric/digeino/tools/research/websearch"
	"github.com/originaleric/digeino/tools/storage"
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

	// 1.3. UI/UX 技术质量审查工具
	auditTool, err := ui_ux.NewUIUXAuditTool(ctx)
	if err == nil {
		tools = append(tools, auditTool)
	}

	// 1.4. UI/UX 设计审查工具
	critiqueTool, err := ui_ux.NewUIUXCritiqueTool(ctx)
	if err == nil {
		tools = append(tools, critiqueTool)
	}

	// 1.5. UI/UX 设计系统标准化工具
	normalizeTool, err := ui_ux.NewUIUXNormalizeTool(ctx)
	if err == nil {
		tools = append(tools, normalizeTool)
	}

	// 1.6. UI/UX 参考文档检索工具
	referenceTool, err := ui_ux.NewUIUXReferenceTool(ctx)
	if err == nil {
		tools = append(tools, referenceTool)
	}

	// 1.7. 临时存储工具（供 audit/critique 使用，需 context 注入 workspace_path 或 agent_session_id）
	writeReviewFileTool, err := storage.NewWriteReviewFileTool(ctx)
	if err == nil {
		tools = append(tools, writeReviewFileTool)
	}

	// 2. 微信推送工具（如果已启用）
	wxTool, err := wx.NewWeChatMessageTool(ctx)
	if err == nil {
		tools = append(tools, wxTool)
	}

	// 3. 企业微信推送工具（如果已启用）
	wecomTool, err := wx.NewWeComMessageTool(ctx)
	if err == nil {
		tools = append(tools, wecomTool)
	}
	wecomImageTool, err := wx.NewWeComImageTool(ctx)
	if err == nil {
		tools = append(tools, wecomImageTool)
	}
	wecomTextCardTool, err := wx.NewWeComTextCardTool(ctx)
	if err == nil {
		tools = append(tools, wecomTextCardTool)
	}
	wecomCustomerTool, err := wx.NewWeComCustomerMessageTool(ctx)
	if err == nil {
		tools = append(tools, wecomCustomerTool)
	}
	wecomCustomerImageTool, err := wx.NewWeComCustomerImageTool(ctx)
	if err == nil {
		tools = append(tools, wecomCustomerImageTool)
	}
	wecomCustomerVoiceTool, err := wx.NewWeComCustomerVoiceTool(ctx)
	if err == nil {
		tools = append(tools, wecomCustomerVoiceTool)
	}
	wecomCustomerVideoTool, err := wx.NewWeComCustomerVideoTool(ctx)
	if err == nil {
		tools = append(tools, wecomCustomerVideoTool)
	}
	wecomCustomerFileTool, err := wx.NewWeComCustomerFileTool(ctx)
	if err == nil {
		tools = append(tools, wecomCustomerFileTool)
	}
	wecomCustomerLinkTool, err := wx.NewWeComCustomerLinkTool(ctx)
	if err == nil {
		tools = append(tools, wecomCustomerLinkTool)
	}
	wecomCustomerMiniprogramTool, err := wx.NewWeComCustomerMiniprogramTool(ctx)
	if err == nil {
		tools = append(tools, wecomCustomerMiniprogramTool)
	}
	// 3.1. 接收企业微信客服消息工具（如果已启用回调功能）
	receiveWeComCustomerTool, err := wx.NewReceiveWeComCustomerMessageTool(ctx)
	if err == nil {
		tools = append(tools, receiveWeComCustomerTool)
	}

	// 4. 通用调研系列工具
	grepTool, err := research.NewGrepSearchTool(ctx)
	if err == nil {
		tools = append(tools, grepTool)
	}
	readTool, err := research.NewReadFileTool(ctx)
	if err == nil {
		tools = append(tools, readTool)
	}
	docTool, err := research.NewDocToMarkdownTool(ctx)
	if err == nil {
		tools = append(tools, docTool)
	}
	firecrawlTool, err := research.NewFirecrawlTool(ctx)
	if err == nil {
		tools = append(tools, firecrawlTool)
	}
	webSearchTool, err := websearch.NewWebSearchTool(ctx)
	if err == nil {
		tools = append(tools, webSearchTool)
	}
	semanticSearchTool, err := research.NewSemanticSearchTool(ctx)
	if err == nil {
		tools = append(tools, semanticSearchTool)
	}
	codeIndexTool, err := research.NewCodeIndexTool(ctx)
	if err == nil {
		tools = append(tools, codeIndexTool)
	}
	jinaReaderTool, err := research.NewJinaReaderTool(ctx)
	if err == nil {
		tools = append(tools, jinaReaderTool)
	}
	writeFileTool, err := research.NewWriteFileTool(ctx)
	if err == nil {
		tools = append(tools, writeFileTool)
	}

	// 5. 本地无头浏览器抓取工具（高级反爬场景，如微信公众号）
	localScraperTool, err := research.NewLocalScraperTool(ctx)
	if err == nil {
		tools = append(tools, localScraperTool)
	}
	// 5.1 本地浏览器通用抓取工具（read/screenshot + cookie 复用）
	browserBrowseTool, err := research.NewBrowserBrowseTool(ctx)
	if err == nil {
		tools = append(tools, browserBrowseTool)
	}
	// 5.2 浏览器快照工具（结构化快照，提取可交互元素）
	browserSnapshotTool, err := research.NewBrowserSnapshotTool(ctx)
	if err == nil {
		tools = append(tools, browserSnapshotTool)
	}
	// 5.3 浏览器操作工具（点击、输入、悬停、滚动等交互操作）
	browserActionTool, err := research.NewBrowserActionTool(ctx)
	if err == nil {
		tools = append(tools, browserActionTool)
	}

	// 这里未来可以放入通用的 Google 搜索、计算器等跨项目工具
	return tools, nil
}
