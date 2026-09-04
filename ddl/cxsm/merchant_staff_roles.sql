create table merchant_staff_roles
(
    id         bigint unsigned auto_increment
        primary key,
    staff_id   bigint unsigned                    not null comment '员工ID(merchant_staffs.id)',
    role_id    bigint unsigned                    not null comment '角色ID(sys_roles.id)',
    created_at datetime default CURRENT_TIMESTAMP not null,
    constraint uk_merchant_staff_roles
        unique (staff_id, role_id)
)
    comment '员工角色关联表';

