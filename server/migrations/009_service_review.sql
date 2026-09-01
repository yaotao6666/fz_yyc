-- ============================================================
-- 009_service_review.sql — 阶段四：评价与服务质量分
-- ------------------------------------------------------------
-- 说明：
--  1. 新建 service_reviews 服务评价表（一单一评，幂等）
--  2. RBAC：服务人员菜单(id=4)下新增「服务评价」菜单与「隐藏评价」按钮
--  3. 权限码：service-reviews:view / service-reviews:update
-- ============================================================

SET NAMES utf8mb4;

-- ============================================================
-- 1. 服务评价表 service_reviews
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_reviews');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE service_reviews (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '评价ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID(一单一评)',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '下单用户ID',
  staff_id BIGINT UNSIGNED NOT NULL COMMENT '被评价服务人员ID',
  score TINYINT UNSIGNED NOT NULL COMMENT '总体评分 1-5',
  attitude_score TINYINT UNSIGNED NOT NULL COMMENT '服务态度分 1-5',
  professional_score TINYINT UNSIGNED NOT NULL COMMENT '专业技能分 1-5',
  punctual_score TINYINT UNSIGNED NOT NULL COMMENT '准时守约分 1-5',
  content VARCHAR(512) DEFAULT NULL COMMENT '评价内容',
  images JSON DEFAULT NULL COMMENT '评价图片URL列表',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=正常展示 0=后台隐藏',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_service_reviews_order_id (order_id),
  KEY idx_service_reviews_user_id (user_id),
  KEY idx_service_reviews_staff_id (staff_id),
  KEY idx_service_reviews_status (status),
  KEY idx_service_reviews_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务评价表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. RBAC：服务人员(id=4)下新增「服务评价」菜单 + 「隐藏评价」按钮
-- ============================================================
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(98, 4, 1, '服务评价', '/service-reviews', NULL, 3, 1, 1, 'service-reviews:view'),
(981, 98, 2, '隐藏评价', NULL, NULL, 1, 1, 1, 'service-reviews:update');

-- 授予超管角色（role_id=1）新菜单权限
INSERT IGNORE INTO sys_role_menus (role_id, menu_id)
SELECT 1, id FROM sys_menus WHERE id IN (98, 981);