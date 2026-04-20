package db

import (
	"code-repo/internal/model/entity"

	"gorm.io/gorm"
)

// RepoRepository 定义了仓库相关的数据库操作接口
type RepoRepository interface {
	Create(repo *entity.Repository) error
	GetByID(id uint64) (*entity.Repository, error)
	GetByNameAndOwner(name string, ownerID uint64) (*entity.Repository, error)
	ListByOwner(ownerID uint64, page, pageSize int, keyword string) ([]entity.Repository, int64, error)
	Update(repo *entity.Repository) error
	Delete(id uint64) error
	UpdateFields(repoID uint64, updates map[string]interface{}) error
}

type repoRepository struct {
	db *gorm.DB
}

// NewRepoRepository 创建一个新的仓库持久层实例
func NewRepoRepository(db *gorm.DB) RepoRepository {
	return &repoRepository{db: db}
}

// Create 插入新仓库记录
func (r *repoRepository) Create(repo *entity.Repository) error {
	return r.db.Create(repo).Error
}

// GetByID 根据 ID 查询仓库（包含所有者信息）
func (r *repoRepository) GetByID(id uint64) (*entity.Repository, error) {
	var repo entity.Repository
	err := r.db.Preload("Owner").First(&repo, id).Error
	return &repo, err
}

// GetByNameAndOwner 根据仓库名和所有者查询（用于唯一性检查）
func (r *repoRepository) GetByNameAndOwner(name string, ownerID uint64) (*entity.Repository, error) {
	var repo entity.Repository
	err := r.db.Where("name = ? AND owner_id = ?", name, ownerID).First(&repo).Error
	return &repo, err
}

// ListByOwner 分页查询用户的仓库列表
func (r *repoRepository) ListByOwner(ownerID uint64, page, pageSize int, keyword string) ([]entity.Repository, int64, error) {
	var repos []entity.Repository
	var total int64

	db := r.db.Model(&entity.Repository{}).Where("owner_id = ?", ownerID)

	// 如果有关键词搜索
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}

	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&repos).Error

	return repos, total, err
}

// Update 更新仓库信息
func (r *repoRepository) Update(repo *entity.Repository) error {
	return r.db.Save(repo).Error
}

// Delete 软删除仓库记录
func (r *repoRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.Repository{}, id).Error
}

// UpdateFields 支持部分字段更新，防止覆盖掉不需要修改的字段
func (r *repoRepository) UpdateFields(repoID uint64, updates map[string]interface{}) error {
	return r.db.Model(&entity.Repository{}).Where("id = ?", repoID).Updates(updates).Error
}

// ListPublicRepos 获取公开仓库流（用于探索页面）
func (r *repoRepository) ListPublicRepos(page, pageSize int) ([]entity.Repository, int64, error) {
	var repos []entity.Repository
	var total int64
	db := r.db.Model(&entity.Repository{}).Where("is_public = ?", true)

	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at desc").Preload("Owner").Find(&repos).Error
	return repos, total, err
}
