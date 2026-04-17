// pkg/hash/password.go
package hash

import "golang.org/x/crypto/bcrypt"

// HashPassword 使用 bcrypt 算法对密码进行哈希
func HashPassword(password string) (string, error) {
	// GenerateFromPassword 默认开销为 10，包含加盐过程
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 比较明文密码与哈希值是否匹配
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
