package render

// BlockKind identifies how a slice of assistant text should be presented.
type BlockKind string

const (
	BlockKindMarkdown BlockKind = "markdown"
	BlockKindThinking BlockKind = "thinking"
	BlockKindCode     BlockKind = "code"
)

// Block is one renderable segment of assistant output.
type Block struct {
	Kind     BlockKind `json:"kind"`
	Language string    `json:"language,omitempty"` // set when Kind == BlockKindCode
	Content  string    `json:"content"`
}
