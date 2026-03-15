package ui_ux

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
)

//go:embed data/anti_patterns.csv
var antiPatternsData embed.FS

// AntiPattern 反模式定义
type AntiPattern struct {
	Name          string
	Description   string
	Severity      string // Critical/High/Medium/Low
	Detection     string // 正则表达式或关键词
	Recommendation string
	Command       string // audit/critique/normalize
	Category      string // Typography/Color/Layout/Motion/Interaction
	Examples      string
}

// AntiPatternDetector 反模式检测引擎
type AntiPatternDetector struct {
	patterns []AntiPattern
}

// NewAntiPatternDetector 创建反模式检测引擎
func NewAntiPatternDetector() (*AntiPatternDetector, error) {
	detector := &AntiPatternDetector{}
	if err := detector.loadPatterns(); err != nil {
		return nil, fmt.Errorf("failed to load anti-patterns: %w", err)
	}
	return detector, nil
}

// loadPatterns 从 CSV 文件加载反模式
func (d *AntiPatternDetector) loadPatterns() error {
	f, err := antiPatternsData.Open("data/anti_patterns.csv")
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	headers, err := reader.Read()
	if err != nil {
		return err
	}

	var patterns []AntiPattern
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 构建字段映射
		fieldMap := make(map[string]string)
		for i, h := range headers {
			if i < len(row) {
				fieldMap[h] = row[i]
			}
		}

		pattern := AntiPattern{
			Name:          fieldMap["Name"],
			Description:   fieldMap["Description"],
			Severity:      fieldMap["Severity"],
			Detection:     fieldMap["Detection"],
			Recommendation: fieldMap["Recommendation"],
			Command:       fieldMap["Command"],
			Category:      fieldMap["Category"],
			Examples:      fieldMap["Examples"],
		}

		patterns = append(patterns, pattern)
	}

	d.patterns = patterns
	return nil
}

// DetectFinding 检测结果
type DetectFinding struct {
	Pattern       AntiPattern
	Location      string // 文件路径和行号
	MatchText     string // 匹配的文本
	Severity      string
	Category      string
	Description   string
	Recommendation string
	SuggestedCommand string
}

// Detect 检测代码中的反模式
func (d *AntiPatternDetector) Detect(content string, filePath string) []DetectFinding {
	var findings []DetectFinding

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineNum := i + 1
		for _, pattern := range d.patterns {
			// 尝试正则表达式匹配
			matched, matchText := d.matchPattern(pattern.Detection, line)
			if matched {
				finding := DetectFinding{
					Pattern:         pattern,
					Location:        fmt.Sprintf("%s:%d", filePath, lineNum),
					MatchText:       matchText,
					Severity:        pattern.Severity,
					Category:        pattern.Category,
					Description:     pattern.Description,
					Recommendation:  pattern.Recommendation,
					SuggestedCommand: pattern.Command,
				}
				findings = append(findings, finding)
			}
		}
	}

	return findings
}

// matchPattern 匹配模式（支持正则表达式和关键词）
func (d *AntiPatternDetector) matchPattern(pattern, text string) (bool, string) {
	// 尝试正则表达式匹配
	re, err := regexp.Compile("(?i)" + pattern) // 不区分大小写
	if err == nil {
		matches := re.FindStringSubmatch(text)
		if len(matches) > 0 {
			return true, matches[0]
		}
	}

	// 如果正则表达式失败，尝试关键词匹配
	keywords := strings.Split(pattern, "|")
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if strings.Contains(strings.ToLower(text), strings.ToLower(keyword)) {
			return true, keyword
		}
	}

	return false, ""
}

// DetectByCategory 按类别检测
func (d *AntiPatternDetector) DetectByCategory(content string, filePath string, category string) []DetectFinding {
	allFindings := d.Detect(content, filePath)
	var filtered []DetectFinding
	for _, finding := range allFindings {
		if strings.EqualFold(finding.Category, category) {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// DetectBySeverity 按严重程度检测
func (d *AntiPatternDetector) DetectBySeverity(content string, filePath string, severity string) []DetectFinding {
	allFindings := d.Detect(content, filePath)
	var filtered []DetectFinding
	for _, finding := range allFindings {
		if strings.EqualFold(finding.Severity, severity) {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}
