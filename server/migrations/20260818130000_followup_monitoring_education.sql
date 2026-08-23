-- ============================================================
-- 阶段四：基层健康服务闭环 —— 持续康复随访 / 生命体征监测 / 健康宣教
-- 表：follow_up_tasks / health_monitoring / health_education_articles
-- 并追加「随访任务」「生命体征」「健康宣教」菜单种子（含按钮权限），角色1绑定全部菜单
-- 约束：纯 SQL 单条执行（CREATE TABLE IF NOT EXISTS / INSERT），无存储过程
-- ============================================================

-- ============================================================
-- 1. 随访任务表
-- ============================================================
CREATE TABLE IF NOT EXISTS follow_up_tasks (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id           BIGINT UNSIGNED NOT NULL COMMENT '居民用户ID',
    task_type         TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '类型:1康复随访2租后回访3慢病随访4评估回访',
    source_type       TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '来源:1服务完成2租赁归还3评估完成4手动',
    source_id         BIGINT UNSIGNED DEFAULT NULL COMMENT '来源ID(订单/评估ID)',
    plan_follow_time  DATETIME DEFAULT NULL COMMENT '计划随访时间',
    staff_id          BIGINT UNSIGNED DEFAULT NULL COMMENT '执行服务人员ID(可空=待认领)',
    contact_method    TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '随访方式:1电话2上门3微信',
    status            TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0待执行1已完成2已跳过',
    result            JSON DEFAULT NULL COMMENT '随访结果{contact_method,content,education_article_ids,satisfaction,remark}',
    completed_at      DATETIME DEFAULT NULL COMMENT '完成时间',
    remark            VARCHAR(512) DEFAULT NULL COMMENT '备注',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_follow_up_tasks_user_id (user_id),
    KEY idx_follow_up_tasks_staff_id (staff_id),
    KEY idx_follow_up_tasks_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='随访任务表';

-- ============================================================
-- 2. 生命体征监测表
-- ============================================================
CREATE TABLE IF NOT EXISTS health_monitoring (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id           BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    record_type       TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '类型:1血压2血糖3心率4血氧5体重',
    value             DECIMAL(8,2) NOT NULL DEFAULT 0 COMMENT '测量值',
    unit              VARCHAR(16) DEFAULT NULL COMMENT '单位',
    extra             JSON DEFAULT NULL COMMENT '扩展(如血压高低压)',
    recorded_by       BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '录入人(0=用户本人)',
    recorded_at       DATETIME DEFAULT NULL COMMENT '测量时间',
    remark            VARCHAR(512) DEFAULT NULL COMMENT '备注',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_health_monitoring_user_id (user_id),
    KEY idx_health_monitoring_record_type (record_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生命体征监测表';

-- ============================================================
-- 3. 健康宣教内容表
-- ============================================================
CREATE TABLE IF NOT EXISTS health_education_articles (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    title             VARCHAR(128) NOT NULL COMMENT '标题',
    category          VARCHAR(32) DEFAULT NULL COMMENT '分类',
    cover             VARCHAR(512) DEFAULT NULL COMMENT '封面图URL',
    content           TEXT COMMENT '正文',
    tags              JSON DEFAULT NULL COMMENT '定向慢病标签数组',
    status            TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0草稿1发布',
    publish_at        DATETIME DEFAULT NULL COMMENT '发布时间',
    views             INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '浏览量',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='健康宣教内容表';

-- ============================================================
-- 4. RBAC 菜单种子（随访任务 / 生命体征 / 健康宣教）
-- ============================================================
INSERT INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(86, 8, 1, '随访任务', '/health/follow-ups', NULL, 6, 1, 1, 'followup:view'),
(861, 86, 2, '执行/登记', NULL, NULL, 1, 1, 1, 'followup:update'),
(87, 8, 1, '生命体征', '/health/monitoring', NULL, 7, 1, 1, 'monitor:view'),
(871, 87, 2, '录入', NULL, NULL, 1, 1, 1, 'monitor:create'),
(88, 8, 1, '健康宣教', '/health/education', NULL, 8, 1, 1, 'education:view'),
(881, 88, 2, '新增/编辑', NULL, NULL, 1, 1, 1, 'education:create');

-- 超级管理员角色(1)绑定全部菜单（含新增菜单）
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus;
