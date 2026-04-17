package db

import (
	"code-repo/internal/model/entity"
	"context"

	"gorm.io/gorm"
)

// UserRepository 定义了用户数据访问的接口规范
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint64) error
}

// userRepo 是 UserRepository 接口的 GORM 实现
type userRepo struct {
	db *gorm.DB
}

// NewUserRepository 实例化用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// Create 在数据库中插入一条新的用户记录
func (r *userRepo) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByEmail 根据邮箱地址查询用户信息
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID 根据用户 ID 查询用户信息
func (r *userRepo) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *userRepo) Update(ctx context.Context, user *entity.User) error {
	// Updates 会根据 struct 的非零值字段进行更新
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 软删除用户记录
func (r *userRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}
