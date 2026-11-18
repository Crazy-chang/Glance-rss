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
// 分类顺序取分类表 sort_order，分类内顺序取推荐源 sort_order
func (d *DiscoverService) List(userID int64) (*DiscoverResult, error) {
	var cats []model.RecommendCategory
	if err := d.db.Order("sort_order ASC, id ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	var all []model.RecommendSource
	if err := d.db.Order("sort_order ASC, id ASC").Find(&all).Error; err != nil {
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
			Subscribed:  subscribed[normalizeFeedURL(r.FeedURL)],
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

	// 分类分组：先按分类表顺序建空组，再填充源
	catIndex := make(map[int64]int, len(cats))
	orphanIndex := make(map[string]int, 2)
	for _, c := range cats {
		catIndex[c.ID] = len(result.Categories)
		result.Categories = append(result.Categories, DiscoverCategory{
			Name:    c.Name,
			Sources: []DiscoverSource{},
		})
	}
	for _, r := range all {
		if idx, ok := catIndex[r.CategoryID]; ok {
			result.Categories[idx].Sources = append(result.Categories[idx].Sources, toDTO(r))
			continue
		}
		// 兜底：未关联分类的旧数据按分类名追加到末尾
		name := strings.TrimSpace(r.Category)
		if name == "" {
			name = "未分类"
		}
		idx, ok := orphanIndex[name]
		if !ok {
			idx = len(result.Categories)
			orphanIndex[name] = idx
			result.Categories = append(result.Categories, DiscoverCategory{
				Name:    name,
				Sources: []DiscoverSource{},
			})
		}
		result.Categories[idx].Sources = append(result.Categories[idx].Sources, toDTO(r))
	}

	// 过滤空分类（后台可能新建了尚未添加源的分类）
	filled := make([]DiscoverCategory, 0, len(result.Categories))
	for _, c := range result.Categories {
		if len(c.Sources) > 0 {
			filled = append(filled, c)
		}
	}
	result.Categories = filled
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

// Unsubscribe 取消订阅推荐源：按 feed_url 匹配当前用户的订阅源后删除（级联删除文章）
func (d *DiscoverService) Unsubscribe(userID, recommendID int64) error {
	var rec model.RecommendSource
	if err := d.db.First(&rec, recommendID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSourceMissing
		}
		return err
	}
	var sources []model.Source
	if err := d.db.Where("user_id = ?", userID).Find(&sources).Error; err != nil {
		return err
	}
	target := normalizeFeedURL(rec.FeedURL)
	for _, s := range sources {
		if normalizeFeedURL(s.FeedURL) == target {
			return d.sources.Delete(userID, s.ID)
		}
	}
	return ErrSourceMissing
}

// normalizeFeedURL 规整 feed 地址，用于判断两个地址是否指向同一个订阅源：
// 忽略协议（http/https）、host 大小写、www. 前缀、路径尾部斜杠，保留路径与查询参数。
// 保留查询参数是因为 RSSHub 等服务的不同源靠 ?type=xx 区分。
func normalizeFeedURL(f string) string {
	f = strings.TrimSpace(f)
	u, err := url.Parse(f)
	if err != nil || u.Host == "" {
		return strings.ToLower(f)
	}
	host := strings.TrimPrefix(strings.ToLower(u.Host), "www.")
	path := strings.TrimRight(u.Path, "/")
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		return host + path + "?" + u.RawQuery
	}
	return host + path
}

// userFeedSet 取当前用户已订阅的 feed_url 集合（归一化后比较）
func (d *DiscoverService) userFeedSet(userID int64) (map[string]bool, error) {
	var feeds []string
	if err := d.db.Model(&model.Source{}).
		Where("user_id = ?", userID).
		Pluck("feed_url", &feeds).Error; err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(feeds))
	for _, f := range feeds {
		set[normalizeFeedURL(f)] = true
	}
	return set, nil
}
