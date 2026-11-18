package service

import (
	"time"

	"gorm.io/gorm"

	"glance-wx/backend/model"
)

// AdminUserSourceDTO 后台查看用户订阅时的源信息
type AdminUserSourceDTO struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	FeedURL     string     `json:"feed_url"`
	SiteURL     string     `json:"site_url"`
	Category    string     `json:"category"`
	LastFetchAt *time.Time `json:"last_fetch_at"`
	LastError   string     `json:"last_error"`
	CreatedAt   time.Time  `json:"created_at"`
}

// AdminUserDTO 后台查看用户及其订阅源
type AdminUserDTO struct {
	ID          int64                `json:"id"`
	OpenID      string               `json:"open_id"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	SourceCount int64                `json:"source_count"`
	Sources     []AdminUserSourceDTO `json:"sources"`
}

// AdminUserListResult 用户订阅分页结果
type AdminUserListResult struct {
	List    []AdminUserDTO `json:"list"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	Size    int            `json:"size"`
	HasMore bool           `json:"hasMore"`
}

// ListAdminUsers 分页查询用户及其订阅源（只读）。
// minSourceCount > 0 时只返回订阅源数量 >= minSourceCount 的用户。
func ListAdminUsers(db *gorm.DB, page, size int, minSourceCount int64) (*AdminUserListResult, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	userQuery := db.Model(&model.User{})
	if minSourceCount > 0 {
		userQuery = userQuery.Where(
			"id IN (SELECT user_id FROM sources GROUP BY user_id HAVING COUNT(*) >= ?)",
			minSourceCount,
		)
	}

	var total int64
	if err := userQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []model.User
	if err := userQuery.Order("created_at DESC, id DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&users).Error; err != nil {
		return nil, err
	}

	result := &AdminUserListResult{
		List:    []AdminUserDTO{},
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: int64(page*size) < total,
	}
	if len(users) == 0 {
		return result, nil
	}

	ids := make([]int64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}

	var sources []model.Source
	if err := db.Where("user_id IN ?", ids).Order("id ASC").Find(&sources).Error; err != nil {
		return nil, err
	}

	grouped := make(map[int64][]AdminUserSourceDTO, len(users))
	for _, s := range sources {
		grouped[s.UserID] = append(grouped[s.UserID], AdminUserSourceDTO{
			ID:          s.ID,
			Title:       s.Title,
			FeedURL:     s.FeedURL,
			SiteURL:     s.SiteURL,
			Category:    s.Category,
			LastFetchAt: s.LastFetchAt,
			LastError:   s.LastError,
			CreatedAt:   s.CreatedAt,
		})
	}

	for _, u := range users {
		list := grouped[u.ID]
		if list == nil {
			list = []AdminUserSourceDTO{}
		}
		result.List = append(result.List, AdminUserDTO{
			ID:          u.ID,
			OpenID:      u.OpenID,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
			SourceCount: int64(len(list)),
			Sources:     list,
		})
	}

	return result, nil
}
