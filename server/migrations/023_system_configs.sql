-- ============================================================
-- 023_system_configs.sql — 通用配置表 system_configs（重构 alert_settings）
-- ------------------------------------------------------------
-- 说明：
--  1. 新建通用键值配置表 system_configs（config_key + JSON config_value + remark）
--  2. 写入初始化默认配置（INSERT IGNORE，保证全新库也有配置行）
--  3. 若存在旧单行表 alert_settings，将其 8 个阈值列迁移为 key-value 配置项并覆盖默认值
--  4. 迁移完成后 DROP 旧表 alert_settings（不再保留旧列/旧表）
-- 不改动 REST 接口契约（GET/PUT /alert-settings 仍返回兼容的 typed JSON）。
-- ============================================================

SET NAMES utf8mb4;

-- ============================================================
-- 1. 建表 system_configs（config_key 唯一）
-- ============================================================
CREATE TABLE IF NOT EXISTS system_configs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID',
  config_key VARCHAR(100) NOT NULL COMMENT '配置键',
  config_value TEXT NOT NULL COMMENT '配置值(JSON)',
  remark VARCHAR(255) DEFAULT NULL COMMENT '备注',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通用系统配置表(key-value+JSON+备注)';

-- ============================================================
-- 2. 初始化默认配置（幂等，仅补齐缺失 key）
-- ============================================================
INSERT IGNORE INTO system_configs (config_key, config_value, remark) VALUES
('alert.enabled',                 'true',   '预警总开关'),
('alert.goods_unverified_hours',    '24',   '实物超时未核销(小时)'),
('alert.service_unassigned_hours', '2',    '服务超时未指派(小时)'),
('alert.escort_unfinished_minutes','120',  '陪诊超时未完成(分钟)'),
('alert.service_unstarted_minutes','30',   '指派超时未签到(分钟)'),
('alert.rental_overdue_hours',      '24',  '租赁逾期未归还(小时)'),
('alert.refund_stuck_hours',        '24',  '退款卡在处理中(小时)'),
('service.audio_retain_days',       '30',  '服务录音保留天数(天)');

-- ============================================================
-- 3. 迁移旧表 alert_settings 的值（存在则覆盖默认值）
-- ============================================================
SET @_legacy = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='alert_settings');

SET @_sql = IF(@_legacy > 0,
"INSERT INTO system_configs (config_key, config_value, remark, created_at, updated_at)
 SELECT 'alert.enabled',                 IF(enabled=1,'true','false'),  '预警总开关', created_at, updated_at FROM alert_settings
 ON DUPLICATE KEY UPDATE config_value=VALUES(config_value), remark=VALUES(remark)",
'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_sql = IF(@_legacy > 0,
"INSERT INTO system_configs (config_key, config_value, remark, created_at, updated_at)
 SELECT 'alert.goods_unverified_hours',    CAST(goods_unverified_hours AS CHAR),    '实物超时未核销(小时)', created_at, updated_at FROM alert_settings
 ON DUPLICATE KEY UPDATE config_value=VALUES(config_value)",
'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_sql = IF(@_legacy > 0,
"INSERT INTO system_configs (config_key, config_value, remark, created_at, updated_at)
 SELECT 'alert.service_unassigned_hours',  CAST(service_unassigned_hours AS CHAR),  '服务超时未指派(小时)', created_at, updated_at FROM alert_settings
 ON DUPLICATE KEY UPDATE config_value=VALUES(config_value)",
'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_sql = IF(@_legacy > 0,
"INSERT INTO system_configs (config_key, config_value, remark, created_at, updated_at)
 SELECT 'alert.escort_unfinished_minutes', CAST(escort_unfinished_minutes AS CHAR), '陪诊超时未完成(分钟)', created_at, updated_at FROM alert_settings
 ON DUPLICATE KEY UPDATE config_value=VALUES(config_value)",
'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_sql = IF(@_legacy > 0,
"INSERT INTO system_configs (config_key, config_value, remark, created_at, updated_at)
 SELECT 'alert.service_unstarted_minutes', CAST(service_unstarted_minutes AS CHAR), '指派超时未签到(分钟)', created_at, updated_at FROM alert_settings
 ON DUPLICATE KEY UPDATE config_value=VALUES(config_value)",
'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_sql = IF(@_legacy > 0,
"INSERT INTO system_configs (config_key, config_value, remark, created_at, updated_at)
 SELECT 'alert.rental_overdue_hours',      CAST(rental_overdue_hours AS CHAR),      '租赁逾期未归还(小时)', created_at, updated_at FROM alert_settings
 ON DUPLICATE KEY UPDATE config_value=VALUES(config_value)",
'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_sql = IF(@_legacy > 0,
"INSERT INTO system_configs (config_key, config_value, remark, created_at, updated_at)
 SELECT 'alert.refund_stuck_hours',        CAST(refund_stuck_hours AS CHAR),        '退款卡在处理中(小时)', created_at, updated_at FROM alert_settings
 ON DUPLICATE KEY UPDATE config_value=VALUES(config_value)",
'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 旧表可能缺 service_audio_retain_days 列（早期 schema），仅在列存在时迁移
SET @_audio_col = IF(@_legacy > 0, (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='alert_settings' AND COLUMN_NAME='service_audio_retain_days'), 0);
SET @_sql = IF(@_legacy > 0 AND @_audio_col > 0,
"INSERT INTO system_configs (config_key, config_value, remark, created_at, updated_at)
 SELECT 'service.audio_retain_days',       CAST(service_audio_retain_days AS CHAR), '服务录音保留天数(天)', created_at, updated_at FROM alert_settings
 ON DUPLICATE KEY UPDATE config_value=VALUES(config_value)",
'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 4. 迁移完成后 DROP 旧表（不保留旧列/旧表）
-- ============================================================
SET @_sql = IF(@_legacy > 0, 'DROP TABLE IF EXISTS alert_settings', 'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;