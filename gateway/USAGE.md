# Wrap → Expose 工具开发指南

本文说明如何把 DigEino 里**已有的 Legacy 工具**（进程内 Eino `BaseTools()`）包装成可通过 **HTTP / Collector / MCP / stdio** 调用的网关工具。

策略来自方案：**Keep → Wrap → Expose**。

- **Keep**：不改动 `tools/` 下原有 `NewXxxTool` / `InferTool` 实现。
- **Wrap**：在 `gateway/executor` 增加 `registry.Entry`（元数据 + Handler）。
- **Expose**：在 `gateway/bootstrap.go` 注册；由 `runtime` 统一执行，各出口自动可见。

---

## 1. 先确认要不要 Wrap

适合 Wrap → Expose 的工具通常满足：

- 需要被**远程宿主**、**本地 Collector** 或 **IDE MCP** 调用；
- 输入/输出能整理成 JSON；
- 有明确的安全边界（域名、路径、凭证等）。

不必 Wrap 的情况：

- 只在同一 Go 进程内给 DigFlow/Knowledge 用 → 继续 `BaseTools()` 即可。

---

## 2. 目录与命名约定

| 层级 | 路径 | 说明 |
|------|------|------|
| Legacy 实现 | `tools/<pkg>/` | 原有业务函数，如 `research.BrowserBrowse` |
| Gateway 包装 | `gateway/executor/<name>.go` | 一个文件一个网关工具族；**使用说明**见 [`executor/README.md`](executor/README.md) |
| 注册 | `gateway/bootstrap.go` | `NewRegistry()` 里 `reg.Register(...)` |
| 协议类型 | `gateway/protocol/` | `ToolCall` / `ToolResult`，一般不用改 |

**网关工具名**建议用点分式，与 Eino 内部名区分：

| Legacy（Eino） | Gateway（Expose） |
|----------------|-------------------|
| `browser_browse` | `browser.browse` |
| `research_read` | `file.read` |

---

## 3. Wrap：新增 executor（分步）

### 步骤 3.1 选定 Legacy 函数

在 `tools/` 中找到可复用的入口，例如：

```go
// tools/research/general_researcher.go
func ReadFile(ctx context.Context, req *ReadFileRequest) (*ReadFileResponse, error)
```

确认其请求/响应结构体已有 `json` 标签（`InferTool` 生成的通常已有）。

### 步骤 3.2 新建 `gateway/executor/xxx.go`

模板如下（按实际工具替换）：

```go
package executor

import (
	"context"

	"github.com/originaleric/digeino/gateway/protocol"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/tools/research" // 按包调整
)

const ToolMyExample = "my.example" // 网关侧工具名

// MyExampleEntry 返回 registry.Entry。
func MyExampleEntry(/* 构造时注入的策略参数 */) registry.Entry {
	return registry.Entry{
		Descriptor: protocol.ToolDescriptor{
			Name:        ToolMyExample,
			Description: "一句话说明工具做什么。",
			InputSchema: registry.MustSchema(map[string]any{
				"type":     "object",
				"required": []string{"field_a"},
				"properties": map[string]any{
					"field_a": map[string]string{"type": "string"},
				},
			}),
			OutputSchema: registry.MustSchema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]string{"type": "string"},
				},
			}),
			Capabilities: []string{"example"},
			Risk:         "network", // 或 filesystem / messaging 等
			RequiresUserApproval: false,
		},
		Handler: func(ctx context.Context, call *protocol.ToolCall) (map[string]any, []protocol.Artifact, error) {
			// 1. 解析 input
			in, err := decodeInput[research.MyRequest](call)
			if err != nil {
				return nil, nil, err
			}

			// 2. 策略校验（见第 4 节）
			// ...

			// 3. 调用 Legacy
			resp, err := research.MyFunc(ctx, &in)
			if err != nil {
				return nil, nil, err
			}

			// 4. 转成 map 输出（将进入 ToolResult.output）
			return map[string]any{
				"result": resp.Something,
			}, nil, nil
		},
	}
}
```

**Handler 签名**（固定）：

```go
func(ctx context.Context, call *protocol.ToolCall) (map[string]any, []protocol.Artifact, error)
```

- 返回的 `map[string]any` 会由 `runtime` 序列化为 `ToolResult.output`。
- 大对象（截图、文件）走 `[]protocol.Artifact`，不要塞进 `output`（见 3.5）。

### 步骤 3.3 解析入参

优先复用 `gateway/executor/helpers.go`：

```go
in, err := decodeInput[research.ReadFileRequest](call)
```

若 Legacy 请求类型与网关 input 不一致，可单独定义 `xxxInput` 结构体再手动映射。

### 步骤 3.4 填写 ToolDescriptor

| 字段 | 必填 | 说明 |
|------|------|------|
| `Name` | 是 | 宿主调用时的 `tool` 字段 |
| `Description` | 是 | Manifest / MCP 展示 |
| `InputSchema` | 建议 | JSON Schema，`registry.MustSchema(...)` |
| `OutputSchema` | 可选 | 便于宿主生成 UI |
| `Capabilities` | 可选 | 如 `browser`, `cookie.local` |
| `Risk` | 建议 | `network` / `filesystem` / `messaging` |
| `RequiresUserApproval` | 可选 | 敏感操作为 `true` |

### 步骤 3.5 处理 Artifact（可选）

截图等大二进制：

1. 在 Handler 里拿到 `[]byte` 或 base64；
2. 通过构造时注入的 `artifact.Store` 写入（参考 `browser_browse.go` 的 `artifact.PutBase64PNG`）；
3. 在 `output` 里只返回 `screenshot_artifact_id`，在 `artifacts` 里返回 `protocol.Artifact`。

HTTP 下载：`GET /artifacts/{id}`（需 `Gateway.ArtifactEnabled: true`）。

---

## 4. 策略与安全校验

在 Handler 内、调用 Legacy **之前**完成校验。

### 4.1 访问 URL 的工具

```go
if err := validateCallURL(in.URL, call, configDomains); err != nil {
	return nil, nil, err
}
if err := validateCookieDomain(in.UseCookieDomain, call, configDomains); err != nil {
	return nil, nil, err
}
```

- `configDomains` 来自 `Tools.LocalBrowser.AllowedDomains`；
- `call.Policy.AllowedDomains` 可覆盖单次调用（见 `policy.MergeDomains`）。

### 4.2 读本地文件

```go
if err := validateReadPath(in.Path, allowedPaths); err != nil {
	return nil, nil, err
}
```

`allowedPaths` 对应 `Gateway.AllowedReadPaths`。

### 4.3 工具白名单

无需在 Handler 里写：注册后由 `runtime` 根据 `Gateway.AllowedTools` / `Collector.AllowedTools` 统一拦截。

### 4.4 错误码

策略失败尽量带前缀，便于 `runtime` 映射为结构化错误：

- `DOMAIN_NOT_ALLOWED`
- `TOOL_NOT_ALLOWED`
- `INVALID_INPUT`

使用 `gateway/policy` 中常量或相同字符串前缀。

---

## 5. Expose：注册到 Registry

编辑 `gateway/bootstrap.go` 的 `NewRegistry()`：

```go
reg.Register(executor.MyExampleEntry(domains /* 或其它策略参数 */))
```

有条件注册示例（`file.read`）：

```go
if len(readPaths) > 0 {
	reg.Register(executor.FileReadEntry(readPaths))
}
```

注册后，以下出口**自动**包含新工具，无需分别改代码：

- `digeino gateway` → HTTP `/manifest`、`/tools/call`
- `digeino collector` → WS `collector_manifest`、`tool_result`
- `digeino mcp` → MCP `tools/list`、`tools/call`
- `digeino stdio` → `get_manifest` / `tool_call`

---

## 6. 配置白名单

`config/config.yaml`：

```yaml
Gateway:
  AllowedTools:
    - browser.browse
    - my.example          # 新增网关工具名
  # 若工具读文件：
  AllowedReadPaths:
    - "./storage"
    - "/path/to/workspace"

Collector:
  AllowedTools:
    - my.example          # Collector 侧也要加（或留空继承 Gateway）
```

浏览器类工具还需：

```yaml
Tools:
  LocalBrowser:
    Enabled: true
    AllowedDomains:
      - example.com
```

---

## 7. 验证清单

### 7.1 编译与单测

```bash
go test ./gateway/executor/... ./gateway/runtime/... -count=1
```

可为 executor 加表驱动测试：构造 `protocol.ToolCall`，直接调 `Entry.Handler`。

### 7.2 HTTP Manifest

```bash
go run ./cmd/digeino gateway --config config/config.yaml
curl -s http://127.0.0.1:8787/manifest | jq '.tools[].name'
```

应能看到 `my.example`。

### 7.3 调用

```bash
curl -s -X POST http://127.0.0.1:8787/tools/call \
  -H 'Content-Type: application/json' \
  -d '{
    "type": "tool_call",
    "id": "test_1",
    "tool": "my.example",
    "input": { "field_a": "value" }
  }'
```

### 7.4 MCP（可选）

```bash
go run ./cmd/digeino mcp --config config/config.yaml
```

在 Cursor MCP 配置里指向该命令，确认 IDE 能列出并调用新工具。

---

## 8. 参考实现（仓库内）

| 网关工具 | 文件 | Legacy |
|---------|------|--------|
| `browser.browse` | `gateway/executor/browser_browse.go` | `research.BrowserBrowse` |
| `browser.snapshot` | `gateway/executor/browser_snapshot.go` | `research.BrowserSnapshot` |
| `browser.action` | `gateway/executor/browser_action.go` | `research.BrowserAction` |
| `wechat.article.read` | `gateway/executor/wechat_article.go` | 组合 `BrowserBrowse` |
| `file.read` | `gateway/executor/file_read.go` | `research.ReadFile` |

公共辅助：`gateway/executor/helpers.go`  
注册入口：`gateway/bootstrap.go`  
执行引擎：`gateway/runtime/runtime.go`

---

## 9. 常见问题

**Q：改了 `tools/` 里 Legacy 函数，网关要改吗？**  
若入参/出参兼容，通常只改 Legacy；若 JSON 字段变了，同步改 `InputSchema` 和 `decodeInput` 类型。

**Q：Eino 工具名和网关名必须一致吗？**  
不必。网关名给宿主用；Legacy 名仅进程内 `BaseTools()` 使用。

**Q：注册后 MCP 没有 input schema？**  
当前 MCP 仅暴露 `Description`；完整 schema 在 HTTP `/manifest` 的 `input_schema` 字段。

**Q：敏感工具如何默认关闭？**  
不在 `NewRegistry` 里 `Register`，或不要加入 `AllowedTools`；需要时再开配置项注册（同 `file.read`）。

---

## 10. 最小 Checklist

- [ ] 在 `gateway/executor/` 新增 `XxxEntry()`，实现 `Handler`
- [ ] 填写 `ToolDescriptor`（含 `InputSchema`）
- [ ] 调用前做策略校验（域名/路径等）
- [ ] 在 `gateway/bootstrap.go` 的 `NewRegistry` 中 `Register`
- [ ] 更新 `config/config.yaml` 的 `Gateway.AllowedTools`（及 Collector）
- [ ] 按需配置 `LocalBrowser` / `AllowedReadPaths` / `Artifact`
- [ ] `curl /manifest` 与 `POST /tools/call` 验证

完成以上步骤后，新工具即完成 **Wrap → Expose**。
