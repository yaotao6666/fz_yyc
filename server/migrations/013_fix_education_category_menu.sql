-- ============================================================
-- 修正健康宣教分类菜单 menu_type（PC 端无入口问题修复）
-- 背景：012 将「宣教分类管理」写成 menu_type=2、按钮写成 3，
--       而 sys_menus 规定 1=菜单/目录、2=按钮；前端只渲染 menu_type=1。
-- 本迁移不重跑 012（其含 DROP TABLE，属破坏性操作），用幂等 UPDATE 修正。
-- 编号：013
-- 执行：mysql -h <host> -u <user> -p <db> < server/migrations/013_fix_education_category_menu.sql
-- ============================================================
SET NAMES utf8mb4;

-- 1. 「宣教分类管理」页面菜单：menu_type 2 -> 1（可被侧边栏渲染）
UPDATE `sys_menus`
SET `menu_type` = 1, `updated_at` = NOW()
WHERE `permission` = 'education-categories:view'
  AND `menu_type` != 1;

-- 2. 分类按钮权限节点：menu_type 3 -> 2（与 schema 按钮语义一致）
UPDATE `sys_menus`
SET `menu_type` = 2, `updated_at` = NOW()
WHERE `permission` IN ('education-categories:create', 'education-categories:update', 'education-categories:delete')
  AND `menu_type` != 2;

-- 3. 文章按钮权限节点（012 补齐的 education:update / education:delete）：menu_type 3 -> 2
UPDATE `sys_menus`
SET `menu_type` = 2, `updated_at` = NOW()
WHERE `permission` IN ('education:update', 'education:delete')
  AND `menu_type` != 2;