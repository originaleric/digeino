package ui_ux

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/originaleric/digeino/pkg/tempstorage"
)

// PreviewPatch 单条修改（与前端 / Agent 约定一致）。
type PreviewPatch struct {
	Type     string `json:"type"`
	Selector string `json:"selector,omitempty"`
	Text     string `json:"text,omitempty"`
	Attr     string `json:"attr,omitempty"`
	// Value 用于 html_attr（字符串）或 json_pointer（任意 JSON）
	Value   any    `json:"value,omitempty"`
	Pointer string `json:"pointer,omitempty"`
	File    string `json:"file,omitempty"`
	Old     string `json:"old,omitempty"`
	New     string `json:"new,omitempty"`
	HTML    string `json:"html,omitempty"` // html_inner：替换匹配元素 inner HTML
}

// ApplyPreviewPatchesResult apply_preview_patch 返回。
type ApplyPreviewPatchesResult struct {
	OK           bool     `json:"ok"`
	Revision     int      `json:"revision"`
	UpdatedFiles []string `json:"updated_files"`
	ManifestPath string   `json:"manifest_path"`
	Message      string   `json:"message"`
}

// ApplyPreviewPatches 将补丁应用到会话/工作区内的预览产物。
func ApplyPreviewPatches(ctx context.Context, manifestRel string, baseRevision int, patches []PreviewPatch) (*ApplyPreviewPatchesResult, error) {
	baseDir, manifestRel, manifestAbs, m, err := resolveManifestInput(ctx, manifestRel, "")
	if err != nil {
		return nil, err
	}
	if m.Revision != baseRevision {
		return nil, fmt.Errorf("revision mismatch: manifest has %d, expected %d", m.Revision, baseRevision)
	}

	snapshotted := map[string]struct{}{}
	snapshotOnce := func(rel, abs string) error {
		n := normRel(rel)
		if _, ok := snapshotted[n]; ok {
			return nil
		}
		b, err := osReadLimited(abs)
		if err != nil {
			return err
		}
		if err := writeHistorySnapshot(baseDir, m, m.Revision, n, b); err != nil {
			return err
		}
		snapshotted[n] = struct{}{}
		return nil
	}

	var updated []string
	for i, p := range patches {
		switch p.Type {
		case "html_text", "html_attr", "html_inner":
			rel := m.Entry
			abs, err := resolveUnderBase(baseDir, rel)
			if err != nil {
				return nil, fmt.Errorf("patch %d entry: %w", i, err)
			}
			if err := m.assertWritableRel(rel); err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			if err := snapshotOnce(rel, abs); err != nil {
				return nil, fmt.Errorf("patch %d: snapshot failed: %w", i, err)
			}
			htmlStr, err := readTextFileLimited(abs)
			if err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			newHTML, err := applyHTMLPatch(htmlStr, p)
			if err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			if err := writeTextFile(abs, newHTML); err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			updated = append(updated, rel)

		case "json_pointer":
			if m.EditableModel == nil || *m.EditableModel == "" {
				return nil, fmt.Errorf("patch %d: manifest has no editable_model", i)
			}
			rel := *m.EditableModel
			if err := m.assertWritableRel(rel); err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			abs, err := resolveUnderBase(baseDir, rel)
			if err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			if err := snapshotOnce(rel, abs); err != nil {
				return nil, fmt.Errorf("patch %d: snapshot failed: %w", i, err)
			}
			raw, err := osReadLimited(abs)
			if err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			var root any
			if err := json.Unmarshal(raw, &root); err != nil {
				return nil, fmt.Errorf("patch %d: invalid json: %w", i, err)
			}
			newRoot, err := setJSONPointer(root, p.Pointer, p.Value)
			if err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			out, err := json.MarshalIndent(newRoot, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			if len(out) > maxPatchFileN() {
				return nil, fmt.Errorf("patch %d: result exceeds max size", i)
			}
			if err := os.WriteFile(abs, out, 0644); err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			updated = append(updated, rel)

		case "literal_replace":
			if p.File == "" || p.Old == "" {
				return nil, fmt.Errorf("patch %d: literal_replace requires file and old", i)
			}
			rel := filepath.ToSlash(filepath.Clean(p.File))
			// 统一为平台路径再与 manifest 比较
			rel = strings.TrimPrefix(rel, "./")
			if err := tempstorage.ValidateRelativePath(rel); err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			if err := m.assertWritableRel(rel); err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			if !allowedLiteralExt(rel) {
				return nil, fmt.Errorf("patch %d: file extension not allowed for literal_replace", i)
			}
			abs, err := resolveUnderBase(baseDir, rel)
			if err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			if err := snapshotOnce(rel, abs); err != nil {
				return nil, fmt.Errorf("patch %d: snapshot failed: %w", i, err)
			}
			content, err := readTextFileLimited(abs)
			if err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			n := strings.Count(content, p.Old)
			if n != 1 {
				return nil, fmt.Errorf("patch %d: old string must match exactly once, got %d", i, n)
			}
			newContent := strings.Replace(content, p.Old, p.New, 1)
			if err := writeTextFile(abs, newContent); err != nil {
				return nil, fmt.Errorf("patch %d: %w", i, err)
			}
			updated = append(updated, rel)

		default:
			return nil, fmt.Errorf("patch %d: unknown type %q", i, p.Type)
		}
	}

	manifestOld, err := osReadLimited(manifestAbs)
	if err == nil {
		_ = writeHistorySnapshot(baseDir, m, m.Revision, manifestRel, manifestOld)
	}
	m.Revision++
	if err := writeManifest(manifestAbs, m); err != nil {
		return nil, err
	}

	return &ApplyPreviewPatchesResult{
		OK:           true,
		Revision:     m.Revision,
		UpdatedFiles: updated,
		ManifestPath: manifestAbs,
		Message:      fmt.Sprintf("已应用 %d 条补丁，revision=%d", len(patches), m.Revision),
	}, nil
}

func osReadLimited(abs string) ([]byte, error) {
	b, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	if len(b) > maxPatchFileN() {
		return nil, fmt.Errorf("file exceeds max patch size")
	}
	return b, nil
}

func applyHTMLPatch(htmlStr string, p PreviewPatch) (string, error) {
	// 片段 HTML 需包一层根节点，避免 goquery 输出完整 <html> 文档骨架
	wrapped := "<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body><div id=\"__uiux_root\">" +
		htmlStr + "</div></body></html>"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrapped))
	if err != nil {
		return "", err
	}
	root := doc.Find("#__uiux_root")
	if root.Length() != 1 {
		return "", fmt.Errorf("internal: uiux root wrapper missing")
	}
	if p.Selector == "" {
		return "", fmt.Errorf("html patch requires selector")
	}
	sel := root.Find(p.Selector)
	if sel.Length() == 0 {
		return "", fmt.Errorf("selector %q matches no elements", p.Selector)
	}
	switch p.Type {
	case "html_text":
		sel.Each(func(i int, s *goquery.Selection) {
			s.Empty()
			s.AppendHtml(html.EscapeString(p.Text))
		})
	case "html_attr":
		if p.Attr == "" {
			return "", fmt.Errorf("html_attr requires attr")
		}
		val, err := patchAttrValue(p.Value, p.Text)
		if err != nil {
			return "", err
		}
		sel.SetAttr(p.Attr, val)
	case "html_inner":
		sel.Each(func(i int, s *goquery.Selection) {
			s.SetHtml(p.HTML)
		})
	default:
		return "", fmt.Errorf("internal: unknown html patch type %q", p.Type)
	}
	out, err := root.Html()
	if err != nil {
		return "", err
	}
	return out, nil
}

func patchAttrValue(value any, textFallback string) (string, error) {
	if value == nil {
		if textFallback != "" {
			return textFallback, nil
		}
		return "", fmt.Errorf("html_attr requires value or text")
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		b, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		var s string
		if err := json.Unmarshal(b, &s); err == nil {
			return s, nil
		}
		return strings.Trim(string(b), `"`), nil
	}
}

func allowedLiteralExt(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	_, ok := previewAllowedExtensions()[ext]
	return ok
}

func setJSONPointer(root any, pointer string, value any) (any, error) {
	if pointer == "" || pointer == "/" {
		return value, nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("pointer must start with /")
	}
	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	// RFC6901 unescape ~1 -> /, ~0 -> ~
	for i := range parts {
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(parts[i], "~1", "/"), "~0", "~")
	}

	if root == nil {
		root = map[string]any{}
	}
	return setPath(root, parts, value)
}

func setPath(cur any, parts []string, value any) (any, error) {
	if len(parts) == 0 {
		return value, nil
	}
	if cur == nil {
		if _, err := strconv.Atoi(parts[0]); err == nil {
			cur = []any{}
		} else {
			cur = map[string]any{}
		}
	}
	key := parts[0]
	rest := parts[1:]

	if len(rest) == 0 {
		switch parent := cur.(type) {
		case map[string]any:
			parent[key] = normalizeJSONValue(value)
			return parent, nil
		default:
			m := map[string]any{}
			m[key] = normalizeJSONValue(value)
			return m, nil
		}
	}

	switch parent := cur.(type) {
	case map[string]any:
		child, ok := parent[key]
		if !ok || child == nil {
			next := ""
			if len(rest) > 0 {
				next = rest[0]
			}
			if _, err := strconv.Atoi(next); err == nil {
				idx, _ := strconv.Atoi(next)
				child = make([]any, idx+1)
			} else {
				child = map[string]any{}
			}
			parent[key] = child
		}
		newChild, err := setPath(child, rest, value)
		if err != nil {
			return nil, err
		}
		parent[key] = newChild
		return parent, nil

	case []any:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return nil, fmt.Errorf("invalid array index %q", key)
		}
		for len(parent) <= idx {
			parent = append(parent, nil)
		}
		newChild, err := setPath(parent[idx], rest, value)
		if err != nil {
			return nil, err
		}
		parent[idx] = newChild
		return parent, nil

	default:
		m := map[string]any{}
		return setPath(m, parts, value)
	}
}

func normalizeJSONValue(value any) any {
	if value == nil {
		return nil
	}
	// 保持 JSON 解码后的数字/布尔/字符串/map/slice
	return value
}
