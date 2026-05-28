package webhook

import "github.com/originaleric/digeino/config"

// NewConfiguredCollector 创建带默认配置注入的 StatusCollector。
// 统一入口（HTTP/CLI/批处理）可复用此函数，避免各自拼装导致语义不一致。
//
// 规则：
// 1. callback 非空时启用回调 sink；
// 2. store 非空且 Store 配置启用时启用存储 sink；
// 3. webhook 配置可用且 URL 非空时启用 webhook sink；
// 4. 若三类 sink 都不可用，则返回 nil。
func NewConfiguredCollector(
	executionID, appName, requestID string,
	store StatusStoreInterface,
	callback func(status ExecutionStatus),
	buildDefaultURL func() string,
) *StatusCollector {
	webhookCfg := GetWebhookConfig(buildDefaultURL)
	if webhookCfg != nil && webhookCfg.URL == "" {
		// 没有可用 URL 时不启用 webhook sink，避免运行期无意义报错。
		webhookCfg = nil
	}

	storeEnabled := IsStoreEnabled() && store != nil
	callbackEnabled := callback != nil
	webhookEnabled := webhookCfg != nil
	feishuCfg := GetFeishuAPIConfig()
	feishuEnabled := feishuCfg != nil
	weChatEnabled := GetWeChatConfig() != nil
	weComEnabled := GetWeComConfig() != nil

	if !storeEnabled && !callbackEnabled && !webhookEnabled && !feishuEnabled && !weChatEnabled && !weComEnabled {
		return nil
	}

	collector := NewStatusCollector(executionID, appName, requestID)
	if callbackEnabled {
		collector.SetStatusCallback(callback)
	}
	if storeEnabled {
		collector.SetStatusStore(store)
	}
	if webhookEnabled {
		collector.AddWebhookClient(NewWebhookClient(webhookCfg))
	}
	if feishuEnabled {
		collector.AddNotifier(NewFeishuNotifier(NewFeishuClient(*feishuCfg), config.Get().Feishu.NotifyOnEvents))
	}
	if weChatEnabled {
		collector.AddNotifier(NewWeChatNotifier(config.Get().WeChat.NotifyOnEvents))
	}
	if weComEnabled {
		collector.AddNotifier(NewWeComNotifier(config.Get().WeCom.NotifyOnEvents))
	}
	return collector
}
