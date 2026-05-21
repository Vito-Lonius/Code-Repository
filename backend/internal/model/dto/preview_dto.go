package dto

type PreviewResponse struct {
	FileID   uint64 `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
	Content  string `json:"content,omitempty"`
	Language string `json:"language,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}

type TreeNode struct {
	ID       uint64       `json:"id"`
	Name     string       `json:"name"`
	Path     string       `json:"path"`
	IsDir    bool         `json:"is_dir"`
	MimeType string       `json:"mime_type,omitempty"`
	FileSize int64        `json:"file_size,omitempty"`
	Children []*TreeNode  `json:"children,omitempty"`
}

type TreeResponse struct {
	RepoID uint64      `json:"repo_id"`
	Tree   []*TreeNode `json:"tree"`
}

type RawFileResponse struct {
	Content     []byte
	ContentType string
	FileName    string
	FileSize    int64
}
