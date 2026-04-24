package render

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseEmpty(t *testing.T) {
	blocks, err := Parse("", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 0 {
		t.Fatalf("got %d blocks", len(blocks))
	}
}

func TestParseMarkdownOnly(t *testing.T) {
	blocks, err := Parse("# Hi\n\npara", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Kind != BlockKindMarkdown {
		t.Fatalf("got %+v", blocks)
	}
	if blocks[0].Content != "# Hi\n\npara" {
		t.Fatalf("content %q", blocks[0].Content)
	}
}

func TestParseCodeFenceCustomOpening(t *testing.T) {
	in := "intro\n\n~~~go\nx\n~~~\n\noutro"
	blocks, err := Parse(in, Options{
		ThinkingTagPairs: []ThinkingTagPair{},
		CodeFence:        CodeFenceConfig{Open: "~~~", Close: "~~~"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 3 {
		t.Fatalf("got %d blocks: %+v", len(blocks), blocks)
	}
	if blocks[1].Kind != BlockKindCode || blocks[1].Language != "go" || blocks[1].Content != "x" {
		t.Fatalf("block1 %+v", blocks[1])
	}
}

func TestParseCodeFence(t *testing.T) {
	in := "intro\n\n```go\nfmt.Println(\"x\")\n```\n\noutro"
	blocks, err := Parse(in, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 3 {
		t.Fatalf("got %d blocks: %+v", len(blocks), blocks)
	}
	if blocks[0].Kind != BlockKindMarkdown || !strings.Contains(blocks[0].Content, "intro") {
		t.Fatalf("block0 %+v", blocks[0])
	}
	if blocks[1].Kind != BlockKindCode || blocks[1].Language != "go" {
		t.Fatalf("block1 %+v", blocks[1])
	}
	if blocks[1].Content != `fmt.Println("x")` {
		t.Fatalf("code %q", blocks[1].Content)
	}
	if blocks[2].Kind != BlockKindMarkdown || !strings.Contains(blocks[2].Content, "outro") {
		t.Fatalf("block2 %+v", blocks[2])
	}
}

func TestParseThinking(t *testing.T) {
	in := "Hello\n<think>\nstep1\n</think>\n\nDone."
	blocks, err := Parse(in, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 3 {
		t.Fatalf("got %d blocks: %+v", len(blocks), blocks)
	}
	if blocks[0].Kind != BlockKindMarkdown || !strings.HasPrefix(blocks[0].Content, "Hello") {
		t.Fatalf("block0 %+v", blocks[0])
	}
	if blocks[1].Kind != BlockKindThinking || strings.TrimSpace(blocks[1].Content) != "step1" {
		t.Fatalf("block1 %+v", blocks[1])
	}
	if blocks[2].Kind != BlockKindMarkdown || !strings.Contains(blocks[2].Content, "Done") {
		t.Fatalf("block2 %+v", blocks[2])
	}
}

func TestParseThinkingUnclosed(t *testing.T) {
	in := "A<think>still typing"
	blocks, err := Parse(in, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("got %+v", blocks)
	}
	if blocks[0].Content != "A" || blocks[1].Kind != BlockKindThinking {
		t.Fatalf("got %+v", blocks)
	}
}

func TestParseCodeThenThinking(t *testing.T) {
	in := "```\nplain\n```\n\n<think>\n\n</think>\n\nok"
	blocks, err := Parse(in, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) < 2 {
		t.Fatalf("got %+v", blocks)
	}
	if blocks[0].Kind != BlockKindCode {
		t.Fatalf("first %+v", blocks[0])
	}
}

func TestParseThinkingInsideCodeIgnored(t *testing.T) {
	in := "```\n<think>fake</think>\n```"
	blocks, err := Parse(in, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].Kind != BlockKindCode {
		t.Fatalf("got %+v", blocks)
	}
	if blocks[0].Content != "<think>fake</think>" {
		t.Fatalf("content %q", blocks[0].Content)
	}
}

func TestOptionsInvalidPair(t *testing.T) {
	_, err := Parse("x", Options{ThinkingTagPairs: []ThinkingTagPair{{Open: "", Close: "x"}}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBlockJSON(t *testing.T) {
	b := Block{Kind: BlockKindCode, Language: "go", Content: "a"}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"kind":"code"`) {
		t.Fatalf("%s", data)
	}
}
