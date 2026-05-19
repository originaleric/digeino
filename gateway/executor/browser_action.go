package executor

import (
	"context"

	"github.com/originaleric/digeino/gateway/protocol"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/tools/research"
)

const ToolBrowserAction = "browser.action"

// BrowserActionEntry returns registry entry for browser.action.
func BrowserActionEntry(configDomains []string) registry.Entry {
	return registry.Entry{
		Descriptor: protocol.ToolDescriptor{
			Name:        ToolBrowserAction,
			Description: "Perform browser interactions (click, type, scroll, etc.) on a page.",
			InputSchema: registry.MustSchema(map[string]any{
				"type":     "object",
				"required": []string{"url", "action"},
				"properties": map[string]any{
					"url":               map[string]string{"type": "string"},
					"ref":               map[string]string{"type": "string"},
					"selector":          map[string]string{"type": "string"},
					"action":            map[string]string{"type": "string"},
					"text":              map[string]string{"type": "string"},
					"key":               map[string]string{"type": "string"},
					"use_cookie_domain": map[string]string{"type": "string"},
				},
			}),
			Capabilities:         []string{"browser", "web.interact", "cookie.local"},
			Risk:                 "network",
			RequiresUserApproval: true,
		},
		Handler: func(ctx context.Context, call *protocol.ToolCall) (map[string]any, []protocol.Artifact, error) {
			in, err := decodeInput[research.BrowserActionRequest](call)
			if err != nil {
				return nil, nil, err
			}
			if err := validateCallURL(in.URL, call, configDomains); err != nil {
				return nil, nil, err
			}
			if err := validateCookieDomain(in.UseCookieDomain, call, configDomains); err != nil {
				return nil, nil, err
			}
			resp, err := research.BrowserAction(ctx, &in)
			if err != nil {
				return nil, nil, err
			}
			return map[string]any{
				"success": resp.Success,
				"message": resp.Message,
				"url":     resp.URL,
			}, nil, nil
		},
	}
}
