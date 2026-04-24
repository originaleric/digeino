package render

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// parseFileYAML 除 thinking_tag_pairs、code_fence 外，任意键若值为 {open, close} 则自动视为思考区标签对（与 thinking_tag_pairs 合并）。
type parseFileYAML struct {
	ThinkingTagPairs *[]thinkingYAML
	CodeFence        *codeFenceYAML
	ParseExtraPairs  []thinkingYAML `yaml:"-"`
	ParseExtraKeys   []string       `yaml:"-"`
}

func (p *parseFileYAML) UnmarshalYAML(n *yaml.Node) error {
	*p = parseFileYAML{}
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("parse: must be a mapping")
	}
	for i := 0; i < len(n.Content); i += 2 {
		key := n.Content[i].Value
		valNode := n.Content[i+1]
		switch key {
		case "thinking_tag_pairs":
			var list []thinkingYAML
			if err := valNode.Decode(&list); err != nil {
				return fmt.Errorf("parse.thinking_tag_pairs: %w", err)
			}
			p.ThinkingTagPairs = &list
		case "code_fence":
			var cf codeFenceYAML
			if err := valNode.Decode(&cf); err != nil {
				return fmt.Errorf("parse.code_fence: %w", err)
			}
			p.CodeFence = &cf
		case "custom_tag_pairs":
			return fmt.Errorf("parse.custom_tag_pairs is not supported; use parse.<your_key>: {open, close} instead")
		default:
			var pair thinkingYAML
			if err := valNode.Decode(&pair); err != nil {
				return fmt.Errorf("parse.%s: %w", key, err)
			}
			o, c := strings.TrimSpace(pair.Open), strings.TrimSpace(pair.Close)
			if o == "" || c == "" {
				return fmt.Errorf("parse.%s: need non-empty open and close (or use a known key: thinking_tag_pairs, code_fence)", key)
			}
			p.ParseExtraPairs = append(p.ParseExtraPairs, thinkingYAML{Open: o, Close: c})
			p.ParseExtraKeys = append(p.ParseExtraKeys, key)
		}
	}
	return nil
}

func validateParseRenderExtras(parse *parseFileYAML, links ParseRenderLinks) error {
	if links.Extra == nil || len(links.Extra) == 0 {
		return nil
	}
	if parse == nil {
		return fmt.Errorf("parse_render: extra keys require a parse: section with matching open/close blocks")
	}
	for _, k := range parse.ParseExtraKeys {
		prof, ok := links.Extra[k]
		if !ok || prof == "" {
			continue
		}
		if prof != "thinking" {
			return fmt.Errorf("parse_render.%s: open/close blocks only support profile %q, got %q", k, "thinking", prof)
		}
	}
	for k := range links.Extra {
		found := false
		for _, pk := range parse.ParseExtraKeys {
			if pk == k {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("parse_render.%s: no parse.%s with open/close", k, k)
		}
	}
	return nil
}
