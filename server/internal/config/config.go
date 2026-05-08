package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 全局配置
var Config *AppConfig

// AppConfig 应用配置
type AppConfig struct {
	App      App      // 应用配置
	Database Database // 数据库配置
	Redis    Redis    // Redis配置
	JWT      JWT      // JWT配置
	WechatPay WechatPay // 微信支付配置
	Qiniu    Qiniu    // 七牛云配置
}

// App 应用配置
type App struct {
	Name string
	Host string
	Port int
	Env  string
	Debug bool
}

// Database 数据库配置
type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Charset  string
	MaxIdle  int
	MaxOpen  int
	LifeTime int
}

// Redis Redis配置
type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWT JWT配置
type JWT struct {
	Secret string
	Expire int // 分钟
}

// WechatPay 微信支付配置
type WechatPay struct {
	MchID       string
	APIKey      string
	CallbackURL string
}

// Qiniu 七牛云配置
type Qiniu struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

// InitConfig 初始化配置
func InitConfig() error {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 设置配置文件路径
	configPath := filepath.Join(workDir, ".env")
	
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Println("配置文件不存在，使用默认配置")
		// 创建默认配置
		Config = getDefaultConfig()
		return nil
	}

	// 读取配置文件
	viper.SetConfigFile(configPath)
	viper.SetConfigType("env")
	
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	Config = &AppConfig{
		App: App{
			Name: viper.GetString("APP_NAME"),
			Host: viper.GetString("APP_HOST"),
			Port: viper.GetInt("APP_PORT"),
			Env: viper.GetString("APP_ENV"),
			Debug: viper.GetBool("APP_DEBUG"),
		},
		Database: Database{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetInt("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			Charset:  viper.GetString("DB_CHARSET"),
			MaxIdle:  viper.GetInt("DB_MAX_IDLE"),
			MaxOpen:  viper.GetInt("DB_MAX_OPEN"),
			LifeTime: viper.GetInt("DB_LIFE_TIME"),
		},
		Redis: Redis{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWT{
			Secret: viper.GetString("JWT_SECRET"),
			Expire: viper.GetInt("JWT_EXPIRE"),
		},
		WechatPay: WechatPay{
			MchID:       viper.GetString("WECHAT_PAY_MCH_ID"),
			APIKey:      viper.GetString("WECHAT_PAY_API_KEY"),
			CallbackURL: viper.GetString("WECHAT_PAY_CALLBACK_URL"),
		},
		Qiniu: Qiniu{
			AccessKey: viper.GetString("QINIU_ACCESS_KEY"),
			SecretKey: viper.GetString("QINIU_SECRET_KEY"),
			Bucket:    viper.GetString("QINIU_BUCKET"),
			Domain:    viper.GetString("QINIU_DOMAIN"),
		},
	}

	return nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *AppConfig {
	return &AppConfig{
		App: App{
			Name: "fz_yyc_api",
			Host: "0.0.0.0",
			Port: 8080,
			Env: "development",
			Debug: true,
		},
		Database: Database{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "root123456",
			Name:     "fz_yyc_api",
			Charset:  "utf8mb4",
			MaxIdle:  10,
			MaxOpen:  100,
			LifeTime: 3600,
		},
		Redis: Redis{
			Host:     "localhost",
			Port:     6379,
			Password: "redis123",
			DB:       0,
		},
		JWT: JWT{
			Secret: "default-secret-change-in-production",
			Expire: 720,
		},
	}
}

// GetDSN 获取数据库连接字符串
func (d *Database) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name, d.Charset)
}

// GetAddr 获取应用地址
func (a *App) GetAddr() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
