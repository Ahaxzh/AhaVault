package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"ahavault/server/internal/config"
)

// RedisClient 全局 Redis 客户端
var RedisClient *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *config.Config) error {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  cfg.Redis.Timeout,
		ReadTimeout:  cfg.Redis.Timeout,
		WriteTimeout: cfg.Redis.Timeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	// 设置全局实例
	RedisClient = client

	return nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if RedisClient == nil {
		return nil
	}

	return RedisClient.Close()
}

// RedisHealthCheck Redis 健康检查
func RedisHealthCheck(ctx context.Context) error {
	if RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}

	return RedisClient.Ping(ctx).Err()
}

// GetRedis 获取 Redis 客户端
func GetRedis() *redis.Client {
	return RedisClient
}

// RedisStats Redis 连接池统计信息
type RedisStats struct {
	Hits     uint32 // 命中次数
	Misses   uint32 // 未命中次数
	Timeouts uint32 // 超时次数
	TotalConns uint32 // 总连接数
	IdleConns  uint32 // 空闲连接数
	StaleConns uint32 // 过期连接数
}

// GetRedisStats 获取 Redis 统计信息
func GetRedisStats() (*RedisStats, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	poolStats := RedisClient.PoolStats()

	return &RedisStats{
		Hits:       poolStats.Hits,
		Misses:     poolStats.Misses,
		Timeouts:   poolStats.Timeouts,
		TotalConns: poolStats.TotalConns,
		IdleConns:  poolStats.IdleConns,
		StaleConns: poolStats.StaleConns,
	}, nil
}

// ==========================================
// Redis 常用操作封装
// ==========================================

// Set 设置键值（带过期时间）
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}

	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// Get 获取键值
func Get(ctx context.Context, key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("redis not initialized")
	}

	return RedisClient.Get(ctx, key).Result()
}

// Del 删除键
func Del(ctx context.Context, keys ...string) error {
	if RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}

	return RedisClient.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, keys ...string) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("redis not initialized")
	}

	return RedisClient.Exists(ctx, keys...).Result()
}

// Incr 自增
func Incr(ctx context.Context, key string) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("redis not initialized")
	}

	return RedisClient.Incr(ctx, key).Result()
}

// Expire 设置过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}

	return RedisClient.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func TTL(ctx context.Context, key string) (time.Duration, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("redis not initialized")
	}

	return RedisClient.TTL(ctx, key).Result()
}

// SetNX 仅当键不存在时设置（分布式锁）
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if RedisClient == nil {
		return false, fmt.Errorf("redis not initialized")
	}

	return RedisClient.SetNX(ctx, key, value, expiration).Result()
}

// HSet 设置哈希字段
func HSet(ctx context.Context, key string, values ...interface{}) error {
	if RedisClient == nil {
		return fmt.Errorf("redis not initialized")
	}

	return RedisClient.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段
func HGet(ctx context.Context, key, field string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("redis not initialized")
	}

	return RedisClient.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	return RedisClient.HGetAll(ctx, key).Result()
}
