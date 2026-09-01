-- ============================================================
-- 007_coupon_system.sql — 阶段二：优惠券营销闭环
-- ------------------------------------------------------------
-- 说明：
--  1. 新建 coupon_templates 券模板表、user_coupons 用户券表（幂等）
--  2. RBAC：商品管理目录(id=3)下新增「优惠券管理」菜单与按钮权限节点
--  3. 权限码：coupon-templates:view/create/update/delete、user-coupons:view
-- ============================================================

SET NAMES utf8mb4;

-- ============================================================
-- 1. 券模板表 coupon_templates
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='coupon_templates');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE coupon_templates (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '券模板ID',
  name VARCHAR(64) NOT NULL COMMENT '券名称',
  type TINYINT UNSIGNED NOT NULL COMMENT '券类型: 1=满减券 2=折扣券',
  threshold_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '使用门槛金额(0=无门槛)',
  discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '满减面值(满减券)',
  discount_rate DECIMAL(3,2) NOT NULL DEFAULT 0 COMMENT '折扣率如0.90(折扣券)',
  total_count INT NOT NULL DEFAULT 0 COMMENT '发行总量(0=不限)',
  received_count INT NOT NULL DEFAULT 0 COMMENT '已领取数量',
  per_user_limit INT NOT NULL DEFAULT 1 COMMENT '每人限领数量',
  valid_type TINYINT UNSIGNED NOT NULL COMMENT '有效期类型: 1=固定期限 2=领取后N天有效',
  valid_start_at DATETIME DEFAULT NULL COMMENT '固定期限开始时间',
  valid_end_at DATETIME DEFAULT NULL COMMENT '固定期限结束时间',
  valid_days INT DEFAULT NULL COMMENT '领取后有效天数',
  apply_scope TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '适用范围: 1=全场 2=指定分类 3=指定商品',
  scope_ids JSON DEFAULT NULL COMMENT '适用范围ID列表',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 0=停用',
  remark VARCHAR(512) DEFAULT NULL COMMENT '备注',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券模板表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. 用户券表 user_coupons
-- ============================================================
SET @_tbl = (SELECT COUNT(*) FROM information_schema.TABLES
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='user_coupons');
SET @_sql = IF(@_tbl > 0, 'SELECT 1',
"CREATE TABLE user_coupons (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '用户券ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  template_id BIGINT UNSIGNED NOT NULL COMMENT '券模板ID',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态: 1=未使用 2=已使用 3=已过期 4=已作废',
  source TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '来源: 1=自主领取 2=系统发放(30天唤回) 3=运营手动发放',
  received_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
  expired_at DATETIME NOT NULL COMMENT '过期时间(领取时计算落库)',
  used_at DATETIME DEFAULT NULL COMMENT '核销时间',
  order_id BIGINT UNSIGNED DEFAULT NULL COMMENT '核销关联订单ID',
  order_no VARCHAR(32) DEFAULT NULL COMMENT '核销关联订单号',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  KEY idx_user_coupons_user_id (user_id),
  KEY idx_user_coupons_template_id (template_id),
  KEY idx_user_coupons_order_id (order_id),
  KEY idx_user_coupons_status_expired (status, expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券表'");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 3. RBAC：商品管理目录(id=3)下新增「优惠券管理」菜单 + 按钮权限
-- ============================================================
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(96, 3, 1, '优惠券管理', '/coupon-templates', NULL, 4, 1, 1, 'coupon-templates:view'),
(960, 96, 2, '新增模板', NULL, NULL, 1, 1, 1, 'coupon-templates:create'),
(961, 96, 2, '编辑模板', NULL, NULL, 2, 1, 1, 'coupon-templates:update'),
(962, 96, 2, '删除模板', NULL, NULL, 3, 1, 1, 'coupon-templates:delete'),
(963, 96, 2, '启停模板', NULL, NULL, 4, 1, 1, 'coupon-templates:update'),
(964, 96, 2, '手动发放', NULL, NULL, 5, 1, 1, 'coupon-templates:create'),
(965, 96, 2, '领取/使用记录', NULL, NULL, 6, 1, 1, 'user-coupons:view');

-- 绑定超级管理员(role_id=1)
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus WHERE id IN (96, 960, 961, 962, 963, 964, 965);
