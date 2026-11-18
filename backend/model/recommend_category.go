package model

import "time"

// RecommendCategory 发现页推荐分类（后台可维护，sort_order 越小越靠前）
type RecommendCategory struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	SortOrder   int       `gorm:"not null;default:0;index" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
