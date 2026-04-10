package db

import (
	"fmt"
	"log"
	"time"

	"code-repo/pkg/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitPostgres 初始化数据库连接
func InitPostgres(cfg utils.DatabaseConfig) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 PostgreSQL 失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取底层 sql.DB 失败: %v", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("PostgreSQL 数据库连接成功!")

	// TODO: 稍后在这里添加 AutoMigrate 以自动建表
	// err = DB.AutoMigrate(&entity.User{}, &entity.Repository{})
}
