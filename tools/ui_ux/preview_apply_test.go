package ui_ux

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/originaleric/digeino/pkg/tempstorage"
)

func TestApplyPreviewPatches_HTMLAndJSON(t *testing.T) {
	dir := t.TempDir()
	ctx := context.WithValue(context.Background(), tempstorage.ContextKeyWorkspacePath, dir)

	entry := "preview/index.html"
	// 片段 HTML（推荐：由模型输出可预览片段，工具会包一层 #__uiux_root）
	html := `<h1 data-uiux-id="t1">Hello</h1><img data-uiux-id="i1" src="a.png" alt="x">`
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(entry)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, entry), []byte(html), 0644); err != nil {
		t.Fatal(err)
	}

	modelPath := "preview/content.json"
	cj := `{"hero":{"title":"old"}}`
	if err := os.WriteFile(filepath.Join(dir, modelPath), []byte(cj), 0644); err != nil {
		t.Fatal(err)
	}

	mfname := "preview/preview-manifest.json"
	manifest := PreviewManifest{
		ArtifactID:    "test-art",
		Kind:          "static_html",
		Revision:      1,
		Entry:         entry,
		EditableModel: stringPtr(modelPath),
	}
	absM := filepath.Join(dir, mfname)
	if err := os.MkdirAll(filepath.Dir(absM), 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeManifest(absM, &manifest); err != nil {
		t.Fatal(err)
	}

	res, err := ApplyPreviewPatches(ctx, mfname, 1, []PreviewPatch{
		{Type: "html_text", Selector: "[data-uiux-id='t1']", Text: "Hi"},
		{Type: "html_attr", Selector: "[data-uiux-id='i1']", Attr: "src", Value: "b.png"},
		{Type: "json_pointer", Pointer: "/hero/title", Value: "new"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Revision != 2 {
		t.Fatalf("unexpected result: %+v", res)
	}

	outHTML, err := os.ReadFile(filepath.Join(dir, entry))
	if err != nil {
		t.Fatal(err)
	}
	s := string(outHTML)
	if strings.Contains(s, "Hello") || !strings.Contains(s, "Hi") || !strings.Contains(s, `src="b.png"`) {
		t.Fatalf("html patch failed: %s", s)
	}

	outJ, err := os.ReadFile(filepath.Join(dir, modelPath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(outJ), `"new"`) {
		t.Fatalf("json patch failed: %s", outJ)
	}

	m2, err := readManifestBytes(absM)
	if err != nil {
		t.Fatal(err)
	}
	if m2.Revision != 2 {
		t.Fatalf("manifest revision %d", m2.Revision)
	}

	// 默认启用 history：应保留 revision=1 的快照（包含 entry/editable_model/manifest）
	hEntry := filepath.Join(dir, "preview/history/test-art/rev-000001", entry)
	if _, err := os.Stat(hEntry); err != nil {
		t.Fatalf("history entry snapshot missing: %v", err)
	}
	hModel := filepath.Join(dir, "preview/history/test-art/rev-000001", modelPath)
	if _, err := os.Stat(hModel); err != nil {
		t.Fatalf("history model snapshot missing: %v", err)
	}
	hManifest := filepath.Join(dir, "preview/history/test-art/rev-000001", mfname)
	if _, err := os.Stat(hManifest); err != nil {
		t.Fatalf("history manifest snapshot missing: %v", err)
	}
}

func stringPtr(s string) *string { return &s }
