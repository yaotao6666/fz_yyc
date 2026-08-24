-- ============================================================
-- 005_mini_program_banners.sql — 小程序首页轮播图配置
-- ------------------------------------------------------------
-- 说明：
--  1. 创建 mini_program_banners 表（商家维度配置 C 端轮播图）
--  2. RBAC：商家资料下新增「小程序轮播图」菜单 + 增删改查按钮权限
-- ============================================================

SET NAMES utf8mb4;

-- ============================================================
-- 1. mini_program_banners 表
-- ============================================================

CREATE TABLE IF NOT EXISTS mini_program_banners (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '轮播图ID',
  merchant_id  BIGINT UNSIGNED NOT NULL COMMENT '所属商家ID',
  title        VARCHAR(128)    DEFAULT NULL COMMENT '轮播图标题(仅后台识别)',
  image        VARCHAR(512)    NOT NULL COMMENT '轮播图图片URL(七牛私有路径)',
  link_type    VARCHAR(16)     NOT NULL DEFAULT 'none' COMMENT '跳转类型: none=无跳转 product=商品详情 category=分类页 url=外部链接',
  link_value   VARCHAR(256)    DEFAULT NULL COMMENT '跳转目标值: product=product_id category=空 url=链接',
  sort         INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '排序值(越小越靠前)',
  status       TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
  created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_mpb_merchant (merchant_id),
  KEY idx_mpb_status_sort (status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小程序首页轮播图表';

-- ============================================================
-- 2. RBAC：商家资料下新增「小程序轮播图」菜单与按钮权限
--    父菜单：商家资料（merchant-profile，sys_menus.id 需要确认）
--    约定：使用新的独立菜单节点 ID 段
-- ============================================================

-- 主菜单：小程序轮播图（顶级菜单，放在商家资料后）
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(95, 0, 1, '小程序轮播图', '/miniprogram-banners', '📷', 15, 1, 1, 'banners:view');

-- 按钮权限：查看/新增/编辑/删除/排序
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(951, 95, 2, '新增轮播图', NULL, NULL, 1, 1, 1, 'banners:create'),
(952, 95, 2, '编辑轮播图', NULL, NULL, 2, 1, 1, 'banners:update'),
(953, 95, 2, '删除轮播图', NULL, NULL, 3, 1, 1, 'banners:delete'),
(954, 95, 2, '启用/禁用', NULL, NULL, 4, 1, 1, 'banners:status');

-- 绑定超级管理员(role_id=1)
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus WHERE id IN (95, 951, 952, 953, 954);
