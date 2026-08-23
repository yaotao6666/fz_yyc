-- ============================================================
-- 阶段二：康复辅具适配建议
-- 表：fitting_recommendations
-- 并追加「适配建议」菜单种子（含按钮权限），角色1绑定全部菜单
-- 约束：纯 SQL 单条执行（CREATE TABLE IF NOT EXISTS / INSERT），无存储过程
-- ============================================================

-- ============================================================
-- 1. 康复辅具适配建议表
-- ============================================================
CREATE TABLE IF NOT EXISTS fitting_recommendations (
    id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id              BIGINT UNSIGNED NOT NULL COMMENT '居民用户ID',
    assessment_id        BIGINT UNSIGNED DEFAULT NULL COMMENT '关联评估记录ID',
    symptom_desc         VARCHAR(512) DEFAULT NULL COMMENT '症状/需求描述',
    fitting_result       VARCHAR(512) DEFAULT NULL COMMENT '适配结论',
    recommended_products JSON DEFAULT NULL COMMENT '推荐商品快照[{product_id,name,reason,sale_type}]',
    staff_id             BIGINT UNSIGNED DEFAULT NULL COMMENT '生成建议的服务人员ID',
    status               TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0草稿1已确认2已下单',
    order_id             BIGINT UNSIGNED DEFAULT NULL COMMENT '关联订单ID',
    created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_fitting_recommendations_user_id (user_id),
    KEY idx_fitting_recommendations_staff_id (staff_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='康复辅具适配建议表';

-- ============================================================
-- 2. RBAC 菜单种子（适配建议）
-- ============================================================
INSERT INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(84, 8, 1, '适配建议', '/health/fitting', NULL, 4, 1, 1, 'fitting:view'),
(841, 84, 2, '编辑/确认', NULL, NULL, 1, 1, 1, 'fitting:update');

-- 超级管理员角色(1)绑定全部菜单（含新增菜单）
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus;
