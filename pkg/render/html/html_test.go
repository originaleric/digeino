package html

import (
	"strings"
	"testing"

	"github.com/originaleric/digeino/pkg/render"
)

func TestBlocksToHTML(t *testing.T) {
	blocks := []render.Block{
		{Kind: render.BlockKindMarkdown, Content: "# T\n\nhi"},
		{Kind: render.BlockKindCode, Language: "go", Content: `fmt.Println("x")`},
	}
	out, err := BlocksToHTML(blocks)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<h1") || !strings.Contains(out, "T") {
		t.Fatalf("missing heading: %s", out)
	}
	if !strings.Contains(out, `language-go`) || !strings.Contains(out, `fmt.Println`) {
		t.Fatalf("missing code: %s", out)
	}
}

func TestWrapDocument(t *testing.T) {
	s := WrapDocument(`a<b>`, "<p>x</p>")
	if !strings.Contains(s, "<title>a&lt;b&gt;</title>") {
		t.Fatalf("title not escaped: %s", s)
	}
	if !strings.Contains(s, `class="llm-render-doc"`) || !strings.Contains(s, "<p>x</p>") {
		t.Fatalf("expected fragment inside doc root: %s", s)
	}
}

func TestBlocksToHTMLWithConfig_CustomClass(t *testing.T) {
	pres := render.HTMLPresentationConfig{
		Code: render.HTMLCodePresentation{
			Outer: render.HTMLTagClass{Tag: "pre", Class: "custom-pre"},
			Inner: struct {
				Tag                 string `yaml:"tag"`
				Class               string `yaml:"class"`
				LanguageClassPrefix string `yaml:"language_class_prefix"`
			}{Tag: "code", LanguageClassPrefix: "lang-"},
		},
	}
	blocks := []render.Block{
		{Kind: render.BlockKindCode, Language: "go", Content: "x"},
	}
	out, err := BlocksToHTMLWithConfig(blocks, &pres)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `class="custom-pre"`) || !strings.Contains(out, `class="lang-go"`) {
		t.Fatalf("%s", out)
	}
}
