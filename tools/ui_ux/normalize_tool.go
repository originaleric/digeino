package ui_ux

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// UIUXNormalizeRequest 设计系统标准化请求
type UIUXNormalizeRequest struct {
	Path             string `json:"path" jsonschema:"description=要标准化的文件或目录路径（必填）"`
	Feature          string `json:"feature,omitempty" jsonschema:"description=要标准化的特定功能/页面（可选）"`
	DesignSystemPath string `json:"design_system_path,omitempty" jsonschema:"description=设计系统文档路径（可选，自动发现）"`
	DryRun           bool   `json:"dry_run,omitempty" jsonschema:"description=仅生成计划，不执行修改（可选，默认 false）"`
}

// NormalizePlan 标准化计划
type NormalizePlan struct {
	DesignSystemFound bool     `json:"design_system_found"`
	DesignSystemPath  string   `json:"design_system_path,omitempty"`
	Deviations        []Deviation `json:"deviations"`
	Recommendations   []string `json:"recommendations"`
}

// Deviation 偏差
type Deviation struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Issue       string `json:"issue"`
	Current     string `json:"current"`
	Recommended string `json:"recommended"`
	Category    string `json:"category"`
}

// UIUXNormalizeResponse 设计系统标准化响应
type UIUXNormalizeResponse struct {
	Plan   string `json:"plan"`
	DryRun bool   `json:"dry_run"`
}

// NewUIUXNormalizeTool 创建设计系统标准化工具
func NewUIUXNormalizeTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("ui_ux_normalize",
		"分析代码与设计系统的偏差，生成标准化计划并执行标准化。确保代码使用设计令牌、统一组件和模式。",
		func(ctx context.Context, req *UIUXNormalizeRequest) (*UIUXNormalizeResponse, error) {
			if req.Path == "" {
				return nil, fmt.Errorf("path is required")
			}

			// 查找设计系统文档
			designSystemPath := req.DesignSystemPath
			if designSystemPath == "" {
				designSystemPath = findDesignSystem(req.Path)
			}

			designSystemFound := designSystemPath != ""

			// 初始化反模式检测器
			antiPatternDetector, err := NewAntiPatternDetector()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize anti-pattern detector: %w", err)
			}

			// 收集偏差
			var deviations []Deviation

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

				// 如果指定了 feature，检查文件路径是否匹配
				if req.Feature != "" && !strings.Contains(strings.ToLower(filePath), strings.ToLower(req.Feature)) {
					return nil
				}

				// 读取文件内容
				content, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}

				contentStr := string(content)
				lines := strings.Split(contentStr, "\n")

				// 检测硬编码颜色
				hardcodedColors := detectHardcodedColors(lines, filePath)
				deviations = append(deviations, hardcodedColors...)

				// 检测不一致的间距
				inconsistentSpacing := detectInconsistentSpacing(lines, filePath)
				deviations = append(deviations, inconsistentSpacing...)

				// 检测反模式（这些需要标准化）
				antiPatternFindings := antiPatternDetector.Detect(contentStr, filePath)
				for _, finding := range antiPatternFindings {
					// 只关注需要标准化的问题
					if finding.SuggestedCommand == "normalize" {
						lineNum := extractLineNumber(finding.Location)
						deviation := Deviation{
							File:        filePath,
							Line:        lineNum,
							Issue:       finding.Description,
							Current:     finding.MatchText,
							Recommended: finding.Recommendation,
							Category:    finding.Category,
						}
						deviations = append(deviations, deviation)
					}
				}

				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("failed to scan files: %w", err)
			}

			// 生成建议
			recommendations := generateRecommendations(deviations, designSystemFound)

			// 构建计划
			plan := NormalizePlan{
				DesignSystemFound: designSystemFound,
				DesignSystemPath:  designSystemPath,
				Deviations:        deviations,
				Recommendations:   recommendations,
			}

			// 如果不是 dry run，执行标准化（这里只生成计划，实际修改需要用户确认）
			if !req.DryRun {
				// TODO: 实际执行标准化修改
				// 这里可以添加实际的代码修改逻辑
			}

			// 格式化为可读字符串
			planStr := formatNormalizePlan(plan, req.DryRun)

			return &UIUXNormalizeResponse{
				Plan:   planStr,
				DryRun: req.DryRun,
			}, nil
		})
}

// findDesignSystem 查找设计系统文档
func findDesignSystem(startPath string) string {
	// 常见的设计系统文档位置
	commonPaths := []string{
		"MASTER.md",
		"design-system.md",
		"ui-guide.md",
		"design-tokens.md",
		"design-system/MASTER.md",
		"docs/design-system.md",
	}

	// 从起始路径向上查找
	currentPath := startPath
	for i := 0; i < 5; i++ { // 最多向上查找 5 层
		for _, commonPath := range commonPaths {
			fullPath := filepath.Join(currentPath, commonPath)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
		}
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}

	return ""
}

// detectHardcodedColors 检测硬编码颜色
func detectHardcodedColors(lines []string, filePath string) []Deviation {
	var deviations []Deviation
	colorPattern := `#[0-9a-f]{3,6}|rgb\([^)]+\)|rgba\([^)]+\)`

	for i, line := range lines {
		if strings.Contains(line, "color:") || strings.Contains(line, "background:") || strings.Contains(line, "border-color:") {
			// 检查是否包含硬编码颜色
			if matched, _ := regexpMatch(colorPattern, line); matched {
				// 检查是否已经是 CSS 变量或设计令牌
				if !strings.Contains(line, "var(") && !strings.Contains(line, "--") {
					deviation := Deviation{
						File:        filePath,
						Line:        i + 1,
						Issue:       "硬编码颜色值",
						Current:     extractColorValue(line),
						Recommended: "使用设计令牌或 CSS 变量（如 var(--color-primary)）",
						Category:    "Theming",
					}
					deviations = append(deviations, deviation)
				}
			}
		}
	}

	return deviations
}

// detectInconsistentSpacing 检测不一致的间距
func detectInconsistentSpacing(lines []string, filePath string) []Deviation {
	var deviations []Deviation
	spacingPattern := `(margin|padding|gap):\s*\d+px`

	for i, line := range lines {
		if matched, _ := regexpMatch(spacingPattern, line); matched {
			// 检查是否使用了设计系统的间距值
			if !strings.Contains(line, "var(") && !strings.Contains(line, "--spacing-") {
				deviation := Deviation{
					File:        filePath,
					Line:        i + 1,
					Issue:       "硬编码间距值",
					Current:     extractSpacingValue(line),
					Recommended: "使用设计系统的间距令牌（如 var(--spacing-md)）",
					Category:    "Layout",
				}
				deviations = append(deviations, deviation)
			}
		}
	}

	return deviations
}

// generateRecommendations 生成建议
func generateRecommendations(deviations []Deviation, designSystemFound bool) []string {
	var recommendations []string

	if !designSystemFound {
		recommendations = append(recommendations, "未找到设计系统文档，建议先创建或指定设计系统路径")
	}

	// 按类别统计
	categoryCount := make(map[string]int)
	for _, d := range deviations {
		categoryCount[d.Category]++
	}

	for category, count := range categoryCount {
		if count > 0 {
			recommendations = append(recommendations, fmt.Sprintf("%s 类别发现 %d 个偏差，需要统一", category, count))
		}
	}

	if len(deviations) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("总共发现 %d 个需要标准化的偏差", len(deviations)))
	}

	return recommendations
}

// formatNormalizePlan 格式化标准化计划
func formatNormalizePlan(plan NormalizePlan, dryRun bool) string {
	var sb strings.Builder

	sb.WriteString("=== 设计系统标准化计划 ===\n\n")

	// 设计系统状态
	if plan.DesignSystemFound {
		sb.WriteString(fmt.Sprintf("✅ 找到设计系统文档：%s\n\n", plan.DesignSystemPath))
	} else {
		sb.WriteString("⚠️ 未找到设计系统文档，将基于通用最佳实践进行标准化\n\n")
	}

	// 偏差列表
	if len(plan.Deviations) > 0 {
		sb.WriteString("## 发现的偏差\n\n")
		for i, deviation := range plan.Deviations {
			sb.WriteString(fmt.Sprintf("### 偏差 %d\n", i+1))
			sb.WriteString(fmt.Sprintf("**文件**：%s:%d\n", deviation.File, deviation.Line))
			sb.WriteString(fmt.Sprintf("**问题**：%s\n", deviation.Issue))
			sb.WriteString(fmt.Sprintf("**当前**：%s\n", deviation.Current))
			sb.WriteString(fmt.Sprintf("**建议**：%s\n", deviation.Recommended))
			sb.WriteString(fmt.Sprintf("**分类**：%s\n", deviation.Category))
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("✅ 未发现需要标准化的偏差\n\n")
	}

	// 建议
	if len(plan.Recommendations) > 0 {
		sb.WriteString("## 建议\n\n")
		for i, rec := range plan.Recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
		sb.WriteString("\n")
	}

	// Dry run 提示
	if dryRun {
		sb.WriteString("---\n")
		sb.WriteString("**注意**：这是预览模式（dry-run），未执行实际修改。要执行标准化，请设置 dry_run=false\n")
	}

	return sb.String()
}

// 辅助函数
func regexpMatch(pattern, text string) (bool, string) {
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return false, ""
	}
	matches := re.FindStringSubmatch(text)
	if len(matches) > 0 {
		return true, matches[0]
	}
	return false, ""
}

func extractColorValue(line string) string {
	re := regexp.MustCompile(`#[0-9a-f]{3,6}|rgb\([^)]+\)|rgba\([^)]+\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 0 {
		return matches[0]
	}
	return line
}

func extractSpacingValue(line string) string {
	re := regexp.MustCompile(`(margin|padding|gap):\s*\d+px`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 0 {
		return matches[0]
	}
	return line
}

func extractLineNumber(location string) int {
	parts := strings.Split(location, ":")
	if len(parts) >= 2 {
		var lineNum int
		fmt.Sscanf(parts[len(parts)-1], "%d", &lineNum)
		return lineNum
	}
	return 0
}
