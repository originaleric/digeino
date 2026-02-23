package research

import (
	"context"
	"fmt"
	"time"

	htmlmd "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

// LocalScraperRequest 本地无头浏览器抓取请求
type LocalScraperRequest struct {
	URL string `json:"url" jsonschema:"required,description=要抓取的网页 URL（建议为微信公众号文章等复杂反爬页面）"`
}

// LocalScraperResponse 本地无头浏览器抓取响应
type LocalScraperResponse struct {
	// 提取到的正文纯文本
	Text string `json:"text" jsonschema:"description=提取到的正文纯文本内容"`
	// 将正文 HTML 转换后的 Markdown（便于与 firecrawl/jina 输出对齐）
	Markdown string `json:"markdown" jsonschema:"description=基于正文 HTML 转换得到的 Markdown 内容"`
}

// LocalScrape 使用本地无头浏览器抓取网页正文内容
// 主要面向微信公众平台等对 HTTP 爬虫有强反爬策略的站点。
func LocalScrape(ctx context.Context, req *LocalScraperRequest) (*LocalScraperResponse, error) {
	if req == nil || req.URL == "" {
		return nil, fmt.Errorf("url 不能为空")
	}

	// 为整次抓取设置一个总超时时间，避免单次调用时间过长
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// 启动本地 Chromium，无头模式 + no-sandbox，便于在服务器环境运行
	l := launcher.New().
		Headless(true).
		NoSandbox(true)

	browserURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("启动本地浏览器失败: %w", err)
	}

	browser := rod.New().ControlURL(browserURL).Timeout(60 * time.Second)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("连接本地浏览器失败: %w", err)
	}
	defer func() {
		_ = browser.Close()
	}()

	// 使用 stealth 插件创建页面，隐藏自动化特征
	page, err := stealth.Page(browser)
	if err != nil {
		return nil, fmt.Errorf("创建 stealth 页面失败: %w", err)
	}
	defer func() {
		_ = page.Close()
	}()

	// 导航到目标地址
	if err := page.Timeout(30 * time.Second).Navigate(req.URL); err != nil {
		return nil, fmt.Errorf("页面导航失败: %w", err)
	}

	// 等待页面加载完成
	page.MustWaitLoad()

	// 微信公众号文章正文通常在 #js_content 中
	el, err := page.Timeout(20 * time.Second).Element("#js_content")
	if err != nil {
		return nil, fmt.Errorf("未能在页面中找到正文容器 (#js_content): %w", err)
	}

	// 获取 HTML 和纯文本
	html, err := el.HTML()
	if err != nil {
		return nil, fmt.Errorf("读取正文 HTML 失败: %w", err)
	}

	text, err := el.Text()
	if err != nil {
		return nil, fmt.Errorf("读取正文内容失败: %w", err)
	}

	// 将 HTML 转换为 Markdown
	converter := htmlmd.NewConverter("", true, nil)
	md, err := converter.ConvertString(html)
	if err != nil {
		// 转换失败时不直接中断，返回文本并带上错误信息
		return &LocalScraperResponse{
			Text:     text,
			Markdown: "",
		}, fmt.Errorf("HTML 转 Markdown 失败: %w", err)
	}

	return &LocalScraperResponse{
		Text:     text,
		Markdown: md,
	}, nil
}

// NewLocalScraperTool 创建本地无头浏览器抓取工具
func NewLocalScraperTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool(
		"research_local_scraper",
		"使用本地无头浏览器 + stealth 绕过复杂反爬（如微信公众号），抓取网页正文内容。",
		LocalScrape,
	)
}

