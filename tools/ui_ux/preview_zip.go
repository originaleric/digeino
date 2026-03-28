package ui_ux

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/originaleric/digeino/pkg/tempstorage"
)

// BuildPreviewZIP 将 manifest 描述的 entry、assets、editable_model 及 manifest 本身打入 zip。
func BuildPreviewZIP(ctx context.Context, manifestRel string) ([]byte, error) {
	if manifestRel == "" {
		return nil, fmt.Errorf("manifest_path is required")
	}
	baseDir, err := tempstorage.GetBasePath(ctx)
	if err != nil {
		return nil, err
	}
	manifestAbs, err := resolveUnderBase(baseDir, manifestRel)
	if err != nil {
		return nil, err
	}
	m, err := readManifestBytes(manifestAbs)
	if err != nil {
		return nil, err
	}

	files := []string{manifestRel}
	files = append(files, m.Entry)
	files = append(files, m.Assets...)
	if m.EditableModel != nil && *m.EditableModel != "" {
		files = append(files, *m.EditableModel)
	}

	seen := map[string]struct{}{}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, rel := range files {
		r := normRel(rel)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		abs, err := resolveUnderBase(baseDir, r)
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("%s: %w", r, err)
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("%s: %w", r, err)
		}
		if len(data) > maxPatchFileN()*5 {
			_ = zw.Close()
			return nil, fmt.Errorf("%s: file too large for zip export", r)
		}
		w, err := zw.Create(filepath.ToSlash(r))
		if err != nil {
			_ = zw.Close()
			return nil, err
		}
		if _, err := w.Write(data); err != nil {
			_ = zw.Close()
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WritePreviewZIPToFile 写入 zip 到 ctx 对应工作区/会话下的相对路径。
func WritePreviewZIPToFile(ctx context.Context, manifestRel, zipRelFilename string) (path string, size int64, err error) {
	data, err := BuildPreviewZIP(ctx, manifestRel)
	if err != nil {
		return "", 0, err
	}
	path, err = tempstorage.SaveBytesForReview(ctx, zipRelFilename, data)
	if err != nil {
		return "", 0, err
	}
	return path, int64(len(data)), nil
}
