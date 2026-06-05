package ui_ux

import (
	"context"
	"fmt"
	"os"
	"strings"

	openaiModel "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/originaleric/digeino/config"
)

// NewChatModelFromConfig 从配置创建 ChatModel（参考 DigFlow 的配置方式）
// 根据 config.yaml 中的 ChatModel.Type 选择对应的模型提供商
func NewChatModelFromConfig(ctx context.Context) (model.ChatModel, error) {
	cfg := config.Get()
	chatModelCfg := cfg.ChatModel
	modelType := strings.ToLower(strings.TrimSpace(chatModelCfg.Type))
	if modelType == "" {
		return nil, fmt.Errorf("ChatModel.Type is required but not configured")
	}

	// 处理环境变量（参考 DigFlow 的 processEnvVars）
	processedConfig, err := processEnvVars(chatModelCfg.Config)
	if err != nil {
		return nil, err
	}

	switch modelType {
	case "qwen":
		return newQwenModel(ctx, processedConfig)
	case "openai":
		return newOpenAIModel(ctx, processedConfig)
	default:
		return nil, fmt.Errorf("unsupported chat model type: %s. Supported: qwen, openai", chatModelCfg.Type)
	}
}

// processEnvVars 处理环境变量（参考 DigFlow 的实现）
// 支持 ${VAR_NAME} 格式的环境变量替换
func processEnvVars(config map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for k, v := range config {
		if str, ok := v.(string); ok {
			// 如果是环境变量格式 ${VAR_NAME}
			if len(str) > 3 && str[:2] == "${" && str[len(str)-1:] == "}" {
				envKey := str[2 : len(str)-1]
				if envVal := os.Getenv(envKey); envVal != "" {
					result[k] = envVal
				} else {
					return nil, fmt.Errorf("ChatModel.Config.%s references missing environment variable %s", k, envKey)
				}
			} else {
				result[k] = v
			}
		} else {
			result[k] = v
		}
	}

	return result, nil
}

// getString 获取字符串配置值
func getString(config map[string]interface{}, key string, defaultValue string) string {
	if val, ok := config[key]; ok {
		if str, ok := val.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return defaultValue
}

// requireString 获取必填字符串配置值，避免隐式使用任何模型或服务地址兜底。
func requireString(config map[string]interface{}, key string) (string, error) {
	val := getString(config, key, "")
	if val == "" {
		return "", fmt.Errorf("ChatModel.Config.%s is required but not configured", key)
	}
	return val, nil
}

// getFloat32Ptr 获取 float32 指针配置值
func getFloat32Ptr(config map[string]interface{}, key string, defaultValue float32) *float32 {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case float32:
			return &v
		case float64:
			f32 := float32(v)
			return &f32
		}
	}
	return &defaultValue
}

// newQwenModel 创建 Qwen 模型
func newQwenModel(ctx context.Context, cfg map[string]interface{}) (model.ChatModel, error) {
	apiKey, err := requireString(cfg, "ApiKey")
	if err != nil {
		return nil, err
	}
	baseURL, err := requireString(cfg, "BaseUrl")
	if err != nil {
		return nil, err
	}
	modelName, err := requireString(cfg, "Model")
	if err != nil {
		return nil, err
	}

	chatModelConfig := &openaiModel.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
		ExtraFields: map[string]any{
			"enable_thinking": false,
		},
	}

	// 处理 Temperature（如果配置了）
	if temp := getFloat32Ptr(cfg, "Temperature", 0.7); temp != nil {
		chatModelConfig.Temperature = temp
	}

	chatModel, err := openaiModel.NewChatModel(ctx, chatModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Qwen model: %w", err)
	}

	return chatModel, nil
}

// newOpenAIModel 创建 OpenAI 模型
func newOpenAIModel(ctx context.Context, cfg map[string]interface{}) (model.ChatModel, error) {
	apiKey, err := requireString(cfg, "ApiKey")
	if err != nil {
		return nil, err
	}
	baseURL, err := requireString(cfg, "BaseUrl")
	if err != nil {
		return nil, err
	}
	modelName, err := requireString(cfg, "Model")
	if err != nil {
		return nil, err
	}

	chatModelConfig := &openaiModel.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	}

	// 处理 Temperature（如果配置了）
	if temp := getFloat32Ptr(cfg, "Temperature", 0.7); temp != nil {
		chatModelConfig.Temperature = temp
	}

	chatModel, err := openaiModel.NewChatModel(ctx, chatModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI model: %w", err)
	}

	return chatModel, nil
}
