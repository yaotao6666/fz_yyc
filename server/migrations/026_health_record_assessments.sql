-- ============================================================
-- 026_health_record_assessments.sql — 健康档案「评估记录」查看权限
-- ------------------------------------------------------------
-- 说明：
--  1. 健康档案列表操作列将评估记录独立展示，需在「健康档案」(父菜单 id=81)
--     下新增按钮节点，权限码与后端接口 RBAC 一致（health:assessment）
--  2. 前端健康档案操作列「评估记录」按钮 v-permission、后端接口
--     GET /health-records/:id/assessments 的 RBAC("health:assessment")、
--     sys_menus.permission 三处保持一致
--  3. 绑定超管角色（role_id=1）
--  幂等：按 permission 去重，已存在则跳过。
-- ============================================================
SET NAMES utf8mb4;

-- 1. 容器化安装时若脚本重复执行，避免主键重复（幂等插入）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT 81, '评估记录', '', '', 2, 2, 'health:assessment', 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'health:assessment');

-- 2. 授予超管角色（role_id=1）该按钮权限
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`, `created_at`)
SELECT 1, `id`, NOW()
FROM `sys_menus`
WHERE `permission` = 'health:assessment';