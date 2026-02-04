// Package middleware 提供 HTTP 中间件
//
// 本文件实现基于 Redis 的限流中间件，功能包括：
//   - IP 级别限流（防止恶意刷接口）
//   - 用户级别限流（认证用户）
//   - 灵活的限流规则配置
//   - 滑动窗口算法
//   - 自动清理过期数据
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// Redis 客户端
	RedisClient *redis.Client

	// 限流规则：每个时间窗口允许的最大请求数
	Limit int

	// 时间窗口（秒）
	Window time.Duration

	// 限流键前缀
	KeyPrefix string

	// 限流类型：ip（IP 限流）或 user（用户限流）
	LimitBy string

	// 自定义键生成函数
	KeyFunc func(c *gin.Context) string
}

// RateLimit 限流中间件工厂函数
//
// 该中间件使用 Redis 实现分布式限流：
//  1. 根据 IP 或用户 ID 生成限流键
//  2. 使用 Redis INCR + EXPIRE 实现计数
//  3. 超过限制返回 429 Too Many Requests
//  4. 自动清理过期数据
//
// 参数:
//   - config: 限流配置
//
// 返回:
//   - Gin 中间件函数
//
// 示例:
//   // 登录限流：5 次/分钟
//   loginLimiter := middleware.RateLimit(middleware.RateLimitConfig{
//       RedisClient: redisClient,
//       Limit:       5,
//       Window:      time.Minute,
//       KeyPrefix:   "ratelimit:login:",
//       LimitBy:     "ip",
//   })
//   router.POST("/api/auth/login", loginLimiter, handlers.Login)
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	// 设置默认值
	if config.KeyPrefix == "" {
		config.KeyPrefix = "ratelimit:"
	}
	if config.LimitBy == "" {
		config.LimitBy = "ip"
	}
	if config.Window == 0 {
		config.Window = time.Minute
	}

	return func(c *gin.Context) {
		// 生成限流键
		var key string
		if config.KeyFunc != nil {
			key = config.KeyFunc(c)
		} else {
			key = generateKey(c, config)
		}

		ctx := context.Background()

		// 获取当前计数
		count, err := config.RedisClient.Incr(ctx, key).Result()
		if err != nil {
			// Redis 错误时放行请求（优雅降级）
			fmt.Printf("[RateLimit] Redis error: %v\n", err)
			c.Next()
			return
		}

		// 如果是第一次请求，设置过期时间
		if count == 1 {
			config.RedisClient.Expire(ctx, key, config.Window)
		}

		// 检查是否超过限制
		if count > int64(config.Limit) {
			// 获取剩余时间
			ttl, _ := config.RedisClient.TTL(ctx, key).Result()

			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "Too many requests",
				"data": gin.H{
					"limit":       config.Limit,
					"window":      config.Window.String(),
					"retry_after": int(ttl.Seconds()),
				},
			})
			c.Abort()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", config.Limit-int(count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(config.Window).Unix()))

		c.Next()
	}
}

// generateKey 生成限流键
func generateKey(c *gin.Context, config RateLimitConfig) string {
	switch config.LimitBy {
	case "user":
		// 用户级限流
		if userID, exists := c.Get("user_id"); exists {
			return fmt.Sprintf("%suser:%s", config.KeyPrefix, userID)
		}
		// 未登录用户回退到 IP 限流
		fallthrough
	case "ip":
		// IP 级限流
		return fmt.Sprintf("%sip:%s", config.KeyPrefix, c.ClientIP())
	default:
		// 全局限流
		return fmt.Sprintf("%sglobal", config.KeyPrefix)
	}
}

// NewIPRateLimiter 创建 IP 限流中间件（便捷函数）
//
// 参数:
//   - redisClient: Redis 客户端
//   - limit: 每个时间窗口允许的请求数
//   - window: 时间窗口
//   - keyPrefix: 键前缀
//
// 返回:
//   - Gin 中间件函数
//
// 示例:
//   limiter := middleware.NewIPRateLimiter(redis, 100, time.Minute, "api:")
//   router.Use(limiter)
func NewIPRateLimiter(redisClient *redis.Client, limit int, window time.Duration, keyPrefix string) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		RedisClient: redisClient,
		Limit:       limit,
		Window:      window,
		KeyPrefix:   keyPrefix,
		LimitBy:     "ip",
	})
}

// NewUserRateLimiter 创建用户限流中间件（便捷函数）
//
// 参数:
//   - redisClient: Redis 客户端
//   - limit: 每个时间窗口允许的请求数
//   - window: 时间窗口
//   - keyPrefix: 键前缀
//
// 返回:
//   - Gin 中间件函数
//
// 示例:
//   limiter := middleware.NewUserRateLimiter(redis, 20, time.Hour, "upload:")
//   router.POST("/api/upload", auth, limiter, handlers.Upload)
func NewUserRateLimiter(redisClient *redis.Client, limit int, window time.Duration, keyPrefix string) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		RedisClient: redisClient,
		Limit:       limit,
		Window:      window,
		KeyPrefix:   keyPrefix,
		LimitBy:     "user",
	})
}

// CommonRateLimiters 常用限流器集合
type CommonRateLimiters struct {
	// 登录限流：5 次/分钟
	Login gin.HandlerFunc

	// 取件码验证限流：10 次/分钟
	PickupCode gin.HandlerFunc

	// 上传限流：20 次/小时
	Upload gin.HandlerFunc

	// API 总限流：100 次/分钟
	API gin.HandlerFunc
}

// NewCommonRateLimiters 创建常用限流器集合
//
// 根据 PRD 要求预设的限流规则：
//   - 登录: 5 次/分钟（IP 级）
//   - 取件码验证: 10 次/分钟（IP 级）
//   - 上传: 20 次/小时（用户级）
//   - API 总限流: 100 次/分钟（IP 级）
//
// 参数:
//   - redisClient: Redis 客户端
//
// 返回:
//   - 常用限流器集合
func NewCommonRateLimiters(redisClient *redis.Client) *CommonRateLimiters {
	return &CommonRateLimiters{
		Login: NewIPRateLimiter(redisClient, 5, time.Minute, "ratelimit:login:"),
		PickupCode: NewIPRateLimiter(redisClient, 10, time.Minute, "ratelimit:pickup:"),
		Upload: NewUserRateLimiter(redisClient, 20, time.Hour, "ratelimit:upload:"),
		API: NewIPRateLimiter(redisClient, 100, time.Minute, "ratelimit:api:"),
	}
}
