package ui_ux

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/pkg/tempstorage"
)

// PreviewManifest 描述可预览产物（与 docs/ideas 方案一致）。
type PreviewManifest struct {
	ArtifactID    string   `json:"artifact_id"`
	Kind          string   `json:"kind"`
	Revision      int      `json:"revision"`
	Entry         string   `json:"entry"`
	Assets        []string `json:"assets,omitempty"`
	EditableModel *string  `json:"editable_model,omitempty"`
}

func maxPatchFileN() int {
	if c := config.Get(); c != nil && c.UIUX.Preview.MaxPatchFileBytes > 0 {
		return c.UIUX.Preview.MaxPatchFileBytes
	}
	return 5 * 1024 * 1024
}

func resolveUnderBase(baseDir, rel string) (abs string, err error) {
	if err := tempstorage.ValidateRelativePath(rel); err != nil {
		return "", err
	}
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	rel = filepath.Clean(strings.ReplaceAll(rel, "/", string(filepath.Separator)))
	full := filepath.Join(baseAbs, rel)
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	if !isSubpath(baseAbs, fullAbs) {
		return "", fmt.Errorf("path escapes base directory")
	}
	return fullAbs, nil
}

func isSubpath(base, child string) bool {
	base = filepath.Clean(base)
	child = filepath.Clean(child)
	if base == child {
		return true
	}
	sep := string(os.PathSeparator)
	return strings.HasPrefix(child+sep, base+sep)
}

// allowedRelPaths 返回 manifest 声明的可写相对路径集合（含 entry、assets、editable_model）。
func (m *PreviewManifest) allowedRelPaths() []string {
	out := []string{m.Entry}
	out = append(out, m.Assets...)
	if m.EditableModel != nil && *m.EditableModel != "" {
		out = append(out, *m.EditableModel)
	}
	return out
}

func normRel(p string) string {
	return filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "./")))
}

func (m *PreviewManifest) assertWritableRel(rel string) error {
	if rel == "" {
		return fmt.Errorf("empty relative path")
	}
	if err := tempstorage.ValidateRelativePath(rel); err != nil {
		return err
	}
	relNorm := normRel(rel)
	for _, a := range m.allowedRelPaths() {
		if normRel(a) == relNorm {
			return nil
		}
	}
	return fmt.Errorf("file %q is not listed in preview manifest (entry/assets/editable_model)", rel)
}

func readManifestBytes(path string) (*PreviewManifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) > maxPatchFileN() {
		return nil, fmt.Errorf("manifest file exceeds max size")
	}
	var m PreviewManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.Entry == "" {
		return nil, fmt.Errorf("manifest.entry is required")
	}
	if m.Kind == "" {
		return nil, fmt.Errorf("manifest.kind is required")
	}
	return &m, nil
}

func writeManifest(path string, m *PreviewManifest) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func readTextFileLimited(abs string) (string, error) {
	b, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	if len(b) > maxPatchFileN() {
		return "", fmt.Errorf("file exceeds max patch size (%d bytes)", maxPatchFileN())
	}
	return string(b), nil
}

func writeTextFile(abs, content string) error {
	if len(content) > maxPatchFileN() {
		return fmt.Errorf("content exceeds max patch size (%d bytes)", maxPatchFileN())
	}
	return os.WriteFile(abs, []byte(content), 0644)
}
