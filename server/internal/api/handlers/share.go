package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ahavault/server/internal/middleware"
	"ahavault/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ShareHandler 分享处理器
type ShareHandler struct {
	shareService *services.ShareService
}

// NewShareHandler 创建分享处理器
func NewShareHandler(shareService *services.ShareService) *ShareHandler {
	return &ShareHandler{
		shareService: shareService,
	}
}

// CreateShareRequest 创建分享请求
type CreateShareRequest struct {
	FileIDs      []string `json:"file_ids" binding:"required"`
	ExpiresIn    int64    `json:"expires_in" binding:"required"` // 秒数
	MaxDownloads int      `json:"max_downloads"`
	Password     string   `json:"password"`
}

// GetShareRequest 获取分享请求
type GetShareRequest struct {
	Password string `json:"password"`
}

// SaveToVaultRequest 转存请求
type SaveToVaultRequest struct {
	FileIDs  []string `json:"file_ids" binding:"required"`
	Password string   `json:"password"`
}

// CreateShare 创建分享
func (h *ShareHandler) CreateShare(c *gin.Context) {
	var req CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"error":   err.Error(),
		})
		return
	}

	userID := middleware.GetUserID(c)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid user ID",
		})
		return
	}

	// 转换文件 ID
	fileIDs := make([]uuid.UUID, len(req.FileIDs))
	for i, id := range req.FileIDs {
		fileUUID, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "Invalid file ID: " + id,
			})
			return
		}
		fileIDs[i] = fileUUID
	}

	// 创建分享
	serviceReq := &services.CreateShareRequest{
		FileIDs:      fileIDs,
		ExpiresIn:    time.Duration(req.ExpiresIn) * time.Second,
		MaxDownloads: req.MaxDownloads,
		Password:     req.Password,
	}

	session, err := h.shareService.CreateShare(userUUID, serviceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Share created successfully",
		"data":    session,
	})
}

// GetShareByCode 通过取件码获取分享
func (h *ShareHandler) GetShareByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Pickup code is required",
		})
		return
	}

	var req GetShareRequest
	c.ShouldBindJSON(&req)

	session, files, err := h.shareService.GetShareByCode(code, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Success",
		"data": gin.H{
			"session": session,
			"files":   files,
		},
	})
}

// SaveToVault 转存到文件柜
func (h *ShareHandler) SaveToVault(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Pickup code is required",
		})
		return
	}

	var req SaveToVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"error":   err.Error(),
		})
		return
	}

	userID := middleware.GetUserID(c)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid user ID",
		})
		return
	}

	// 转换文件 ID
	fileIDs := make([]uuid.UUID, len(req.FileIDs))
	for i, id := range req.FileIDs {
		fileUUID, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "Invalid file ID: " + id,
			})
			return
		}
		fileIDs[i] = fileUUID
	}

	savedIDs, err := h.shareService.SaveToVault(code, req.Password, fileIDs, userUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Files saved to vault successfully",
		"data": gin.H{
			"saved_ids": savedIDs,
		},
	})
}

// StopShare 停止分享
func (h *ShareHandler) StopShare(c *gin.Context) {
	shareID := c.Param("id")
	shareUUID, err := uuid.Parse(shareID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid share ID",
		})
		return
	}

	userID := middleware.GetUserID(c)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid user ID",
		})
		return
	}

	if err := h.shareService.StopShare(shareUUID, userUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Share stopped successfully",
	})
}

// ListMyShares 获取我的分享列表
func (h *ShareHandler) ListMyShares(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid user ID",
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	shares, total, err := h.shareService.ListMyShares(userUUID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Success",
		"data": gin.H{
			"shares":    shares,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
