package render

import "testing"

func TestParseStablePrefixIncompleteFenceTilde(t *testing.T) {
	in := "before\n~~~go\nx := 1"
	res, err := ParseStablePrefix(in, Options{
		ThinkingTagPairs: []ThinkingTagPair{},
		CodeFence:        CodeFenceConfig{Open: "~~~", Close: "~~~"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Blocks) != 1 || res.Blocks[0].Kind != BlockKindMarkdown {
		t.Fatalf("blocks %+v", res.Blocks)
	}
}

func TestParseStablePrefixIncompleteFence(t *testing.T) {
	in := "before\n```go\nx := 1"
	res, err := ParseStablePrefix(in, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Blocks) != 1 || res.Blocks[0].Kind != BlockKindMarkdown {
		t.Fatalf("blocks %+v", res.Blocks)
	}
	if res.Remainder != "```go\nx := 1" && res.Remainder != "before\n```go\nx := 1" {
		// remainder should start at opening fence line
		t.Fatalf("remainder %q", res.Remainder)
	}
}

func TestParseStablePrefixComplete(t *testing.T) {
	in := "ok\n\n```\nc\n```\n\ndone"
	res, err := ParseStablePrefix(in, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Remainder != "" {
		t.Fatalf("remainder %q", res.Remainder)
	}
	if len(res.Blocks) < 2 {
		t.Fatalf("blocks %+v", res.Blocks)
	}
}

func TestParseStablePrefixUnclosedThinking(t *testing.T) {
	in := "hi<think>tail"
	res, err := ParseStablePrefix(in, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Blocks) != 1 || res.Blocks[0].Content != "hi" {
		t.Fatalf("blocks %+v", res.Blocks)
	}
	if res.Remainder != "<think>tail" {
		t.Fatalf("remainder %q", res.Remainder)
	}
}
