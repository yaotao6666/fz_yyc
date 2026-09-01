-- ============================================================
-- 去重：健康宣教菜单节点（012 被重复执行产生冗余行）
-- 现象：education 相关 permission 存在多条记录，侧边栏会重复渲染入口
-- 策略：按 permission 分组，保留最小 id，删除其余及其 sys_role_menus 引用
-- 幂等：重复执行时无重复组，临时表为空，不产生副作用
-- 编号：014
-- 执行：mysql -h <host> -u <user> -p <db> < server/migrations/014_dedupe_education_menu.sql
-- ============================================================
SET NAMES utf8mb4;

-- 1. 找出每个 permission 中需删除的重复 id（保留最小 id）
DROP TEMPORARY TABLE IF EXISTS tmp_edu_dup;
CREATE TEMPORARY TABLE tmp_edu_dup AS
SELECT m.`id`
FROM `sys_menus` m
JOIN (
  SELECT `permission`, MIN(`id`) AS keep_id
  FROM `sys_menus`
  WHERE `permission` IN (
    'education:view', 'education:create', 'education:update', 'education:delete',
    'education-categories:view', 'education-categories:create', 'education-categories:update', 'education-categories:delete'
  )
  GROUP BY `permission`
  HAVING COUNT(*) > 1
) g ON m.`permission` = g.`permission` AND m.`id` > g.keep_id;

-- 2. 先删除角色授权引用，避免孤儿数据
DELETE rm FROM `sys_role_menus` rm
JOIN `tmp_edu_dup` t ON rm.`menu_id` = t.`id`;

-- 3. 删除重复菜单节点
DELETE m FROM `sys_menus` m
JOIN `tmp_edu_dup` t ON m.`id` = t.`id`;

DROP TEMPORARY TABLE IF EXISTS tmp_edu_dup;