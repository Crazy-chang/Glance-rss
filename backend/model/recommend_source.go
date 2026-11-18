package model

import "time"

// RecommendSource 发现页推荐源（后台可维护）
// Category 保留分类名称快照，便于无 join 的简单查询；CategoryID 为真实关联
type RecommendSource struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	FeedURL     string    `gorm:"size:500;not null" json:"feed_url"`
	SiteURL     string    `gorm:"size:500" json:"site_url"`
	Category    string    `gorm:"size:50;not null;index" json:"category"`
	CategoryID  int64     `gorm:"not null;default:0;index" json:"category_id"`
	Description string    `gorm:"size:500" json:"description"`
	HotRank     int       `gorm:"not null;default:0" json:"hot_rank"`         // >0 进入热门订阅，按升序排列
	SortOrder   int       `gorm:"not null;default:0;index" json:"sort_order"` // 分类内排序，越小越靠前
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
