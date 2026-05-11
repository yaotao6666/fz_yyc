package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
)

// Config 全局配置
var Config *AppConfig

// AppConfig 应用配置
type AppConfig struct {
	App       App       // 应用配置
	Database  Database  // 数据库配置
	Redis     Redis     // Redis配置
	JWT       JWT       // JWT配置
	WechatPay WechatPay // 微信支付配置
	Wechat    Wechat    // 微信小程序配置
	Qiniu     Qiniu     // 七牛云配置
}

// App 应用配置
type App struct {
	Name  string
	Host  string
	Port  int
	Env   string
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
	UploadURL string
}

// Wechat 微信小程序配置
type Wechat struct {
	AppID     string // 小程序AppID
	AppSecret string // 小程序AppSecret
}

// InitConfig 初始化配置
func InitConfig() error {
	// 先设置默认配置
	Config = getDefaultConfig()

	// 尝试从 .env 文件读取
	workDir, err := os.Getwd()
	if err == nil {
		configPath := filepath.Join(workDir, ".env")
		if _, err := os.Stat(configPath); err == nil {
			viper.SetConfigFile(configPath)
			viper.SetConfigType("env")

			if err := viper.ReadInConfig(); err == nil {
				// 从 .env 文件覆盖配置
				mergeConfig()
				return nil
			}
		}
	}

	// 尝试从环境变量读取（优先级最高）
	mergeEnvConfig()

	log.Println("配置初始化完成")
	return nil
}

// mergeConfig 从 .env 文件合并配置
func mergeConfig() {
	if viper.IsSet("APP_NAME") {
		Config.App.Name = viper.GetString("APP_NAME")
	}
	if viper.IsSet("APP_HOST") {
		Config.App.Host = viper.GetString("APP_HOST")
	}
	if viper.IsSet("APP_PORT") {
		Config.App.Port = viper.GetInt("APP_PORT")
	}
	if viper.IsSet("APP_ENV") {
		Config.App.Env = viper.GetString("APP_ENV")
	}
	if viper.IsSet("APP_DEBUG") {
		Config.App.Debug = viper.GetBool("APP_DEBUG")
	}

	if viper.IsSet("DB_HOST") {
		Config.Database.Host = viper.GetString("DB_HOST")
	}
	if viper.IsSet("DB_PORT") {
		Config.Database.Port = viper.GetInt("DB_PORT")
	}
	if viper.IsSet("DB_USER") {
		Config.Database.User = viper.GetString("DB_USER")
	}
	if viper.IsSet("DB_PASSWORD") {
		Config.Database.Password = viper.GetString("DB_PASSWORD")
	}
	if viper.IsSet("DB_NAME") {
		Config.Database.Name = viper.GetString("DB_NAME")
	}
	if viper.IsSet("DB_CHARSET") {
		Config.Database.Charset = viper.GetString("DB_CHARSET")
	}
	if viper.IsSet("DB_MAX_IDLE") {
		Config.Database.MaxIdle = viper.GetInt("DB_MAX_IDLE")
	}
	if viper.IsSet("DB_MAX_OPEN") {
		Config.Database.MaxOpen = viper.GetInt("DB_MAX_OPEN")
	}
	if viper.IsSet("DB_LIFE_TIME") {
		Config.Database.LifeTime = viper.GetInt("DB_LIFE_TIME")
	}

	if viper.IsSet("REDIS_HOST") {
		Config.Redis.Host = viper.GetString("REDIS_HOST")
	}
	if viper.IsSet("REDIS_PORT") {
		Config.Redis.Port = viper.GetInt("REDIS_PORT")
	}
	if viper.IsSet("REDIS_PASSWORD") {
		Config.Redis.Password = viper.GetString("REDIS_PASSWORD")
	}
	if viper.IsSet("REDIS_DB") {
		Config.Redis.DB = viper.GetInt("REDIS_DB")
	}

	if viper.IsSet("JWT_SECRET") {
		Config.JWT.Secret = viper.GetString("JWT_SECRET")
	}
	if viper.IsSet("JWT_EXPIRE") {
		Config.JWT.Expire = viper.GetInt("JWT_EXPIRE")
	}

	if viper.IsSet("WECHAT_PAY_MCH_ID") {
		Config.WechatPay.MchID = viper.GetString("WECHAT_PAY_MCH_ID")
	}
	if viper.IsSet("WECHAT_PAY_API_KEY") {
		Config.WechatPay.APIKey = viper.GetString("WECHAT_PAY_API_KEY")
	}
	if viper.IsSet("WECHAT_PAY_CALLBACK_URL") {
		Config.WechatPay.CallbackURL = viper.GetString("WECHAT_PAY_CALLBACK_URL")
	}

	if viper.IsSet("QINIU_ACCESS_KEY") {
		Config.Qiniu.AccessKey = viper.GetString("QINIU_ACCESS_KEY")
	}
	if viper.IsSet("QINIU_SECRET_KEY") {
		Config.Qiniu.SecretKey = viper.GetString("QINIU_SECRET_KEY")
	}
	if viper.IsSet("QINIU_BUCKET") {
		Config.Qiniu.Bucket = viper.GetString("QINIU_BUCKET")
	}
	if viper.IsSet("QINIU_DOMAIN") {
		Config.Qiniu.Domain = viper.GetString("QINIU_DOMAIN")
	}
	if viper.IsSet("QINIU_UPLOAD_URL") {
		Config.Qiniu.UploadURL = viper.GetString("QINIU_UPLOAD_URL")
	}
}

// mergeEnvConfig 从环境变量合并配置（优先级最高）
func mergeEnvConfig() {
	if env := os.Getenv("APP_HOST"); env != "" {
		Config.App.Host = env
	}
	if env := os.Getenv("APP_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			Config.App.Port = port
		}
	}
	if env := os.Getenv("APP_ENV"); env != "" {
		Config.App.Env = env
	}
	if env := os.Getenv("APP_DEBUG"); env != "" {
		Config.App.Debug = env == "true" || env == "1"
	}

	if env := os.Getenv("DB_HOST"); env != "" {
		Config.Database.Host = env
	}
	if env := os.Getenv("DB_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			Config.Database.Port = port
		}
	}
	if env := os.Getenv("DB_USER"); env != "" {
		Config.Database.User = env
	}
	if env := os.Getenv("DB_PASSWORD"); env != "" {
		Config.Database.Password = env
	}
	if env := os.Getenv("DB_NAME"); env != "" {
		Config.Database.Name = env
	}
	if env := os.Getenv("DB_CHARSET"); env != "" {
		Config.Database.Charset = env
	}
	if env := os.Getenv("DB_MAX_IDLE"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			Config.Database.MaxIdle = val
		}
	}
	if env := os.Getenv("DB_MAX_OPEN"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			Config.Database.MaxOpen = val
		}
	}
	if env := os.Getenv("DB_LIFE_TIME"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			Config.Database.LifeTime = val
		}
	}

	if env := os.Getenv("REDIS_HOST"); env != "" {
		Config.Redis.Host = env
	}
	if env := os.Getenv("REDIS_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			Config.Redis.Port = port
		}
	}
	if env := os.Getenv("REDIS_PASSWORD"); env != "" {
		Config.Redis.Password = env
	}
	if env := os.Getenv("REDIS_DB"); env != "" {
		if db, err := strconv.Atoi(env); err == nil {
			Config.Redis.DB = db
		}
	}

	if env := os.Getenv("JWT_SECRET"); env != "" {
		Config.JWT.Secret = env
	}
	if env := os.Getenv("JWT_EXPIRE"); env != "" {
		if expire, err := strconv.Atoi(env); err == nil {
			Config.JWT.Expire = expire
		}
	}

	if env := os.Getenv("WECHAT_APP_ID"); env != "" {
		Config.Wechat.AppID = env
	}
	if env := os.Getenv("WECHAT_APP_SECRET"); env != "" {
		Config.Wechat.AppSecret = env
	}
	if env := os.Getenv("QINIU_ACCESS_KEY"); env != "" {
		Config.Qiniu.AccessKey = env
	}
	if env := os.Getenv("QINIU_SECRET_KEY"); env != "" {
		Config.Qiniu.SecretKey = env
	}
	if env := os.Getenv("QINIU_BUCKET"); env != "" {
		Config.Qiniu.Bucket = env
	}
	if env := os.Getenv("QINIU_DOMAIN"); env != "" {
		Config.Qiniu.Domain = env
	}
	if env := os.Getenv("QINIU_UPLOAD_URL"); env != "" {
		Config.Qiniu.UploadURL = env
	}
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *AppConfig {
	return &AppConfig{
		App: App{
			Name:  "fz_yyc_api",
			Host:  "0.0.0.0",
			Port:  8080,
			Env:   "development",
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
		Wechat: Wechat{
			AppID:     "",
			AppSecret: "",
		},
		Qiniu: Qiniu{
			AccessKey: "",
			SecretKey: "",
			Bucket:    "",
			Domain:    "",
			UploadURL: "",
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
