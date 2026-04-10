package dto

type CreateRepositoryRequest struct {
	Name       string `json:"name" binding:"required,min=1,max=255"`
	Visibility string `json:"visibility" binding:"required,oneof=public private"`
	OwnerID    uint64 `json:"owner_id" binding:"required"`
}

type UpdateRepositoryRequest struct {
	Name       string `json:"name" binding:"omitempty,min=1,max=255"`
	Visibility string `json:"visibility" binding:"omitempty,oneof=public private"`
}

type DeleteRepositoryRequest struct {
	ID uint64 `json:"id" binding:"required"`
}

type RepositoryResponse struct {
	ID         uint64 `json:"id"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
	OwnerID    uint64 `json:"owner_id"`
}
