package service

import "errors"

// 业务错误（service 层返回，handler 层映射为具体提示语）
var (
	ErrInvalidCode    = errors.New("登录凭证无效")
	ErrArticleMissing = errors.New("文章不存在")
	ErrSourceMissing  = errors.New("订阅源不存在")
	ErrFeedDuplicate  = errors.New("该 RSS 源已订阅")
	ErrFeedInvalid    = errors.New("RSS 地址无效或无法访问")
	ErrSettingValue   = errors.New("设置项取值不合法")
)
