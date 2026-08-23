-- ============================================================
-- 阶段一：居民健康档案 + 人群健康评估
-- 表：health_records / health_assessment_forms / health_assessments
-- 并追加「健康服务」菜单种子（含按钮权限），角色1绑定全部菜单
-- 约束：纯 SQL 单条执行（CREATE TABLE IF NOT EXISTS / INSERT），无存储过程
-- ============================================================

-- ============================================================
-- 1. 居民健康档案表
-- ============================================================
CREATE TABLE IF NOT EXISTS health_records (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id           BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    real_name         VARCHAR(64) DEFAULT NULL COMMENT '真实姓名',
    gender            TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '性别:1男2女',
    birth_date        VARCHAR(16) DEFAULT NULL COMMENT '出生日期',
    id_card           VARCHAR(32) DEFAULT NULL COMMENT '身份证号',
    phone             VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    emergency_contact VARCHAR(64) DEFAULT NULL COMMENT '紧急联系人',
    emergency_phone   VARCHAR(20) DEFAULT NULL COMMENT '紧急联系电话',
    address           VARCHAR(256) DEFAULT NULL COMMENT '常住地址',
    height_cm         DECIMAL(5,1) DEFAULT NULL COMMENT '身高(cm)',
    weight_kg         DECIMAL(5,1) DEFAULT NULL COMMENT '体重(kg)',
    blood_type        VARCHAR(8) DEFAULT NULL COMMENT '血型',
    past_history      JSON DEFAULT NULL COMMENT '既往病史数组',
    allergy_history   JSON DEFAULT NULL COMMENT '过敏史数组',
    family_history    JSON DEFAULT NULL COMMENT '家族病史数组',
    surgery_history   JSON DEFAULT NULL COMMENT '手术史数组',
    medication_list   JSON DEFAULT NULL COMMENT '长期用药数组',
    chronic_tags      JSON DEFAULT NULL COMMENT '慢病标签数组',
    smoking           VARCHAR(32) DEFAULT NULL COMMENT '吸烟情况',
    drinking          VARCHAR(32) DEFAULT NULL COMMENT '饮酒情况',
    assessment_level  VARCHAR(32) DEFAULT NULL COMMENT '最近一次评估等级',
    remark            VARCHAR(512) DEFAULT NULL COMMENT '备注',
    status            TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '0未建档1正常2已归档',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_health_records_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='居民健康档案表';

-- ============================================================
-- 2. 健康评估量表表
-- ============================================================
CREATE TABLE IF NOT EXISTS health_assessment_forms (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name        VARCHAR(64) NOT NULL COMMENT '量表名称',
    dimension   VARCHAR(32) DEFAULT NULL COMMENT '评估维度',
    description TEXT DEFAULT NULL COMMENT '量表说明',
    questions   JSON DEFAULT NULL COMMENT '题目数组[{key,title,options:[{label,score}]}]',
    score_rule  JSON DEFAULT NULL COMMENT '评分规则[{min,max,level,conclusion}]',
    version     INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '版本号',
    status      TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0草稿1启用',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='健康评估量表表';

-- ============================================================
-- 3. 健康评估记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS health_assessments (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id       BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    form_id       BIGINT UNSIGNED NOT NULL COMMENT '量表ID',
    form_name     VARCHAR(64) DEFAULT NULL COMMENT '量表名称快照',
    assessor_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1自助2服务人员',
    staff_id      BIGINT UNSIGNED DEFAULT NULL COMMENT '评估服务人员ID',
    answers       JSON DEFAULT NULL COMMENT '答案{key:选中label}',
    total_score   DECIMAL(6,1) DEFAULT NULL COMMENT '总分',
    level         VARCHAR(32) DEFAULT NULL COMMENT '评估等级',
    conclusion    VARCHAR(512) DEFAULT NULL COMMENT '评估结论',
    suggestions   JSON DEFAULT NULL COMMENT '建议数组',
    symptom_desc  VARCHAR(512) DEFAULT NULL COMMENT '症状描述',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_health_assessments_user_id (user_id),
    KEY idx_health_assessments_form_id (form_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='健康评估记录表';

-- ============================================================
-- 4. RBAC 菜单种子（健康服务）
-- ============================================================
INSERT INTO sys_menus (id, parent_id, menu_type, name, path, icon, sort, status, visible, permission) VALUES
(8, 0, 1, '健康服务', '/health', 'FirstAidKit', 8, 1, 1, NULL),
(81, 8, 1, '健康档案', '/health/records', NULL, 1, 1, 1, 'health:view'),
(811, 81, 2, '编辑档案', NULL, NULL, 1, 1, 1, 'health:update'),
(82, 8, 1, '评估量表', '/health/assessment-forms', NULL, 2, 1, 1, 'assessment:view'),
(821, 82, 2, '新增/编辑量表', NULL, NULL, 1, 1, 1, 'assessment:config'),
(83, 8, 1, '评估记录', '/health/assessments', NULL, 3, 1, 1, 'assessment:view'),
(831, 83, 2, '手动登记', NULL, NULL, 1, 1, 1, 'assessment:create');

-- 超级管理员角色(1)绑定全部菜单（含新增菜单）
INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus;
