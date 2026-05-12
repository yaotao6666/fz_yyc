package database

import (
	"fmt"
	"log"
	"time"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/models"

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

	if ensureErr := ensureMerchantStaffColumns(DB); ensureErr != nil {
		return fmt.Errorf("初始化商家员工扩展字段失败: %w", ensureErr)
	}

	if ensureErr := ensureUserBehaviorEventsTable(DB); ensureErr != nil {
		return fmt.Errorf("初始化用户行为事件表失败: %w", ensureErr)
	}

	if ensureErr := ensureUserTables(DB); ensureErr != nil {
		return fmt.Errorf("初始化用户表扩展字段失败: %w", ensureErr)
	}

	if ensureErr := ensureOrderColumns(DB); ensureErr != nil {
		return fmt.Errorf("初始化订单扩展字段失败: %w", ensureErr)
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

func ensureMerchantStaffColumns(db *gorm.DB) error {
	if err := ensureMerchantStaffColumn(
		db,
		"openid",
		"ADD COLUMN openid VARCHAR(64) DEFAULT NULL COMMENT '微信OpenID' AFTER phone",
	); err != nil {
		return err
	}

	if err := ensureMerchantStaffColumn(
		db,
		"wechat_bound_at",
		"ADD COLUMN wechat_bound_at DATETIME DEFAULT NULL COMMENT '微信绑定时间' AFTER openid",
	); err != nil {
		return err
	}

	if err := ensureMerchantStaffColumn(
		db,
		"unionid",
		"ADD COLUMN unionid VARCHAR(64) DEFAULT NULL COMMENT '微信UnionID' AFTER openid",
	); err != nil {
		return err
	}

	if err := ensureMerchantStaffColumn(
		db,
		"notify_enabled",
		"ADD COLUMN notify_enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '订单提示音开关' AFTER role",
	); err != nil {
		return err
	}

	if err := ensureMerchantStaffColumn(
		db,
		"browse_notify_enabled",
		"ADD COLUMN browse_notify_enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '顾客浏览提示音开关' AFTER notify_enabled",
	); err != nil {
		return err
	}

	if err := ensureMerchantStaffColumn(
		db,
		"last_wechat_login_at",
		"ADD COLUMN last_wechat_login_at DATETIME DEFAULT NULL COMMENT '最后一次微信快捷登录时间' AFTER last_login_at",
	); err != nil {
		return err
	}

	return ensureMerchantStaffIndex(
		db,
		"idx_merchant_staffs_openid",
		"CREATE INDEX idx_merchant_staffs_openid ON merchant_staffs (openid)",
	)
}

func ensureMerchantStaffColumn(db *gorm.DB, columnName string, addSQL string) error {
	var count int64
	queryErr := db.Raw(`
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'merchant_staffs'
		  AND COLUMN_NAME = ?
	`, columnName).Scan(&count).Error
	if queryErr != nil {
		return queryErr
	}

	if count > 0 {
		return nil
	}

	return db.Exec("ALTER TABLE merchant_staffs " + addSQL).Error
}

func ensureMerchantStaffIndex(db *gorm.DB, indexName string, createSQL string) error {
	var count int64
	queryErr := db.Raw(`
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'merchant_staffs'
		  AND INDEX_NAME = ?
	`, indexName).Scan(&count).Error
	if queryErr != nil {
		return queryErr
	}

	if count > 0 {
		return nil
	}

	return db.Exec(createSQL).Error
}

func ensureUserBehaviorEventsTable(db *gorm.DB) error {
	return db.AutoMigrate(&models.UserBehaviorEvent{})
}

func ensureUserTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.UserVisit{},
	)
}

func ensureOrderColumns(db *gorm.DB) error {
	var count int64
	queryErr := db.Raw(`
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'orders'
		  AND COLUMN_NAME = 'completed_by_name'
	`).Scan(&count).Error
	if queryErr != nil {
		return queryErr
	}

	if count > 0 {
		return nil
	}

	return db.Exec(
		"ALTER TABLE orders ADD COLUMN completed_by_name VARCHAR(64) DEFAULT NULL COMMENT '核销人' AFTER completed_at",
	).Error
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
