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

// GET /api/admin/system/config 读取全局配置
func (h *Handler) systemConfig(c *gin.Context) {
	cfg, err := h.recommend.GetAppConfig()
	if err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"batch_read_limit": cfg.BatchReadLimit})
}

// configRequest 全局配置更新请求体
type configRequest struct {
	BatchReadLimit int `json:"batch_read_limit"`
}

// POST /api/admin/system/config 更新全局配置
func (h *Handler) systemConfigUpdate(c *gin.Context) {
	var req configRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.recommend.UpdateBatchReadLimit(middleware.CurrentAdmin(c), req.BatchReadLimit); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}
