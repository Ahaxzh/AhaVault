// Package handlers 提供 HTTP 请求处理器测试
//
// 本文件测试 Tus 上传接口的功能：
//   - 创建上传会话
//   - 分片上传
//   - 查询上传进度
//   - 删除上传会话
//   - 秒传检测
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ahavault/server/internal/models"
	"ahavault/server/internal/services"
	"ahavault/server/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUploadTestEnv 设置上传测试环境
func setupUploadTestEnv(t *testing.T) (*UploadHandler, *gin.Engine, uuid.UUID, func()) {
	// 初始化数据库（使用与 services 层相同的测试辅助函数）
	db := setupTestDB(t)

	// 初始化存储引擎（内存）
	storageEngine := storage.NewMemoryEngine()

	// 初始化 KEK
	kek := []byte("test-master-key-1234567890123456")

	// 初始化服务
	fileService := services.NewFileService(db, storageEngine, kek)

	// 创建处理器
	handler := NewUploadHandler(fileService)

	// 创建测试用户
	user := &models.User{
		Email:        "upload@test.com",
		Password:     "hashed_password",
		StorageQuota: 10 * 1024 * 1024 * 1024, // 10GB
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(user).Error)

	// 设置 Gin 路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 模拟认证中间件
	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID.String())
		c.Next()
	})

	router.POST("/api/tus/upload", handler.CreateUpload)
	router.PATCH("/api/tus/upload/:id", handler.UploadChunk)
	router.HEAD("/api/tus/upload/:id", handler.GetUploadProgress)
	router.DELETE("/api/tus/upload/:id", handler.DeleteUpload)
	router.OPTIONS("/api/tus/upload", handler.Options)

	// 清理函数
	cleanup := func() {
		// SQLite 内存数据库会自动清理
	}

	return handler, router, user.ID, cleanup
}

// TestCreateUpload 测试创建上传会话
func TestCreateUpload(t *testing.T) {
	_, router, _, cleanup := setupUploadTestEnv(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    CreateUploadRequest
		expectedStatus int
		expectedCode   int
		checkResponse  func(t *testing.T, resp map[string]interface{})
	}{
		{
			name: "valid upload session creation",
			requestBody: CreateUploadRequest{
				Filename: "test.txt",
				Size:     1024,
				Hash:     "a" + string(make([]byte, 63)), // 模拟 SHA-256
			},
			expectedStatus: http.StatusCreated,
			expectedCode:   0,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				assert.NotEmpty(t, data["upload_id"])
				assert.NotEmpty(t, data["upload_url"])
				assert.Equal(t, float64(0), data["upload_offset"])
				assert.Equal(t, float64(1024), data["upload_length"])
			},
		},
		{
			name: "missing filename",
			requestBody: CreateUploadRequest{
				Size: 1024,
				Hash: "a" + string(make([]byte, 63)),
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   400,
		},
		{
			name: "invalid size",
			requestBody: CreateUploadRequest{
				Filename: "test.txt",
				Size:     0,
				Hash:     "a" + string(make([]byte, 63)),
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   400,
		},
		{
			name: "missing hash",
			requestBody: CreateUploadRequest{
				Filename: "test.txt",
				Size:     1024,
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建请求
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/api/tus/upload", bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// 发送请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, float64(tt.expectedCode), resp["code"])

			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestUploadChunk 测试分片上传
func TestUploadChunk(t *testing.T) {
	_, router, _, cleanup := setupUploadTestEnv(t)
	defer cleanup()

	uploadID := uuid.New().String()

	tests := []struct {
		name           string
		uploadID       string
		headers        map[string]string
		body           []byte
		expectedStatus int
	}{
		{
			name:     "valid chunk upload",
			uploadID: uploadID,
			headers: map[string]string{
				"Content-Type":   "application/offset+octet-stream",
				"Upload-Offset":  "0",
				"Content-Length": "13",
				"Upload-Metadata-Filename": "test.txt",
				"Upload-Metadata-Size":     "13",
			},
			body:           []byte("Hello, World!"),
			expectedStatus: http.StatusOK,
		},
		{
			name:     "missing upload offset",
			uploadID: uploadID,
			headers: map[string]string{
				"Content-Type":   "application/offset+octet-stream",
				"Content-Length": "13",
			},
			body:           []byte("Hello, World!"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPatch, "/api/tus/upload/"+tt.uploadID, bytes.NewReader(tt.body))
			require.NoError(t, err)

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetUploadProgress 测试查询上传进度
func TestGetUploadProgress(t *testing.T) {
	_, router, _, cleanup := setupUploadTestEnv(t)
	defer cleanup()

	uploadID := uuid.New().String()

	req, err := http.NewRequest(http.MethodHead, "/api/tus/upload/"+uploadID, nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1.0.0", w.Header().Get("Tus-Resumable"))
	assert.NotEmpty(t, w.Header().Get("Upload-Offset"))
	assert.NotEmpty(t, w.Header().Get("Upload-Length"))
}

// TestDeleteUpload 测试删除上传会话
func TestDeleteUpload(t *testing.T) {
	_, router, _, cleanup := setupUploadTestEnv(t)
	defer cleanup()

	uploadID := uuid.New().String()

	req, err := http.NewRequest(http.MethodDelete, "/api/tus/upload/"+uploadID, nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "1.0.0", w.Header().Get("Tus-Resumable"))
}

// TestOptions 测试 OPTIONS 请求
func TestOptions(t *testing.T) {
	_, router, _, cleanup := setupUploadTestEnv(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodOptions, "/api/tus/upload", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "1.0.0", w.Header().Get("Tus-Resumable"))
	assert.Equal(t, "1.0.0", w.Header().Get("Tus-Version"))
	assert.NotEmpty(t, w.Header().Get("Tus-Extension"))
	assert.NotEmpty(t, w.Header().Get("Tus-Max-Size"))
}
