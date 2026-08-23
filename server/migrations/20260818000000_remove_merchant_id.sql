-- ============================================================
-- 深度重构：去除全链路 merchant_id
-- 约束：纯 SQL 单条执行，禁止存储过程 / BEGIN...END / WHILE / IF...THEN
--       仅允许 SET @var + PREPARE + EXECUTE 模式；先删外键 → 删索引 → 删列
-- merchants 表保留为全局单例配置（id=1），其余表删除 merchant_id 关联
-- ============================================================

-- ============================================================
-- 1. categories：删外键 / 索引 / merchant_id 列
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE categories DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND INDEX_NAME='idx_categories_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE categories DROP INDEX idx_categories_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND INDEX_NAME='idx_categories_sort' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE categories DROP INDEX idx_categories_sort');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE categories DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND INDEX_NAME='idx_categories_sort' LIMIT 1);
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1', 'CREATE INDEX idx_categories_sort ON categories (sort)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. products：删外键 / 索引 / merchant_id 列，重建单列索引
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE products DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND INDEX_NAME='idx_products_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE products DROP INDEX idx_products_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND INDEX_NAME='idx_products_sales' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE products DROP INDEX idx_products_sales');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND INDEX_NAME='idx_products_sort' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE products DROP INDEX idx_products_sort');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND INDEX_NAME='idx_products_product_type' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE products DROP INDEX idx_products_product_type');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE products DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND INDEX_NAME='idx_products_sales' LIMIT 1);
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1', 'CREATE INDEX idx_products_sales ON products (sales)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND INDEX_NAME='idx_products_sort' LIMIT 1);
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1', 'CREATE INDEX idx_products_sort ON products (sort)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='products' AND INDEX_NAME='idx_products_product_type' LIMIT 1);
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1', 'CREATE INDEX idx_products_product_type ON products (product_type, status)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 3. orders：删外键 / 索引 / merchant_id 列，重建组合索引
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE orders DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP INDEX idx_orders_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_order_type' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP INDEX idx_orders_order_type');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_scheduled_at' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP INDEX idx_orders_scheduled_at');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE orders DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_order_type' LIMIT 1);
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1', 'CREATE INDEX idx_orders_order_type ON orders (order_type, status)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='orders' AND INDEX_NAME='idx_orders_scheduled_at' LIMIT 1);
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1', 'CREATE INDEX idx_orders_scheduled_at ON orders (scheduled_at)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 4. order_items：删外键 / 索引 / merchant_id 列
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='order_items'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE order_items DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='order_items' AND INDEX_NAME='idx_order_items_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE order_items DROP INDEX idx_order_items_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='order_items' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE order_items DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 5. user_visits：删外键 / 索引 / merchant_id 列
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='user_visits'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE user_visits DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='user_visits' AND INDEX_NAME='idx_user_visits_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE user_visits DROP INDEX idx_user_visits_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='user_visits' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE user_visits DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 6. user_behavior_events：删索引 / merchant_id 列
-- ============================================================
SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='user_behavior_events' AND INDEX_NAME='idx_user_behavior_events_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE user_behavior_events DROP INDEX idx_user_behavior_events_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='user_behavior_events' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE user_behavior_events DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 7. merchant_delivery_settings：删外键 / 唯一索引 / merchant_id 列
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_delivery_settings'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE merchant_delivery_settings DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_delivery_settings' AND INDEX_NAME='uk_merchant_delivery_settings_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE merchant_delivery_settings DROP INDEX uk_merchant_delivery_settings_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_delivery_settings' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchant_delivery_settings DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 8. merchant_fees：删外键 / 索引 / merchant_id 列
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_fees'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE merchant_fees DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_fees' AND INDEX_NAME='idx_merchant_fees_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE merchant_fees DROP INDEX idx_merchant_fees_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_fees' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchant_fees DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 9. merchant_rates：删外键 / 索引 / merchant_id 列
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_rates'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE merchant_rates DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_rates' AND INDEX_NAME='idx_merchant_rates_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE merchant_rates DROP INDEX idx_merchant_rates_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_rates' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchant_rates DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 10. merchant_staffs：删外键 / 唯一索引 / merchant_id 列，username 重建唯一索引
-- ============================================================
SET @_fk = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_staffs'
    AND REFERENCED_TABLE_NAME='merchants' LIMIT 1);
SET @_sql = IF(@_fk IS NULL, 'SELECT 1', CONCAT('ALTER TABLE merchant_staffs DROP FOREIGN KEY ', @_fk));
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_staffs' AND INDEX_NAME='uk_merchant_staffs_merchant_username' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE merchant_staffs DROP INDEX uk_merchant_staffs_merchant_username');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_staffs' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE merchant_staffs DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='merchant_staffs' AND INDEX_NAME='uk_merchant_staffs_username' LIMIT 1);
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1', 'CREATE UNIQUE INDEX uk_merchant_staffs_username ON merchant_staffs (username)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 11. service_staffs（GORM 自动建表）：删索引 / merchant_id 列
-- ============================================================
SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_staffs' AND INDEX_NAME='idx_service_staffs_merchant_id' LIMIT 1);
SET @_sql = IF(@_idx IS NULL, 'SELECT 1', 'ALTER TABLE service_staffs DROP INDEX idx_service_staffs_merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_staffs' AND COLUMN_NAME='merchant_id');
SET @_sql = IF(@_col IS NULL, 'SELECT 1', 'ALTER TABLE service_staffs DROP COLUMN merchant_id');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
