package main

import (
	"log"
	"os"

	"code-repo/internal/repository/db"
	"code-repo/internal/repository/storage"
	"code-repo/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 确定配置文件路径（支持本地调试与 Docker 容器双环境）
	configPath := "configs/config.yaml"
	if os.Getenv("APP_ENV") == "docker" {
		configPath = "/root/configs/config.yaml"
	}

	// 2. 加载配置
	log.Println("正在启动代码托管平台后端服务...")
	utils.LoadConfig(configPath)

	// 3. 初始化基础设施
	db.InitPostgres(utils.Config.Database)
	storage.InitMinio(utils.Config.Minio)

	// 初始化 Redis (待补充)

	// 4. 启动 Gin HTTP 服务
	log.Printf("基础设施初始化完毕！服务正运行在端口: %s\n", utils.Config.Server.Port)

	r := gin.Default()

	// 定义一个简单的健康检查接口，防止程序因无路由而看起来像没启动
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 启动服务（这会阻塞主线程，让容器保持运行）
	if err := r.Run(":" + utils.Config.Server.Port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
