# pkg/render

与具体框架无关：将 **助手文本** 解析为可渲染块（`markdown`、`thinking`、`code`）。面向模型返回之后的 **宿主侧** 管线，**不是** 给大模型用的 Tool。

## 子包

| 路径 | 作用 |
|------|------|
| `github.com/originaleric/digeino/pkg/render` | 核心：`Parse`、`ParseStablePrefix`、类型 — **不** `import cloudwego/eino` |
| `github.com/originaleric/digeino/pkg/render/html` | `BlocksToHTMLWithConfig`、`WrapDocumentWithConfig`（呈现由 `config.yaml` 的 `render:` 或代码传入） |
| `github.com/originaleric/digeino/pkg/render/eino` | `rendereino`：`schema.Message` → 文本 → `Parse` |

## 如何使用

### 0. YAML 配置（推荐维护方式）

库内嵌默认：`pkg/render/config/config.yaml`（随 `go:embed` 打包），包含 **`parse:`**（块识别）、**`parse_render:`**（parse 规则与 `render:` 小节名的对应表）、**`render:`**（呈现：标签/class、markdown 消毒、`WrapDocument` 的 `document`）。旧键 **`html:`** 仍可读，与 **`render:`** 同时存在时 **`render:` 覆盖 `html:`**。`Parse(..., Options{})` 未指定时仍用嵌入文件中的思考标签与 **\`\`\`** 围栏；`BlocksToHTML(nil)` / `WrapDocument` 会通过 `DefaultHTMLPresentation()` 读入嵌入文件里的 **`render:`**。

**推荐**：一次加载完整配置，解析与导出 HTML 共用：

```go
rc, err := render.LoadRenderConfigFromFile("config/render/config.yaml")
if err != nil {
    return err
}
blocks, err := render.Parse(text, rc.Options)
body, err := renderhtml.BlocksToHTMLWithConfig(blocks, &rc.Render)
full := renderhtml.WrapDocumentWithConfig("标题", body, &rc.Render)
```

仅解析、沿用嵌入的 HTML 呈现时，可继续 `LoadOptionsFromFile` / `OptionsFromYAML`。嵌入全文用 `render.LoadEmbeddedRenderConfig()`。

`thinking_tag_pairs` 行为：

| YAML 中 `thinking_tag_pairs` | 效果 |
|------------------------------|------|
| 省略或 `null` | 等价零值 `Options{}`，`Parse` 时 withDefaults 使用**嵌入**列表 |
| `[]` 空列表 | **关闭**思考区识别 |

**`parse` 扩展键**：与 `thinking_tag_pairs`、`code_fence` **并列**写任意键名（如 `my_tag`），值为 **`open` / `close`** 时自动并入思考区识别（仍为 **thinking** 块，默认走 **`render.thinking`**）。**`parse_render`** 下可写同名键 **`my_tag: thinking`** 显式对应（可省略；若写了 **`parse_render.my_tag`** 则必须有 **`parse.my_tag`**）。

**`parse_render:`（可选）**：写明「parse 里的哪一块」对应 **`render:` 下的哪个小节名**（便于自定义命名时一眼对齐）。省略时等价于 `thinking_tag_pairs→thinking`、`code_fence→code`、`markdown→markdown`。加载结果在 **`RenderConfig.ParseRender`**。若改名，例如 `code_fence: snippet`，则 **`render:` 里必须用 `snippet:`** 配代码块样式；`document` 键名固定，不参与映射。

**`parse:`（可选，块识别）**：把「怎么认出 code / thinking」收拢到同一段；根级 `thinking_tag_pairs` 仍可单独使用。

| 键 | 含义 |
|----|------|
| `parse.thinking_tag_pairs` | 若 YAML 里**写出**该键（含 `[]`），则**整表覆盖**根级 `thinking_tag_pairs`；省略则沿用根级 |
| `parse.code_fence.open` / `close` | 与 `thinking_tag_pairs` 相同键名：起始行 / 闭合行前缀；省略 `close` 时默认等于 `open`。旧键 `opening` 仍可读作 `open` |

代码里也可直接设 `render.Options{ CodeFence: render.CodeFenceConfig{Open: "~~~", Close: "~~~"}, ThinkingTagPairs: ... }`。块类型枚举（`markdown` / `thinking` / `code`）仍在解析器内固定；**`render:`**（及旧 **`html:`**）只控制 **生成 HTML** 时的标签、class、消毒与整页样式。

**`render:` 小节键约定**：在默认 **`parse_render`** 下，小节名为 `thinking` / `code` / `markdown`（及固定的 `document`）。`thinking` 与 `code` 使用同一形状——`outer`、`inner`，每项为 `tag` + `class`。`markdown` 主要为 `sanitize`（goldmark 之后的 HTML 消毒）。兼容旧 YAML：`thinking` 的 `wrapper_*` / `content_*`、`code` 的 `pre_class` 等会在加载时迁到 `outer`/`inner`；`sanitize_policy` 可读作 `sanitize`。

| 段 | 键 | 含义 |
|----|-----|------|
| `thinking` | `outer` | 思考块最外层容器（如 `aside` + `llm-thinking`） |
| `thinking` | `inner` | 内层包转义后的正文（如 `pre` + `llm-thinking-pre`） |
| `code` | `outer` | 代码块外层（通常为 `pre` + `llm-code`） |
| `code` | `inner` | 内层（通常为 `code`）；`language_class_prefix` 挂在此项：有语言时在 `class` 上追加 `前缀+语言`（如 `language-go`） |
| `markdown` | `sanitize` | `ugc` \| `strict` \| `noop` |
| `document` | `lang`、`inline_css` | 仅整页 `WrapDocument*`：`lang` → `<html lang>`；样式写在 `.llm-render-doc` 与块 class 上（键名始终为 `document`，不受 `parse_render` 改名影响） |

### 1. 解析完整助手文本（终态）

```go
import "github.com/originaleric/digeino/pkg/render"

blocks, err := render.Parse(assistantText, render.Options{})
if err != nil {
    return err
}
for _, b := range blocks {
    switch b.Kind {
    case render.BlockKindMarkdown:
        // b.Content
    case render.BlockKindThinking:
        // b.Content
    case render.BlockKindCode:
        // b.Language, b.Content
    }
}
```

自定义思考标签（不配则用 `DefaultThinkingTagPairs()`）：

```go
blocks, err := render.Parse(text, render.Options{
    ThinkingTagPairs: []render.ThinkingTagPair{
        {Open: "<my_think>", Close: "</my_think>"},
    },
})
```

### 2. 流式过程中的稳定前缀

在缓冲区每次变长后调用；已结构化的部分在 `Blocks`，未闭合的围栏/思考留在 `Remainder`，可原样拼在 UI 尾部。

```go
res, err := render.ParseStablePrefix(buf.String(), render.Options{})
if err != nil {
    return err
}
// res.Blocks — 已可安全渲染的块
// res.Remainder — 仍待流式补全的原文后缀
```

流式结束后应再执行一次 `render.Parse` 得到与终态一致的 `[]Block`。

### 3. 转成 HTML 片段或整页

```go
import (
    "github.com/originaleric/digeino/pkg/render"
    renderhtml "github.com/originaleric/digeino/pkg/render/html"
)

// 使用与 Parse 同一份 YAML 中的 render: 段
rc, err := render.LoadRenderConfigFromFile("config/render/config.yaml")
if err != nil {
    return err
}
body, err := renderhtml.BlocksToHTMLWithConfig(blocks, &rc.Render)
if err != nil {
    return err
}
full := renderhtml.WrapDocumentWithConfig("标题", body, &rc.Render)
```

仍可用 `BlocksToHTML(blocks)` / `WrapDocument`（内部等价于 `nil` 配置 → `DefaultHTMLPresentation()`，即嵌入 `config.yaml` 的 `render:`）。

**说明**：`BlocksToHTML` 只产出 **HTML 片段**（无 `<html>`/`<body>`）。`WrapDocument` 才生成整页，并把片段包在 `<div class="llm-render-doc">` 里；`document.inline_css` 里的版式请优先写 `.llm-render-doc` 与块 class（`.llm-thinking`、`.llm-code`），不要依赖 `body`，以免和「仅片段」用法混淆。

### 4. 从 Eino `schema.Message` 解析

包名为 `rendereino`，路径为 `.../pkg/render/eino`。

```go
import (
    "github.com/originaleric/digeino/pkg/render"
    "github.com/originaleric/digeino/pkg/render/eino"
)

blocks, err := rendereino.ParseMessage(msg, render.Options{})
```

`MessageToRenderableText` 会把 `ReasoningContent`、正文 `Content` 以及 `ToolCalls` 摘要拼成可解析字符串；也可先取字符串再 `render.Parse`。

### 5. 写入 HTML 文件（DigEino + tempstorage）

`context` 需通过 `tempstorage` 约定注入 `workspace_path` 或 `agent_session_id`（见 `tempstorage` 包说明）。

```go
import "github.com/originaleric/digeino/internal/renderexport"

path, err := renderexport.SaveAssistantHTML(ctx, "chat_render/out.html", blocks, "会话标题")
```

## 支持的格式与自定义

### 解析阶段（`Parse` / `ParseStablePrefix`）产出三类块

思考区标签建议用 **`config/config.yaml`（库内参考）或自有 `config.yml`** 维护（见上文「YAML 配置」），无需改 Go 代码即可增删厂商格式。

| `BlockKind` | 含义 | 识别方式（当前实现） |
|-------------|------|----------------------|
| `markdown` | 正文 Markdown | 与代码围栏、思考区错开后的普通文本段 |
| `thinking` | 思考 / 推理 | 由 `Options.ThinkingTagPairs` 成对标签包住的内容（默认含 `<think>`、`<reasoning>` 等，见 `DefaultThinkingTagPairs`） |
| `code` | 代码 | CommonMark 式围栏：起始行以 `Options.CodeFence.Open` 为前缀，闭合行以 `Close` 为前缀（默认与 `Open` 相同）；首行前缀后可带语言标识 |

块类型枚举（`markdown` / `thinking` / `code`）仍在库内固定；思考标签与围栏前缀可通过 YAML **`parse:`** 或 **`Options`** 配置。

### HTML 输出（`pkg/render/html`）

默认结构由 **`config.yaml` → `render:`** 定义（可与内置 `pkg/render/config/config.yaml` 对齐后自行覆盖；旧 **`html:`** 仍可读）：

- **markdown**：goldmark → `markdown.sanitize`（`ugc` / `strict` / `noop`）。
- **thinking**：`thinking.outer` / `thinking.inner`（`tag` + `class`，标签经白名单校验）。
- **code**：`code.outer` / `code.inner`；`language_class_prefix` + 语言名。

若需完全自定义 DOM，仍可遍历 `[]Block` 自写渲染；或 fork 改 `html` 包。

### 你可以怎样自定义？

1. **无需改库**：YAML 的 `thinking_tag_pairs` / `parse:`，或 `render.Options{ ThinkingTagPairs, CodeFence }`，即可适配思考标签与代码围栏前缀。
2. **自行维护 / 扩展**：本仓库开源，可在 `pkg/render` 内增加新 `BlockKind`，或 fork 后改 `parse.go` / `html` 包。

## 流式（Cursor 式）

- **实时 UI**：在客户端缓冲 token；可在不断增长的字符串上按需调用 `ParseStablePrefix`，使已 **闭合** 的围栏代码 / 思考区变成结构化块；将 `Remainder` 当作未定型原文展示。
- **结束**：流式结束后对完整助手字符串调用一次 `Parse`。

## 与 DigEino 集成

- 按 workspace / session 路径导出 HTML：`internal/renderexport.SaveAssistantHTML`（基于 `tempstorage`）。
- Webhook / `StatusCollector`：可选 — 在 `OnComplete` 中调用 `render.Parse`（或 `rendereino.ParseMessage`），若需要结构化事件再把 `[]Block` 挂到载荷上。

## 拆成独立 Go 模块

本目录结构便于日后迁到例如 `github.com/yourorg/llm-render`：**`pkg/render` 与 `pkg/render/html`** 保持零 eino；**`pkg/render/eino`** 留在 DigEino，或放到单独小模块里并同时依赖二者。

## CI 约束（建议）

确保 `pkg/render/*.go` 与 `pkg/render/html/*.go` 下的文件 **不要** `import github.com/cloudwego/eino`（**`pkg/render/eino/`** 子目录除外）。
