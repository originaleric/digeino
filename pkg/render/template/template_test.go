package rendertemplate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderTemplateEscapesHTML(t *testing.T) {
	out, err := RenderTemplate(`<p>{{.Name}}</p>`, map[string]any{"Name": `<script>alert(1)</script>`}, TemplateRenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "<script>") || !strings.Contains(out, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped script, got %q", out)
	}
}

func TestRenderTemplateStrictMissingKey(t *testing.T) {
	_, err := RenderTemplate(`{{.Missing}}`, map[string]any{}, TemplateRenderOptions{Strict: true})
	assertTemplateErrorKind(t, err, TemplateErrorExecute)
}

func TestRenderTemplateFuncs(t *testing.T) {
	out, err := RenderTemplate(`{{label .Risk}}`, map[string]any{"Risk": "high"}, TemplateRenderOptions{
		Funcs: map[string]any{
			"label": func(s string) string {
				return "risk-" + s
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out != "risk-high" {
		t.Fatalf("out=%q", out)
	}
}

func TestRenderTemplateRejectsInvalidFunc(t *testing.T) {
	_, err := RenderTemplate(`{{bad .X}}`, map[string]any{"X": "x"}, TemplateRenderOptions{
		Funcs: map[string]any{"bad": "not-a-func"},
	})
	assertTemplateErrorKind(t, err, TemplateErrorFunc)
}

func TestRenderTemplateFromFileWithBaseDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "card.tmpl"), []byte(`<strong>{{.Title}}</strong>`), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := RenderTemplateFromFile("card.tmpl", map[string]any{"Title": "Hello"}, TemplateRenderOptions{BaseDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if out != "<strong>Hello</strong>" {
		t.Fatalf("out=%q", out)
	}
}

func TestRenderTemplateFromFileRejectsBaseDirEscape(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "card.tmpl")
	if err := os.WriteFile(outside, []byte(`x`), 0644); err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(dir, outside)
	if err != nil {
		t.Fatal(err)
	}
	_, err = RenderTemplateFromFile(rel, nil, TemplateRenderOptions{BaseDir: dir})
	assertTemplateErrorKind(t, err, TemplateErrorPath)
}

func TestRenderTemplateFromFileNotFound(t *testing.T) {
	_, err := RenderTemplateFromFile("missing.tmpl", nil, TemplateRenderOptions{BaseDir: t.TempDir()})
	assertTemplateErrorKind(t, err, TemplateErrorNotFound)
}

func TestRenderTemplateFromFileCacheInvalidatesOnFileChange(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "card.tmpl")
	if err := os.WriteFile(p, []byte(`one {{.Name}}`), 0644); err != nil {
		t.Fatal(err)
	}
	opts := TemplateRenderOptions{BaseDir: dir, Cache: true}
	out, err := RenderTemplateFromFile("card.tmpl", map[string]any{"Name": "A"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if out != "one A" {
		t.Fatalf("out=%q", out)
	}
	time.Sleep(2 * time.Millisecond)
	if err := os.WriteFile(p, []byte(`two {{.Name}}!`), 0644); err != nil {
		t.Fatal(err)
	}
	out, err = RenderTemplateFromFile("card.tmpl", map[string]any{"Name": "B"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if out != "two B!" {
		t.Fatalf("expected cache invalidation, got %q", out)
	}
}

func TestRenderTemplateInlineCacheUsesCacheKey(t *testing.T) {
	opts := TemplateRenderOptions{Cache: true, CacheKey: "inline-card"}
	out, err := RenderTemplate(`{{.Name}}`, map[string]any{"Name": "A"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if out != "A" {
		t.Fatalf("out=%q", out)
	}
	out, err = RenderTemplate(`{{.Name}}`, map[string]any{"Name": "B"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if out != "B" {
		t.Fatalf("out=%q", out)
	}
}

func assertTemplateErrorKind(t *testing.T, err error, kind TemplateErrorKind) {
	t.Helper()
	var te *TemplateError
	if !errors.As(err, &te) {
		t.Fatalf("expected TemplateError %q, got %T: %v", kind, err, err)
	}
	if te.Kind != kind {
		t.Fatalf("kind=%q want %q err=%v", te.Kind, kind, err)
	}
}
