// Package service 业务逻辑层：handler 只收参/回包，业务规则集中在此
package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一响应：{"code":0,"msg":"ok","data":...}
// 业务错误 code 直接复用 HTTP 状态码（400/401/404/500）

// OK 成功响应
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": data,
	})
}

// Fail 业务错误响应，status 同时作为 HTTP 状态码与响应 code
func Fail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{
		"code": status,
		"msg":  msg,
		"data": gin.H{},
	})
}

// FailBadRequest 400：参数或业务规则错误，msg 为可展示给用户的具体原因
func FailBadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, msg)
}

// FailUnauthorized 401：未登录或登录失效（前端收到后静默重新登录）
func FailUnauthorized(c *gin.Context) {
	Fail(c, http.StatusUnauthorized, "登录已失效，请重新登录")
}

// FailNotFound 404：资源不存在
func FailNotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, msg)
}

// FailServer 500：不向前端暴露内部错误细节
func FailServer(c *gin.Context) {
	Fail(c, http.StatusInternalServerError, "系统错误，请稍后再试")
}

// PageResult 分页统一结构
type PageResult struct {
	List    any   `json:"list"`
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	Size    int   `json:"size"`
	HasMore bool  `json:"hasMore"`
}

// NewPageResult 构造分页结果
func NewPageResult(list any, total int64, page, size int) PageResult {
	return PageResult{
		List:    list,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: int64(page*size) < total,
	}
}
