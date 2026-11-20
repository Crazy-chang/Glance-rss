package model

import "time"

// RecommendCategory 发现页推荐分类（后台可维护，sort_order 越小越靠前）
type RecommendCategory struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	SortOrder   int       `gorm:"not null;default:0;index" json:"sort_order"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"` // 停用后不在发现页展示，其下推荐源一并隐藏
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
