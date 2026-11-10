package service

import (
	"errors"
	"net/url"
	"sort"
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
	// 分类顺序与分类内源顺序均以 seed 行序为准；数据库 id 仅作兜底
	var all []model.RecommendSource
	if err := d.db.Order("id ASC").Find(&all).Error; err != nil {
		return nil, err
	}
	catRank := make(map[string]int, 8)
	for i, name := range model.RecommendCategoryOrder() {
		catRank[name] = i
	}
	unknownRank := len(catRank)
	rankOf := func(category string) int {
		if r, ok := catRank[category]; ok {
			return r
		}
		return unknownRank
	}
	seedIndex := model.RecommendSeedIndex()
	unknownSeed := len(seedIndex)
	seedOf := func(feedURL string) int {
		if i, ok := seedIndex[feedURL]; ok {
			return i
		}
		return unknownSeed
	}
	sort.SliceStable(all, func(i, j int) bool {
		if ri, rj := rankOf(all[i].Category), rankOf(all[j].Category); ri != rj {
			return ri < rj
		}
		if si, sj := seedOf(all[i].FeedURL), seedOf(all[j].FeedURL); si != sj {
			return si < sj
		}
		return all[i].ID < all[j].ID
	})

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
	for _, s := range sources {
		if s.FeedURL == rec.FeedURL || normalizeFeedURL(s.FeedURL) == normalizeFeedURL(rec.FeedURL) {
			return d.sources.Delete(userID, s.ID)
		}
	}
	return ErrSourceMissing
}

// normalizeFeedURL 规整 feed 地址（小写 host + 去尾部斜杠路径），与 userFeedSet 的匹配口径一致
func normalizeFeedURL(f string) string {
	u, err := url.Parse(f)
	if err != nil {
		return f
	}
	return strings.ToLower(u.Host) + strings.TrimRight(u.Path, "/")
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
