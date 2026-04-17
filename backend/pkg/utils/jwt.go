// pkg/utils/jwt.go
package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 定义了载荷中存储的信息
type CustomClaims struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// 注意：在生产环境中，Secret 应该从 config.yaml 中读取
var jwtSecret = []byte("your_super_secret_key_change_me")

// GenerateToken 生成一个 JWT Token
func GenerateToken(userID uint64, role string, expireDuration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "code-repository",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析并验证 JWT Token
func ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
