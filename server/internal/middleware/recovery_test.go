package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRecovery 测试 Recovery 中间件
func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should recover from panic", func(t *testing.T) {
		r := gin.New()
		r.Use(Recovery())
		r.GET("/panic", func(c *gin.Context) {
			panic("something went wrong")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, float64(500), resp["code"])
		assert.Equal(t, "Internal server error", resp["message"])
		assert.Contains(t, resp["error"], "something went wrong")
	})

	t.Run("should not affect normal requests", func(t *testing.T) {
		r := gin.New()
		r.Use(Recovery())
		r.GET("/ok", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ok", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})
}

// TestRecoveryWithWriter 测试带自定义日志功能的 Recovery 中间件
func TestRecoveryWithWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should call custom log function", func(t *testing.T) {
		var capturedErr interface{}
		var capturedStack []byte

		logFunc := func(err interface{}, stack []byte) {
			capturedErr = err
			capturedStack = stack
		}

		r := gin.New()
		r.Use(RecoveryWithWriter(logFunc))
		r.GET("/panic", func(c *gin.Context) {
			panic("custom panic")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "custom panic", capturedErr)
		assert.NotEmpty(t, capturedStack)
	})
}
