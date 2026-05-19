package executor

import (
	"context"

	"github.com/originaleric/digeino/gateway/artifact"
	"github.com/originaleric/digeino/gateway/protocol"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/tools/research"
)

const ToolBrowserBrowse = "browser.browse"

type browserBrowseInput struct {
	URL             string `json:"url"`
	Action          string `json:"action,omitempty"`
	Mode            string `json:"mode,omitempty"`
	TabID           string `json:"tab_id,omitempty"`
	WaitSelector    string `json:"wait_selector,omitempty"`
	ContentSelector string `json:"content_selector,omitempty"`
	UseCookieDomain string `json:"use_cookie_domain,omitempty"`
}

// BrowserBrowseEntry returns registry entry for browser.browse.
func BrowserBrowseEntry(configDomains []string, artStore artifact.Store) registry.Entry {
	return registry.Entry{
		Descriptor: protocol.ToolDescriptor{
			Name:        ToolBrowserBrowse,
			Description: "Open a URL in a local browser and extract rendered content.",
			InputSchema: registry.MustSchema(map[string]any{
				"type":     "object",
				"required": []string{"url"},
				"properties": map[string]any{
					"url":               map[string]string{"type": "string"},
					"action":            map[string]string{"type": "string"},
					"mode":              map[string]string{"type": "string"},
					"wait_selector":     map[string]string{"type": "string"},
					"content_selector":  map[string]string{"type": "string"},
					"use_cookie_domain": map[string]string{"type": "string"},
				},
			}),
			OutputSchema: registry.MustSchema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":      map[string]string{"type": "string"},
					"text":       map[string]string{"type": "string"},
					"markdown":   map[string]string{"type": "string"},
					"source_url": map[string]string{"type": "string"},
				},
			}),
			Capabilities: []string{"browser", "web.read", "cookie.local"},
			Risk:         "network",
		},
		Handler: func(ctx context.Context, call *protocol.ToolCall) (map[string]any, []protocol.Artifact, error) {
			in, err := decodeInput[browserBrowseInput](call)
			if err != nil {
				return nil, nil, err
			}
			if err := validateCallURL(in.URL, call, configDomains); err != nil {
				return nil, nil, err
			}
			if err := validateCookieDomain(in.UseCookieDomain, call, configDomains); err != nil {
				return nil, nil, err
			}

			resp, err := research.BrowserBrowse(ctx, &research.BrowserBrowseRequest{
				URL:             in.URL,
				Action:          in.Action,
				Mode:            in.Mode,
				TabID:           in.TabID,
				WaitSelector:    in.WaitSelector,
				ContentSelector: in.ContentSelector,
				UseCookieDomain: in.UseCookieDomain,
			})
			if err != nil {
				return nil, nil, err
			}

			out := map[string]any{
				"source_url": resp.URL,
				"title":      resp.Title,
			}
			if resp.Text != "" {
				out["text"] = resp.Text
			}
			if resp.Markdown != "" {
				out["markdown"] = resp.Markdown
			}
			if resp.TabID != "" {
				out["tab_id"] = resp.TabID
			}
			if resp.Summary != "" {
				out["summary"] = resp.Summary
			}
			var artifacts []protocol.Artifact
			if resp.ScreenshotBase != "" {
				artID := call.ID + "_screenshot"
				if artStore != nil {
					art, err := artifact.PutBase64PNG(ctx, artStore, artID, resp.ScreenshotBase)
					if err == nil {
						artifacts = append(artifacts, art)
						out["screenshot_artifact_id"] = art.ID
					}
				} else {
					artifacts = append(artifacts, protocol.Artifact{
						ID:   artID,
						Type: "image/png",
						Name: "page.png",
						URI:  "digeino-artifact://" + artID,
					})
					out["screenshot_artifact_id"] = artID
				}
			}
			return out, artifacts, nil
		},
	}
}
