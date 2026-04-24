// Package rendereino adapts cloudwego/eino/schema.Message to pkg/render inputs.
package rendereino

import (
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/originaleric/digeino/pkg/render"
)

// MessageToRenderableText builds UTF-8 text suitable for render.Parse.
// It prepends ReasoningContent as a thinking-tagged region when present,
// then Content, then a short appendix for ToolCalls when present.
func MessageToRenderableText(msg *schema.Message) string {
	if msg == nil {
		return ""
	}
	var b strings.Builder
	if msg.ReasoningContent != "" {
		b.WriteString("<think>\n")
		b.WriteString(msg.ReasoningContent)
		b.WriteString("\n</think>\n\n")
	}
	b.WriteString(msg.Content)
	if len(msg.ToolCalls) > 0 {
		b.WriteString("\n\n<!-- tool_calls -->\n")
		for _, tc := range msg.ToolCalls {
			b.WriteString(tc.Function.Name)
			b.WriteString(": ")
			b.WriteString(tc.Function.Arguments)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// ParseMessage runs render.Parse on MessageToRenderableText(msg).
func ParseMessage(msg *schema.Message, opts render.Options) ([]render.Block, error) {
	return render.Parse(MessageToRenderableText(msg), opts)
}
