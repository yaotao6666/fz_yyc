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

	if err := ensureMerchantStaffNotifyEnabledColumn(DB); err != nil {
		return fmt.Errorf("初始化商家员工提示音字段失败: %w", err)
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

func ensureMerchantStaffNotifyEnabledColumn(db *gorm.DB) error {
	var count int64
	queryErr := db.Raw(`
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'merchant_staffs'
		  AND COLUMN_NAME = 'notify_enabled'
	`).Scan(&count).Error
	if queryErr != nil {
		return queryErr
	}

	if count > 0 {
		return nil
	}

	return db.Exec(`
		ALTER TABLE merchant_staffs
		ADD COLUMN notify_enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '订单提示音开关' AFTER role
	`).Error
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
