package render

import (
	"strings"
	"sync"
)

// HTMLTagClass is a single element: tag name + optional CSS class（thinking/code 共用此形状）.
type HTMLTagClass struct {
	Tag   string `yaml:"tag"`
	Class string `yaml:"class"`
}

// HTMLPresentationConfig drives pkg/render/html output.
type HTMLPresentationConfig struct {
	Thinking HTMLThinkingPresentation `yaml:"thinking"`
	Code     HTMLCodePresentation     `yaml:"code"`
	Markdown HTMLMarkdownPresentation `yaml:"markdown"`
	Document HTMLDocumentPresentation `yaml:"document"`
}

// HTMLThinkingPresentation：外层 outer、内层 inner，与 code 字段命名一致。
type HTMLThinkingPresentation struct {
	Outer HTMLTagClass `yaml:"outer"`
	Inner HTMLTagClass `yaml:"inner"`
}

// HTMLCodePresentation：外层 pre、内层 code（tag 可配但默认 pre/code，且受白名单约束）。
type HTMLCodePresentation struct {
	Outer HTMLTagClass `yaml:"outer"`
	Inner struct {
		Tag                 string `yaml:"tag"`
		Class               string `yaml:"class"`
		LanguageClassPrefix string `yaml:"language_class_prefix"`
	} `yaml:"inner"`
}

// HTMLMarkdownPresentation：仅一项策略。新键 sanitize；旧键 sanitize_policy 仍可读。
type HTMLMarkdownPresentation struct {
	Sanitize       string `yaml:"sanitize"`
	SanitizePolicy string `yaml:"sanitize_policy"`
}

func (m HTMLMarkdownPresentation) normalizedSanitize() string {
	if strings.TrimSpace(m.Sanitize) != "" {
		return strings.TrimSpace(m.Sanitize)
	}
	return strings.TrimSpace(m.SanitizePolicy)
}

// HTMLDocumentPresentation configures WrapDocument shell.
type HTMLDocumentPresentation struct {
	Lang      string `yaml:"lang"`
	InlineCSS string `yaml:"inline_css"`
}

// thinkingRawYAML 解码 thinking 段（支持新 outer/inner 与旧 wrapper_* / content_*）。
type thinkingRawYAML struct {
	Outer HTMLTagClass `yaml:"outer"`
	Inner HTMLTagClass `yaml:"inner"`
	// legacy
	WrapperTag   string `yaml:"wrapper_tag"`
	WrapperClass string `yaml:"wrapper_class"`
	ContentTag   string `yaml:"content_tag"`
	ContentClass string `yaml:"content_class"`
}

// codeRawYAML 解码 code 段（支持新 outer/inner 与旧 pre_class）。
type codeRawYAML struct {
	Outer HTMLTagClass `yaml:"outer"`
	Inner struct {
		Tag                 string `yaml:"tag"`
		Class               string `yaml:"class"`
		LanguageClassPrefix string `yaml:"language_class_prefix"`
	} `yaml:"inner"`
	PreClass            string `yaml:"pre_class"`
	LanguageClassPrefix string `yaml:"language_class_prefix"`
}

// HTMLPresentationConfigRaw 供 yaml.Unmarshal + MigrateHTMLPresentationRaw 使用。
type HTMLPresentationConfigRaw struct {
	Thinking thinkingRawYAML              `yaml:"thinking"`
	Code     codeRawYAML                  `yaml:"code"`
	Markdown HTMLMarkdownPresentation     `yaml:"markdown"`
	Document HTMLDocumentPresentation     `yaml:"document"`
}

// MigrateHTMLPresentationRaw 将原始 YAML 结构转为规范化配置（含旧键兼容）。
func MigrateHTMLPresentationRaw(raw HTMLPresentationConfigRaw) HTMLPresentationConfig {
	out := HTMLPresentationConfig{
		Markdown: raw.Markdown,
		Document: raw.Document,
	}
	th := raw.Thinking
	hasNew := th.Outer.Tag != "" || th.Outer.Class != "" || th.Inner.Tag != "" || th.Inner.Class != ""
	hasLeg := th.WrapperTag != "" || th.WrapperClass != "" || th.ContentTag != "" || th.ContentClass != ""
	if hasNew {
		out.Thinking.Outer = th.Outer
		out.Thinking.Inner = th.Inner
	} else if hasLeg {
		out.Thinking.Outer = HTMLTagClass{Tag: th.WrapperTag, Class: th.WrapperClass}
		out.Thinking.Inner = HTMLTagClass{Tag: th.ContentTag, Class: th.ContentClass}
	}
	co := raw.Code
	hasNewCode := co.Outer.Tag != "" || co.Outer.Class != "" || co.Inner.Tag != "" || co.Inner.Class != "" || co.Inner.LanguageClassPrefix != ""
	hasLegCode := co.PreClass != "" || co.LanguageClassPrefix != ""
	if hasNewCode {
		out.Code.Outer = co.Outer
		out.Code.Inner = co.Inner
	} else if hasLegCode {
		out.Code.Outer = HTMLTagClass{Tag: "pre", Class: co.PreClass}
		out.Code.Inner.Tag = "code"
		out.Code.Inner.LanguageClassPrefix = co.LanguageClassPrefix
	}
	return out
}

var (
	htmlPresMu     sync.Mutex
	htmlPresCached *HTMLPresentationConfig
	htmlPresLoaded bool
)

func DefaultHTMLPresentation() HTMLPresentationConfig {
	htmlPresMu.Lock()
	defer htmlPresMu.Unlock()
	if htmlPresLoaded && htmlPresCached != nil {
		return *htmlPresCached
	}
	rc, err := renderConfigFromYAMLBytes(embeddedDefaultYAML)
	if err != nil {
		c := NormalizeHTMLPresentation(BuiltinHTMLPresentationDefaults())
		htmlPresCached = &c
		htmlPresLoaded = true
		return *htmlPresCached
	}
	htmlPresCached = &rc.Render
	htmlPresLoaded = true
	return *htmlPresCached
}

func BuiltinHTMLPresentationDefaults() HTMLPresentationConfig {
	return HTMLPresentationConfig{
		Thinking: HTMLThinkingPresentation{
			Outer: HTMLTagClass{Tag: "aside", Class: "llm-thinking"},
			Inner: HTMLTagClass{Tag: "pre", Class: "llm-thinking-pre"},
		},
		Code: HTMLCodePresentation{
			Outer: HTMLTagClass{Tag: "pre", Class: "llm-code"},
			Inner: struct {
				Tag                 string `yaml:"tag"`
				Class               string `yaml:"class"`
				LanguageClassPrefix string `yaml:"language_class_prefix"`
			}{Tag: "code", Class: "", LanguageClassPrefix: "language-"},
		},
		Markdown: HTMLMarkdownPresentation{Sanitize: "ugc"},
		Document: HTMLDocumentPresentation{
			Lang:      "zh-CN",
			InlineCSS: builtinDocumentInlineCSS(),
		},
	}
}

func builtinDocumentInlineCSS() string {
	return `.llm-render-doc{font-family:system-ui,sans-serif;line-height:1.5;max-width:52rem;margin:1rem auto;padding:0 1rem;}
.llm-thinking{opacity:0.85;border-left:3px solid #888;padding-left:0.75rem;margin:1rem 0;}
.llm-thinking-pre{white-space:pre-wrap;margin:0;font-size:0.9rem;}
.llm-code{overflow:auto;padding:0.75rem;background:#f6f8fa;border-radius:6px;}`
}

func NormalizeHTMLPresentation(c HTMLPresentationConfig) HTMLPresentationConfig {
	def := BuiltinHTMLPresentationDefaults()
	if c.Thinking.Outer.Tag == "" {
		c.Thinking.Outer.Tag = def.Thinking.Outer.Tag
	}
	c.Thinking.Outer.Tag = sanitizeHTMLTagName(c.Thinking.Outer.Tag, allowedThinkingWrapperTags, def.Thinking.Outer.Tag)
	if c.Thinking.Outer.Class == "" {
		c.Thinking.Outer.Class = def.Thinking.Outer.Class
	}
	if c.Thinking.Inner.Tag == "" {
		c.Thinking.Inner.Tag = def.Thinking.Inner.Tag
	}
	c.Thinking.Inner.Tag = sanitizeHTMLTagName(c.Thinking.Inner.Tag, allowedThinkingContentTags, def.Thinking.Inner.Tag)
	if c.Thinking.Inner.Class == "" {
		c.Thinking.Inner.Class = def.Thinking.Inner.Class
	}
	if c.Code.Outer.Tag == "" {
		c.Code.Outer.Tag = def.Code.Outer.Tag
	}
	c.Code.Outer.Tag = sanitizeHTMLTagName(c.Code.Outer.Tag, allowedCodeOuterTags, def.Code.Outer.Tag)
	if c.Code.Outer.Class == "" {
		c.Code.Outer.Class = def.Code.Outer.Class
	}
	if c.Code.Inner.Tag == "" {
		c.Code.Inner.Tag = def.Code.Inner.Tag
	}
	c.Code.Inner.Tag = sanitizeHTMLTagName(c.Code.Inner.Tag, allowedCodeInnerTags, def.Code.Inner.Tag)
	if c.Code.Inner.LanguageClassPrefix == "" {
		c.Code.Inner.LanguageClassPrefix = def.Code.Inner.LanguageClassPrefix
	}
	pol := strings.ToLower(strings.TrimSpace(c.Markdown.normalizedSanitize()))
	if pol == "" {
		pol = strings.ToLower(def.Markdown.normalizedSanitize())
	}
	c.Markdown.Sanitize = pol
	c.Markdown.SanitizePolicy = ""
	if c.Document.Lang == "" {
		c.Document.Lang = def.Document.Lang
	}
	if strings.TrimSpace(c.Document.InlineCSS) == "" {
		c.Document.InlineCSS = def.Document.InlineCSS
	}
	return c
}

var (
	allowedThinkingWrapperTags = map[string]struct{}{
		"aside": {}, "div": {}, "section": {}, "figure": {}, "article": {},
	}
	allowedThinkingContentTags = map[string]struct{}{
		"pre": {}, "div": {},
	}
	allowedCodeOuterTags  = map[string]struct{}{"pre": {}}
	allowedCodeInnerTags  = map[string]struct{}{"code": {}}
)

func sanitizeHTMLTagName(tag string, allow map[string]struct{}, fallback string) string {
	t := strings.ToLower(strings.TrimSpace(tag))
	if _, ok := allow[t]; ok {
		return t
	}
	return fallback
}

func mergeHTMLPresentation(base, o HTMLPresentationConfig) HTMLPresentationConfig {
	if strings.TrimSpace(o.Thinking.Outer.Tag) != "" {
		base.Thinking.Outer.Tag = o.Thinking.Outer.Tag
	}
	if strings.TrimSpace(o.Thinking.Outer.Class) != "" {
		base.Thinking.Outer.Class = o.Thinking.Outer.Class
	}
	if strings.TrimSpace(o.Thinking.Inner.Tag) != "" {
		base.Thinking.Inner.Tag = o.Thinking.Inner.Tag
	}
	if strings.TrimSpace(o.Thinking.Inner.Class) != "" {
		base.Thinking.Inner.Class = o.Thinking.Inner.Class
	}
	if strings.TrimSpace(o.Code.Outer.Tag) != "" {
		base.Code.Outer.Tag = o.Code.Outer.Tag
	}
	if strings.TrimSpace(o.Code.Outer.Class) != "" {
		base.Code.Outer.Class = o.Code.Outer.Class
	}
	if strings.TrimSpace(o.Code.Inner.Tag) != "" {
		base.Code.Inner.Tag = o.Code.Inner.Tag
	}
	if strings.TrimSpace(o.Code.Inner.Class) != "" {
		base.Code.Inner.Class = o.Code.Inner.Class
	}
	if strings.TrimSpace(o.Code.Inner.LanguageClassPrefix) != "" {
		base.Code.Inner.LanguageClassPrefix = o.Code.Inner.LanguageClassPrefix
	}
	if strings.TrimSpace(o.Markdown.normalizedSanitize()) != "" {
		base.Markdown.Sanitize = o.Markdown.normalizedSanitize()
		base.Markdown.SanitizePolicy = ""
	}
	if strings.TrimSpace(o.Document.Lang) != "" {
		base.Document.Lang = o.Document.Lang
	}
	if strings.TrimSpace(o.Document.InlineCSS) != "" {
		base.Document.InlineCSS = o.Document.InlineCSS
	}
	return base
}
