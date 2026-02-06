// Package middleware 提供 HTTP 中间件测试
//
// 本文件测试 CORS 跨域中间件的功能：
//   - CORS 响应头设置
//   - OPTIONS 预检请求处理
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

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "should set CORS headers for GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
			},
		},
		{
			name:           "should handle OPTIONS preflight request",
			method:         http.MethodOptions,
			expectedStatus: http.StatusNoContent,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
			},
		},
		{
			name:           "should set CORS headers for POST request",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORS())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			router.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			router.OPTIONS("/test", func(c *gin.Context) {
				// This should not be reached because CORS middleware handles it
				c.JSON(http.StatusOK, gin.H{"message": "should not reach"})
			})

			req, _ := http.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkHeaders(t, w)
		})
	}
}
