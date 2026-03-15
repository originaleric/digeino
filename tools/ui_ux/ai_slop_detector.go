package ui_ux

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
)

//go:embed data/ai_slop_patterns.csv
var aiSlopPatternsData embed.FS

// AISlopPattern AI Slop 特征定义
type AISlopPattern struct {
	Pattern     string
	Description string
	Weight      int    // 权重（用于计算分数）
	Detection   string // 正则表达式或关键词
	Category    string // Color/Typography/Layout/Effect/Animation
	Examples    string
}

// AISlopDetector AI Slop 检测器
type AISlopDetector struct {
	patterns []AISlopPattern
}

// NewAISlopDetector 创建 AI Slop 检测器
func NewAISlopDetector() (*AISlopDetector, error) {
	detector := &AISlopDetector{}
	if err := detector.loadPatterns(); err != nil {
		return nil, fmt.Errorf("failed to load AI slop patterns: %w", err)
	}
	return detector, nil
}

// loadPatterns 从 CSV 文件加载 AI Slop 特征
func (d *AISlopDetector) loadPatterns() error {
	f, err := aiSlopPatternsData.Open("data/ai_slop_patterns.csv")
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	headers, err := reader.Read()
	if err != nil {
		return err
	}

	var patterns []AISlopPattern
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

		// 解析权重
		weight := 5 // 默认权重
		if wStr := fieldMap["Weight"]; wStr != "" {
			fmt.Sscanf(wStr, "%d", &weight)
		}

		pattern := AISlopPattern{
			Pattern:     fieldMap["Pattern"],
			Description: fieldMap["Description"],
			Weight:      weight,
			Detection:   fieldMap["Detection"],
			Category:    fieldMap["Category"],
			Examples:    fieldMap["Examples"],
		}

		patterns = append(patterns, pattern)
	}

	d.patterns = patterns
	return nil
}

// AISlopFinding AI Slop 检测结果
type AISlopFinding struct {
	Pattern     AISlopPattern
	Location    string // 文件路径和行号
	MatchText   string // 匹配的文本
	Weight      int
	Category    string
	Description string
}

// Detect 检测代码中的 AI Slop 特征
func (d *AISlopDetector) Detect(content string, filePath string) []AISlopFinding {
	var findings []AISlopFinding

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineNum := i + 1
		for _, pattern := range d.patterns {
			// 尝试正则表达式匹配
			matched, matchText := d.matchPattern(pattern.Detection, line)
			if matched {
				finding := AISlopFinding{
					Pattern:     pattern,
					Location:    fmt.Sprintf("%s:%d", filePath, lineNum),
					MatchText:   matchText,
					Weight:      pattern.Weight,
					Category:    pattern.Category,
					Description: pattern.Description,
				}
				findings = append(findings, finding)
			}
		}
	}

	return findings
}

// matchPattern 匹配模式（支持正则表达式和关键词）
func (d *AISlopDetector) matchPattern(pattern, text string) (bool, string) {
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

// CalculateScore 计算 AI Slop 分数（0-100）
func (d *AISlopDetector) CalculateScore(findings []AISlopFinding) int {
	if len(findings) == 0 {
		return 0
	}

	totalWeight := 0
	for _, finding := range findings {
		totalWeight += finding.Weight
	}

	// 分数 = (总权重 / 最大可能权重) * 100
	// 假设每个文件最多检测到 10 个特征，每个特征最大权重 10
	maxPossibleWeight := 10 * 10
	score := (totalWeight * 100) / maxPossibleWeight
	if score > 100 {
		score = 100
	}

	return score
}

// GetVerdict 获取 AI Slop 判定结果
func (d *AISlopDetector) GetVerdict(findings []AISlopFinding) string {
	if len(findings) == 0 {
		return "PASS - 未检测到明显的 AI Slop 特征"
	}

	score := d.CalculateScore(findings)
	count := len(findings)

	if score >= 70 {
		return fmt.Sprintf("FAIL - 检测到 %d 个 AI Slop 特征，分数 %d/100（严重）", count, score)
	} else if score >= 40 {
		return fmt.Sprintf("WARNING - 检测到 %d 个 AI Slop 特征，分数 %d/100（中等）", count, score)
	} else {
		return fmt.Sprintf("PASS - 检测到 %d 个 AI Slop 特征，分数 %d/100（轻微）", count, score)
	}
}
