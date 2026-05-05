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

	t.Run("文件不存在应报错", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(999)).Return(nil, errors.New("not found"))

		err := svc.DeleteFile(context.Background(), 1, 999)
		assert.Error(t, err)
		assert.Equal(t, "文件不存在", err.Error())
	})

	t.Run("仓库所有者可删除文件", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "test.txt", IsDir: false, ObjectKey: "repos/1/test.txt", UploaderID: 1,
		}, nil)
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{
			ID: 1, OwnerID: 1,
		}, nil)
		mockFileRepo.EXPECT().Delete(uint64(1)).Return(nil)

		err := svc.DeleteFile(context.Background(), 1, 1)
		assert.NoError(t, err)
	})
}

func TestFileService_ListFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("成功列出仓库文件", func(t *testing.T) {
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1}, nil)
		mockFileRepo.EXPECT().ListByRepo(uint64(1), "").Return([]entity.File{
			{ID: 1, FileName: "src", IsDir: true, RepoID: 1},
			{ID: 2, FileName: "main.go", IsDir: false, RepoID: 1},
		}, nil)

		resp, err := svc.ListFiles(context.Background(), 1, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), resp.Total)
		assert.Len(t, resp.Items, 2)
	})

	t.Run("仓库不存在应报错", func(t *testing.T) {
		mockRepoRepo.EXPECT().GetByID(uint64(999)).Return(nil, errors.New("not found"))

		_, err := svc.ListFiles(context.Background(), 999, "")
		assert.Error(t, err)
		assert.Equal(t, "仓库不存在", err.Error())
	})

	t.Run("按目录路径过滤", func(t *testing.T) {
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1}, nil)
		mockFileRepo.EXPECT().ListByRepo(uint64(1), "/src").Return([]entity.File{
			{ID: 3, FileName: "main.go", IsDir: false, RepoID: 1},
		}, nil)

		resp, err := svc.ListFiles(context.Background(), 1, "/src")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
	})
}

func TestFileService_RenameFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("权限不足应拒绝重命名", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "old.go", Path: "/old.go", UploaderID: 1,
		}, nil)
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{
			ID: 1, OwnerID: 2,
		}, nil)

		_, err := svc.RenameFile(context.Background(), 3, 1, "new.go")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "权限不足")
	})

	t.Run("文件不存在应报错", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(999)).Return(nil, errors.New("not found"))

		_, err := svc.RenameFile(context.Background(), 1, 999, "new.go")
		assert.Error(t, err)
		assert.Equal(t, "文件不存在", err.Error())
	})

	t.Run("重命名目录成功", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "olddir", Path: "/olddir", IsDir: true, UploaderID: 1,
		}, nil)
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1, OwnerID: 1}, nil)
		mockFileRepo.EXPECT().GetByPath(uint64(1), "/newdir").Return(nil, gorm.ErrRecordNotFound)
		mockFileRepo.EXPECT().Update(gomock.Any()).Return(nil)

		resp, err := svc.RenameFile(context.Background(), 1, 1, "newdir")
		assert.NoError(t, err)
		assert.Equal(t, "newdir", resp.FileName)
		assert.Equal(t, "/newdir", resp.Path)
	})
}

func TestFileService_MoveFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("权限不足应拒绝移动", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "file.go", Path: "/file.go", UploaderID: 1,
		}, nil)
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{
			ID: 1, OwnerID: 2,
		}, nil)

		_, err := svc.MoveFile(context.Background(), 3, 1, "/new/file.go")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "权限不足")
	})

	t.Run("文件不存在应报错", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(999)).Return(nil, errors.New("not found"))

		_, err := svc.MoveFile(context.Background(), 1, 999, "/new/path")
		assert.Error(t, err)
		assert.Equal(t, "文件不存在", err.Error())
	})

	t.Run("目标路径已存在同名文件应报错", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "file.go", Path: "/file.go", IsDir: false, UploaderID: 1,
		}, nil)
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1, OwnerID: 1}, nil)
		mockFileRepo.EXPECT().GetByPath(uint64(1), "/dest/file.go").Return(&entity.File{ID: 2, FileName: "file.go"}, nil)

		_, err := svc.MoveFile(context.Background(), 1, 1, "/dest/file.go")
		assert.Error(t, err)
		assert.Equal(t, "目标路径已存在同名文件", err.Error())
	})

	t.Run("移动目录成功", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "src", Path: "/src", IsDir: true, UploaderID: 1,
		}, nil)
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1, OwnerID: 1}, nil)
		mockFileRepo.EXPECT().GetByPath(uint64(1), "/lib/src").Return(nil, gorm.ErrRecordNotFound)
		mockFileRepo.EXPECT().Update(gomock.Any()).Return(nil)

		resp, err := svc.MoveFile(context.Background(), 1, 1, "/lib/src")
		assert.NoError(t, err)
		assert.Equal(t, "/lib/src", resp.Path)
	})
}

func TestFileService_DownloadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("文件不存在应报错", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(999)).Return(nil, errors.New("not found"))

		_, _, err := svc.DownloadFile(context.Background(), 999)
		assert.Error(t, err)
		assert.Equal(t, "文件不存在", err.Error())
	})

	t.Run("不能下载目录", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "src", IsDir: true,
		}, nil)

		_, _, err := svc.DownloadFile(context.Background(), 1)
		assert.Error(t, err)
		assert.Equal(t, "不能下载目录", err.Error())
	})

	t.Run("ObjectKey为空应报错", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, RepoID: 1, FileName: "empty.txt", IsDir: false, ObjectKey: "",
		}, nil)

		_, _, err := svc.DownloadFile(context.Background(), 1)
		assert.Error(t, err)
		assert.Equal(t, "文件存储键为空", err.Error())
	})
}

func TestFileService_UploadInit(t *testing.T) {
	t.Run("仓库不存在应报错", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockFileRepo := db.NewMockFileRepository(ctrl)
		mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
		mockRepoRepo := db.NewMockRepoRepository(ctrl)
		svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

		req := dto.UploadInitRequest{RepoID: 999, FileName: "big.zip", FilePath: "/big.zip", FileSize: 100, ChunkSize: 10, TotalChunks: 10}
		mockRepoRepo.EXPECT().GetByID(uint64(999)).Return(nil, errors.New("not found"))

		_, err := svc.UploadInit(context.Background(), 1, req)
		assert.Error(t, err)
		assert.Equal(t, "仓库不存在", err.Error())
	})

	t.Run("新任务创建成功", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockFileRepo := db.NewMockFileRepository(ctrl)
		mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
		mockRepoRepo := db.NewMockRepoRepository(ctrl)
		svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

		req := dto.UploadInitRequest{RepoID: 1, FileName: "big.zip", FilePath: "/big.zip", FileSize: 100, ChunkSize: 10, TotalChunks: 10}
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1}, nil)
		mockUploadRepo.EXPECT().GetByUploadID(gomock.Any()).Return(nil, errors.New("not found"))
		mockUploadRepo.EXPECT().Create(gomock.Any()).Return(nil)

		resp, err := svc.UploadInit(context.Background(), 1, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.UploadID)
		assert.Equal(t, 0, resp.UploadedChunks)
	})

	t.Run("断点续传返回已上传分块数", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockFileRepo := db.NewMockFileRepository(ctrl)
		mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
		mockRepoRepo := db.NewMockRepoRepository(ctrl)
		svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

		req := dto.UploadInitRequest{RepoID: 1, FileName: "big.zip", FilePath: "/big.zip", FileSize: 100, ChunkSize: 10, TotalChunks: 10}
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1}, nil)
		mockUploadRepo.EXPECT().GetByUploadID(gomock.Any()).Return(&entity.UploadTask{
			ID: 1, UploadID: "resume-id", Status: "uploading", UploadedChunks: 5,
		}, nil)

		resp, err := svc.UploadInit(context.Background(), 1, req)
		assert.NoError(t, err)
		assert.Equal(t, 5, resp.UploadedChunks)
	})
}

func TestFileService_UploadChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("上传任务不存在应报错", func(t *testing.T) {
		req := dto.UploadChunkRequest{UploadID: "not-exist", ChunkIndex: 0}
		mockUploadRepo.EXPECT().GetByUploadID("not-exist").Return(nil, errors.New("not found"))

		_, err := svc.UploadChunk(context.Background(), req, []byte("data"))
		assert.Error(t, err)
		assert.Equal(t, "上传任务不存在", err.Error())
	})

	t.Run("任务已结束应报错", func(t *testing.T) {
		req := dto.UploadChunkRequest{UploadID: "done-id", ChunkIndex: 0}
		mockUploadRepo.EXPECT().GetByUploadID("done-id").Return(&entity.UploadTask{
			UploadID: "done-id", Status: "completed", TotalChunks: 10,
		}, nil)

		_, err := svc.UploadChunk(context.Background(), req, []byte("data"))
		assert.Error(t, err)
		assert.Equal(t, "上传任务已结束", err.Error())
	})

	t.Run("分块索引超出范围应报错", func(t *testing.T) {
		req := dto.UploadChunkRequest{UploadID: "active-id", ChunkIndex: 20}
		mockUploadRepo.EXPECT().GetByUploadID("active-id").Return(&entity.UploadTask{
			UploadID: "active-id", Status: "uploading", TotalChunks: 10,
		}, nil)

		_, err := svc.UploadChunk(context.Background(), req, []byte("data"))
		assert.Error(t, err)
		assert.Equal(t, "分块索引超出范围", err.Error())
	})

	t.Run("重复上传同一分块直接返回", func(t *testing.T) {
		req := dto.UploadChunkRequest{UploadID: "active-id", ChunkIndex: 2}
		mockUploadRepo.EXPECT().GetByUploadID("active-id").Return(&entity.UploadTask{
			UploadID: "active-id", Status: "uploading", TotalChunks: 10,
			UploadedChunkIndices: "0,1,2,3",
		}, nil)

		resp, err := svc.UploadChunk(context.Background(), req, []byte("data"))
		assert.NoError(t, err)
		assert.Equal(t, 2, resp.ChunkIndex)
		assert.True(t, resp.UploadedChunks >= 4)
	})
}

func TestFileService_UploadComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockUploadRepo := db.NewMockUploadTaskRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewFileService(mockFileRepo, mockUploadRepo, mockRepoRepo)

	t.Run("上传任务不存在应报错", func(t *testing.T) {
		req := dto.UploadCompleteRequest{UploadID: "not-exist"}
		mockUploadRepo.EXPECT().GetByUploadID("not-exist").Return(nil, errors.New("not found"))

		_, err := svc.UploadComplete(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "上传任务不存在", err.Error())
	})

	t.Run("分块未全部上传应报错", func(t *testing.T) {
		req := dto.UploadCompleteRequest{UploadID: "partial-id"}
		mockUploadRepo.EXPECT().GetByUploadID("partial-id").Return(&entity.UploadTask{
			UploadID: "partial-id", Status: "uploading", TotalChunks: 10, UploadedChunks: 8,
		}, nil)

		_, err := svc.UploadComplete(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "分块未全部上传")
	})

	t.Run("任务状态异常应报错", func(t *testing.T) {
		req := dto.UploadCompleteRequest{UploadID: "bad-id"}
		mockUploadRepo.EXPECT().GetByUploadID("bad-id").Return(&entity.UploadTask{
			UploadID: "bad-id", Status: "completed", TotalChunks: 10, UploadedChunks: 10,
		}, nil)

		_, err := svc.UploadComplete(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "上传任务状态异常")
	})
}
