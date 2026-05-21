package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"
	"code-repo/internal/repository/storage"
)

type FileService interface {
	UploadSimple(ctx context.Context, repoID uint64, uploaderID uint64, filePath string, file multipart.File, header *multipart.FileHeader) (*dto.FileResponse, error)
	UploadInit(ctx context.Context, uploaderID uint64, req dto.UploadInitRequest) (*dto.UploadInitResponse, error)
	UploadChunk(ctx context.Context, req dto.UploadChunkRequest, data []byte) (*dto.UploadChunkResponse, error)
	UploadComplete(ctx context.Context, req dto.UploadCompleteRequest) (*dto.FileResponse, error)
	GetFileDetail(ctx context.Context, fileID uint64) (*dto.FileResponse, error)
	ListFiles(ctx context.Context, repoID uint64, parentPath string) (*dto.FileListResponse, error)
	DownloadFile(ctx context.Context, fileID uint64) (io.Reader, *entity.File, error)
	DeleteFile(ctx context.Context, userID uint64, fileID uint64) error
	CreateDir(ctx context.Context, userID uint64, req dto.CreateDirRequest) (*dto.FileResponse, error)
	RenameFile(ctx context.Context, userID uint64, fileID uint64, newName string) (*dto.FileResponse, error)
	MoveFile(ctx context.Context, userID uint64, fileID uint64, newPath string) (*dto.FileResponse, error)
}

type fileService struct {
	fileRepo     db.FileRepository
	uploadRepo   db.UploadTaskRepository
	repoRepo     db.RepoRepository
}

func NewFileService(fileRepo db.FileRepository, uploadRepo db.UploadTaskRepository, repoRepo db.RepoRepository) FileService {
	return &fileService{
		fileRepo:   fileRepo,
		uploadRepo: uploadRepo,
		repoRepo:   repoRepo,
	}
}

func (s *fileService) UploadSimple(ctx context.Context, repoID uint64, uploaderID uint64, filePath string, file multipart.File, header *multipart.FileHeader) (*dto.FileResponse, error) {
	if _, err := s.repoRepo.GetByID(repoID); err != nil {
		return nil, errors.New("仓库不存在")
	}

	existing, _ := s.fileRepo.GetByPath(repoID, filePath)
	if existing != nil && existing.ID > 0 {
		objectKey := storage.BuildObjectKey(repoID, filePath)
		if err := storage.DeleteFile(ctx, objectKey); err != nil {
			return nil, fmt.Errorf("删除旧文件失败: %v", err)
		}
		existing.FileSize = header.Size
		existing.MimeType = header.Header.Get("Content-Type")
		if existing.MimeType == "" {
			existing.MimeType = "application/octet-stream"
		}
		existing.UploaderID = uploaderID
		if err := s.fileRepo.Update(existing); err != nil {
			return nil, err
		}
		objectKey = storage.BuildObjectKey(repoID, filePath)
		if err := storage.UploadFileFromMultipart(ctx, objectKey, file, header); err != nil {
			return nil, fmt.Errorf("文件上传失败: %v", err)
		}
		return s.convertToResponse(existing), nil
	}

	objectKey := storage.BuildObjectKey(repoID, filePath)
	if err := storage.UploadFileFromMultipart(ctx, objectKey, file, header); err != nil {
		return nil, fmt.Errorf("文件上传失败: %v", err)
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	fileEntity := &entity.File{
		RepoID:     repoID,
		FileName:   header.Filename,
		Path:       filePath,
		IsDir:      false,
		MimeType:   mimeType,
		FileSize:   header.Size,
		ObjectKey:  objectKey,
		Status:     "completed",
		UploaderID: uploaderID,
	}

	if err := s.fileRepo.Create(fileEntity); err != nil {
		_ = storage.DeleteFile(ctx, objectKey)
		return nil, fmt.Errorf("文件记录保存失败: %v", err)
	}

	return s.convertToResponse(fileEntity), nil
}

func (s *fileService) UploadInit(ctx context.Context, uploaderID uint64, req dto.UploadInitRequest) (*dto.UploadInitResponse, error) {
	if _, err := s.repoRepo.GetByID(req.RepoID); err != nil {
		return nil, errors.New("仓库不存在")
	}

	uploadID := generateUploadID(req.RepoID, uploaderID, req.FileName, req.FileSize)

	existingTask, _ := s.uploadRepo.GetByUploadID(uploadID)
	if existingTask != nil && existingTask.ID > 0 {
		if existingTask.Status == "uploading" {
			return &dto.UploadInitResponse{
				UploadID:       uploadID,
				UploadedChunks: existingTask.UploadedChunks,
			}, nil
		}
	}

	task := &entity.UploadTask{
		UploadID:   uploadID,
		RepoID:     req.RepoID,
		FileName:   req.FileName,
		FilePath:   req.FilePath,
		FileSize:   req.FileSize,
		ChunkSize:  req.ChunkSize,
		TotalChunks: req.TotalChunks,
		MimeType:   req.MimeType,
		Status:     "uploading",
		UploaderID: uploaderID,
	}

	if err := s.uploadRepo.Create(task); err != nil {
		return nil, fmt.Errorf("创建上传任务失败: %v", err)
	}

	return &dto.UploadInitResponse{
		UploadID:       uploadID,
		UploadedChunks: 0,
	}, nil
}

func (s *fileService) UploadChunk(ctx context.Context, req dto.UploadChunkRequest, data []byte) (*dto.UploadChunkResponse, error) {
	task, err := s.uploadRepo.GetByUploadID(req.UploadID)
	if err != nil {
		return nil, errors.New("上传任务不存在")
	}
	if task.Status != "uploading" {
		return nil, errors.New("上传任务已结束")
	}
	if req.ChunkIndex < 0 || req.ChunkIndex >= task.TotalChunks {
		return nil, errors.New("分块索引超出范围")
	}

	uploadedSet := parseChunkIndices(task.UploadedChunkIndices)
	if uploadedSet[req.ChunkIndex] {
		return &dto.UploadChunkResponse{
			UploadID:       req.UploadID,
			ChunkIndex:     req.ChunkIndex,
			UploadedChunks: len(uploadedSet),
			Completed:      len(uploadedSet) >= task.TotalChunks,
		}, nil
	}

	if err := storage.UploadChunk(ctx, req.UploadID, req.ChunkIndex, data); err != nil {
		return nil, fmt.Errorf("分块上传失败: %v", err)
	}

	uploadedSet[req.ChunkIndex] = true
	newUploadedChunks := len(uploadedSet)
	newIndices := serializeChunkIndices(uploadedSet)

	if err := s.uploadRepo.UpdateUploadedChunks(req.UploadID, newUploadedChunks, newIndices); err != nil {
		return nil, err
	}

	completed := newUploadedChunks >= task.TotalChunks
	if completed {
		_ = s.uploadRepo.UpdateStatus(req.UploadID, "merging")
	}

	return &dto.UploadChunkResponse{
		UploadID:       req.UploadID,
		ChunkIndex:     req.ChunkIndex,
		UploadedChunks: newUploadedChunks,
		Completed:      completed,
	}, nil
}

func (s *fileService) UploadComplete(ctx context.Context, req dto.UploadCompleteRequest) (*dto.FileResponse, error) {
	task, err := s.uploadRepo.GetByUploadID(req.UploadID)
	if err != nil {
		return nil, errors.New("上传任务不存在")
	}
	if task.Status != "merging" && task.Status != "uploading" {
		return nil, fmt.Errorf("上传任务状态异常: %s", task.Status)
	}
	if task.UploadedChunks < task.TotalChunks {
		return nil, fmt.Errorf("分块未全部上传: %d/%d", task.UploadedChunks, task.TotalChunks)
	}

	objectKey := storage.BuildObjectKey(task.RepoID, task.FilePath)

	existing, _ := s.fileRepo.GetByPath(task.RepoID, task.FilePath)
	if existing != nil && existing.ID > 0 {
		_ = storage.DeleteFile(ctx, existing.ObjectKey)
	}

	if err := storage.MergeChunks(ctx, req.UploadID, task.TotalChunks, objectKey); err != nil {
		return nil, fmt.Errorf("分块合并失败: %v", err)
	}

	_ = storage.DeleteChunks(ctx, req.UploadID, task.TotalChunks)

	if existing != nil && existing.ID > 0 {
		existing.FileName = task.FileName
		existing.FileSize = task.FileSize
		existing.MimeType = task.MimeType
		existing.ObjectKey = objectKey
		existing.Status = "completed"
		existing.UploaderID = task.UploaderID
		if err := s.fileRepo.Update(existing); err != nil {
			return nil, err
		}
		_ = s.uploadRepo.UpdateStatus(req.UploadID, "completed")
		return s.convertToResponse(existing), nil
	}

	fileEntity := &entity.File{
		RepoID:     task.RepoID,
		FileName:   task.FileName,
		Path:       task.FilePath,
		IsDir:      false,
		MimeType:   task.MimeType,
		FileSize:   task.FileSize,
		ObjectKey:  objectKey,
		UploadID:   req.UploadID,
		ChunkCount: task.TotalChunks,
		Status:     "completed",
		UploaderID: task.UploaderID,
	}

	if err := s.fileRepo.Create(fileEntity); err != nil {
		return nil, fmt.Errorf("文件记录保存失败: %v", err)
	}

	_ = s.uploadRepo.UpdateStatus(req.UploadID, "completed")
	return s.convertToResponse(fileEntity), nil
}

func (s *fileService) GetFileDetail(ctx context.Context, fileID uint64) (*dto.FileResponse, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, errors.New("文件不存在")
	}
	return s.convertToResponse(file), nil
}

func (s *fileService) ListFiles(ctx context.Context, repoID uint64, parentPath string) (*dto.FileListResponse, error) {
	if _, err := s.repoRepo.GetByID(repoID); err != nil {
		return nil, errors.New("仓库不存在")
	}

	files, err := s.fileRepo.ListByRepo(repoID, parentPath)
	if err != nil {
		return nil, err
	}

	items := make([]dto.FileResponse, 0, len(files))
	for _, f := range files {
		items = append(items, *s.convertToResponse(&f))
	}

	return &dto.FileListResponse{
		Total: int64(len(items)),
		Items: items,
	}, nil
}

func (s *fileService) DownloadFile(ctx context.Context, fileID uint64) (io.Reader, *entity.File, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, nil, errors.New("文件不存在")
	}
	if file.IsDir {
		return nil, nil, errors.New("不能下载目录")
	}
	if file.ObjectKey == "" {
		return nil, nil, errors.New("文件存储键为空")
	}

	obj, err := storage.DownloadFile(ctx, file.ObjectKey)
	if err != nil {
		return nil, nil, fmt.Errorf("获取文件内容失败: %v", err)
	}

	return obj, file, nil
}

func (s *fileService) DeleteFile(ctx context.Context, userID uint64, fileID uint64) error {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return errors.New("文件不存在")
	}

	repo, err := s.repoRepo.GetByID(file.RepoID)
	if err != nil {
		return errors.New("仓库不存在")
	}
	if repo.OwnerID != userID {
		return errors.New("权限不足，只有仓库所有者可以删除文件")
	}

	if !file.IsDir && file.ObjectKey != "" {
		_ = storage.DeleteFile(ctx, file.ObjectKey)
	}

	return s.fileRepo.Delete(fileID)
}

func (s *fileService) CreateDir(ctx context.Context, userID uint64, req dto.CreateDirRequest) (*dto.FileResponse, error) {
	if _, err := s.repoRepo.GetByID(req.RepoID); err != nil {
		return nil, errors.New("仓库不存在")
	}

	dirPath := filepath.Join(req.Path, req.DirName)
	dirPath = strings.ReplaceAll(dirPath, "\\", "/")

	existing, _ := s.fileRepo.GetByPath(req.RepoID, dirPath)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("目录已存在")
	}

	dirEntity := &entity.File{
		RepoID:     req.RepoID,
		FileName:   req.DirName,
		Path:       dirPath,
		IsDir:      true,
		MimeType:   "directory",
		FileSize:   0,
		Status:     "completed",
		UploaderID: userID,
	}

	if err := s.fileRepo.Create(dirEntity); err != nil {
		return nil, fmt.Errorf("目录创建失败: %v", err)
	}

	return s.convertToResponse(dirEntity), nil
}

func (s *fileService) RenameFile(ctx context.Context, userID uint64, fileID uint64, newName string) (*dto.FileResponse, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, errors.New("文件不存在")
	}

	repo, err := s.repoRepo.GetByID(file.RepoID)
	if err != nil {
		return nil, errors.New("仓库不存在")
	}
	if repo.OwnerID != userID {
		return nil, errors.New("权限不足")
	}

	parentDir := filepath.Dir(file.Path)
	parentDir = strings.ReplaceAll(parentDir, "\\", "/")
	newPath := filepath.Join(parentDir, newName)
	newPath = strings.ReplaceAll(newPath, "\\", "/")

	existing, _ := s.fileRepo.GetByPath(file.RepoID, newPath)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("目标路径已存在同名文件")
	}

	oldObjectKey := file.ObjectKey
	file.FileName = newName
	file.Path = newPath
	if !file.IsDir {
		file.ObjectKey = storage.BuildObjectKey(file.RepoID, newPath)
		if oldObjectKey != "" {
			obj, err := storage.DownloadFile(ctx, oldObjectKey)
			if err == nil {
				info, err := storage.GetFileInfo(ctx, oldObjectKey)
				if err == nil {
					_ = storage.UploadFile(ctx, file.ObjectKey, obj, info.Size, info.ContentType)
					_ = storage.DeleteFile(ctx, oldObjectKey)
				}
			}
		}
	}

	if err := s.fileRepo.Update(file); err != nil {
		return nil, err
	}

	return s.convertToResponse(file), nil
}

func (s *fileService) MoveFile(ctx context.Context, userID uint64, fileID uint64, newPath string) (*dto.FileResponse, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, errors.New("文件不存在")
	}

	repo, err := s.repoRepo.GetByID(file.RepoID)
	if err != nil {
		return nil, errors.New("仓库不存在")
	}
	if repo.OwnerID != userID {
		return nil, errors.New("权限不足")
	}

	newPath = strings.ReplaceAll(newPath, "\\", "/")
	existing, _ := s.fileRepo.GetByPath(file.RepoID, newPath)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("目标路径已存在同名文件")
	}

	oldObjectKey := file.ObjectKey
	file.Path = newPath
	if !file.IsDir {
		file.ObjectKey = storage.BuildObjectKey(file.RepoID, newPath)
		if oldObjectKey != "" {
			obj, err := storage.DownloadFile(ctx, oldObjectKey)
			if err == nil {
				info, err := storage.GetFileInfo(ctx, oldObjectKey)
				if err == nil {
					_ = storage.UploadFile(ctx, file.ObjectKey, obj, info.Size, info.ContentType)
					_ = storage.DeleteFile(ctx, oldObjectKey)
				}
			}
		}
	}

	if err := s.fileRepo.Update(file); err != nil {
		return nil, err
	}

	return s.convertToResponse(file), nil
}

func (s *fileService) convertToResponse(file *entity.File) *dto.FileResponse {
	return &dto.FileResponse{
		ID:         file.ID,
		RepoID:     file.RepoID,
		FileName:   file.FileName,
		Path:       file.Path,
		IsDir:      file.IsDir,
		MimeType:   file.MimeType,
		FileSize:   file.FileSize,
		Status:     file.Status,
		UploaderID: file.UploaderID,
		CreatedAt:  file.CreatedAt,
		UpdatedAt:  file.UpdatedAt,
	}
}

func generateUploadID(repoID uint64, uploaderID uint64, fileName string, fileSize int64) string {
	raw := fmt.Sprintf("%d_%d_%s_%d_%d", repoID, uploaderID, fileName, fileSize, time.Now().UnixNano())
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h[:16])
}

func parseChunkIndices(s string) map[int]bool {
	result := make(map[int]bool)
	if s == "" {
		return result
	}
	parts := strings.Split(s, ",")
	for _, p := range parts {
		if idx, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			result[idx] = true
		}
	}
	return result
}

func serializeChunkIndices(m map[int]bool) string {
	indices := make([]int, 0, len(m))
	for k := range m {
		indices = append(indices, k)
	}
	sort.Ints(indices)
	parts := make([]string, len(indices))
	for i, idx := range indices {
		parts[i] = strconv.Itoa(idx)
	}
	return strings.Join(parts, ",")
}
