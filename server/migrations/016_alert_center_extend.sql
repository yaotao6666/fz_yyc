-- ============================================================
-- 016_alert_center_extend.sql — 预警中心扩展（订单级预警）
-- ------------------------------------------------------------
-- 说明：
--  1. service_alert_events.staff_id 改为可空（订单级预警可空）+ 新增 summary 列
--  2. orders 增列 assigned_at（指派/接单时间，超时未签到预警依据）
--  3. 新建单行配置表 alert_settings（预警阈值总开关）
--  4. RBAC：预警中心(id=97)下新增「预警设置」按钮
-- ============================================================

SET NAMES utf8mb4;

-- ============================================================
-- 1. service_alert_events.staff_id 改为可空（幂等：仅当当前为 NOT NULL 时执行）
-- ============================================================
SET @_col = (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_alert_events'
    AND COLUMN_NAME='staff_id' AND IS_NULLABLE='NO');
SET @_sql = IF(@_col > 0,
  "ALTER TABLE service_alert_events MODIFY staff_id BIGINT UNSIGNED DEFAULT NULL COMMENT '服务人员ID(订单级预警可空)'",
  'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. service_alert_events 新增 summary 列（address 之后）
-- ============================================================
SET @_col = (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_alert_events' AND COLUMN_NAME='summary');
SET @_sql = IF(@_col = 0,
  "ALTER TABLE service_alert_events ADD COLUMN summary VARCHAR(256) DEFAULT NULL COMMENT '预警摘要(订单级预警触发说明)' AFTER address",
  'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 3. orders 新增 assigned_at 列（指派/接单时间）
-- ============================================================
SET @_col = (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='assigned_at');
SET @_sql = IF(@_col = 0,
  "ALTER TABLE orders ADD COLUMN assigned_at DATETIME DEFAULT NULL COMMENT '指派/接单时间(超时未签到预警依据)'",
  'SELECT 1');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 4. 单行配置表 alert_settings（预警阈值总开关）
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='alert_settings');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE alert_settings (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID',
  enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '预警总开关',
  goods_unverified_hours INT NOT NULL DEFAULT 24 COMMENT '实物超时未核销(小时)',
  service_unassigned_hours INT NOT NULL DEFAULT 2 COMMENT '服务超时未指派(小时)',
  escort_unfinished_minutes INT NOT NULL DEFAULT 120 COMMENT '陪诊超时未完成(分钟)',
  service_unstarted_minutes INT NOT NULL DEFAULT 30 COMMENT '指派超时未签到(分钟)',
  rental_overdue_hours INT NOT NULL DEFAULT 24 COMMENT '租赁逾期未归还(小时)',
  refund_stuck_hours INT NOT NULL DEFAULT 24 COMMENT '退款卡在处理中(小时)',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='预警配置表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 保证默认配置行存在
INSERT IGNORE INTO alert_settings (id) VALUES (1);

-- ============================================================
-- 5. RBAC：预警中心(id=97)下新增「预警设置」按钮
-- ============================================================
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(972, 97, 2, '预警设置', NULL, NULL, 2, 1, 1, 'alert-settings:update');

-- 授予超管角色（role_id=1）新菜单权限
INSERT IGNORE INTO sys_role_menus (role_id, menu_id)
SELECT 1, id FROM sys_menus WHERE id = 972;
