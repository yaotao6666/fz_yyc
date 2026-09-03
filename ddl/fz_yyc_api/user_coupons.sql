create table user_coupons
(
    id          bigint unsigned auto_increment comment '用户券ID'
        primary key,
    user_id     bigint unsigned                            not null comment '用户ID',
    template_id bigint unsigned                            not null comment '券模板ID',
    status      tinyint unsigned default '1'               not null comment '状态: 1=未使用 2=已使用 3=已过期 4=已作废',
    source      tinyint unsigned default '1'               not null comment '来源: 1=自主领取 2=系统发放(30天唤回) 3=运营手动发放',
    received_at datetime         default CURRENT_TIMESTAMP not null comment '领取时间',
    expired_at  datetime                                   not null comment '过期时间(领取时计算落库)',
    used_at     datetime                                   null comment '核销时间',
    order_id    bigint unsigned                            null comment '核销关联订单ID',
    order_no    varchar(32)                                null comment '核销关联订单号',
    created_at  datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at  datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
)
    comment '用户优惠券表' charset = utf8mb4;

create index idx_user_coupons_order_id
    on user_coupons (order_id);

create index idx_user_coupons_status_expired
    on user_coupons (status, expired_at);

create index idx_user_coupons_template_id
    on user_coupons (template_id);

create index idx_user_coupons_user_id
    on user_coupons (user_id);

