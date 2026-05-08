package database

import (
	"fmt"
	"log"
	"time"

	"fz_yyc_api/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.Database) error {
	// 配置日志
	newLogger := logger.Default.LogMode(logger.Info)

	// 连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.LifeTime) * time.Second)

	log.Println("数据库连接成功")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
