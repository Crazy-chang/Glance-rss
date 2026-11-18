package model

import "time"

// AdminUser 后台管理员（与小程序用户体系完全独立）
type AdminUser struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"size:100;not null" json:"-"`
	Role         string     `gorm:"size:20;not null;default:admin" json:"role"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// AdminLog 后台操作日志（保留最近若干条）
type AdminLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID   int64     `gorm:"index" json:"admin_id"`
	Username  string    `gorm:"size:50" json:"username"`
	Action    string    `gorm:"size:50;not null" json:"action"`
	Detail    string    `gorm:"size:500" json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}
