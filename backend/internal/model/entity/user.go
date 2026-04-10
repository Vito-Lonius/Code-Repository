// internal/model/entity/user.go
package entity

import (
	"time"

	"gorm.io/gorm"
)

// User 用户实体，对应数据库 users 表
type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string         `gorm:"column:password_hash;size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:100;not null;default:''" json:"nickname"`
	AvatarURL    string         `gorm:"column:avatar_url;size:500;default:''" json:"avatar_url"`
	Role         string         `gorm:"size:20;not null;default:'user'" json:"role"`     // user/admin
	Status       string         `gorm:"size:20;not null;default:'active'" json:"status"` // active/banned
	JwtVersion   int            `gorm:"column:jwt_version;default:0;not null" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
