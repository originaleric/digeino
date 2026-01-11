package ui_ux

import (
	"strings"
)

var domainKeywords = map[string][]string{
	"color":      {"color", "palette", "hex", "#", "rgb"},
	"chart":      {"chart", "graph", "visualization", "trend", "bar", "pie", "scatter", "heatmap", "funnel"},
	"landing":    {"landing", "page", "cta", "conversion", "hero", "testimonial", "pricing", "section"},
	"product":    {"saas", "ecommerce", "e-commerce", "fintech", "healthcare", "gaming", "portfolio", "crypto", "dashboard"},
	"prompt":     {"prompt", "css", "implementation", "variable", "checklist", "tailwind"},
	"style":      {"style", "design", "ui", "minimalism", "glassmorphism", "neumorphism", "brutalism", "dark mode", "flat", "aurora"},
	"ux":         {"ux", "usability", "accessibility", "wcag", "touch", "scroll", "animation", "keyboard", "navigation", "mobile"},
	"typography": {"font", "typography", "heading", "serif", "sans"},
}

// DetectDomain auto-detects the most relevant domain from the query.
func DetectDomain(query string) string {
	queryLower := strings.ToLower(query)
	queryPadded := " " + queryLower + " "
	scores := make(map[string]int)

	for domain, keywords := range domainKeywords {
		score := 0
		for _, kw := range keywords {
			if strings.Contains(queryPadded, " "+kw+" ") ||
				strings.Contains(queryPadded, " "+kw+",") ||
				strings.Contains(queryPadded, " "+kw+".") ||
				strings.Contains(queryPadded, " "+kw+"?") ||
				strings.Contains(queryPadded, " "+kw+"!") {
				score++
			}
		}
		scores[domain] = score
	}

	bestDomain := "style"
	maxScore := 0

	for _, domain := range []string{"color", "chart", "landing", "product", "prompt", "style", "ux", "typography"} {
		if scores[domain] > maxScore {
			maxScore = scores[domain]
			bestDomain = domain
		}
	}

	return bestDomain
}
