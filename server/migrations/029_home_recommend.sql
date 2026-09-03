-- ============================================================
-- 029_home_recommend.sql — 小程序首页推荐商品/服务配置
-- ------------------------------------------------------------
-- 说明：
--  1. 创建 store_home_recommends 表（商家维度配置 C 端首页推荐项，
--     可指向实物商品(产品类型1/2)或服务(产品类型3/4；参考 products.id)
--  2. 新增「首页推荐」菜单与按钮权限，绑定超管角色
--  幂等：按 permission 去重，已存在则跳过。
-- ============================================================
SET NAMES utf8mb4;

-- 1. store_home_recommends 表
CREATE TABLE IF NOT EXISTS store_home_recommends (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '推荐ID',
  merchant_id  BIGINT UNSIGNED NOT NULL COMMENT '所属商家ID',
  product_id   BIGINT UNSIGNED NOT NULL COMMENT '目标商品/服务ID(products.id)',
  target_type  TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '推荐对象类型: 1=实物商品(产品类型1/2) 2=服务(产品类型3/4)',
  title        VARCHAR(128)     DEFAULT NULL COMMENT '展示标题(空则用商品名)',
  sort         INT UNSIGNED     NOT NULL DEFAULT 0 COMMENT '排序值(越小越靠前)',
  status       TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
  created_at   DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at   DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_shr_merchant (merchant_id),
  KEY idx_shr_status_sort (status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小程序首页推荐商品/服务表';

-- 2. RBAC：新增「首页推荐」顶级菜单与按钮权限
-- 顶级菜单：首页推荐（顶级菜单 menu_type=1，放轮播图菜单之后）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT 0, '首页推荐', '/home-recommends', '⭐', 16, 1, 'home-recommend:view', 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'home-recommend:view');

-- 按钮权限：新增/编辑/删除/启停（挂在「首页推荐」菜单下）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT id, '新增推荐', NULL, NULL, 1, 2, 'home-recommend:create', 1, 1, NOW(), NOW()
FROM `sys_menus`
WHERE `permission` = 'home-recommend:view'
  AND NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'home-recommend:create');

INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT id, '编辑推荐', NULL, NULL, 2, 2, 'home-recommend:update', 1, 1, NOW(), NOW()
FROM `sys_menus`
WHERE `permission` = 'home-recommend:view'
  AND NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'home-recommend:update');

INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT id, '删除推荐', NULL, NULL, 3, 2, 'home-recommend:delete', 1, 1, NOW(), NOW()
FROM `sys_menus`
WHERE `permission` = 'home-recommend:view'
  AND NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'home-recommend:delete');

INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT id, '启用/禁用', NULL, NULL, 4, 2, 'home-recommend:status', 1, 1, NOW(), NOW()
FROM `sys_menus`
WHERE `permission` = 'home-recommend:view'
  AND NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'home-recommend:status');

-- 绑定超管角色（role_id=1）
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`, `created_at`)
SELECT 1, `id`, NOW()
FROM `sys_menus`
WHERE `permission` IN ('home-recommend:view', 'home-recommend:create', 'home-recommend:update', 'home-recommend:delete', 'home-recommend:status');