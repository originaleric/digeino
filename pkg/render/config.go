package render

import (
	"bytes"
	_ "embed"
	"io"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed config/config.yaml
var embeddedDefaultYAML []byte

type thinkingYAML struct {
	Open  string `yaml:"open"`
	Close string `yaml:"close"`
}

type configFileYAML struct {
	SchemaVersion int `yaml:"schema_version"`
	// ThinkingTagPairs nil = 键省略；非 nil 空切片 = 显式 thinking_tag_pairs: []。
	ThinkingTagPairs *[]thinkingYAML `yaml:"thinking_tag_pairs"`
	Parse            *parseFileYAML  `yaml:"parse"`
	ParseRender      *ParseRenderLinks     `yaml:"parse_render"`
	Render           *capturedYAMLNode     `yaml:"render"`
	HTML             *htmlPresentationYAML `yaml:"html"` // deprecated: use render:
}

type codeFenceYAML struct {
	Open    string `yaml:"open"`
	Close   string `yaml:"close"`
	Opening string `yaml:"opening"` // deprecated: use open; still read for compatibility
}

func mergeCodeFenceYAML(dst *CodeFenceConfig, src *codeFenceYAML) {
	if dst == nil || src == nil {
		return
	}
	op := strings.TrimSpace(src.Open)
	if op == "" {
		op = strings.TrimSpace(src.Opening)
	}
	if op == "" {
		return
	}
	cl := strings.TrimSpace(src.Close)
	dst.Open = op
	if cl != "" {
		dst.Close = cl
	} else {
		dst.Close = op
	}
}

// htmlPresentationYAML 解码 html: 段，支持 outer/inner 与旧键名（见 MigrateHTMLPresentationRaw）。
type htmlPresentationYAML struct {
	inner HTMLPresentationConfig
}

func (h *htmlPresentationYAML) UnmarshalYAML(n *yaml.Node) error {
	var raw HTMLPresentationConfigRaw
	if err := n.Decode(&raw); err != nil {
		return err
	}
	h.inner = MigrateHTMLPresentationRaw(raw)
	return nil
}

// RenderConfig is parse Options plus presentation (HTML output) from one YAML file.
type RenderConfig struct {
	Options
	// ParseRender is the effective parse→render profile map after loading (defaults if omitted).
	ParseRender ParseRenderLinks
	// Render is the normalized presentation config (formerly YAML key html:).
	Render HTMLPresentationConfig
}

func renderConfigFromYAMLBytes(data []byte) (RenderConfig, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		def := NormalizeHTMLPresentation(HTMLPresentationConfig{})
		links, _ := normalizeParseRenderLinks(nil)
		return RenderConfig{ParseRender: links, Render: def}, nil
	}
	var root configFileYAML
	if err := yaml.Unmarshal(data, &root); err != nil {
		return RenderConfig{}, err
	}
	var thinkList []thinkingYAML
	var thinkingExplicit bool
	if root.Parse != nil && root.Parse.ThinkingTagPairs != nil {
		// 用 make(…,0) 作底，避免 append(nil, 空切片...) 仍为 nil，导致无法区分「显式 []」与「省略」。
		thinkList = append(make([]thinkingYAML, 0), *root.Parse.ThinkingTagPairs...)
		thinkingExplicit = true
	} else if root.ThinkingTagPairs != nil {
		thinkList = append(make([]thinkingYAML, 0), *root.ThinkingTagPairs...)
		thinkingExplicit = true
	}
	if root.Parse != nil {
		thinkList = append(thinkList, root.Parse.ParseExtraPairs...)
	}
	if !thinkingExplicit && len(thinkList) == 0 {
		thinkList = nil
	}
	opt, err := optionsFromThinkingYAMLList(thinkList)
	if err != nil {
		return RenderConfig{}, err
	}
	if root.Parse != nil && root.Parse.CodeFence != nil {
		mergeCodeFenceYAML(&opt.CodeFence, root.Parse.CodeFence)
	}
	links, err := normalizeParseRenderLinks(root.ParseRender)
	if err != nil {
		return RenderConfig{}, err
	}
	if err := validateParseRenderExtras(root.Parse, links); err != nil {
		return RenderConfig{}, err
	}
	pres := BuiltinHTMLPresentationDefaults()
	if root.HTML != nil {
		pres = mergeHTMLPresentation(pres, root.HTML.inner)
	}
	if root.Render != nil && root.Render.Node != nil {
		rpres, err := presentationFromRenderYAML(root.Render.Node, links)
		if err != nil {
			return RenderConfig{}, err
		}
		pres = mergeHTMLPresentation(pres, rpres)
	}
	pres = NormalizeHTMLPresentation(pres)
	return RenderConfig{Options: opt, ParseRender: links, Render: pres}, nil
}

func optionsFromThinkingYAMLList(list []thinkingYAML) (Options, error) {
	if list == nil {
		return Options{}, nil
	}
	pairs := make([]ThinkingTagPair, 0, len(list))
	for _, p := range list {
		pairs = append(pairs, ThinkingTagPair{
			Open:  strings.TrimSpace(p.Open),
			Close: strings.TrimSpace(p.Close),
		})
	}
	opt := Options{ThinkingTagPairs: pairs}.withDefaults()
	if err := validateOptions(opt); err != nil {
		return Options{}, err
	}
	return opt, nil
}

// RenderConfigFromYAML loads full render config (parse + render presentation) from YAML.
func RenderConfigFromYAML(data []byte) (RenderConfig, error) {
	return renderConfigFromYAMLBytes(data)
}

// LoadRenderConfigFromFile reads a YAML file into RenderConfig.
func LoadRenderConfigFromFile(path string) (RenderConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return RenderConfig{}, err
	}
	return renderConfigFromYAMLBytes(b)
}

// LoadRenderConfigFromReader reads YAML from r into RenderConfig.
func LoadRenderConfigFromReader(r io.Reader) (RenderConfig, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return RenderConfig{}, err
	}
	return renderConfigFromYAMLBytes(b)
}

// LoadEmbeddedRenderConfig returns RenderConfig from embedded config/config.yaml.
func LoadEmbeddedRenderConfig() (RenderConfig, error) {
	return renderConfigFromYAMLBytes(embeddedDefaultYAML)
}

// OptionsFromYAML parses YAML into Options (ignores render: / html: sections).
//
//   - 空内容或仅空白：返回零值 Options（Parse 时 withDefaults 仍会补全默认思考标签与 ``` 围栏）。
//   - 省略 thinking_tag_pairs 或值为 null：返回零值 Options，同上。
//   - thinking_tag_pairs: [] 空列表：显式关闭思考区识别（仅 markdown 与代码围栏）。
//   - parse.thinking_tag_pairs：若 YAML 中写出该键（含空列表），则覆盖根级 thinking_tag_pairs。
//   - parse 下其它键：值为 {open, close} 时自动并入思考区识别；parse_render 下可写同名键: thinking（可省略，默认按 thinking 块处理）。
//   - parse.code_fence：与 thinking 相同使用 open / close（闭合行前缀）；省略 close 时与 open 相同。旧键 opening 仍可读作 open。
func OptionsFromYAML(data []byte) (Options, error) {
	rc, err := renderConfigFromYAMLBytes(data)
	if err != nil {
		return Options{}, err
	}
	return rc.Options, nil
}

// LoadOptionsFromFile reads a YAML file (e.g. config.yml) and returns Options.
// Semantics match OptionsFromYAML.
func LoadOptionsFromFile(path string) (Options, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Options{}, err
	}
	return OptionsFromYAML(b)
}

// LoadOptionsFromReader reads YAML from r until EOF.
func LoadOptionsFromReader(r io.Reader) (Options, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Options{}, err
	}
	return OptionsFromYAML(b)
}

// LoadEmbeddedDefaultOptions returns Options from the embedded config/config.yaml.
func LoadEmbeddedDefaultOptions() (Options, error) {
	return OptionsFromYAML(embeddedDefaultYAML)
}

var (
	embeddedPairsMu     sync.Mutex
	embeddedPairsCached []ThinkingTagPair
	embeddedPairsErr    error
	embeddedPairsLoaded bool
)

func loadEmbeddedThinkingPairs() ([]ThinkingTagPair, error) {
	embeddedPairsMu.Lock()
	defer embeddedPairsMu.Unlock()
	if embeddedPairsLoaded {
		return embeddedPairsCached, embeddedPairsErr
	}
	embeddedPairsLoaded = true
	opt, err := OptionsFromYAML(embeddedDefaultYAML)
	if err != nil {
		embeddedPairsErr = err
		return nil, err
	}
	if len(opt.ThinkingTagPairs) == 0 {
		embeddedPairsErr = &OptionError{Msg: "embedded config/config.yaml: thinking_tag_pairs is empty"}
		return nil, embeddedPairsErr
	}
	embeddedPairsCached = make([]ThinkingTagPair, len(opt.ThinkingTagPairs))
	copy(embeddedPairsCached, opt.ThinkingTagPairs)
	return embeddedPairsCached, nil
}

func defaultThinkingTagPairsFromEmbed() []ThinkingTagPair {
	pairs, err := loadEmbeddedThinkingPairs()
	if err != nil || len(pairs) == 0 {
		return []ThinkingTagPair{
			{Open: "<think>", Close: "</think>"},
			{Open: "<redacted_reasoning>", Close: "</redacted_reasoning>"},
			{Open: "<reasoning>", Close: "</reasoning>"},
		}
	}
	out := make([]ThinkingTagPair, len(pairs))
	copy(out, pairs)
	return out
}
