-- 系统一致性检测修复迁移
-- 创建时间: 2026-05-15
-- 关联检测: system-consistency-check
-- 使用存储过程安全处理已存在的列

-- ===== UP =====

DELIMITER //

-- 1. Activity表新增service_provider_id字段（SP数据隔离）
DROP PROCEDURE IF EXISTS add_column_if_not_exists//
CREATE PROCEDURE add_column_if_not_exists(
  IN table_name_param VARCHAR(100),
  IN column_name_param VARCHAR(100),
  IN column_definition VARCHAR(500)
)
BEGIN
  DECLARE col_exists INT DEFAULT 0;
  SELECT COUNT(*) INTO col_exists
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = table_name_param
    AND COLUMN_NAME = column_name_param;

  IF col_exists = 0 THEN
    SET @sql = CONCAT('ALTER TABLE `', table_name_param, '` ADD COLUMN `', column_name_param, '` ', column_definition);
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END//

CALL add_column_if_not_exists('activities', 'service_provider_id', 'BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT "服务商ID" AFTER `id`')//
UPDATE `activities` SET `service_provider_id` = 1 WHERE `service_provider_id` = 0//

-- 2. cloud_printers表补充缺失列
CALL add_column_if_not_exists('cloud_printers', 'api_url', 'VARCHAR(255) NOT NULL DEFAULT "" COMMENT "API地址" AFTER `brand`')//
CALL add_column_if_not_exists('cloud_printers', 'print_types', 'JSON COMMENT "打印类型" AFTER `api_url`')//
CALL add_column_if_not_exists('cloud_printers', 'auto_print', 'TINYINT(1) NOT NULL DEFAULT 1 COMMENT "自动打印" AFTER `print_types`')//
CALL add_column_if_not_exists('cloud_printers', 'is_default', 'TINYINT(1) NOT NULL DEFAULT 0 COMMENT "是否默认" AFTER `auto_print`')//
CALL add_column_if_not_exists('cloud_printers', 'print_count', 'INT NOT NULL DEFAULT 0 COMMENT "打印次数" AFTER `is_default`')//
CALL add_column_if_not_exists('cloud_printers', 'last_print_at', 'DATETIME DEFAULT NULL COMMENT "最后打印时间" AFTER `print_count`')//

DROP PROCEDURE IF EXISTS add_column_if_not_exists//

DELIMITER ;

-- ===== DOWN =====
-- ALTER TABLE `activities` DROP COLUMN `service_provider_id`;
-- ALTER TABLE `cloud_printers` DROP COLUMN `api_url`;
-- ALTER TABLE `cloud_printers` DROP COLUMN `print_types`;
-- ALTER TABLE `cloud_printers` DROP COLUMN `auto_print`;
-- ALTER TABLE `cloud_printers` DROP COLUMN `is_default`;
-- ALTER TABLE `cloud_printers` DROP COLUMN `print_count`;
-- ALTER TABLE `cloud_printers` DROP COLUMN `last_print_at`;
