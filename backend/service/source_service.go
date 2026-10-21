package service

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"glance-wx/backend/fetcher"
	"glance-wx/backend/model"
)

// SourceService 订阅源管理
type SourceService struct {
	db      *gorm.DB
	fetcher *fetcher.Fetcher
	baseURL string
}

func NewSourceService(db *gorm.DB, f *fetcher.Fetcher, baseURL string) *SourceService {
	return &SourceService{db: db, fetcher: f, baseURL: strings.TrimRight(baseURL, "/")}
}

// SourceDTO 订阅源列表项（含文章数与未读数）
type SourceDTO struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	FeedURL      string `json:"feed_url"`
	SiteURL      string `json:"site_url"`
	FaviconURL   string `json:"favicon_url"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	LastFetchAt  int64  `json:"last_fetch_at"`
	LastError    string `json:"last_error"`
	ArticleCount int64  `json:"article_count"`
	UnreadCount  int64  `json:"unread_count"`
}

type sourceCountRow struct {
	SourceID     int64 `gorm:"column:source_id"`
	ArticleCount int64 `gorm:"column:article_count"`
	UnreadCount  int64 `gorm:"column:unread_count"`
}

// List 当前用户的全部订阅源（新订阅在前）
func (s *SourceService) List(userID int64) ([]SourceDTO, error) {
	var sources []model.Source
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Find(&sources).Error; err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return []SourceDTO{}, nil
	}

	ids := make([]int64, 0, len(sources))
	for _, src := range sources {
		ids = append(ids, src.ID)
	}

	// 一次聚合查出文章总数与未读数
	var rows []sourceCountRow
	err := s.db.Table("articles a").
		Select(`a.source_id AS source_id,
			COUNT(*) AS article_count,
			SUM(CASE WHEN COALESCE(st.is_read, 0) = 0 THEN 1 ELSE 0 END) AS unread_count`).
		Joins("LEFT JOIN user_article_states st ON st.article_id = a.id AND st.user_id = ?", userID).
		Where("a.source_id IN ?", ids).
		Group("a.source_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[int64]sourceCountRow, len(rows))
	for _, r := range rows {
		counts[r.SourceID] = r
	}

	list := make([]SourceDTO, 0, len(sources))
	for _, src := range sources {
		dto := s.toDTO(src)
		if c, ok := counts[src.ID]; ok {
			dto.ArticleCount = c.ArticleCount
			dto.UnreadCount = c.UnreadCount
		}
		list = append(list, dto)
	}
	return list, nil
}

// Add 校验 RSS 地址 → 查重 → 创建订阅源 → 异步首次抓取
func (s *SourceService) Add(ctx context.Context, userID int64, feedURL string) (*SourceDTO, error) {
	feedURL = strings.TrimSpace(feedURL)
	if !isHTTPURL(feedURL) {
		return nil, ErrFeedInvalid
	}

	// 同一用户同一 feed 不可重复订阅
	var dup int64
	if err := s.db.Model(&model.Source{}).
		Where("user_id = ? AND feed_url = ?", userID, feedURL).
		Count(&dup).Error; err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, ErrFeedDuplicate
	}

	// 同步校验 RSS 可解析（接口直接给出失败原因）
	meta, err := s.fetcher.ValidateFeed(ctx, feedURL)
	if err != nil {
		return nil, ErrFeedInvalid
	}

	return s.CreateFromFeed(userID, feedURL,
		firstNonEmpty(meta.Title, feedURL), meta.Link, meta.Description, "")
}

// CreateFromFeed 用已校验的信息创建订阅源并异步首次抓取（手动添加与推荐订阅共用）
func (s *SourceService) CreateFromFeed(userID int64, feedURL, title, siteURL, description, category string) (*SourceDTO, error) {
	// 同一用户同一 feed 不可重复订阅
	var dup int64
	if err := s.db.Model(&model.Source{}).
		Where("user_id = ? AND feed_url = ?", userID, feedURL).
		Count(&dup).Error; err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, ErrFeedDuplicate
	}

	source := model.Source{
		UserID:      userID,
		Title:       truncateRunes(firstNonEmpty(title, feedURL), 200),
		FeedURL:     feedURL,
		SiteURL:     siteURL,
		Description: truncateRunes(description, 200),
		Category:    category,
	}
	if err := s.db.Create(&source).Error; err != nil {
		return nil, err
	}

	// 立即首次抓取（异步，不阻塞接口）
	go func(id int64) {
		fetchCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = s.fetcher.FetchSource(fetchCtx, id)
	}(source.ID)

	dto := s.toDTO(source)
	return &dto, nil
}

// Rename 更新订阅源标题（仅允许所属用户操作）
func (s *SourceService) Rename(userID, sourceID int64, title string) error {
	title = truncateRunes(title, 200)
	if title == "" {
		return ErrSourceTitleEmpty
	}
	res := s.db.Model(&model.Source{}).
		Where("id = ? AND user_id = ?", sourceID, userID).
		Update("title", title)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrSourceMissing
	}
	return nil
}

// Delete 删除订阅源，级联删除其文章与用户文章状态，并清理 favicon 文件
func (s *SourceService) Delete(userID, sourceID int64) error {
	var src model.Source
	err := s.db.Where("id = ? AND user_id = ?", sourceID, userID).Take(&src).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrSourceMissing
	}
	if err != nil {
		return err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 先收集该源文章 id，再删状态、文章、源
		var articleIDs []int64
		if err := tx.Model(&model.Article{}).
			Where("source_id = ?", sourceID).
			Pluck("id", &articleIDs).Error; err != nil {
			return err
		}
		if len(articleIDs) > 0 {
			if err := tx.Where("user_id = ? AND article_id IN ?", userID, articleIDs).
				Delete(&model.UserArticleState{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("source_id = ?", sourceID).
			Delete(&model.Article{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Source{}, sourceID).Error
	})
	if err != nil {
		return err
	}

	// favicon 文件清理失败不影响删除结果
	if src.FaviconPath != "" {
		if rmErr := os.Remove(src.FaviconPath); rmErr != nil && !os.IsNotExist(rmErr) {
			// 仅忽略，不返回错误
			_ = rmErr
		}
	}
	return nil
}

// GetOwned 取属于该用户的订阅源（供 favicon 接口校验归属）
func (s *SourceService) GetOwned(userID, sourceID int64) (*model.Source, error) {
	var src model.Source
	err := s.db.Where("id = ? AND user_id = ?", sourceID, userID).Take(&src).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSourceMissing
	}
	if err != nil {
		return nil, err
	}
	return &src, nil
}

func (s *SourceService) toDTO(src model.Source) SourceDTO {
	dto := SourceDTO{
		ID:          src.ID,
		Title:       src.Title,
		FeedURL:     src.FeedURL,
		SiteURL:     src.SiteURL,
		Category:    src.Category,
		Description: src.Description,
		LastError:   src.LastError,
	}
	if src.LastFetchAt != nil {
		dto.LastFetchAt = src.LastFetchAt.Unix()
	}
	if src.FaviconPath != "" {
		dto.FaviconURL = s.baseURL + "/api/favicon/" + strconv.FormatInt(src.ID, 10)
	}
	return dto
}

func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func truncateRunes(s string, n int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= n {
		return strings.TrimSpace(s)
	}
	return string(runes[:n])
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
