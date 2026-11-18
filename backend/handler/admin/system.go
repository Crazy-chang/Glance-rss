package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/middleware"
	"glance-wx/backend/service"
)

// POST /api/admin/system/reset
func (h *Handler) systemReset(c *gin.Context) {
	if err := h.recommend.Reset(middleware.CurrentAdmin(c)); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}

// GET /api/admin/system/logs?limit=100
func (h *Handler) systemLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	logs, err := service.ListAdminLogs(h.db, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, logs)
}
