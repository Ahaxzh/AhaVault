package database

import (
	"context"
	"fmt"
	"time"

	"ahavault/server/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// InitPostgreSQL 初始化 PostgreSQL 连接
func InitPostgreSQL(cfg *config.Config) error {
	// 配置 GORM 日志级别
	logLevel := logger.Silent
	if cfg.App.Debug {
		logLevel = logger.Info
	}

	// 打开数据库连接
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // 启用预编译语句缓存
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 设置全局实例
	DB = db

	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// HealthCheck 数据库健康检查
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Stats 数据库连接池统计信息
type Stats struct {
	MaxOpenConnections int           // 最大打开连接数
	OpenConnections    int           // 当前打开连接数
	InUse              int           // 正在使用的连接数
	Idle               int           // 空闲连接数
	WaitCount          int64         // 等待连接的总次数
	WaitDuration       time.Duration // 等待连接的总时间
	MaxIdleClosed      int64         // 因超过最大空闲时间而关闭的连接数
	MaxLifetimeClosed  int64         // 因超过最大生命周期而关闭的连接数
}

// GetStats 获取数据库连接池统计信息
func GetStats() (*Stats, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}

	dbStats := sqlDB.Stats()

	return &Stats{
		MaxOpenConnections: dbStats.MaxOpenConnections,
		OpenConnections:    dbStats.OpenConnections,
		InUse:              dbStats.InUse,
		Idle:               dbStats.Idle,
		WaitCount:          dbStats.WaitCount,
		WaitDuration:       dbStats.WaitDuration,
		MaxIdleClosed:      dbStats.MaxIdleClosed,
		MaxLifetimeClosed:  dbStats.MaxLifetimeClosed,
	}, nil
}
