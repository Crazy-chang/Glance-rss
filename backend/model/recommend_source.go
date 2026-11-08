package model

import "time"

// RecommendCategoryOrder 分类区块的展示顺序（取 defaultRecommendSources 中首次出现顺序）
// SQL 的 category ASC 对中文按 UTF-8 字节序排序，结果与 seed 书写顺序无关，故在内存中排序
func RecommendCategoryOrder() []string {
	seen := make(map[string]bool, 8)
	order := make([]string, 0, 8)
	for _, s := range defaultRecommendSources {
		if !seen[s.Category] {
			seen[s.Category] = true
			order = append(order, s.Category)
		}
	}
	return order
}

// RecommendSeedIndex 返回 seed 中 feed_url → 数组下标 的映射
// 发现页分类内的源顺序以此为准：seed 文件中越靠前，展示越靠前
func RecommendSeedIndex() map[string]int {
	idx := make(map[string]int, len(defaultRecommendSources))
	for i, s := range defaultRecommendSources {
		idx[s.FeedURL] = i
	}
	return idx
}

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
