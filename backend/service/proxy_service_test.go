package service

import (
	"strings"
	"testing"
)

func TestValidatePublicURL(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"正常 https 公网 IP", "https://1.1.1.1/a.png", false},
		{"正常 http 公网 IP", "http://8.8.8.8/x.jpg", false},
		{"回环 127.0.0.1", "http://127.0.0.1/x.png", true},
		{"回环 localhost 之外的 127 段", "http://127.1.2.3/x.png", true},
		{"10 段私网", "http://10.0.0.5/x.png", true},
		{"172.16 段私网", "http://172.16.0.1/x.png", true},
		{"172.31 段私网", "http://172.31.255.254/x.png", true},
		{"172.32 公网", "http://172.32.0.1/x.png", false},
		{"192.168 段私网", "http://192.168.1.1/x.png", true},
		{"链路本地", "http://169.254.1.1/x.png", true},
		{"IPv6 回环", "http://[::1]/x.png", true},
		{"ftp 协议", "ftp://1.1.1.1/x.png", true},
		{"javascript 协议", "javascript:alert(1)", true},
		{"空地址", "", true},
		{"无 host", "http:///x.png", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := validatePublicURL(c.raw)
			if (err != nil) != c.wantErr {
				t.Errorf("validatePublicURL(%q) err=%v, wantErr=%v", c.raw, err, c.wantErr)
			}
		})
	}
}

func TestRewriteImageProxy(t *testing.T) {
	prefix := "http://localhost:8080/api/img-proxy?url="
	content := `<p><img src="https://cdn.example.com/a.webp" alt="x"></p>
		<a href="https://example.com/post">链接不被改写</a>`
	got := RewriteImageProxy(content, prefix)

	if !strings.Contains(got, prefix) {
		t.Errorf("图片地址未被改写为代理地址: %s", got)
	}
	if strings.Contains(got, `src="https://cdn.example.com/a.webp"`) {
		t.Errorf("原始图片地址仍存在: %s", got)
	}
	if !strings.Contains(got, `href="https://example.com/post"`) {
		t.Errorf("普通链接被误改写: %s", got)
	}
}
