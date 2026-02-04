package api

import (
	"ahavault/server/internal/api/handlers"
	"ahavault/server/internal/middleware"
	"ahavault/server/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(
	router *gin.Engine,
	userService *services.UserService,
	fileService *services.FileService,
	shareService *services.ShareService,
) {
	// 创建处理器
	authHandler := handlers.NewAuthHandler(userService)
	fileHandler := handlers.NewFileHandler(fileService)
	shareHandler := handlers.NewShareHandler(shareService)

	// 应用全局中间件
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())

	// API 路由组
	api := router.Group("/api")
	{
		// 认证路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// 公开分享路由（无需认证）
		public := api.Group("/public")
		{
			public.POST("/shares/:code", shareHandler.GetShareByCode)
		}

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.Auth(userService))
		{
			// 用户路由
			user := authenticated.Group("/user")
			{
				user.GET("/me", authHandler.GetCurrentUser)
			}

			// 文件路由
			files := authenticated.Group("/files")
			{
				files.GET("", fileHandler.ListFiles)
				files.POST("/check", fileHandler.CheckInstantUpload)
				files.POST("", fileHandler.CreateFileMetadata)
				files.POST("/upload", fileHandler.UploadFile)
				files.GET("/:id/download", fileHandler.DownloadFile)
				files.DELETE("/:id", fileHandler.DeleteFile)
			}

			// 分享路由
			shares := authenticated.Group("/shares")
			{
				shares.GET("", shareHandler.ListMyShares)
				shares.POST("", shareHandler.CreateShare)
				shares.POST("/:code/save", shareHandler.SaveToVault)
				shares.DELETE("/:id", shareHandler.StopShare)
			}
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
