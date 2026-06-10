package ocr

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/originaleric/digeino/config"
)

// ImageOCR 执行图片 OCR（工具与 Client 共用入口）。
func ImageOCR(ctx context.Context, req *OCRRequest) (*OCRResponse, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	return client.Recognize(ctx, req)
}

// NewImageOCRTool 创建 image_ocr 工具（需在配置中启用 Tools.OCR）。
func NewImageOCRTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get()
	if cfg.Tools.OCR.Enabled == nil || !*cfg.Tools.OCR.Enabled {
		return nil, fmt.Errorf("OCR tool is not enabled in config (Tools.OCR.Enabled)")
	}
	if _, err := NewClient(); err != nil {
		return nil, err
	}
	return utils.InferTool(
		"image_ocr",
		"对图片进行 OCR 识别，支持 URL、Base64 或本地文件路径。可指定任务类型（plain_text、layout、table、receipt、form）及是否返回版面块与坐标。",
		ImageOCR,
	)
}
