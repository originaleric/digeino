package digeino

import (
	"encoding/json"
	"fmt"
)

// AgentState 是在多 Agent 流程中传递的全局状态
// 注意：业务特定的结构体（如 Outline, Pages, Design 等）应通过 Extensions 字段存储
// 各个项目可以在自己的包中定义业务结构体，并使用 GetBusinessData/SetBusinessData 方法访问
type AgentState struct {
	SessionID       string                 `json:"session_id"`
	Query           string                 `json:"query"`            // 用户的原始需求
	Status          string                 `json:"status"`           // 当前状态
	ResearchSummary string                 `json:"research_summary"` // Researcher 节点的输出总结
	Extensions      map[string]interface{} `json:"extensions,omitempty"` // 扩展字段，用于存储项目特定的数据
	Error           error                  `json:"error"`            // 错误信息
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

// GetBusinessData 获取业务数据结构（通过 JSON 序列化/反序列化进行类型转换）
// 用于获取存储在 Extensions 中的复杂业务对象
// 示例：var outline *DocumentOutline; err := state.GetBusinessData("outline", &outline)
func (s *AgentState) GetBusinessData(key string, target interface{}) error {
	val, ok := s.GetExtension(key)
	if !ok {
		return fmt.Errorf("key %s not found in extensions", key)
	}
	// 使用 JSON 序列化/反序列化进行类型转换
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal business data for key %s: %w", key, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal business data for key %s: %w", key, err)
	}
	return nil
}

// SetBusinessData 设置业务数据结构
// 用于将复杂业务对象存储到 Extensions 中
// 示例：state.SetBusinessData("outline", outline)
func (s *AgentState) SetBusinessData(key string, value interface{}) {
	s.SetExtension(key, value)
}
