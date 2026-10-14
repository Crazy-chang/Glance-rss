package handler

import (
	"github.com/gin-gonic/gin"

	"glance-wx/backend/service"
)

type loginRequest struct {
	Code string `json:"code"`
}

// POST /api/auth/login
func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		service.FailBadRequest(c, "缺少登录凭证")
		return
	}

	result, err := h.auth.Login(c.Request.Context(), req.Code)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, result)
}
