// Package middleware 提供 HTTP 中间件测试
//
// 本文件测试 Auth 认证中间件的功能：
//   - JWT Token 验证
//   - Bearer 格式解析
//   - 用户 ID 上下文存储
//
// 作者: AhaVault Team
// 创建时间: 2026-02-06
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected string
	}{
		{
			name: "should return user ID when set",
			setup: func(c *gin.Context) {
				c.Set("user_id", "test-user-123")
			},
			expected: "test-user-123",
		},
		{
			name: "should return empty string when not set",
			setup: func(c *gin.Context) {
				// No setup, user_id not set
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := GetUserID(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	// Use a mock user service that returns nil - we're testing header parsing
	router.Use(Auth(nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "missing Bearer prefix",
			authHeader: "token-without-bearer",
		},
		{
			name:       "wrong prefix",
			authHeader: "Basic sometoken",
		},
		{
			name:       "only Bearer without token",
			authHeader: "Bearer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Auth(nil))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "Invalid authorization header format")
		})
	}
}
