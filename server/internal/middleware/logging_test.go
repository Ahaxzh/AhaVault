package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestLogger 测试 Logger 中间件
func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 由于 Logger 默认直接打印到 stdout，这里主要测试不会 panic 且正确调用 next
	t.Run("should process request correctly", func(t *testing.T) {
		r := gin.New()
		r.Use(Logger())
		r.GET("/test-log", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test-log", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestLoggerWithConfig 测试带配置的 Logger 中间件
func TestLoggerWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should skip configured paths", func(t *testing.T) {
		var logCalled bool
		config := LoggerConfig{
			SkipPaths: []string{"/health"},
			LogWriter: func(entry LogEntry) {
				logCalled = true
			},
		}

		r := gin.New()
		r.Use(LoggerWithConfig(config))
		r.GET("/health", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, logCalled, "should skip logging for /health")
	})

	t.Run("should log non-skipped paths", func(t *testing.T) {
		var capturedEntry LogEntry
		config := LoggerConfig{
			SkipPaths: []string{"/health"},
			LogWriter: func(entry LogEntry) {
				capturedEntry = entry
			},
			EnableQueryParams: true,
			EnableUserAgent:   true,
		}

		r := gin.New()
		r.Use(LoggerWithConfig(config))
		r.GET("/api/data", func(c *gin.Context) {
			c.Set("user_id", "user-123")
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/data?q=test", nil)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/api/data", capturedEntry.Path)
		assert.Equal(t, "GET", capturedEntry.Method)
		assert.Equal(t, "q=test", capturedEntry.Query)
		assert.Equal(t, "TestAgent/1.0", capturedEntry.UserAgent)
		assert.Equal(t, "user-123", capturedEntry.UserID)
		assert.Equal(t, 200, capturedEntry.StatusCode)
	})
}
