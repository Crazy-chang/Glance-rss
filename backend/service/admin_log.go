package service

import (
	"gorm.io/gorm"

	"glance-wx/backend/model"
)

// adminLogKeep 操作日志最多保留的条数
const adminLogKeep = 500

// WriteAdminLog 记录后台写操作，超出上限时清理最旧记录（失败不影响主流程）
func WriteAdminLog(db *gorm.DB, actor AdminActor, action, detail string) {
	entry := model.AdminLog{
		AdminID:  actor.ID,
		Username: actor.Username,
		Action:   action,
		Detail:   truncateRunes(detail, 500),
	}
	if err := db.Create(&entry).Error; err != nil {
		return
	}
	var count int64
	if err := db.Model(&model.AdminLog{}).Count(&count).Error; err != nil || count <= adminLogKeep {
		return
	}
	db.Where("id NOT IN (?)",
		db.Model(&model.AdminLog{}).Select("id").Order("id DESC").Limit(adminLogKeep)).
		Delete(&model.AdminLog{})
}

// ListAdminLogs 最近的登录与操作记录
func ListAdminLogs(db *gorm.DB, limit int) ([]model.AdminLog, error) {
	if limit <= 0 || limit > adminLogKeep {
		limit = 100
	}
	var logs []model.AdminLog
	if err := db.Order("id DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	if logs == nil {
		logs = []model.AdminLog{}
	}
	return logs, nil
}

// 日志动作名（前端直接展示，与后台页面文案保持一致）
const (
	LogCategoryCreate = "新增分类"
	LogCategoryUpdate = "修改分类"
	LogCategoryDelete = "删除分类"
	LogCategorySort   = "分类排序"
	LogSourceCreate   = "新增推荐源"
	LogSourceUpdate   = "修改推荐源"
	LogSourceDelete   = "删除推荐源"
	LogSourceSort     = "推荐源排序"
	LogSourceImport   = "导入推荐数据"
	LogSystemReset    = "恢复初始设置"
)
