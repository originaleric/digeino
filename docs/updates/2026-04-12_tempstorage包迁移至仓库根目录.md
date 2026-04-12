# 2026-04-12 tempstorage 包迁移至仓库根目录

## 概述

将临时存储实现从 `pkg/tempstorage` 迁至仓库根目录 `tempstorage/`，与 `learning`、`status`、`webhook` 等顶层功能包保持一致；原 `pkg/` 目录已移除（此前仅包含该包）。

## 变更摘要

| 项目 | 说明 |
|------|------|
| 新路径 | `tempstorage/tempstorage.go`，包名仍为 `tempstorage` |
| 旧路径 | ~~`pkg/tempstorage/tempstorage.go`~~（已删除） |
| import | `github.com/originaleric/digeino/tempstorage`（不再使用 `.../pkg/tempstorage`） |

## 受影响的代码引用

以下文件中的 import 已更新为 `github.com/originaleric/digeino/tempstorage`：

- `tools/storage/write_review_file.go`
- `tools/ui_ux/preview_apply.go`
- `tools/ui_ux/preview_apply_test.go`
- `tools/ui_ux/preview_manifest.go`
- `tools/ui_ux/preview_runtime.go`
- `tools/ui_ux/preview_tools.go`
- `tools/ui_ux/preview_zip.go`

## 文档同步

与路径相关的说明已对齐，包括：

- `docs/ideas/2026-03-28_UI预览与用户编辑工具方案.md`
- `docs/updates/2026-03-18_临时存储统一方案实施完成.md`
- `docs/updates/2026-03-28_UI预览与Patch工具实施.md`
- `docs/ideas/2026-03-18_临时存储统一方案.md`
- `tools/ui_ux/README.md`

## 对外集成说明

若其他仓库（如 DigPdf、DigFlow宿主）曾引用 `github.com/originaleric/digeino/pkg/tempstorage`，需改为：

```go
import "github.com/originaleric/digeino/tempstorage"
```

API（如 `SaveForReview`、`SaveBytesForReview`、`ValidateRelativePath`、`GetBasePath`、Context 键常量）未变。

## 相关背景

- 原实施记录：[2026-03-18 临时存储统一方案实施完成](./2026-03-18_临时存储统一方案实施完成.md)
