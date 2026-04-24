package uimodel

import (
	"fmt"
	"strings"

	"github.com/originaleric/digeino/pkg/render"
)

// ParseOutputMode normalizes query/header values（默认 both）。
func ParseOutputMode(s string) OutputMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "both":
		return OutputBoth
	case "blocks":
		return OutputBlocks
	case "ui_model", "uimodel":
		return OutputUIModel
	default:
		return OutputBoth
	}
}

// BuildUIModel maps blocks to cards using Mapping（props：markdown / text / language+code）。
func BuildUIModel(blocks []render.Block, m Mapping) (UIModel, error) {
	m, err := normalizeMapping(m)
	if err != nil {
		return UIModel{}, err
	}
	out := UIModel{SchemaVersion: 1}
	for i, b := range blocks {
		kind := string(b.Kind)
		prof, ok := m.BlockProfiles[kind]
		if !ok {
			return UIModel{}, fmt.Errorf("block_profiles: missing profile for kind %q", kind)
		}
		c := Card{
			Type: prof.Card,
			Meta: &BlockRef{Index: i, Kind: kind},
		}
		switch b.Kind {
		case render.BlockKindMarkdown:
			c.Props = map[string]any{"markdown": b.Content}
		case render.BlockKindThinking:
			c.Props = map[string]any{"text": b.Content}
		case render.BlockKindCode:
			c.Props = map[string]any{
				"language": b.Language,
				"code":     b.Content,
			}
		default:
			return UIModel{}, fmt.Errorf("unsupported block kind %q", b.Kind)
		}
		out.Cards = append(out.Cards, c)
	}
	return out, nil
}

// BuildHybridPayload builds 路线 C 推荐载荷；mode 控制是否包含 blocks / ui_model（SSE done 可整包下发）。
func BuildHybridPayload(mode OutputMode, blocks []render.Block, m Mapping) (*HybridPayload, error) {
	m, err := normalizeMapping(m)
	if err != nil {
		return nil, err
	}
	mode = ParseOutputMode(string(mode))
	h := MappingFingerprint(m)
	p := &HybridPayload{
		MappingVersion:     m.MappingVersion,
		MappingSource:      m.MappingSource,
		MappingHash:        h,
		MappingChangedAt:   strings.TrimSpace(m.MappingChangedAt),
		MappingContentType: "application/json",
	}
	switch mode {
	case OutputBlocks:
		p.Blocks = blocks
	case OutputUIModel:
		ui, err := BuildUIModel(blocks, m)
		if err != nil {
			return nil, err
		}
		p.UIModel = &ui
	case OutputBoth:
		p.Blocks = blocks
		ui, err := BuildUIModel(blocks, m)
		if err != nil {
			return nil, err
		}
		p.UIModel = &ui
	default:
		p.Blocks = blocks
		ui, err := BuildUIModel(blocks, m)
		if err != nil {
			return nil, err
		}
		p.UIModel = &ui
	}
	return p, nil
}
