package handler

import (
	"github.com/gin-gonic/gin"

	"glance-wx/backend/service"
)

// GET /api/source/list
func (h *Handler) sourceList(c *gin.Context) {
	list, err := h.source.List(uid(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, gin.H{"list": list})
}

type sourceAddRequest struct {
	FeedURL string `json:"feed_url"`
}

// POST /api/source/add
func (h *Handler) sourceAdd(c *gin.Context) {
	var req sourceAddRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.FeedURL == "" {
		service.FailBadRequest(c, "请输入 RSS 地址")
		return
	}
	dto, err := h.source.Add(c.Request.Context(), uid(c), req.FeedURL)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, dto)
}

// POST /api/source/delete/:id
func (h *Handler) sourceDelete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.source.Delete(uid(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, gin.H{"id": id})
}
