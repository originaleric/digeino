package ocr

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const maxImageRedirects = 10

// newSecureImageHTTPClient 创建用于图片下载的 HTTP 客户端。
// DNS 仅解析一次，校验通过后以 ip:port 直连，避免校验与 Dial 之间的重绑定窗口。
// HTTPS 的 SNI/证书校验仍由 net/http 按请求 URL 主机名处理。
func newSecureImageHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	dialer := &net.Dialer{Timeout: timeout}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := resolveValidatedHostIPs(ctx, host)
			if err != nil {
				return nil, err
			}
			var lastErr error
			for _, ip := range ips {
				ipAddr := net.JoinHostPort(ip.String(), port)
				conn, dialErr := dialer.DialContext(ctx, network, ipAddr)
				if dialErr == nil {
					return conn, nil
				}
				lastErr = dialErr
			}
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, fmt.Errorf("no addresses to dial for %q", host)
		},
		DisableKeepAlives: true,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxImageRedirects {
				return fmt.Errorf("too many redirects")
			}
			return validateImageURL(req.URL.String())
		},
	}
}

// resolveValidatedHostIPs 解析主机并校验全部 IP（仅解析一次，供直连使用）。
func resolveValidatedHostIPs(ctx context.Context, host string) ([]net.IP, error) {
	cfg := ocrCfg()
	blockPrivate := cfg.BlockPrivateNetworks == nil || *cfg.BlockPrivateNetworks

	if ip := net.ParseIP(host); ip != nil {
		if blockPrivate {
			if err := checkIPNotPrivate(ip); err != nil {
				return nil, err
			}
		}
		return []net.IP{ip}, nil
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, newOCRError(CodeURLNotAllowed, "dns lookup failed: "+err.Error())
	}
	if len(addrs) == 0 {
		return nil, newOCRError(CodeURLNotAllowed, "dns lookup returned no addresses")
	}
	ips := make([]net.IP, 0, len(addrs))
	for _, a := range addrs {
		if blockPrivate {
			if err := checkIPNotPrivate(a.IP); err != nil {
				return nil, err
			}
		}
		ips = append(ips, a.IP)
	}
	return ips, nil
}

func checkIPNotPrivate(ip net.IP) error {
	if ip == nil {
		return newOCRError(CodeURLNotAllowed, "invalid resolved IP")
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return newOCRError(CodeURLNotAllowed, "resolved to private or loopback address")
	}
	return nil
}
