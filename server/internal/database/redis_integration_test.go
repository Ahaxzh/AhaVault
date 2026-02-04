// Package database 提供数据库连接管理
//
// 本文件实现 Redis 的集成测试（需要真实 Redis）
//
// 作者: AhaVault Team
// 创建时间: 2026-02-05
package database

import (
	"context"
	"os"
	"testing"
	"time"

	"ahavault/server/internal/config"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisIntegration Redis 集成测试
//
// 注意：此测试需要真实的 Redis 实例运行
// 运行前请启动: docker-compose up -d
func TestRedisIntegration(t *testing.T) {
	// 检查是否跳过集成测试
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 配置测试环境变量
	setupRedisTestEnv(t)

	// 加载配置
	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	// 初始化 Redis
	err = InitRedis(cfg)
	require.NoError(t, err, "Failed to initialize Redis")
	defer CloseRedis()

	// 验证全局实例已设置
	assert.NotNil(t, RedisClient, "RedisClient global instance should be set")
	assert.NotNil(t, GetRedis(), "GetRedis() should return non-nil")

	t.Run("健康检查", testRedisHealthCheck)
	t.Run("连接池统计", testRedisPoolStats)
	t.Run("基本键值操作", testRedisBasicOps)
	t.Run("过期时间操作", testRedisExpiration)
	t.Run("哈希操作", testRedisHashOps)
	t.Run("分布式锁", testRedisLock)
}

// testRedisHealthCheck 测试 Redis 健康检查
func testRedisHealthCheck(t *testing.T) {
	ctx := context.Background()

	err := RedisHealthCheck(ctx)
	assert.NoError(t, err, "Health check should succeed")

	// 测试超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond) // 确保超时

	err = RedisHealthCheck(ctx)
	// 超时可能会失败，但不应该 panic
	if err != nil {
		t.Logf("Health check with timeout context failed as expected: %v", err)
	}
}

// testRedisPoolStats 测试连接池统计
func testRedisPoolStats(t *testing.T) {
	stats, err := GetRedisStats()
	require.NoError(t, err, "Failed to get Redis stats")
	require.NotNil(t, stats, "Stats should not be nil")

	// 验证统计信息的合理性
	assert.GreaterOrEqual(t, stats.TotalConns, uint32(0), "TotalConns should be >= 0")
	assert.GreaterOrEqual(t, stats.IdleConns, uint32(0), "IdleConns should be >= 0")

	t.Logf("Redis Pool Stats: Total=%d, Idle=%d, Hits=%d, Misses=%d",
		stats.TotalConns, stats.IdleConns, stats.Hits, stats.Misses)
}

// testRedisBasicOps 测试基本键值操作
func testRedisBasicOps(t *testing.T) {
	ctx := context.Background()
	testKey := "test:integration:" + time.Now().Format("20060102150405")
	testValue := "test_value_123"

	// 测试 Set
	err := Set(ctx, testKey, testValue, 5*time.Minute)
	require.NoError(t, err, "Set should succeed")

	// 测试 Get
	value, err := Get(ctx, testKey)
	require.NoError(t, err, "Get should succeed")
	assert.Equal(t, testValue, value, "Value should match")

	// 测试 Exists
	count, err := Exists(ctx, testKey)
	require.NoError(t, err, "Exists should succeed")
	assert.Equal(t, int64(1), count, "Key should exist")

	// 测试 Del
	err = Del(ctx, testKey)
	require.NoError(t, err, "Del should succeed")

	// 验证删除
	_, err = Get(ctx, testKey)
	assert.Equal(t, redis.Nil, err, "Should return redis.Nil for deleted key")
}

// testRedisExpiration 测试过期时间操作
func testRedisExpiration(t *testing.T) {
	ctx := context.Background()
	testKey := "test:expiration:" + time.Now().Format("20060102150405")

	// 设置键值
	err := Set(ctx, testKey, "value", 1*time.Second)
	require.NoError(t, err, "Set should succeed")

	// 获取 TTL
	ttl, err := TTL(ctx, testKey)
	require.NoError(t, err, "TTL should succeed")
	assert.Greater(t, ttl, time.Duration(0), "TTL should be > 0")
	assert.LessOrEqual(t, ttl, 1*time.Second, "TTL should be <= 1s")

	// 更新过期时间
	err = Expire(ctx, testKey, 5*time.Second)
	require.NoError(t, err, "Expire should succeed")

	// 验证新的 TTL
	ttl, err = TTL(ctx, testKey)
	require.NoError(t, err, "TTL should succeed")
	assert.Greater(t, ttl, 1*time.Second, "TTL should be > 1s after update")

	// 清理
	Del(ctx, testKey)
}

// testRedisHashOps 测试哈希操作
func testRedisHashOps(t *testing.T) {
	ctx := context.Background()
	hashKey := "test:hash:" + time.Now().Format("20060102150405")

	// 测试 HSet
	err := HSet(ctx, hashKey, "field1", "value1", "field2", "value2")
	require.NoError(t, err, "HSet should succeed")

	// 测试 HGet
	value, err := HGet(ctx, hashKey, "field1")
	require.NoError(t, err, "HGet should succeed")
	assert.Equal(t, "value1", value, "Value should match")

	// 测试 HGetAll
	allValues, err := HGetAll(ctx, hashKey)
	require.NoError(t, err, "HGetAll should succeed")
	assert.Len(t, allValues, 2, "Should have 2 fields")
	assert.Equal(t, "value1", allValues["field1"], "field1 should match")
	assert.Equal(t, "value2", allValues["field2"], "field2 should match")

	// 清理
	Del(ctx, hashKey)
}

// testRedisLock 测试分布式锁（SetNX）
func testRedisLock(t *testing.T) {
	ctx := context.Background()
	lockKey := "test:lock:" + time.Now().Format("20060102150405")

	// 第一次获取锁应该成功
	acquired, err := SetNX(ctx, lockKey, "owner1", 5*time.Second)
	require.NoError(t, err, "SetNX should succeed")
	assert.True(t, acquired, "First lock acquisition should succeed")

	// 第二次获取锁应该失败（键已存在）
	acquired, err = SetNX(ctx, lockKey, "owner2", 5*time.Second)
	require.NoError(t, err, "SetNX should not error")
	assert.False(t, acquired, "Second lock acquisition should fail")

	// 释放锁
	err = Del(ctx, lockKey)
	require.NoError(t, err, "Del should succeed")

	// 再次获取锁应该成功
	acquired, err = SetNX(ctx, lockKey, "owner3", 5*time.Second)
	require.NoError(t, err, "SetNX should succeed")
	assert.True(t, acquired, "Lock acquisition after delete should succeed")

	// 清理
	Del(ctx, lockKey)
}

// testRedisIncr 测试自增操作
func TestRedisIncr(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupRedisTestEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	err = InitRedis(cfg)
	require.NoError(t, err)
	defer CloseRedis()

	ctx := context.Background()
	counterKey := "test:counter:" + time.Now().Format("20060102150405")

	// 第一次自增
	count, err := Incr(ctx, counterKey)
	require.NoError(t, err, "Incr should succeed")
	assert.Equal(t, int64(1), count, "First incr should return 1")

	// 第二次自增
	count, err = Incr(ctx, counterKey)
	require.NoError(t, err, "Incr should succeed")
	assert.Equal(t, int64(2), count, "Second incr should return 2")

	// 第三次自增
	count, err = Incr(ctx, counterKey)
	require.NoError(t, err, "Incr should succeed")
	assert.Equal(t, int64(3), count, "Third incr should return 3")

	// 清理
	Del(ctx, counterKey)
}

// TestInitRedisError 测试初始化错误场景
func TestInitRedisError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 测试无效配置
	invalidCfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     "invalid-host-12345",
			Port:     9999,
			Password: "invalid",
			DB:       0,
			Timeout:  1 * time.Second,
		},
	}

	err := InitRedis(invalidCfg)
	assert.Error(t, err, "Should fail with invalid config")
	assert.Contains(t, err.Error(), "failed to connect to redis", "Error should mention connection failure")
}

// TestCloseRedis 测试关闭 Redis 连接
func TestCloseRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupRedisTestEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	// 初始化
	err = InitRedis(cfg)
	require.NoError(t, err)

	// 关闭
	err = CloseRedis()
	assert.NoError(t, err, "CloseRedis should succeed")

	// 再次关闭应该安全
	err = CloseRedis()
	assert.NoError(t, err, "CloseRedis should be idempotent")
}

// setupRedisTestEnv 设置 Redis 测试环境变量
func setupRedisTestEnv(t *testing.T) {
	// 设置 Redis 连接环境变量
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "ahavault_dev")
	os.Setenv("REDIS_DB", "0")

	// 设置必需的环境变量
	os.Setenv("POSTGRES_PASSWORD", "ahavault_dev") // config.Load() 验证需要
	os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	t.Cleanup(func() {
		// 测试结束后清理
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("APP_MASTER_KEY")
	})
}
