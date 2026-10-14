package model

import "time"

// RecommendSource 发现页内置推荐源（用户后续直接维护数据库）
type RecommendSource struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	FeedURL     string    `gorm:"size:500;not null" json:"feed_url"`
	SiteURL     string    `gorm:"size:500" json:"site_url"`
	Category    string    `gorm:"size:50;not null;index" json:"category"`
	Description string    `gorm:"size:500" json:"description"`
	HotRank     int       `gorm:"not null;default:0" json:"hot_rank"` // >0 进入热门订阅，按升序排列
	CreatedAt   time.Time `json:"created_at"`
}
