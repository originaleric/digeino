package executor

import (
	"github.com/originaleric/digeino/gateway/artifact"
	"github.com/originaleric/digeino/gateway/registry"
	xplatform "github.com/originaleric/digeino/tools/platform/x"
)

const ToolXPostRead = "x.post.read"

var xDefaultDomains = []string{"x.com", "twitter.com"}

// XPostReadEntry returns registry entry for x.post.read.
func XPostReadEntry(configDomains []string, artStore artifact.Store) registry.Entry {
	return platformReadEntry(
		ToolXPostRead,
		"Read an X (Twitter) post via local browser with structured output.",
		[]string{"browser", "x", "cookie.local", "platform.read"},
		xDefaultDomains,
		configDomains,
		artStore,
		xplatform.ReadPost,
	)
}
