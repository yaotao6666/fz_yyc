create table sys_role_menus
(
    id         bigint unsigned auto_increment
        primary key,
    role_id    bigint unsigned                    not null comment '角色ID',
    menu_id    bigint unsigned                    not null comment '菜单ID',
    created_at datetime default CURRENT_TIMESTAMP not null,
    constraint uk_sys_role_menus
        unique (role_id, menu_id)
)
    comment '角色菜单关联表';

