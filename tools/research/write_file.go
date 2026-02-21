package research

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	// 验证路径
	if req.Path == "" {
		return nil, fmt.Errorf("文件路径不能为空")
	}

	// 默认模式为覆盖
	mode := req.Mode
	if mode == "" {
		mode = "overwrite"
	}
	if mode != "overwrite" && mode != "append" {
		return nil, fmt.Errorf("写入模式必须是 overwrite 或 append")
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("无法解析文件路径: %w", err)
	}

	// 创建目录（如果不存在）
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建目录: %w", err)
	}

	// 根据模式写入文件
	var file *os.File
	var errOpen error

	if mode == "append" {
		// 追加模式：打开文件，如果不存在则创建
		file, errOpen = os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		// 覆盖模式：创建新文件（如果存在则覆盖）
		file, errOpen = os.Create(absPath)
	}

	if errOpen != nil {
		return nil, fmt.Errorf("无法打开文件: %w", errOpen)
	}
	defer file.Close()

	// 写入内容
	if _, err := file.WriteString(req.Content); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	// 获取文件信息用于返回
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("无法获取文件信息: %w", err)
	}

	message := fmt.Sprintf("文件已成功写入（%s模式），大小：%d 字节", mode, fileInfo.Size())

	return &WriteFileResponse{
		Success: true,
		Path:    absPath,
		Message: message,
	}, nil
}

// --- Factory Function ---

func NewWriteFileTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("research_write_file", "将内容写入文件，支持覆盖和追加模式，自动创建目录", WriteFile)
}
