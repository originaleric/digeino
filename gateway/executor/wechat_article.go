package executor

import (
	"github.com/originaleric/digeino/gateway/artifact"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/tools/platform/wechat"
)

const ToolWechatArticleRead = "wechat.article.read"

var wechatDefaultDomains = []string{"mp.weixin.qq.com", "weixin.qq.com"}

// WechatArticleReadEntry returns registry entry for wechat.article.read.
func WechatArticleReadEntry(configDomains []string, artStore artifact.Store) registry.Entry {
	return platformReadEntry(
		ToolWechatArticleRead,
		"Read a WeChat official account article via local browser with cookie reuse.",
		[]string{"browser", "wechat", "cookie.local", "platform.read"},
		wechatDefaultDomains,
		configDomains,
		artStore,
		wechat.ReadArticle,
	)
}
