package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code-repo/internal/model/dto"
	"code-repo/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPreviewHandler_PreviewFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockPreviewService(ctrl)
	handler := NewPreviewHandler(mockSvc)

	t.Run("无效的文件ID应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/abc", nil)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		handler.PreviewFile(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("文件不存在应返回404", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/999", nil)
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		mockSvc.EXPECT().PreviewFile(gomock.Any(), uint64(999)).Return(nil, assert.AnError)

		handler.PreviewFile(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("正常预览代码文件", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		mockSvc.EXPECT().PreviewFile(gomock.Any(), uint64(1)).Return(&dto.PreviewResponse{
			FileID: 1, FileName: "main.go", FileType: "code", Language: "go",
		}, nil)

		handler.PreviewFile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "main.go")
		assert.Contains(t, w.Body.String(), "code")
	})
}

func TestPreviewHandler_GetDirectoryTree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockPreviewService(ctrl)
	handler := NewPreviewHandler(mockSvc)

	t.Run("无效的仓库ID应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/repos/abc/tree", nil)
		c.Params = gin.Params{{Key: "repo_id", Value: "abc"}}

		handler.GetDirectoryTree(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("正常获取目录树", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/repos/1/tree", nil)
		c.Params = gin.Params{{Key: "repo_id", Value: "1"}}

		mockSvc.EXPECT().GetDirectoryTree(gomock.Any(), uint64(1)).Return(&dto.TreeResponse{
			RepoID: 1,
			Tree: []*dto.TreeNode{
				{ID: 1, Name: "src", Path: "/src", IsDir: true},
			},
		}, nil)

		handler.GetDirectoryTree(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "src")
	})
}

func TestPreviewHandler_GetRawFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockPreviewService(ctrl)
	handler := NewPreviewHandler(mockSvc)

	t.Run("无效的文件ID应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/abc/raw", nil)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		handler.GetRawFile(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("文件不存在应返回404", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/999/raw", nil)
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		mockSvc.EXPECT().GetRawFile(gomock.Any(), uint64(999)).Return(nil, assert.AnError)

		handler.GetRawFile(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestPreviewHandler_PreviewImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockPreviewService(ctrl)
	handler := NewPreviewHandler(mockSvc)

	t.Run("非图片文件应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/1/image", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		mockSvc.EXPECT().PreviewFile(gomock.Any(), uint64(1)).Return(&dto.PreviewResponse{
			FileID: 1, FileType: "code",
		}, nil)

		handler.PreviewImage(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "不是图片类型")
	})
}

func TestPreviewHandler_PreviewMedia(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockPreviewService(ctrl)
	handler := NewPreviewHandler(mockSvc)

	t.Run("非音视频文件应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/1/media", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		mockSvc.EXPECT().PreviewFile(gomock.Any(), uint64(1)).Return(&dto.PreviewResponse{
			FileID: 1, FileType: "image",
		}, nil)

		handler.PreviewMedia(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "不是音视频类型")
	})
}

func TestPreviewHandler_PreviewPDF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockPreviewService(ctrl)
	handler := NewPreviewHandler(mockSvc)

	t.Run("非PDF文件应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/1/pdf", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		mockSvc.EXPECT().PreviewFile(gomock.Any(), uint64(1)).Return(&dto.PreviewResponse{
			FileID: 1, FileType: "image",
		}, nil)

		handler.PreviewPDF(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "不是PDF类型")
	})
}

func TestPreviewHandler_GetPreviewInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockPreviewService(ctrl)
	handler := NewPreviewHandler(mockSvc)

	t.Run("正常获取预览信息", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/preview/1/info", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		mockSvc.EXPECT().PreviewFile(gomock.Any(), uint64(1)).Return(&dto.PreviewResponse{
			FileID: 1, FileName: "photo.png", FileType: "image", FileSize: 102400,
		}, nil)

		handler.GetPreviewInfo(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "photo.png")
	})
}

func TestPreviewHandler_ListDir(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockPreviewService(ctrl)
	handler := NewPreviewHandler(mockSvc)

	t.Run("无效的仓库ID应返回400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/repos/abc/dir", nil)
		c.Params = gin.Params{{Key: "repo_id", Value: "abc"}}

		handler.ListDir(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("正常列出目录内容", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/repos/1/dir", nil)
		c.Params = gin.Params{{Key: "repo_id", Value: "1"}}

		mockSvc.EXPECT().GetDirectoryTree(gomock.Any(), uint64(1)).Return(&dto.TreeResponse{
			RepoID: 1,
			Tree: []*dto.TreeNode{
				{ID: 1, Name: "src", Path: "/src", IsDir: true},
				{ID: 2, Name: "main.go", Path: "/main.go", IsDir: false},
			},
		}, nil)

		handler.ListDir(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
