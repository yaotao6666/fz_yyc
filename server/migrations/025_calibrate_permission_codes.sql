-- ============================================================
-- 025_calibrate_permission_codes.sql — 校准 sys_menus 与代码 RBAC 权限码一致
-- ------------------------------------------------------------
-- 背景：
--  1. 后端商品/分类“编辑”接口使用 products:update / categories:update，
--     但 sys_menus 仅有 products:create / categories:create（按钮名“新增/编辑”兼任），
--     导致拥有相应角色权限的非超管账号调用编辑接口时被 RBAC 拒绝(403)。
--     方案：按权限码命名规范独立补齐 products:update / categories:update 节点。
--  2. 健康宣教创建走 education:create，编辑/删除将分别对齐 education:update /
--     education:delete（DB 已存在这两个节点），代码端已同步改用，方向一致。
--  3. service-products:create/update/delete 为 006 遗留死节点：
--     「服务管理」页面复用商品接口（products:*），前端 ServiceProductsView 未施加
--     service-products:* 权限控制，代码亦无任何 RBAC 引用，需清理。
-- ------------------------------------------------------------
SET NAMES utf8mb4;

-- ============================================================
-- 1. 商品管理页面(31)下补充「编辑商品」按钮节点
-- ============================================================
SET @sort_31 := (SELECT COALESCE(MAX(`sort`), -1) FROM `sys_menus` WHERE `parent_id` = 31);
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT 31, '编辑商品', '', '', @sort_31 + 1, 2, 'products:update', 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'products:update');

-- 分类管理页面(32)下补充「编辑分类」按钮节点
SET @sort_32 := (SELECT COALESCE(MAX(`sort`), -1) FROM `sys_menus` WHERE `parent_id` = 32);
INSERT INTO `sys_menus` (`parent_id`, `name`, `path`, `icon`, `sort`, `menu_type`, `permission`, `visible`, `status`, `created_at`, `updated_at`)
SELECT 32, '编辑分类', '', '', @sort_32 + 1, 2, 'categories:update', 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `sys_menus` WHERE `permission` = 'categories:update');

-- 同步修正：原「新增/编辑」按钮语义归为「新增商品/新增分类」（仅改名，不影响 permission）
UPDATE `sys_menus` SET `name` = '新增商品' WHERE `permission` = 'products:create';
UPDATE `sys_menus` SET `name` = '新增分类' WHERE `permission` = 'categories:create';

-- ============================================================
-- 2. 清理 service-products 死节点（先删角色授权引用，再删菜单）
-- ============================================================
DELETE rm FROM `sys_role_menus` rm
JOIN `sys_menus` m ON rm.`menu_id` = m.`id`
WHERE m.`permission` IN ('service-products:create', 'service-products:update', 'service-products:delete');

DELETE FROM `sys_menus` WHERE `permission` IN ('service-products:create', 'service-products:update', 'service-products:delete');

-- ============================================================
-- 3. 授予超管角色(role_id=1)新增权限码（原子补齐，幂等）
-- ============================================================
INSERT IGNORE INTO `sys_role_menus` (`role_id`, `menu_id`, `created_at`)
SELECT 1, `id`, NOW()
FROM `sys_menus`
WHERE `permission` IN ('products:update', 'categories:update');