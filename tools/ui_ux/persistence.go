package ui_ux

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/originaleric/digeino/config"
)

// PersistenceManager 持久化管理器
type PersistenceManager struct {
	baseDir string
	appName string // 应用/agent 名称，用于隔离不同应用的存储
}

// NewPersistenceManager 创建持久化管理器
// baseDir: 基础目录，如果为空则从配置读取
// appName: 应用/agent 名称，用于隔离存储（可选）
func NewPersistenceManager(baseDir string, appName string) *PersistenceManager {
	if baseDir == "" {
		// 从配置读取
		cfg := config.Get()
		if cfg.UIUX.Storage.BaseDir != "" {
			baseDir = cfg.UIUX.Storage.BaseDir
		} else {
			baseDir = "storage/app/ui_ux"
		}
	}

	return &PersistenceManager{
		baseDir: baseDir,
		appName: appName,
	}
}

// GetBaseDir 获取基础目录
func (p *PersistenceManager) GetBaseDir() string {
	return p.baseDir
}

// GetAppName 获取应用名称
func (p *PersistenceManager) GetAppName() string {
	return p.appName
}

// PersistDesignSystem 保存设计系统到文件（Master + Overrides 模式）
func (p *PersistenceManager) PersistDesignSystem(ds *DesignSystem, projectSlug string, pageName string) error {
	if projectSlug == "" {
		projectSlug = strings.ToLower(strings.ReplaceAll(ds.ProjectName, " ", "-"))
	}

	// 构建存储路径：如果指定了 appName，则添加应用隔离层
	var designSystemDir string
	if p.appName != "" {
		// 应用隔离：{baseDir}/{app-name}/design-system/{project}
		designSystemDir = filepath.Join(p.baseDir, p.appName, "design-system", projectSlug)
	} else {
		// 默认：{baseDir}/design-system/{project}
		designSystemDir = filepath.Join(p.baseDir, "design-system", projectSlug)
	}
	pagesDir := filepath.Join(designSystemDir, "pages")

	// 创建目录
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// 生成并保存 MASTER.md
	masterFile := filepath.Join(designSystemDir, "MASTER.md")
	masterContent := p.formatMasterMD(ds)
	if err := os.WriteFile(masterFile, []byte(masterContent), 0644); err != nil {
		return fmt.Errorf("failed to write MASTER.md: %w", err)
	}

	// 如果指定了页面名称，创建页面覆盖文件
	if pageName != "" {
		pageFile := filepath.Join(pagesDir, strings.ToLower(strings.ReplaceAll(pageName, " ", "-"))+".md")
		pageContent := p.formatPageOverrideMD(ds, pageName)
		if err := os.WriteFile(pageFile, []byte(pageContent), 0644); err != nil {
			return fmt.Errorf("failed to write page override: %w", err)
		}
	}

	return nil
}

// formatMasterMD 格式化 Master 文件
func (p *PersistenceManager) formatMasterMD(ds *DesignSystem) string {
	var sb strings.Builder
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	sb.WriteString("# Design System Master File\n\n")
	sb.WriteString("> **LOGIC:** When building a specific page, first check `design-system/pages/[page-name].md`.\n")
	sb.WriteString("> If that file exists, its rules **override** this Master file.\n")
	sb.WriteString("> If not, strictly follow the rules below.\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n", ds.ProjectName))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", timestamp))
	sb.WriteString(fmt.Sprintf("**Category:** %s\n\n", ds.Category))
	sb.WriteString("---\n\n")

	// Global Rules
	sb.WriteString("## Global Rules\n\n")

	// Color Palette
	sb.WriteString("### Color Palette\n\n")
	sb.WriteString("| Role | Hex | CSS Variable |\n")
	sb.WriteString("|------|-----|--------------|\n")
	sb.WriteString(fmt.Sprintf("| Primary | `%s` | `--color-primary` |\n", ds.Colors.Primary))
	sb.WriteString(fmt.Sprintf("| Secondary | `%s` | `--color-secondary` |\n", ds.Colors.Secondary))
	sb.WriteString(fmt.Sprintf("| CTA/Accent | `%s` | `--color-cta` |\n", ds.Colors.CTA))
	sb.WriteString(fmt.Sprintf("| Background | `%s` | `--color-background` |\n", ds.Colors.Background))
	sb.WriteString(fmt.Sprintf("| Text | `%s` | `--color-text` |\n", ds.Colors.Text))
	sb.WriteString("\n")
	if ds.Colors.Notes != "" {
		sb.WriteString(fmt.Sprintf("**Color Notes:** %s\n\n", ds.Colors.Notes))
	}

	// Typography
	sb.WriteString("### Typography\n\n")
	sb.WriteString(fmt.Sprintf("- **Heading Font:** %s\n", ds.Typography.Heading))
	sb.WriteString(fmt.Sprintf("- **Body Font:** %s\n", ds.Typography.Body))
	if ds.Typography.Mood != "" {
		sb.WriteString(fmt.Sprintf("- **Mood:** %s\n", ds.Typography.Mood))
	}
	if ds.Typography.GoogleFontsURL != "" {
		sb.WriteString(fmt.Sprintf("- **Google Fonts:** [%s + %s](%s)\n", ds.Typography.Heading, ds.Typography.Body, ds.Typography.GoogleFontsURL))
	}
	sb.WriteString("\n")
	if ds.Typography.CSSImport != "" {
		sb.WriteString("**CSS Import:**\n")
		sb.WriteString("```css\n")
		sb.WriteString(ds.Typography.CSSImport)
		sb.WriteString("\n```\n\n")
	}

	// Spacing Variables
	sb.WriteString("### Spacing Variables\n\n")
	sb.WriteString("| Token | Value | Usage |\n")
	sb.WriteString("|-------|-------|-------|\n")
	sb.WriteString("| `--space-xs` | `4px` / `0.25rem` | Tight gaps |\n")
	sb.WriteString("| `--space-sm` | `8px` / `0.5rem` | Icon gaps, inline spacing |\n")
	sb.WriteString("| `--space-md` | `16px` / `1rem` | Standard padding |\n")
	sb.WriteString("| `--space-lg` | `24px` / `1.5rem` | Section padding |\n")
	sb.WriteString("| `--space-xl` | `32px` / `2rem` | Large gaps |\n")
	sb.WriteString("| `--space-2xl` | `48px` / `3rem` | Section margins |\n")
	sb.WriteString("| `--space-3xl` | `64px` / `4rem` | Hero padding |\n\n")

	// Shadow Depths
	sb.WriteString("### Shadow Depths\n\n")
	sb.WriteString("| Level | Value | Usage |\n")
	sb.WriteString("|-------|-------|-------|\n")
	sb.WriteString("| `--shadow-sm` | `0 1px 2px rgba(0,0,0,0.05)` | Subtle lift |\n")
	sb.WriteString("| `--shadow-md` | `0 4px 6px rgba(0,0,0,0.1)` | Cards, buttons |\n")
	sb.WriteString("| `--shadow-lg` | `0 10px 15px rgba(0,0,0,0.1)` | Modals, dropdowns |\n")
	sb.WriteString("| `--shadow-xl` | `0 20px 25px rgba(0,0,0,0.15)` | Hero images, featured cards |\n\n")

	// Component Specs
	sb.WriteString("---\n\n")
	sb.WriteString("## Component Specs\n\n")

	// Buttons
	sb.WriteString("### Buttons\n\n")
	sb.WriteString("```css\n")
	sb.WriteString("/* Primary Button */\n")
	sb.WriteString(".btn-primary {\n")
	sb.WriteString(fmt.Sprintf("  background: %s;\n", ds.Colors.CTA))
	sb.WriteString("  color: white;\n")
	sb.WriteString("  padding: 12px 24px;\n")
	sb.WriteString("  border-radius: 8px;\n")
	sb.WriteString("  font-weight: 600;\n")
	sb.WriteString("  transition: all 200ms ease;\n")
	sb.WriteString("  cursor: pointer;\n")
	sb.WriteString("}\n\n")
	sb.WriteString(".btn-primary:hover {\n")
	sb.WriteString("  opacity: 0.9;\n")
	sb.WriteString("  transform: translateY(-1px);\n")
	sb.WriteString("}\n\n")
	sb.WriteString("/* Secondary Button */\n")
	sb.WriteString(".btn-secondary {\n")
	sb.WriteString("  background: transparent;\n")
	sb.WriteString(fmt.Sprintf("  color: %s;\n", ds.Colors.Primary))
	sb.WriteString(fmt.Sprintf("  border: 2px solid %s;\n", ds.Colors.Primary))
	sb.WriteString("  padding: 12px 24px;\n")
	sb.WriteString("  border-radius: 8px;\n")
	sb.WriteString("  font-weight: 600;\n")
	sb.WriteString("  transition: all 200ms ease;\n")
	sb.WriteString("  cursor: pointer;\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	// Cards
	sb.WriteString("### Cards\n\n")
	sb.WriteString("```css\n")
	sb.WriteString(".card {\n")
	sb.WriteString(fmt.Sprintf("  background: %s;\n", ds.Colors.Background))
	sb.WriteString("  border-radius: 12px;\n")
	sb.WriteString("  padding: 24px;\n")
	sb.WriteString("  box-shadow: var(--shadow-md);\n")
	sb.WriteString("  transition: all 200ms ease;\n")
	sb.WriteString("  cursor: pointer;\n")
	sb.WriteString("}\n\n")
	sb.WriteString(".card:hover {\n")
	sb.WriteString("  box-shadow: var(--shadow-lg);\n")
	sb.WriteString("  transform: translateY(-2px);\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	// Style Guidelines
	sb.WriteString("---\n\n")
	sb.WriteString("## Style Guidelines\n\n")
	sb.WriteString(fmt.Sprintf("**Style:** %s\n\n", ds.Style.Name))
	if ds.Style.Keywords != "" {
		sb.WriteString(fmt.Sprintf("**Keywords:** %s\n\n", ds.Style.Keywords))
	}
	if ds.Style.BestFor != "" {
		sb.WriteString(fmt.Sprintf("**Best For:** %s\n\n", ds.Style.BestFor))
	}
	if ds.KeyEffects != "" {
		sb.WriteString(fmt.Sprintf("**Key Effects:** %s\n\n", ds.KeyEffects))
	}

	// Page Pattern
	sb.WriteString("### Page Pattern\n\n")
	sb.WriteString(fmt.Sprintf("**Pattern Name:** %s\n\n", ds.Pattern.Name))
	if ds.Pattern.Conversion != "" {
		sb.WriteString(fmt.Sprintf("- **Conversion Strategy:** %s\n", ds.Pattern.Conversion))
	}
	if ds.Pattern.CTAPlacement != "" {
		sb.WriteString(fmt.Sprintf("- **CTA Placement:** %s\n", ds.Pattern.CTAPlacement))
	}
	if ds.Pattern.Sections != "" {
		sb.WriteString(fmt.Sprintf("- **Section Order:** %s\n", ds.Pattern.Sections))
	}
	sb.WriteString("\n")

	// Anti-Patterns
	sb.WriteString("---\n\n")
	sb.WriteString("## Anti-Patterns (Do NOT Use)\n\n")
	if ds.AntiPatterns != "" {
		antiList := strings.Split(ds.AntiPatterns, "+")
		for _, anti := range antiList {
			if trimmed := strings.TrimSpace(anti); trimmed != "" {
				sb.WriteString(fmt.Sprintf("- ❌ %s\n", trimmed))
			}
		}
	}
	sb.WriteString("\n")
	sb.WriteString("### Additional Forbidden Patterns\n\n")
	sb.WriteString("- ❌ **Emojis as icons** — Use SVG icons (Heroicons, Lucide, Simple Icons)\n")
	sb.WriteString("- ❌ **Missing cursor:pointer** — All clickable elements must have cursor:pointer\n")
	sb.WriteString("- ❌ **Layout-shifting hovers** — Avoid scale transforms that shift layout\n")
	sb.WriteString("- ❌ **Low contrast text** — Maintain 4.5:1 minimum contrast ratio\n")
	sb.WriteString("- ❌ **Instant state changes** — Always use transitions (150-300ms)\n")
	sb.WriteString("- ❌ **Invisible focus states** — Focus states must be visible for a11y\n\n")

	// Pre-Delivery Checklist
	sb.WriteString("---\n\n")
	sb.WriteString("## Pre-Delivery Checklist\n\n")
	sb.WriteString("Before delivering any UI code, verify:\n\n")
	sb.WriteString("- [ ] No emojis used as icons (use SVG instead)\n")
	sb.WriteString("- [ ] All icons from consistent icon set (Heroicons/Lucide)\n")
	sb.WriteString("- [ ] `cursor-pointer` on all clickable elements\n")
	sb.WriteString("- [ ] Hover states with smooth transitions (150-300ms)\n")
	sb.WriteString("- [ ] Light mode: text contrast 4.5:1 minimum\n")
	sb.WriteString("- [ ] Focus states visible for keyboard navigation\n")
	sb.WriteString("- [ ] `prefers-reduced-motion` respected\n")
	sb.WriteString("- [ ] Responsive: 375px, 768px, 1024px, 1440px\n")
	sb.WriteString("- [ ] No content hidden behind fixed navbars\n")
	sb.WriteString("- [ ] No horizontal scroll on mobile\n\n")

	return sb.String()
}

// formatPageOverrideMD 格式化页面覆盖文件
func (p *PersistenceManager) formatPageOverrideMD(ds *DesignSystem, pageName string) string {
	var sb strings.Builder
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	pageTitle := strings.Title(strings.ReplaceAll(strings.ReplaceAll(pageName, "-", " "), "_", " "))

	sb.WriteString(fmt.Sprintf("# %s Page Overrides\n\n", pageTitle))
	sb.WriteString(fmt.Sprintf("> **PROJECT:** %s\n", ds.ProjectName))
	sb.WriteString(fmt.Sprintf("> **Generated:** %s\n", timestamp))
	sb.WriteString("\n")
	sb.WriteString("> ⚠️ **IMPORTANT:** Rules in this file **override** the Master file (`design-system/MASTER.md`).\n")
	sb.WriteString("> Only deviations from the Master are documented here. For all other rules, refer to the Master.\n\n")
	sb.WriteString("---\n\n")

	// Page-specific rules
	sb.WriteString("## Page-Specific Rules\n\n")
	sb.WriteString("### Layout Overrides\n\n")
	sb.WriteString("- No overrides — use Master layout\n\n")
	sb.WriteString("### Spacing Overrides\n\n")
	sb.WriteString("- No overrides — use Master spacing\n\n")
	sb.WriteString("### Typography Overrides\n\n")
	sb.WriteString("- No overrides — use Master typography\n\n")
	sb.WriteString("### Color Overrides\n\n")
	sb.WriteString("- No overrides — use Master colors\n\n")
	sb.WriteString("### Component Overrides\n\n")
	sb.WriteString("- No overrides — use Master component specs\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("## Page-Specific Components\n\n")
	sb.WriteString("- No unique components for this page\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("## Recommendations\n\n")
	sb.WriteString("- Refer to MASTER.md for all design rules\n")
	sb.WriteString("- Add specific overrides as needed for this page\n\n")

	return sb.String()
}
