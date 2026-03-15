package ui_ux

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// UIUXCritiqueRequest UX 设计审查请求
type UIUXCritiqueRequest struct {
	Path  string `json:"path" jsonschema:"description=要审查的文件或目录路径（必填）"`
	Area  string `json:"area,omitempty" jsonschema:"description=要审查的特定区域/功能（可选）"`
	Focus string `json:"focus,omitempty" jsonschema:"description=重点关注领域（可选，如 visual-hierarchy, color-usage）"`
}

// CritiqueIssue 审查问题
type CritiqueIssue struct {
	What        string `json:"what"`
	WhyItMatters string `json:"why_it_matters"`
	Fix         string `json:"fix"`
	Command     string `json:"command"`
}

// CritiqueReport UX 设计审查报告
type CritiqueReport struct {
	AntiPatternsVerdict string        `json:"anti_patterns_verdict"`
	OverallImpression   string        `json:"overall_impression"`
	WhatsWorking        []string      `json:"whats_working"`
	PriorityIssues      []CritiqueIssue `json:"priority_issues"`
	MinorObservations   []string      `json:"minor_observations"`
}

// UIUXCritiqueResponse UX 设计审查响应
type UIUXCritiqueResponse struct {
	Report string `json:"report"`
}

// NewUIUXCritiqueTool 创建 UX 设计审查工具
func NewUIUXCritiqueTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("ui_ux_critique",
		"从 UX 设计角度评估界面效果，包括 AI Slop 检测、视觉层次、信息架构、情感共鸣等。提供设计改进建议和优先级排序。",
		func(ctx context.Context, req *UIUXCritiqueRequest) (*UIUXCritiqueResponse, error) {
			if req.Path == "" {
				return nil, fmt.Errorf("path is required")
			}

			// 初始化检测器
			antiPatternDetector, err := NewAntiPatternDetector()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize anti-pattern detector: %w", err)
			}

			aiSlopDetector, err := NewAISlopDetector()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize AI slop detector: %w", err)
			}

			// 收集所有发现
			var allAISlopFindings []AISlopFinding
			var allAntiPatternFindings []DetectFinding
			var positiveFindings []string

			// 扫描文件
			err = filepath.Walk(req.Path, func(filePath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// 只处理代码文件
				if info.IsDir() {
					return nil
				}

				ext := strings.ToLower(filepath.Ext(filePath))
				if ext != ".tsx" && ext != ".ts" && ext != ".jsx" && ext != ".js" && ext != ".css" && ext != ".scss" && ext != ".vue" && ext != ".svelte" {
					return nil
				}

				// 如果指定了 area，检查文件路径是否匹配
				if req.Area != "" && !strings.Contains(strings.ToLower(filePath), strings.ToLower(req.Area)) {
					return nil
				}

				// 读取文件内容
				content, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}

				contentStr := string(content)

				// 检测 AI Slop（最重要的检查）
				aiSlopFindings := aiSlopDetector.Detect(contentStr, filePath)
				allAISlopFindings = append(allAISlopFindings, aiSlopFindings...)

				// 检测反模式
				antiPatternFindings := antiPatternDetector.Detect(contentStr, filePath)
				allAntiPatternFindings = append(allAntiPatternFindings, antiPatternFindings...)

				// 检测正面特征（简单启发式）
				if strings.Contains(contentStr, "focus") && strings.Contains(contentStr, "outline") {
					positiveFindings = append(positiveFindings, "良好的焦点指示器设计")
				}
				if strings.Contains(contentStr, "prefers-reduced-motion") {
					positiveFindings = append(positiveFindings, "支持减少动画偏好")
				}
				if strings.Contains(contentStr, "aria-label") || strings.Contains(contentStr, "aria-labelledby") {
					positiveFindings = append(positiveFindings, "良好的 ARIA 标签使用")
				}

				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("failed to scan files: %w", err)
			}

			// 生成 AI Slop 判定
			aiSlopVerdict := aiSlopDetector.GetVerdict(allAISlopFindings)

			// 生成整体印象
			overallImpression := generateOverallImpression(allAISlopFindings, allAntiPatternFindings)

			// 提取正面发现（去重）
			whatsWorking := deduplicateStrings(positiveFindings)
			if len(whatsWorking) == 0 {
				whatsWorking = []string{"代码结构清晰", "文件组织良好"}
			}

			// 生成优先级问题
			priorityIssues := generatePriorityIssues(allAISlopFindings, allAntiPatternFindings, req.Focus)

			// 生成次要观察
			minorObservations := generateMinorObservations(allAntiPatternFindings)

			// 构建报告
			report := CritiqueReport{
				AntiPatternsVerdict: aiSlopVerdict,
				OverallImpression:   overallImpression,
				WhatsWorking:        whatsWorking,
				PriorityIssues:      priorityIssues,
				MinorObservations:    minorObservations,
			}

			// 格式化为可读字符串
			reportStr := formatCritiqueReport(report)

			return &UIUXCritiqueResponse{Report: reportStr}, nil
		})
}

// generateOverallImpression 生成整体印象
func generateOverallImpression(aiSlopFindings []AISlopFinding, antiPatternFindings []DetectFinding) string {
	totalIssues := len(aiSlopFindings) + len(antiPatternFindings)

	if totalIssues == 0 {
		return "设计质量良好，未发现明显的设计问题。"
	}

	if len(aiSlopFindings) > 5 {
		return "设计缺乏独特性，使用了多个常见的 AI 生成模板和模式。建议重新思考设计方向，创造更独特的视觉语言。"
	}

	if len(antiPatternFindings) > 10 {
		return "发现多个设计问题，需要系统性地改进设计质量和一致性。"
	}

	return "设计基本良好，但存在一些可以改进的地方。"
}

// generatePriorityIssues 生成优先级问题
func generatePriorityIssues(aiSlopFindings []AISlopFinding, antiPatternFindings []DetectFinding, focus string) []CritiqueIssue {
	var issues []CritiqueIssue

	// 优先处理 AI Slop 问题
	for i, finding := range aiSlopFindings {
		if i >= 3 { // 最多 3 个 AI Slop 问题
			break
		}

		// 如果指定了 focus，检查是否匹配
		if focus != "" {
			if !strings.Contains(strings.ToLower(finding.Category), strings.ToLower(focus)) {
				continue
			}
		}

		issue := CritiqueIssue{
			What:        finding.Description,
			WhyItMatters: "设计缺乏独特性，看起来像 AI 生成，影响品牌识别度",
			Fix:         fmt.Sprintf("避免使用 %s，创造更独特的设计方案", finding.Description),
			Command:     "normalize",
		}
		issues = append(issues, issue)
	}

	// 添加关键反模式问题
	criticalCount := 0
	for _, finding := range antiPatternFindings {
		if finding.Severity == "Critical" && criticalCount < 2 {
			issue := CritiqueIssue{
				What:        finding.Description,
				WhyItMatters: getImpactByCategory(finding.Category),
				Fix:         finding.Recommendation,
				Command:     finding.SuggestedCommand,
			}
			issues = append(issues, issue)
			criticalCount++
		}
	}

	// 如果问题少于 3 个，添加高优先级问题
	if len(issues) < 3 {
		for _, finding := range antiPatternFindings {
			if finding.Severity == "High" && len(issues) < 5 {
				issue := CritiqueIssue{
					What:        finding.Description,
					WhyItMatters: getImpactByCategory(finding.Category),
					Fix:         finding.Recommendation,
					Command:     finding.SuggestedCommand,
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

// generateMinorObservations 生成次要观察
func generateMinorObservations(antiPatternFindings []DetectFinding) []string {
	var observations []string

	for _, finding := range antiPatternFindings {
		if finding.Severity == "Low" || finding.Severity == "Medium" {
			observations = append(observations, fmt.Sprintf("%s: %s", finding.Category, finding.Description))
			if len(observations) >= 5 {
				break
			}
		}
	}

	return observations
}

// formatCritiqueReport 格式化审查报告
func formatCritiqueReport(report CritiqueReport) string {
	var sb strings.Builder

	sb.WriteString("=== UX 设计审查报告 ===\n\n")

	// AI Slop 判定
	sb.WriteString("## AI Slop 判定\n")
	sb.WriteString(report.AntiPatternsVerdict)
	sb.WriteString("\n\n")

	// 整体印象
	sb.WriteString("## 整体印象\n")
	sb.WriteString(report.OverallImpression)
	sb.WriteString("\n\n")

	// 做得好的地方
	sb.WriteString("## 做得好的地方\n")
	for i, item := range report.WhatsWorking {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
	sb.WriteString("\n")

	// 优先级问题
	sb.WriteString("## 优先级问题\n")
	for i, issue := range report.PriorityIssues {
		sb.WriteString(fmt.Sprintf("### 问题 %d\n", i+1))
		sb.WriteString(fmt.Sprintf("**问题**：%s\n", issue.What))
		sb.WriteString(fmt.Sprintf("**为什么重要**：%s\n", issue.WhyItMatters))
		sb.WriteString(fmt.Sprintf("**修复建议**：%s\n", issue.Fix))
		if issue.Command != "" {
			sb.WriteString(fmt.Sprintf("**建议命令**：%s\n", issue.Command))
		}
		sb.WriteString("\n")
	}

	// 次要观察
	if len(report.MinorObservations) > 0 {
		sb.WriteString("## 次要观察\n")
		for i, obs := range report.MinorObservations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, obs))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// deduplicateStrings 去重字符串切片
func deduplicateStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
