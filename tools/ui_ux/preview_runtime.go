package ui_ux

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/pkg/tempstorage"
)

var defaultAllowedExtensions = []string{".tsx", ".jsx", ".ts", ".js", ".html", ".htm", ".json", ".css", ".md"}

func previewAllowedExtensions() map[string]struct{} {
	set := map[string]struct{}{}
	exts := defaultAllowedExtensions
	if cfg := config.Get(); cfg != nil && len(cfg.UIUX.Preview.AllowedExtensions) > 0 {
		exts = cfg.UIUX.Preview.AllowedExtensions
	}
	for _, ext := range exts {
		e := strings.TrimSpace(strings.ToLower(ext))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		set[e] = struct{}{}
	}
	return set
}

func previewHistoryEnabled() bool {
	cfg := config.Get()
	if cfg == nil || cfg.UIUX.Preview.HistoryEnabled == nil {
		return true
	}
	return *cfg.UIUX.Preview.HistoryEnabled
}

func previewHistoryDir() string {
	if cfg := config.Get(); cfg != nil && strings.TrimSpace(cfg.UIUX.Preview.HistoryDir) != "" {
		return filepath.ToSlash(filepath.Clean(strings.TrimSpace(cfg.UIUX.Preview.HistoryDir)))
	}
	return "preview/history"
}

func resolveManifestInput(ctx context.Context, manifestPath, artifactID string) (baseDir, manifestRel, manifestAbs string, m *PreviewManifest, err error) {
	baseDir, err = tempstorage.GetBasePath(ctx)
	if err != nil {
		return "", "", "", nil, err
	}

	mp := strings.TrimSpace(manifestPath)
	if mp != "" {
		if err := tempstorage.ValidateRelativePath(mp); err != nil {
			return "", "", "", nil, fmt.Errorf("manifest_path: %w", err)
		}
		manifestRel = normRel(mp)
		manifestAbs, err = resolveUnderBase(baseDir, manifestRel)
		if err != nil {
			return "", "", "", nil, fmt.Errorf("manifest_path: %w", err)
		}
		m, err = readManifestBytes(manifestAbs)
		if err != nil {
			return "", "", "", nil, err
		}
		return baseDir, manifestRel, manifestAbs, m, nil
	}

	aid := strings.TrimSpace(artifactID)
	if aid == "" {
		return "", "", "", nil, fmt.Errorf("manifest_path or artifact_id is required")
	}
	manifestRel, manifestAbs, err = findManifestByArtifactID(baseDir, aid)
	if err != nil {
		return "", "", "", nil, err
	}
	m, err = readManifestBytes(manifestAbs)
	if err != nil {
		return "", "", "", nil, err
	}
	return baseDir, manifestRel, manifestAbs, m, nil
}

func findManifestByArtifactID(baseDir, artifactID string) (rel string, abs string, err error) {
	try := []string{"preview/preview-manifest.json"}
	for _, c := range try {
		a, e := resolveUnderBase(baseDir, c)
		if e != nil {
			continue
		}
		b, e := os.ReadFile(a)
		if e != nil {
			continue
		}
		var v struct {
			ArtifactID string `json:"artifact_id"`
		}
		if json.Unmarshal(b, &v) == nil && strings.TrimSpace(v.ArtifactID) == artifactID {
			return c, a, nil
		}
	}

	const maxScanFiles = 20000
	scanned := 0
	var foundRel, foundAbs string
	err = filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			switch name {
			case ".git", ".specstory", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			return nil
		}
		scanned++
		if scanned > maxScanFiles {
			return filepath.SkipAll
		}
		b, e := os.ReadFile(path)
		if e != nil || len(b) == 0 || len(b) > maxPatchFileN() {
			return nil
		}
		var v struct {
			ArtifactID string `json:"artifact_id"`
		}
		if json.Unmarshal(b, &v) != nil {
			return nil
		}
		if strings.TrimSpace(v.ArtifactID) == artifactID {
			foundAbs = path
			r, e := filepath.Rel(baseDir, path)
			if e != nil {
				return e
			}
			foundRel = filepath.ToSlash(filepath.Clean(r))
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return "", "", err
	}
	if foundRel == "" {
		return "", "", fmt.Errorf("artifact_id %q not found under base directory", artifactID)
	}
	return foundRel, foundAbs, nil
}

func writeHistorySnapshot(baseDir string, m *PreviewManifest, revision int, rel string, data []byte) error {
	if !previewHistoryEnabled() {
		return nil
	}
	if m == nil {
		return fmt.Errorf("manifest is nil")
	}
	if err := tempstorage.ValidateRelativePath(rel); err != nil {
		return err
	}

	aid := strings.TrimSpace(m.ArtifactID)
	if aid == "" {
		aid = "unknown-artifact"
	}
	hdir := previewHistoryDir()
	historyRel := filepath.ToSlash(filepath.Join(hdir, aid, fmt.Sprintf("rev-%06d", revision), rel))
	historyAbs, err := resolveUnderBase(baseDir, historyRel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(historyAbs), 0755); err != nil {
		return err
	}
	return os.WriteFile(historyAbs, data, 0644)
}

