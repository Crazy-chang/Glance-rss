package fetcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"

	"glance-wx/backend/model"
)

const ua = "GlanceRSS/1.0 (+https://github.com/glance-wx; RSS Reader)"

// Fetcher RSS 抓取器，scheduler 与手动添加/刷新共用
type Fetcher struct {
	db     *gorm.DB
	client *http.Client
}

// New 创建抓取器，timeout 为单源整体抓取超时
func New(db *gorm.DB, timeout time.Duration) *Fetcher {
	return &Fetcher{
		db: db,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// FeedMeta RSS 校验成功后提取的订阅源元数据
type FeedMeta struct {
	Title       string
	Link        string
	Description string
	ImageURL    string
}

// newParser 创建带自定义 UA 与超时的 gofeed 解析器
func (f *Fetcher) newParser() *gofeed.Parser {
	fp := gofeed.NewParser()
	fp.Client = &http.Client{
		Timeout: f.client.Timeout,
		Transport: &userAgentTransport{
			base: http.DefaultTransport,
			ua:   ua,
		},
	}
	return fp
}

// ValidateFeed 同步解析 RSS 地址，校验是否可用并返回订阅源元数据
func (f *Fetcher) ValidateFeed(ctx context.Context, feedURL string) (*FeedMeta, error) {
	parseCtx, cancel := context.WithTimeout(ctx, f.client.Timeout)
	defer cancel()

	feed, err := f.newParser().ParseURLWithContext(feedURL, parseCtx)
	if err != nil {
		return nil, err
	}
	meta := &FeedMeta{
		Title:       strings.TrimSpace(feed.Title),
		Link:        strings.TrimSpace(feed.Link),
		Description: strings.TrimSpace(feed.Description),
	}
	if feed.Image != nil {
		meta.ImageURL = strings.TrimSpace(feed.Image.URL)
	}
	return meta, nil
}

// FetchSource 抓取单个订阅源：解析 → 清洗入库 → 裁剪条数 → 更新 favicon/抓取状态。
// 任何失败都会写入 sources.last_error 并返回错误。
func (f *Fetcher) FetchSource(ctx context.Context, sourceID int64) error {
	var src model.Source
	if err := f.db.First(&src, sourceID).Error; err != nil {
		return fmt.Errorf("订阅源不存在: %w", err)
	}

	parseCtx, cancel := context.WithTimeout(ctx, f.client.Timeout)
	defer cancel()

	feed, err := f.newParser().ParseURLWithContext(src.FeedURL, parseCtx)
	if err != nil {
		f.markError(&src, fmt.Sprintf("抓取失败: %s", truncateErr(err.Error())))
		return err
	}

	// 用 feed 信息补全订阅源元数据（首次抓取时这些字段通常为空或为 URL）
	f.fillSourceMeta(&src, feed)

	// favicon（失败不影响文章入库）
	if src.FaviconPath == "" {
		feedImage := ""
		if feed.Image != nil {
			feedImage = feed.Image.URL
		}
		if path, err := downloadFavicon(ctx, f.client, sourceID, src.SiteURL, feedImage); err == nil {
			src.FaviconPath = path
		} else {
			slog.Debug("favicon 下载失败", "source", src.Title, "err", err)
		}
	}

	base := firstNonEmpty(src.SiteURL, feed.Link, src.FeedURL)
	newCount := 0
	for _, item := range feed.Items {
		created, err := f.saveItem(&src, item, base)
		if err != nil {
			slog.Warn("文章入库失败", "source", src.Title, "guid", item.GUID, "err", err)
			continue
		}
		if created {
			newCount++
		}
	}

	// 按属主设置裁剪单源最大条数
	if err := f.trimOldest(src.ID, src.UserID); err != nil {
		slog.Warn("裁剪旧文章失败", "source", src.Title, "err", err)
	}

	now := time.Now().UTC()
	src.LastFetchAt = &now
	src.LastError = ""
	if err := f.db.Save(&src).Error; err != nil {
		return err
	}

	slog.Info("订阅源抓取完成", "source", src.Title, "new", newCount, "items", len(feed.Items))
	return nil
}

// fillSourceMeta 用 feed 信息补全标题/站点/描述
func (f *Fetcher) fillSourceMeta(src *model.Source, feed *gofeed.Feed) {
	changed := false
	if strings.TrimSpace(feed.Title) != "" && (src.Title == "" || src.Title == src.FeedURL) {
		src.Title = strings.TrimSpace(feed.Title)
		changed = true
	}
	if strings.TrimSpace(feed.Link) != "" && src.SiteURL == "" {
		src.SiteURL = strings.TrimSpace(feed.Link)
		changed = true
	}
	if strings.TrimSpace(feed.Description) != "" && src.Description == "" {
		desc := strings.TrimSpace(feed.Description)
		if len([]rune(desc)) > 200 {
			desc = string([]rune(desc)[:200])
		}
		src.Description = desc
		changed = true
	}
	if changed {
		f.db.Model(src).Updates(map[string]any{
			"title":       src.Title,
			"site_url":    src.SiteURL,
			"description": src.Description,
		})
	}
}

// saveItem 清洗并写入单篇文章，返回是否新创建
func (f *Fetcher) saveItem(src *model.Source, item *gofeed.Item, base string) (bool, error) {
	guid := strings.TrimSpace(item.GUID)
	if guid == "" {
		guid = strings.TrimSpace(item.Link) // GUID 缺失回退 link
	}
	if guid == "" {
		return false, nil // 两者皆空无法去重，跳过
	}

	// 去重
	var existing int64
	f.db.Model(&model.Article{}).
		Where("source_id = ? AND guid = ?", src.ID, guid).
		Count(&existing)
	if existing > 0 {
		return false, nil
	}

	rawContent := firstNonEmpty(item.Content, item.Description, "")
	content := SanitizeContent(rawContent, base)

	summarySource := firstNonEmpty(item.Description, item.Content)
	summary := ExtractSummary(summarySource)

	publishedAt := time.Now().UTC()
	if item.PublishedParsed != nil {
		publishedAt = item.PublishedParsed.UTC()
	}

	article := model.Article{
		SourceID:    src.ID,
		GUID:        guid,
		Link:        strings.TrimSpace(item.Link),
		Title:       strings.TrimSpace(firstNonEmpty(item.Title, "(无标题)")),
		Author:      joinAuthors(item.Authors),
		Summary:     summary,
		Content:     content,
		PublishedAt: publishedAt,
		FetchedAt:   time.Now().UTC(),
	}
	if err := f.db.Create(&article).Error; err != nil {
		return false, err
	}
	return true, nil
}

// trimOldest 超过属主设置的单源最大条数时，删除最旧的文章及其状态记录
func (f *Fetcher) trimOldest(sourceID, userID int64) error {
	var setting model.Setting
	maxItems := model.DefaultMaxItems
	if err := f.db.Where("user_id = ?", userID).First(&setting).Error; err == nil && setting.MaxItems > 0 {
		maxItems = setting.MaxItems
	}

	return f.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DELETE FROM articles
			WHERE source_id = ?
			  AND id NOT IN (
				SELECT id FROM articles
				WHERE source_id = ?
				ORDER BY published_at DESC, id DESC
				LIMIT ?
			  )`, sourceID, sourceID, maxItems).Error; err != nil {
			return err
		}
		// 清理已不存在文章的孤儿状态记录
		return tx.Exec(`
			DELETE FROM user_article_states
			WHERE user_id = ?
			  AND article_id NOT IN (SELECT id FROM articles)`, userID).Error
	})
}

func (f *Fetcher) markError(src *model.Source, msg string) {
	now := time.Now().UTC()
	src.LastFetchAt = &now
	src.LastError = msg
	f.db.Model(src).Updates(map[string]any{
		"last_fetch_at": now,
		"last_error":    msg,
	})
}

type userAgentTransport struct {
	base http.RoundTripper
	ua   string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.base == nil {
		return nil, errors.New("no transport")
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("User-Agent", t.ua)
	return t.base.RoundTrip(clone)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func joinAuthors(authors []*gofeed.Person) string {
	if len(authors) == 0 {
		return ""
	}
	names := make([]string, 0, len(authors))
	for _, a := range authors {
		if a != nil && strings.TrimSpace(a.Name) != "" {
			names = append(names, strings.TrimSpace(a.Name))
		}
	}
	return strings.Join(names, ", ")
}

func truncateErr(msg string) string {
	if len(msg) > 280 {
		return msg[:280]
	}
	return msg
}
