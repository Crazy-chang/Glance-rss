package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/middleware"
	"glance-wx/backend/service"
)

type sourceUpdateRequest struct {
	ID int64 `json:"id"`
	service.AdminSourceInput
}

// GET /api/admin/recommend/list?category_id=1
func (h *Handler) recommendList(c *gin.Context) {
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	list, err := h.recommend.ListSources(categoryID)
	if err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, list)
}

// POST /api/admin/recommend/create
func (h *Handler) recommendCreate(c *gin.Context) {
	var req service.AdminSourceInput
	if !bindJSON(c, &req) {
		return
	}
	source, err := h.recommend.CreateSource(middleware.CurrentAdmin(c), req)
	if err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, source)
}

// POST /api/admin/recommend/update
func (h *Handler) recommendUpdate(c *gin.Context) {
	var req sourceUpdateRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.ID <= 0 {
		service.FailBadRequest(c, "参数错误")
		return
	}
	if err := h.recommend.UpdateSource(middleware.CurrentAdmin(c), req.ID, req.AdminSourceInput); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}

// POST /api/admin/recommend/delete
func (h *Handler) recommendDelete(c *gin.Context) {
	var req idRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.ID <= 0 {
		service.FailBadRequest(c, "参数错误")
		return
	}
	if err := h.recommend.DeleteSource(middleware.CurrentAdmin(c), req.ID); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}

// POST /api/admin/recommend/reorder
func (h *Handler) recommendReorder(c *gin.Context) {
	var req reorderRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.recommend.ReorderSources(middleware.CurrentAdmin(c), req.IDs); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}

// GET /api/admin/recommend/export
func (h *Handler) recommendExport(c *gin.Context) {
	data, err := h.recommend.Export()
	if err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, data)
}

// POST /api/admin/recommend/import
func (h *Handler) recommendImport(c *gin.Context) {
	var req service.RecommendExport
	if !bindJSON(c, &req) {
		return
	}
	if err := h.recommend.Import(middleware.CurrentAdmin(c), req); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}
