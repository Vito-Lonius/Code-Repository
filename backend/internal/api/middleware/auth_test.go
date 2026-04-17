package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code-repo/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("未携带授权头应返回401", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/protected", func(c *gin.Context) { c.Status(200) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "请求未携带认证令牌")
	})

	t.Run("携带合法Token应通过并提取ID", func(t *testing.T) {
		expectedID := uint64(888)
		token, _ := utils.GenerateToken(expectedID, "user", time.Hour)

		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/protected", func(c *gin.Context) {
			val, _ := c.Get("userID")
			assert.Equal(t, expectedID, val.(uint64))
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
