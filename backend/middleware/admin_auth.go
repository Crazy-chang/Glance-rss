package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/service"
)

// 后台上下文键（与小程序用户区分）
const (
	ContextAdminID       = "currentAdminID"
	ContextAdminUsername = "currentAdminUsername"
)

// AdminAuth 后台鉴权：只接受后台独立签发的令牌，小程序用户令牌无法通过
func AdminAuth(auth *service.AdminAuthService) gin.HandlerFunc {
	const bearerPrefix = "Bearer "
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, bearerPrefix) {
			service.FailUnauthorized(c)
			c.Abort()
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
		if token == "" {
			service.FailUnauthorized(c)
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(token)
		if err != nil {
			service.FailUnauthorized(c)
			c.Abort()
			return
		}
		c.Set(ContextAdminID, claims.AdminID)
		c.Set(ContextAdminUsername, claims.Username)
		c.Next()
	}
}

// CurrentAdmin 从上下文取当前管理员
func CurrentAdmin(c *gin.Context) service.AdminActor {
	id, _ := c.Get(ContextAdminID)
	adminID, _ := id.(int64)
	name, _ := c.Get(ContextAdminUsername)
	username, _ := name.(string)
	return service.AdminActor{ID: adminID, Username: username}
}
