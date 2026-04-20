package v1

import (
	"net/http"
	"strconv"

	"code-repo/internal/model/dto"
	"code-repo/internal/service"

	"github.com/gin-gonic/gin"
)

type RepoHandler struct {
	repoService service.RepoService
}

func NewRepoHandler(repoService service.RepoService) *RepoHandler {
	return &RepoHandler{repoService: repoService}
}

// Create 创建仓库
// POST /api/v1/repos
func (h *RepoHandler) Create(c *gin.Context) {
	var req dto.CreateRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数验证失败: " + err.Error()})
		return
	}

	// 模拟获取用户 ID (后续应从 JWT 中间件 c.Get("userID") 获取)
	// 为了让你先跑通，这里假设当前登录用户 ID 为 1
	userID := uint64(1)

	resp, err := h.repoService.CreateRepo(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetDetail 获取仓库详情
// GET /api/v1/repos/:id
func (h *RepoHandler) GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的仓库ID"})
		return
	}

	resp, err := h.repoService.GetRepoDetail(uint64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete 删除仓库
// DELETE /api/v1/repos/:id
func (h *RepoHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint64) // 从 JWT 中间件获取

	if err := h.repoService.DeleteRepo(userID, uint64(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "仓库已成功删除"})
}
