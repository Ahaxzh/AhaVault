package middleware

import (
	"net/http"
	"strings"

	"ahavault/server/internal/services"
	"github.com/gin-gonic/gin"
)

// Auth JWT 认证中间件
func Auth(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 token
		claims, err := userService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户 ID 存储到上下文
		userID := (*claims)["user_id"].(string)
		c.Set("user_id", userID)

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行普通认证
		Auth(userService)(c)
		if c.IsAborted() {
			return
		}

		// 检查是否是管理员
		userID := GetUserID(c)
		user, err := userService.GetUserByID(userID)
		if err != nil || !user.IsAdmin() {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}
