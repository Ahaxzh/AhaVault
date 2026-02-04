package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config 全局配置结构
type Config struct {
	// 应用配置
	App AppConfig

	// 数据库配置
	Database DatabaseConfig

	// Redis 配置
	Redis RedisConfig

	// 存储配置
	Storage StorageConfig

	// 加密配置
	Crypto CryptoConfig

	// 服务器配置
	Server ServerConfig

	// 业务配置
	Business BusinessConfig
}

// AppConfig 应用配置
type AppConfig struct {
	Env        string // 环境：dev, staging, production
	Debug      bool   // 调试模式
	LogLevel   string // 日志级别：debug, info, warn, error
	InviteCode string // 邀请码
	Version    string // 应用版本
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int           // 数据库编号 (0-15)
	PoolSize int           // 连接池大小
	Timeout  time.Duration // 超时时间
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type string // 存储类型：local, s3

	// Local 存储配置
	LocalPath string

	// S3 存储配置
	S3Endpoint  string
	S3Region    string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3UseSSL    bool
}

// CryptoConfig 加密配置
type CryptoConfig struct {
	MasterKey []byte // KEK (Key Encryption Key) - 32 bytes
	JWTSecret string // JWT 签名密钥
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	GinMode      string // debug, release, test
}

// BusinessConfig 业务配置
type BusinessConfig struct {
	// 文件相关
	MaxFileSize      int64 // 单文件最大大小（字节）
	AllowedMimeTypes []string

	// 存储配额
	DefaultUserQuota int64 // 新用户默认配额（字节）

	// 分享相关
	ShareCodeLength     int           // 取件码长度
	DefaultShareExpiry  time.Duration // 默认分享有效期
	MaxShareExpiry      time.Duration // 最大分享有效期
	MaxFilesPerShare    int           // 单次分享最大文件数
	MaxActiveSharesUser int           // 单用户最大活跃分享数

	// 垃圾回收
	GCRetentionDays     int           // 软删除保留天数
	GCCleanupInterval   time.Duration // GC 清理间隔
	GCFragmentRetention time.Duration // 上传碎片保留时间

	// 注册控制
	RegistrationEnabled bool // 是否开启注册
	InviteCodeRequired  bool // 是否需要邀请码
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载 .env 文件（忽略错误，允许纯环境变量部署）
	_ = godotenv.Load()

	cfg := &Config{}

	// 加载各模块配置
	if err := cfg.loadAppConfig(); err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	if err := cfg.loadDatabaseConfig(); err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	if err := cfg.loadRedisConfig(); err != nil {
		return nil, fmt.Errorf("failed to load redis config: %w", err)
	}

	if err := cfg.loadStorageConfig(); err != nil {
		return nil, fmt.Errorf("failed to load storage config: %w", err)
	}

	if err := cfg.loadCryptoConfig(); err != nil {
		return nil, fmt.Errorf("failed to load crypto config: %w", err)
	}

	if err := cfg.loadServerConfig(); err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	if err := cfg.loadBusinessConfig(); err != nil {
		return nil, fmt.Errorf("failed to load business config: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// loadAppConfig 加载应用配置
func (c *Config) loadAppConfig() error {
	c.App = AppConfig{
		Env:        getEnvOrDefault("APP_ENV", "dev"),
		Debug:      getEnvAsBool("APP_DEBUG", true),
		LogLevel:   getEnvOrDefault("LOG_LEVEL", "debug"),
		InviteCode: getEnvOrDefault("APP_INVITE_CODE", ""),
		Version:    getEnvOrDefault("APP_VERSION", "0.1.0"),
	}
	return nil
}

// loadDatabaseConfig 加载数据库配置
func (c *Config) loadDatabaseConfig() error {
	c.Database = DatabaseConfig{
		Host:            getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            getEnvAsInt("POSTGRES_PORT", 5432),
		User:            getEnvOrDefault("POSTGRES_USER", "ahavault"),
		Password:        getEnvOrDefault("POSTGRES_PASSWORD", ""),
		DBName:          getEnvOrDefault("POSTGRES_DB", "ahavault"),
		SSLMode:         getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}
	return nil
}

// loadRedisConfig 加载 Redis 配置
func (c *Config) loadRedisConfig() error {
	c.Redis = RedisConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     getEnvAsInt("REDIS_PORT", 6379),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
		PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		Timeout:  getEnvAsDuration("REDIS_TIMEOUT", 5*time.Second),
	}
	return nil
}

// loadStorageConfig 加载存储配置
func (c *Config) loadStorageConfig() error {
	c.Storage = StorageConfig{
		Type:      getEnvOrDefault("STORAGE_TYPE", "local"),
		LocalPath: getEnvOrDefault("STORAGE_PATH", "/data/storage"),

		// S3 配置
		S3Endpoint:  getEnvOrDefault("S3_ENDPOINT", ""),
		S3Region:    getEnvOrDefault("S3_REGION", "us-east-1"),
		S3Bucket:    getEnvOrDefault("S3_BUCKET", "ahavault"),
		S3AccessKey: getEnvOrDefault("S3_ACCESS_KEY", ""),
		S3SecretKey: getEnvOrDefault("S3_SECRET_KEY", ""),
		S3UseSSL:    getEnvAsBool("S3_USE_SSL", true),
	}
	return nil
}

// loadCryptoConfig 加载加密配置
func (c *Config) loadCryptoConfig() error {
	masterKeyHex := os.Getenv("APP_MASTER_KEY")
	if masterKeyHex == "" {
		return fmt.Errorf("APP_MASTER_KEY is required")
	}

	// 解码 HEX 字符串
	masterKey, err := hex.DecodeString(masterKeyHex)
	if err != nil {
		return fmt.Errorf("invalid APP_MASTER_KEY format (must be 64-char HEX): %w", err)
	}

	if len(masterKey) != 32 {
		return fmt.Errorf("APP_MASTER_KEY must be 32 bytes (64 hex chars), got %d bytes", len(masterKey))
	}

	// 读取 JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-please-change-in-production"
	}

	c.Crypto = CryptoConfig{
		MasterKey: masterKey,
		JWTSecret: jwtSecret,
	}
	return nil
}

// loadServerConfig 加载服务器配置
func (c *Config) loadServerConfig() error {
	c.Server = ServerConfig{
		Host:         getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
		Port:         getEnvAsInt("SERVER_PORT", 8080),
		ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		GinMode:      getEnvOrDefault("GIN_MODE", "debug"),
	}
	return nil
}

// loadBusinessConfig 加载业务配置
func (c *Config) loadBusinessConfig() error {
	c.Business = BusinessConfig{
		// 文件相关
		MaxFileSize:      getEnvAsInt64("MAX_FILE_SIZE", 2*1024*1024*1024), // 2GB
		AllowedMimeTypes: parseCommaSeparated(getEnvOrDefault("ALLOWED_MIME_TYPES", "image/*,video/*,application/pdf,application/zip")),

		// 存储配额
		DefaultUserQuota: getEnvAsInt64("DEFAULT_USER_QUOTA", 10*1024*1024*1024), // 10GB

		// 分享相关
		ShareCodeLength:     getEnvAsInt("SHARE_CODE_LENGTH", 8),
		DefaultShareExpiry:  getEnvAsDuration("DEFAULT_SHARE_EXPIRY", 24*time.Hour),
		MaxShareExpiry:      getEnvAsDuration("MAX_SHARE_EXPIRY", 7*24*time.Hour),
		MaxFilesPerShare:    getEnvAsInt("MAX_FILES_PER_SHARE", 100),
		MaxActiveSharesUser: getEnvAsInt("MAX_ACTIVE_SHARES_USER", 50),

		// 垃圾回收
		GCRetentionDays:     getEnvAsInt("GC_RETENTION_DAYS", 7),
		GCCleanupInterval:   getEnvAsDuration("GC_CLEANUP_INTERVAL", 1*time.Hour),
		GCFragmentRetention: getEnvAsDuration("GC_FRAGMENT_RETENTION", 24*time.Hour),

		// 注册控制
		RegistrationEnabled: getEnvAsBool("REGISTRATION_ENABLED", true),
		InviteCodeRequired:  getEnvAsBool("INVITE_CODE_REQUIRED", false),
	}
	return nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证数据库配置
	if c.Database.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	// 验证存储配置
	if c.Storage.Type != "local" && c.Storage.Type != "s3" {
		return fmt.Errorf("STORAGE_TYPE must be 'local' or 's3', got: %s", c.Storage.Type)
	}

	if c.Storage.Type == "s3" {
		if c.Storage.S3AccessKey == "" || c.Storage.S3SecretKey == "" {
			return fmt.Errorf("S3_ACCESS_KEY and S3_SECRET_KEY are required for S3 storage")
		}
	}

	// 验证加密配置
	if len(c.Crypto.MasterKey) != 32 {
		return fmt.Errorf("master key must be 32 bytes")
	}

	// 验证业务配置
	if c.Business.ShareCodeLength < 6 || c.Business.ShareCodeLength > 12 {
		return fmt.Errorf("SHARE_CODE_LENGTH must be between 6 and 12, got: %d", c.Business.ShareCodeLength)
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddr 获取 Redis 地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr 获取服务器地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// ==========================================
// 辅助函数
// ==========================================

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量作为整数
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsInt64 获取环境变量作为 int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsBool 获取环境变量作为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsDuration 获取环境变量作为时间间隔
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// parseCommaSeparated 解析逗号分隔的字符串
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
