package entity

import (
	"time"

	"gorm.io/gorm"
)

// Repository 代表代码仓库的数据库实体
type Repository struct {
	ID          uint64 `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex:idx_owner_repo" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`

	// 关联 User
	OwnerID uint64 `gorm:"not null;uniqueIndex:idx_owner_repo" json:"owner_id"`
	Owner   User   `gorm:"foreignKey:OwnerID" json:"owner"`

	// 仓库属性
	IsPublic      bool   `gorm:"default:true" json:"is_public"`
	Path          string `gorm:"type:varchar(255);not null" json:"path"` // 磁盘上的物理路径
	DefaultBranch string `gorm:"type:varchar(50);default:main" json:"default_branch"`

	// 统计信息
	StarCount int `gorm:"default:0" json:"star_count"`
	ForkCount int `gorm:"default:0" json:"fork_count"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Repository) TableName() string {
	return "repositories"
}
