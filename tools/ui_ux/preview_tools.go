package ui_ux

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/google/uuid"
	"github.com/originaleric/digeino/pkg/tempstorage"
)

// WritePreviewManifestRequest 注册/覆盖预览产物清单。
type WritePreviewManifestRequest struct {
	ManifestFilename string   `json:"manifest_filename,omitempty" jsonschema:"description=相对路径，默认 preview/preview-manifest.json"`
	ArtifactID       string   `json:"artifact_id,omitempty" jsonschema:"description=可选；为空则生成 UUID"`
	Kind             string   `json:"kind" jsonschema:"description=static_html 或 react_sandpack"`
	Entry            string   `json:"entry" jsonschema:"description=主 HTML 或入口文件相对路径，如 preview/index.html"`
	Assets           []string `json:"assets,omitempty" jsonschema:"description=静态资源相对路径列表"`
	EditableModel    string   `json:"editable_model,omitempty" jsonschema:"description=可选，content.json 等可编辑数据文件"`
	InitialRevision  int      `json:"initial_revision,omitempty" jsonschema:"description=初始 revision，默认 1"`
}

// WritePreviewManifestResponse 写 manifest 结果。
type WritePreviewManifestResponse struct {
	Path         string `json:"path"`
	ArtifactID   string `json:"artifact_id"`
	Revision     int    `json:"revision"`
	RelativePath string `json:"relative_path"`
	Message      string `json:"message"`
}

// NewWritePreviewManifestTool 创建 write_preview_manifest 工具。
func NewWritePreviewManifestTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("write_preview_manifest",
		"写入或覆盖 preview-manifest.json，描述可预览产物（entry、assets、editable_model）。需在 context 注入 workspace_path 或 agent_session_id。应在 write_review_file 写好页面文件后调用。",
		func(ctx context.Context, req *WritePreviewManifestRequest) (*WritePreviewManifestResponse, error) {
			mfname := strings.TrimSpace(req.ManifestFilename)
			if mfname == "" {
				mfname = "preview/preview-manifest.json"
			}
			if err := tempstorage.ValidateRelativePath(mfname); err != nil {
				return nil, err
			}
			kind := strings.TrimSpace(req.Kind)
			if kind == "" {
				return nil, fmt.Errorf("kind is required")
			}
			if !isAllowedPreviewKind(kind) {
				return nil, fmt.Errorf("kind must be one of: static_html, react_sandpack")
			}
			entry := strings.TrimSpace(req.Entry)
			if entry == "" {
				return nil, fmt.Errorf("entry is required")
			}
			if err := tempstorage.ValidateRelativePath(entry); err != nil {
				return nil, fmt.Errorf("entry: %w", err)
			}
			artifactID := strings.TrimSpace(req.ArtifactID)
			if artifactID == "" {
				artifactID = uuid.NewString()
			}
			rev := req.InitialRevision
			if rev <= 0 {
				rev = 1
			}
			var editable *string
			if em := strings.TrimSpace(req.EditableModel); em != "" {
				if err := tempstorage.ValidateRelativePath(em); err != nil {
					return nil, fmt.Errorf("editable_model: %w", err)
				}
				editable = &em
			}
			var assets []string
			for _, a := range req.Assets {
				a = strings.TrimSpace(a)
				if a == "" {
					continue
				}
				if err := tempstorage.ValidateRelativePath(a); err != nil {
					return nil, fmt.Errorf("asset %q: %w", a, err)
				}
				assets = append(assets, a)
			}
			m := &PreviewManifest{
				ArtifactID:    artifactID,
				Kind:          kind,
				Revision:      rev,
				Entry:         entry,
				Assets:        assets,
				EditableModel: editable,
			}
			baseDir, err := tempstorage.GetBasePath(ctx)
			if err != nil {
				return nil, err
			}
			manifestAbs, err := resolveUnderBase(baseDir, mfname)
			if err != nil {
				return nil, err
			}
			if err := writeManifest(manifestAbs, m); err != nil {
				return nil, err
			}
			return &WritePreviewManifestResponse{
				Path:         manifestAbs,
				ArtifactID:   artifactID,
				Revision:     rev,
				RelativePath: mfname,
				Message:      "preview manifest 已写入，可据此做 iframe 预览；用户修改请用 apply_preview_patch（需 base_revision 与当前 revision 一致）",
			}, nil
		})
}

// ApplyPreviewPatchRequest 对预览产物应用结构化补丁。
type ApplyPreviewPatchRequest struct {
	ManifestPath  string         `json:"manifest_path,omitempty" jsonschema:"description=manifest 相对路径，如 preview/preview-manifest.json；与 artifact_id 二选一，manifest_path 优先"`
	ArtifactID    string         `json:"artifact_id,omitempty" jsonschema:"description=可选：仅传 artifact_id 时自动定位 manifest"`
	BaseRevision  int            `json:"base_revision" jsonschema:"description=必须与当前 manifest.revision 一致"`
	Patches       []PreviewPatch `json:"patches" jsonschema:"description=补丁列表，type 支持 html_text、html_attr、html_inner、json_pointer、literal_replace"`
}

// ApplyPreviewPatchResponse 应用补丁结果。
type ApplyPreviewPatchResponse struct {
	ApplyPreviewPatchesResult
}

// NewApplyPreviewPatchTool 创建 apply_preview_patch 工具。
func NewApplyPreviewPatchTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("apply_preview_patch",
		"对预览产物应用结构化补丁（html_text / html_attr / html_inner / json_pointer / literal_replace）。manifest_path 与 artifact_id 二选一；base_revision 须与当前 manifest 一致，成功后 revision+1。需 context 注入 workspace_path 或 agent_session_id。",
		func(ctx context.Context, req *ApplyPreviewPatchRequest) (*ApplyPreviewPatchResponse, error) {
			if req.ManifestPath == "" && req.ArtifactID == "" {
				return nil, fmt.Errorf("manifest_path or artifact_id is required")
			}
			if len(req.Patches) == 0 {
				return nil, fmt.Errorf("patches is required")
			}
			var manifestRel string
			if req.ManifestPath != "" {
				manifestRel = req.ManifestPath
			} else {
				_, mr, _, _, err := resolveManifestInput(ctx, "", req.ArtifactID)
				if err != nil {
					return nil, err
				}
				manifestRel = mr
			}
			res, err := ApplyPreviewPatches(ctx, manifestRel, req.BaseRevision, req.Patches)
			if err != nil {
				return nil, err
			}
			return &ApplyPreviewPatchResponse{ApplyPreviewPatchesResult: *res}, nil
		})
}

// ExportPreviewBundleRequest 导出 zip。
type ExportPreviewBundleRequest struct {
	ManifestPath string `json:"manifest_path,omitempty" jsonschema:"description=manifest 相对路径；与 artifact_id 二选一，manifest_path 优先"`
	ArtifactID   string `json:"artifact_id,omitempty" jsonschema:"description=可选：仅传 artifact_id 时自动定位 manifest"`
	ZipFilename  string `json:"zip_filename,omitempty" jsonschema:"description=输出 zip 相对路径，默认 preview/bundle.zip"`
}

// ExportPreviewBundleResponse zip 写入结果。
type ExportPreviewBundleResponse struct {
	Path         string `json:"path"`
	BytesWritten int64  `json:"bytes_written"`
	Message      string `json:"message"`
}

// NewExportPreviewBundleTool 创建 export_preview_bundle 工具。
func NewExportPreviewBundleTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("export_preview_bundle",
		"将 manifest 中的 entry、assets、editable_model 及 manifest 本身打包为 zip，写入工作区/会话目录，返回 zip 绝对路径。manifest_path 与 artifact_id 二选一。需 context 注入 workspace_path 或 agent_session_id。",
		func(ctx context.Context, req *ExportPreviewBundleRequest) (*ExportPreviewBundleResponse, error) {
			if req.ManifestPath == "" && req.ArtifactID == "" {
				return nil, fmt.Errorf("manifest_path or artifact_id is required")
			}
			manifestRel := req.ManifestPath
			if manifestRel == "" {
				_, mr, _, _, err := resolveManifestInput(ctx, "", req.ArtifactID)
				if err != nil {
					return nil, err
				}
				manifestRel = mr
			}
			zname := strings.TrimSpace(req.ZipFilename)
			if zname == "" {
				zname = "preview/bundle.zip"
			}
			if !strings.HasSuffix(strings.ToLower(zname), ".zip") {
				zname += ".zip"
			}
			if err := tempstorage.ValidateRelativePath(zname); err != nil {
				return nil, err
			}
			path, n, err := WritePreviewZIPToFile(ctx, manifestRel, zname)
			if err != nil {
				return nil, err
			}
			return &ExportPreviewBundleResponse{
				Path:         path,
				BytesWritten: n,
				Message:      "预览产物已导出为 zip",
			}, nil
		})
}

func isAllowedPreviewKind(kind string) bool {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "static_html", "react_sandpack":
		return true
	default:
		return false
	}
}
