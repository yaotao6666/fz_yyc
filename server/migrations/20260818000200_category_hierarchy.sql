-- ============================================================
-- 商品分类支持三级分类：categories 增加 parent_id / level / 索引
-- 层级语义：level=1 顶级(无父)、level=2 二级、level=3 三级，不允许超过三级
-- ============================================================

-- 1. 新增 parent_id（父分类ID，空=一级）
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND COLUMN_NAME='parent_id');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE categories ADD COLUMN parent_id BIGINT UNSIGNED NULL COMMENT '父分类ID(空=一级)' AFTER name");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2. 新增 level（层级）
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND COLUMN_NAME='level');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE categories ADD COLUMN level TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '层级: 1=一级 2=二级 3=三级' AFTER parent_id");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 3. 存量数据回归：所有现有分类保持为一级（不自动构建历史父子关系）
UPDATE categories SET parent_id = NULL, level = 1 WHERE level IS NULL OR level = 0 OR parent_id IS NOT NULL;

-- 4. 新增 parent_id 索引
SET @_idx = (SELECT INDEX_NAME FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories' AND INDEX_NAME='idx_categories_parent');
SET @_sql = IF(@_idx IS NOT NULL, 'SELECT 1', 'CREATE INDEX idx_categories_parent ON categories (parent_id)');
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;