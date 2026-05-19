package executor

import (
	"context"

	"github.com/originaleric/digeino/gateway/protocol"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/tools/research"
)

const ToolBrowserSnapshot = "browser.snapshot"

// BrowserSnapshotEntry returns registry entry for browser.snapshot.
func BrowserSnapshotEntry(configDomains []string) registry.Entry {
	return registry.Entry{
		Descriptor: protocol.ToolDescriptor{
			Name:        ToolBrowserSnapshot,
			Description: "Capture a structured accessibility snapshot of a page with interactive elements.",
			InputSchema: registry.MustSchema(map[string]any{
				"type":     "object",
				"required": []string{"url"},
				"properties": map[string]any{
					"url":               map[string]string{"type": "string"},
					"filter":            map[string]string{"type": "string"},
					"max_depth":         map[string]string{"type": "integer"},
					"wait_selector":     map[string]string{"type": "string"},
					"use_cookie_domain": map[string]string{"type": "string"},
				},
			}),
			Capabilities: []string{"browser", "web.read", "cookie.local"},
			Risk:         "network",
		},
		Handler: func(ctx context.Context, call *protocol.ToolCall) (map[string]any, []protocol.Artifact, error) {
			in, err := decodeInput[research.BrowserSnapshotRequest](call)
			if err != nil {
				return nil, nil, err
			}
			if err := validateCallURL(in.URL, call, configDomains); err != nil {
				return nil, nil, err
			}
			if err := validateCookieDomain(in.UseCookieDomain, call, configDomains); err != nil {
				return nil, nil, err
			}
			resp, err := research.BrowserSnapshot(ctx, &in)
			if err != nil {
				return nil, nil, err
			}
			return map[string]any{
				"source_url": resp.URL,
				"title":      resp.Title,
				"elements":   resp.Elements,
			}, nil, nil
		},
	}
}
