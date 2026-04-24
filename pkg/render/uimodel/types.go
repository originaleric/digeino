// Package uimodel maps render.Block slices to a small, JSON-friendly card model (路线 C：与 blocks 并存，供前端优先消费并可回退)。
package uimodel

import "github.com/originaleric/digeino/pkg/render"

// OutputMode selects which fields appear on HybridPayload (API / SSE done 事件可复用).
type OutputMode string

const (
	OutputBlocks  OutputMode = "blocks"
	OutputUIModel OutputMode = "ui_model"
	OutputBoth    OutputMode = "both"
)

// UIModel is a versioned list of cards for React (or other) hosts.
type UIModel struct {
	SchemaVersion int    `json:"schema_version"`
	Cards         []Card `json:"cards"`
}

// Card is one UI card; Type names are host-defined conventions (首版默认 MarkdownCard / ReasoningCard / CodeCard)。
type Card struct {
	Type  string         `json:"type"`
	Props map[string]any `json:"props"`
	Meta  *BlockRef      `json:"meta,omitempty"`
}

// BlockRef ties a card back to the source block index (便于前端与 blocks 对齐、调试)。
type BlockRef struct {
	Index int    `json:"index"`
	Kind  string `json:"kind"`
}

// HybridPayload is the recommended API shape for 路线 C：blocks 为真源，ui_model 为可选派生。
type HybridPayload struct {
	Blocks             []render.Block `json:"blocks,omitempty"`
	UIModel            *UIModel       `json:"ui_model,omitempty"`
	MappingVersion     string         `json:"mapping_version,omitempty"`
	MappingSource      string         `json:"mapping_source,omitempty"`
	MappingHash        string         `json:"mapping_hash,omitempty"`
	MappingChangedAt   string         `json:"mapping_changed_at,omitempty"`   // RFC3339，可选
	MappingContentType string         `json:"mapping_content_type,omitempty"` // 例如 application/json
}
