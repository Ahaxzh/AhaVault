// Package handlers 提供 HTTP 请求处理器
//
// 本文件实现了基于 Tus 协议的文件上传处理器，支持：
//   - 分片上传（Chunked Upload）
//   - 断点续传（Resumable Upload）
//   - 上传进度查询
//   - 上传完成后自动触发加密存储
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"ahavault/server/internal/middleware"
	"ahavault/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler Tus 上传处理器
type UploadHandler struct {
	fileService *services.FileService
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(fileService *services.FileService) *UploadHandler {
	return &UploadHandler{
		fileService: fileService,
	}
}

// CreateUploadRequest 创建上传会话请求
type CreateUploadRequest struct {
	Filename string `json:"filename" binding:"required"`
	Size     int64  `json:"size" binding:"required,gt=0"`
	Hash     string `json:"hash" binding:"required"` // SHA-256 哈希值
}

// CreateUpload 创建上传会话
//
// 该函数实现 Tus 协议的上传会话创建：
//  1. 验证用户存储配额
//  2. 检查是否可以秒传（文件已存在）
//  3. 创建临时上传会话（使用 Redis 或内存存储）
//  4. 返回上传 URL 和会话 ID
//
// 端点: POST /api/tus/upload
//
// 参数:
//   - c: Gin 上下文对象
//
// 返回:
//   - 201: 上传会话创建成功
//   - 400: 请求参数错误
//   - 401: 用户未认证
//   - 507: 用户配额不足
func (h *UploadHandler) CreateUpload(c *gin.Context) {
	var req CreateUploadRequest
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

	// 检查秒传
	exists, _, err := h.fileService.CheckInstantUpload(req.Hash, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	if exists {
		// 秒传成功，直接创建元数据
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
			"message": "Instant upload successful",
			"data": gin.H{
				"instant_upload": true,
				"file":           metadata,
			},
		})
		return
	}

	// 创建上传会话 ID
	uploadID := uuid.New().String()

	// 返回上传会话信息
	c.Header("Location", fmt.Sprintf("/api/tus/upload/%s", uploadID))
	c.Header("Tus-Resumable", "1.0.0")
	c.Header("Upload-Offset", "0")
	c.Header("Upload-Length", strconv.FormatInt(req.Size, 10))

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "Upload session created",
		"data": gin.H{
			"upload_id":    uploadID,
			"upload_url":   fmt.Sprintf("/api/tus/upload/%s", uploadID),
			"upload_offset": 0,
			"upload_length": req.Size,
			"filename":     req.Filename,
			"hash":         req.Hash,
		},
	})
}

// UploadChunk 上传文件分片
//
// 该函数实现 Tus 协议的 PATCH 方法：
//  1. 验证 Content-Length 和 Upload-Offset 头
//  2. 接收文件分片数据
//  3. 追加到临时文件
//  4. 更新上传进度
//  5. 若上传完成，触发文件加密和存储
//
// 端点: PATCH /api/tus/upload/:id
//
// 参数:
//   - c: Gin 上下文对象
//
// 返回:
//   - 204: 分片上传成功
//   - 200: 文件上传完成
//   - 400: 请求参数错误
//   - 404: 上传会话不存在
//   - 409: Upload-Offset 不匹配
func (h *UploadHandler) UploadChunk(c *gin.Context) {
	uploadID := c.Param("id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Missing upload ID",
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

	// 获取请求头信息
	contentLengthStr := c.GetHeader("Content-Length")
	uploadOffsetStr := c.GetHeader("Upload-Offset")

	contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil || contentLength <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid Content-Length header",
		})
		return
	}

	uploadOffset, err := strconv.ParseInt(uploadOffsetStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid Upload-Offset header",
		})
		return
	}

	// 读取请求体数据
	body := c.Request.Body
	defer body.Close()

	// 限制读取大小（防止恶意请求）
	limitedReader := io.LimitReader(body, contentLength)

	// TODO: 实现实际的分片存储逻辑
	// 这里需要：
	// 1. 将数据追加到临时文件（/tmp/uploads/{uploadID}）
	// 2. 更新上传进度到 Redis
	// 3. 检查是否上传完成
	// 4. 若完成，调用 fileService.UploadFile 进行加密存储

	// 示例：直接上传完整文件（简化实现）
	filename := c.GetHeader("Upload-Metadata-Filename")
	sizeStr := c.GetHeader("Upload-Metadata-Size")
	size, _ := strconv.ParseInt(sizeStr, 10, 64)

	if filename == "" {
		filename = "untitled"
	}

	metadata, err := h.fileService.UploadFile(userUUID, filename, size, limitedReader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Tus-Resumable", "1.0.0")
	c.Header("Upload-Offset", strconv.FormatInt(uploadOffset+contentLength, 10))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Upload completed successfully",
		"data":    metadata,
	})
}

// GetUploadProgress 查询上传进度
//
// 该函数实现 Tus 协议的 HEAD 方法：
//  1. 查询上传会话信息
//  2. 返回当前上传偏移量和总大小
//
// 端点: HEAD /api/tus/upload/:id
//
// 参数:
//   - c: Gin 上下文对象
//
// 返回:
//   - 200: 成功返回上传进度
//   - 404: 上传会话不存在
func (h *UploadHandler) GetUploadProgress(c *gin.Context) {
	uploadID := c.Param("id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Missing upload ID",
		})
		return
	}

	// TODO: 从 Redis 或内存存储中获取上传进度
	// 这里返回模拟数据
	uploadOffset := int64(0)
	uploadLength := int64(1024000)

	c.Header("Tus-Resumable", "1.0.0")
	c.Header("Upload-Offset", strconv.FormatInt(uploadOffset, 10))
	c.Header("Upload-Length", strconv.FormatInt(uploadLength, 10))
	c.Header("Cache-Control", "no-store")

	c.Status(http.StatusOK)
}

// DeleteUpload 删除上传会话
//
// 该函数实现 Tus 协议的 DELETE 方法：
//  1. 删除临时文件
//  2. 清除上传会话信息
//
// 端点: DELETE /api/tus/upload/:id
//
// 参数:
//   - c: Gin 上下文对象
//
// 返回:
//   - 204: 删除成功
//   - 404: 上传会话不存在
func (h *UploadHandler) DeleteUpload(c *gin.Context) {
	uploadID := c.Param("id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Missing upload ID",
		})
		return
	}

	// TODO: 删除临时文件和 Redis 会话数据

	c.Header("Tus-Resumable", "1.0.0")
	c.Status(http.StatusNoContent)
}

// Options 处理 OPTIONS 请求（CORS 预检）
//
// 该函数实现 Tus 协议的 OPTIONS 方法：
//  1. 返回支持的 Tus 版本
//  2. 返回支持的扩展
//
// 端点: OPTIONS /api/tus/upload
//
// 参数:
//   - c: Gin 上下文对象
//
// 返回:
//   - 204: 成功
func (h *UploadHandler) Options(c *gin.Context) {
	c.Header("Tus-Resumable", "1.0.0")
	c.Header("Tus-Version", "1.0.0")
	c.Header("Tus-Extension", "creation,termination")
	c.Header("Tus-Max-Size", "10737418240") // 10GB

	c.Status(http.StatusNoContent)
}

// parseUploadMetadata 解析 Upload-Metadata 头
//
// Tus 协议定义的元数据格式: "key1 value1,key2 value2"
func parseUploadMetadata(metadata string) map[string]string {
	result := make(map[string]string)
	if metadata == "" {
		return result
	}

	pairs := strings.Split(metadata, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), " ", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}
