package ui_ux

import (
	"fmt"
	"strings"
)

// DesignSystem 设计系统结构
type DesignSystem struct {
	ProjectName  string
	Category     string
	Pattern      *Pattern
	Style        *Style
	Colors       *Colors
	Typography   *Typography
	KeyEffects   string
	AntiPatterns string
}

// Pattern 布局模式
type Pattern struct {
	Name          string
	Sections      string
	CTAPlacement  string
	ColorStrategy string
	Conversion    string
}

// Style 样式信息
type Style struct {
	Name         string
	Type         string
	Effects      string
	Keywords     string
	BestFor      string
	Performance  string
	Accessibility string
}

// Colors 配色方案
type Colors struct {
	Primary    string
	Secondary  string
	CTA        string
	Background string
	Text       string
	Notes      string
}

// Typography 字体排版
type Typography struct {
	Heading       string
	Body          string
	Mood          string
	BestFor       string
	GoogleFontsURL string
	CSSImport     string
}

// DesignSystemGenerator 设计系统生成器
type DesignSystemGenerator struct {
	service   *UIUXService
	reasoning *ReasoningEngine
}

// NewDesignSystemGenerator 创建设计系统生成器
func NewDesignSystemGenerator() (*DesignSystemGenerator, error) {
	reasoning, err := NewReasoningEngine()
	if err != nil {
		return nil, err
	}

	return &DesignSystemGenerator{
		service:   NewUIUXService(),
		reasoning: reasoning,
	}, nil
}

// GenerateDesignSystem 生成完整设计系统
func (g *DesignSystemGenerator) GenerateDesignSystem(query string, projectName string) (*DesignSystem, error) {
	// Step 1: 搜索 product 领域获取产品类型
	productResult, err := g.service.Search(query, "product", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to search product: %w", err)
	}

	category := "General"
	if len(productResult.Results) > 0 {
		category = productResult.Results[0]["Product Type"]
	}

	// Step 2: 获取推理规则
	reasoning := g.reasoning.FindRule(category)
	stylePriority := reasoning.StylePriority

	// Step 3: 多领域并行搜索
	styleResult, err := g.service.Search(query, "style", 3)
	if err != nil {
		return nil, fmt.Errorf("failed to search style: %w", err)
	}

	colorResult, err := g.service.Search(query, "color", 2)
	if err != nil {
		return nil, fmt.Errorf("failed to search color: %w", err)
	}

	typographyResult, err := g.service.Search(query, "typography", 2)
	if err != nil {
		return nil, fmt.Errorf("failed to search typography: %w", err)
	}

	landingResult, err := g.service.Search(query, "landing", 2)
	if err != nil {
		return nil, fmt.Errorf("failed to search landing: %w", err)
	}

	// Step 4: 选择最佳匹配
	bestStyle := g.selectBestStyleMatch(styleResult.Results, stylePriority)
	bestColor := g.selectBestMatch(colorResult.Results, nil)
	bestTypography := g.selectBestMatch(typographyResult.Results, nil)
	bestLanding := g.selectBestMatch(landingResult.Results, nil)

	// Step 5: 构建设计系统
	styleEffects := bestStyle["Effects & Animation"]
	if styleEffects == "" {
		styleEffects = reasoning.KeyEffects
	}

	ds := &DesignSystem{
		ProjectName:  projectName,
		Category:     category,
		KeyEffects:    styleEffects,
		AntiPatterns: reasoning.AntiPatterns,
		Pattern: &Pattern{
			Name:          bestLanding["Pattern Name"],
			Sections:      bestLanding["Section Order"],
			CTAPlacement:  bestLanding["Primary CTA Placement"],
			ColorStrategy: bestLanding["Color Strategy"],
			Conversion:    bestLanding["Conversion Optimization"],
		},
		Style: &Style{
			Name:         bestStyle["Style Category"],
			Type:         bestStyle["Type"],
			Effects:      bestStyle["Effects & Animation"],
			Keywords:     bestStyle["Keywords"],
			BestFor:      bestStyle["Best For"],
			Performance:  bestStyle["Performance"],
			Accessibility: bestStyle["Accessibility"],
		},
		Colors: &Colors{
			Primary:    bestColor["Primary (Hex)"],
			Secondary:  bestColor["Secondary (Hex)"],
			CTA:        bestColor["CTA (Hex)"],
			Background: bestColor["Background (Hex)"],
			Text:       bestColor["Text (Hex)"],
			Notes:      bestColor["Notes"],
		},
		Typography: &Typography{
			Heading:       bestTypography["Heading Font"],
			Body:          bestTypography["Body Font"],
			Mood:          bestTypography["Mood/Style Keywords"],
			BestFor:       bestTypography["Best For"],
			GoogleFontsURL: bestTypography["Google Fonts URL"],
			CSSImport:     bestTypography["CSS Import"],
		},
	}

	// 如果 Pattern Name 为空，使用推理规则的推荐模式
	if ds.Pattern.Name == "" {
		ds.Pattern.Name = reasoning.RecommendedPattern
	}

	// 如果 Typography Mood 为空，使用推理规则
	if ds.Typography.Mood == "" {
		ds.Typography.Mood = reasoning.TypographyMood
	}

	return ds, nil
}

// selectBestStyleMatch 选择最佳样式匹配（基于优先级关键词）
func (g *DesignSystemGenerator) selectBestStyleMatch(results []map[string]string, priorityKeywords []string) map[string]string {
	if len(results) == 0 {
		return make(map[string]string)
	}

	if len(priorityKeywords) == 0 {
		return results[0]
	}

	// 1. 精确样式名称匹配
	for _, priority := range priorityKeywords {
		priorityLower := strings.ToLower(strings.TrimSpace(priority))
		for _, result := range results {
			styleName := strings.ToLower(result["Style Category"])
			if strings.Contains(styleName, priorityLower) || strings.Contains(priorityLower, styleName) {
				return result
			}
		}
	}

	// 2. 关键词匹配评分
	type scoredResult struct {
		result map[string]string
		score  int
	}

	scored := make([]scoredResult, 0, len(results))
	for _, result := range results {
		score := 0
		resultStr := strings.ToLower(fmt.Sprintf("%v", result))
		
		for _, kw := range priorityKeywords {
			kwLower := strings.ToLower(strings.TrimSpace(kw))
			
			// 样式名称匹配（高分）
			if strings.Contains(strings.ToLower(result["Style Category"]), kwLower) {
				score += 10
			}
			// 关键词字段匹配（中分）
			if strings.Contains(strings.ToLower(result["Keywords"]), kwLower) {
				score += 3
			}
			// 其他字段匹配（低分）
			if strings.Contains(resultStr, kwLower) {
				score += 1
			}
		}
		
		if score > 0 {
			scored = append(scored, scoredResult{result: result, score: score})
		}
	}

	// 选择最高分
	if len(scored) > 0 {
		best := scored[0]
		for _, s := range scored[1:] {
			if s.score > best.score {
				best = s
			}
		}
		return best.result
	}

	return results[0]
}

// selectBestMatch 选择最佳匹配（通用方法）
func (g *DesignSystemGenerator) selectBestMatch(results []map[string]string, priorityKeywords []string) map[string]string {
	if len(results) == 0 {
		return make(map[string]string)
	}
	return results[0]
}
