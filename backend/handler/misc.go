package handler

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/service"
)

// POST /api/refresh/start 异步触发当前用户全部订阅源刷新
func (h *Handler) refreshStart(c *gin.Context) {
	// RefreshUserSources 内部并发执行并立即返回；
	// 用独立 context，请求返回后后台抓取继续执行
	h.sched.RefreshUserSources(context.Background(), uid(c))
	service.OK(c, gin.H{"refreshing": true})
}

// GET /api/favicon/:sourceId 返回订阅源本地 favicon
func (h *Handler) favicon(c *gin.Context) {
	sourceID, ok := parseIDParam(c, "sourceId")
	if !ok {
		return
	}
	src, err := h.source.GetOwned(uid(c), sourceID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	if src.FaviconPath == "" {
		service.FailNotFound(c, "图标不存在")
		return
	}
	if _, err := os.Stat(src.FaviconPath); err != nil {
		service.FailNotFound(c, "图标不存在")
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.File(src.FaviconPath)
}

// GET /api/img-proxy?url=&referer= 正文图片代理（白名单 + SSRF 防护）
func (h *Handler) imgProxy(c *gin.Context) {
	target := c.Query("url")
	if target == "" {
		service.FailBadRequest(c, "缺少图片地址")
		return
	}
	referer := c.Query("referer")

	img, err := h.proxy.Fetch(c.Request.Context(), target, referer)
	if err != nil {
		// 把代理层的具体错误原因写进响应头，方便前端排查
		c.Header("X-Proxy-Error", err.Error())
		service.Fail(c, 403, "图片无法访问: "+err.Error())
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(200, img.ContentType, img.Body)
}
