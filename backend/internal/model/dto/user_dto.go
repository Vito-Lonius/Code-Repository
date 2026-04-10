// internal/model/dto/user_dto.go
package dto

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
	Nickname string `json:"nickname" binding:"omitempty,min=1,max=100"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应（含JWT）
type LoginResponse struct {
	Token     string        `json:"token"`
	TokenType string        `json:"token_type"` // Bearer
	ExpiresAt int64         `json:"expires_at"` // Unix timestamp
	User      *UserResponse `json:"user"`
}

// UserResponse 用户信息响应（公开字段）
type UserResponse struct {
	ID        uint64 `json:"id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"` // ISO 8601
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname" binding:"omitempty,min=1,max=100"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,url"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=32"`
}
