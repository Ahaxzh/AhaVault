// Package handlers 提供 HTTP 请求处理器测试
//
// 本文件测试文件下载接口的功能：
//   - 通过取件码下载文件
//   - HTTP Range 请求支持
//   - 访问密码验证
//   - 下载次数限制
//   - 文件预览
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ahavault/server/internal/models"
	"ahavault/server/internal/services"
	"ahavault/server/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDownloadTestEnv 设置下载测试环境
func setupDownloadTestEnv(t *testing.T) (*DownloadHandler, *gin.Engine, *models.User, *models.FileMetadata, *models.ShareSession, func()) {
	// 初始化数据库
	db := setupTestDB(t)

	// 初始化存储引擎
	storageEngine := storage.NewMemoryEngine()

	// 初始化 KEK
	kek := []byte("test-master-key-1234567890123456")

	// 初始化服务
	fileService := services.NewFileService(db, storageEngine, kek)
	shareService := services.NewShareService(db, fileService)

	// 创建处理器
	handler := NewDownloadHandler(shareService, fileService)

	// 创建测试用户
	user := &models.User{
		Email:        "download@test.com",
		Password:     "hashed_password",
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(user).Error)

	// 上传测试文件
	fileContent := []byte("This is a test file for download.")
	metadata, err := fileService.UploadFile(user.ID, "test.txt", int64(len(fileContent)), bytes.NewReader(fileContent))
	require.NoError(t, err)

	// 创建分享
	share, err := shareService.CreateShare(user.ID, &services.CreateShareRequest{
		FileIDs:      []uuid.UUID{metadata.ID},
		ExpiresIn:    3600 * time.Second,
		MaxDownloads: 10,
		Password:     "",
	})
	require.NoError(t, err)

	// 设置 Gin 路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/api/download/:code", handler.DownloadByPickupCode)
	router.GET("/api/download/:code/preview", handler.DownloadPreview)

	cleanup := func() {
		// 清理资源
	}

	return handler, router, user, metadata, share, cleanup
}

// TestDownloadByPickupCode 测试通过取件码下载文件
func TestDownloadByPickupCode(t *testing.T) {
	_, router, _, _, share, cleanup := setupDownloadTestEnv(t)
	defer cleanup()

	tests := []struct {
		name           string
		pickupCode     string
		password       string
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "valid download",
			pickupCode:     share.PickupCode,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "attachment; filename=\"test.txt\"", w.Header().Get("Content-Disposition"))
				assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
				assert.NotEmpty(t, w.Header().Get("Content-Length"))
				assert.Equal(t, "bytes", w.Header().Get("Accept-Ranges"))
			},
		},
		{
			name:           "invalid pickup code",
			pickupCode:     "INVALID1",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/download/" + tt.pickupCode
			if tt.password != "" {
				url += "?password=" + tt.password
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

// TestDownloadWithRange 测试 Range 下载
func TestDownloadWithRange(t *testing.T) {
	_, router, _, _, share, cleanup := setupDownloadTestEnv(t)
	defer cleanup()

	tests := []struct {
		name           string
		rangeHeader    string
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "range request bytes=0-10",
			rangeHeader:    "bytes=0-10",
			expectedStatus: http.StatusPartialContent,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Header().Get("Content-Range"), "bytes 0-10")
				assert.Equal(t, "11", w.Header().Get("Content-Length"))

				body, _ := io.ReadAll(w.Body)
				assert.Len(t, body, 11)
			},
		},
		{
			name:           "range request bytes=5-",
			rangeHeader:    "bytes=5-",
			expectedStatus: http.StatusPartialContent,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Header().Get("Content-Range"), "bytes 5-")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/download/"+share.PickupCode, nil)
			require.NoError(t, err)

			if tt.rangeHeader != "" {
				req.Header.Set("Range", tt.rangeHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

// TestDownloadPreview 测试文件预览
func TestDownloadPreview(t *testing.T) {
	_, router, _, _, share, cleanup := setupDownloadTestEnv(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodGet, "/api/download/"+share.PickupCode+"/preview", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, share.PickupCode, data["pickup_code"])
	assert.NotNil(t, data["expires_at"])
}

// TestParseRange 测试 Range 解析函数
func TestParseRange(t *testing.T) {
	tests := []struct {
		name        string
		rangeHeader string
		size        int64
		expected    []httpRange
	}{
		{
			name:        "bytes=0-1023",
			rangeHeader: "bytes=0-1023",
			size:        2048,
			expected:    []httpRange{{start: 0, end: 1023}},
		},
		{
			name:        "bytes=1024-",
			rangeHeader: "bytes=1024-",
			size:        2048,
			expected:    []httpRange{{start: 1024, end: 2047}},
		},
		{
			name:        "bytes=-1024",
			rangeHeader: "bytes=-1024",
			size:        2048,
			expected:    []httpRange{{start: 1024, end: 2047}},
		},
		{
			name:        "invalid format",
			rangeHeader: "invalid",
			size:        2048,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRange(tt.rangeHeader, tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}
