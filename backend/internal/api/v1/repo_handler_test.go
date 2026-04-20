package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code-repo/internal/model/dto"
	"code-repo/internal/service" // 确保这里有 MockService

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRepoHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. 创建 MockService (需要先执行 mockgen 生成 service 的 mock)
	// 注意：如果没生成 service mock，建议先运行：
	// mockgen -source=internal/service/repo_service.go -destination=internal/service/mock_service.go
	mockService := service.NewMockRepoService(ctrl)
	handler := NewRepoHandler(mockService)

	t.Run("无效的JSON输入应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 模拟发送空 body
		c.Request, _ = http.NewRequest("POST", "/api/v1/repos", bytes.NewBufferString("{invalid}"))

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("正常创建请求", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 模拟中间件设置 userID
		c.Set("userID", uint64(1))

		reqBody := dto.CreateRepoRequest{Name: "api-test"}
		jsonStr, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("POST", "/api/v1/repos", bytes.NewBuffer(jsonStr))
		c.Request.Header.Set("Content-Type", "application/json")

		// 设定 MockService 期望
		mockService.EXPECT().CreateRepo(uint64(1), gomock.Any()).Return(&dto.RepoResponse{Name: "api-test"}, nil)

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "api-test")
	})
}
