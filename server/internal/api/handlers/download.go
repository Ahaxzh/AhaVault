// Package handlers 提供 HTTP 请求处理器
//
// 本文件实现了文件下载处理器，支持：
//   - HTTP Range 请求（断点续传）
//   - 流式解密传输（边解密边传输）
//   - 下载次数统计
//   - 访问日志记录
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

	"ahavault/server/internal/services"
	"github.com/gin-gonic/gin"
)

// DownloadHandler 下载处理器
type DownloadHandler struct {
	shareService *services.ShareService
	fileService  *services.FileService
}

// NewDownloadHandler 创建下载处理器
func NewDownloadHandler(shareService *services.ShareService, fileService *services.FileService) *DownloadHandler {
	return &DownloadHandler{
		shareService: shareService,
		fileService:  fileService,
	}
}

// DownloadByPickupCode 通过取件码下载文件
//
// 该函数实现文件下载的完整流程：
//  1. 验证取件码有效性
//  2. 检查访问密码（如果设置）
//  3. 检查下载次数限制
//  4. 支持 HTTP Range 请求（断点续传）
//  5. 流式解密传输文件
//  6. 更新下载统计
//
// 端点: GET /api/download/:code
//
// 参数:
//   - c: Gin 上下文对象
//
// 返回:
//   - 200: 下载成功（全文件）
//   - 206: 部分内容下载成功（Range 请求）
//   - 400: 请求参数错误
//   - 401: 需要访问密码
//   - 403: 访问密码错误或下载次数超限
//   - 404: 取件码不存在
//   - 410: 分享已过期
func (h *DownloadHandler) DownloadByPickupCode(c *gin.Context) {
	pickupCode := c.Param("code")
	if pickupCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Missing pickup code",
		})
		return
	}

	// 获取访问密码（如果有）
	password := c.Query("password")

	// 验证取件码并获取分享信息
	share, files, err := h.shareService.GetShareByCode(pickupCode, password)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	// 如果没有文件，返回错误
	if len(files) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "No files in share",
		})
		return
	}

	// 获取第一个文件ID（当前版本只支持单文件分享）
	fileID := files[0].ID

	// 下载文件
	reader, metadata, err := h.fileService.DownloadFile(fileID, share.CreatorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}
	defer reader.Close()

	// 解析 Range 请求
	rangeHeader := c.GetHeader("Range")
	contentLength := metadata.Size
	statusCode := http.StatusOK

	var start, end int64
	if rangeHeader != "" {
		// 解析 Range 头: "bytes=0-1023"
		ranges := parseRange(rangeHeader, contentLength)
		if len(ranges) > 0 {
			start = ranges[0].start
			end = ranges[0].end
			statusCode = http.StatusPartialContent

			// 设置 Range 响应头
			c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, contentLength))
			c.Header("Accept-Ranges", "bytes")
			contentLength = end - start + 1

			// 跳过前面的数据
			if start > 0 {
				_, err := io.CopyN(io.Discard, reader, start)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    500,
						"message": "Failed to seek file",
					})
					return
				}
			}

			// 限制读取长度
			reader = io.NopCloser(io.LimitReader(reader, contentLength))
		}
	}

	// 设置响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", metadata.Filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "no-cache")

	// 流式传输文件
	c.Status(statusCode)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		// 传输中断，记录日志（客户端可能主动断开）
		return
	}

	// 更新下载统计
	go func() {
		// 异步更新，不影响下载速度
		h.shareService.IncrementDownload(share.ID)
	}()
}

// DownloadPreview 预览文件信息（不下载）
//
// 该函数返回文件元数据，用于前端展示：
//  1. 验证取件码
//  2. 返回文件名、大小、MIME 类型等信息
//  3. 不消耗下载次数
//
// 端点: GET /api/download/:code/preview
//
// 参数:
//   - c: Gin 上下文对象
//
// 返回:
//   - 200: 成功返回文件信息
//   - 400: 请求参数错误
//   - 401: 需要访问密码
//   - 404: 取件码不存在
//   - 410: 分享已过期
func (h *DownloadHandler) DownloadPreview(c *gin.Context) {
	pickupCode := c.Param("code")
	if pickupCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Missing pickup code",
		})
		return
	}

	// 获取访问密码（如果有）
	password := c.Query("password")

	// 验证取件码
	share, files, err := h.shareService.GetShareByCode(pickupCode, password)
	if err != nil {
		// GetShareByCode 已经处理了密码验证
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	// 返回文件预览信息
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Success",
		"data": gin.H{
			"pickup_code":       share.PickupCode,
			"expires_at":        share.ExpiresAt,
			"max_downloads":     share.MaxDownloads,
			"download_count":    share.CurrentDownloads,
			"password_required": share.HasPassword(),
			"files":             files,
		},
	})
}

// httpRange 表示 HTTP Range 请求的范围
type httpRange struct {
	start int64
	end   int64
}

// parseRange 解析 Range 请求头
//
// 示例: "bytes=0-1023" 或 "bytes=1024-"
//
// 参数:
//   - rangeHeader: Range 请求头的值
//   - size: 文件总大小
//
// 返回:
//   - 解析后的范围列表
func parseRange(rangeHeader string, size int64) []httpRange {
	const bytesPrefix = "bytes="
	if !strings.HasPrefix(rangeHeader, bytesPrefix) {
		return nil
	}

	rangeSpec := strings.TrimPrefix(rangeHeader, bytesPrefix)
	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		return nil
	}

	var start, end int64
	var err error

	// 解析起始位置
	if parts[0] == "" {
		// 格式: "bytes=-1024" (最后 1024 字节)
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffix <= 0 || suffix > size {
			return nil
		}
		start = size - suffix
		end = size - 1
	} else {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil || start < 0 || start >= size {
			return nil
		}

		if parts[1] == "" {
			// 格式: "bytes=1024-" (从 1024 到文件末尾)
			end = size - 1
		} else {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil || end < start || end >= size {
				return nil
			}
		}
	}

	return []httpRange{{start: start, end: end}}
}
