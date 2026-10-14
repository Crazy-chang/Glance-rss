// Package fetcher favicon 下载：优先 RSS 自带图标，其次站点 /favicon.ico
package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// FaviconDir favicon 本地存放目录（相对后端启动目录）
const FaviconDir = "data/favicons"

const maxFaviconBytes = 100 * 1024

// downloadFavicon 依次尝试 RSS feed 自带图标、站点 /favicon.ico，
// 成功后保存到 data/favicons/<sourceID>.<ext>，返回相对路径
func downloadFavicon(ctx context.Context, client *http.Client, sourceID int64, siteURL, feedImage string) (string, error) {
	var candidates []string
	if feedImage != "" {
		candidates = append(candidates, feedImage)
	}
	if siteURL != "" {
		if u, err := url.Parse(siteURL); err == nil && u.Host != "" {
			candidates = append(candidates, u.Scheme+"://"+u.Host+"/favicon.ico")
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("无可用 favicon 地址")
	}

	if err := os.MkdirAll(FaviconDir, 0o755); err != nil {
		return "", err
	}

	var lastErr error
	for _, candidate := range candidates {
		ext := faviconExt(candidate)
		savePath := filepath.Join(FaviconDir, fmt.Sprintf("%d%s", sourceID, ext))

		if err := fetchAndSaveImage(ctx, client, candidate, savePath); err != nil {
			lastErr = err
			continue
		}
		return savePath, nil
	}
	return "", lastErr
}

func fetchAndSaveImage(ctx context.Context, client *http.Client, imageURL, savePath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("favicon 状态码 %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "image/") {
		return fmt.Errorf("favicon 非图片类型: %s", ct)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFaviconBytes+1))
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return fmt.Errorf("favicon 内容为空")
	}
	if len(body) > maxFaviconBytes {
		return fmt.Errorf("favicon 超过 100KB")
	}

	tmpPath := savePath + ".tmp"
	if err := os.WriteFile(tmpPath, body, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, savePath)
}

func faviconExt(imageURL string) string {
	// 去掉 query 再取后缀
	if i := strings.IndexByte(imageURL, '?'); i >= 0 {
		imageURL = imageURL[:i]
	}
	switch strings.ToLower(filepath.Ext(imageURL)) {
	case ".png":
		return ".png"
	case ".jpg", ".jpeg":
		return ".jpg"
	case ".gif":
		return ".gif"
	case ".svg":
		return ".svg"
	case ".webp":
		return ".webp"
	default:
		return ".ico"
	}
}
