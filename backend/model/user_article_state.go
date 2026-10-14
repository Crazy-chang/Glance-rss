package model

import "time"

// UserArticleState 用户维度的文章状态（已读/收藏）
type UserArticleState struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64      `gorm:"not null;uniqueIndex:idx_state_user_article,priority:1;index:idx_state_user_read" json:"user_id"`
	ArticleID   int64      `gorm:"not null;uniqueIndex:idx_state_user_article,priority:2;index:idx_state_user_fav" json:"article_id"`
	IsRead      bool       `gorm:"not null;default:0;index:idx_state_user_read,priority:2" json:"is_read"`
	ReadAt      *time.Time `json:"read_at"`
	IsFavorited bool       `gorm:"not null;default:0;index:idx_state_user_fav,priority:2" json:"is_favorited"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
