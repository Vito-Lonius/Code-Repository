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
	"code-repo/internal/api/middleware"
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
	dbConn.AutoMigrate(&entity.User{}, &entity.Repository{}, &entity.File{}, &entity.UploadTask{})

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

	// --- File 模块 ---
	fileRepo := db.NewFileRepository(dbConn)
	uploadTaskRepo := db.NewUploadTaskRepository(dbConn)
	fileSvc := service.NewFileService(fileRepo, uploadTaskRepo, repoDB)
	fileHandler := v1.NewFileHandler(fileSvc)

	// --- Preview 模块 ---
	previewSvc := service.NewPreviewService(fileRepo, repoDB)
	previewHandler := v1.NewPreviewHandler(previewSvc)

	// --- 中间件 ---
	authMW := middleware.AuthMiddleware()

	// 7. 注册路由
	r := gin.Default()

	apiV1 := r.Group("/api/v1")
	{
		// 用户相关接口
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)

		// 仓库相关接口
		apiV1.POST("/repos", repoHandler.Create)
		apiV1.GET("/repos/:id", repoHandler.GetDetail)
		apiV1.DELETE("/repos/:id", repoHandler.Delete)

		// 文件相关接口（需要认证）
		fileGroup := apiV1.Group("")
		fileGroup.Use(authMW)
		{
			fileGroup.POST("/files/upload", fileHandler.UploadSimple)
			fileGroup.POST("/files/upload/init", fileHandler.UploadInit)
			fileGroup.POST("/files/upload/chunk", fileHandler.UploadChunk)
			fileGroup.POST("/files/upload/complete", fileHandler.UploadComplete)
			fileGroup.GET("/files/:id", fileHandler.GetFileDetail)
			fileGroup.GET("/files/:id/download", fileHandler.DownloadFile)
			fileGroup.DELETE("/files/:id", fileHandler.DeleteFile)
			fileGroup.GET("/files", fileHandler.ListFiles)
			fileGroup.POST("/files/dir", fileHandler.CreateDir)
			fileGroup.PUT("/files/:id/rename", fileHandler.RenameFile)
			fileGroup.PUT("/files/:id/move", fileHandler.MoveFile)

			// 预览相关接口（需要认证）
			fileGroup.GET("/preview/:id", previewHandler.PreviewFile)
			fileGroup.GET("/preview/:id/info", previewHandler.GetPreviewInfo)
			fileGroup.GET("/preview/:id/raw", previewHandler.GetRawFile)
			fileGroup.GET("/preview/:id/image", previewHandler.PreviewImage)
			fileGroup.GET("/preview/:id/media", previewHandler.PreviewMedia)
			fileGroup.GET("/preview/:id/pdf", previewHandler.PreviewPDF)
			fileGroup.GET("/repos/:repo_id/tree", previewHandler.GetDirectoryTree)
			fileGroup.GET("/repos/:repo_id/dir", previewHandler.ListDir)
		}
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
