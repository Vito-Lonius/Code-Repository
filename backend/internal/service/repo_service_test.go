package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"
	"code-repo/pkg/utils"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRepoService_CreateRepo(t *testing.T) {
	// 1. 初始化 Mock 控制器
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 2. 创建 Mock 依赖
	mockRepoDB := db.NewMockRepoRepository(ctrl)

	// 3. 准备临时测试目录模拟物理磁盘
	tempDir, err := os.MkdirTemp("", "repo_test_root_*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试完成后清理

	// 4. 模拟全局配置
	utils.Config.Git.RootPath = tempDir
	utils.Config.Git.DefaultBranch = "main"
	utils.Config.Server.Port = 8080

	// 5. 初始化 Service
	service := NewRepoService(mockRepoDB)

	t.Run("成功创建仓库", func(t *testing.T) {
		userID := uint64(1)
		req := dto.CreateRepoRequest{
			Name:        "test-project",
			Description: "unit test repo",
			IsPublic:    true,
		}

		// 定义 Mock 行为
		// 检查重名：返回 nil 表示不存在
		mockRepoDB.EXPECT().
			GetByNameAndOwner(req.Name, userID).
			Return(nil, nil)

		// 检查创建：接收任意 repository 实体并返回成功
		mockRepoDB.EXPECT().
			Create(gomock.Any()).
			Return(nil)

		// 检查获取详情：返回填充后的实体
		mockRepoDB.EXPECT().
			GetByID(gomock.Any()).
			Return(&entity.Repository{
				ID:            uint64(1),
				Name:          req.Name,
				OwnerID:       userID,
				Owner:         entity.User{Nickname: "Tester"},
				DefaultBranch: "main",
			}, nil)

		// 执行方法
		resp, err := service.CreateRepo(userID, req)

		// 断言
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Name, resp.Name)

		// 物理检查：验证 .git 目录和关键文件是否存在
		repoPath := filepath.Join(tempDir, "1", "test-project.git")
		assert.DirExists(t, repoPath)
		assert.FileExists(t, filepath.Join(repoPath, "HEAD"))
	})

	t.Run("数据库写入失败时应回滚物理目录", func(t *testing.T) {
		userID := uint64(2)
		req := dto.CreateRepoRequest{
			Name: "fail-repo",
		}

		// 定义 Mock 行为
		mockRepoDB.EXPECT().
			GetByNameAndOwner(req.Name, userID).
			Return(nil, nil)

		// 模拟数据库插入失败
		mockRepoDB.EXPECT().
			Create(gomock.Any()).
			Return(errors.New("database insert error"))

		// 执行方法
		_, err := service.CreateRepo(userID, req)

		// 断言：应该报错
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "仓库数据存库失败")

		// 核心断言：物理目录应该已经被 RemoveAll 清理掉
		repoPath := filepath.Join(tempDir, "2", "fail-repo.git")
		_, statErr := os.Stat(repoPath)
		assert.True(t, os.IsNotExist(statErr), "数据库失败后物理目录必须不存在")
	})

	t.Run("仓库名已存在应返回错误", func(t *testing.T) {
		userID := uint64(1)
		req := dto.CreateRepoRequest{Name: "existing-repo"}

		// 模拟数据库已存在该记录
		mockRepoDB.EXPECT().
			GetByNameAndOwner(req.Name, userID).
			Return(&entity.Repository{ID: uint64(100)}, nil)

		_, err := service.CreateRepo(userID, req)

		assert.Error(t, err)
		assert.Equal(t, "该用户下已存在同名仓库", err.Error())
	})
}

func TestRepoService_DeleteRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoDB := db.NewMockRepoRepository(ctrl)

	tempDir, _ := os.MkdirTemp("", "repo_delete_test_*")
	defer os.RemoveAll(tempDir)

	utils.Config.Git.RootPath = tempDir
	service := NewRepoService(mockRepoDB)

	t.Run("成功删除仓库并清理物理目录", func(t *testing.T) {
		userID := uint64(1)
		repoID := uint64(10)
		repoPath := filepath.Join(tempDir, "1", "to-delete.git")

		// 事前准备：创建物理目录
		_ = os.MkdirAll(repoPath, 0755)

		mockRepoDB.EXPECT().
			GetByID(repoID).
			Return(&entity.Repository{
				ID:      uint64(repoID),
				OwnerID: uint64(userID),
				Path:    repoPath,
			}, nil)

		mockRepoDB.EXPECT().
			Delete(repoID).
			Return(nil)

		err := service.DeleteRepo(userID, repoID)

		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// 检查物理目录（由于删除通常是异步或快速执行，测试中需确保清理逻辑生效）
		_, err = os.Stat(repoPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("非所有者删除应拒绝", func(t *testing.T) {
		userID := uint64(1)
		otherUserID := uint64(2)
		repoID := uint64(11)

		mockRepoDB.EXPECT().
			GetByID(repoID).
			Return(&entity.Repository{
				ID:      uint64(repoID),
				OwnerID: uint64(otherUserID),
			}, nil)

		err := service.DeleteRepo(userID, repoID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "权限不足")
	})
}
