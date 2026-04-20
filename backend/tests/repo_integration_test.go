package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"

	"github.com/stretchr/testify/assert"
)

func TestRepo_Integration(t *testing.T) {
	repoName := "integration-test-repo"

	// 清理之前失败的残留数据（注意 Unscoped 以防软删除干扰）
	testDB.Unscoped().Where("name = ?", repoName).Delete(&entity.Repository{})

	t.Run("成功创建仓库并验证物理路径", func(t *testing.T) {
		req := dto.CreateRepoRequest{
			Name:        repoName,
			Description: "集成测试仓库描述",
			IsPublic:    true,
		}
		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		request, _ := http.NewRequest("POST", "/api/v1/repos", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, request)

		// 验证响应
		assert.Equal(t, http.StatusCreated, w.Code)

		// 验证数据库
		var repo entity.Repository
		err := testDB.Where("name = ?", repoName).First(&repo).Error
		assert.NoError(t, err)

		// 验证物理路径是否存在
		if _, err := os.Stat(repo.Path); os.IsNotExist(err) {
			t.Errorf("Git 仓库物理目录未创建: %s", repo.Path)
		} else {
			// 测试成功后清理物理目录
			os.RemoveAll(repo.Path)
		}
	})
}
