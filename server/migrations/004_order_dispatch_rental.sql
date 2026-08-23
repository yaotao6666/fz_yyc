-- ============================================================
-- 004_order_dispatch_rental.sql — 订单派单 + 租赁到期提醒 + 续租
-- ------------------------------------------------------------
-- 说明：
--  1. orders 新增租赁到期/续租关联列（幂等）
--  2. RBAC：订单管理下新增派单/续租权限，新增「租赁到期提醒」菜单
-- ============================================================

-- ============================================================
-- 1. orders 扩展列
-- ============================================================

-- rental_end_at: 租赁到期时间（支付成功时=paid_at+最长租赁时长）
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='rental_end_at');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN rental_end_at DATETIME DEFAULT NULL COMMENT '租赁到期时间(支付时=paid_at+租赁时长)' AFTER rental_return_remark");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- parent_order_id: 续租关联原订单ID
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='parent_order_id');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN parent_order_id BIGINT UNSIGNED DEFAULT NULL COMMENT '续租关联原订单ID(0/空=普通订单)' AFTER rental_end_at");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- renew_flag: 是否续租单
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='renew_flag');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN renew_flag TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否续租单:1=续租 0=非' AFTER parent_order_id");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- rental_end_at 加速到期提醒查询的索引
SET @_idx = (SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_rental_end_at');
SET @_sql = IF(@_idx > 0, 'SELECT 1',
  "ALTER TABLE orders ADD INDEX idx_orders_rental_end_at (rental_end_at)");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. RBAC：订单管理下新增派单/续租按钮 + 租赁到期提醒菜单
-- ============================================================
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(93, 2, 1, '租赁到期提醒', '/orders/rental-due', NULL, 4, 1, 1, 'orderrental:view'),
(930, 745, 2, '派单', NULL, NULL, 4, 1, 1, 'order:dispatch'),
(931, 745, 2, '续租', NULL, NULL, 5, 1, 1, 'order:renew'),
(932, 93, 2, '续租', NULL, NULL, 1, 1, 1, 'orderrental:renew'),
(933, 93, 2, '归还', NULL, NULL, 2, 1, 1, 'orderrental:return');

-- 绑定超级管理员(role_id=1)
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus WHERE id IN (93, 930, 931, 932, 933);