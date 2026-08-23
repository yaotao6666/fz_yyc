-- ============================================================
-- 003_service_staff_audit.sql — 服务人员审核机制 + 资质文件
-- ------------------------------------------------------------
-- 说明：
--  1. service_staffs 新增资质与审核扩展列（幂等）
--  2. 新增 service_staff_audit_records 审核记录表（注册/变更/资质提交留痕）
--  3. RBAC：服务人员顶级下新增「审核列表」菜单与独立审核权限
-- 命名约定：服务人员相关表统一 `service_` 前缀（service_staffs / service_staff_audit_records）
-- 与启用状态解耦：status 表示启用/禁用，audit_status 表示审核态。
-- ============================================================

-- ============================================================
-- 1. service_staffs 扩展列
-- ============================================================

-- qualifications: 资质材料 URL 数组 JSON（[{type,name,url}])
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_staffs' AND COLUMN_NAME='qualifications');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE service_staffs ADD COLUMN qualifications JSON DEFAULT NULL COMMENT '资质材料列表JSON' AFTER avatar");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- audit_status: 0=无/已通过 1=待审核（与 status 解耦）
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_staffs' AND COLUMN_NAME='audit_status');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE service_staffs ADD COLUMN audit_status tinyint unsigned NOT NULL DEFAULT 0 COMMENT '审核状态: 0=无/已通过 1=待审核' AFTER status");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- pending_fields: 正在审核中的待变更字段快照 JSON
SET @_col = (SELECT COLUMN_NAME FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='service_staffs' AND COLUMN_NAME='pending_fields');
SET @_sql = IF(@_col IS NOT NULL, 'SELECT 1',
  "ALTER TABLE service_staffs ADD COLUMN pending_fields JSON DEFAULT NULL COMMENT '审核中待变更字段快照' AFTER audit_status");
PREPARE stmt FROM @_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. 审核记录表 service_staff_audit_records
-- ============================================================
CREATE TABLE IF NOT EXISTS `service_staff_audit_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `staff_id` bigint unsigned NOT NULL COMMENT '服务人员ID(service_staffs.id)',
  `audit_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '审核类型:1=注册申请 2=信息变更 3=资质提交 4=状态变更',
  `apply_type` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '申请来源:1=自注册 2=PC添加 3=端上变更',
  `before_data` json DEFAULT NULL COMMENT '变更前字段快照',
  `after_data` json DEFAULT NULL COMMENT '变更后字段快照(审核通过后回写)',
  `qualifications` json DEFAULT NULL COMMENT '资质材料URL列表',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '审核状态:0=待审 1=通过 2=驳回',
  `reviewer_id` bigint unsigned DEFAULT NULL COMMENT '审核人工员ID(merchant_staffs.id)',
  `review_remark` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '审核备注',
  `review_at` datetime DEFAULT NULL COMMENT '审核时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_staff_audit_staff_id` (`staff_id`),
  KEY `idx_staff_audit_type` (`audit_type`),
  KEY `idx_staff_audit_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务人员审核记录表';

-- ============================================================
-- 3. RBAC：服务人员顶部(4)下新增「审核列表」菜单 + 权限
-- ============================================================
INSERT IGNORE INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(75, 4, 1, '审核列表', '/staff/audits', NULL, 2, 1, 1, 'staffaudit:view'),
(751, 75, 2, '审核通过', NULL, NULL, 1, 1, 1, 'staffaudit:approve'),
(752, 75, 2, '驳回', NULL, NULL, 2, 1, 1, 'staffaudit:reject');

-- 绑定超级管理员(role_id=1)
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus WHERE id IN (75, 751, 752);