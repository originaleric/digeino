package ui_ux

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ReasoningRule 推理规则
type ReasoningRule struct {
	UICategory        string
	RecommendedPattern string
	StylePriority     []string
	ColorMood         string
	TypographyMood    string
	KeyEffects        string
	DecisionRules     map[string]string // JSON 条件
	AntiPatterns      string
	Severity          string
}

// ReasoningEngine 推理引擎
type ReasoningEngine struct {
	rules []ReasoningRule
}

// NewReasoningEngine 创建推理引擎并加载规则
func NewReasoningEngine() (*ReasoningEngine, error) {
	engine := &ReasoningEngine{}
	if err := engine.loadRules(); err != nil {
		return nil, fmt.Errorf("failed to load reasoning rules: %w", err)
	}
	return engine, nil
}

// loadRules 从 CSV 文件加载推理规则
func (e *ReasoningEngine) loadRules() error {
	f, err := designData.Open("data/ui-reasoning.csv")
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	headers, err := reader.Read()
	if err != nil {
		return err
	}

	var rules []ReasoningRule
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		ruleMap := make(map[string]string)
		for i, h := range headers {
			if i < len(row) {
				ruleMap[h] = row[i]
			}
		}

		// 解析 Style Priority（用 + 分隔）
		stylePriorityStr := ruleMap["Style_Priority"]
		var stylePriority []string
		if stylePriorityStr != "" {
			parts := strings.Split(stylePriorityStr, "+")
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					stylePriority = append(stylePriority, trimmed)
				}
			}
		}

		// 解析 Decision Rules（JSON）
		decisionRules := make(map[string]string)
		decisionRulesStr := ruleMap["Decision_Rules"]
		if decisionRulesStr != "" {
			if err := json.Unmarshal([]byte(decisionRulesStr), &decisionRules); err != nil {
				// 如果解析失败，忽略错误
				decisionRules = make(map[string]string)
			}
		}

		rule := ReasoningRule{
			UICategory:        ruleMap["UI_Category"],
			RecommendedPattern: ruleMap["Recommended_Pattern"],
			StylePriority:     stylePriority,
			ColorMood:         ruleMap["Color_Mood"],
			TypographyMood:    ruleMap["Typography_Mood"],
			KeyEffects:        ruleMap["Key_Effects"],
			DecisionRules:     decisionRules,
			AntiPatterns:      ruleMap["Anti_Patterns"],
			Severity:          ruleMap["Severity"],
		}

		rules = append(rules, rule)
	}

	e.rules = rules
	return nil
}

// FindRule 根据产品类型查找匹配的推理规则
func (e *ReasoningEngine) FindRule(category string) *ReasoningRule {
	categoryLower := strings.ToLower(category)

	// 1. 精确匹配
	for _, rule := range e.rules {
		if strings.ToLower(rule.UICategory) == categoryLower {
			return &rule
		}
	}

	// 2. 部分匹配（关键词包含）
	for _, rule := range e.rules {
		uiCatLower := strings.ToLower(rule.UICategory)
		if strings.Contains(uiCatLower, categoryLower) || strings.Contains(categoryLower, uiCatLower) {
			return &rule
		}
	}

	// 3. 关键词匹配（分词匹配）
	categoryWords := strings.Fields(categoryLower)
	for _, rule := range e.rules {
		uiCatLower := strings.ToLower(rule.UICategory)
		uiCatWords := strings.Fields(strings.ReplaceAll(strings.ReplaceAll(uiCatLower, "/", " "), "-", " "))
		
		matched := false
		for _, catWord := range categoryWords {
			for _, uiWord := range uiCatWords {
				if catWord == uiWord || strings.Contains(catWord, uiWord) || strings.Contains(uiWord, catWord) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if matched {
			return &rule
		}
	}

	// 4. 返回默认规则
	return &ReasoningRule{
		RecommendedPattern: "Hero + Features + CTA",
		StylePriority:     []string{"Minimalism", "Flat Design"},
		ColorMood:         "Professional",
		TypographyMood:    "Clean",
		KeyEffects:        "Subtle hover transitions",
		AntiPatterns:      "",
		Severity:          "MEDIUM",
	}
}
