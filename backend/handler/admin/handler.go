// Package admin 后台管理接口（与小程序用户接口完全独立）
package admin

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"glance-wx/backend/middleware"
	"glance-wx/backend/service"
)

// Handler 后台接口依赖
type Handler struct {
	db        *gorm.DB
	auth      *service.AdminAuthService
	recommend *service.AdminRecommendService
}

// Deps 构造 Handler 的依赖集合
type Deps struct {
	DB        *gorm.DB
	Auth      *service.AdminAuthService
	Recommend *service.AdminRecommendService
}

func New(d Deps) *Handler {
	return &Handler{db: d.DB, auth: d.Auth, recommend: d.Recommend}
}

// Register 注册 /api/admin 下的全部接口
func (h *Handler) Register(g *gin.RouterGroup) {
	g.POST("/login", h.login)

	authed := g.Group("").Use(middleware.AdminAuth(h.auth))
	{
		authed.POST("/logout", h.logout)

		// 分类
		authed.GET("/category/list", h.categoryList)
		authed.POST("/category/create", h.categoryCreate)
		authed.POST("/category/update", h.categoryUpdate)
		authed.POST("/category/delete", h.categoryDelete)
		authed.POST("/category/status", h.categoryStatus)
		authed.POST("/category/reorder", h.categoryReorder)

		// 推荐源
		authed.GET("/recommend/list", h.recommendList)
		authed.POST("/recommend/create", h.recommendCreate)
		authed.POST("/recommend/update", h.recommendUpdate)
		authed.POST("/recommend/delete", h.recommendDelete)
		authed.POST("/recommend/status", h.recommendStatus)
		authed.POST("/recommend/reorder", h.recommendReorder)
		authed.POST("/recommend/import", h.recommendImport)
		authed.GET("/recommend/export", h.recommendExport)

		// 用户订阅（只读）
		authed.GET("/user/list", h.userList)

		// 系统
		authed.POST("/system/reset", h.systemReset)
		authed.GET("/system/logs", h.systemLogs)
	}
}

// idRequest 仅含主键的请求体（删除等只按 id 操作的接口）
type idRequest struct {
	ID int64 `json:"id"`
}

// bindJSON 解析请求体，失败直接回 400
func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		service.FailBadRequest(c, "参数错误")
		return false
	}
	return true
}

// handleError 把 service 错误映射为响应
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAdminCredentials):
		service.FailBadRequest(c, "用户名或密码错误")
	case errors.Is(err, service.ErrCategoryMissing):
		service.FailNotFound(c, "分类不存在")
	case errors.Is(err, service.ErrCategoryNameEmpty):
		service.FailBadRequest(c, "分类名称不能为空")
	case errors.Is(err, service.ErrCategoryDuplicate):
		service.FailBadRequest(c, "分类名称已存在")
	case errors.Is(err, service.ErrCategoryHasSources):
		service.FailBadRequest(c, "该分类下还有推荐源，请先删除或移动")
	case errors.Is(err, service.ErrRecommendMissing):
		service.FailNotFound(c, "推荐源不存在")
	case errors.Is(err, service.ErrRecommendTitle):
		service.FailBadRequest(c, "推荐源名称不能为空")
	case errors.Is(err, service.ErrRecommendFeedURL):
		service.FailBadRequest(c, "RSS 地址不合法")
	case errors.Is(err, service.ErrRecommendDuplicate):
		service.FailBadRequest(c, "该 RSS 地址已存在")
	case errors.Is(err, service.ErrImportInvalid):
		service.FailBadRequest(c, "导入数据格式不正确，请检查分类名称与 RSS 地址")
	default:
		slog.Error("后台接口内部错误", "path", c.FullPath(), "err", err)
		service.FailServer(c)
	}
}
