-- ============================================================
-- 020_printer_menu.sql — 打印机管理菜单与权限
-- ------------------------------------------------------------
-- 说明：
--  1. 在「系统管理」(id=7) 父菜单下新增「打印机管理」页面节点
--  2. 页面节点权限码 printers:view，按钮 create/update/delete
--     （前端路由 /system/printers 的 meta.permission、报表审批权限码与后端打印接口 RBAC 一致）
--  3. 绑定超管角色（role_id=1）
-- ============================================================
SET NAMES utf8mb4;

-- 1. 计算父菜单下最大 sort，保证追加在现有节点之后
SET @parent_id := 7;
SET @sort_max := (SELECT COALESCE(MAX(`sort`), -1) FROM `sys_menus` WHERE `parent_id` = @parent_id);

-- 2. 新增「打印机管理」页面节点（menu_type 1=菜单/目录）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
VALUES
  (@parent_id, '打印机管理', '/system/printers', 'Printer', @sort_max + 1, 1, 'printers:view', 1, 1, NOW(), NOW());

SET @printer_menu_id := LAST_INSERT_ID();

-- 3. 新增按钮权限节点（menu_type 2=按钮）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
VALUES
  (@printer_menu_id, '新增打印机', '', '', 1, 2, 'printers:create', 1, 1, NOW(), NOW()),
  (@printer_menu_id, '编辑/测试',   '', '', 2, 2, 'printers:update', 1, 1, NOW(), NOW()),
  (@printer_menu_id, '删除打印机', '', '', 3, 2, 'printers:delete', 1, 1, NOW(), NOW());

-- 4. 授予超管角色（role_id=1）打印机管理相关权限
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`, `created_at`)
SELECT 1, `id`, NOW()
FROM `sys_menus`
WHERE `permission` IN ('printers:view', 'printers:create', 'printers:update', 'printers:delete');