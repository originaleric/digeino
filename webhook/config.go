package webhook

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/originaleric/digeino/config"
)

var feishuCredentialsMissingWarn sync.Once

// GetWebhookConfig 从全局配置获取 WebhookConfig
func GetWebhookConfig(buildDefaultURL func() string) *config.WebhookConfig {
	cfg := config.Get().Status.Webhook
	if cfg.Enabled != nil && !*cfg.Enabled {
		return nil
	}

	// 这里需要注意，Config 是 WebhookConfig 类型
	webhookConfig := &cfg.Config
	if webhookConfig.URL == "" {
		if cfg.URL != "" {
			webhookConfig.URL = cfg.URL
		} else if buildDefaultURL != nil {
			webhookConfig.URL = buildDefaultURL()
		}
	}

	if webhookConfig.Method == "" {
		webhookConfig.Method = "POST"
	}
	if webhookConfig.Timeout == 0 {
		webhookConfig.Timeout = 5
	}
	if webhookConfig.RetryCount == 0 {
		webhookConfig.RetryCount = 3
	}
	if webhookConfig.RetryDelay == 0 {
		webhookConfig.RetryDelay = 1000
	}

	return webhookConfig
}

// IsStoreEnabled 检查 Store 是否启用
func IsStoreEnabled() bool {
	cfg := config.Get().Status.Store
	if cfg.Enabled != nil && !*cfg.Enabled {
		return false
	}

	return true
}

// BuildDefaultWebhookURL 构建默认的本地 webhook URL
func BuildDefaultWebhookURL(scheme, host string) string {
	if host == "" {
		port := config.Get().HttpServer.Api.Port
		if port == "" {
			host = "localhost"
		} else {
			if strings.HasPrefix(port, ":") {
				host = "localhost" + port
			} else {
				host = "localhost:" + port
			}
		}
	}

	return fmt.Sprintf("%s://%s/api/v1/webhook/status", scheme, host)
}

// GetFeishuAPIConfig 获取飞书 API 配置（仅用于消息发送）。
func GetFeishuAPIConfig() *config.FeishuAPIConfig {
	cfg := config.Get().Feishu
	if cfg.Enabled != nil && !*cfg.Enabled {
		return nil
	}
	if cfg.SendViaAPI != nil && !*cfg.SendViaAPI {
		return nil
	}

	apiCfg := cfg.API
	if apiCfg.Enabled != nil && !*apiCfg.Enabled {
		return nil
	}
	if strings.TrimSpace(apiCfg.BaseURL) == "" {
		apiCfg.BaseURL = "https://open.feishu.cn"
	}
	if apiCfg.Timeout <= 0 {
		apiCfg.Timeout = 5
	}
	if apiCfg.RetryDelayMs <= 0 {
		apiCfg.RetryDelayMs = 500
	}
	if apiCfg.ReceiveIDType == "" {
		apiCfg.ReceiveIDType = "chat_id"
	}
	if strings.TrimSpace(apiCfg.AppID) == "" || strings.TrimSpace(apiCfg.AppSecret) == "" {
		feishuCredentialsMissingWarn.Do(func() {
			log.Printf("digeino: Feishu enabled but AppID/AppSecret empty, runtime notification sink disabled")
		})
		return nil
	}
	return &apiCfg
}

func GetWeChatConfig() *config.WeChatConfig {
	cfg := config.Get().WeChat
	if cfg.Enabled != nil && !*cfg.Enabled {
		return nil
	}
	if strings.TrimSpace(cfg.AppID) == "" || strings.TrimSpace(cfg.AppSecret) == "" {
		return nil
	}
	if len(cfg.OpenIDs) == 0 {
		return nil
	}
	return &cfg
}

func GetWeComConfig() *config.WeComConfig {
	cfg := config.Get().WeCom
	if cfg.Enabled != nil && !*cfg.Enabled {
		return nil
	}
	if strings.TrimSpace(cfg.CorpID) == "" {
		return nil
	}
	for _, app := range cfg.Applications {
		if app.AgentID > 0 && strings.TrimSpace(app.AgentSecret) != "" && strings.TrimSpace(app.ToUser) != "" {
			return &cfg
		}
	}
	return nil
}
