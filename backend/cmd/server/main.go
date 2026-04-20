package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "code-repo/internal/api/v1"
	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"
	"code-repo/internal/repository/storage"
	"code-repo/internal/service"
	"code-repo/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	configPath := "configs/config.yaml"
	if os.Getenv("APP_ENV") == "docker" {
		configPath = "/root/configs/config.yaml"
	}
	utils.LoadConfig(configPath)

	// 2. 构造数据库连接 DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=Asia/Shanghai",
		utils.Config.Database.User,
		utils.Config.Database.Password,
		utils.Config.Database.Host,
		utils.Config.Database.Port,
		utils.Config.Database.DBName,
		utils.Config.Database.SSLMode,
	)

	// 3. 初始化数据库连接（带重试逻辑，解决 Docker 容器启动顺序问题）
	var dbConn *gorm.DB
	var err error
	for i := 0; i < 5; i++ {
		dbConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("正在等待数据库就绪 (%d/5)...", i+1)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("数据库连接最终失败: %v", err)
	}

	// 4. 自动迁移数据库表结构
	// 包含 User 和 Repository 实体
	dbConn.AutoMigrate(&entity.User{}, &entity.Repository{})

	// 5. 初始化其他基础设施
	storage.InitMinio(utils.Config.Minio) //

	// 6. 依赖注入 (Dependency Injection)

	// --- User 模块 ---
	userRepo := db.NewUserRepository(dbConn)
	userSvc := service.NewUserService(userRepo)
	userHandler := v1.NewUserHandler(userSvc)

	// --- Repository 模块 ---
	repoDB := db.NewRepoRepository(dbConn)
	repoSvc := service.NewRepoService(repoDB)
	repoHandler := v1.NewRepoHandler(repoSvc)

	// 7. 注册路由
	r := gin.Default()

	apiV1 := r.Group("/api/v1")
	{
		// 用户相关接口
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)

		// 仓库相关接口
		apiV1.POST("/repos", repoHandler.Create)       // 创建仓库
		apiV1.GET("/repos/:id", repoHandler.GetDetail) // 获取仓库详情
		apiV1.DELETE("/repos/:id", repoHandler.Delete) // 删除仓库
	}

	// 8. 启动服务与优雅关机
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", utils.Config.Server.Port),
		Handler: r,
	}

	go func() {
		log.Printf("代码托管平台后端服务运行在端口: %d", utils.Config.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务强制关闭:", err)
	}
	log.Println("服务已安全退出")
}
