package research

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// JinaReaderRequest 网页读取请求
type JinaReaderRequest struct {
	URL string `json:"url" jsonschema:"required,description=要读取并转换为 Markdown 的网页 URL"`
}

// JinaReaderResponse 网页读取响应
type JinaReaderResponse struct {
	Markdown string `json:"markdown" jsonschema:"description=转换后的 Markdown 内容"`
}

// JinaReader 使用 r.jina.ai 将网页转换为 Markdown
// 这是一个极其强大的抓取方案，能够有效绕过微信公众号等平台的反爬验证
func JinaReader(ctx context.Context, req *JinaReaderRequest) (*JinaReaderResponse, error) {
	if req.URL == "" {
		return nil, fmt.Errorf("URL 不能为空")
	}

	// 构造 Jina Reader URL
	jinaURL := fmt.Sprintf("https://r.jina.ai/%s", req.URL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", jinaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Header 为 text/event-stream
	// 实验证明 Jina Reader 的免费模式对 event-stream 的反爬支持更佳
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求 Jina Reader 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jina Reader 返回错误: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 简单的 EventStream 解析逻辑 (Jina Reader 的格式通常是 data: {...})
	// 我们尝试寻找 data: 开头的行并提取其中的 content
	content := string(body)

	// 如果是 event-stream 格式，尝试提取数据部分
	// 为简单起见，如果包含 "data: {" 则尝试剥离前缀，
	// 但其实 Jina Reader 在 event-stream 模式下如果是完整返回，
	// 我们也可以直接返回内容并让 LLM 行文时自动忽略协议头，
	// 或者尝试进行精细化提取：

	extracted := parseJinaEventStream(content)
	if extracted != "" {
		content = extracted
	}

	return &JinaReaderResponse{
		Markdown: content,
	}, nil
}

func parseJinaEventStream(raw string) string {
	// 简单的状态机提取 data 字段
	// Jina Reader 的 event stream 格式示例:
	// data: {"content": "...", "title": "..."}

	// 这里我们寻找最后一个 data: 块，或者所有 data 块的内容拼接
	// 为简单起见，我们直接处理常见格式
	if !strings.Contains(raw, "data: {") {
		return ""
	}

	type jinaData struct {
		Content string `json:"content"`
	}

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			jsonData := strings.TrimPrefix(line, "data: ")
			var d jinaData
			if err := json.Unmarshal([]byte(jsonData), &d); err == nil && d.Content != "" {
				return d.Content // 拿到内容直接返回
			}
		}
	}
	return ""
}

// NewJinaReaderTool 创建 Jina Reader 网页读取工具
func NewJinaReaderTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("research_jina_reader", "使用 Jina Reader 深度读取网页内容并转换为 Markdown。相比普通爬虫，它能更有效地处理微信公众号、知乎等复杂页面的反爬拦截。", JinaReader)
}
