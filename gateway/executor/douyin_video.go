package executor

import (
	"github.com/originaleric/digeino/gateway/artifact"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/tools/platform/douyin"
)

const ToolDouyinVideoRead = "douyin.video.read"

var douyinDefaultDomains = []string{"douyin.com", "www.douyin.com", "iesdouyin.com"}

// DouyinVideoReadEntry returns registry entry for douyin.video.read.
func DouyinVideoReadEntry(configDomains []string, artStore artifact.Store) registry.Entry {
	return platformReadEntry(
		ToolDouyinVideoRead,
		"Read a Douyin video page via local browser with structured output.",
		[]string{"browser", "douyin", "cookie.local", "platform.read"},
		douyinDefaultDomains,
		configDomains,
		artStore,
		douyin.ReadVideo,
	)
}
