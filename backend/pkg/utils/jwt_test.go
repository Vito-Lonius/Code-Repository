package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	userID := uint64(101)
	role := "admin"
	duration := time.Hour

	// 1. 测试生成 Token
	token, err := GenerateToken(userID, role, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 2. 测试解析有效 Token
	claims, err := ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)

	// 3. 测试解析过期 Token
	expiredToken, _ := GenerateToken(userID, role, -time.Hour) // 设置一小时前过期
	_, err = ParseToken(expiredToken)
	assert.Error(t, err, "过期的 Token 应该返回错误")
}
