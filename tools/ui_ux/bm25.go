package ui_ux

import (
	"math"
	"regexp"
	"strings"
)

var (
	rePunct = regexp.MustCompile(`[^\w\s]`)
)

// BM25 implements the BM25 ranking algorithm.
type BM25 struct {
	k1         float64
	b          float64
	corpus     [][]string
	docLengths []int
	avgdl      float64
	idf        map[string]float64
	docFreqs   map[string]int
	N          int
}

// NewBM25 creates a new BM25 instance.
func NewBM25(k1, b float64) *BM25 {
	return &BM25{
		k1:       k1,
		b:        b,
		idf:      make(map[string]float64),
		docFreqs: make(map[string]int),
	}
}

// Tokenize converts text into a slice of tokens.
func (b *BM25) Tokenize(text string) []string {
	// Lowercase and remove punctuation
	text = strings.ToLower(text)
	text = rePunct.ReplaceAllString(text, " ")

	words := strings.Fields(text)
	var tokens []string
	for _, w := range words {
		if len(w) > 2 {
			tokens = append(tokens, w)
		}
	}
	return tokens
}

// Fit builds the BM25 index from documents.
func (b *BM25) Fit(documents []string) {
	b.corpus = make([][]string, len(documents))
	b.docLengths = make([]int, len(documents))
	b.N = len(documents)

	if b.N == 0 {
		return
	}

	totalLen := 0
	for i, doc := range documents {
		tokens := b.Tokenize(doc)
		b.corpus[i] = tokens
		b.docLengths[i] = len(tokens)
		totalLen += len(tokens)

		seen := make(map[string]bool)
		for _, token := range tokens {
			if !seen[token] {
				b.docFreqs[token]++
				seen[token] = true
			}
		}
	}

	b.avgdl = float64(totalLen) / float64(b.N)

	for word, freq := range b.docFreqs {
		b.idf[word] = math.Log((float64(b.N)-float64(freq)+0.5)/(float64(freq)+0.5) + 1.0)
	}
}

// ScoreRank represents a document index and its score.
type ScoreRank struct {
	Index int
	Score float64
}

// Score all documents against the query.
func (b *BM25) Score(query string) []ScoreRank {
	queryTokens := b.Tokenize(query)
	scores := make([]ScoreRank, b.N)

	for i, doc := range b.corpus {
		score := 0.0
		docLen := b.docLengths[i]

		termFreqs := make(map[string]int)
		for _, token := range doc {
			termFreqs[token]++
		}

		for _, token := range queryTokens {
			if idfVal, ok := b.idf[token]; ok {
				tf := float64(termFreqs[token])
				numerator := tf * (b.k1 + 1.0)
				denominator := tf + b.k1*(1.0-b.b+b.b*float64(docLen)/b.avgdl)
				score += idfVal * numerator / denominator
			}
		}
		scores[i] = ScoreRank{Index: i, Score: score}
	}

	return scores
}
