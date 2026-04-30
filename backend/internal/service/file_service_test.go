package service

import (
	"context"
	"errors"
	"testing"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestFileService_CreateDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("成功创建目录", func(t *testing.T) {
		userID := uint64(1)
		req := dto.CreateDirRequest{
			RepoID:  1,
			Path:    "/",
			DirName: "src",
		}

		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1, OwnerID: 1}, nil)
		mockFileRepo.EXPECT().GetByPath(uint64(1), "/src").Return(nil, gorm.ErrRecordNotFound)
		mockFileRepo.EXPECT().Create(gomock.Any()).Return(nil)

		resp, err := svc.CreateDir(context.Background(), userID, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "src", resp.FileName)
		assert.True(t, resp.IsDir)
	})

	t.Run("目录已存在应报错", func(t *testing.T) {
		userID := uint64(1)
		req := dto.CreateDirRequest{
			RepoID:  1,
			Path:    "/",
			DirName: "src",
		}

		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1, OwnerID: 1}, nil)
		mockFileRepo.EXPECT().GetByPath(uint64(1), "/src").Return(&entity.File{ID: 1, FileName: "src", IsDir: true}, nil)

		_, err := svc.CreateDir(context.Background(), userID, req)
		assert.Error(t, err)
		assert.Equal(t, "目录已存在", err.Error())
	})

	t.Run("仓库不存在应报错", func(t *testing.T) {
		req := dto.CreateDirRequest{RepoID: 999, Path: "/", DirName: "src"}

		mockRepoRepo.EXPECT().GetByID(uint64(999)).Return(nil, errors.New("not found"))

		_, err := svc.CreateDir(context.Background(), 1, req)
		assert.Error(t, err)
		assert.Equal(t, "仓库不存在", err.Error())
	})
}

func TestFileService_GetFileDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("获取文件详情成功", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "main.go", Path: "/main.go",
			IsDir: false, MimeType: "text/x-go", FileSize: 2048, Status: "completed", UploaderID: 1,
		}, nil)

		resp, err := svc.GetFileDetail(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, "main.go", resp.FileName)
		assert.Equal(t, "text/x-go", resp.MimeType)
	})

	t.Run("文件不存在应报错", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(999)).Return(nil, errors.New("not found"))

		_, err := svc.GetFileDetail(context.Background(), 999)
		assert.Error(t, err)
		assert.Equal(t, "文件不存在", err.Error())
	})
}

func TestFileService_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("权限不足应拒绝删除", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "test.txt", UploaderID: 1,
		}, nil)
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{
			ID: 1, OwnerID: 2,
		}, nil)

		err := svc.DeleteFile(context.Background(), 3, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "权限不足")
	})
}
