package render

import "strings"

// Parse splits assistant output into blocks (markdown, thinking, code).
// It does not depend on any agent framework; input must be complete text (not a stream chunk).
func Parse(input string, opts Options) ([]Block, error) {
	opts = opts.withDefaults()
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	input = normalizeNL(input)
	segs := splitByCodeFences(input, opts.CodeFence.Open, opts.CodeFence.Close)
	var all []Block
	for _, seg := range segs {
		if seg.isCode {
			all = append(all, Block{
				Kind:     BlockKindCode,
				Language: seg.lang,
				Content:  seg.content,
			})
			continue
		}
		if strings.TrimSpace(seg.content) == "" {
			continue
		}
		chunks := parseThinkingSegment(seg.content, opts.ThinkingTagPairs)
		chunks = mergeAdjacentMarkdown(chunks)
		all = append(all, chunks...)
	}
	all = mergeAdjacentMarkdown(all)
	all = trimEmptyBlocks(all)
	return all, nil
}

func normalizeNL(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

type rawSegment struct {
	isCode  bool
	lang    string
	content string
}

func isFenceClose(trimmed, fenceClosePrefix string) bool {
	if !strings.HasPrefix(trimmed, fenceClosePrefix) {
		return false
	}
	rest := strings.TrimSpace(trimmed[len(fenceClosePrefix):])
	return rest == ""
}

func splitByCodeFences(input, fenceOpen, fenceClose string) []rawSegment {
	if input == "" {
		return nil
	}
	if fenceOpen == "" {
		fenceOpen = "```"
	}
	if fenceClose == "" {
		fenceClose = fenceOpen
	}
	lines := strings.Split(input, "\n")
	var out []rawSegment
	i := 0
	for i < len(lines) {
		trim := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trim, fenceOpen) {
			lang := strings.TrimSpace(trim[len(fenceOpen):])
			i++
			var body []string
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				if isFenceClose(t, fenceClose) {
					break
				}
				body = append(body, lines[i])
				i++
			}
			out = append(out, rawSegment{
				isCode:  true,
				lang:    lang,
				content: strings.Join(body, "\n"),
			})
			if i < len(lines) && isFenceClose(strings.TrimSpace(lines[i]), fenceClose) {
				i++
			}
			continue
		}
		var md []string
		for i < len(lines) {
			t := strings.TrimSpace(lines[i])
			if strings.HasPrefix(t, fenceOpen) {
				break
			}
			md = append(md, lines[i])
			i++
		}
		out = append(out, rawSegment{isCode: false, content: strings.Join(md, "\n")})
	}
	return out
}

func parseThinkingSegment(s string, pairs []ThinkingTagPair) []Block {
	earliest := -1
	var chosen ThinkingTagPair
	for _, p := range pairs {
		idx := strings.Index(s, p.Open)
		if idx >= 0 && (earliest < 0 || idx < earliest) {
			earliest = idx
			chosen = p
		}
	}
	if earliest < 0 {
		if s == "" {
			return nil
		}
		return []Block{{Kind: BlockKindMarkdown, Content: s}}
	}
	var blocks []Block
	if earliest > 0 {
		blocks = append(blocks, Block{Kind: BlockKindMarkdown, Content: s[:earliest]})
	}
	rest := s[earliest+len(chosen.Open):]
	closeIdx := strings.Index(rest, chosen.Close)
	if closeIdx < 0 {
		blocks = append(blocks, Block{Kind: BlockKindThinking, Content: rest})
		return blocks
	}
	think := rest[:closeIdx]
	blocks = append(blocks, Block{Kind: BlockKindThinking, Content: think})
	after := rest[closeIdx+len(chosen.Close):]
	blocks = append(blocks, parseThinkingSegment(after, pairs)...)
	return blocks
}

func mergeAdjacentMarkdown(blocks []Block) []Block {
	if len(blocks) == 0 {
		return blocks
	}
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if b.Kind == BlockKindMarkdown && len(out) > 0 && out[len(out)-1].Kind == BlockKindMarkdown {
			prev := &out[len(out)-1]
			if prev.Content != "" && b.Content != "" && !strings.HasSuffix(prev.Content, "\n") {
				prev.Content += "\n"
			}
			prev.Content += b.Content
			continue
		}
		out = append(out, b)
	}
	return out
}

func trimEmptyBlocks(blocks []Block) []Block {
	var out []Block
	for _, b := range blocks {
		if b.Kind == BlockKindMarkdown && strings.TrimSpace(b.Content) == "" {
			continue
		}
		out = append(out, b)
	}
	return out
}
