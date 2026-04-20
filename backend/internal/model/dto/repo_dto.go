package dto

import "time"

// CreateRepoRequest 创建仓库的请求体
type CreateRepoRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"` // 仓库名，必填
	Description string `json:"description" binding:"max=255"`         // 描述，选填
	IsPublic    bool   `json:"is_public"`                             // 是否公开，默认为 true
}

// UpdateRepoRequest 更新仓库信息的请求体
type UpdateRepoRequest struct {
	Description *string `json:"description" binding:"max=255"` // 使用指针允许仅更新该字段或传空字符串
	IsPublic    *bool   `json:"is_public"`                     // 使用指针区分“不更新”与“设置为 false”
}

// RepoResponse 仓库详情返回体
type RepoResponse struct {
	ID            uint64    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	OwnerID       uint64    `json:"owner_id"`
	OwnerNickname string    `json:"owner_nickname"`
	IsPublic      bool      `json:"is_public"`
	DefaultBranch string    `json:"default_branch"`
	CloneURL      string    `json:"clone_url"` // HTTP 克隆地址，如 http://localhost:8080/git/user/repo.git
	StarCount     int       `json:"star_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RepoListQuery 仓库列表查询参数（用于分页和搜索）
type RepoListQuery struct {
	Page     int    `form:"page,default=1"`       // 第几页
	PageSize int    `form:"page_size,default=10"` // 每页条数
	Keyword  string `form:"keyword"`              // 搜索关键词
	OwnerID  uint   `form:"owner_id"`             // 按用户筛选
}

// RepoListResponse 仓库列表分页返回
type RepoListResponse struct {
	Total int64          `json:"total"`
	Items []RepoResponse `json:"items"`
}
