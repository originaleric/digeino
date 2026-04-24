# 2026-04-24 pkg/render 配置与解析呈现管线整理

## 概述

对 `pkg/render` 的 **YAML 配置形态** 与 **加载逻辑** 做了一轮整理：呈现段统一命名、解析与呈现之间的对应关系可配置、代码围栏与思考标签的键名对齐，并支持在 **`parse` / `parse_render` 下并列扩展键** 而不增加单独的 `custom_tag_pairs` 段。业务侧加载完整配置后仍通过 `RenderConfig` 一次性得到 `Options` 与呈现配置。

## 呈现配置（原 `html:` → `render:`）

- YAML 顶层推荐使用 **`render:`** 描述「块 → HTML」的样式与策略（`thinking` / `code` 的 **outer、inner**，`markdown` 的 **sanitize**，以及 **document** 整页壳）。
- 旧键 **`html:`** 仍可读；若 **`render:`** 与 **`html:`** 同时存在，**`render:` 优先覆盖**。
- Go 侧 **`RenderConfig`** 中呈现结果字段由 **`HTML` 改为 `Render`**（类型仍为 `HTMLPresentationConfig`），调用 `BlocksToHTMLWithConfig` / `WrapDocumentWithConfig` 时传入 **`&rc.Render`**。

## 解析配置

- **`parse.code_fence`** 使用 **`open` / `close`**（与思考标签一致）；省略 `close` 时默认等于 `open`。旧键 **`opening`** 仍可读作 `open`。
- **`parse_render:`** 声明内置三段映射：`thinking_tag_pairs` → `render` 小节名、`code_fence` → 代码段小节名、`markdown` → 正文段小节名（默认分别为 `thinking`、`code`、`markdown`）。加载后的有效映射在 **`RenderConfig.ParseRender`**。
- 可在 **`parse_render`** 下增加 **扩展键**（例如 `my_tag: thinking`），须与 **`parse.my_tag`** 的 `{ open, close }` 成对出现；仅写 `parse_render` 扩展键而不写对应 `parse` 块会报错。
- **`parse:`** 下除 **`thinking_tag_pairs`**、**`code_fence`** 外，可 **并列任意键名**；值为 **`open` / `close`** 时自动并入 **思考区识别**（块类型仍为 thinking，默认走 **`render.thinking`**）。**`parse_render` 下同名扩展键可省略**；若写了则必须与 `parse` 中的块一致。
- 根级 **`thinking_tag_pairs`** 改为 **指针类型语义**：键省略与 **`thinking_tag_pairs: []`**（显式关闭思考识别）可区分；合并列表时注意空切片与 `append` 行为已处理。

## 不再使用的配置

- **`custom_tag_pairs`**（根级或 `parse` 下）已移除；若 YAML 仍保留该键，标准解码不会消费它（等同于无效键）。请改为 **`parse.<你的键名>: { open, close }`**。

## 主要涉及路径

- `pkg/render/config/config.yaml`（嵌入默认）、`pkg/render/config/config.example.yml`
- `pkg/render/config.go`、`pkg/render/config_render.go`、`pkg/render/config_parse_yaml.go`
- `pkg/render/options.go`、`pkg/render/parse.go`、`pkg/render/prefix.go`
- `pkg/render/html_pres.go`、`pkg/render/html/html.go`
- `pkg/render/README.md`
- `pkg/render/config_test.go` 等单测

## 使用提示

- 只关心解析、不关心呈现时，仍可用 **`OptionsFromYAML`** / **`LoadOptionsFromFile`**（忽略 `render:` / `html:`）。
- 需要 **Parse + HTML** 时推荐 **`LoadRenderConfigFromFile`**，对 **`rc.Options`** 做 **`Parse`**，对 **`&rc.Render`** 做 **`BlocksToHTMLWithConfig`** / **`WrapDocumentWithConfig`**。
- 更细的字段说明与表格见 **`pkg/render/README.md`**。

## 路线 C：`pkg/render/uimodel`（块 → 卡片 JSON）

- 新增 **`pkg/render/uimodel`**：`BuildHybridPayload`、`BuildUIModel`、`BuiltinMapping` / `LoadMappingFromFile`，以及 **`mapping_hash`**（`MappingFingerprint`）。
- 示例映射：**`pkg/render/uimodel/config/uimodel.example.yaml`**
- 设计说明见 **`docs/ideas/2026-04-24_前后端渲染映射规则治理方案.md`**
