package db

import (
	"code-repo/internal/model/entity"

	"gorm.io/gorm"
)

type FileRepository interface {
	Create(file *entity.File) error
	GetByID(id uint64) (*entity.File, error)
	GetByPath(repoID uint64, path string) (*entity.File, error)
	ListByRepo(repoID uint64, parentPath string) ([]entity.File, error)
	Update(file *entity.File) error
	Delete(id uint64) error
	DeleteByRepo(repoID uint64) error
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(file *entity.File) error {
	return r.db.Create(file).Error
}

func (r *fileRepository) GetByID(id uint64) (*entity.File, error) {
	var file entity.File
	err := r.db.First(&file, id).Error
	return &file, err
}

func (r *fileRepository) GetByPath(repoID uint64, path string) (*entity.File, error) {
	var file entity.File
	err := r.db.Where("repo_id = ? AND path = ?", repoID, path).First(&file).Error
	return &file, err
}

func (r *fileRepository) ListByRepo(repoID uint64, parentPath string) ([]entity.File, error) {
	var files []entity.File
	query := r.db.Where("repo_id = ?", repoID)
	if parentPath != "" {
		query = query.Where("path LIKE ?", parentPath+"%")
	}
	err := query.Order("is_dir DESC, file_name ASC").Find(&files).Error
	return files, err
}

func (r *fileRepository) Update(file *entity.File) error {
	return r.db.Save(file).Error
}

func (r *fileRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.File{}, id).Error
}

func (r *fileRepository) DeleteByRepo(repoID uint64) error {
	return r.db.Where("repo_id = ?", repoID).Delete(&entity.File{}).Error
}
