package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/originaleric/digeino/gateway/policy"
	"github.com/originaleric/digeino/gateway/protocol"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/tools/research"
)

const ToolWechatArticleRead = "wechat.article.read"

type wechatArticleInput struct {
	URL    string   `json:"url"`
	Format []string `json:"format,omitempty"`
}

// WechatArticleReadEntry returns registry entry for wechat.article.read.
func WechatArticleReadEntry(configDomains []string) registry.Entry {
	return registry.Entry{
		Descriptor: protocol.ToolDescriptor{
			Name:        ToolWechatArticleRead,
			Description: "Read a WeChat official account article via local browser with cookie reuse.",
			InputSchema: registry.MustSchema(map[string]any{
				"type":     "object",
				"required": []string{"url"},
				"properties": map[string]any{
					"url":    map[string]string{"type": "string"},
					"format": map[string]any{"type": "array", "items": map[string]string{"type": "string"}},
				},
			}),
			OutputSchema: registry.MustSchema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":      map[string]string{"type": "string"},
					"text":       map[string]string{"type": "string"},
					"markdown":   map[string]string{"type": "string"},
					"source_url": map[string]string{"type": "string"},
					"metadata":   map[string]string{"type": "object"},
				},
			}),
			Capabilities:         []string{"browser", "wechat", "cookie.local"},
			Risk:                 "network",
			RequiresUserApproval: false,
		},
		Handler: func(ctx context.Context, call *protocol.ToolCall) (map[string]any, []protocol.Artifact, error) {
			var in wechatArticleInput
			if err := json.Unmarshal(call.Input, &in); err != nil {
				return nil, nil, fmt.Errorf("%s: %w", policy.CodeInvalidInput, err)
			}
			domains := policy.MergeDomains(&call.Policy, configDomains)
			if len(domains) == 0 {
				domains = []string{"mp.weixin.qq.com", "weixin.qq.com"}
			}
			if err := policy.ValidateURLDomain(in.URL, domains, nil); err != nil {
				return nil, nil, err
			}

			waitSelector := "#js_article"
			contentSelector := "#js_article"
			resp, err := research.BrowserBrowse(ctx, &research.BrowserBrowseRequest{
				URL:             in.URL,
				Action:          "read",
				Mode:            "full",
				WaitSelector:    waitSelector,
				ContentSelector: contentSelector,
				UseCookieDomain: "mp.weixin.qq.com",
			})
			if err != nil {
				// Fallback: try #js_content container used by some article layouts.
				resp, err = research.BrowserBrowse(ctx, &research.BrowserBrowseRequest{
					URL:             in.URL,
					Action:          "read",
					Mode:            "full",
					WaitSelector:    "#js_content",
					ContentSelector: "#js_content",
					UseCookieDomain: "mp.weixin.qq.com",
				})
				if err != nil {
					return nil, nil, err
				}
			}

			formats := normalizeFormats(in.Format)
			out := map[string]any{
				"source_url": resp.URL,
				"title":      resp.Title,
				"metadata": map[string]any{
					"captured_at": time.Now().Format(time.RFC3339),
				},
			}
			if containsFormat(formats, "text") {
				out["text"] = resp.Text
			}
			if containsFormat(formats, "markdown") {
				out["markdown"] = resp.Markdown
			}
			if len(formats) == 0 {
				out["text"] = resp.Text
				out["markdown"] = resp.Markdown
			}
			return out, nil, nil
		},
	}
}

func normalizeFormats(formats []string) []string {
	out := make([]string, 0, len(formats))
	for _, f := range formats {
		f = strings.ToLower(strings.TrimSpace(f))
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func containsFormat(formats []string, want string) bool {
	for _, f := range formats {
		if f == want {
			return true
		}
	}
	return false
}
