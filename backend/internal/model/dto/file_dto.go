package dto

import "time"

type UploadInitRequest struct {
	RepoID      uint64 `json:"repo_id" binding:"required"`
	FileName    string `json:"file_name" binding:"required,min=1,max=255"`
	FilePath    string `json:"file_path" binding:"required,min=1"`
	FileSize    int64  `json:"file_size" binding:"required,gt=0"`
	ChunkSize   int64  `json:"chunk_size" binding:"required,gt=0"`
	TotalChunks int    `json:"total_chunks" binding:"required,gt=0"`
	MimeType    string `json:"mime_type" binding:"max=100"`
}

type UploadInitResponse struct {
	UploadID    string `json:"upload_id"`
	UploadedChunks int `json:"uploaded_chunks"`
}

type UploadChunkRequest struct {
	UploadID    string `form:"upload_id" binding:"required"`
	ChunkIndex  int    `form:"chunk_index" binding:"required,gte=0"`
}

type UploadChunkResponse struct {
	UploadID       string `json:"upload_id"`
	ChunkIndex     int    `json:"chunk_index"`
	UploadedChunks int    `json:"uploaded_chunks"`
	Completed      bool   `json:"completed"`
}

type UploadCompleteRequest struct {
	UploadID string `json:"upload_id" binding:"required"`
}

type FileResponse struct {
	ID         uint64    `json:"id"`
	RepoID     uint64    `json:"repo_id"`
	FileName   string    `json:"file_name"`
	Path       string    `json:"path"`
	IsDir      bool      `json:"is_dir"`
	MimeType   string    `json:"mime_type"`
	FileSize   int64     `json:"file_size"`
	Status     string    `json:"status"`
	UploaderID uint64    `json:"uploader_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FileListQuery struct {
	RepoID uint64 `form:"repo_id" binding:"required"`
	Path   string `form:"path"`
}

type FileListResponse struct {
	Total int64          `json:"total"`
	Items []FileResponse `json:"items"`
}

type CreateDirRequest struct {
	RepoID  uint64 `json:"repo_id" binding:"required"`
	Path    string `json:"path" binding:"required,min=1"`
	DirName string `json:"dir_name" binding:"required,min=1,max=255"`
}

type RenameFileRequest struct {
	NewName string `json:"new_name" binding:"required,min=1,max=255"`
}

type MoveFileRequest struct {
	NewPath string `json:"new_path" binding:"required,min=1"`
}
