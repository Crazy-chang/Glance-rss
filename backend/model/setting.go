package model

import "time"

// 设置取值常量
const (
	ThemeWhite  = "white"  // 白色（默认）
	ThemeBlack  = "black"  // 黑色
	ThemeYellow = "yellow" // 黄色
	ThemeGreen  = "green"  // 绿色

	FontSizeSmall  = "small"  // 正文 15
	FontSizeMedium = "medium" // 正文 17（默认）
	FontSizeLarge  = "large"  // 正文 19

	LineHeightCompact = "compact" // 行距 1.45
	LineHeightNormal  = "normal"  // 行距 1.65（默认）
	LineHeightLoose   = "loose"   // 行距 1.85

	TextAlignLeft    = "left"    // 左对齐（默认）
	TextAlignJustify = "justify" // 两端对齐

	RefreshInterval30m    = "30m"    // 30 分钟（默认）
	RefreshInterval1h     = "1h"     // 1 小时
	RefreshInterval2h     = "2h"     // 2 小时
	RefreshIntervalManual = "manual" // 手动，不参与定时抓取

	SortOrderDesc = "desc" // 最新在前（默认）
	SortOrderAsc  = "asc"  // 最早在前
)

// DefaultMaxItems 默认单源最大条数
const DefaultMaxItems = 300

// Setting 用户设置（随用户创建自动生成一条默认值）
type Setting struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64     `gorm:"not null;uniqueIndex" json:"user_id"`
	Theme           string    `gorm:"size:20;not null;default:white" json:"theme"`
	FontSize        string    `gorm:"size:20;not null;default:medium" json:"font_size"`
	LineHeight      string    `gorm:"size:20;not null;default:normal" json:"line_height"`
	TextAlign       string    `gorm:"size:20;not null;default:left" json:"text_align"`
	MaxItems        int       `gorm:"not null;default:300" json:"max_items"`
	RefreshInterval string    `gorm:"size:20;not null;default:1h" json:"refresh_interval"`
	SortOrder       string    `gorm:"size:10;not null;default:desc" json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DefaultSetting 生成用户的默认设置
func DefaultSetting(userID int64) Setting {
	return Setting{
		UserID:          userID,
		Theme:           ThemeWhite,
		FontSize:        FontSizeMedium,
		LineHeight:      LineHeightNormal,
		TextAlign:       TextAlignLeft,
		MaxItems:        DefaultMaxItems,
		RefreshInterval: RefreshInterval1h,
		SortOrder:       SortOrderDesc,
	}
}
