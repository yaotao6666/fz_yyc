-- 阶段五 8.1 抹除照护计划/随访任务/生命体征
-- 前置：专项备份已完成（server/backups/p5_health_cleanup_pre_*.sql）
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 1) 撤销 sys_role_menus 对应菜单权限（先删关联再删菜单，避免 FK）
DELETE FROM sys_role_menus WHERE menu_id IN (85,851,852,853, 86,861, 87,871);

-- 2) 移除 sys_menus 中 照护计划/随访任务/生命体征 相关节点
DELETE FROM sys_menus WHERE id IN (85,851,852,853, 86,861, 87,871);

-- 3) 丢弃 4 张与业务无关的健康服务表（按依赖顺序：先子表后父表）
DROP TABLE IF EXISTS care_visits;
DROP TABLE IF EXISTS care_plans;
DROP TABLE IF EXISTS follow_up_tasks;
DROP TABLE IF EXISTS health_monitoring;

SET FOREIGN_KEY_CHECKS = 1;
