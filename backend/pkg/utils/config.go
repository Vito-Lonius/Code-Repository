package utils

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig 是全局配置的载体
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Minio    MinioConfig    `mapstructure:"minio"`
	Git      GitConfig      `mapstructure:"git"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Jwt      JwtConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"` // 修改为 int，与 yaml 匹配
	Mode string `mapstructure:"mode"`
}

type JwtConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MinioConfig struct {
	Endpoint   string `mapstructure:"endpoint"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	UseSSL     bool   `mapstructure:"use_ssl"`
	BucketName string `mapstructure:"bucket_name"`
	TempBucket string `mapstructure:"temp_bucket"`
}

// GitConfig 对应 yaml 中的 git 存储配置
type GitConfig struct {
	RootPath      string `mapstructure:"root_path"` // ✨ 必须与 YAML 中的 key 一致
	DefaultBranch string `mapstructure:"default_branch"`
}

var Config AppConfig

// LoadConfig 读取并解析配置文件
func LoadConfig(path string) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// Viper 使用 mapstructure 标签进行反序列化
	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	log.Println("配置文件加载成功!")
}
