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
	return db.AutoMigrate(
		&User{},
		&Source{},
		&Article{},
		&UserArticleState{},
		&Setting{},
		&RecommendSource{},
	)
}

// SeedRecommendSources 把 defaultRecommendSources 中缺失的推荐源 Upsert 到数据库
// 按 FeedURL 去重：已存在则保留数据库中的版本（不覆盖 Category/Title 等字段），不存在则新增
func SeedRecommendSources(db *gorm.DB) error {
	var existing []RecommendSource
	if err := db.Find(&existing).Error; err != nil {
		return err
	}
	existMap := make(map[string]bool, len(existing))
	for _, s := range existing {
		existMap[s.FeedURL] = true
	}
	var toInsert []RecommendSource
	for _, s := range defaultRecommendSources {
		if !existMap[s.FeedURL] {
			toInsert = append(toInsert, s)
		}
	}
	if len(toInsert) == 0 {
		return nil
	}
	return db.Create(&toInsert).Error
}
