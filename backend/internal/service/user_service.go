package service

import (
	"context"
	"errors"
	"time"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"
	"code-repo/pkg/hash"
	"code-repo/pkg/utils"

	"gorm.io/gorm"
)

type UserService struct {
	repo db.UserRepository
}

func NewUserService(repo db.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register 处理用户注册逻辑
func (s *UserService) Register(ctx context.Context, req dto.RegisterRequest) error {
	// 1. 检查邮箱是否已被注册
	_, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		return errors.New("该邮箱已被注册")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 2. 对密码进行哈希加密
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// 3. 构造用户实体并保存
	user := &entity.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Nickname:     req.Nickname,
		Username:     req.Username,
		Role:         "user",   // 默认角色
		Status:       "active", // 默认状态
	}

	return s.repo.Create(ctx, user)
}

// Login 处理用户登录并返回 Token
func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1. 根据邮箱查询用户
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号或密码错误")
		}
		return nil, err
	}

	// 2. 验证密码是否匹配
	if !hash.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("账号或密码错误")
	}

	// 3. 签发 JWT Token
	// 假设过期时间为 24 小时
	token, err := utils.GenerateToken(user.ID, user.Role, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	// 4. 封装 DTO 响应
	return &dto.LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		User: &dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetProfile 获取用户个人资料
func (s *UserService) GetProfile(ctx context.Context, userID uint64) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}
