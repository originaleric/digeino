package research

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	htmlmd "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/originaleric/digeino/config"
)

var cookieDomainPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

// BrowserBrowseRequest 本地浏览器抓取请求
type BrowserBrowseRequest struct {
	URL             string `json:"url" jsonschema:"required,description=需要访问的目标网页 URL"`
	Action          string `json:"action,omitempty" jsonschema:"description=动作类型，支持 read 或 screenshot，默认 read"`
	WaitSelector    string `json:"wait_selector,omitempty" jsonschema:"description=可选 CSS 选择器，等待目标元素出现后再提取"`
	UseCookieDomain string `json:"use_cookie_domain,omitempty" jsonschema:"description=可选 Cookie 域名（如 weixin.qq.com），用于加载并复用该域的 Cookie"`
}

// BrowserBrowseResponse 本地浏览器抓取响应
type BrowserBrowseResponse struct {
	URL            string `json:"url" jsonschema:"description=最终抓取的 URL"`
	Title          string `json:"title" jsonschema:"description=页面标题"`
	Text           string `json:"text,omitempty" jsonschema:"description=提取的正文纯文本"`
	Markdown       string `json:"markdown,omitempty" jsonschema:"description=转换后的正文 Markdown"`
	ScreenshotBase string `json:"screenshot_base64,omitempty" jsonschema:"description=页面截图 base64（action=screenshot 时返回）"`
}

// BrowserBrowse 使用本地 go-rod 浏览器读取动态页面或截图
func BrowserBrowse(ctx context.Context, req *BrowserBrowseRequest) (*BrowserBrowseResponse, error) {
	if req == nil || strings.TrimSpace(req.URL) == "" {
		return nil, fmt.Errorf("url 不能为空")
	}

	cfg := normalizeLocalBrowserConfig(config.Get().Tools.LocalBrowser)
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		action = "read"
	}
	if action != "read" && action != "screenshot" {
		return nil, fmt.Errorf("action 仅支持 read 或 screenshot")
	}

	targetURL, err := validateURL(req.URL)
	if err != nil {
		return nil, err
	}
	if err := checkAllowedDomain(targetURL.Hostname(), cfg.AllowedDomains); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TotalTimeoutSec)*time.Second)
	defer cancel()

	manager := getBrowserManager()
	session, err := manager.acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer manager.release(session)

	page := session.page.Timeout(time.Duration(cfg.TotalTimeoutSec) * time.Second)

	if req.UseCookieDomain != "" {
		if err := checkAllowedDomain(req.UseCookieDomain, cfg.AllowedDomains); err != nil {
			return nil, fmt.Errorf("cookie 域名不在白名单: %w", err)
		}
		if err := loadDomainCookies(page, req.UseCookieDomain, cfg.CookieStoreDir, targetURL); err != nil {
			return nil, err
		}
	}

	if err := page.Timeout(time.Duration(cfg.NavigateTimeoutSec) * time.Second).Navigate(targetURL.String()); err != nil {
		return nil, fmt.Errorf("页面导航失败: %w", err)
	}
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("页面加载失败: %w", err)
	}
	if req.WaitSelector != "" {
		if _, err := page.Timeout(time.Duration(cfg.WaitSelectorTimeoutSec) * time.Second).Element(req.WaitSelector); err != nil {
			return nil, fmt.Errorf("等待选择器失败: %w", err)
		}
	}

	title, _ := pageTitle(page)
	resp := &BrowserBrowseResponse{
		URL:   targetURL.String(),
		Title: title,
	}

	switch action {
	case "screenshot":
		imgBytes, err := page.Screenshot(true, nil)
		if err != nil {
			return nil, fmt.Errorf("页面截图失败: %w", err)
		}
		resp.ScreenshotBase = base64.StdEncoding.EncodeToString(imgBytes)
	default:
		htmlContent, textContent, err := extractMainContent(page, cfg.WaitSelectorTimeoutSec)
		if err != nil {
			return nil, err
		}
		converter := htmlmd.NewConverter("", true, nil)
		md, err := converter.ConvertString(htmlContent)
		if err != nil {
			return nil, fmt.Errorf("HTML 转 Markdown 失败: %w", err)
		}
		resp.Text = strings.TrimSpace(textContent)
		resp.Markdown = strings.TrimSpace(md)
	}

	cookieDomain := strings.TrimSpace(req.UseCookieDomain)
	if cookieDomain == "" {
		cookieDomain = targetURL.Hostname()
	}
	if err := saveDomainCookies(page, cookieDomain, cfg.CookieStoreDir, targetURL); err != nil {
		return resp, fmt.Errorf("页面抓取成功，但保存 Cookie 失败: %w", err)
	}

	return resp, nil
}

func validateURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("无效 url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("仅支持 http/https URL")
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("url 缺少 host")
	}
	return u, nil
}

func checkAllowedDomain(host string, allowedDomains []string) error {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return fmt.Errorf("空域名")
	}
	if len(allowedDomains) == 0 {
		return nil
	}
	for _, domain := range allowedDomains {
		d := strings.ToLower(strings.TrimSpace(domain))
		if d == "" {
			continue
		}
		if h == d || strings.HasSuffix(h, "."+d) {
			return nil
		}
	}
	return fmt.Errorf("目标域名 %s 不在允许列表", host)
}

func sanitizeCookieDomain(domain string) (string, error) {
	d := strings.ToLower(strings.TrimSpace(domain))
	d = strings.TrimPrefix(d, ".")
	if d == "" {
		return "", fmt.Errorf("cookie 域名为空")
	}
	if !cookieDomainPattern.MatchString(d) {
		return "", fmt.Errorf("cookie 域名格式非法")
	}
	return d, nil
}

func cookieFilePath(cookieStoreDir, domain string) (string, error) {
	d, err := sanitizeCookieDomain(domain)
	if err != nil {
		return "", err
	}
	return filepath.Join(cookieStoreDir, d+".json"), nil
}

func loadDomainCookies(page interface {
	SetCookies(cookies []*proto.NetworkCookieParam) error
}, domain, cookieStoreDir string, targetURL *url.URL) error {
	cookiePath, err := cookieFilePath(cookieStoreDir, domain)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(cookiePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取 cookie 文件失败: %w", err)
	}

	var cookies []*proto.NetworkCookieParam
	if err := json.Unmarshal(data, &cookies); err != nil {
		return fmt.Errorf("解析 cookie 文件失败: %w", err)
	}
	for _, c := range cookies {
		if c.URL == "" {
			c.URL = targetURL.Scheme + "://" + targetURL.Host
		}
	}
	if len(cookies) == 0 {
		return nil
	}
	if err := page.SetCookies(cookies); err != nil {
		return fmt.Errorf("注入 cookie 失败: %w", err)
	}
	return nil
}

func saveDomainCookies(page interface {
	Cookies(urls []string) ([]*proto.NetworkCookie, error)
}, domain, cookieStoreDir string, targetURL *url.URL) error {
	safeDomain, err := sanitizeCookieDomain(domain)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cookieStoreDir, 0o755); err != nil {
		return fmt.Errorf("创建 cookie 目录失败: %w", err)
	}

	cookies, err := page.Cookies([]string{targetURL.String()})
	if err != nil {
		return fmt.Errorf("读取页面 cookie 失败: %w", err)
	}

	params := make([]*proto.NetworkCookieParam, 0, len(cookies))
	for _, c := range cookies {
		domainValue := c.Domain
		if domainValue == "" {
			domainValue = safeDomain
		}
		params = append(params, &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   domainValue,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
			SameSite: c.SameSite,
			Expires:  c.Expires,
		})
	}

	data, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 cookie 失败: %w", err)
	}
	cookiePath := filepath.Join(cookieStoreDir, safeDomain+".json")
	if err := os.WriteFile(cookiePath, data, 0o600); err != nil {
		return fmt.Errorf("写入 cookie 文件失败: %w", err)
	}
	return nil
}

func pageTitle(page *rod.Page) (string, error) {
	el, err := page.Element("title")
	if err != nil {
		return "", err
	}
	return el.Text()
}

func extractMainContent(page *rod.Page, waitSelectorTimeoutSec int) (string, string, error) {
	selectors := []string{"#js_content", "article", "main", "[role='main']", "body"}
	timeout := time.Duration(waitSelectorTimeoutSec) * time.Second
	for _, selector := range selectors {
		el, err := page.Timeout(timeout).Element(selector)
		if err != nil {
			continue
		}
		htmlContent, err := el.HTML()
		if err != nil {
			continue
		}
		textContent, err := el.Text()
		if err != nil {
			continue
		}
		if strings.TrimSpace(textContent) == "" {
			continue
		}
		return htmlContent, textContent, nil
	}
	return "", "", fmt.Errorf("未能提取页面正文")
}

func NewBrowserBrowseTool(ctx context.Context) (tool.BaseTool, error) {
	cfg := config.Get().Tools.LocalBrowser
	if cfg.Enabled == nil || !*cfg.Enabled {
		return nil, fmt.Errorf("browser_browse tool is not enabled in config")
	}
	return utils.InferTool(
		"browser_browse",
		"使用本地 go-rod + stealth 浏览器访问动态网页，支持 read/screenshot、wait_selector 与 cookie 域复用。",
		BrowserBrowse,
	)
}

