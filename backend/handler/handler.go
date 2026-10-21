// Package handler HTTP 接口层：只做收参、调用 service、统一回包
package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/config"
	"glance-wx/backend/middleware"
	"glance-wx/backend/scheduler"
	"glance-wx/backend/service"
)

// Handler 持有全部 service 依赖
type Handler struct {
	cfg      *config.Config
	auth     *service.AuthService
	article  *service.ArticleService
	source   *service.SourceService
	discover *service.DiscoverService
	setting  *service.SettingService
	proxy    *service.ProxyService
	sched    *scheduler.Scheduler
}

// Deps 构造 Handler 的依赖集合
type Deps struct {
	Cfg      *config.Config
	Auth     *service.AuthService
	Article  *service.ArticleService
	Source   *service.SourceService
	Discover *service.DiscoverService
	Setting  *service.SettingService
	Proxy    *service.ProxyService
	Sched    *scheduler.Scheduler
}

func New(d Deps) *Handler {
	return &Handler{
		cfg:      d.Cfg,
		auth:     d.Auth,
		article:  d.Article,
		source:   d.Source,
		discover: d.Discover,
		setting:  d.Setting,
		proxy:    d.Proxy,
		sched:    d.Sched,
	}
}

// Register 在指定路由组上注册全部接口
func (h *Handler) Register(api *gin.RouterGroup) {
	api.GET("/ping", h.ping)
	api.POST("/auth/login", h.login)

	authed := api.Group("").Use(middleware.Auth(h.auth))
	{
		// 文章
		authed.GET("/article/list", h.articleList)
		authed.GET("/article/detail/:id", h.articleDetail)
		authed.POST("/article/read/:id", h.articleMarkRead)
		authed.POST("/article/unread/:id", h.articleMarkUnread)
		authed.POST("/article/favorite/:id", h.articleFavorite)
		authed.POST("/article/unfavorite/:id", h.articleUnfavorite)
		authed.POST("/article/clear-read", h.articleClearRead)

		// 订阅源
		authed.GET("/source/list", h.sourceList)
		authed.POST("/source/add", h.sourceAdd)
		authed.POST("/source/rename/:id", h.sourceRename)
		authed.POST("/source/delete/:id", h.sourceDelete)

		// 发现
		authed.GET("/discover/list", h.discoverList)
		authed.POST("/discover/subscribe/:id", h.subscribeRecommend)

		// 设置
		authed.GET("/setting/info", h.settingInfo)
		authed.POST("/setting/update", h.settingUpdate)

		// 刷新与资源
		authed.POST("/refresh/start", h.refreshStart)
		authed.GET("/favicon/:sourceId", h.favicon)
		authed.GET("/img-proxy", h.imgProxy)
	}
}

func (h *Handler) ping(c *gin.Context) {
	service.OK(c, gin.H{"pong": true})
}

// uid 取当前登录用户 ID
func uid(c *gin.Context) int64 {
	return middleware.CurrentUserID(c)
}

// parseIDParam 解析路径中的整数 id
func parseIDParam(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		service.FailBadRequest(c, "参数错误")
		return 0, false
	}
	return id, true
}

// handleServiceError 把 service 错误映射为 HTTP 响应；未知错误记日志后返回系统错误
func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrArticleMissing):
		service.FailNotFound(c, "文章不存在")
	case errors.Is(err, service.ErrSourceMissing):
		service.FailNotFound(c, "订阅源不存在")
	case errors.Is(err, service.ErrSourceTitleEmpty):
		service.FailBadRequest(c, "订阅源名称不能为空")
	case errors.Is(err, service.ErrFeedInvalid):
		service.FailBadRequest(c, "RSS地址无效或无法访问")
	case errors.Is(err, service.ErrFeedDuplicate):
		service.FailBadRequest(c, "该订阅源已存在")
	case errors.Is(err, service.ErrSettingValue):
		service.FailBadRequest(c, "设置项取值不合法")
	case errors.Is(err, service.ErrInvalidCode):
		service.FailBadRequest(c, "登录失败，请重试")
	default:
		slog.Error("接口内部错误", "path", c.FullPath(), "err", err)
		service.FailServer(c)
	}
}
