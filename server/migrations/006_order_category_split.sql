-- ============================================================
-- 006_order_category_split.sql — 订单口径统一 + 商品/服务管理拆分
-- ------------------------------------------------------------
-- 说明：
--  1. orders 新增服务对象档案ID（服务订单必填）与收货区县列（幂等）
--  2. RBAC：商品管理目录下新增「服务管理」菜单与按钮权限节点
--  3. 订单分类口径：实物订单 order_type ∈ {1,2}，服务订单 order_type ∈ {3,4,5,6}
--     （仅约定，无数据变更；后端接口按 category 参数过滤）
-- ============================================================

SET NAMES utf8mb4;

-- ============================================================
-- 1. orders 扩展列
-- ============================================================

-- record_id: 服务订单绑定的健康档案ID（服务单下单必填，实物单为空）
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='record_id');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN record_id BIGINT UNSIGNED DEFAULT NULL COMMENT '服务订单绑定的健康档案ID(服务单必填)' AFTER assigned_staff_id");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- delivery_district: 收货区县（服务订单区域匹配用，下单时从收货地址提取）
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='delivery_district');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN delivery_district VARCHAR(32) DEFAULT NULL COMMENT '收货区县(服务订单区域匹配用)' AFTER record_id");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- record_id 索引（档案维度查服务订单）
SET @_idx = (SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_record_id');
SET @_sql = IF(@_idx > 0, 'SELECT 1',
  "ALTER TABLE orders ADD INDEX idx_orders_record_id (record_id)");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. RBAC：商品管理目录(id=3)下新增「服务管理」菜单 + 按钮权限
--    服务项目（康养套餐/陪诊服务）与实物商品（零售/租赁/资讯）拆分管理
-- ============================================================
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(94, 3, 1, '服务管理', '/service-products', NULL, 3, 1, 1, 'service-products:view'),
(940, 94, 2, '新增服务', NULL, NULL, 1, 1, 1, 'service-products:create'),
(941, 94, 2, '编辑服务', NULL, NULL, 2, 1, 1, 'service-products:update'),
(942, 94, 2, '删除服务', NULL, NULL, 3, 1, 1, 'service-products:delete');

-- 绑定超级管理员(role_id=1)
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus WHERE id IN (94, 940, 941, 942);
