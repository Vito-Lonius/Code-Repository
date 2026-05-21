package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code-repo/internal/model/dto"
	"code-repo/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFileHandler_GetFileDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockFileService(ctrl)
	handler := NewFileHandler(mockSvc)

	t.Run("无效的文件ID应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/files/abc", nil)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		handler.GetFileDetail(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("文件不存在应返回404", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/files/999", nil)
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		mockSvc.EXPECT().GetFileDetail(gomock.Any(), uint64(999)).Return(nil, assert.AnError)

		handler.GetFileDetail(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("正常获取文件详情", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/files/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		mockSvc.EXPECT().GetFileDetail(gomock.Any(), uint64(1)).Return(&dto.FileResponse{
			ID: 1, FileName: "main.go", MimeType: "text/x-go",
		}, nil)

		handler.GetFileDetail(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "main.go")
	})
}

func TestFileHandler_DeleteFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockFileService(ctrl)
	handler := NewFileHandler(mockSvc)

	t.Run("无效的文件ID应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))
		c.Request, _ = http.NewRequest("DELETE", "/api/v1/files/abc", nil)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		handler.DeleteFile(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("正常删除文件", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))
		c.Request, _ = http.NewRequest("DELETE", "/api/v1/files/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		mockSvc.EXPECT().DeleteFile(gomock.Any(), uint64(1), uint64(1)).Return(nil)

		handler.DeleteFile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "文件已成功删除")
	})
}

func TestFileHandler_CreateDir(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockFileService(ctrl)
	handler := NewFileHandler(mockSvc)

	t.Run("无效的JSON输入应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))
		c.Request, _ = http.NewRequest("POST", "/api/v1/files/dir", bytes.NewBufferString("{invalid}"))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateDir(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("正常创建目录", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))

		reqBody := dto.CreateDirRequest{RepoID: 1, Path: "/", DirName: "src"}
		jsonStr, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("POST", "/api/v1/files/dir", bytes.NewBuffer(jsonStr))
		c.Request.Header.Set("Content-Type", "application/json")

		mockSvc.EXPECT().CreateDir(gomock.Any(), uint64(1), reqBody).Return(&dto.FileResponse{
			ID: 1, FileName: "src", IsDir: true,
		}, nil)

		handler.CreateDir(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "src")
	})
}

func TestFileHandler_RenameFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockFileService(ctrl)
	handler := NewFileHandler(mockSvc)

	t.Run("无效的文件ID应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))
		c.Request, _ = http.NewRequest("PUT", "/api/v1/files/abc/rename", nil)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		handler.RenameFile(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("正常重命名文件", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		reqBody := dto.RenameFileRequest{NewName: "renamed.go"}
		jsonStr, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/files/1/rename", bytes.NewBuffer(jsonStr))
		c.Request.Header.Set("Content-Type", "application/json")

		mockSvc.EXPECT().RenameFile(gomock.Any(), uint64(1), uint64(1), "renamed.go").Return(&dto.FileResponse{
			ID: 1, FileName: "renamed.go",
		}, nil)

		handler.RenameFile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "renamed.go")
	})
}

func TestFileHandler_MoveFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockFileService(ctrl)
	handler := NewFileHandler(mockSvc)

	t.Run("正常移动文件", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		reqBody := dto.MoveFileRequest{NewPath: "/lib/main.go"}
		jsonStr, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/files/1/move", bytes.NewBuffer(jsonStr))
		c.Request.Header.Set("Content-Type", "application/json")

		mockSvc.EXPECT().MoveFile(gomock.Any(), uint64(1), uint64(1), "/lib/main.go").Return(&dto.FileResponse{
			ID: 1, Path: "/lib/main.go",
		}, nil)

		handler.MoveFile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "/lib/main.go")
	})
}

func TestFileHandler_ListFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockFileService(ctrl)
	handler := NewFileHandler(mockSvc)

	t.Run("无效的仓库ID应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/files?repo_id=abc", nil)

		handler.ListFiles(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("正常列出文件", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/files?repo_id=1", nil)

		mockSvc.EXPECT().ListFiles(gomock.Any(), uint64(1), "").Return(&dto.FileListResponse{
			Total: 1,
			Items: []dto.FileResponse{{ID: 1, FileName: "main.go"}},
		}, nil)

		handler.ListFiles(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "main.go")
	})
}

func TestFileHandler_UploadInit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockFileService(ctrl)
	handler := NewFileHandler(mockSvc)

	t.Run("无效的JSON输入应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))
		c.Request, _ = http.NewRequest("POST", "/api/v1/files/upload/init", bytes.NewBufferString("{invalid}"))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UploadInit(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("正常初始化分块上传", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint64(1))

		reqBody := dto.UploadInitRequest{RepoID: 1, FileName: "big.zip", FilePath: "/big.zip", FileSize: 100, ChunkSize: 10, TotalChunks: 10}
		jsonStr, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("POST", "/api/v1/files/upload/init", bytes.NewBuffer(jsonStr))
		c.Request.Header.Set("Content-Type", "application/json")

		mockSvc.EXPECT().UploadInit(gomock.Any(), uint64(1), reqBody).Return(&dto.UploadInitResponse{
			UploadID: "test-upload-id", UploadedChunks: 0,
		}, nil)

		handler.UploadInit(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "test-upload-id")
	})
}
