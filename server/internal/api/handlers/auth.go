package handlers

import (
	"net/http"

	"ahavault/server/internal/middleware"
	"ahavault/server/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	InviteCode string `json:"invite_code"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"error":   err.Error(),
		})
		return
	}

	// 从系统配置读取是否需要邀请码
	// 开发阶段默认关闭邀请码功能
	requireInviteCode := false
	validInviteCode := ""

	serviceReq := &services.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.userService.Register(serviceReq, requireInviteCode, validInviteCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Registration successful",
		"data":    resp,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"error":   err.Error(),
		})
		return
	}

	serviceReq := &services.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.userService.Login(serviceReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Login successful",
		"data":    resp,
	})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Unauthorized",
		})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "User not found",
		})
		return
	}

	// 隐藏敏感字段
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Success",
		"data":    user,
	})
}

// Logout 用户登出（客户端删除 token）
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Logout successful",
	})
}
