package model

import "time"

// Source 用户订阅源
type Source struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64      `gorm:"not null;uniqueIndex:idx_source_user_feed,priority:1;index" json:"user_id"`
	Title        string     `gorm:"size:200;not null" json:"title"`
	FeedURL      string     `gorm:"size:500;not null;uniqueIndex:idx_source_user_feed,priority:2" json:"feed_url"`
	SiteURL      string     `gorm:"size:500" json:"site_url"`
	FaviconPath  string     `gorm:"size:300" json:"favicon_path"` // 本地相对路径，空表示未下载
	Category     string     `gorm:"size:50" json:"category"`
	Description  string     `gorm:"size:500" json:"description"`
	ETag         string     `gorm:"size:500" json:"etag"`          // HTTP 条件请求：If-None-Match
	LastModified string     `gorm:"size:100" json:"last_modified"` // HTTP 条件请求：If-Modified-Since
	LastFetchAt  *time.Time `json:"last_fetch_at"`
	LastError    string     `gorm:"size:300" json:"last_error"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
