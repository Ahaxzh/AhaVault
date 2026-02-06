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
	// Create handlers
	authHandler := handlers.NewAuthHandler(userService)
	fileHandler := handlers.NewFileHandler(fileService)
	shareHandler := handlers.NewShareHandler(shareService)
	downloadHandler := handlers.NewDownloadHandler(shareService, fileService)

	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())

	// API Routes Group
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Public share routes (public)
		public := api.Group("/public")
		{
			public.POST("/shares/:code", shareHandler.GetShareByCode)
			public.GET("/download/:code", downloadHandler.DownloadByPickupCode)
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

			// Tus Upload Routes
			tusHandler := handlers.NewTusHandler(fileService, "./tmp/tus_uploads")
			// We handle both base path and wildcards for Tus protocol (POST, HEAD, PATCH, OPTIONS, DELETE)
			authenticated.Any("/tus/upload", tusHandler.GinHandler)
			authenticated.Any("/tus/upload/*any", tusHandler.GinHandler)
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
