-- ============================================================
-- 024_system_config_menu.sql — 系统配置页菜单与权限
-- ------------------------------------------------------------
-- 说明：
--  1. 在「系统管理」(id=7) 父菜单下新增「系统配置」页面节点
--  2. 页面节点权限码 systemconfig:view，按钮 systemconfig:update
--     （前端路由 /system/config 的 meta.permission、RBAC 权限码、
--      后端 /system-configs 接口 RBAC 三者一致）
--  3. 绑定超管角色（role_id=1）
-- ============================================================
SET NAMES utf8mb4;

-- 1. 计算父菜单下最大 sort，保证追加在现有节点之后
SET @parent_id := 7;
SET @sort_max := (SELECT COALESCE(MAX(`sort`), -1) FROM `sys_menus` WHERE `parent_id` = @parent_id);

-- 2. 新增「系统配置」页面节点（menu_type 1=菜单/目录）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
VALUES
  (@parent_id, '系统配置', '/system/config', 'Setting', @sort_max + 1, 1, 'systemconfig:view', 1, 1, NOW(), NOW());

SET @config_menu_id := LAST_INSERT_ID();

-- 3. 新增按钮权限节点（menu_type 2=按钮）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
VALUES
  (@config_menu_id, '新增/编辑配置', '', '', 1, 2, 'systemconfig:update', 1, 1, NOW(), NOW());

-- 4. 授予超管角色（role_id=1）系统配置相关权限
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`, `created_at`)
SELECT 1, `id`, NOW()
FROM `sys_menus`
WHERE `permission` IN ('systemconfig:view', 'systemconfig:update');