// Package fetcher 负责 RSS 抓取、正文清洗与文章入库
package fetcher

import (
	"bytes"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

// MaxContentBytes 单篇正文最大字节数，超出截断（防极端大文章）
const MaxContentBytes = 500 * 1024

// SummaryRunes 摘要保留的纯文本字数
const SummaryRunes = 100

// 白名单标签与属性：只保留阅读必需的结构化标签
var sanitizerPolicy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements(
		"p", "h2", "h3", "strong", "em",
		"ul", "li", "blockquote", "a", "pre", "code", "img",
	)
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "alt").OnElements("img")
	// 允许相对链接，Sanitize 之后统一由 resolveURLs 转绝对地址
	p.RequireParseableURLs(false)
	return p
}()

// SanitizeContent 清洗正文：
//  1. bluemonday 白名单过滤（script/事件属性等全部移除）
//  2. a[href] / img[src] 相对链接转绝对链接（base 为站点地址）
//  3. 超过 MaxContentBytes 截断到最后一个完整标签边界
func SanitizeContent(rawHTML string, base string) string {
	cleaned := sanitizerPolicy.Sanitize(rawHTML)
	cleaned = resolveURLs(cleaned, base)
	return truncateContent(cleaned, MaxContentBytes)
}

// ExtractSummary 先白名单清洗，再去掉全部 HTML 标签，折叠空白，取前 100 个字（按 rune）
func ExtractSummary(rawHTML string) string {
	plain := htmlToPlainText(sanitizerPolicy.Sanitize(rawHTML))
	plain = collapseWhitespace(plain)
	return truncateRunes(plain, SummaryRunes)
}

// resolveURLs 将 HTML 中 a/img 的相对 URL 解析为绝对 URL
func resolveURLs(content, base string) string {
	baseURL, err := url.Parse(base)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return content // 没有可用的基准地址，原样返回
	}

	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return content
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			rewriteAttr(n, "href", baseURL)
			rewriteAttr(n, "src", baseURL)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return content
	}
	// html.Parse 会补全 html/head/body，只取 body 内容
	rendered := buf.String()
	if start := strings.Index(rendered, "<body>"); start >= 0 {
		start += len("<body>")
		end := strings.LastIndex(rendered, "</body>")
		if end > start {
			return rendered[start:end]
		}
	}
	return rendered
}

// rewriteAttr 将节点指定属性的相对地址解析为绝对地址，非法地址移除该属性
func rewriteAttr(n *html.Node, attrName string, baseURL *url.URL) {
	for i, attr := range n.Attr {
		if attr.Key != attrName || strings.TrimSpace(attr.Val) == "" {
			continue
		}
		u, err := url.Parse(strings.TrimSpace(attr.Val))
		if err != nil {
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			return
		}
		switch u.Scheme {
		case "", "http", "https":
			n.Attr[i].Val = baseURL.ResolveReference(u).String()
		default:
			// javascript:、data: 等非 http(s) 协议一律移除
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			return
		}
	}
}

// truncateContent 在不超过 limit 字节的前提下，截断到最后一个 '>' 边界
func truncateContent(content string, limit int) string {
	if len(content) <= limit {
		return content
	}
	cut := strings.LastIndex(content[:limit], ">")
	if cut <= 0 {
		return content[:limit]
	}
	return content[:cut+1]
}

// htmlToPlainText 遍历 DOM 提取纯文本
func htmlToPlainText(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		// 块级元素与 br 补空格，避免文字粘连
		if n.Type == html.ElementNode {
			switch n.Data {
			case "p", "br", "div", "li", "h1", "h2", "h3", "h4", "tr", "blockquote", "pre":
				sb.WriteByte(' ')
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return sb.String()
}

// collapseWhitespace 将连续空白（含换行/制表符）折叠为单个空格并去除首尾空白
func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n])
}
