package handlers

import (
	"io"
	"net/http"
	"strconv"

	"ahavault/server/internal/middleware"
	"ahavault/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FileHandler 文件处理器
type FileHandler struct {
	fileService *services.FileService
}

// NewFileHandler 创建文件处理器
func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// CheckInstantUploadRequest 秒传检测请求
type CheckInstantUploadRequest struct {
	Hash string `json:"hash" binding:"required"`
}

// CreateFileMetadataRequest 创建文件元数据请求（秒传）
type CreateFileMetadataRequest struct {
	Hash     string `json:"hash" binding:"required"`
	Filename string `json:"filename" binding:"required"`
	Size     int64  `json:"size" binding:"required"`
}

// CheckInstantUpload 秒传检测
func (h *FileHandler) CheckInstantUpload(c *gin.Context) {
	var req CheckInstantUploadRequest
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

	exists, blob, err := h.fileService.CheckInstantUpload(req.Hash, userUUID)
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
			"exists": exists,
			"blob":   blob,
		},
	})
}

// CreateFileMetadata 创建文件元数据（秒传）
func (h *FileHandler) CreateFileMetadata(c *gin.Context) {
	var req CreateFileMetadataRequest
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

	metadata, err := h.fileService.CreateFileMetadata(userUUID, req.Hash, req.Filename, req.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "File created successfully",
		"data":    metadata,
	})
}

// UploadFile 上传文件
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "Invalid user ID",
		})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "No file uploaded",
			"error":   err.Error(),
		})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to open file",
			"error":   err.Error(),
		})
		return
	}
	defer src.Close()

	// 上传文件
	metadata, err := h.fileService.UploadFile(userUUID, file.Filename, file.Size, src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "File uploaded successfully",
		"data":    metadata,
	})
}

// DownloadFile 下载文件
func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileID := c.Param("id")
	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid file ID",
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

	reader, metadata, err := h.fileService.DownloadFile(fileUUID, userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+metadata.Filename)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(metadata.Size, 10))

	// 流式传输文件
	if _, err := io.Copy(c.Writer, reader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to send file",
		})
		return
	}
}

// DeleteFile 删除文件
func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid file ID",
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

	if err := h.fileService.DeleteFile(fileUUID, userUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "File deleted successfully",
	})
}

// ListFiles 获取文件列表
func (h *FileHandler) ListFiles(c *gin.Context) {
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

	files, total, err := h.fileService.ListFiles(userUUID, page, pageSize)
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
			"files": files,
			"total": total,
			"page":  page,
			"page_size": pageSize,
		},
	})
}
