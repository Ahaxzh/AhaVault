package main

import (
	"fmt"
	"log"

	"ahavault/server/internal/api"
	"ahavault/server/internal/config"
	"ahavault/server/internal/database"
	"ahavault/server/internal/models"
	"ahavault/server/internal/services"
	"ahavault/server/internal/storage"
	"ahavault/server/internal/tasks"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置 Gin 模式
	if cfg.Server.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 连接数据库
	if err := database.InitPostgreSQL(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// 自动迁移数据库
	log.Println("Starting database migration...")
	
	// DROP VIEW specifically to allow GORM to alter column types
	log.Println("Dropping legacy views...")
	if err := database.DB.Exec("DROP VIEW IF EXISTS user_storage_stats CASCADE").Error; err != nil {
		log.Printf("Warning: Failed to drop view user_storage_stats: %v", err)
	}
	if err := database.DB.Exec("DROP VIEW IF EXISTS active_shares CASCADE").Error; err != nil {
		log.Printf("Warning: Failed to drop view active_shares: %v", err)
	}

	// Fix schema drift: Drop 'password' column if it exists
	log.Println("Fixing schema drift: Dropping 'password' column...")
	if err := database.DB.Exec("ALTER TABLE users DROP COLUMN IF EXISTS password").Error; err != nil {
		log.Printf("Warning: Failed to drop password column: %v", err)
	}

	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.FileMetadata{},
		&models.FileBlob{},
		&models.ShareSession{},
		&models.ShareFile{},
		&models.UploadSession{},
		&models.AuditLog{},
		&models.SystemSetting{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 连接 Redis
	if err := database.InitRedis(cfg); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer database.CloseRedis()

	// 初始化存储引擎
	var storageEngine storage.Engine
	switch cfg.Storage.Type {
	case "local":
		storageEngine, err = storage.NewLocalEngine(cfg.Storage.LocalPath)
		if err != nil {
			log.Fatalf("Failed to initialize local storage: %v", err)
		}
	default:
		log.Fatalf("Unsupported storage type: %s", cfg.Storage.Type)
	}

	// 创建服务实例
	userService := services.NewUserService(database.DB, cfg.Crypto.JWTSecret)
	fileService := services.NewFileService(database.DB, storageEngine, cfg.Crypto.MasterKey)
	shareService := services.NewShareService(database.DB, fileService)

	// 启动后台任务调度器
	scheduler := tasks.NewScheduler(database.DB, storageEngine)
	if err := scheduler.Start(); err != nil {
		log.Printf("Warning: Failed to start background scheduler: %v", err)
	} else {
		log.Println("Background task scheduler started successfully")
	}
	defer scheduler.Stop()

	// 创建 Gin 路由
	router := gin.Default()

	// 设置路由
	api.SetupRoutes(router, userService, fileService, shareService)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting AhaVault server on %s (environment: %s)", addr, cfg.App.Env)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
