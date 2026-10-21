package service

import (
	"errors"
	"slices"

	"gorm.io/gorm"

	"glance-wx/backend/model"
)

// SettingService 用户设置读写
type SettingService struct {
	db *gorm.DB
}

func NewSettingService(db *gorm.DB) *SettingService {
	return &SettingService{db: db}
}

// SettingDTO 设置返回结构（字段与前端一一对应）
type SettingDTO struct {
	Theme           string `json:"theme"`
	FontSize        string `json:"font_size"`
	MaxItems        int    `json:"max_items"`
	RefreshInterval string `json:"refresh_interval"`
	SortOrder       string `json:"sort_order"`
}

// 合法取值白名单
var (
	validThemes    = []string{model.ThemeWhite, model.ThemeBlack, model.ThemeYellow, model.ThemeGreen}
	validFonts     = []string{model.FontSizeSmall, model.FontSizeMedium, model.FontSizeLarge}
	validItems     = []int{100, 200, 300, 500}
	validIntervals = []string{
		model.RefreshInterval30m, model.RefreshInterval1h,
		model.RefreshInterval2h, model.RefreshIntervalManual,
	}
	validSortOrders = []string{model.SortOrderDesc, model.SortOrderAsc}
)

// UpdateSettingParams 更新请求（指针区分"未传"与"传零值"）
type UpdateSettingParams struct {
	Theme           *string `json:"theme"`
	FontSize        *string `json:"font_size"`
	MaxItems        *int    `json:"max_items"`
	RefreshInterval *string `json:"refresh_interval"`
	SortOrder       *string `json:"sort_order"`
}

// Info 取用户设置（正常情况登录时已自动创建）
func (s *SettingService) Info(userID int64) (*SettingDTO, error) {
	var setting model.Setting
	err := s.db.Where("user_id = ?", userID).Take(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 兜底：历史数据缺失时补建默认设置
		setting = model.DefaultSetting(userID)
		if err := s.db.Create(&setting).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return toSettingDTO(setting), nil
}

// Update 校验白名单后更新，返回最新设置
func (s *SettingService) Update(userID int64, p UpdateSettingParams) (*SettingDTO, error) {
	if p.Theme != nil && !slices.Contains(validThemes, *p.Theme) {
		return nil, ErrSettingValue
	}
	if p.FontSize != nil && !slices.Contains(validFonts, *p.FontSize) {
		return nil, ErrSettingValue
	}
	if p.MaxItems != nil && !slices.Contains(validItems, *p.MaxItems) {
		return nil, ErrSettingValue
	}
	if p.RefreshInterval != nil && !slices.Contains(validIntervals, *p.RefreshInterval) {
		return nil, ErrSettingValue
	}
	if p.SortOrder != nil && !slices.Contains(validSortOrders, *p.SortOrder) {
		return nil, ErrSettingValue
	}

	var setting model.Setting
	err := s.db.Where("user_id = ?", userID).Take(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setting = model.DefaultSetting(userID)
		if err := s.db.Create(&setting).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if p.Theme != nil {
		setting.Theme = *p.Theme
		updates["theme"] = *p.Theme
	}
	if p.FontSize != nil {
		setting.FontSize = *p.FontSize
		updates["font_size"] = *p.FontSize
	}
	if p.MaxItems != nil {
		setting.MaxItems = *p.MaxItems
		updates["max_items"] = *p.MaxItems
	}
	if p.RefreshInterval != nil {
		setting.RefreshInterval = *p.RefreshInterval
		updates["refresh_interval"] = *p.RefreshInterval
	}
	if p.SortOrder != nil {
		setting.SortOrder = *p.SortOrder
		updates["sort_order"] = *p.SortOrder
	}
	if len(updates) > 0 {
		if err := s.db.Model(&setting).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return toSettingDTO(setting), nil
}

func toSettingDTO(s model.Setting) *SettingDTO {
	return &SettingDTO{
		Theme:           s.Theme,
		FontSize:        s.FontSize,
		MaxItems:        s.MaxItems,
		RefreshInterval: s.RefreshInterval,
		SortOrder:       s.SortOrder,
	}
}
