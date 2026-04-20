package db

import (
	"code-repo/internal/model/entity"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// 使用更标准的 SQLite 内存连接字符串
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法连接到内存数据库: %v", err)
	}

	// 执行自动迁移
	err = db.AutoMigrate(&entity.Repository{}, &entity.User{})
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

func TestRepoRepository_CreateAndGet(t *testing.T) {
	// 传入 t 以便在 setup 失败时直接终止
	dbConn := setupTestDB(t)
	repo := NewRepoRepository(dbConn)

	t.Run("保存并按名称查询", func(t *testing.T) {
		testRepo := &entity.Repository{
			Name:    "test-sql-repo",
			OwnerID: 1,
			Path:    "/tmp/test",
		}

		err := repo.Create(testRepo)
		assert.NoError(t, err)
		assert.NotZero(t, testRepo.ID)

		found, err := repo.GetByNameAndOwner("test-sql-repo", 1)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, testRepo.Path, found.Path)
	})

	t.Run("测试分页查询", func(t *testing.T) {
		// 清理之前的数据（对于内存数据库，每个测试运行最好是独立的）
		dbConn.Exec("DELETE FROM repositories")

		// 插入 5 条测试数据
		for i := 1; i <= 5; i++ {
			repoName := fmt.Sprintf("list-test-%d", i) // 生成 list-test-1, list-test-2...
			err := repo.Create(&entity.Repository{
				Name:    repoName,
				OwnerID: 2,
				Path:    "/tmp/" + repoName,
			})
			assert.NoError(t, err) // 顺便检查每次插入是否成功
		}

		repos, total, err := repo.ListByOwner(2, 1, 3, "") // 第1页，每页3条
		assert.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, repos, 3)
	})
}
