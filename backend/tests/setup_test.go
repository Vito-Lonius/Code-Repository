package tests

import (
	"fmt"
	"os"
	"testing"

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

var (
	testDB     *gorm.DB
	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	// 1. 加载配置
	utils.LoadConfig("../configs/config.yaml")

	// 2. 宿主机运行环境适配 (核心修复点)
	dbHost := utils.Config.Database.Host
	if os.Getenv("IN_DOCKER") != "true" {
		dbHost = "127.0.0.1"
		// ⚠️ 修复：根据 docker-compose.yml，Minio API 映射到了 9005
		utils.Config.Minio.Endpoint = "127.0.0.1:9005"
		utils.Config.Minio.UseSSL = false
	}

	// 3. 构造 DSN (使用 KV 格式，在 Windows 上更稳健)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		dbHost,
		utils.Config.Database.User,
		utils.Config.Database.Password,
		utils.Config.Database.DBName,
		utils.Config.Database.Port,
	)

	var err error
	// 使用显式配置打开连接
	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		fmt.Printf("无法连接测试数据库: %v\n", err)
		os.Exit(1)
	}

	// 4. 自动迁移并初始化基础设施
	testDB.AutoMigrate(&entity.User{}, &entity.Repository{})
	storage.InitMinio(utils.Config.Minio)

	// 5. 依赖注入 (DI)
	userRepo := db.NewUserRepository(testDB)
	userSvc := service.NewUserService(userRepo)
	userHandler := v1.NewUserHandler(userSvc)

	repoDB := db.NewRepoRepository(testDB)
	repoSvc := service.NewRepoService(repoDB)
	repoHandler := v1.NewRepoHandler(repoSvc)

	// 6. 配置测试路由
	gin.SetMode(gin.TestMode)
	testRouter = gin.Default()
	apiV1 := testRouter.Group("/api/v1")
	{
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)
		apiV1.POST("/repos", repoHandler.Create)
		apiV1.GET("/repos/:id", repoHandler.GetDetail)
	}

	os.Exit(m.Run())
}
