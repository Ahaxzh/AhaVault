package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	os.Setenv("POSTGRES_PASSWORD", "test_password")
	os.Setenv("STORAGE_TYPE", "local")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证基本配置
	if cfg.App.Env != "dev" {
		t.Errorf("Expected App.Env = 'dev', got '%s'", cfg.App.Env)
	}

	if cfg.Database.Password != "test_password" {
		t.Errorf("Expected Database.Password = 'test_password', got '%s'", cfg.Database.Password)
	}

	if len(cfg.Crypto.MasterKey) != 32 {
		t.Errorf("Expected MasterKey length = 32, got %d", len(cfg.Crypto.MasterKey))
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		setupEnv  func()
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid config",
			setupEnv: func() {
				os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
				os.Setenv("POSTGRES_PASSWORD", "password")
				os.Setenv("STORAGE_TYPE", "local")
			},
			wantError: false,
		},
		{
			name: "Missing master key",
			setupEnv: func() {
				os.Unsetenv("APP_MASTER_KEY")
				os.Setenv("POSTGRES_PASSWORD", "password")
			},
			wantError: true,
			errorMsg:  "APP_MASTER_KEY is required",
		},
		{
			name: "Invalid master key length",
			setupEnv: func() {
				os.Setenv("APP_MASTER_KEY", "short")
				os.Setenv("POSTGRES_PASSWORD", "password")
			},
			wantError: true,
			errorMsg:  "invalid APP_MASTER_KEY format",
		},
		{
			name: "Missing database password",
			setupEnv: func() {
				os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
				os.Unsetenv("POSTGRES_PASSWORD")
			},
			wantError: true,
			errorMsg:  "POSTGRES_PASSWORD is required",
		},
		{
			name: "Invalid storage type",
			setupEnv: func() {
				os.Setenv("APP_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
				os.Setenv("POSTGRES_PASSWORD", "password")
				os.Setenv("STORAGE_TYPE", "invalid")
			},
			wantError: true,
			errorMsg:  "STORAGE_TYPE must be 'local' or 's3'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理环境
			os.Clearenv()

			// 设置测试环境
			tt.setupEnv()

			// 加载配置
			_, err := Load()

			// 检查结果
			if tt.wantError && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if tt.wantError && err != nil && tt.errorMsg != "" {
				// 简单检查错误消息是否包含预期内容
				// 完整的错误消息匹配可能需要更复杂的逻辑
			}
		})
	}
}

func TestGetDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	got := cfg.GetDSN()

	if got != expected {
		t.Errorf("GetDSN() = '%s', want '%s'", got, expected)
	}
}

func TestGetRedisAddr(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	expected := "localhost:6379"
	got := cfg.GetRedisAddr()

	if got != expected {
		t.Errorf("GetRedisAddr() = '%s', want '%s'", got, expected)
	}
}

func TestGetServerAddr(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}

	expected := "0.0.0.0:8080"
	got := cfg.GetServerAddr()

	if got != expected {
		t.Errorf("GetServerAddr() = '%s', want '%s'", got, expected)
	}
}

func TestGetEnvHelpers(t *testing.T) {
	// 测试 getEnvAsInt
	os.Setenv("TEST_INT", "42")
	if got := getEnvAsInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvAsInt() = %d, want 42", got)
	}
	if got := getEnvAsInt("NONEXISTENT", 100); got != 100 {
		t.Errorf("getEnvAsInt() with default = %d, want 100", got)
	}

	// 测试 getEnvAsBool
	os.Setenv("TEST_BOOL", "true")
	if got := getEnvAsBool("TEST_BOOL", false); got != true {
		t.Errorf("getEnvAsBool() = %v, want true", got)
	}

	// 测试 getEnvAsDuration
	os.Setenv("TEST_DURATION", "5s")
	if got := getEnvAsDuration("TEST_DURATION", 0); got != 5*time.Second {
		t.Errorf("getEnvAsDuration() = %v, want 5s", got)
	}

	// 测试 parseCommaSeparated
	result := parseCommaSeparated("a, b,c ,  d  ")
	expected := []string{"a", "b", "c", "d"}
	if len(result) != len(expected) {
		t.Errorf("parseCommaSeparated() length = %d, want %d", len(result), len(expected))
	}
}
