package rendereino

import (
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/originaleric/digeino/pkg/render"
)

func TestMessageToRenderableTextToolCalls(t *testing.T) {
	msg := &schema.Message{
		Content: "answer",
		ToolCalls: []schema.ToolCall{
			{Function: schema.FunctionCall{Name: "fn", Arguments: `{"a":1}`}},
		},
	}
	s := MessageToRenderableText(msg)
	if !strings.Contains(s, "fn") || !strings.Contains(s, "answer") {
		t.Fatalf("%q", s)
	}
	blocks, err := ParseMessage(msg, render.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) < 1 {
		t.Fatalf("%+v", blocks)
	}
}

func TestMessageToRenderableTextReasoning(t *testing.T) {
	msg := &schema.Message{
		ReasoningContent: "plan",
		Content:          "out",
	}
	s := MessageToRenderableText(msg)
	if !strings.Contains(s, "plan") || !strings.Contains(s, "out") {
		t.Fatalf("%q", s)
	}
	blocks, err := ParseMessage(msg, render.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var sawThink bool
	for _, b := range blocks {
		if b.Kind == render.BlockKindThinking && strings.Contains(b.Content, "plan") {
			sawThink = true
		}
	}
	if !sawThink {
		t.Fatalf("%+v", blocks)
	}
}
