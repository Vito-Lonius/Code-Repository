package v1

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"code-repo/internal/model/dto"
	"code-repo/internal/service"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileSvc service.FileService
}

func NewFileHandler(fileSvc service.FileService) *FileHandler {
	return &FileHandler{fileSvc: fileSvc}
}

func (h *FileHandler) UploadSimple(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)

	repoIDStr := c.PostForm("repo_id")
	repoID, err := strconv.ParseUint(repoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的仓库ID"})
		return
	}

	filePath := c.PostForm("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件路径不能为空"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}
	defer file.Close()

	resp, err := h.fileSvc.UploadSimple(c.Request.Context(), repoID, userID, filePath, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *FileHandler) UploadInit(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)

	var req dto.UploadInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数验证失败: " + err.Error()})
		return
	}

	resp, err := h.fileSvc.UploadInit(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) UploadChunk(c *gin.Context) {
	var req dto.UploadChunkRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数验证失败: " + err.Error()})
		return
	}

	file, _, err := c.Request.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传分块数据"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取分块数据失败"})
		return
	}

	resp, err := h.fileSvc.UploadChunk(c.Request.Context(), req, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) UploadComplete(c *gin.Context) {
	var req dto.UploadCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数验证失败: " + err.Error()})
		return
	}

	resp, err := h.fileSvc.UploadComplete(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) GetFileDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	resp, err := h.fileSvc.GetFileDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Query("repo_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的仓库ID"})
		return
	}

	parentPath := c.Query("path")

	resp, err := h.fileSvc.ListFiles(c.Request.Context(), repoID, parentPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	reader, file, err := h.fileSvc.DownloadFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(file.FileName)))
	c.Header("Content-Type", file.MimeType)
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, reader)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	if err := h.fileSvc.DeleteFile(c.Request.Context(), userID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件已成功删除"})
}

func (h *FileHandler) CreateDir(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)

	var req dto.CreateDirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数验证失败: " + err.Error()})
		return
	}

	resp, err := h.fileSvc.CreateDir(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *FileHandler) RenameFile(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var req dto.RenameFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数验证失败: " + err.Error()})
		return
	}

	resp, err := h.fileSvc.RenameFile(c.Request.Context(), userID, id, req.NewName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) MoveFile(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var req dto.MoveFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数验证失败: " + err.Error()})
		return
	}

	resp, err := h.fileSvc.MoveFile(c.Request.Context(), userID, id, req.NewPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
