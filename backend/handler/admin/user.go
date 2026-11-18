package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

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
