package research

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// --- Write File Tool ---

type WriteFileRequest struct {
	Path    string `json:"path" jsonschema:"required,description=文件路径（相对路径或绝对路径）"`
	Content string `json:"content" jsonschema:"required,description=要写入的文件内容"`
	Mode    string `json:"mode" jsonschema:"description=写入模式：overwrite（覆盖，默认）或 append（追加）"`
}

type WriteFileResponse struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// WriteFile 写入文件工具，支持覆盖和追加模式
func WriteFile(ctx context.Context, req *WriteFileRequest) (*WriteFileResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	// 默认模式为覆盖
	mode := req.Mode
	if mode == "" {
		mode = "overwrite"
	}
	if mode != "overwrite" && mode != "append" {
		return nil, fmt.Errorf("写入模式必须是 overwrite 或 append")
	}

	file, absPath, err := openAllowedWriteFile(req.Path, mode)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := file.WriteString(req.Content); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}
	if err := file.Sync(); err != nil {
		return nil, fmt.Errorf("同步文件失败: %w", err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("无法获取文件信息: %w", err)
	}

	return &WriteFileResponse{
		Success: true,
		Path:    absPath,
		Message: fmt.Sprintf("文件已成功写入（%s模式），大小：%d 字节", mode, fileInfo.Size()),
	}, nil
}

// --- Factory Function ---

func NewWriteFileTool(ctx context.Context) (tool.BaseTool, error) {
	if !secureFileAccessSupported() {
		return nil, fmt.Errorf("research_write_file tool is disabled because secure file access is not supported on this platform")
	}
	if len(allowedPathsFor(pathAccessWrite)) == 0 {
		return nil, fmt.Errorf("research_write_file tool is disabled without Gateway.AllowedWritePaths")
	}
	return utils.InferTool("research_write_file", "将内容写入文件，支持覆盖和追加模式，自动创建目录", WriteFile)
}
