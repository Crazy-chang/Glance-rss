package service

import "errors"

// 业务错误（service 层返回，handler 层映射为具体提示语）
var (
	ErrInvalidCode      = errors.New("登录凭证无效")
	ErrArticleMissing   = errors.New("文章不存在")
	ErrSourceMissing    = errors.New("订阅源不存在")
	ErrFeedDuplicate    = errors.New("该 RSS 源已订阅")
	ErrFeedInvalid      = errors.New("RSS 地址无效或无法访问")
	ErrSettingValue     = errors.New("设置项取值不合法")
	ErrSourceTitleEmpty = errors.New("订阅源名称不能为空")
	ErrBatchReadOff     = errors.New("一键阅读未开启")
)

// 后台管理错误
var (
	ErrAdminCredentials   = errors.New("用户名或密码错误")
	ErrCategoryMissing    = errors.New("分类不存在")
	ErrCategoryNameEmpty  = errors.New("分类名称不能为空")
	ErrCategoryDuplicate  = errors.New("分类名称已存在")
	ErrCategoryHasSources = errors.New("该分类下还有推荐源，请先删除或移动")
	ErrRecommendMissing   = errors.New("推荐源不存在")
	ErrRecommendTitle     = errors.New("推荐源名称不能为空")
	ErrRecommendFeedURL   = errors.New("RSS 地址不合法")
	ErrRecommendDuplicate = errors.New("该 RSS 地址已存在")
	ErrImportInvalid      = errors.New("导入数据格式不正确")
	ErrUserMissing        = errors.New("用户不存在")
	ErrConfigValue        = errors.New("配置取值不合法")
)
