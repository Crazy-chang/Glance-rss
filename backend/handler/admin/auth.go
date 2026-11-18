package admin

import (
	"github.com/gin-gonic/gin"

	"glance-wx/backend/middleware"
	"glance-wx/backend/service"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// POST /api/admin/login
func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if !bindJSON(c, &req) {
		return
	}
	token, actor, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}
	service.WriteAdminLog(h.db, *actor, "登录后台", "管理员登录成功")
	service.OK(c, gin.H{"token": token, "username": actor.Username})
}

// POST /api/admin/logout （令牌无状态，前端清除本地令牌即可，这里仅记日志）
func (h *Handler) logout(c *gin.Context) {
	actor := middleware.CurrentAdmin(c)
	service.WriteAdminLog(h.db, actor, "退出后台", "管理员退出登录")
	service.OK(c, gin.H{"ok": true})
}
