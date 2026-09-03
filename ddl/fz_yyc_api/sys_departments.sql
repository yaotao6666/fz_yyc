create table sys_departments
(
    id         bigint unsigned auto_increment
        primary key,
    parent_id  bigint unsigned  default '0'               not null comment '父部门ID(0=顶级)',
    name       varchar(64)                                not null comment '部门名称',
    leader     varchar(64)                                null comment '负责人',
    phone      varchar(20)                                null comment '联系电话',
    sort       int unsigned     default '0'               not null comment '排序值',
    status     tinyint unsigned default '1'               not null comment '状态: 1=启用 0=禁用',
    created_at datetime         default CURRENT_TIMESTAMP not null,
    updated_at datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '系统部门表';

create index idx_sys_departments_parent_id
    on sys_departments (parent_id);

