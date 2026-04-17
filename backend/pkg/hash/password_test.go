package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHash(t *testing.T) {
	password := "Secret123456!"

	// 1. 测试加密逻辑
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "加密后的值不应与原密码相同")

	// 2. 测试正确密码的比对
	assert.True(t, CheckPasswordHash(password, hash), "正确密码应该验证通过")

	// 3. 测试错误密码的比对
	assert.False(t, CheckPasswordHash("WrongPass", hash), "错误密码应该验证失败")
}
