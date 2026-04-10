package utils

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig 是全局配置的载体
type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Minio    MinioConfig
	// 暂时省略 Redis, Git, SonarQube 等，按需在此处补充对应结构体
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type MinioConfig struct {
	Endpoint   string
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	UseSSL     bool   `mapstructure:"use_ssl"`
	BucketName string `mapstructure:"bucket_name"`
	TempBucket string `mapstructure:"temp_bucket"`
}

var Config AppConfig

// LoadConfig 读取并解析配置文件
func LoadConfig(path string) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv() // 允许从环境变量覆盖配置

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	log.Println("配置文件加载成功!")
}
