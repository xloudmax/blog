package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// 定义 JWT 密钥
var jwtSecret = []byte("JNU_technicians_club")

// JWTAuthMiddleware 返回一个 JWT 认证的中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization 字段
		tokenString := c.GetHeader("Authorization")

		// 检查 Authorization 字段是否存在，并且是否以 "Bearer " 开头
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token not provided or invalid"})
			c.Abort()
			return
		}

		// 提取 token 字符串
		tokenString = tokenString[len("Bearer "):]

		// 解析 token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		// 检查 token 是否有效
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 提取用户信息
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// 将用户信息存储到上下文中，便于控制器中使用
			if username, exists := claims["username"]; exists {
				c.Set("username", username)
			}
			if role, exists := claims["role"]; exists {
				c.Set("role", role)
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT claims"})
			c.Abort()
			return
		}

		// 如果 token 有效，继续处理请求
		c.Next()
	}
}
