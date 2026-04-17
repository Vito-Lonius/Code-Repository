package service

import (
	"context"
	"testing"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// 定义 Mock 对象
type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Create(ctx context.Context, u *entity.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *MockUserRepo) GetByEmail(ctx context.Context, e string) (*entity.User, error) {
	args := m.Called(ctx, e)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *MockUserRepo) GetByID(ctx context.Context, id uint64) (*entity.User, error) { return nil, nil }
func (m *MockUserRepo) Update(ctx context.Context, u *entity.User) error             { return nil }
func (m *MockUserRepo) Delete(ctx context.Context, id uint64) error                  { return nil }

func TestUserService_Register(t *testing.T) {
	mockRepo := new(MockUserRepo)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	t.Run("邮箱冲突测试", func(t *testing.T) {
		req := dto.RegisterRequest{Email: "exists@test.com"}
		// 模拟数据库能查到该邮箱
		mockRepo.On("GetByEmail", ctx, req.Email).Return(&entity.User{}, nil).Once()

		err := svc.Register(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, "该邮箱已被注册", err.Error())
	})

	t.Run("注册成功测试", func(t *testing.T) {
		req := dto.RegisterRequest{Email: "new@test.com", Password: "password123"}
		// 模拟数据库查不到该邮箱
		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("Create", ctx, mock.Anything).Return(nil).Once()

		err := svc.Register(ctx, req)
		assert.NoError(t, err)
	})
}
