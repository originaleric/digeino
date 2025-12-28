package webhook

import (
	"fmt"
	"strings"

	"github.com/originaleric/digeino/config"
)

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
