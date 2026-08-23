create table order_items
(
    id           bigint unsigned auto_increment
        primary key,
    order_id     bigint unsigned                        not null,
    product_id   bigint unsigned                        not null,
    product_name varchar(128)                           not null,
    image        varchar(512)                           null,
    price        decimal(10, 2)                         not null,
    quantity     int unsigned default '1'               not null,
    spec_info    json                                   null,
    subtotal     decimal(10, 2)                         not null,
    created_at   datetime     default CURRENT_TIMESTAMP not null,
    constraint fk_order_items_order
        foreign key (order_id) references orders (id)
            on delete cascade,
    constraint fk_order_items_product
        foreign key (product_id) references products (id)
)
    comment '订单商品表' collate = utf8mb4_unicode_ci;

create index idx_order_items_order_id
    on order_items (order_id);

create index idx_order_items_product_id
    on order_items (product_id);

