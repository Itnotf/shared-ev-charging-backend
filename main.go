package main

import (
	"shared-charge/config"
	"shared-charge/controllers"
	"shared-charge/middleware"
	"shared-charge/models"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title 充电位共享小程序 API
// @version 1.0
// @description 为微信群里的充电共享场景提供API服务
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化日志
	cfg := config.GetConfig()
	if err := utils.InitLogger(cfg.Log.Mode, cfg.Log.Level, cfg.Log.FilePath); err != nil {
		panic("初始化日志失败: " + err.Error())
	}

	// 初始化数据库
	models.InitDB()

	// 初始化MinIO客户端
	if err := utils.InitMinioClient(); err != nil {
		utils.Fatal("初始化MinIO客户端失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(config.GetConfig().Server.Mode)

	// 创建路由（禁用默认日志）
	r := gin.New()
	r.Use(middleware.TraceMiddleware())       // 添加trace中间件
	r.Use(middleware.PerformanceMiddleware()) // 添加性能监控中间件
	r.Use(middleware.CORS())

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", controllers.WechatLogin)
			auth.POST("/refresh", middleware.AuthMiddleware(), controllers.RefreshToken)
		}

		// 用户相关
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware())
		{
			users.GET("/profile", controllers.GetUserProfile)
			users.POST("/profile", controllers.UpdateUserProfile) // 新增的路由
			users.GET("/price", controllers.GetUserPrice)
		}

		// 预约相关
		reservations := api.Group("/reservations")
		reservations.Use(middleware.AuthMiddleware())
		{
			reservations.GET("", controllers.GetReservations)
			reservations.POST("", controllers.CreateReservation)
			reservations.DELETE(":id", controllers.DeleteReservation)
			reservations.GET("/current", controllers.GetCurrentReservation)
			reservations.GET("/current-status", controllers.GetCurrentStatus)
		}

		// 充电记录相关
		records := api.Group("/records")
		records.Use(middleware.AuthMiddleware())
		{
			records.GET("", controllers.GetRecords)
			records.POST("", controllers.CreateRecord)
			records.GET("/unsubmitted", controllers.GetUnsubmittedRecords)
		}

		// 文件上传
		upload := api.Group("/upload")
		upload.Use(middleware.AuthMiddleware())
		{
			upload.POST("/image", controllers.UploadImage)
		}

		// 统计相关
		statistics := api.Group("/statistics")
		statistics.Use(middleware.AuthMiddleware())
		{
			statistics.GET("/monthly", controllers.GetMonthlyStatistics)
			statistics.GET("/daily", controllers.GetDailyStatistics)
			statistics.GET("/monthly-shift", controllers.GetMonthlyShiftStatistics)
		}

	}

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "充电位共享小程序服务运行正常",
		})
	})

	// 启动服务器
	port := config.GetConfig().Server.Port
	utils.Info("服务器启动在端口 %s", port)
	if err := r.Run(":" + port); err != nil {
		utils.Fatal("服务器启动失败: %v", err)
	}
}
