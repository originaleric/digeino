package storage

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/pkg/tempstorage"
)

// WriteReviewFileRequest 写入审查文件请求
type WriteReviewFileRequest struct {
	Filename string `json:"filename" jsonschema:"required,description=文件名（如 review.css、design.html），支持子目录如 sub/review.tsx"`
	Content  string `json:"content" jsonschema:"required,description=要写入的文件内容"`
}

// WriteReviewFileResponse 写入审查文件响应
type WriteReviewFileResponse struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// NewWriteReviewFileTool 创建 write_review_file 工具。
// 需在 context 中注入 workspace_path 或 agent_session_id，返回写入后的绝对路径供 ui_ux_audit / ui_ux_critique 使用。
func NewWriteReviewFileTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("write_review_file",
		"将内容写入临时/工作空间文件，返回绝对路径。需 context 注入 workspace_path 或 agent_session_id。与 ui_ux_audit、ui_ux_critique 配合使用：先写入得到 path，再传入审查工具。",
		func(ctx context.Context, req *WriteReviewFileRequest) (*WriteReviewFileResponse, error) {
			if req.Filename == "" {
				return nil, fmt.Errorf("filename is required")
			}
			path, err := tempstorage.SaveForReview(ctx, req.Filename, req.Content)
			if err != nil {
				return nil, fmt.Errorf("write_review_file: %w", err)
			}
			return &WriteReviewFileResponse{
				Path:    path,
				Message: "文件已写入，path 可传入 ui_ux_audit 或 ui_ux_critique 进行审查",
			}, nil
		})
}
