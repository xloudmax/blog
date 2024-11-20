package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminAuthMiddleware 检查用户是否为管理员角色
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户角色
		role, exists := c.Get("role")

		// 记录访问尝试的日志，包括请求的URL和方法，以及用户信息（若有）
		if username, ok := c.Get("username"); ok {
			zap.S().Infof("Admin access attempt by user: %s, Role: %s, URL: %s, Method: %s",
				username, role, c.Request.URL.Path, c.Request.Method)
		} else {
			zap.S().Infof("Anonymous admin access attempt, Role: %s, URL: %s, Method: %s",
				role, c.Request.URL.Path, c.Request.Method)
		}

		// 检查角色是否为管理员
		if !exists || role != "admin" {
			zap.S().Warn("Unauthorized admin access attempt")
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can access this resource"})
			c.Abort()
			return
		}

		// 如果是管理员角色，则继续处理请求
		zap.S().Info("Admin access granted")
		c.Next()
	}
}
