package db

import (
	"code-repo/internal/model/entity"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupFileTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法连接到内存数据库: %v", err)
	}
	err = db.AutoMigrate(&entity.File{}, &entity.UploadTask{}, &entity.Repository{}, &entity.User{})
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

func TestFileRepository_CreateAndGet(t *testing.T) {
	dbConn := setupFileTestDB(t)
	repo := NewFileRepository(dbConn)

	t.Run("创建文件并按ID查询", func(t *testing.T) {
		file := &entity.File{
			RepoID:     1,
			FileName:   "test.txt",
			Path:       "/test.txt",
			IsDir:      false,
			MimeType:   "text/plain",
			FileSize:   100,
			ObjectKey:  "repos/1/test.txt",
			Status:     "completed",
			UploaderID: 1,
		}

		err := repo.Create(file)
		assert.NoError(t, err)
		assert.NotZero(t, file.ID)

		found, err := repo.GetByID(file.ID)
		assert.NoError(t, err)
		assert.Equal(t, "test.txt", found.FileName)
		assert.Equal(t, "/test.txt", found.Path)
	})

	t.Run("按路径查询文件", func(t *testing.T) {
		found, err := repo.GetByPath(1, "/test.txt")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test.txt", found.FileName)
	})

	t.Run("路径不存在时返回错误", func(t *testing.T) {
		_, err := repo.GetByPath(1, "/not-exist.txt")
		assert.Error(t, err)
	})
}

func TestFileRepository_ListByRepo(t *testing.T) {
	dbConn := setupFileTestDB(t)
	repo := NewFileRepository(dbConn)

	files := []entity.File{
		{RepoID: 1, FileName: "a.txt", Path: "/a.txt", IsDir: false, MimeType: "text/plain", Status: "completed", UploaderID: 1},
		{RepoID: 1, FileName: "b.txt", Path: "/b.txt", IsDir: false, MimeType: "text/plain", Status: "completed", UploaderID: 1},
		{RepoID: 1, FileName: "src", Path: "/src", IsDir: true, MimeType: "directory", Status: "completed", UploaderID: 1},
		{RepoID: 2, FileName: "other.txt", Path: "/other.txt", IsDir: false, Status: "completed", UploaderID: 1},
	}
	for _, f := range files {
		_ = repo.Create(&f)
	}

	t.Run("查询仓库1的文件列表", func(t *testing.T) {
		result, err := repo.ListByRepo(1, "")
		assert.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("查询仓库2的文件列表", func(t *testing.T) {
		result, err := repo.ListByRepo(2, "")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestFileRepository_Delete(t *testing.T) {
	dbConn := setupFileTestDB(t)
	repo := NewFileRepository(dbConn)

	file := &entity.File{
		RepoID: 1, FileName: "del.txt", Path: "/del.txt",
		IsDir: false, Status: "completed", UploaderID: 1,
	}
	_ = repo.Create(file)

	err := repo.Delete(file.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(file.ID)
	assert.Error(t, err)
}

func TestUploadTaskRepository_CreateAndGet(t *testing.T) {
	dbConn := setupFileTestDB(t)
	repo := NewUploadTaskRepository(dbConn)

	t.Run("创建上传任务并查询", func(t *testing.T) {
		task := &entity.UploadTask{
			UploadID:    "test-upload-123",
			RepoID:      1,
			FileName:    "bigfile.zip",
			FilePath:    "/bigfile.zip",
			FileSize:    100 * 1024 * 1024,
			ChunkSize:   5 * 1024 * 1024,
			TotalChunks: 20,
			Status:      "uploading",
			UploaderID:  1,
		}

		err := repo.Create(task)
		assert.NoError(t, err)
		assert.NotZero(t, task.ID)

		found, err := repo.GetByUploadID("test-upload-123")
		assert.NoError(t, err)
		assert.Equal(t, 20, found.TotalChunks)
		assert.Equal(t, "uploading", found.Status)
	})

	t.Run("更新已上传分块数", func(t *testing.T) {
		err := repo.UpdateUploadedChunks("test-upload-123", 10, "0,1,2,3,4,5,6,7,8,9")
		assert.NoError(t, err)

		found, _ := repo.GetByUploadID("test-upload-123")
		assert.Equal(t, 10, found.UploadedChunks)
	})

	t.Run("更新任务状态", func(t *testing.T) {
		err := repo.UpdateStatus("test-upload-123", "completed")
		assert.NoError(t, err)

		found, _ := repo.GetByUploadID("test-upload-123")
		assert.Equal(t, "completed", found.Status)
	})
}
