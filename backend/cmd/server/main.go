package main

import (
	"log"
	"os"

	v1 "code-repo/internal/api/v1" // 引入你的 API 层
	"code-repo/internal/repository/db"
	"code-repo/internal/repository/storage"
	"code-repo/internal/service"
	"code-repo/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 确定配置文件路径
	configPath := "configs/config.yaml"
	if os.Getenv("APP_ENV") == "docker" {
		configPath = "/root/configs/config.yaml"
	}

	// 2. 加载配置
	log.Println("正在启动代码托管平台后端服务...")
	utils.LoadConfig(configPath)

	// 3. 初始化基础设施
	// 注意：确保 db.InitPostgres 会初始化全局变量 db.DB 或返回实例
	db.InitPostgres(utils.Config.Database)
	storage.InitMinio(utils.Config.Minio)

	// TODO: 初始化 Redis

	// 4. 依赖注入 (Dependency Injection)
	// 假设 db.DB 是你在 InitPostgres 中初始化的全局 *gorm.DB 实例
	userRepo := db.NewUserRepository(db.DB)
	userSvc := service.NewUserService(userRepo)
	userHandler := v1.NewUserHandler(userSvc)

	// 5. 启动 Gin HTTP 服务并注册路由
	r := gin.Default()

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 注册 API 路由
	apiV1 := r.Group("/api/v1")
	{
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)
		// 注意：GetProfile 需要 AuthMiddleware 保护，建议后续引入后加上
		// authGroup := apiV1.Group("/")
		// authGroup.Use(middleware.AuthMiddleware())
		// authGroup.GET("/user/profile", userHandler.GetProfile)
	}

	log.Printf("基础设施初始化完毕！服务正运行在端口: %s\n", utils.Config.Server.Port)
	if err := r.Run(":" + utils.Config.Server.Port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
