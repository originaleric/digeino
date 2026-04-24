package render

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseRenderLinks maps parse-time areas to keys under the render: section.
// Defaults: thinking_tag_pairs→thinking, code_fence→code, markdown→markdown.
// Extra keys (e.g. my_tag: thinking) must match a parse.<same_key> open/close block.
type ParseRenderLinks struct {
	ThinkingTagPairs string `yaml:"thinking_tag_pairs"`
	CodeFence        string `yaml:"code_fence"`
	Markdown         string `yaml:"markdown"`
	Extra            map[string]string `yaml:"-"`
}

func (p *ParseRenderLinks) UnmarshalYAML(n *yaml.Node) error {
	*p = ParseRenderLinks{}
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("parse_render: must be a mapping")
	}
	extras := make(map[string]string)
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i].Value
		vNode := n.Content[i+1]
		var v string
		if err := vNode.Decode(&v); err != nil {
			return fmt.Errorf("parse_render.%s: %w", k, err)
		}
		switch k {
		case "thinking_tag_pairs":
			p.ThinkingTagPairs = v
		case "code_fence":
			p.CodeFence = v
		case "markdown":
			p.Markdown = v
		default:
			extras[k] = v
		}
	}
	if len(extras) > 0 {
		p.Extra = extras
	}
	return nil
}

func defaultParseRenderLinks() ParseRenderLinks {
	return ParseRenderLinks{
		ThinkingTagPairs: "thinking",
		CodeFence:        "code",
		Markdown:         "markdown",
	}
}

func normalizeParseRenderLinks(p *ParseRenderLinks) (ParseRenderLinks, error) {
	d := defaultParseRenderLinks()
	out := d
	if p != nil {
		out = *p
	}
	if out.Extra == nil {
		out.Extra = map[string]string{}
	} else {
		cp := make(map[string]string, len(out.Extra))
		for k, v := range out.Extra {
			cp[k] = v
		}
		out.Extra = cp
	}
	if out.ThinkingTagPairs == "" {
		out.ThinkingTagPairs = d.ThinkingTagPairs
	}
	if out.CodeFence == "" {
		out.CodeFence = d.CodeFence
	}
	if out.Markdown == "" {
		out.Markdown = d.Markdown
	}
	if out.ThinkingTagPairs == out.CodeFence || out.ThinkingTagPairs == out.Markdown || out.CodeFence == out.Markdown {
		return ParseRenderLinks{}, fmt.Errorf(
			"parse_render: render profile keys must be unique (thinking_tag_pairs→%q, code_fence→%q, markdown→%q)",
			out.ThinkingTagPairs, out.CodeFence, out.Markdown,
		)
	}
	if out.ThinkingTagPairs == "document" || out.CodeFence == "document" || out.Markdown == "document" {
		return ParseRenderLinks{}, fmt.Errorf("parse_render: profile key %q is reserved for render.document", "document")
	}
	for k, v := range out.Extra {
		vv := strings.ToLower(strings.TrimSpace(v))
		if vv != "thinking" && vv != "code" && vv != "markdown" {
			return ParseRenderLinks{}, fmt.Errorf("parse_render.%s: profile must be thinking|code|markdown, got %q", k, v)
		}
		out.Extra[k] = vv
	}
	return out, nil
}

// capturedYAMLNode holds a raw mapping node for render: (dynamic profile keys).
type capturedYAMLNode struct {
	Node *yaml.Node
}

func (c *capturedYAMLNode) UnmarshalYAML(n *yaml.Node) error {
	c.Node = n
	return nil
}

// presentationFromRenderYAML decodes render: using parse_render profile names.
// The key "document" is always the document shell (not remapped).
func presentationFromRenderYAML(n *yaml.Node, links ParseRenderLinks) (HTMLPresentationConfig, error) {
	if n == nil {
		return HTMLPresentationConfig{}, nil
	}
	if n.Kind != yaml.MappingNode {
		return HTMLPresentationConfig{}, fmt.Errorf("render: must be a mapping")
	}
	nodes := make(map[string]*yaml.Node)
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i].Value
		nodes[k] = n.Content[i+1]
	}
	var raw HTMLPresentationConfigRaw
	if node, ok := nodes[links.ThinkingTagPairs]; ok {
		var th thinkingRawYAML
		if err := node.Decode(&th); err != nil {
			return HTMLPresentationConfig{}, fmt.Errorf("render.%s: %w", links.ThinkingTagPairs, err)
		}
		raw.Thinking = th
	}
	if node, ok := nodes[links.CodeFence]; ok {
		var co codeRawYAML
		if err := node.Decode(&co); err != nil {
			return HTMLPresentationConfig{}, fmt.Errorf("render.%s: %w", links.CodeFence, err)
		}
		raw.Code = co
	}
	if node, ok := nodes[links.Markdown]; ok {
		var md HTMLMarkdownPresentation
		if err := node.Decode(&md); err != nil {
			return HTMLPresentationConfig{}, fmt.Errorf("render.%s: %w", links.Markdown, err)
		}
		raw.Markdown = md
	}
	if node, ok := nodes["document"]; ok {
		var doc HTMLDocumentPresentation
		if err := node.Decode(&doc); err != nil {
			return HTMLPresentationConfig{}, fmt.Errorf("render.document: %w", err)
		}
		raw.Document = doc
	}
	return MigrateHTMLPresentationRaw(raw), nil
}
