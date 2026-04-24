// Package renderexport writes assistant render output via tempstorage (DigEino integration).
package renderexport

import (
	"context"

	"github.com/originaleric/digeino/pkg/render"
	renderhtml "github.com/originaleric/digeino/pkg/render/html"
	"github.com/originaleric/digeino/tempstorage"
)

// SaveAssistantHTML renders blocks to a self-contained HTML file using tempstorage path rules.
// ctx must carry workspace_path or agent_session_id (see tempstorage).
// filename is relative under the resolved base (e.g. "chat_render/session.html").
func SaveAssistantHTML(ctx context.Context, filename string, blocks []render.Block, title string) (absPath string, err error) {
	if title == "" {
		title = "Assistant"
	}
	body, err := renderhtml.BlocksToHTML(blocks)
	if err != nil {
		return "", err
	}
	doc := renderhtml.WrapDocument(title, body)
	return tempstorage.SaveForReview(ctx, filename, doc)
}
