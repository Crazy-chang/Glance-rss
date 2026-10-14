package handler

import (
	"github.com/gin-gonic/gin"

	"glance-wx/backend/service"
)

// GET /api/discover/list
func (h *Handler) discoverList(c *gin.Context) {
	result, err := h.discover.List(uid(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, result)
}

// POST /api/discover/subscribe/:id
func (h *Handler) subscribeRecommend(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	dto, err := h.discover.Subscribe(uid(c), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, dto)
}
