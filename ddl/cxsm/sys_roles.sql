create table sys_roles
(
    id         bigint unsigned auto_increment
        primary key,
    name       varchar(64)                                not null comment '角色名称',
    code       varchar(64)                                not null comment '角色编码(唯一)',
    remark     varchar(256)                               null comment '备注',
    status     tinyint unsigned default '1'               not null comment '状态: 1=启用 0=禁用',
    created_at datetime         default CURRENT_TIMESTAMP not null,
    updated_at datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint uk_sys_roles_code
        unique (code)
)
    comment '系统角色表';

