package tests

import (
	"fmt"
	"os"
	"testing"

	v1 "code-repo/internal/api/v1"
	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"
	"code-repo/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	// 1. 直接连接 Docker 环境中的 Postgres 容器
	// host=postgres 对应 docker-compose 中的服务名
	dsn := "host=localhost user=user password=password dbname=code_repository port=5432 sslmode=disable"

	var err error
	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("无法连接测试数据库: %v\n", err)
		os.Exit(1)
	}

	// 2. 自动迁移表结构
	testDB.AutoMigrate(&entity.User{})

	// 3. 依赖注入
	userRepo := db.NewUserRepository(testDB)
	userSvc := service.NewUserService(userRepo)
	userHandler := v1.NewUserHandler(userSvc)

	// 4. 配置测试路由
	gin.SetMode(gin.TestMode)
	testRouter = gin.Default()
	apiV1 := testRouter.Group("/api/v1")
	{
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)
	}

	// 5. 执行所有测试
	os.Exit(m.Run())
}
