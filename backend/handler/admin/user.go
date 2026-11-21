package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/middleware"
	"glance-wx/backend/service"
)

// GET /api/admin/user/list?page=1&size=20&min_sources=1
func (h *Handler) userList(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	minSources, _ := strconv.ParseInt(c.Query("min_sources"), 10, 64)
	result, err := service.ListAdminUsers(h.db, page, size, minSources)
	if err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, result)
}

// POST /api/admin/user/batch-read 开启/关闭用户的一键阅读
func (h *Handler) userBatchRead(c *gin.Context) {
	var req statusRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.ID <= 0 {
		service.FailBadRequest(c, "参数错误")
		return
	}
	if err := service.SetUserBatchRead(h.db, middleware.CurrentAdmin(c), req.ID, req.Enabled); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}
