package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/originaleric/digeino/tools/research"
)

// Browse 使用本地浏览器按平台规格采集页面。
func Browse(ctx context.Context, url string, in ReadInput, spec BrowseSpec) (*research.BrowserBrowseResponse, error) {
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("url 不能为空")
	}
	action := strings.TrimSpace(spec.Action)
	if action == "" {
		if in.IncludeScreenshot {
			action = "screenshot"
		} else {
			action = "read"
		}
	}
	cookieDomain := strings.TrimSpace(in.UseCookieDomain)
	if cookieDomain == "" {
		cookieDomain = spec.CookieDomain
	}
	waitSelector := strings.TrimSpace(in.WaitSelector)
	if waitSelector == "" {
		waitSelector = spec.WaitSelector
	}
	return research.BrowserBrowse(ctx, &research.BrowserBrowseRequest{
		URL:             url,
		Action:          action,
		Mode:            "full",
		WaitSelector:    waitSelector,
		ContentSelector: spec.ContentSelector,
		UseCookieDomain: cookieDomain,
		MetadataScript:  spec.MetadataScript,
	})
}
