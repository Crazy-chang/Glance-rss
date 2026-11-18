package service

import (
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"glance-wx/backend/model"
)

// AdminRecommendService 后台推荐分类与推荐源管理
type AdminRecommendService struct {
	db *gorm.DB
}

func NewAdminRecommendService(db *gorm.DB) *AdminRecommendService {
	return &AdminRecommendService{db: db}
}

// ---------- 分类 ----------

// AdminCategoryDTO 后台分类列表项（附源数量）
type AdminCategoryDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	SourceCount int64  `json:"source_count"`
}

type categoryCountRow struct {
	CategoryID int64 `gorm:"column:category_id"`
	Total      int64 `gorm:"column:total"`
}

// ListCategories 按展示顺序返回全部分类
func (s *AdminRecommendService) ListCategories() ([]AdminCategoryDTO, error) {
	var cats []model.RecommendCategory
	if err := s.db.Order("sort_order ASC, id ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	var rows []categoryCountRow
	if err := s.db.Model(&model.RecommendSource{}).
		Select("category_id, COUNT(*) AS total").
		Group("category_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := make(map[int64]int64, len(rows))
	for _, r := range rows {
		counts[r.CategoryID] = r.Total
	}
	list := make([]AdminCategoryDTO, 0, len(cats))
	for _, c := range cats {
		list = append(list, AdminCategoryDTO{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			SortOrder:   c.SortOrder,
			SourceCount: counts[c.ID],
		})
	}
	return list, nil
}

// CreateCategory 新增分类，排到末尾
func (s *AdminRecommendService) CreateCategory(actor AdminActor, name, description string) (*model.RecommendCategory, error) {
	name = truncateRunes(name, 50)
	if name == "" {
		return nil, ErrCategoryNameEmpty
	}
	dup, err := s.categoryNameExists(name, 0)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, ErrCategoryDuplicate
	}

	var maxOrder int
	if err := s.db.Model(&model.RecommendCategory{}).
		Select("COALESCE(MAX(sort_order),0)").Scan(&maxOrder).Error; err != nil {
		return nil, err
	}
	cat := model.RecommendCategory{
		Name:        name,
		Description: truncateRunes(description, 500),
		SortOrder:   maxOrder + 10,
	}
	if err := s.db.Create(&cat).Error; err != nil {
		return nil, err
	}
	WriteAdminLog(s.db, actor, LogCategoryCreate, "新增分类："+cat.Name)
	return &cat, nil
}

// UpdateCategory 修改分类名称/描述，同步推荐源中的分类名快照
func (s *AdminRecommendService) UpdateCategory(actor AdminActor, id int64, name, description string) error {
	name = truncateRunes(name, 50)
	if name == "" {
		return ErrCategoryNameEmpty
	}
	var cat model.RecommendCategory
	if err := s.db.First(&cat, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryMissing
		}
		return err
	}
	dup, err := s.categoryNameExists(name, id)
	if err != nil {
		return err
	}
	if dup {
		return ErrCategoryDuplicate
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.RecommendCategory{}).Where("id = ?", id).
			Updates(map[string]any{
				"name":        name,
				"description": truncateRunes(description, 500),
			}).Error; err != nil {
			return err
		}
		return tx.Model(&model.RecommendSource{}).Where("category_id = ?", id).
			Update("category", name).Error
	})
	if err != nil {
		return err
	}
	WriteAdminLog(s.db, actor, LogCategoryUpdate, cat.Name+" → "+name)
	return nil
}

// DeleteCategory 删除空分类；分类下仍有推荐源时拒绝
func (s *AdminRecommendService) DeleteCategory(actor AdminActor, id int64) error {
	var cat model.RecommendCategory
	if err := s.db.First(&cat, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryMissing
		}
		return err
	}
	var count int64
	if err := s.db.Model(&model.RecommendSource{}).
		Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrCategoryHasSources
	}
	if err := s.db.Delete(&model.RecommendCategory{}, id).Error; err != nil {
		return err
	}
	WriteAdminLog(s.db, actor, LogCategoryDelete, "删除分类："+cat.Name)
	return nil
}

// ReorderCategories 按传入 id 顺序重排分类
func (s *AdminRecommendService) ReorderCategories(actor AdminActor, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.RecommendCategory{}).Where("id = ?", id).
				Update("sort_order", (i+1)*10).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	WriteAdminLog(s.db, actor, LogCategorySort, "调整分类顺序")
	return nil
}

func (s *AdminRecommendService) categoryNameExists(name string, excludeID int64) (bool, error) {
	tx := s.db.Model(&model.RecommendCategory{}).Where("name = ?", name)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ---------- 推荐源 ----------

// AdminSourceInput 新增/修改推荐源的入参
type AdminSourceInput struct {
	Title       string `json:"title"`
	FeedURL     string `json:"feed_url"`
	SiteURL     string `json:"site_url"`
	Description string `json:"description"`
	CategoryID  int64  `json:"category_id"`
	HotRank     int    `json:"hot_rank"`
}

// ListSources 推荐源列表；categoryID > 0 时只返回该分类下的源
func (s *AdminRecommendService) ListSources(categoryID int64) ([]model.RecommendSource, error) {
	var list []model.RecommendSource
	tx := s.db.Model(&model.RecommendSource{})
	if categoryID > 0 {
		tx = tx.Where("category_id = ?", categoryID)
	}
	if err := tx.Order("category_id ASC, sort_order ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.RecommendSource{}
	}
	return list, nil
}

// CreateSource 新增推荐源，排在所属分类末尾
func (s *AdminRecommendService) CreateSource(actor AdminActor, in AdminSourceInput) (*model.RecommendSource, error) {
	source, err := s.prepareSource(in)
	if err != nil {
		return nil, err
	}
	dup, err := s.feedExists(source.FeedURL, 0)
	if err != nil {
		return nil, err
	}
	if dup {
		return nil, ErrRecommendDuplicate
	}
	if err := s.db.Model(&model.RecommendSource{}).
		Where("category_id = ?", source.CategoryID).
		Select("COALESCE(MAX(sort_order),0)").Scan(&source.SortOrder).Error; err != nil {
		return nil, err
	}
	source.SortOrder += 10
	if err := s.db.Create(&source).Error; err != nil {
		return nil, err
	}
	WriteAdminLog(s.db, actor, LogSourceCreate, "新增推荐源："+source.Title)
	return &source, nil
}

// UpdateSource 修改推荐源；变更分类时移动到新分类末尾
func (s *AdminRecommendService) UpdateSource(actor AdminActor, id int64, in AdminSourceInput) error {
	var existing model.RecommendSource
	if err := s.db.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendMissing
		}
		return err
	}
	source, err := s.prepareSource(in)
	if err != nil {
		return err
	}
	dup, err := s.feedExists(source.FeedURL, id)
	if err != nil {
		return err
	}
	if dup {
		return ErrRecommendDuplicate
	}

	updates := map[string]any{
		"title":       source.Title,
		"feed_url":    source.FeedURL,
		"site_url":    source.SiteURL,
		"description": source.Description,
		"hot_rank":    source.HotRank,
		"category_id": source.CategoryID,
		"category":    source.Category,
	}
	if existing.CategoryID != source.CategoryID {
		var maxOrder int
		if err := s.db.Model(&model.RecommendSource{}).
			Where("category_id = ?", source.CategoryID).
			Select("COALESCE(MAX(sort_order),0)").Scan(&maxOrder).Error; err != nil {
			return err
		}
		updates["sort_order"] = maxOrder + 10
	}
	if err := s.db.Model(&model.RecommendSource{}).Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return err
	}
	WriteAdminLog(s.db, actor, LogSourceUpdate, "修改推荐源："+source.Title)
	return nil
}

// DeleteSource 删除推荐源
func (s *AdminRecommendService) DeleteSource(actor AdminActor, id int64) error {
	var source model.RecommendSource
	if err := s.db.First(&source, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendMissing
		}
		return err
	}
	if err := s.db.Delete(&model.RecommendSource{}, id).Error; err != nil {
		return err
	}
	WriteAdminLog(s.db, actor, LogSourceDelete, "删除推荐源："+source.Title)
	return nil
}

// ReorderSources 按传入 id 顺序重排推荐源（只传同一分类下的源）
func (s *AdminRecommendService) ReorderSources(actor AdminActor, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.RecommendSource{}).Where("id = ?", id).
				Update("sort_order", (i+1)*10).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	WriteAdminLog(s.db, actor, LogSourceSort, "调整推荐源顺序")
	return nil
}

func (s *AdminRecommendService) prepareSource(in AdminSourceInput) (model.RecommendSource, error) {
	title := truncateRunes(in.Title, 200)
	feedURL := strings.TrimSpace(in.FeedURL)
	if title == "" {
		return model.RecommendSource{}, ErrRecommendTitle
	}
	if !isHTTPURL(feedURL) {
		return model.RecommendSource{}, ErrRecommendFeedURL
	}
	var cat model.RecommendCategory
	if err := s.db.First(&cat, in.CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.RecommendSource{}, ErrCategoryMissing
		}
		return model.RecommendSource{}, err
	}
	hotRank := in.HotRank
	if hotRank < 0 {
		hotRank = 0
	}
	return model.RecommendSource{
		Title:       title,
		FeedURL:     feedURL,
		SiteURL:     strings.TrimSpace(in.SiteURL),
		Description: truncateRunes(in.Description, 500),
		CategoryID:  cat.ID,
		Category:    cat.Name,
		HotRank:     hotRank,
	}, nil
}

func (s *AdminRecommendService) feedExists(feedURL string, excludeID int64) (bool, error) {
	tx := s.db.Model(&model.RecommendSource{}).Select("id, feed_url")
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	var rows []model.RecommendSource
	if err := tx.Find(&rows).Error; err != nil {
		return false, err
	}
	target := normalizeFeedURL(feedURL)
	for _, r := range rows {
		if normalizeFeedURL(r.FeedURL) == target {
			return true, nil
		}
	}
	return false, nil
}

// ---------- 导入 / 导出 / 恢复初始 ----------

// RecommendExportSource 导入导出用推荐源结构
type RecommendExportSource struct {
	Title       string `json:"title"`
	FeedURL     string `json:"feed_url"`
	SiteURL     string `json:"site_url"`
	Description string `json:"description"`
	HotRank     int    `json:"hot_rank"`
}

// RecommendExportCategory 导入导出用分类结构（嵌套其下推荐源）
type RecommendExportCategory struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Sources     []RecommendExportSource `json:"sources"`
}

// RecommendExport 完整推荐数据快照
type RecommendExport struct {
	Categories []RecommendExportCategory `json:"categories"`
}

// Export 导出全部分类与推荐源（按当前展示顺序）
func (s *AdminRecommendService) Export() (*RecommendExport, error) {
	var cats []model.RecommendCategory
	if err := s.db.Order("sort_order ASC, id ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	var sources []model.RecommendSource
	if err := s.db.Order("sort_order ASC, id ASC").Find(&sources).Error; err != nil {
		return nil, err
	}
	grouped := make(map[int64][]RecommendExportSource, len(cats))
	for _, src := range sources {
		grouped[src.CategoryID] = append(grouped[src.CategoryID], RecommendExportSource{
			Title:       src.Title,
			FeedURL:     src.FeedURL,
			SiteURL:     src.SiteURL,
			Description: src.Description,
			HotRank:     src.HotRank,
		})
	}
	out := &RecommendExport{Categories: make([]RecommendExportCategory, 0, len(cats))}
	for _, c := range cats {
		items := grouped[c.ID]
		if items == nil {
			items = []RecommendExportSource{}
		}
		out.Categories = append(out.Categories, RecommendExportCategory{
			Name:        c.Name,
			Description: c.Description,
			Sources:     items,
		})
	}
	return out, nil
}

// Import 用导入数据整体替换推荐分类与推荐源
func (s *AdminRecommendService) Import(actor AdminActor, data RecommendExport) error {
	if len(data.Categories) == 0 {
		return ErrImportInvalid
	}

	names := make(map[string]bool, len(data.Categories))
	feedSeen := make(map[string]bool)
	prepared := make([]model.RecommendCategory, 0, len(data.Categories))
	preparedSources := make([][]model.RecommendSource, 0, len(data.Categories))

	for i, c := range data.Categories {
		name := truncateRunes(c.Name, 50)
		if name == "" {
			return ErrImportInvalid
		}
		if names[name] {
			return ErrCategoryDuplicate
		}
		names[name] = true
		cat := model.RecommendCategory{
			Name:        name,
			Description: truncateRunes(c.Description, 500),
			SortOrder:   (i + 1) * 10,
		}
		prepared = append(prepared, cat)

		items := make([]model.RecommendSource, 0, len(c.Sources))
		for j, src := range c.Sources {
			title := truncateRunes(src.Title, 200)
			feedURL := strings.TrimSpace(src.FeedURL)
			if title == "" || !isHTTPURL(feedURL) {
				return ErrImportInvalid
			}
			key := normalizeFeedURL(feedURL)
			if feedSeen[key] {
				return ErrRecommendDuplicate
			}
			feedSeen[key] = true
			hotRank := src.HotRank
			if hotRank < 0 {
				hotRank = 0
			}
			items = append(items, model.RecommendSource{
				Title:       title,
				FeedURL:     feedURL,
				SiteURL:     strings.TrimSpace(src.SiteURL),
				Description: truncateRunes(src.Description, 500),
				Category:    name,
				HotRank:     hotRank,
				SortOrder:   (j + 1) * 10,
			})
		}
		preparedSources = append(preparedSources, items)
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.RecommendSource{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&model.RecommendCategory{}).Error; err != nil {
			return err
		}
		for i := range prepared {
			if err := tx.Create(&prepared[i]).Error; err != nil {
				return err
			}
			items := preparedSources[i]
			for j := range items {
				items[j].CategoryID = prepared[i].ID
			}
			if len(items) > 0 {
				if err := tx.Create(&items).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	WriteAdminLog(s.db, actor, LogSourceImport, "导入推荐数据，分类 "+strconv.Itoa(len(prepared))+" 个")
	return nil
}

// Reset 清空并恢复内置推荐数据
func (s *AdminRecommendService) Reset(actor AdminActor) error {
	if err := model.ResetRecommendData(s.db); err != nil {
		return err
	}
	WriteAdminLog(s.db, actor, LogSystemReset, "恢复内置推荐分类与推荐源")
	return nil
}
