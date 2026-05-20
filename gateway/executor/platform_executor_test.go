package executor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/originaleric/digeino/gateway/protocol"
	"github.com/originaleric/digeino/tools/platform"
)

func TestPlatformReadInvalidURL(t *testing.T) {
	t.Parallel()
	entry := XiaohongshuNoteReadEntry([]string{"xiaohongshu.com"}, nil)
	raw, _ := json.Marshal(map[string]any{"url": "https://evil.example/note"})
	call := &protocol.ToolCall{
		ID:    "call_1",
		Tool:  ToolXiaohongshuNoteRead,
		Input: raw,
		Policy: protocol.CallPolicy{
			AllowedDomains: []string{"xiaohongshu.com"},
		},
	}
	_, _, err := entry.Handler(context.Background(), call)
	if err == nil {
		t.Fatal("expected domain error")
	}
}

func TestPlatformReadInvalidInput(t *testing.T) {
	t.Parallel()
	entry := DouyinVideoReadEntry(nil, nil)
	call := &protocol.ToolCall{
		ID:    "call_2",
		Tool:  ToolDouyinVideoRead,
		Input: json.RawMessage(`{}`),
	}
	_, _, err := entry.Handler(context.Background(), call)
	if err == nil {
		t.Fatal("expected invalid input")
	}
}

func TestPlatformOutputSchemaRegistered(t *testing.T) {
	t.Parallel()
	names := []string{ToolWechatArticleRead, ToolXiaohongshuNoteRead, ToolDouyinVideoRead, ToolXPostRead}
	entries := []struct {
		name  string
		entry func() protocol.ToolDescriptor
	}{
		{ToolWechatArticleRead, func() protocol.ToolDescriptor { return WechatArticleReadEntry(nil, nil).Descriptor }},
		{ToolXiaohongshuNoteRead, func() protocol.ToolDescriptor { return XiaohongshuNoteReadEntry(nil, nil).Descriptor }},
		{ToolDouyinVideoRead, func() protocol.ToolDescriptor { return DouyinVideoReadEntry(nil, nil).Descriptor }},
		{ToolXPostRead, func() protocol.ToolDescriptor { return XPostReadEntry(nil, nil).Descriptor }},
	}
	for i, e := range entries {
		if e.entry().Name != names[i] {
			t.Fatalf("tool name mismatch: got %s want %s", e.entry().Name, names[i])
		}
	}
	_ = platform.NowRFC3339()
}
