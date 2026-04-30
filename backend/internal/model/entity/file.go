package entity

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	RepoID    uint64         `gorm:"not null;index" json:"repo_id"`
	FileName  string         `gorm:"type:varchar(255);not null" json:"file_name"`
	Path      string         `gorm:"type:varchar(500);not null" json:"path"`
	IsDir     bool           `gorm:"default:false" json:"is_dir"`
	MimeType  string         `gorm:"type:varchar(100);default:''" json:"mime_type"`
	FileSize  int64          `gorm:"default:0" json:"file_size"`
	ObjectKey string         `gorm:"type:varchar(500);default:''" json:"object_key"`
	UploadID  string         `gorm:"type:varchar(100);default:''" json:"upload_id"`
	ChunkCount int           `gorm:"default:0" json:"chunk_count"`
	UploadedChunks int       `gorm:"default:0" json:"uploaded_chunks"`
	Status    string         `gorm:"type:varchar(20);default:completed" json:"status"`
	UploaderID uint64        `gorm:"not null" json:"uploader_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (File) TableName() string {
	return "files"
}

type UploadTask struct {
	ID                  uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UploadID            string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"upload_id"`
	RepoID              uint64         `gorm:"not null;index" json:"repo_id"`
	FileName            string         `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath            string         `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize            int64          `gorm:"not null" json:"file_size"`
	ChunkSize           int64          `gorm:"not null" json:"chunk_size"`
	TotalChunks         int            `gorm:"not null" json:"total_chunks"`
	UploadedChunks      int            `gorm:"default:0" json:"uploaded_chunks"`
	UploadedChunkIndices string        `gorm:"type:text;default:''" json:"uploaded_chunk_indices"`
	MimeType            string         `gorm:"type:varchar(100);default:''" json:"mime_type"`
	Status              string         `gorm:"type:varchar(20);default:uploading" json:"status"`
	UploaderID          uint64         `gorm:"not null" json:"uploader_id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UploadTask) TableName() string {
	return "upload_tasks"
}
