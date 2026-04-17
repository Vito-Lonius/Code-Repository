package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code-repo/internal/model/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_Register_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &UserHandler{svc: nil} // 仅测试校验逻辑，不调 service
	r := gin.New()
	r.POST("/register", handler.Register)

	t.Run("邮箱格式错误应返回400", func(t *testing.T) {
		body, _ := json.Marshal(dto.RegisterRequest{
			Email:    "invalid-email",
			Password: "short",
		})
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
