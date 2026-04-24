// Package html converts render.Block slices to sanitized HTML fragments or full documents.
// It does not import cloudwego/eino.
package html

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/originaleric/digeino/pkg/render"
	"github.com/yuin/goldmark"
)

var gm = goldmark.New()

// BlocksToHTML renders blocks using HTMLPresentation from embedded config (see config.yaml render:).
// Equivalent to BlocksToHTMLWithConfig(blocks, nil).
func BlocksToHTML(blocks []render.Block) (string, error) {
	return BlocksToHTMLWithConfig(blocks, nil)
}

// BlocksToHTMLWithConfig renders blocks using pres; nil loads DefaultHTMLPresentation().
func BlocksToHTMLWithConfig(blocks []render.Block, pres *render.HTMLPresentationConfig) (string, error) {
	var c render.HTMLPresentationConfig
	if pres != nil {
		c = render.NormalizeHTMLPresentation(*pres)
	} else {
		c = render.DefaultHTMLPresentation()
	}
	sanitize := markdownSanitizer(c.Markdown.Sanitize)

	var buf strings.Builder
	for _, b := range blocks {
		switch b.Kind {
		case render.BlockKindMarkdown:
			var out bytes.Buffer
			if err := gm.Convert([]byte(b.Content), &out); err != nil {
				return "", err
			}
			buf.WriteString(sanitize(out.String()))
		case render.BlockKindThinking:
			fmt.Fprintf(&buf, "<%s", c.Thinking.Outer.Tag)
			if c.Thinking.Outer.Class != "" {
				fmt.Fprintf(&buf, ` class="%s"`, html.EscapeString(c.Thinking.Outer.Class))
			}
			buf.WriteByte('>')
			fmt.Fprintf(&buf, "<%s", c.Thinking.Inner.Tag)
			if c.Thinking.Inner.Class != "" {
				fmt.Fprintf(&buf, ` class="%s"`, html.EscapeString(c.Thinking.Inner.Class))
			}
			buf.WriteString(">")
			buf.WriteString(html.EscapeString(b.Content))
			fmt.Fprintf(&buf, "</%s></%s>", c.Thinking.Inner.Tag, c.Thinking.Outer.Tag)
		case render.BlockKindCode:
			fmt.Fprintf(&buf, "<%s", c.Code.Outer.Tag)
			if c.Code.Outer.Class != "" {
				fmt.Fprintf(&buf, ` class="%s"`, html.EscapeString(c.Code.Outer.Class))
			}
			buf.WriteByte('>')
			fmt.Fprintf(&buf, "<%s", c.Code.Inner.Tag)
			codeClass := codeInnerClassAttr(c.Code.Inner.Class, c.Code.Inner.LanguageClassPrefix, b.Language)
			if codeClass != "" {
				fmt.Fprintf(&buf, ` class="%s"`, codeClass)
			}
			buf.WriteString(">")
			buf.WriteString(html.EscapeString(b.Content))
			fmt.Fprintf(&buf, "</%s></%s>", c.Code.Inner.Tag, c.Code.Outer.Tag)
		default:
			buf.WriteString(html.EscapeString(b.Content))
		}
	}
	return buf.String(), nil
}

func codeInnerClassAttr(staticClass, langPrefix, lang string) string {
	var parts []string
	if strings.TrimSpace(staticClass) != "" {
		parts = append(parts, html.EscapeString(strings.TrimSpace(staticClass)))
	}
	if strings.TrimSpace(lang) != "" {
		parts = append(parts, html.EscapeString(langPrefix+lang))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func markdownSanitizer(policyName string) func(string) string {
	switch strings.ToLower(strings.TrimSpace(policyName)) {
	case "strict":
		p := bluemonday.StrictPolicy()
		return p.Sanitize
	case "noop", "none", "raw":
		return func(s string) string { return s }
	default:
		p := bluemonday.UGCPolicy()
		return p.Sanitize
	}
}

// WrapDocument wraps a fragment in a minimal HTML5 shell using document settings from DefaultHTMLPresentation().
func WrapDocument(title, bodyHTML string) string {
	return WrapDocumentWithConfig(title, bodyHTML, nil)
}

// WrapDocumentWithConfig uses pres.Document for lang and inline CSS; nil uses DefaultHTMLPresentation().
func WrapDocumentWithConfig(title, bodyHTML string, pres *render.HTMLPresentationConfig) string {
	var doc render.HTMLDocumentPresentation
	if pres != nil {
		p := render.NormalizeHTMLPresentation(*pres)
		doc = p.Document
	} else {
		doc = render.DefaultHTMLPresentation().Document
	}
	t := html.EscapeString(title)
	lang := html.EscapeString(doc.Lang)
	css := doc.InlineCSS
	// 片段本身无 body；整页导出时用包裹层，避免 inline_css 依赖 body 选择器（与 BlocksToHTML 输出语义一致）。
	const docRootClass = "llm-render-doc"
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>
%s
</style>
</head>
<body>
<div class="%s">
%s
</div>
</body>
</html>`, lang, t, css, docRootClass, bodyHTML)
}
