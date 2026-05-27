package ocr

// OCRRequest 统一 OCR 输入（与方案 Schema 对齐）。
type OCRRequest struct {
	ImageURL     string   `json:"image_url,omitempty" jsonschema_description:"图片 URL，适合云端可访问图片"`
	ImageBase64  string   `json:"image_base64,omitempty" jsonschema_description:"图片 Base64（可带 data:image/...;base64, 前缀）"`
	FilePath     string   `json:"file_path,omitempty" jsonschema_description:"本地文件路径，需在配置的 AllowedFilePaths 前缀内"`
	MimeType     string   `json:"mime_type,omitempty" jsonschema_description:"图片 MIME 类型，如 image/png"`
	Task         string   `json:"task,omitempty" jsonschema_description:"任务类型：plain_text、layout、table、form、invoice"`
	Languages    []string `json:"languages,omitempty" jsonschema_description:"语言提示，如 zh、en"`
	ReturnBBox   bool     `json:"return_bbox,omitempty" jsonschema_description:"是否返回文字坐标"`
	ReturnLayout bool     `json:"return_layout,omitempty" jsonschema_description:"是否返回版面结构"`
}

// OCRBlock 版面块。
type OCRBlock struct {
	Type       string    `json:"type,omitempty"`
	Text       string    `json:"text"`
	BBox       []float64 `json:"bbox,omitempty"`
	Confidence float64   `json:"confidence,omitempty"`
}

// OCRTable 表格（预留结构化扩展）。
type OCRTable struct {
	Rows [][]string `json:"rows,omitempty"`
}

// OCRUsage 模型用量。
type OCRUsage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
}

// OCRResponse 统一 OCR 输出。
type OCRResponse struct {
	Text       string     `json:"text"`
	Blocks     []OCRBlock `json:"blocks,omitempty"`
	Tables     []OCRTable `json:"tables,omitempty"`
	Confidence float64    `json:"confidence,omitempty"`
	Provider   string     `json:"provider"`
	Model      string     `json:"model"`
	Usage      *OCRUsage  `json:"usage,omitempty"`
}

// resolvedImage 内部：解析后的图片载荷。
type resolvedImage struct {
	DataURL  string // data:image/png;base64,... 或 https://...
	MimeType string
	Source   string // url | base64 | file
}
