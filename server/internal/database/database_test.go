package database

import (
	"context"
	"os"
	"testing"
	"time"

	"ahavault/server/internal/config"
)

func TestInitPostgreSQL(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "ahavault")
	os.Setenv("POSTGRES_PASSWORD", "ahavault_dev_2026")
	os.Setenv("POSTGRES_DB", "ahavault")
	os.Setenv("POSTGRES_SSLMODE", "disable")

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Failed to load config (skipping test): %v", err)
	}

	// 测试初始化
	err = InitPostgreSQL(cfg)
	if err != nil {
		t.Skipf("Failed to connect to PostgreSQL (skipping test): %v", err)
	}
	defer Close()

	// 测试健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// 测试获取统计信息
	stats, err := GetStats()
	if err != nil {
		t.Errorf("GetStats failed: %v", err)
	}

	if stats.MaxOpenConnections == 0 {
		t.Error("MaxOpenConnections should not be 0")
	}
}

func TestGetDB(t *testing.T) {
	// 未初始化时应该返回 nil
	DB = nil
	if GetDB() != nil {
		t.Error("GetDB() should return nil when not initialized")
	}
}

func TestInitRedis(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	os.Setenv("POSTGRES_PASSWORD", "test")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "redis_dev_2026")

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Failed to load config (skipping test): %v", err)
	}

	// 测试初始化
	err = InitRedis(cfg)
	if err != nil {
		t.Skipf("Failed to connect to Redis (skipping test): %v", err)
	}
	defer CloseRedis()

	// 测试健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisHealthCheck(ctx); err != nil {
		t.Errorf("RedisHealthCheck failed: %v", err)
	}

	// 测试统计信息
	stats, err := GetRedisStats()
	if err != nil {
		t.Errorf("GetRedisStats failed: %v", err)
	}

	if stats == nil {
		t.Error("RedisStats should not be nil")
	}
}

func TestRedisOperations(t *testing.T) {
	// 设置测试环境
	os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	os.Setenv("POSTGRES_PASSWORD", "test")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "redis_dev_2026")

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Failed to load config (skipping test): %v", err)
	}

	if err := InitRedis(cfg); err != nil {
		t.Skipf("Failed to connect to Redis (skipping test): %v", err)
	}
	defer CloseRedis()

	ctx := context.Background()

	// 测试 Set/Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		if err := Set(ctx, key, value, 10*time.Second); err != nil {
			t.Errorf("Set failed: %v", err)
		}

		got, err := Get(ctx, key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}

		if got != value {
			t.Errorf("Get() = %s, want %s", got, value)
		}

		// 清理
		Del(ctx, key)
	})

	// 测试 Exists
	t.Run("Exists", func(t *testing.T) {
		key := "test_exists"
		Set(ctx, key, "value", 10*time.Second)

		count, err := Exists(ctx, key)
		if err != nil {
			t.Errorf("Exists failed: %v", err)
		}

		if count != 1 {
			t.Errorf("Exists() = %d, want 1", count)
		}

		Del(ctx, key)
	})

	// 测试 Incr
	t.Run("Incr", func(t *testing.T) {
		key := "test_counter"

		val, err := Incr(ctx, key)
		if err != nil {
			t.Errorf("Incr failed: %v", err)
		}

		if val != 1 {
			t.Errorf("Incr() = %d, want 1", val)
		}

		Del(ctx, key)
	})

	// 测试 SetNX (分布式锁)
	t.Run("SetNX", func(t *testing.T) {
		key := "test_lock"

		// 第一次应该成功
		ok, err := SetNX(ctx, key, "locked", 10*time.Second)
		if err != nil {
			t.Errorf("SetNX failed: %v", err)
		}
		if !ok {
			t.Error("SetNX should return true for first call")
		}

		// 第二次应该失败（键已存在）
		ok, err = SetNX(ctx, key, "locked", 10*time.Second)
		if err != nil {
			t.Errorf("SetNX failed: %v", err)
		}
		if ok {
			t.Error("SetNX should return false for second call")
		}

		Del(ctx, key)
	})

	// 测试 Hash 操作
	t.Run("Hash Operations", func(t *testing.T) {
		key := "test_hash"

		// HSet
		if err := HSet(ctx, key, "field1", "value1", "field2", "value2"); err != nil {
			t.Errorf("HSet failed: %v", err)
		}

		// HGet
		val, err := HGet(ctx, key, "field1")
		if err != nil {
			t.Errorf("HGet failed: %v", err)
		}
		if val != "value1" {
			t.Errorf("HGet() = %s, want value1", val)
		}

		// HGetAll
		all, err := HGetAll(ctx, key)
		if err != nil {
			t.Errorf("HGetAll failed: %v", err)
		}
		if len(all) != 2 {
			t.Errorf("HGetAll() returned %d fields, want 2", len(all))
		}

		Del(ctx, key)
	})
}
