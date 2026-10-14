package service

import (
	"errors"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"glance-wx/backend/model"
)

// DiscoverService 发现页（热门订阅 + 分类推荐）
type DiscoverService struct {
	db      *gorm.DB
	sources *SourceService
}

func NewDiscoverService(db *gorm.DB, sources *SourceService) *DiscoverService {
	return &DiscoverService{db: db, sources: sources}
}

// DiscoverSource 推荐源项，附带当前用户是否已订阅
type DiscoverSource struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	FeedURL     string `json:"feed_url"`
	SiteURL     string `json:"site_url"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Subscribed  bool   `json:"subscribed"`
}

// DiscoverCategory 分类区块
type DiscoverCategory struct {
	Name    string           `json:"name"`
	Sources []DiscoverSource `json:"sources"`
}

// DiscoverResult 发现页数据
type DiscoverResult struct {
	Hot        []DiscoverSource   `json:"hot"`
	Categories []DiscoverCategory `json:"categories"`
}

// List 返回热门订阅与分类推荐，并标记当前用户已订阅的源
func (d *DiscoverService) List(userID int64) (*DiscoverResult, error) {
	var all []model.RecommendSource
	if err := d.db.Order("category ASC, hot_rank ASC, id ASC").Find(&all).Error; err != nil {
		return nil, err
	}

	subscribed, err := d.userFeedSet(userID)
	if err != nil {
		return nil, err
	}

	toDTO := func(r model.RecommendSource) DiscoverSource {
		return DiscoverSource{
			ID:          r.ID,
			Title:       r.Title,
			FeedURL:     r.FeedURL,
			SiteURL:     r.SiteURL,
			Category:    r.Category,
			Description: r.Description,
			Subscribed:  subscribed[r.FeedURL],
		}
	}

	result := &DiscoverResult{
		Hot:        []DiscoverSource{},
		Categories: []DiscoverCategory{},
	}

	// 热门单独按 hot_rank 升序
	var hot []model.RecommendSource
	if err := d.db.Where("hot_rank > 0").Order("hot_rank ASC, id ASC").Find(&hot).Error; err != nil {
		return nil, err
	}
	for _, r := range hot {
		result.Hot = append(result.Hot, toDTO(r))
	}

	// 分类分组（保持查询出的分类首次出现顺序）
	catIndex := map[string]int{}
	for _, r := range all {
		idx, ok := catIndex[r.Category]
		if !ok {
			result.Categories = append(result.Categories, DiscoverCategory{
				Name:    r.Category,
				Sources: []DiscoverSource{},
			})
			idx = len(result.Categories) - 1
			catIndex[r.Category] = idx
		}
		result.Categories[idx].Sources = append(result.Categories[idx].Sources, toDTO(r))
	}
	return result, nil
}

// Subscribe 订阅推荐源：查重后创建并异步首次抓取
func (d *DiscoverService) Subscribe(userID, recommendID int64) (*SourceDTO, error) {
	var rec model.RecommendSource
	if err := d.db.First(&rec, recommendID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSourceMissing
		}
		return nil, err
	}
	return d.sources.CreateFromFeed(userID, rec.FeedURL, rec.Title, rec.SiteURL, rec.Description, rec.Category)
}

// userFeedSet 取当前用户已订阅的 feed_url 集合（规整 host 后比较，兼容尾部斜杠差异）
func (d *DiscoverService) userFeedSet(userID int64) (map[string]bool, error) {
	var feeds []string
	if err := d.db.Model(&model.Source{}).
		Where("user_id = ?", userID).
		Pluck("feed_url", &feeds).Error; err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(feeds))
	for _, f := range feeds {
		set[f] = true
		if u, err := url.Parse(f); err == nil {
			set[strings.ToLower(u.Host)+strings.TrimRight(u.Path, "/")] = true
		}
	}
	return set, nil
}
