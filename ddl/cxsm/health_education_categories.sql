create table health_education_categories
(
    id         bigint unsigned auto_increment comment '分类ID'
        primary key,
    parent_id  bigint unsigned  default '0'               not null comment '父分类ID，0=一级',
    name       varchar(64)                                not null comment '分类名称',
    sort       int              default 0                 not null comment '排序（小在前）',
    status     tinyint unsigned default '1'               not null comment '状态:1启用0停用',
    created_at datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_hec_parent_name
        unique (parent_id, name)
)
    comment '健康宣教分类表（阶段五 8.3，两级树形）';

create index idx_hec_parent_id
    on health_education_categories (parent_id);

create index idx_hec_status
    on health_education_categories (status);

