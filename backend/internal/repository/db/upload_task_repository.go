package db

import (
	"code-repo/internal/model/entity"

	"gorm.io/gorm"
)

type UploadTaskRepository interface {
	Create(task *entity.UploadTask) error
	GetByUploadID(uploadID string) (*entity.UploadTask, error)
	UpdateUploadedChunks(uploadID string, uploadedChunks int, indices string) error
	UpdateStatus(uploadID string, status string) error
	Delete(uploadID string) error
}

type uploadTaskRepository struct {
	db *gorm.DB
}

func NewUploadTaskRepository(db *gorm.DB) UploadTaskRepository {
	return &uploadTaskRepository{db: db}
}

func (r *uploadTaskRepository) Create(task *entity.UploadTask) error {
	return r.db.Create(task).Error
}

func (r *uploadTaskRepository) GetByUploadID(uploadID string) (*entity.UploadTask, error) {
	var task entity.UploadTask
	err := r.db.Where("upload_id = ?", uploadID).First(&task).Error
	return &task, err
}

func (r *uploadTaskRepository) UpdateUploadedChunks(uploadID string, uploadedChunks int, indices string) error {
	return r.db.Model(&entity.UploadTask{}).Where("upload_id = ?", uploadID).
		Updates(map[string]interface{}{"uploaded_chunks": uploadedChunks, "uploaded_chunk_indices": indices}).Error
}

func (r *uploadTaskRepository) UpdateStatus(uploadID string, status string) error {
	return r.db.Model(&entity.UploadTask{}).Where("upload_id = ?", uploadID).
		Updates(map[string]interface{}{"status": status}).Error
}

func (r *uploadTaskRepository) Delete(uploadID string) error {
	return r.db.Where("upload_id = ?", uploadID).Delete(&entity.UploadTask{}).Error
}
