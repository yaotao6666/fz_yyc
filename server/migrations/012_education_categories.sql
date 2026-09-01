-- ============================================================
-- 阶段五 8.3 宣教独立：分类表建 + 文章挂分类 + 存量回填 + 后台菜单
-- 编号：012（多档案为 011）
-- 执行：mysql -h <host> -u <user> -p <db> < server/migrations/012_education_categories.sql
-- ============================================================
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------------------------------------
-- 1. 建表 health_education_categories（两级分类）
-- ----------------------------------------------------------
DROP TABLE IF EXISTS `health_education_categories`;
CREATE TABLE `health_education_categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `parent_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '父分类ID，0=一级',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分类名称',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序（小在前）',
  `status` tinyint unsigned NOT NULL DEFAULT 1 COMMENT '状态:1启用0停用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hec_parent_name` (`parent_id`, `name`),
  KEY `idx_hec_parent_id` (`parent_id`),
  KEY `idx_hec_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='健康宣教分类表（阶段五 8.3，两级树形）';

-- ----------------------------------------------------------
-- 2. health_education_articles 新增 category_id 字段及索引（幂等：先查是否已加）
-- ----------------------------------------------------------
SET @col_exist := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'health_education_articles' AND COLUMN_NAME = 'category_id'
);
SET @sql_add_col := IF(@col_exist = 0,
  'ALTER TABLE `health_education_articles` ADD COLUMN `category_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''分类ID（关联health_education_categories）'' AFTER `title`',
  'SELECT 1');
PREPARE stmt FROM @sql_add_col; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx_exist := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'health_education_articles' AND INDEX_NAME = 'idx_hea_category_id'
);
SET @sql_add_idx := IF(@idx_exist = 0,
  'ALTER TABLE `health_education_articles` ADD INDEX `idx_hea_category_id` (`category_id`)',
  'SELECT 1');
PREPARE stmt FROM @sql_add_idx; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ----------------------------------------------------------
-- 3. 存量回填：按 category 字符串去重 -> 插入一级分类 -> 回填 category_id
-- ----------------------------------------------------------
-- 3.1 插入去重后的一级分类（parent_id=0）。相同(parent_id,name) 由唯一键去重。
INSERT INTO `health_education_categories` (`parent_id`, `name`, `sort`, `status`, `created_at`, `updated_at`)
SELECT
  0 AS `parent_id`,
  CASE WHEN TRIM(IFNULL(`category`, '')) = '' THEN '未分类' ELSE TRIM(`category`) END AS `name`,
  0 AS `sort`,
  1 AS `status`,
  NOW() AS `created_at`,
  NOW() AS `updated_at`
FROM `health_education_articles`
GROUP BY `name`
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- 3.2 回填 category_id（按字符串匹配）
UPDATE `health_education_articles` art
LEFT JOIN `health_education_categories` cat
       ON cat.`parent_id` = 0
      AND cat.`name` = CASE WHEN TRIM(IFNULL(art.`category`, '')) = '' THEN '未分类' ELSE TRIM(art.`category`) END
SET art.`category_id` = COALESCE(cat.`id`, 0);

-- 3.3 兜底：确保存在「未分类」，并回填 category_id 仍为 0 的文章
INSERT IGNORE INTO `health_education_categories` (`parent_id`, `name`, `sort`, `status`, `created_at`, `updated_at`)
VALUES (0, '未分类', 999, 1, NOW(), NOW());

UPDATE `health_education_articles` art
JOIN `health_education_categories` cat ON cat.`parent_id` = 0 AND cat.`name` = '未分类'
SET art.`category_id` = cat.`id`
WHERE art.`category_id` = 0;

-- ----------------------------------------------------------
-- 4. sys_menus：健康宣教独立板块（分类管理 + 内容管理按钮权限）
--    注：sys_menus 无 component 字段，仅 parent_id/name/path/icon/sort/menu_type/permission/visible/status
-- ----------------------------------------------------------
SET @parent_health_id := (
  SELECT `id` FROM `sys_menus`
  WHERE `permission` = 'health:view' OR `name` = '健康管理' OR `name` = '健康服务'
  ORDER BY `id` ASC LIMIT 1
);

SET @sort_max := (SELECT COALESCE(MAX(`sort`), 0) FROM `sys_menus` WHERE `parent_id` = @parent_health_id);

-- 4.1 宣教分类管理（菜单 + 按钮权限）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
VALUES
  (@parent_health_id, '宣教分类管理', '/health/education-categories', 'Folder', @sort_max + 1, 2, 'education-categories:view', 1, 1, NOW(), NOW());

SET @cat_menu_id := LAST_INSERT_ID();

INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
VALUES
  (@cat_menu_id, '新建分类', '', '', 1, 3, 'education-categories:create', 1, 1, NOW(), NOW()),
  (@cat_menu_id, '编辑分类', '', '', 2, 3, 'education-categories:update', 1, 1, NOW(), NOW()),
  (@cat_menu_id, '删除分类', '', '', 3, 3, 'education-categories:delete', 1, 1, NOW(), NOW());

-- 4.2 原有健康宣教内容菜单补齐按钮权限（education:view 页面菜单已存在）
SET @edu_menu_id := (
  SELECT `id` FROM `sys_menus` WHERE `permission` = 'education:view' ORDER BY `id` ASC LIMIT 1
);
INSERT IGNORE INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
VALUES
  (@edu_menu_id, '新建文章', '', '', 1, 3, 'education:create', 1, 1, NOW(), NOW()),
  (@edu_menu_id, '编辑文章', '', '', 2, 3, 'education:update', 1, 1, NOW(), NOW()),
  (@edu_menu_id, '删除文章', '', '', 3, 3, 'education:delete', 1, 1, NOW(), NOW());

-- 4.3 超管角色（默认 id=1）赋予新增按钮权限
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`, `created_at`)
SELECT 1, `id`, NOW()
FROM `sys_menus`
WHERE `permission` IN (
  'education-categories:view',
  'education-categories:create',
  'education-categories:update',
  'education-categories:delete',
  'education:create',
  'education:update',
  'education:delete'
);

SET FOREIGN_KEY_CHECKS = 1;
