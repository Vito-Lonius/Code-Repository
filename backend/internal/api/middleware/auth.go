package middleware

import (
	"net/http"
	"strings"

	"code-repo/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 身份认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请求未携带认证令牌"})
			c.Abort() // 终止后续处理逻辑
			return
		}

		// 2. 检查格式是否为 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误，请使用 Bearer 格式"})
			c.Abort()
			return
		}

		// 3. 解析并验证 Token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的令牌"})
			c.Abort()
			return
		}

		// 4. 将解析出的关键信息存入上下文，方便后续 Handler 获取
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)

		c.Next() // 继续执行后续逻辑
	}
}

// AdminMiddleware 权限检查中间件（在 AuthMiddleware 之后使用）
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，仅限管理员访问"})
			c.Abort()
			return
		}
		c.Next()
	}
}
