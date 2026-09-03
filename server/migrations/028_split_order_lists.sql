-- ============================================================
-- 028_split_order_lists.sql — 订单列表拆分为「实物订单/服务订单」两个页面
-- ------------------------------------------------------------
-- 说明：
--  1. 在「订单管理」(父菜单 id=2) 下新增「实物订单」「服务订单」两个页面节点，
--     权限码分别为 order:goods / order:service（前端路由 /orders/goods、/orders/service）
--  2. 后端订单查询接口 GET /orders、GET /orders/:order_id 已改用 RBACAny
--     任一订单权限(orders:view/order:goods/order:service)即可访问
--  3. 停用旧的「订单列表」(id=745, permission=orders:view)
--  4. 授权超管角色(role_id=1) 两个新权限码
--  幂等：按 permission 去重，已存在则跳过。
-- ============================================================
SET NAMES utf8mb4;

-- 1. 新增「实物订单」页面节点（menu_type 1=菜单/目录）
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT 2, '实物订单', '/orders/goods', '', 1, 1, 'order:goods', 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'order:goods');

-- 2. 新增「服务订单」页面节点
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT 2, '服务订单', '/orders/service', '', 2, 1, 'order:service', 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'order:service');

-- 3. 停用旧的「订单列表」节点
UPDATE `sys_menus` SET `status` = 0 WHERE `id` = 745 AND `permission` = 'orders:view' AND `name` = '订单列表';

-- 4. 授予超管角色（role_id=1）两个新权限码（原子补齐，幂等）
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`, `created_at`)
SELECT 1, `id`, NOW()
FROM `sys_menus`
WHERE `permission` IN ('order:goods', 'order:service');