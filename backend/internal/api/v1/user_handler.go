package v1

import (
	"net/http"

	"code-repo/internal/model/dto"
	"code-repo/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 初始化用户处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register 注册接口：POST /api/v1/register
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	// 1. 绑定并校验参数（基于 dto 中的 binding 标签）
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数校验失败: " + err.Error()})
		return
	}

	// 2. 调用 Service 层执行注册逻辑
	if err := h.svc.Register(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusCreated, gin.H{"message": "注册成功"})
}

// Login 登录接口：POST /api/v1/login
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 调用 Service 执行登录并获取 Token
	if resp, err := h.svc.Login(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, resp)
	}
}

// GetProfile 获取个人资料：GET /api/v1/user/profile
// 注意：此接口通常需要配合 Auth 中间件使用
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 从上下文获取中间件解析出的 userID (假设键名为 "userID")
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID := userIDVal.(uint64)

	resp, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
