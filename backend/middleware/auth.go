// Package middleware HTTP 中间件
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"glance-wx/backend/service"
)

// ContextUserID gin.Context 中存放当前用户 ID 的键
const ContextUserID = "currentUserID"

// Auth JWT 鉴权：优先从 Authorization: Bearer <token> 取 token；
// Image 组件无法自定义请求头，favicon/img-proxy 允许通过 ?token= 传递
func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
		if token == "" {
			token = strings.TrimSpace(c.Query("token"))
		}
		if token == "" {
			service.FailUnauthorized(c)
			c.Abort()
			return
		}

		userID, err := auth.ParseToken(token)
		if err != nil {
			service.FailUnauthorized(c)
			c.Abort()
			return
		}
		c.Set(ContextUserID, userID)
		c.Next()
	}
}

// CurrentUserID 从上下文取当前登录用户 ID
func CurrentUserID(c *gin.Context) int64 {
	v, _ := c.Get(ContextUserID)
	id, _ := v.(int64)
	return id
}
