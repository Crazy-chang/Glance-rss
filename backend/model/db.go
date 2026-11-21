// Package model 定义全部 GORM 模型与数据库初始化
package model

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 打开 SQLite 数据库并确保目录存在
func Open(dsn string) (*gorm.DB, error) {
	if dir := filepath.Dir(dsn); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %w", err)
		}
	}

	// glebarez（modernc 纯 Go 驱动）支持 _pragma 查询参数
	fullDSN := dsn + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true, // 首次登录等正常的"未查到"不打错误日志
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(sqlite.Open(fullDSN), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	return db, nil
}

// AutoMigrate 自动建表
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&Source{},
		&Article{},
		&UserArticleState{},
		&Setting{},
		&RecommendCategory{},
		&RecommendSource{},
		&AdminUser{},
		&AdminLog{},
		&AppConfig{},
	); err != nil {
		return err
	}
	_, err := EnsureAppConfig(db)
	return err
}

// BackfillFavoritedAt 为历史收藏补齐收藏时间（旧数据没有该字段，用 updated_at 近似）
// 仅处理 is_favorited=1 且 favorited_at 为空的行，可重复执行
func BackfillFavoritedAt(db *gorm.DB) error {
	return db.Model(&UserArticleState{}).
		Where("is_favorited = 1 AND favorited_at IS NULL").
		Update("favorited_at", gorm.Expr("updated_at")).Error
}

// SeedRecommendData 推荐数据初始化与旧数据迁移（幂等，启动时调用）：
//  1. 分类表为空时写入内置分类
//  2. 历史推荐源（仅有 category 文本、无 category_id）回填分类关联与排序
//  3. 推荐源表为空时写入内置推荐源
//
// 非空表不再补写，避免后台删除的数据在重启后被 seed 复活
func SeedRecommendData(db *gorm.DB) error {
	if err := seedRecommendCategories(db); err != nil {
		return err
	}
	if err := backfillRecommendCategoryIDs(db); err != nil {
		return err
	}
	return seedRecommendSources(db)
}

func seedRecommendCategories(db *gorm.DB) error {
	var count int64
	if err := db.Model(&RecommendCategory{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	cats := make([]RecommendCategory, 0, len(defaultRecommendCategories))
	for i, c := range defaultRecommendCategories {
		c.SortOrder = (i + 1) * 10
		cats = append(cats, c)
	}
	return db.Create(&cats).Error
}

// backfillRecommendCategoryIDs 把旧版 category 文本迁移为分类表关联
// 对 seed 中存在同名 feed_url 的源，以 seed 分类为准（修正历史上误归到"其他"的数据）
func backfillRecommendCategoryIDs(db *gorm.DB) error {
	var legacy []RecommendSource
	if err := db.Where("category_id = 0").Find(&legacy).Error; err != nil {
		return err
	}
	if len(legacy) == 0 {
		return nil
	}

	var cats []RecommendCategory
	if err := db.Order("sort_order ASC, id ASC").Find(&cats).Error; err != nil {
		return err
	}
	byName := make(map[string]*RecommendCategory, len(cats))
	maxCatOrder := 0
	for i := range cats {
		byName[cats[i].Name] = &cats[i]
		if cats[i].SortOrder > maxCatOrder {
			maxCatOrder = cats[i].SortOrder
		}
	}

	seedCategory := make(map[string]string, len(defaultRecommendSources))
	for _, s := range defaultRecommendSources {
		seedCategory[s.FeedURL] = s.Category
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, r := range legacy {
			name := r.Category
			if seedName, ok := seedCategory[r.FeedURL]; ok && seedName != "" {
				name = seedName
			}
			if name == "" {
				name = "未分类"
			}
			cat, ok := byName[name]
			if !ok {
				maxCatOrder += 10
				cat = &RecommendCategory{Name: name, SortOrder: maxCatOrder}
				if err := tx.Create(cat).Error; err != nil {
					return err
				}
				byName[name] = cat
			}

			updates := map[string]any{"category_id": cat.ID, "category": cat.Name}
			if r.SortOrder == 0 {
				var maxOrder int
				if err := tx.Model(&RecommendSource{}).
					Where("category_id = ?", cat.ID).
					Select("COALESCE(MAX(sort_order),0)").
					Scan(&maxOrder).Error; err != nil {
					return err
				}
				updates["sort_order"] = maxOrder + 10
			}
			if err := tx.Model(&RecommendSource{}).
				Where("id = ?", r.ID).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// seedRecommendSources 仅在推荐源表为空时写入内置推荐源
func seedRecommendSources(db *gorm.DB) error {
	var count int64
	if err := db.Model(&RecommendSource{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var cats []RecommendCategory
	if err := db.Find(&cats).Error; err != nil {
		return err
	}
	catID := make(map[string]int64, len(cats))
	for _, c := range cats {
		catID[c.Name] = c.ID
	}

	nextOrder := make(map[int64]int, len(cats))
	toInsert := make([]RecommendSource, 0, len(defaultRecommendSources))
	for _, s := range defaultRecommendSources {
		id, ok := catID[s.Category]
		if !ok {
			continue // 分类被调整过则跳过，避免写入孤儿数据
		}
		nextOrder[id] += 10
		s.CategoryID = id
		s.SortOrder = nextOrder[id]
		toInsert = append(toInsert, s)
	}
	if len(toInsert) == 0 {
		return nil
	}
	return db.Create(&toInsert).Error
}

// ResetRecommendData 清空推荐分类与推荐源，恢复内置初始数据（后台"恢复初始设置"）
func ResetRecommendData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&RecommendSource{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&RecommendCategory{}).Error; err != nil {
			return err
		}
		if err := seedRecommendCategories(tx); err != nil {
			return err
		}
		return seedRecommendSources(tx)
	})
}
