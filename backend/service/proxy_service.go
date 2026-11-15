package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MaxImageSize 代理图片大小上限 5MB
const MaxImageSize = 5 * 1024 * 1024

// browserUA 伪装成手机浏览器 UA：部分 CDN（知乎/微博/小红书等）会拒绝非浏览器 UA 的抓取
const browserUA = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) " +
	"AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"

// ProxyService 正文图片代理。
// 域名不做白名单（RSS 图片普遍托管在第三方 CDN），但保留 SSRF 防护：
// 仅 http(s)、解析 DNS 后拦截内网/回环/链路本地地址、仅放行图片类型。
type ProxyService struct {
	client *http.Client
}

func NewProxyService() *ProxyService {
	p := &ProxyService{}
	// 不自动跟随重定向，每一跳都重新做内网校验
	p.client = &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return p
}

// ImageResult 拉取结果
type ImageResult struct {
	ContentType string
	Body        []byte
}

// Fetch 校验并拉取图片，最多跟随 3 次合法重定向
func (p *ProxyService) Fetch(ctx context.Context, rawURL string) (*ImageResult, error) {
	current := strings.TrimSpace(rawURL)
	const maxRedirects = 3
	for hop := 0; hop <= maxRedirects; hop++ {
		target, err := validatePublicURL(current)
		if err != nil {
			return nil, err
		}

		resp, err := p.fetchWithHotlinkFallback(ctx, target)
		if err != nil {
			return nil, err
		}

		// 重定向：关闭 body 后校验下一跳
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			loc := resp.Header.Get("Location")
			resp.Body.Close()
			if loc == "" {
				return nil, fmt.Errorf("重定向地址缺失")
			}
			// 相对地址基于当前地址解析
			base, _ := url.Parse(current)
			if ref, err := url.Parse(loc); err == nil && base != nil {
				current = base.ResolveReference(ref).String()
				continue
			}
			return nil, fmt.Errorf("重定向地址非法")
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("源站返回状态码 %d", resp.StatusCode)
		}
		ct := resp.Header.Get("Content-Type")
		if idx := strings.IndexByte(ct, ';'); idx >= 0 {
			ct = ct[:idx]
		}
		ct = strings.TrimSpace(strings.ToLower(ct))
		if ct != "" && !strings.HasPrefix(ct, "image/") {
			resp.Body.Close()
			return nil, fmt.Errorf("目标非图片类型: %s", ct)
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, MaxImageSize+1))
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if len(body) > MaxImageSize {
			return nil, fmt.Errorf("图片超过大小上限")
		}

		if ct == "" {
			ct = http.DetectContentType(body)
			if !strings.HasPrefix(ct, "image/") {
				return nil, fmt.Errorf("目标内容非图片")
			}
		}
		return &ImageResult{ContentType: ct, Body: body}, nil
	}
	return nil, fmt.Errorf("重定向次数过多")
}

// fetchWithHotlinkFallback 防盗链兜底：
// 先以浏览器 UA、不带 Referer 请求（多数站点允许空 Referer）；
// 若被拒绝（401/403/404），再带「图片域名 Referer」重试一次，
// 以满足「只允许本站引用」的策略。
func (p *ProxyService) fetchWithHotlinkFallback(ctx context.Context, target *url.URL) (*http.Response, error) {
	resp, err := p.doGet(ctx, target, "")
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		resp.Body.Close()
		return p.doGet(ctx, target, target.Scheme+"://"+target.Host+"/")
	}
	return resp, nil
}

func (p *ProxyService) doGet(ctx context.Context, target *url.URL, referer string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	return p.client.Do(req)
}

// validatePublicURL 校验协议为公网 http(s)，并解析 DNS 确认目标不是内网地址
func validatePublicURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("图片地址非法")
	}

	host := strings.ToLower(u.Hostname())
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("目标地址被禁止")
		}
		return u, nil
	}

	// 域名：解析后逐个 IP 校验，防 DNS rebinding 到内网
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("域名解析失败")
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("目标地址被禁止")
		}
	}
	return u, nil
}

// isBlockedIP 拦截回环/私有/链路本地/未指定/组播地址
func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast()
}
