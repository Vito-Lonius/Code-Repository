package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"
	"code-repo/internal/repository/storage"
)

type PreviewService interface {
	PreviewFile(ctx context.Context, fileID uint64) (*dto.PreviewResponse, error)
	GetDirectoryTree(ctx context.Context, repoID uint64) (*dto.TreeResponse, error)
	GetRawFile(ctx context.Context, fileID uint64) (*dto.RawFileResponse, error)
}

type previewService struct {
	fileRepo db.FileRepository
	repoRepo db.RepoRepository
}

func NewPreviewService(fileRepo db.FileRepository, repoRepo db.RepoRepository) PreviewService {
	return &previewService{fileRepo: fileRepo, repoRepo: repoRepo}
}

var codeLanguageMap = map[string]string{
	".go": "go", ".py": "python", ".js": "javascript", ".ts": "typescript",
	".jsx": "jsx", ".tsx": "tsx", ".java": "java", ".c": "c", ".cpp": "cpp",
	".h": "c", ".hpp": "cpp", ".cs": "csharp", ".rb": "ruby", ".php": "php",
	".rs": "rust", ".swift": "swift", ".kt": "kotlin", ".scala": "scala",
	".sh": "shell", ".bash": "shell", ".zsh": "shell", ".ps1": "powershell",
	".sql": "sql", ".r": "r", ".m": "matlab", ".lua": "lua",
	".html": "html", ".htm": "html", ".css": "css", ".scss": "scss",
	".less": "less", ".svg": "svg", ".xml": "xml", ".yaml": "yaml",
	".yml": "yaml", ".json": "json", ".toml": "toml", ".ini": "ini",
	".md": "markdown", ".rst": "rst", ".tex": "latex",
	".dart": "dart", ".vue": "vue", ".svelte": "svelte",
	".dockerfile": "dockerfile", ".makefile": "makefile",
	".proto": "protobuf", ".graphql": "graphql",
	".tf": "hcl", ".hcl": "hcl",
}

var textMimePrefixes = []string{
	"text/", "application/json", "application/xml", "application/javascript",
	"application/typescript", "application/x-yaml", "application/x-sh",
	"application/xhtml+xml",
}

var textExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".java": true,
	".c": true, ".cpp": true, ".h": true, ".hpp": true, ".cs": true,
	".rb": true, ".php": true, ".rs": true, ".swift": true, ".kt": true,
	".scala": true, ".sh": true, ".bash": true, ".zsh": true,
	".sql": true, ".r": true, ".lua": true, ".dart": true,
	".html": true, ".htm": true, ".css": true, ".scss": true,
	".less": true, ".xml": true, ".yaml": true, ".yml": true,
	".json": true, ".toml": true, ".ini": true, ".cfg": true,
	".md": true, ".rst": true, ".tex": true, ".txt": true,
	".vue": true, ".svelte": true, ".proto": true, ".graphql": true,
	".tf": true, ".hcl": true, ".env": true, ".gitignore": true,
	".dockerignore": true, ".editorconfig": true, ".log": true,
	".csv": true, ".tsv": true, ".conf": true, ".properties": true,
}

var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".bmp": true, ".webp": true, ".ico": true, ".tiff": true, ".tif": true,
	".avif": true, ".heif": true, ".heic": true,
}

var videoExtensions = map[string]bool{
	".mp4": true, ".webm": true, ".ogg": true, ".avi": true,
	".mov": true, ".mkv": true, ".flv": true, ".wmv": true,
}

var audioExtensions = map[string]bool{
	".mp3": true, ".wav": true, ".flac": true, ".aac": true,
	".ogg": true, ".m4a": true, ".wma": true, ".opus": true,
}

var pdfExtensions = map[string]bool{
	".pdf": true,
}

var officeExtensions = map[string]bool{
	".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".odt": true, ".ods": true, ".odp": true,
}

const maxTextPreviewSize = 2 * 1024 * 1024

func classifyFile(fileName string, mimeType string) (fileType string, language string) {
	ext := strings.ToLower(filepath.Ext(fileName))

	if codeLanguageMap[ext] != "" {
		return "code", codeLanguageMap[ext]
	}

	if imageExtensions[ext] {
		return "image", ""
	}
	if videoExtensions[ext] {
		return "video", ""
	}
	if audioExtensions[ext] {
		return "audio", ""
	}
	if pdfExtensions[ext] {
		return "pdf", ""
	}
	if officeExtensions[ext] {
		return "office", ""
	}

	if textExtensions[ext] {
		return "text", ""
	}

	for _, prefix := range textMimePrefixes {
		if strings.HasPrefix(mimeType, prefix) {
			return "text", ""
		}
	}

	if strings.HasPrefix(mimeType, "image/") {
		return "image", ""
	}
	if strings.HasPrefix(mimeType, "video/") {
		return "video", ""
	}
	if strings.HasPrefix(mimeType, "audio/") {
		return "audio", ""
	}

	return "binary", ""
}

func (s *previewService) PreviewFile(ctx context.Context, fileID uint64) (*dto.PreviewResponse, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, errors.New("文件不存在")
	}
	if file.IsDir {
		return nil, errors.New("不能预览目录")
	}

	fileType, language := classifyFile(file.FileName, file.MimeType)

	resp := &dto.PreviewResponse{
		FileID:   file.ID,
		FileName: file.FileName,
		MimeType: file.MimeType,
		FileSize: file.FileSize,
		FileType: fileType,
		Language: language,
		Encoding: "utf-8",
	}

	if (fileType == "code" || fileType == "text") && file.FileSize <= maxTextPreviewSize {
		content, err := s.readTextContent(ctx, file.ObjectKey, file.FileSize)
		if err == nil {
			resp.Content = content
		}
	}

	return resp, nil
}

func (s *previewService) readTextContent(ctx context.Context, objectKey string, fileSize int64) (string, error) {
	if objectKey == "" {
		return "", errors.New("文件存储键为空")
	}

	obj, err := storage.DownloadFile(ctx, objectKey)
	if err != nil {
		return "", err
	}
	defer obj.Close()

	size := fileSize
	if size <= 0 || size > maxTextPreviewSize {
		size = maxTextPreviewSize
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.CopyN(buf, obj, size); err != nil && err != io.EOF {
		return "", err
	}

	content := buf.String()
	if !isValidUTF8(content) {
		return "", errors.New("文件内容不是有效的UTF-8文本")
	}

	return content, nil
}

func isValidUTF8(s string) bool {
	for _, r := range s {
		if r == '\uFFFD' {
			return false
		}
	}
	return true
}

func (s *previewService) GetDirectoryTree(ctx context.Context, repoID uint64) (*dto.TreeResponse, error) {
	if _, err := s.repoRepo.GetByID(repoID); err != nil {
		return nil, errors.New("仓库不存在")
	}

	files, err := s.fileRepo.ListByRepo(repoID, "")
	if err != nil {
		return nil, err
	}

	tree := buildTree(files)

	return &dto.TreeResponse{
		RepoID: repoID,
		Tree:   tree,
	}, nil
}

func buildTree(files []entity.File) []*dto.TreeNode {
	root := &dto.TreeNode{
		Path:     "/",
		IsDir:    true,
		Children: make([]*dto.TreeNode, 0),
	}

	nodeMap := make(map[string]*dto.TreeNode)
	nodeMap["/"] = root

	for i := range files {
		f := &files[i]
		node := &dto.TreeNode{
			ID:       f.ID,
			Name:     f.FileName,
			Path:     f.Path,
			IsDir:    f.IsDir,
			MimeType: f.MimeType,
			FileSize: f.FileSize,
		}
		if f.IsDir {
			node.Children = make([]*dto.TreeNode, 0)
		}
		nodeMap[f.Path] = node

		parentPath := parentDirPath(f.Path)
		parent, exists := nodeMap[parentPath]
		if !exists {
			parent = &dto.TreeNode{
				Name:     filepath.Base(parentPath),
				Path:     parentPath,
				IsDir:    true,
				Children: make([]*dto.TreeNode, 0),
			}
			nodeMap[parentPath] = parent
		}
		parent.Children = append(parent.Children, node)
	}

	return root.Children
}

func parentDirPath(path string) string {
	cleaned := strings.TrimSuffix(path, "/")
	idx := strings.LastIndex(cleaned, "/")
	if idx <= 0 {
		return "/"
	}
	return cleaned[:idx]
}

func (s *previewService) GetRawFile(ctx context.Context, fileID uint64) (*dto.RawFileResponse, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, errors.New("文件不存在")
	}
	if file.IsDir {
		return nil, errors.New("不能获取目录的原始内容")
	}
	if file.ObjectKey == "" {
		return nil, errors.New("文件存储键为空")
	}

	obj, err := storage.DownloadFile(ctx, file.ObjectKey)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}

	contentType := file.MimeType
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = detectContentType(data, file.FileName)
	}

	return &dto.RawFileResponse{
		Content:     data,
		ContentType: contentType,
		FileName:    file.FileName,
		FileSize:    int64(len(data)),
	}, nil
}

func detectContentType(data []byte, fileName string) string {
	if len(data) > 0 {
		detected := http.DetectContentType(data)
		if detected != "application/octet-stream" {
			return detected
		}
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	mimeMap := map[string]string{
		".go": "text/x-go", ".py": "text/x-python", ".js": "text/javascript",
		".ts": "text/typescript", ".json": "application/json", ".xml": "application/xml",
		".html": "text/html", ".css": "text/css", ".md": "text/markdown",
		".yaml": "application/x-yaml", ".yml": "application/x-yaml",
		".svg": "image/svg+xml", ".pdf": "application/pdf",
	}
	if m, ok := mimeMap[ext]; ok {
		return m
	}
	return "application/octet-stream"
}
