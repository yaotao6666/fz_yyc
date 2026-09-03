create table coupon_templates
(
    id               bigint unsigned auto_increment comment '券模板ID'
        primary key,
    name             varchar(64)                                not null comment '券名称',
    type             tinyint unsigned                           not null comment '券类型: 1=满减券 2=折扣券',
    threshold_amount decimal(10, 2)   default 0.00              not null comment '使用门槛金额(0=无门槛)',
    discount_amount  decimal(10, 2)   default 0.00              not null comment '满减面值(满减券)',
    discount_rate    decimal(3, 2)    default 0.00              not null comment '折扣率如0.90(折扣券)',
    total_count      int              default 0                 not null comment '发行总量(0=不限)',
    received_count   int              default 0                 not null comment '已领取数量',
    per_user_limit   int              default 1                 not null comment '每人限领数量',
    valid_type       tinyint unsigned                           not null comment '有效期类型: 1=固定期限 2=领取后N天有效',
    valid_start_at   datetime                                   null comment '固定期限开始时间',
    valid_end_at     datetime                                   null comment '固定期限结束时间',
    valid_days       int                                        null comment '领取后有效天数',
    apply_scope      tinyint unsigned default '1'               not null comment '适用范围: 1=全场 2=指定分类 3=指定商品',
    scope_ids        json                                       null comment '适用范围ID列表',
    status           tinyint unsigned default '1'               not null comment '状态: 1=启用 0=停用',
    remark           varchar(512)                               null comment '备注',
    created_at       datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at       datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
)
    comment '优惠券模板表' charset = utf8mb4;

