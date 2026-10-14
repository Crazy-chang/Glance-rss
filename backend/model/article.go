package model

import "time"

// Article 文章（入库时即抓取并清洗好正文，列表接口不返回 content）
type Article struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceID    int64     `gorm:"not null;uniqueIndex:idx_article_source_guid,priority:1;index:idx_article_source_pub,priority:1" json:"source_id"`
	GUID        string    `gorm:"size:500;not null;uniqueIndex:idx_article_source_guid,priority:2" json:"guid"`
	Link        string    `gorm:"size:1000" json:"link"`
	Title       string    `gorm:"size:500;not null" json:"title"`
	Author      string    `gorm:"size:200" json:"author"`
	Summary     string    `gorm:"size:300" json:"summary"` // 纯文本前 100 字
	Content     string    `gorm:"type:text" json:"content,omitempty"`
	PublishedAt time.Time `gorm:"not null;index:idx_article_source_pub,priority:2;index" json:"published_at"` // 缺失时回退抓取时间
	FetchedAt   time.Time `gorm:"not null" json:"fetched_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
