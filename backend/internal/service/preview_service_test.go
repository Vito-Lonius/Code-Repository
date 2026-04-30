package service

import (
	"context"
	"testing"

	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPreviewService_PreviewFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewPreviewService(mockFileRepo, mockRepoRepo)

	t.Run("代码文件预览", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(1)).Return(&entity.File{
			ID: 1, FileName: "main.go", Path: "/main.go",
			IsDir: false, MimeType: "text/x-go", FileSize: 512,
			ObjectKey: "repos/1/main.go", Status: "completed",
		}, nil)

		resp, err := svc.PreviewFile(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "code", resp.FileType)
		assert.Equal(t, "go", resp.Language)
		assert.Equal(t, "main.go", resp.FileName)
	})

	t.Run("图片文件预览", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(2)).Return(&entity.File{
			ID: 2, FileName: "photo.png", Path: "/photo.png",
			IsDir: false, MimeType: "image/png", FileSize: 102400,
			ObjectKey: "repos/1/photo.png", Status: "completed",
		}, nil)

		resp, err := svc.PreviewFile(context.Background(), 2)
		assert.NoError(t, err)
		assert.Equal(t, "image", resp.FileType)
	})

	t.Run("PDF文件预览", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(3)).Return(&entity.File{
			ID: 3, FileName: "doc.pdf", Path: "/doc.pdf",
			IsDir: false, MimeType: "application/pdf", FileSize: 204800,
			ObjectKey: "repos/1/doc.pdf", Status: "completed",
		}, nil)

		resp, err := svc.PreviewFile(context.Background(), 3)
		assert.NoError(t, err)
		assert.Equal(t, "pdf", resp.FileType)
	})

	t.Run("视频文件预览", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(4)).Return(&entity.File{
			ID: 4, FileName: "intro.mp4", Path: "/intro.mp4",
			IsDir: false, MimeType: "video/mp4", FileSize: 10485760,
			ObjectKey: "repos/1/intro.mp4", Status: "completed",
		}, nil)

		resp, err := svc.PreviewFile(context.Background(), 4)
		assert.NoError(t, err)
		assert.Equal(t, "video", resp.FileType)
	})

	t.Run("音频文件预览", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(5)).Return(&entity.File{
			ID: 5, FileName: "song.mp3", Path: "/song.mp3",
			IsDir: false, MimeType: "audio/mpeg", FileSize: 5242880,
			ObjectKey: "repos/1/song.mp3", Status: "completed",
		}, nil)

		resp, err := svc.PreviewFile(context.Background(), 5)
		assert.NoError(t, err)
		assert.Equal(t, "audio", resp.FileType)
	})

	t.Run("Office文件预览", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(6)).Return(&entity.File{
			ID: 6, FileName: "report.docx", Path: "/report.docx",
			IsDir: false, MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			FileSize: 30720, ObjectKey: "repos/1/report.docx", Status: "completed",
		}, nil)

		resp, err := svc.PreviewFile(context.Background(), 6)
		assert.NoError(t, err)
		assert.Equal(t, "office", resp.FileType)
	})

	t.Run("Markdown文件预览", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(7)).Return(&entity.File{
			ID: 7, FileName: "README.md", Path: "/README.md",
			IsDir: false, MimeType: "text/markdown", FileSize: 2048,
			ObjectKey: "repos/1/README.md", Status: "completed",
		}, nil)

		resp, err := svc.PreviewFile(context.Background(), 7)
		assert.NoError(t, err)
		assert.Equal(t, "code", resp.FileType)
		assert.Equal(t, "markdown", resp.Language)
	})

	t.Run("目录不能预览", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(8)).Return(&entity.File{
			ID: 8, FileName: "src", Path: "/src", IsDir: true,
		}, nil)

		_, err := svc.PreviewFile(context.Background(), 8)
		assert.Error(t, err)
		assert.Equal(t, "不能预览目录", err.Error())
	})

	t.Run("文件不存在", func(t *testing.T) {
		mockFileRepo.EXPECT().GetByID(uint64(999)).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.PreviewFile(context.Background(), 999)
		assert.Error(t, err)
		assert.Equal(t, "文件不存在", err.Error())
	})
}

func TestPreviewService_GetDirectoryTree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFileRepo := db.NewMockFileRepository(ctrl)
	mockRepoRepo := db.NewMockRepoRepository(ctrl)

	svc := NewPreviewService(mockFileRepo, mockRepoRepo)

	t.Run("构建目录树", func(t *testing.T) {
		mockRepoRepo.EXPECT().GetByID(uint64(1)).Return(&entity.Repository{ID: 1}, nil)
		mockFileRepo.EXPECT().ListByRepo(uint64(1), "").Return([]entity.File{
			{ID: 1, FileName: "README.md", Path: "/README.md", IsDir: false, MimeType: "text/markdown", FileSize: 1024},
			{ID: 2, FileName: "src", Path: "/src", IsDir: true, MimeType: "directory"},
			{ID: 3, FileName: "main.go", Path: "/src/main.go", IsDir: false, MimeType: "text/x-go", FileSize: 2048},
			{ID: 4, FileName: "utils", Path: "/src/utils", IsDir: true, MimeType: "directory"},
			{ID: 5, FileName: "helper.go", Path: "/src/utils/helper.go", IsDir: false, MimeType: "text/x-go", FileSize: 512},
		}, nil)

		resp, err := svc.GetDirectoryTree(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint64(1), resp.RepoID)
		assert.NotEmpty(t, resp.Tree)
	})

	t.Run("仓库不存在", func(t *testing.T) {
		mockRepoRepo.EXPECT().GetByID(uint64(999)).Return(nil, gorm.ErrRecordNotFound)

		_, err := svc.GetDirectoryTree(context.Background(), 999)
		assert.Error(t, err)
		assert.Equal(t, "仓库不存在", err.Error())
	})
}

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		mimeType  string
		wantType  string
		wantLang  string
	}{
		{"Go代码", "main.go", "text/x-go", "code", "go"},
		{"Python代码", "app.py", "text/x-python", "code", "python"},
		{"JavaScript", "index.js", "text/javascript", "code", "javascript"},
		{"TypeScript", "app.ts", "text/typescript", "code", "typescript"},
		{"Java代码", "Main.java", "text/x-java", "code", "java"},
		{"Rust代码", "lib.rs", "text/x-rust", "code", "rust"},
		{"Markdown", "README.md", "text/markdown", "code", "markdown"},
		{"HTML", "index.html", "text/html", "code", "html"},
		{"CSS", "style.css", "text/css", "code", "css"},
		{"JSON数据", "config.json", "application/json", "code", "json"},
		{"YAML配置", "docker-compose.yml", "application/x-yaml", "code", "yaml"},
		{"Shell脚本", "deploy.sh", "text/x-sh", "code", "shell"},
		{"SQL脚本", "schema.sql", "text/x-sql", "code", "sql"},
		{"PNG图片", "logo.png", "image/png", "image", ""},
		{"JPG图片", "photo.jpg", "image/jpeg", "image", ""},
		{"GIF图片", "anim.gif", "image/gif", "image", ""},
		{"WebP图片", "pic.webp", "image/webp", "image", ""},
		{"SVG图片", "icon.svg", "image/svg+xml", "code", "svg"},
		{"PDF文档", "doc.pdf", "application/pdf", "pdf", ""},
		{"Word文档", "report.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "office", ""},
		{"Excel文档", "data.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "office", ""},
		{"PPT文档", "slides.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation", "office", ""},
		{"MP4视频", "intro.mp4", "video/mp4", "video", ""},
		{"WebM视频", "clip.webm", "video/webm", "video", ""},
		{"MP3音频", "song.mp3", "audio/mpeg", "audio", ""},
		{"WAV音频", "voice.wav", "audio/wav", "audio", ""},
		{"纯文本", "notes.txt", "text/plain", "text", ""},
		{"CSV数据", "data.csv", "text/csv", "text", ""},
		{"未知二进制", "data.bin", "application/octet-stream", "binary", ""},
		{"Vue组件", "App.vue", "text/x-vue", "code", "vue"},
		{"Dockerfile", "Dockerfile", "text/x-dockerfile", "text", ""},
		{"环境配置", ".env", "text/plain", "text", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotLang := classifyFile(tt.fileName, tt.mimeType)
			assert.Equal(t, tt.wantType, gotType, "fileType mismatch for %s", tt.fileName)
			assert.Equal(t, tt.wantLang, gotLang, "language mismatch for %s", tt.fileName)
		})
	}
}
