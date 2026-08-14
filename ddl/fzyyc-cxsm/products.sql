create table products
(
    id             bigint unsigned auto_increment
        primary key,
    merchant_id    bigint unsigned                            not null,
    category_id    bigint unsigned                            null,
    name           varchar(128)                               not null,
    description    text                                       null,
    images         json                                       null,
    price          decimal(10, 2)                             not null,
    original_price decimal(10, 2)                             null,
    stock          int unsigned     default '0'               not null,
    unit           varchar(16)      default '份'              not null,
    sales          int unsigned     default '0'               not null,
    sort           int unsigned     default '0'               not null,
    status         tinyint unsigned default '1'               not null,
    deleted_at     datetime                                   null,
    created_at     datetime         default CURRENT_TIMESTAMP not null,
    updated_at     datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint fk_products_category
        foreign key (category_id) references categories (id)
            on delete set null,
    constraint fk_products_merchant
        foreign key (merchant_id) references merchants (id)
            on delete cascade
)
    comment '商品表' collate = utf8mb4_unicode_ci;

create index idx_products_category_id
    on products (category_id);

create index idx_products_merchant_id
    on products (merchant_id);

create index idx_products_sales
    on products (merchant_id, sales);

create index idx_products_sort
    on products (merchant_id, sort);

create index idx_products_status
    on products (status);

