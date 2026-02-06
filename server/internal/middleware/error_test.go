// Package middleware 提供 HTTP 中间件测试
//
// 本文件测试 ErrorHandler 全局错误处理中间件的功能：
//   - Panic 恢复
//   - 错误响应格式化
//   - Gin 错误链处理
//
// 作者: AhaVault Team
// 创建时间: 2026-02-06
package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		handler        gin.HandlerFunc
		expectedStatus int
		checkBody      func(t *testing.T, body string)
	}{
		{
			name: "should pass through normal request",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "success")
			},
		},
		{
			name: "should recover from panic",
			handler: func(c *gin.Context) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Internal server error")
			},
		},
		{
			name: "should handle gin errors",
			handler: func(c *gin.Context) {
				_ = c.Error(errors.New("custom error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "custom error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ErrorHandler())
			router.GET("/test", tt.handler)

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkBody(t, w.Body.String())
		})
	}
}
