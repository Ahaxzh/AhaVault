// Package database 提供数据库连接管理
//
// 本文件实现 PostgreSQL 的集成测试（需要真实数据库）
//
// 作者: AhaVault Team
// 创建时间: 2026-02-05
package database

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"ahavault/server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLIntegration PostgreSQL 集成测试
//
// 注意：此测试需要真实的 PostgreSQL 实例运行
// 运行前请启动: docker-compose up -d
func TestPostgreSQLIntegration(t *testing.T) {
	// 检查是否跳过集成测试
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 配置测试环境变量
	setupTestEnv(t)

	// 加载配置
	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	// 初始化数据库
	err = InitPostgreSQL(cfg)
	require.NoError(t, err, "Failed to initialize PostgreSQL")
	defer Close()

	// 验证全局实例已设置
	assert.NotNil(t, DB, "DB global instance should be set")
	assert.NotNil(t, GetDB(), "GetDB() should return non-nil")

	t.Run("健康检查", testHealthCheck)
	t.Run("连接池统计", testConnectionPoolStats)
	t.Run("数据库操作", testDatabaseOperations)
}

// testHealthCheck 测试健康检查功能
func testHealthCheck(t *testing.T) {
	ctx := context.Background()

	err := HealthCheck(ctx)
	assert.NoError(t, err, "Health check should succeed")

	// 测试超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond) // 确保超时

	err = HealthCheck(ctx)
	// 超时可能会失败，但不应该 panic
	if err != nil {
		t.Logf("Health check with timeout context failed as expected: %v", err)
	}
}

// testConnectionPoolStats 测试连接池统计
func testConnectionPoolStats(t *testing.T) {
	stats, err := GetStats()
	require.NoError(t, err, "Failed to get stats")
	require.NotNil(t, stats, "Stats should not be nil")

	// 验证统计信息的合理性
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 1, "MaxOpenConnections should be >= 1")
	assert.GreaterOrEqual(t, stats.OpenConnections, 0, "OpenConnections should be >= 0")
	assert.GreaterOrEqual(t, stats.Idle, 0, "Idle connections should be >= 0")
	assert.GreaterOrEqual(t, stats.InUse, 0, "InUse connections should be >= 0")

	t.Logf("Connection Pool Stats: Max=%d, Open=%d, InUse=%d, Idle=%d",
		stats.MaxOpenConnections, stats.OpenConnections, stats.InUse, stats.Idle)
}

// testDatabaseOperations 测试基本数据库操作
func testDatabaseOperations(t *testing.T) {
	// 测试创建表
	type TestTable struct {
		ID        uint      `gorm:"primaryKey"`
		Name      string    `gorm:"size:100"`
		CreatedAt time.Time `gorm:"autoCreateTime"`
	}

	err := DB.AutoMigrate(&TestTable{})
	require.NoError(t, err, "Failed to auto migrate test table")

	// 测试插入数据
	testData := TestTable{
		Name: "integration_test_" + time.Now().Format("20060102150405"),
	}

	result := DB.Create(&testData)
	require.NoError(t, result.Error, "Failed to insert test data")
	assert.NotZero(t, testData.ID, "ID should be auto-generated")

	// 测试查询数据
	var found TestTable
	result = DB.First(&found, testData.ID)
	require.NoError(t, result.Error, "Failed to query test data")
	assert.Equal(t, testData.Name, found.Name, "Name should match")

	// 测试更新数据
	found.Name = "updated_" + found.Name
	result = DB.Save(&found)
	require.NoError(t, result.Error, "Failed to update test data")

	// 验证更新
	var updated TestTable
	result = DB.First(&updated, testData.ID)
	require.NoError(t, result.Error, "Failed to query updated data")
	assert.Equal(t, found.Name, updated.Name, "Updated name should match")

	// 测试删除数据
	result = DB.Delete(&updated)
	require.NoError(t, result.Error, "Failed to delete test data")

	// 验证删除
	result = DB.First(&TestTable{}, testData.ID)
	assert.Error(t, result.Error, "Should return error for deleted record")

	// 清理：删除测试表
	err = DB.Migrator().DropTable(&TestTable{})
	require.NoError(t, err, "Failed to drop test table")
}

// TestInitPostgreSQLError 测试初始化错误场景
func TestInitPostgreSQLError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 测试无效配置
	invalidCfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host-12345",
			Port:     9999,
			User:     "invalid",
			Password: "invalid",
			DBName:   "invalid",
		},
	}

	err := InitPostgreSQL(invalidCfg)
	assert.Error(t, err, "Should fail with invalid config")
	// 错误消息可能是 "failed to connect to database" 或 "failed to ping database"
	assert.True(t,
		strings.Contains(err.Error(), "failed to connect") ||
		strings.Contains(err.Error(), "failed to ping"),
		"Error should mention connection/ping failure, got: %v", err)
}

// TestCloseDatabase 测试关闭数据库连接
func TestCloseDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupTestEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	// 初始化
	err = InitPostgreSQL(cfg)
	require.NoError(t, err)

	// 关闭
	err = Close()
	assert.NoError(t, err, "Close should succeed")

	// 再次关闭应该安全
	err = Close()
	assert.NoError(t, err, "Close should be idempotent")
}

// setupTestEnv 设置测试环境变量
func setupTestEnv(t *testing.T) {
	// 设置数据库连接环境变量
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "ahavault")
	os.Setenv("POSTGRES_PASSWORD", "ahavault_dev")
	os.Setenv("POSTGRES_DB", "ahavault")

	// 设置必需的 APP_MASTER_KEY (测试用 64 字符 HEX)
	os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	t.Cleanup(func() {
		// 测试结束后清理
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DB")
		os.Unsetenv("APP_MASTER_KEY")
	})
}
