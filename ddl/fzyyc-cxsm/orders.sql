create table orders
(
    id                      bigint unsigned auto_increment
        primary key,
    order_no                varchar(32)                   not null comment '订单编号',
    user_id                 bigint unsigned               not null comment '下单用户ID',
    total_amount            decimal(10, 2)   default 0.00 not null comment '商品总金额(元)',
    delivery_fee            decimal(10, 2)   default 0.00 not null comment '配送费(元)',
    discount_amount         decimal(10, 2)   default 0.00 not null comment '优惠减免金额(元)',
    pay_amount              decimal(10, 2)   default 0.00 not null comment '实付金额(元)=总金额+配送费-优惠',
    delivery_type           tinyint unsigned default '1'  not null comment '取餐方式: 1=配送 2=堂食 3=自提',
    delivery_distance       decimal(5, 2)                 null comment '配送距离(km,配送时用户选择)',
    delivery_address        varchar(256)                  null comment '收货地址(配送时填写)',
    contact_name            varchar(64)                   null comment '联系人姓名(配送时填写)',
    contact_phone           varchar(20)                   null comment '联系电话(配送时填写)',
    pickup_point_id         bigint unsigned               null comment '自提点ID(自提订单)',
    pickup_point_name       varchar(64)                   null comment '自提点名称快照',
    pickup_point_address    varchar(256)                  null comment '自提点地址快照',
    pickup_point_lat        decimal(10, 6)                null comment '自提点纬度快照',
    pickup_point_lng        decimal(10, 6)                null comment '自提点经度快照',
    status                  tinyint unsigned default '1'  not null comment '订单状态: 1=待支付 2=已支付 3=已完成 4=已取消 5=退款中 6=已退款',
    remark                  varchar(256)                  null comment '用户备注',
    verify_code             varchar(16)                   null comment '核销码(已支付订单出示给商家)',
    transaction_id          varchar(64)                   null comment '微信支付交易单号',
    paid_at                 datetime(3)                   null comment '支付完成时间',
    pay_notify_payload      json                          null comment '微信支付回调原始报文JSON',
    completed_at            datetime(3)                   null comment '核销完成时间',
    completed_by_name       varchar(64)                   null comment '核销操作人姓名',
    cancelled_at            datetime(3)                   null comment '取消时间',
    refunded_at             datetime(3)                   null comment '退款完成时间(status=6时写入)',
    created_at              datetime(3)                   null comment '创建时间',
    updated_at              datetime(3)                   null comment '更新时间',
    constraint idx_orders_order_no
        unique (order_no),
    constraint order_no
        unique (order_no),
    constraint uk_orders_order_no
        unique (order_no),
    constraint fk_orders_user
        foreign key (user_id) references users (id)
)
    comment '订单表' collate = utf8mb4_unicode_ci;

create index idx_orders_created_at
    on orders (created_at);

create index idx_orders_paid_at
    on orders (paid_at);

create index idx_orders_pickup_point_id
    on orders (pickup_point_id);

create index idx_orders_status
    on orders (status);

create index idx_orders_user_id
    on orders (user_id);

create index idx_orders_verify_code
    on orders (verify_code);

