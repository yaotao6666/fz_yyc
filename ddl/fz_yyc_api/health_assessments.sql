create table health_assessments
(
    id            bigint unsigned auto_increment
        primary key,
    user_id       bigint unsigned                            not null comment '用户ID',
    record_id     bigint unsigned                            null comment '关联档案ID（多档案）',
    form_id       bigint unsigned                            not null comment '量表ID',
    form_name     varchar(64)                                null comment '量表名称快照',
    assessor_type tinyint unsigned default '1'               not null comment '1自助2服务人员',
    staff_id      bigint unsigned                            null comment '评估服务人员ID',
    answers       json                                       null comment '答案{key:选中label}',
    total_score   decimal(6, 1)                              null comment '总分',
    level         varchar(32)                                null comment '评估等级',
    conclusion    varchar(512)                               null comment '评估结论',
    suggestions   json                                       null comment '建议数组',
    symptom_desc  varchar(512)                               null comment '症状描述',
    created_at    datetime         default CURRENT_TIMESTAMP not null
)
    comment '健康评估记录表';

create index idx_health_assessments_form_id
    on health_assessments (form_id);

create index idx_health_assessments_record_id
    on health_assessments (record_id);

create index idx_health_assessments_user_id
    on health_assessments (user_id);

