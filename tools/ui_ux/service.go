package ui_ux

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
)

//go:embed data/*.csv data/stacks/*.csv
var designData embed.FS

// CSVConfigEntry defines the configuration for each design domain
type CSVConfigEntry struct {
	File       string
	SearchCols []string
	OutputCols []string
}

// CSVConfigs maps domain keys to their respective CSV configurations
var CSVConfigs = map[string]CSVConfigEntry{
	"style": {
		File:       "data/styles.csv",
		SearchCols: []string{"Style Category", "Keywords", "Best For", "Type"},
		OutputCols: []string{"Style Category", "Type", "Keywords", "Primary Colors", "Effects & Animation", "Best For", "Performance", "Accessibility", "Framework Compatibility", "Complexity"},
	},
	"prompt": {
		File:       "data/prompts.csv",
		SearchCols: []string{"Style Category", "AI Prompt Keywords (Copy-Paste Ready)", "CSS/Technical Keywords"},
		OutputCols: []string{"Style Category", "AI Prompt Keywords (Copy-Paste Ready)", "CSS/Technical Keywords", "Implementation Checklist"},
	},
	"color": {
		File:       "data/colors.csv",
		SearchCols: []string{"Product Type", "Keywords", "Notes"},
		OutputCols: []string{"Product Type", "Keywords", "Primary (Hex)", "Secondary (Hex)", "CTA (Hex)", "Background (Hex)", "Text (Hex)", "Border (Hex)", "Notes"},
	},
	"chart": {
		File:       "data/charts.csv",
		SearchCols: []string{"Data Type", "Keywords", "Best Chart Type", "Accessibility Notes"},
		OutputCols: []string{"Data Type", "Keywords", "Best Chart Type", "Secondary Options", "Color Guidance", "Accessibility Notes", "Library Recommendation", "Interactive Level"},
	},
	"landing": {
		File:       "data/landing.csv",
		SearchCols: []string{"Pattern Name", "Keywords", "Conversion Optimization", "Section Order"},
		OutputCols: []string{"Pattern Name", "Keywords", "Section Order", "Primary CTA Placement", "Color Strategy", "Conversion Optimization"},
	},
	"product": {
		File:       "data/products.csv",
		SearchCols: []string{"Product Type", "Keywords", "Primary Style Recommendation", "Key Considerations"},
		OutputCols: []string{"Product Type", "Keywords", "Primary Style Recommendation", "Secondary Styles", "Landing Page Pattern", "Dashboard Style (if applicable)", "Color Palette Focus"},
	},
	"ux": {
		File:       "data/ux-guidelines.csv",
		SearchCols: []string{"Category", "Issue", "Description", "Platform"},
		OutputCols: []string{"Category", "Issue", "Platform", "Description", "Do", "Don't", "Code Example Good", "Code Example Bad", "Severity"},
	},
	"typography": {
		File:       "data/typography.csv",
		SearchCols: []string{"Font Pairing Name", "Category", "Mood/Style Keywords", "Best For", "Heading Font", "Body Font"},
		OutputCols: []string{"Font Pairing Name", "Category", "Heading Font", "Body Font", "Mood/Style Keywords", "Best For", "Google Fonts URL", "CSS Import", "Tailwind Config", "Notes"},
	},
}

// StackConfigs maps technology stacks to their respective CSV files
var StackConfigs = map[string]string{
	"html-tailwind": "data/stacks/html-tailwind.csv",
	"react":         "data/stacks/react.csv",
	"nextjs":        "data/stacks/nextjs.csv",
	"vue":           "data/stacks/vue.csv",
	"nuxtjs":        "data/stacks/nuxtjs.csv",
	"nuxt-ui":       "data/stacks/nuxt-ui.csv",
	"svelte":        "data/stacks/svelte.csv",
	"swiftui":       "data/stacks/swiftui.csv",
	"react-native":  "data/stacks/react-native.csv",
	"flutter":       "data/stacks/flutter.csv",
}

// StackCols defines search and output columns for stack-specific data
var StackCols = struct {
	SearchCols []string
	OutputCols []string
}{
	SearchCols: []string{"Category", "Guideline", "Description", "Do", "Don't"},
	OutputCols: []string{"Category", "Guideline", "Description", "Do", "Don't", "Code Good", "Code Bad", "Severity", "Docs URL"},
}

// UIUXService manages design intelligence data and search
type UIUXService struct{}

// NewUIUXService creates a new UIUXService using embedded data
func NewUIUXService() *UIUXService {
	return &UIUXService{}
}

func (s *UIUXService) loadCSV(filename string) ([]map[string]string, error) {
	f, err := designData.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	var results []map[string]string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		item := make(map[string]string)
		for i, h := range headers {
			if i < len(row) {
				item[h] = row[i]
			}
		}
		results = append(results, item)
	}
	return results, nil
}

// SearchResult represents the output of a design intelligence search
type SearchResult struct {
	Domain  string              `json:"domain"`
	Query   string              `json:"query"`
	File    string              `json:"file"`
	Count   int                 `json:"count"`
	Results []map[string]string `json:"results"`
}

// Search performs a BM25 search across design domains
func (s *UIUXService) Search(query string, domain string, maxResults int) (*SearchResult, error) {
	if domain == "" {
		domain = DetectDomain(query)
	}

	config, ok := CSVConfigs[domain]
	if !ok {
		config = CSVConfigs["style"]
		domain = "style"
	}

	data, err := s.loadCSV(config.File)
	if err != nil {
		return nil, fmt.Errorf("failed to load csv %s: %w", config.File, err)
	}

	documents := make([]string, len(data))
	for i, row := range data {
		var builder strings.Builder
		for _, col := range config.SearchCols {
			builder.WriteString(row[col])
			builder.WriteString(" ")
		}
		documents[i] = builder.String()
	}

	bm25 := NewBM25(1.5, 0.75)
	bm25.Fit(documents)
	ranked := bm25.Score(query)

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	var finalResults []map[string]string
	for i := 0; i < len(ranked) && len(finalResults) < maxResults; i++ {
		if ranked[i].Score > 0 {
			origRow := data[ranked[i].Index]
			filteredRow := make(map[string]string)
			for _, col := range config.OutputCols {
				if val, ok := origRow[col]; ok {
					filteredRow[col] = val
				}
			}
			finalResults = append(finalResults, filteredRow)
		}
	}

	return &SearchResult{
		Domain:  domain,
		Query:   query,
		File:    config.File,
		Count:   len(finalResults),
		Results: finalResults,
	}, nil
}

// SearchStack performs a BM25 search for technology-specific design guidelines
func (s *UIUXService) SearchStack(query string, stack string, maxResults int) (*SearchResult, error) {
	filename, ok := StackConfigs[stack]
	if !ok {
		return nil, fmt.Errorf("unknown stack: %s", stack)
	}

	data, err := s.loadCSV(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load stack csv %s: %w", filename, err)
	}

	documents := make([]string, len(data))
	for i, row := range data {
		var builder strings.Builder
		for _, col := range StackCols.SearchCols {
			builder.WriteString(row[col])
			builder.WriteString(" ")
		}
		documents[i] = builder.String()
	}

	bm25 := NewBM25(1.5, 0.75)
	bm25.Fit(documents)
	ranked := bm25.Score(query)

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	var finalResults []map[string]string
	for i := 0; i < len(ranked) && len(finalResults) < maxResults; i++ {
		if ranked[i].Score > 0 {
			origRow := data[ranked[i].Index]
			filteredRow := make(map[string]string)
			for _, col := range StackCols.OutputCols {
				if val, ok := origRow[col]; ok {
					filteredRow[col] = val
				}
			}
			finalResults = append(finalResults, filteredRow)
		}
	}

	return &SearchResult{
		Domain:  "stack",
		Query:   query,
		File:    filename,
		Count:   len(finalResults),
		Results: finalResults,
	}, nil
}

// GenerateDesignSystem 生成完整设计系统
func (s *UIUXService) GenerateDesignSystem(query string, projectName string) (*DesignSystem, error) {
	generator, err := NewDesignSystemGenerator()
	if err != nil {
		return nil, err
	}
	return generator.GenerateDesignSystem(query, projectName)
}
