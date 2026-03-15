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

// UIUXAuditRequest 技术质量审查请求
type UIUXAuditRequest struct {
	Path          string `json:"path" jsonschema:"description=要审查的文件或目录路径（必填）"`
	Area          string `json:"area,omitempty" jsonschema:"description=要审查的特定区域/功能（可选，如 header, form）"`
	SeverityFilter string `json:"severity_filter,omitempty" jsonschema:"description=严重程度过滤（可选，如 critical,high）"`
}

// AuditFinding 审查发现
type AuditFinding struct {
	Location         string `json:"location"`
	Severity         string `json:"severity"`
	Category         string `json:"category"`
	Description      string `json:"description"`
	Impact           string `json:"impact"`
	Recommendation   string `json:"recommendation"`
	SuggestedCommand string `json:"suggested_command"`
}

// AuditReport 审查报告
type AuditReport struct {
	AntiPatternsVerdict string         `json:"anti_patterns_verdict"`
	ExecutiveSummary    ExecutiveSummary `json:"executive_summary"`
	Findings            []AuditFinding  `json:"findings"`
}

// ExecutiveSummary 执行摘要
type ExecutiveSummary struct {
	TotalIssues int            `json:"total_issues"`
	BySeverity  map[string]int `json:"by_severity"`
	TopIssues   []string       `json:"top_issues"`
}

// UIUXAuditResponse 技术质量审查响应
type UIUXAuditResponse struct {
	Report string `json:"report"`
}

// NewUIUXAuditTool 创建技术质量审查工具
func NewUIUXAuditTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("ui_ux_audit",
		"对前端代码进行全面的技术质量审查，包括可访问性、性能、主题、响应式设计和反模式检测。生成详细的审查报告，包含问题位置、严重程度和修复建议。",
		func(ctx context.Context, req *UIUXAuditRequest) (*UIUXAuditResponse, error) {
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
			var allFindings []AuditFinding
			var allAISlopFindings []AISlopFinding

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

				// 检测反模式
				antiPatternFindings := antiPatternDetector.Detect(contentStr, filePath)
				for _, finding := range antiPatternFindings {
					// 应用严重程度过滤
					if req.SeverityFilter != "" {
						severities := strings.Split(req.SeverityFilter, ",")
						matched := false
						for _, s := range severities {
							if strings.EqualFold(strings.TrimSpace(s), finding.Severity) {
								matched = true
								break
							}
						}
						if !matched {
							continue
						}
					}

					allFindings = append(allFindings, AuditFinding{
						Location:         finding.Location,
						Severity:         finding.Severity,
						Category:         finding.Category,
						Description:      finding.Description,
						Impact:           fmt.Sprintf("影响：%s", getImpactByCategory(finding.Category)),
						Recommendation:   finding.Recommendation,
						SuggestedCommand: finding.SuggestedCommand,
					})
				}

				// 检测 AI Slop
				aiSlopFindings := aiSlopDetector.Detect(contentStr, filePath)
				allAISlopFindings = append(allAISlopFindings, aiSlopFindings...)
				for _, finding := range aiSlopFindings {
					allFindings = append(allFindings, AuditFinding{
						Location:         finding.Location,
						Severity:         getSeverityByWeight(finding.Weight),
						Category:         "AI Slop",
						Description:      finding.Description,
						Impact:           "设计缺乏独特性，看起来像 AI 生成",
						Recommendation:   fmt.Sprintf("避免使用 %s，创造更独特的设计", finding.Description),
						SuggestedCommand: "critique",
					})
				}

				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("failed to scan files: %w", err)
			}

			// 生成 AI Slop 判定
			aiSlopVerdict := aiSlopDetector.GetVerdict(allAISlopFindings)

			// 生成执行摘要
			summary := generateExecutiveSummary(allFindings)

			// 构建报告
			report := AuditReport{
				AntiPatternsVerdict: aiSlopVerdict,
				ExecutiveSummary:    summary,
				Findings:            allFindings,
			}

			// 格式化为可读字符串
			reportStr := formatAuditReport(report)

			return &UIUXAuditResponse{Report: reportStr}, nil
		})
}

// generateExecutiveSummary 生成执行摘要
func generateExecutiveSummary(findings []AuditFinding) ExecutiveSummary {
	summary := ExecutiveSummary{
		BySeverity: make(map[string]int),
		TopIssues:  []string{},
	}

	for _, finding := range findings {
		summary.TotalIssues++
		summary.BySeverity[finding.Severity]++
	}

	// 提取前 5 个问题
	for i := 0; i < len(findings) && i < 5; i++ {
		summary.TopIssues = append(summary.TopIssues, fmt.Sprintf("%s: %s", findings[i].Severity, findings[i].Description))
	}

	return summary
}

// formatAuditReport 格式化审查报告
func formatAuditReport(report AuditReport) string {
	var sb strings.Builder

	sb.WriteString("=== 技术质量审查报告 ===\n\n")

	// AI Slop 判定
	sb.WriteString("## AI Slop 判定\n")
	sb.WriteString(report.AntiPatternsVerdict)
	sb.WriteString("\n\n")

	// 执行摘要
	sb.WriteString("## 执行摘要\n")
	sb.WriteString(fmt.Sprintf("总问题数：%d\n", report.ExecutiveSummary.TotalIssues))
	sb.WriteString("按严重程度分布：\n")
	for severity, count := range report.ExecutiveSummary.BySeverity {
		sb.WriteString(fmt.Sprintf("  - %s: %d\n", severity, count))
	}
	if len(report.ExecutiveSummary.TopIssues) > 0 {
		sb.WriteString("\n主要问题：\n")
		for i, issue := range report.ExecutiveSummary.TopIssues {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, issue))
		}
	}
	sb.WriteString("\n")

	// 详细发现
	sb.WriteString("## 详细发现\n\n")
	for i, finding := range report.Findings {
		sb.WriteString(fmt.Sprintf("### 问题 %d\n", i+1))
		sb.WriteString(fmt.Sprintf("**位置**：%s\n", finding.Location))
		sb.WriteString(fmt.Sprintf("**严重程度**：%s\n", finding.Severity))
		sb.WriteString(fmt.Sprintf("**分类**：%s\n", finding.Category))
		sb.WriteString(fmt.Sprintf("**描述**：%s\n", finding.Description))
		sb.WriteString(fmt.Sprintf("**影响**：%s\n", finding.Impact))
		sb.WriteString(fmt.Sprintf("**建议**：%s\n", finding.Recommendation))
		if finding.SuggestedCommand != "" {
			sb.WriteString(fmt.Sprintf("**建议命令**：%s\n", finding.SuggestedCommand))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// getImpactByCategory 根据类别获取影响描述
func getImpactByCategory(category string) string {
	impacts := map[string]string{
		"Typography":     "影响设计的独特性和品牌识别度",
		"Color":          "影响视觉一致性和可访问性",
		"Layout":         "影响用户体验和信息架构",
		"Animation":      "影响性能和用户体验",
		"Effect":         "影响设计质量和性能",
		"Accessibility":  "影响可访问性和合规性",
		"Performance":    "影响页面加载速度和用户体验",
		"Theming":        "影响主题一致性和维护性",
		"Responsive":     "影响移动端用户体验",
		"Interaction":    "影响用户交互体验",
		"UX Writing":     "影响用户体验和清晰度",
	}
	if impact, ok := impacts[category]; ok {
		return impact
	}
	return "影响代码质量和用户体验"
}

// getSeverityByWeight 根据权重获取严重程度
func getSeverityByWeight(weight int) string {
	if weight >= 8 {
		return "Critical"
	} else if weight >= 5 {
		return "High"
	} else if weight >= 3 {
		return "Medium"
	}
	return "Low"
}
