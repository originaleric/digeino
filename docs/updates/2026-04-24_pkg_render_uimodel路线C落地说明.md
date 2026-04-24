# 2026-04-24 pkg/render uimodel 路线C落地说明

## 概述

本次在 `pkg/render` 内新增 `uimodel` 子包，落地“路线 C（混合治理）”：

- 后端保留 `blocks` 作为真源数据；
- 可选生成 `ui_model`（卡片 JSON）供前端优先渲染；
- 同时输出 `mapping_version`、`mapping_source`、`mapping_hash` 等治理元数据，支持灰度、回滚与跨端一致性追踪。

该改动不改变既有 `Parse -> []Block -> HTML` 主链路，属于增量能力。

## 新增内容

### 1) `pkg/render/uimodel` 子包

- `types.go`
  - 定义 `UIModel`、`Card`、`HybridPayload`、`OutputMode`（`blocks` / `ui_model` / `both`）。
- `mapping.go`
  - 定义映射配置 `Mapping` 与 `BlockProfile`；
  - 提供 `BuiltinMapping()`；
  - 提供 `LoadMappingFromBytes/File/Reader()`，并与内置默认映射做缺省合并。
- `hash.go`
  - 提供 `MappingFingerprint()`，生成稳定 `mapping_hash`（SHA-256）。
- `build.go`
  - 提供 `BuildUIModel()`（`[]Block -> UIModel`）；
  - 提供 `BuildHybridPayload()`（按输出模式生成混合载荷）；
  - 提供 `ParseOutputMode()`（默认 `both`）。
- `config/uimodel.example.yaml`
  - 提供业务可复制的映射配置样例（`block_profiles`）。

### 2) 文档更新

- `pkg/render/README.md`
  - 新增 `uimodel` 子包说明；
  - 新增「3b. 路线 C：块 -> 卡片 JSON」使用示例；
  - 补充前端“优先 `ui_model`、失败回退 `blocks`”策略建议。
- `docs/ideas/2026-04-24_前后端渲染映射规则治理方案.md`
  - 增加“DigEino 已落地（路线 C 首版）”说明与路径指引。

## 协议形状（建议）

可由业务后端按需返回：

- `blocks: Block[]`（可选）
- `ui_model: { schema_version, cards[] }`（可选）
- `mapping_version` / `mapping_source` / `mapping_hash` / `mapping_changed_at`（可选）

推荐默认输出模式为 `both`，便于前端灰度与回退。

## 首版边界

- 块类型仍限定三类：`markdown` / `thinking` / `code`。
- 首版卡片类型默认：
  - `MarkdownCard`
  - `ReasoningCard`
  - `CodeCard`
- 不引入表达式引擎与数据库动态规则，先以 YAML + 代码加载实现。

## 测试与验证

新增 `pkg/render/uimodel/uimodel_test.go`，覆盖：

- 默认映射构建；
- 输出模式分支；
- 映射覆盖与缺省合并；
- 指纹稳定性；
- 非法配置校验。

`go test ./pkg/render/...` 通过。
