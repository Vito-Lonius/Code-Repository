package entity

import (
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string         `gorm:"uniqueIndex:idx_owner_repo_name;size:255;not null" json:"name"`
	Visibility      string         `gorm:"size:255;not null" json:"visibility"` // public/private
	OwnerID         uint64         `gorm:"uniqueIndex:idx_owner_repo_name;not null" json:"owner_id"`
	CollaboratorsID []uint64       `gorm:"type:json" json:"collaborators_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
