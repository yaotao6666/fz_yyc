-- ============================================================
-- 第0期：CXSM 清理外卖残留 + 业务字段扩展迁移
-- 约束：纯 SQL 单条执行，禁止使用存储过程 / BEGIN...END / WHILE / IF...THEN
--       仅允许使用 SET @var = ... + PREPARE stmt FROM @sql + EXECUTE stmt 模式
--       多列/多索引删除必须展开为多条语句，不允许循环
-- ============================================================

-- ------------------------------------------------------------
-- Step 1. 删除8张废弃表（DRP-TABLE 规则）
-- ------------------------------------------------------------
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS merchant_pickup_points;
DROP TABLE IF EXISTS cloud_printers;
DROP TABLE IF EXISTS print_logs;
DROP TABLE IF EXISTS merchant_full_reduction_rules;
DROP TABLE IF EXISTS coupons;
DROP TABLE IF EXISTS coupon_records;
DROP TABLE IF EXISTS service_provider_sps;
DROP TABLE IF EXISTS service_providers;
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- Step 2. merchants 表：删 SP 外键 / 删 SP 索引 / 删除残留外卖字段
-- ============================================================

-- 2a. 若存在指向 service_providers 的外键则删除
SET @_fk_name = (
  SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'merchants'
    AND REFERENCED_TABLE_NAME = 'service_providers'
  LIMIT 1
);
SET @_sql = IF(@_fk_name IS NULL, 'SELECT 1',
               CONCAT('ALTER TABLE merchants DROP FOREIGN KEY ', @_fk_name));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2b. 删除 SP 相关索引（逐一尝试 3 个已知索引名）
SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND INDEX_NAME='idx_merchants_service_provider_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP INDEX idx_merchants_service_provider_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND INDEX_NAME='idx_merchants_sp_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP INDEX idx_merchants_sp_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND INDEX_NAME='service_provider_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP INDEX service_provider_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2c. 删除 merchants 残留列（逐一尝试 7 列）
--     service_provider_id / takeout_enabled / dine_in_enabled / pickup_enabled / pickup_enabled_2 / min_order_amount / qrcode_url / qr_code_url
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='service_provider_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP COLUMN service_provider_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='takeout_enabled');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP COLUMN takeout_enabled');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='dine_in_enabled');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP COLUMN dine_in_enabled');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='pickup_enabled');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP COLUMN pickup_enabled');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='pickup_enabled_2');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP COLUMN pickup_enabled_2');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='min_order_amount');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP COLUMN min_order_amount');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='qrcode_url');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP COLUMN qrcode_url');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchants' AND COLUMN_NAME='qr_code_url');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchants DROP COLUMN qr_code_url');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- Step 3. categories 表：新增 category_type
-- ============================================================
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND COLUMN_NAME='category_type');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE categories ADD COLUMN category_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '分类类型: 1=商品分类 2=服务分类' AFTER name");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- Step 4. products 表：新增 product_type + service_content JSON + 索引
-- ============================================================
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND COLUMN_NAME='product_type');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE products ADD COLUMN product_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '商品类型: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务 5=科普活动 6=长护险服务' AFTER unit");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND COLUMN_NAME='service_content');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE products ADD COLUMN service_content JSON NULL COMMENT '服务型商品的内容描述JSON（康养套餐项/陪诊项目/活动详情）' AFTER product_type");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND INDEX_NAME='idx_products_product_type');
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1',
  'CREATE INDEX idx_products_product_type ON products (merchant_id, product_type, status)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- Step 5. orders 表：删外卖字段 + 新增业务字段 + 索引
-- ============================================================
-- 5a. 删除外卖相关列（逐一尝试 9 列）
SET @_cols_to_drop = 'delivery_type|delivery_distance|pickup_point_id|pickup_point_name|pickup_point_address|pickup_point_lat|pickup_point_lng|verify_code|completed_by_name';

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='delivery_type');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN delivery_type');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='delivery_distance');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN delivery_distance');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='pickup_point_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN pickup_point_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='pickup_point_name');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN pickup_point_name');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='pickup_point_address');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN pickup_point_address');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='pickup_point_lat');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN pickup_point_lat');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='pickup_point_lng');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN pickup_point_lng');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='verify_code');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN verify_code');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='completed_by_name');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN completed_by_name');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 5b. 新增 order_type / biz_status / scheduled_at / assigned_staff_id / actual_started_at / actual_ended_at
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='order_type');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN order_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '订单类型: 1=零售 2=租赁 3=康养上门 4=陪诊 5=科普体验 6=长护险服务' AFTER merchant_id");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='biz_status');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN biz_status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '业务子状态(按order_type语义不同)' AFTER status");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='scheduled_at');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN scheduled_at DATETIME(3) NULL COMMENT '预约服务开始时间(康养/陪诊/科普)' AFTER refunded_at");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='assigned_staff_id');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN assigned_staff_id BIGINT UNSIGNED NULL COMMENT '指派的服务人员ID(service_staffs.id)' AFTER scheduled_at");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='actual_started_at');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN actual_started_at DATETIME(3) NULL COMMENT '实际服务开始时间(签到时间)' AFTER assigned_staff_id");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='actual_ended_at');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE orders ADD COLUMN actual_ended_at DATETIME(3) NULL COMMENT '实际服务结束时间(签退时间)' AFTER actual_started_at");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 5c. 新增 3 个索引
SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_order_type');
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1',
  'CREATE INDEX idx_orders_order_type ON orders (merchant_id, order_type, status)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_scheduled_at');
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1',
  'CREATE INDEX idx_orders_scheduled_at ON orders (merchant_id, scheduled_at)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_assigned_staff');
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1',
  'CREATE INDEX idx_orders_assigned_staff ON orders (assigned_staff_id, status)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- Step 6. order_items 表：补索引
-- ============================================================
SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='order_items' AND INDEX_NAME='idx_order_items_order_type_ref');
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1',
  'CREATE INDEX idx_order_items_order_type_ref ON order_items (order_id, product_id)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- Step 7. merchant_delivery_settings：保留不删
-- ============================================================

-- ============================================================
-- Step 8. 4 个辅助表：删残留 service_provider_id 列
--         announcements / activities / merchant_fees / merchant_profit_sharing_records
-- ============================================================

-- 8a. announcements: 先删 SP FK 再删列
SET @_fk_name = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='announcements'
    AND REFERENCED_TABLE_NAME='service_providers' LIMIT 1);
SET @_sql = IF(@_fk_name IS NULL, 'SELECT 1', CONCAT('ALTER TABLE announcements DROP FOREIGN KEY ', @_fk_name));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='announcements' AND COLUMN_NAME='service_provider_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE announcements DROP COLUMN service_provider_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 8b. activities
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='activities' AND COLUMN_NAME='service_provider_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE activities DROP COLUMN service_provider_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 8c. merchant_fees
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_fees' AND COLUMN_NAME='service_provider_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchant_fees DROP COLUMN service_provider_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 8d. merchant_profit_sharing_records：先删 SP FK（若存在），再删列
SET @_fk_name = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_profit_sharing_records'
    AND REFERENCED_TABLE_NAME='service_providers' LIMIT 1);
SET @_sql = IF(@_fk_name IS NULL, 'SELECT 1', CONCAT('ALTER TABLE merchant_profit_sharing_records DROP FOREIGN KEY ', @_fk_name));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_profit_sharing_records' AND COLUMN_NAME='service_provider_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchant_profit_sharing_records DROP COLUMN service_provider_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
