package handler

import (
	"github.com/gin-gonic/gin"

	"glance-wx/backend/service"
)

// GET /api/setting/info
func (h *Handler) settingInfo(c *gin.Context) {
	dto, err := h.setting.Info(uid(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, dto)
}

// POST /api/setting/update
func (h *Handler) settingUpdate(c *gin.Context) {
	var params service.UpdateSettingParams
	if err := c.ShouldBindJSON(&params); err != nil {
		service.FailBadRequest(c, "参数错误")
		return
	}
	dto, err := h.setting.Update(uid(c), params)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	service.OK(c, dto)
}
