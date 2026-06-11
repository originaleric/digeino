# Image OCR 工具说明

`tools/ocr` 提供统一的 `image_ocr` 工具与 Go Client。工具输入支持图片 URL、Base64 和受白名单保护的本地文件路径，输出统一为 `OCRResponse`，便于宿主在流程节点、Agent 工具或业务服务中复用。

## Provider 选择

`Tools.OCR.Provider` 用来表达接入协议，而不是固定表达某个模型品牌：

| Provider | 适用场景 | 配置节点 |
|----------|----------|----------|
| `openai-compatible-vision` | OpenAI Chat Completions 兼容的视觉模型，如 Qwen VL、DashScope 兼容模式、自建 OpenAI compatible gateway | `Tools.OCR.OpenAICompatible` |
| `multipart-ocr-http` | 内部部署或第三方提供的 `multipart/form-data` OCR HTTP 服务 | `Tools.OCR.MultipartOCR` |
| `deepseek-ocr` | 兼容历史 DeepSeek OCR 配置；保留 `chat` 与 `ocr_endpoint` 两种模式 | `Tools.OCR.DeepSeek` |

空 `Provider` 仍会回退到 `deepseek-ocr`，用于兼容已有部署。新配置建议显式填写 `openai-compatible-vision` 或 `multipart-ocr-http`。

## OpenAI 兼容视觉模型

当 OCR 能力来自 OpenAI compatible 的视觉模型时，使用 `openai-compatible-vision`。例如 DashScope / Qwen VL：

```yaml
Tools:
  OCR:
    Enabled: true
    Provider: "openai-compatible-vision"
    OpenAICompatible:
      Model: "qwen3-vl-plus"
      ApiKey: "${QWEN_API_KEY}"
      BaseUrl: "https://dashscope.aliyuncs.com/compatible-mode/v1"
```

`BaseUrl` 填到兼容 API 的版本根路径即可，不要包含 `/chat/completions`。API Key 会按顺序读取配置值、`OPENAI_COMPATIBLE_VISION_API_KEY`、`QWEN_API_KEY`、`DASHSCOPE_API_KEY`、`OPENAI_API_KEY`。

## 通用 Multipart OCR HTTP 服务

当 OCR 是内部部署的 HTTP 服务，且接收 `multipart/form-data` 文件上传时，使用 `multipart-ocr-http`。这个 provider 不绑定 DeepSeek，也可以接 PaddleOCR、Tesseract、DeepSeek-OCR 本地封装服务或其它自研 OCR。

```yaml
Tools:
  OCR:
    Enabled: true
    Provider: "multipart-ocr-http"
    MultipartOCR:
      ApiKey: "${MULTIPART_OCR_API_KEY}"
      BaseUrl: "http://ocr-service.internal"
      OCREndpoint: "/v1/ocr"
      Model: "internal-ocr"
      ResponseTextPath: "data.text"
      FileField: "file"
      PromptField: "prompt"
      LanguageField: "language"
```

请求会上传规范化后的图片二进制，并按配置追加 prompt、language 字段。若内部服务不需要 prompt 或 language，可将对应字段名设为 `"-"`：

```yaml
PromptField: "-"
LanguageField: "-"
```

返回体需要是 JSON，`ResponseTextPath` 用点号路径读取识别文本；例如 `data.text` 对应：

```json
{"data":{"text":"识别结果"}}
```

API Key 会按顺序读取配置值、`MULTIPART_OCR_API_KEY`、`OCR_HTTP_API_KEY`。为空时不发送 `Authorization` 头，适合内网无鉴权服务。

## 输入安全

- `AllowedImageDomains` 可限制 URL 图片来源域名。
- `BlockPrivateNetworks` 用于阻止 URL 下载解析到本机、内网或私有地址。
- `AllowedFilePaths` 非空时才允许读取本地文件路径。
- `AllowedMimeTypes` 和 `MaxImageBytes` 会在图片进入 provider 前统一校验。

## 扩展 Provider

外部包可以实现：

```go
type OCRProvider interface {
	Recognize(ctx context.Context, req *OCRRequest, img *OCRImage) (*OCRResponse, error)
	Name() string
}
```

随后调用 `ocr.RegisterOCRProvider(provider)` 注册。`OCRImage` 会携带已经过安全校验的 `DataURL`、原始 `Data`、`MimeType` 与 `Source`。
