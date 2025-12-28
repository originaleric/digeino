package digeino

// AgentState 是在多 Agent 流程中传递的全局状态
type AgentState struct {
	SessionID       string                 `json:"session_id"`
	Query           string                 `json:"query"`            // 用户的原始需求
	Outline         *DocumentOutline       `json:"outline"`          // 规划的文档大纲
	Pages           []*Page                `json:"pages"`            // 生成的每一页内容
	Status          string                 `json:"status"`           // 当前状态
	Design          *DesignConfig          `json:"design"`           // LLM 生成的设计配置
	InputFiles      []*InputFile           `json:"input_files"`      // 输入的文件列表（PDF, Video 等）
	ResearchSummary string                 `json:"research_summary"` // Researcher 节点的输出总结
	Extensions      map[string]interface{} `json:"extensions,omitempty"` // 扩展字段，用于存储项目特定的数据
	Error           error                  `json:"error"`            // 错误信息
}

// InputFile 输入的文件信息
type InputFile struct {
	Type    string `json:"type"` // "pdf", "video"
	URL     string `json:"url"`
	Name    string `json:"name"`
	Content string `json:"content"` // 提取出的文本内容（由前端或预处理提供）
}

// DesignConfig 设计配置
type DesignConfig struct {
	CSS              string `json:"css"`                // 全局 CSS，包含 h1, p, .page 等各个元素的差异化样式
	GlobalImageStyle string `json:"global_image_style"` // 全局生图风格指令
}

// DocumentOutline 文档大纲
type DocumentOutline struct {
	Title          string        `json:"title"`
	Topic          string        `json:"topic"`
	TargetAudience string        `json:"target_audience"`
	PageOutlines   []PageOutline `json:"pages"`
}

// PageOutline 每一页的规划
type PageOutline struct {
	PageNumber  int    `json:"page_number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	LayoutType  string `json:"layout_type"` // text_only, full_image, text_image_split
}

// Page 最终生成的页面内容
type Page struct {
	PageNumber int        `json:"page_number"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`         // 生成的文字内容
	Image      *PageImage `json:"image,omitempty"` // 生成的图片信息
	Layout     string     `json:"layout"`          // 布局类型
}

// PageImage 图片信息
type PageImage struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt"`
	ImageURL       string `json:"image_url"`
}

// ContextKey 用于在 context 中传递 key
type ContextKey string

const (
	CtxKeySessionID ContextKey = "session_id"
)

// GetExtension 获取扩展字段的值
func (s *AgentState) GetExtension(key string) (interface{}, bool) {
	if s.Extensions == nil {
		return nil, false
	}
	val, ok := s.Extensions[key]
	return val, ok
}

// SetExtension 设置扩展字段的值
func (s *AgentState) SetExtension(key string, value interface{}) {
	if s.Extensions == nil {
		s.Extensions = make(map[string]interface{})
	}
	s.Extensions[key] = value
}

// GetStringExtension 获取字符串类型的扩展字段
func (s *AgentState) GetStringExtension(key string) (string, bool) {
	val, ok := s.GetExtension(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// SetStringExtension 设置字符串类型的扩展字段
func (s *AgentState) SetStringExtension(key string, value string) {
	s.SetExtension(key, value)
}

// GetIntExtension 获取整数类型的扩展字段
func (s *AgentState) GetIntExtension(key string) (int, bool) {
	val, ok := s.GetExtension(key)
	if !ok {
		return 0, false
	}
	// 支持多种数字类型
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// SetIntExtension 设置整数类型的扩展字段
func (s *AgentState) SetIntExtension(key string, value int) {
	s.SetExtension(key, value)
}

// GetBoolExtension 获取布尔类型的扩展字段
func (s *AgentState) GetBoolExtension(key string) (bool, bool) {
	val, ok := s.GetExtension(key)
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// SetBoolExtension 设置布尔类型的扩展字段
func (s *AgentState) SetBoolExtension(key string, value bool) {
	s.SetExtension(key, value)
}
