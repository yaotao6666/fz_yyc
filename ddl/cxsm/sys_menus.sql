create table sys_menus
(
    id         bigint unsigned auto_increment
        primary key,
    parent_id  bigint unsigned  default '0'               not null comment '父菜单ID(0=顶级)',
    menu_type  tinyint unsigned default '1'               not null comment '类型: 1=菜单/目录 2=按钮',
    name       varchar(64)                                not null comment '菜单名称',
    path       varchar(128)                               null comment '前端路由路径(菜单时)',
    icon       varchar(64)                                null comment '菜单图标',
    sort       int unsigned     default '0'               not null comment '排序值(越小越靠前)',
    status     tinyint unsigned default '1'               not null comment '状态: 1=启用 0=禁用',
    visible    tinyint unsigned default '1'               not null comment '是否显示: 1=显示 0=隐藏',
    permission varchar(128)                               null comment '权限标识(如 order:view)',
    created_at datetime         default CURRENT_TIMESTAMP not null,
    updated_at datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '系统菜单表';

create index idx_sys_menus_parent_id
    on sys_menus (parent_id);

create index idx_sys_menus_permission
    on sys_menus (permission);

