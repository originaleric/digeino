package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptionsFromYAML_OmittedUsesNilSlice(t *testing.T) {
	opt, err := OptionsFromYAML([]byte("# only comments\n"))
	if err != nil {
		t.Fatal(err)
	}
	if opt.ThinkingTagPairs != nil {
		t.Fatalf("expected nil, got %v", opt.ThinkingTagPairs)
	}
}

func TestOptionsFromYAML_EmptyListDisablesThinking(t *testing.T) {
	opt, err := OptionsFromYAML([]byte("thinking_tag_pairs: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	if opt.ThinkingTagPairs == nil {
		t.Fatal("expected non-nil empty slice for thinking_tag_pairs: []")
	}
	if len(opt.ThinkingTagPairs) != 0 {
		t.Fatalf("expected len 0, got %v", opt.ThinkingTagPairs)
	}
	blocks, err := Parse("a<think>x</think>b", opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Kind != BlockKindMarkdown {
		t.Fatalf("thinking should not split: %+v", blocks)
	}
}

func TestLoadOptionsFromFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(p, []byte(`thinking_tag_pairs:
  - open: "<x>"
    close: "</x>"
`), 0644); err != nil {
		t.Fatal(err)
	}
	opt, err := LoadOptionsFromFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(opt.ThinkingTagPairs) != 1 || opt.ThinkingTagPairs[0].Open != "<x>" {
		t.Fatalf("%+v", opt.ThinkingTagPairs)
	}
}

func TestLoadEmbeddedDefaultOptions(t *testing.T) {
	opt, err := LoadEmbeddedDefaultOptions()
	if err != nil {
		t.Fatal(err)
	}
	if len(opt.ThinkingTagPairs) < 1 {
		t.Fatalf("%+v", opt.ThinkingTagPairs)
	}
}

func TestDefaultThinkingTagPairsMatchesEmbed(t *testing.T) {
	emb, err := LoadEmbeddedDefaultOptions()
	if err != nil {
		t.Fatal(err)
	}
	def := DefaultThinkingTagPairs()
	if len(def) != len(emb.ThinkingTagPairs) {
		t.Fatalf("len default=%d embed=%d", len(def), len(emb.ThinkingTagPairs))
	}
}

func TestOptionsFromYAML_InvalidPair(t *testing.T) {
	_, err := OptionsFromYAML([]byte(`thinking_tag_pairs:
  - open: ""
    close: "</x>"
`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRenderConfigFromYAML_HTMLMerge(t *testing.T) {
	y := []byte(`thinking_tag_pairs: []
html:
  code:
    pre_class: my-pre
`)
	rc, err := RenderConfigFromYAML(y)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Render.Code.Outer.Class != "my-pre" || rc.Render.Code.Outer.Tag != "pre" {
		t.Fatalf("code outer: %+v", rc.Render.Code.Outer)
	}
	if rc.Render.Thinking.Outer.Class == "" {
		t.Fatal("expected default thinking class merged")
	}
}

func TestRenderConfigFromYAML_NewShape(t *testing.T) {
	y := []byte(`thinking_tag_pairs: []
render:
  thinking:
    outer: { tag: div, class: t-outer }
    inner: { tag: pre, class: t-inner }
  code:
    outer: { tag: pre, class: c-pre }
    inner: { tag: code, language_class_prefix: lang- }
  markdown:
    sanitize: strict
`)
	rc, err := RenderConfigFromYAML(y)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Render.Thinking.Outer.Tag != "div" || rc.Render.Markdown.Sanitize != "strict" {
		t.Fatalf("%+v", rc.Render)
	}
	if rc.ParseRender.CodeFence != "code" || rc.ParseRender.ThinkingTagPairs != "thinking" {
		t.Fatalf("parse_render %+v", rc.ParseRender)
	}
}

func TestRenderConfigFromYAML_CustomProfileKeys(t *testing.T) {
	y := []byte(`thinking_tag_pairs: []
parse_render:
  thinking_tag_pairs: reasoning
  code_fence: snippet
  markdown: prose
render:
  reasoning:
    outer: { tag: div, class: r-out }
  snippet:
    outer: { tag: pre, class: s-pre }
  prose:
    sanitize: strict
`)
	rc, err := RenderConfigFromYAML(y)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Render.Thinking.Outer.Class != "r-out" || rc.Render.Code.Outer.Class != "s-pre" {
		t.Fatalf("thinking/code %+v %+v", rc.Render.Thinking, rc.Render.Code)
	}
	if rc.Render.Markdown.Sanitize != "strict" {
		t.Fatalf("markdown %+v", rc.Render.Markdown)
	}
}

func TestRenderConfigFromYAML_ParseRenderDuplicate(t *testing.T) {
	_, err := RenderConfigFromYAML([]byte(`thinking_tag_pairs: []
parse_render:
  thinking_tag_pairs: x
  code_fence: x
  markdown: m
render:
  x: {}
  m: { sanitize: ugc }
`))
	if err == nil {
		t.Fatal("expected error for duplicate render profile key")
	}
}

func TestRenderConfigFromYAML_RenderOverridesHTML(t *testing.T) {
	y := []byte(`thinking_tag_pairs: []
html:
  code:
    outer: { tag: pre, class: from-html }
render:
  code:
    outer: { tag: pre, class: from-render }
`)
	rc, err := RenderConfigFromYAML(y)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Render.Code.Outer.Class != "from-render" {
		t.Fatalf("expected render to win, got %+v", rc.Render.Code.Outer)
	}
}

func TestOptionsFromYAML_ParseThinkingOverridesRoot(t *testing.T) {
	opt, err := OptionsFromYAML([]byte(`thinking_tag_pairs:
  - open: "<a>"
    close: "</a>"
parse:
  thinking_tag_pairs:
    - open: "<b>"
      close: "</b>"
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(opt.ThinkingTagPairs) != 1 || opt.ThinkingTagPairs[0].Open != "<b>" {
		t.Fatalf("%+v", opt.ThinkingTagPairs)
	}
}

func TestOptionsFromYAML_ParseThinkingEmptyOverridesRoot(t *testing.T) {
	opt, err := OptionsFromYAML([]byte(`thinking_tag_pairs:
  - open: "<a>"
    close: "</a>"
parse:
  thinking_tag_pairs: []
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(opt.ThinkingTagPairs) != 0 {
		t.Fatalf("expected empty, got %+v", opt.ThinkingTagPairs)
	}
}

func TestOptionsFromYAML_ParseExtraTagKey(t *testing.T) {
	opt, err := OptionsFromYAML([]byte(`thinking_tag_pairs:
  - open: "<a>"
    close: "</a>"
parse:
  my_tag:
    open: "<my-tag>"
    close: "</my-tag>"
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(opt.ThinkingTagPairs) != 2 {
		t.Fatalf("want 2 pairs, got %+v", opt.ThinkingTagPairs)
	}
	if opt.ThinkingTagPairs[1].Open != "<my-tag>" {
		t.Fatalf("%+v", opt.ThinkingTagPairs)
	}
	blocks, err := Parse("x<my-tag>y</my-tag>z", opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 3 || blocks[1].Kind != BlockKindThinking || blocks[1].Content != "y" {
		t.Fatalf("%+v", blocks)
	}
}

func TestRenderConfig_ParseRenderExtraRequiresParseBlock(t *testing.T) {
	_, err := RenderConfigFromYAML([]byte(`thinking_tag_pairs: []
parse_render:
  thinking_tag_pairs: thinking
  code_fence: code
  markdown: markdown
  orphan: thinking
`))
	if err == nil {
		t.Fatal("expected error for parse_render.orphan without parse.orphan")
	}
}

func TestOptionsFromYAML_ParseCodeFence(t *testing.T) {
	opt, err := OptionsFromYAML([]byte(`parse:
  code_fence:
    open: "~~~"
    close: "~~~"
`))
	if err != nil {
		t.Fatal(err)
	}
	if opt.CodeFence.Open != "~~~" || opt.CodeFence.Close != "~~~" {
		t.Fatalf("%+v", opt.CodeFence)
	}
}

func TestOptionsFromYAML_ParseCodeFenceDeprecatedOpening(t *testing.T) {
	opt, err := OptionsFromYAML([]byte(`parse:
  code_fence:
    opening: "~~~"
`))
	if err != nil {
		t.Fatal(err)
	}
	if opt.CodeFence.Open != "~~~" || opt.CodeFence.Close != "~~~" {
		t.Fatalf("%+v", opt.CodeFence)
	}
}

func TestOptionsFromYAML_CustomPairParse(t *testing.T) {
	opt, err := OptionsFromYAML([]byte(`thinking_tag_pairs:
  - open: "<think>"
    close: "</think>"
`))
	if err != nil {
		t.Fatal(err)
	}
	blocks, err := Parse("pre<think>mid</think>post", opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 3 || blocks[1].Kind != BlockKindThinking || !strings.Contains(blocks[1].Content, "mid") {
		t.Fatalf("%+v", blocks)
	}
}
