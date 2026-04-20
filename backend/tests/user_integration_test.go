package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"

	"github.com/stretchr/testify/assert"
)

func TestUser_Integration(t *testing.T) {
	testEmail := "docker_test@example.com"
	testPassword := "DockerPass123!"

	// 在测试开始前清理可能遗留的脏数据
	testDB.Where("email = ?", testEmail).Delete(&entity.User{})

	t.Run("在Docker中真实注册", func(t *testing.T) {
		req := dto.RegisterRequest{
			Email:    testEmail,
			Password: testPassword,
			Nickname: "Docker测试员",
		}
		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		request, _ := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(body))
		testRouter.ServeHTTP(w, request)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("测试数据是否真实落库", func(t *testing.T) {
		var user entity.User
		err := testDB.Where("email = ?", testEmail).First(&user).Error

		assert.NoError(t, err)
		assert.Equal(t, "Docker测试员", user.Nickname)
		assert.NotEqual(t, testPassword, user.PasswordHash) // 确保存入的是密文
	})

	t.Run("在Docker中真实登录", func(t *testing.T) {
		req := dto.LoginRequest{
			Email:    testEmail,
			Password: testPassword,
		}
		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		request, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
		testRouter.ServeHTTP(w, request)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "token") // 确保返回了 Token
	})
}
