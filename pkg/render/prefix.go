package render

import "strings"

// PrefixResult is the outcome of ParseStablePrefix for streaming UIs.
type PrefixResult struct {
	Blocks    []Block `json:"blocks"`
	Remainder string  `json:"remainder"`
}

// ParseStablePrefix parses the longest prefix of input that is structurally complete:
// all fenced code blocks are closed, and no thinking tag pair is left unclosed at the end.
// The remainder should be shown as raw / pending text in Cursor-style streaming.
func ParseStablePrefix(input string, opts Options) (PrefixResult, error) {
	opts = opts.withDefaults()
	if err := validateOptions(opts); err != nil {
		return PrefixResult{}, err
	}
	input = normalizeNL(input)
	cut := stableCutIndex(input, opts.ThinkingTagPairs, opts.CodeFence.Open, opts.CodeFence.Close)
	if cut <= 0 {
		return PrefixResult{Remainder: input}, nil
	}
	blocks, err := Parse(input[:cut], opts)
	if err != nil {
		return PrefixResult{}, err
	}
	return PrefixResult{Blocks: blocks, Remainder: input[cut:]}, nil
}

func stableCutIndex(s string, pairs []ThinkingTagPair, fenceOpen, fenceClose string) int {
	if fenceOpen == "" {
		fenceOpen = "```"
	}
	if fenceClose == "" {
		fenceClose = fenceOpen
	}
	lines := strings.Split(s, "\n")
	fenceLine := fenceStableLine(lines, fenceOpen, fenceClose)
	fenceCut := len(s)
	if fenceLine < len(lines) {
		fenceCut = lineByteOffset(lines, fenceLine)
	}
	thinkCut := thinkingStableCut(s, pairs)
	if fenceCut < thinkCut {
		return fenceCut
	}
	return thinkCut
}

// fenceStableLine returns the line index of an opening fence whose fence is not closed by EOF, or len(lines) if none.
func fenceStableLine(lines []string, fenceOpen, fenceClose string) int {
	inFence := false
	openLine := len(lines)
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if !inFence {
			if strings.HasPrefix(trim, fenceOpen) {
				inFence = true
				openLine = i
			}
		} else {
			if isFenceClose(trim, fenceClose) {
				inFence = false
				openLine = len(lines)
			}
		}
	}
	if inFence {
		return openLine
	}
	return len(lines)
}

func lineByteOffset(lines []string, lineIdx int) int {
	off := 0
	for i := 0; i < lineIdx && i < len(lines); i++ {
		off += len(lines[i]) + 1
	}
	return off
}

// thinkingStableCut returns the byte index of the earliest thinking Open that has no matching Close in the suffix, or len(s).
func thinkingStableCut(s string, pairs []ThinkingTagPair) int {
	best := len(s)
	for _, p := range pairs {
		pos := 0
		for {
			i := strings.Index(s[pos:], p.Open)
			if i < 0 {
				break
			}
			abs := pos + i
			if strings.Index(s[abs+len(p.Open):], p.Close) < 0 {
				if abs < best {
					best = abs
				}
			}
			pos = abs + 1
		}
	}
	return best
}
