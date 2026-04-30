package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"code-repo/internal/model/dto"
	"code-repo/internal/service"

	"github.com/gin-gonic/gin"
)

type PreviewHandler struct {
	previewSvc service.PreviewService
}

func NewPreviewHandler(previewSvc service.PreviewService) *PreviewHandler {
	return &PreviewHandler{previewSvc: previewSvc}
}

func (h *PreviewHandler) PreviewFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	resp, err := h.previewSvc.PreviewFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PreviewHandler) GetDirectoryTree(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的仓库ID"})
		return
	}

	resp, err := h.previewSvc.GetDirectoryTree(c.Request.Context(), repoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PreviewHandler) GetRawFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	resp, err := h.previewSvc.GetRawFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", resp.FileName))
	c.Header("Content-Type", resp.ContentType)
	c.Data(http.StatusOK, resp.ContentType, resp.Content)
}

func (h *PreviewHandler) PreviewImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	preview, err := h.previewSvc.PreviewFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if preview.FileType != "image" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该文件不是图片类型"})
		return
	}

	rawResp, err := h.previewSvc.GetRawFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", rawResp.ContentType)
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, rawResp.ContentType, rawResp.Content)
}

func (h *PreviewHandler) PreviewMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	preview, err := h.previewSvc.PreviewFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if preview.FileType != "video" && preview.FileType != "audio" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该文件不是音视频类型"})
		return
	}

	rawResp, err := h.previewSvc.GetRawFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", rawResp.ContentType)
	c.Header("Accept-Ranges", "bytes")
	c.Data(http.StatusOK, rawResp.ContentType, rawResp.Content)
}

func (h *PreviewHandler) PreviewPDF(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	preview, err := h.previewSvc.PreviewFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if preview.FileType != "pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该文件不是PDF类型"})
		return
	}

	rawResp, err := h.previewSvc.GetRawFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", rawResp.FileName))
	c.Data(http.StatusOK, "application/pdf", rawResp.Content)
}

func (h *PreviewHandler) GetPreviewInfo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	resp, err := h.previewSvc.PreviewFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PreviewHandler) ListDir(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的仓库ID"})
		return
	}

	resp, err := h.previewSvc.GetDirectoryTree(c.Request.Context(), repoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []*dto.TreeNode
	path := c.Query("path")
	if path == "" || path == "/" {
		result = resp.Tree
	} else {
		result = findNodesByPath(resp.Tree, path)
	}

	c.JSON(http.StatusOK, gin.H{
		"repo_id": repoID,
		"path":    path,
		"items":   result,
	})
}

func findNodesByPath(nodes []*dto.TreeNode, targetPath string) []*dto.TreeNode {
	for _, node := range nodes {
		if node.Path == targetPath && node.IsDir {
			return node.Children
		}
		if node.IsDir && node.Children != nil {
			if result := findNodesByPath(node.Children, targetPath); result != nil {
				return result
			}
		}
	}
	return nil
}
