create table order_items
(
    id                bigint unsigned auto_increment
        primary key,
    order_id          bigint unsigned                            not null,
    product_id        bigint unsigned                            not null,
    product_name      varchar(128)                               not null,
    image             varchar(512)                               null,
    price             decimal(10, 2)                             not null,
    quantity          int unsigned     default '1'               not null,
    spec_info         json                                       null,
    subtotal          decimal(10, 2)                             not null,
    sale_type         tinyint unsigned default '1'               not null comment '销售类型快照: 1=一口价 2=租赁',
    rental_unit       tinyint unsigned default '0'               not null comment '租赁计费周期快照',
    rental_duration   int unsigned     default '0'               not null comment '租赁时长(下单时选择)',
    unit_rental_price decimal(10, 2)   default 0.00              not null comment '单位租金快照',
    rental_subtotal   decimal(10, 2)   default 0.00              not null comment '租金小计=单价×时长',
    deposit           decimal(10, 2)   default 0.00              not null comment '单商品押金快照',
    deposit_deduct    decimal(10, 2)   default 0.00              not null comment '押金扣除金额(损坏赔偿)',
    created_at        datetime         default CURRENT_TIMESTAMP not null,
    constraint fk_order_items_order
        foreign key (order_id) references orders (id)
            on delete cascade,
    constraint fk_order_items_product
        foreign key (product_id) references products (id)
)
    comment '订单商品表';

create index idx_order_items_order_id
    on order_items (order_id);

create index idx_order_items_product_id
    on order_items (product_id);

