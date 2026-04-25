// Package rendertemplate renders business view models into safe HTML fragments.
package rendertemplate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// TemplateRenderOptions controls template parsing, execution, and caching.
type TemplateRenderOptions struct {
	// Funcs registers safe template helper functions.
	Funcs map[string]any
	// Cache enables compiled template caching. File templates are keyed by path
	// and file metadata; inline templates require CacheKey to opt in.
	Cache bool
	// CacheKey explicitly names an inline or virtual template cache entry.
	CacheKey string
	// BaseDir constrains relative file template paths to a trusted root.
	BaseDir string
	// Strict makes missing map keys or struct fields return execution errors.
	Strict bool
}

// TemplateErrorKind classifies template rendering failures for callers.
type TemplateErrorKind string

const (
	TemplateErrorNotFound TemplateErrorKind = "not_found"
	TemplateErrorPath     TemplateErrorKind = "path"
	TemplateErrorParse    TemplateErrorKind = "parse"
	TemplateErrorExecute  TemplateErrorKind = "execute"
	TemplateErrorFunc     TemplateErrorKind = "func"
)

// TemplateError wraps a rendering failure with a stable kind and context.
type TemplateError struct {
	Kind         TemplateErrorKind
	TemplateName string
	Path         string
	Err          error
}

func (e *TemplateError) Error() string {
	if e == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("render template")
	if e.Kind != "" {
		b.WriteString(" ")
		b.WriteString(string(e.Kind))
	}
	if e.TemplateName != "" {
		b.WriteString(" ")
		b.WriteString(e.TemplateName)
	}
	if e.Path != "" {
		b.WriteString(" ")
		b.WriteString(e.Path)
	}
	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

func (e *TemplateError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type cachedTemplate struct {
	t *htmltemplate.Template
}

var templateCache sync.Map

var templateFuncNameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// RenderTemplate renders templateText with data using html/template escaping.
func RenderTemplate(templateText string, data any, opts TemplateRenderOptions) (string, error) {
	funcKey, err := validateFuncs(opts.Funcs)
	if err != nil {
		return "", newTemplateError(TemplateErrorFunc, inlineTemplateName(opts), "", err)
	}
	name := inlineTemplateName(opts)
	key := ""
	if opts.Cache && strings.TrimSpace(opts.CacheKey) != "" {
		key = cacheKey("inline", opts.CacheKey, hashString(templateText), strictKey(opts.Strict), funcKey)
	}
	t, err := templateFromCacheOrParse(key, name, templateText, opts)
	if err != nil {
		return "", err
	}
	return executeTemplate(t, name, "", data)
}

// RenderTemplateFromFile renders a template file with data using html/template escaping.
func RenderTemplateFromFile(path string, data any, opts TemplateRenderOptions) (string, error) {
	funcKey, err := validateFuncs(opts.Funcs)
	if err != nil {
		return "", newTemplateError(TemplateErrorFunc, filepath.Base(path), path, err)
	}
	absPath, err := resolveTemplatePath(path, opts.BaseDir)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		kind := TemplateErrorPath
		if errors.Is(err, os.ErrNotExist) {
			kind = TemplateErrorNotFound
		}
		return "", newTemplateError(kind, filepath.Base(absPath), absPath, err)
	}
	if info.IsDir() {
		return "", newTemplateError(TemplateErrorPath, filepath.Base(absPath), absPath, fmt.Errorf("template path is a directory"))
	}
	b, err := os.ReadFile(absPath)
	if err != nil {
		kind := TemplateErrorPath
		if errors.Is(err, os.ErrNotExist) {
			kind = TemplateErrorNotFound
		}
		return "", newTemplateError(kind, filepath.Base(absPath), absPath, err)
	}
	name := filepath.Base(absPath)
	key := ""
	if opts.Cache {
		key = cacheKey("file", absPath, fmt.Sprintf("%d", info.ModTime().UnixNano()), fmt.Sprintf("%d", info.Size()), strictKey(opts.Strict), funcKey)
	}
	t, err := templateFromCacheOrParse(key, name, string(b), opts)
	if err != nil {
		if te, ok := err.(*TemplateError); ok {
			te.Path = absPath
			return "", te
		}
		return "", err
	}
	return executeTemplate(t, name, absPath, data)
}

func templateFromCacheOrParse(key, name, text string, opts TemplateRenderOptions) (*htmltemplate.Template, error) {
	if key != "" {
		if v, ok := templateCache.Load(key); ok {
			return v.(cachedTemplate).t, nil
		}
	}
	t, err := parseTemplate(name, text, opts)
	if err != nil {
		return nil, err
	}
	if key != "" {
		v, _ := templateCache.LoadOrStore(key, cachedTemplate{t: t})
		return v.(cachedTemplate).t, nil
	}
	return t, nil
}

func parseTemplate(name, text string, opts TemplateRenderOptions) (*htmltemplate.Template, error) {
	t := htmltemplate.New(name)
	if opts.Strict {
		t = t.Option("missingkey=error")
	}
	if len(opts.Funcs) > 0 {
		var err error
		t, err = withFuncs(t, opts.Funcs)
		if err != nil {
			return nil, newTemplateError(TemplateErrorFunc, name, "", err)
		}
	}
	parsed, err := t.Parse(text)
	if err != nil {
		return nil, newTemplateError(TemplateErrorParse, name, "", err)
	}
	return parsed, nil
}

func executeTemplate(t *htmltemplate.Template, name, path string, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", newTemplateError(TemplateErrorExecute, name, path, err)
	}
	return buf.String(), nil
}

func resolveTemplatePath(path, baseDir string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", newTemplateError(TemplateErrorPath, "", "", fmt.Errorf("template path is empty"))
	}
	if baseDir == "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", newTemplateError(TemplateErrorPath, filepath.Base(path), path, err)
		}
		return filepath.Clean(abs), nil
	}
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", newTemplateError(TemplateErrorPath, filepath.Base(path), path, err)
	}
	baseAbs = filepath.Clean(baseAbs)
	if resolvedBase, err := filepath.EvalSymlinks(baseAbs); err == nil {
		baseAbs = filepath.Clean(resolvedBase)
	}
	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(baseAbs, candidate)
	}
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", newTemplateError(TemplateErrorPath, filepath.Base(path), path, err)
	}
	abs = filepath.Clean(abs)
	if resolvedAbs, err := filepath.EvalSymlinks(abs); err == nil {
		abs = filepath.Clean(resolvedAbs)
	}
	rel, err := filepath.Rel(baseAbs, abs)
	if err != nil {
		return "", newTemplateError(TemplateErrorPath, filepath.Base(path), abs, err)
	}
	if rel == "." || rel == "" {
		return "", newTemplateError(TemplateErrorPath, filepath.Base(path), abs, fmt.Errorf("template path resolves to base directory"))
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", newTemplateError(TemplateErrorPath, filepath.Base(path), abs, fmt.Errorf("template path escapes base dir %q", baseAbs))
	}
	return abs, nil
}

func validateFuncs(funcs map[string]any) (string, error) {
	if len(funcs) == 0 {
		return "funcs:none", nil
	}
	names := make([]string, 0, len(funcs))
	for name, fn := range funcs {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return "", fmt.Errorf("function name must be non-empty")
		}
		if name != trimmed {
			return "", fmt.Errorf("function name %q must not contain surrounding whitespace", name)
		}
		if !templateFuncNameRE.MatchString(name) {
			return "", fmt.Errorf("invalid function name %q", name)
		}
		if fn == nil {
			return "", fmt.Errorf("function %q must be non-nil", name)
		}
		v := reflect.ValueOf(fn)
		if v.Kind() != reflect.Func {
			return "", fmt.Errorf("function %q must be a func", name)
		}
		names = append(names, fmt.Sprintf("%s:%x", name, v.Pointer()))
	}
	sort.Strings(names)
	return strings.Join(names, ","), nil
}

func withFuncs(t *htmltemplate.Template, funcs map[string]any) (out *htmltemplate.Template, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid template funcs: %v", r)
		}
	}()
	return t.Funcs(htmltemplate.FuncMap(funcs)), nil
}

func inlineTemplateName(opts TemplateRenderOptions) string {
	if strings.TrimSpace(opts.CacheKey) != "" {
		return strings.TrimSpace(opts.CacheKey)
	}
	return "inline"
}

func strictKey(strict bool) string {
	if strict {
		return "strict"
	}
	return "missingkey-default"
}

func cacheKey(parts ...string) string {
	return strings.Join(parts, "\x00")
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func newTemplateError(kind TemplateErrorKind, name, path string, err error) *TemplateError {
	return &TemplateError{Kind: kind, TemplateName: name, Path: path, Err: err}
}
