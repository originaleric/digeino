package render

// ThinkingTagPair is an opening/closing tag pair that delimits a reasoning section.
type ThinkingTagPair struct {
	Open  string
	Close string
}

// CodeFenceConfig selects how fenced code blocks are recognized (line-open / line-close).
// Zero value means open/close "```" as in CommonMark/GFM.
// If Close is empty after withDefaults, it is set equal to Open.
type CodeFenceConfig struct {
	// Open is the trimmed-line prefix that starts a fence (e.g. "```", "~~~").
	Open string
	// Close is the trimmed-line prefix that ends a fence; same rule as Open (prefix + only whitespace after).
	Close string
}

// Options configures Parse and prefix helpers.
type Options struct {
	// ThinkingTagPairs defines tags that wrap model "thinking" regions.
	// If nil, DefaultThinkingTagPairs is used.
	ThinkingTagPairs []ThinkingTagPair

	// CodeFence configures fenced code block delimiters. Zero value is open/close "```".
	CodeFence CodeFenceConfig
}

// DefaultThinkingTagPairs returns pairs from the embedded config/config.yaml
// (with a small hardcoded fallback if embed or YAML fails).
func DefaultThinkingTagPairs() []ThinkingTagPair {
	return defaultThinkingTagPairsFromEmbed()
}

func (o Options) withDefaults() Options {
	if o.ThinkingTagPairs == nil {
		o.ThinkingTagPairs = DefaultThinkingTagPairs()
	}
	if o.CodeFence.Open == "" {
		o.CodeFence.Open = "```"
	}
	if o.CodeFence.Close == "" {
		o.CodeFence.Close = o.CodeFence.Open
	}
	return o
}

func validateOptions(o Options) error {
	for i, p := range o.ThinkingTagPairs {
		if p.Open == "" || p.Close == "" {
			return &OptionError{Index: i, Msg: "thinking tag pair must have non-empty Open and Close"}
		}
	}
	if o.CodeFence.Open == "" {
		return &OptionError{Msg: "code fence open must be non-empty"}
	}
	if o.CodeFence.Close == "" {
		return &OptionError{Msg: "code fence close must be non-empty"}
	}
	return nil
}

// OptionError describes invalid Options.
type OptionError struct {
	Index int
	Msg   string
}

func (e *OptionError) Error() string {
	return e.Msg
}
