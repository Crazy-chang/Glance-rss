package model

import (
	"errors"

	"gorm.io/gorm"
)

// 全局配置默认值
const DefaultBatchReadLimit = 5

// AppConfig 全局配置单例（固定 ID=1）
type AppConfig struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	BatchReadLimit int   `gorm:"not null;default:5" json:"batch_read_limit"` // 一键阅读条数，0 表示全部
}

// EnsureAppConfig 保证配置行存在并返回当前配置
func EnsureAppConfig(db *gorm.DB) (AppConfig, error) {
	var cfg AppConfig
	err := db.First(&cfg, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cfg = AppConfig{ID: 1, BatchReadLimit: DefaultBatchReadLimit}
		if err := db.Create(&cfg).Error; err != nil {
			return AppConfig{}, err
		}
		return cfg, nil
	}
	return cfg, err
}

// SetBatchReadLimit 更新一键阅读条数（0 表示全部）
func SetBatchReadLimit(db *gorm.DB, limit int) error {
	if _, err := EnsureAppConfig(db); err != nil {
		return err
	}
	return db.Model(&AppConfig{}).Where("id = 1").
		Update("batch_read_limit", limit).Error
}
