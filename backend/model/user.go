package model

import "time"

// User 用户（静默 openid 体系，无注册流程）
type User struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	OpenID    string `gorm:"size:64;uniqueIndex;not null" json:"-"` // 不下发给前端
	// BatchReadEnabled 是否开放「一键阅读」，由管理后台按用户开启
	BatchReadEnabled bool      `gorm:"not null;default:false" json:"batch_read_enabled"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
