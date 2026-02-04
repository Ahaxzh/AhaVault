package main

import (
	"fmt"
	"log"

	"ahavault/server/internal/api"
	"ahavault/server/internal/config"
	"ahavault/server/internal/database"
	"ahavault/server/internal/services"
	"ahavault/server/internal/storage"
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
