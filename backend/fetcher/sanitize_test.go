package fetcher

import (
	"strings"
	"testing"
)

func TestSanitizeContent_StripsDangerousTags(t *testing.T) {
	raw := `<p>正常段落</p><script>alert('xss')</script><a href="https://a.com" onclick="steal()">链接</a>`
	got := SanitizeContent(raw, "https://example.com/feed.xml")

	if strings.Contains(got, "<script") || strings.Contains(got, "alert") {
		t.Errorf("script 标签未被移除: %s", got)
	}
	if strings.Contains(got, "onclick") || strings.Contains(got, "steal") {
		t.Errorf("事件属性未被移除: %s", got)
	}
	if !strings.Contains(got, "正常段落") || !strings.Contains(got, "链接") {
		t.Errorf("合法内容被误删: %s", got)
	}
}

func TestSanitizeContent_KeepsWhitelist(t *testing.T) {
	raw := `<h2>标题</h2><p>文字 <strong>加粗</strong> <em>斜体</em></p>
		<ul><li>一</li><li>二</li></ul><blockquote>引用</blockquote>
		<pre><code>code()</code></pre><img src="/a.png" alt="图">`
	got := SanitizeContent(raw, "https://example.com/feed.xml")

	for _, tag := range []string{"<h2>", "<p>", "<strong>", "<em>", "<ul>", "<li>", "<blockquote>", "<pre>", "<code>", "<img"} {
		if !strings.Contains(got, tag) {
			t.Errorf("白名单标签 %s 被误删: %s", tag, got)
		}
	}
}

func TestSanitizeContent_DropsNonWhitelist(t *testing.T) {
	raw := `<p>x</p><h1>大标题</h1><ol><li>有序</li></ol><table><tr><td>表格</td></tr></table>`
	got := SanitizeContent(raw, "https://example.com/")

	if strings.Contains(got, "<h1") {
		t.Errorf("h1 应被移除: %s", got)
	}
	if strings.Contains(got, "<ol") || strings.Contains(got, "<table") || strings.Contains(got, "<td") {
		t.Errorf("非白名单标签应被移除: %s", got)
	}
	if !strings.Contains(got, "大标题") && !strings.Contains(got, "有序") && !strings.Contains(got, "表格") {
		t.Errorf("标签内文本应保留")
	}
}

func TestResolveURLs_RelativeToAbsolute(t *testing.T) {
	raw := `<a href="/post/1">文章</a><img src="../img/a.png"><a href="https://other.com/x">外链</a>`
	got := SanitizeContent(raw, "https://example.com/blog/feed.xml")

	if !strings.Contains(got, `href="https://example.com/post/1"`) {
		t.Errorf("绝对路径未正确解析: %s", got)
	}
	if !strings.Contains(got, `src="https://example.com/img/a.png"`) {
		t.Errorf("相对路径未正确解析: %s", got)
	}
	if !strings.Contains(got, `href="https://other.com/x"`) {
		t.Errorf("绝对外链被改写: %s", got)
	}
}

func TestResolveURLs_DropJavascriptScheme(t *testing.T) {
	raw := `<a href="javascript:alert(1)">x</a><a href="mailto:a@b.com">y</a>`
	got := SanitizeContent(raw, "https://example.com/")

	if strings.Contains(got, "javascript:") {
		t.Errorf("javascript 协议未移除: %s", got)
	}
	if strings.Contains(got, "mailto:") {
		t.Errorf("非 http(s) 协议未移除: %s", got)
	}
}

func TestSanitizeContent_Truncate500KB(t *testing.T) {
	// 构造约 600KB 的合法段落
	var sb strings.Builder
	sb.WriteString("<p>")
	for sb.Len() < 600*1024 {
		sb.WriteString("一二三四五六七八九十")
	}
	sb.WriteString("</p>")
	raw := sb.String()

	got := SanitizeContent(raw, "https://example.com/")
	if len(got) > MaxContentBytes {
		t.Errorf("正文超过 500KB 上限，实际 %d 字节", len(got))
	}
	// 截断后不应出现半个标签（以 > 结尾或内容本身未超限）
	if len(got) >= MaxContentBytes && !strings.HasSuffix(got, ">") {
		t.Errorf("截断未落在标签边界，结尾: %q", got[len(got)-20:])
	}
}

func TestSanitizeContent_BadBaseFallback(t *testing.T) {
	raw := `<a href="/x">x</a>`
	// 非法 base 不应导致内容丢失
	got := SanitizeContent(raw, "://not-a-url")
	if !strings.Contains(got, "/x") {
		t.Errorf("base 不可用时应原样保留: %s", got)
	}
}

func TestExtractSummary(t *testing.T) {
	raw := `<p>  第一段内容，<b>加粗</b>文字  </p>
		<script>bad()</script>
		<p>第二段</p>`
	got := ExtractSummary(raw)

	if strings.Contains(got, "<") || strings.Contains(got, "bad()") {
		t.Errorf("摘要不应包含标签或脚本: %q", got)
	}
	if !strings.HasPrefix(got, "第一段内容") {
		t.Errorf("摘要应以正文开头: %q", got)
	}
	if strings.Contains(got, "  ") {
		t.Errorf("空白未折叠: %q", got)
	}
}

func TestExtractSummary_100Runes(t *testing.T) {
	raw := "<p>" + strings.Repeat("字", 250) + "</p>"
	got := ExtractSummary(raw)
	runeCount := len([]rune(got))
	if runeCount != SummaryRunes {
		t.Errorf("摘要应为 %d 字，实际 %d 字", SummaryRunes, runeCount)
	}
}
