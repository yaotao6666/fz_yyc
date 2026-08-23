-- ============================================================
-- 阶段三：居家康养照护
-- 表：care_plans / care_visits
-- 并追加「照护计划」菜单种子（含按钮权限），角色1绑定全部菜单
-- 约束：纯 SQL 单条执行（CREATE TABLE IF NOT EXISTS / INSERT），无存储过程
-- ============================================================

-- ============================================================
-- 1. 照护计划表
-- ============================================================
CREATE TABLE IF NOT EXISTS care_plans (
    id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id            BIGINT UNSIGNED NOT NULL COMMENT '居民用户ID',
    name               VARCHAR(64) NOT NULL COMMENT '计划名称',
    plan_type          TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '类型:1生活照料2基础护理3康复训练4综合康养',
    start_date         DATE DEFAULT NULL COMMENT '开始日期',
    end_date           DATE DEFAULT NULL COMMENT '结束日期',
    frequency          VARCHAR(64) DEFAULT NULL COMMENT '照护频次',
    goals              VARCHAR(512) DEFAULT NULL COMMENT '照护目标',
    items              JSON DEFAULT NULL COMMENT '护理项配置[{name,desc}]',
    assigned_staff_id  BIGINT UNSIGNED DEFAULT NULL COMMENT '指派服务人员ID',
    order_id           BIGINT UNSIGNED DEFAULT NULL COMMENT '关联服务订单ID(可空)',
    status             TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0草稿1执行中2已暂停3已完成',
    created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_care_plans_user_id (user_id),
    KEY idx_care_plans_assigned_staff_id (assigned_staff_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='照护计划表';

-- ============================================================
-- 2. 上门照护记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS care_visits (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    plan_id           BIGINT UNSIGNED DEFAULT NULL COMMENT '关联照护计划ID(可空)',
    order_id          BIGINT UNSIGNED DEFAULT NULL COMMENT '关联服务订单ID(可空)',
    user_id           BIGINT UNSIGNED NOT NULL COMMENT '居民用户ID',
    staff_id          BIGINT UNSIGNED NOT NULL COMMENT '录入服务人员ID',
    visit_at          DATETIME DEFAULT NULL COMMENT '到访时间',
    nursing_items     JSON DEFAULT NULL COMMENT '完成的护理项[{name,done,remark}]',
    vitals            JSON DEFAULT NULL COMMENT '生命体征{blood_pressure,blood_glucose,heart_rate,oxygen,weight}',
    photos            JSON DEFAULT NULL COMMENT '照片URL数组',
    remark            VARCHAR(512) DEFAULT NULL COMMENT '备注',
    follow_up_advice  VARCHAR(512) DEFAULT NULL COMMENT '下次随访建议',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_care_visits_plan_id (plan_id),
    KEY idx_care_visits_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上门照护记录表';

-- ============================================================
-- 3. RBAC 菜单种子（照护计划）
-- ============================================================
INSERT INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(85, 8, 1, '照护计划', '/health/care-plans', NULL, 5, 1, 1, 'care:view'),
(851, 85, 2, '新增/编辑', NULL, NULL, 1, 1, 1, 'care:create');

-- 超级管理员角色(1)绑定全部菜单（含新增菜单）
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus;
