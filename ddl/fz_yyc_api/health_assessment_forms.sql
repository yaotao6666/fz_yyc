create table health_assessment_forms
(
    id          bigint unsigned auto_increment
        primary key,
    name        varchar(64)                                not null comment '量表名称',
    dimension   varchar(32)                                null comment '评估维度',
    description text                                       null comment '量表说明',
    questions   json                                       null comment '题目数组[{key,title,options:[{label,score}]}]',
    score_rule  json                                       null comment '评分规则[{min,max,level,conclusion}]',
    version     int unsigned     default '1'               not null comment '版本号',
    status      tinyint unsigned default '0'               not null comment '0草稿1启用',
    created_at  datetime         default CURRENT_TIMESTAMP not null,
    updated_at  datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '健康评估量表表';

