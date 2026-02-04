package middleware

import (
	"ahavault/server/internal/config"
	"ahavault/server/internal/database"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRedisTestEnv 设置 Redis 测试环境
func setupRedisTestEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// 设置 Redis 连接环境变量
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "ahavault_dev")
	os.Setenv("REDIS_DB", "0")

	// 设置必需的环境变量
	os.Setenv("POSTGRES_PASSWORD", "ahavault_dev")
	os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	t.Cleanup(func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("APP_MASTER_KEY")
	})
}

// TestRateLimit 测试 RateLimit 中间件
func TestRateLimit(t *testing.T) {
	setupRedisTestEnv(t)

	// 加载配置并初始化 Redis
	cfg, err := config.Load()
	require.NoError(t, err)
	err = database.InitRedis(cfg)
	require.NoError(t, err)
	defer database.CloseRedis()

	redisClient := database.GetRedis()
	require.NotNil(t, redisClient)

	gin.SetMode(gin.TestMode)

	t.Run("should limit requests", func(t *testing.T) {
		// 配置限流：允许 2 次/秒
		limit := 2
		window := time.Second
		keyPrefix := "test:ratelimit:ip:" + time.Now().Format("150405")

		limiter := NewIPRateLimiter(redisClient, limit, window, keyPrefix)

		r := gin.New()
		r.Use(limiter)
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		// 第一次请求：允许
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Equal(t, "2", w1.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "1", w1.Header().Get("X-RateLimit-Remaining"))

		// 第二次请求：允许
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Equal(t, "2", w2.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", w2.Header().Get("X-RateLimit-Remaining"))

		// 第三次请求：拒绝 (429)
		w3 := httptest.NewRecorder()
		req3, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusTooManyRequests, w3.Code)

		var resp map[string]interface{}
		json.Unmarshal(w3.Body.Bytes(), &resp)
		assert.Equal(t, float64(429), resp["code"])

		// 清理 Redis 键
		redisClient.Del(redisClient.Context(), keyPrefix+"ip:")
	})

	t.Run("should limit by user", func(t *testing.T) {
		// 配置限流：允许 1 次/秒，按用户
		limit := 1
		window := time.Second
		keyPrefix := "test:ratelimit:user:" + time.Now().Format("150405")

		limiter := NewUserRateLimiter(redisClient, limit, window, keyPrefix)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "user-123") // 模拟认证
			c.Next()
		})
		r.Use(limiter)
		r.GET("/test-user", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		// 第一次请求：允许
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test-user", nil)
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// 第二次请求：拒绝
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test-user", nil)
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)

		// 清理
		redisClient.Del(redisClient.Context(), keyPrefix+"user:user-123")
	})
}
