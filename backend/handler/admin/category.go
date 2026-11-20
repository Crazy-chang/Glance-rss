package admin

import (
	"github.com/gin-gonic/gin"

	"glance-wx/backend/middleware"
	"glance-wx/backend/service"
)

type categoryRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type reorderRequest struct {
	IDs []int64 `json:"ids"`
}

// statusRequest 停用/启用请求体
type statusRequest struct {
	ID      int64 `json:"id"`
	Enabled bool  `json:"enabled"`
}

// GET /api/admin/category/list
func (h *Handler) categoryList(c *gin.Context) {
	list, err := h.recommend.ListCategories()
	if err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, list)
}

// POST /api/admin/category/create
func (h *Handler) categoryCreate(c *gin.Context) {
	var req categoryRequest
	if !bindJSON(c, &req) {
		return
	}
	cat, err := h.recommend.CreateCategory(middleware.CurrentAdmin(c), req.Name, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, cat)
}

// POST /api/admin/category/update
func (h *Handler) categoryUpdate(c *gin.Context) {
	var req categoryRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.ID <= 0 {
		service.FailBadRequest(c, "参数错误")
		return
	}
	if err := h.recommend.UpdateCategory(middleware.CurrentAdmin(c), req.ID, req.Name, req.Description); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}

// POST /api/admin/category/delete
func (h *Handler) categoryDelete(c *gin.Context) {
	var req idRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.ID <= 0 {
		service.FailBadRequest(c, "参数错误")
		return
	}
	if err := h.recommend.DeleteCategory(middleware.CurrentAdmin(c), req.ID); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}

// POST /api/admin/category/status
func (h *Handler) categoryStatus(c *gin.Context) {
	var req statusRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.ID <= 0 {
		service.FailBadRequest(c, "参数错误")
		return
	}
	if err := h.recommend.SetCategoryEnabled(middleware.CurrentAdmin(c), req.ID, req.Enabled); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}

// POST /api/admin/category/reorder
func (h *Handler) categoryReorder(c *gin.Context) {
	var req reorderRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.recommend.ReorderCategories(middleware.CurrentAdmin(c), req.IDs); err != nil {
		handleError(c, err)
		return
	}
	service.OK(c, gin.H{"ok": true})
}
