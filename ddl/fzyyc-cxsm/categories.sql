create table categories
(
    id          bigint unsigned auto_increment
        primary key,
    merchant_id bigint unsigned                            not null,
    name        varchar(64)                                not null,
    sort        int unsigned     default '0'               not null,
    status      tinyint unsigned default '1'               not null,
    created_at  datetime         default CURRENT_TIMESTAMP not null,
    updated_at  datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint fk_categories_merchant
        foreign key (merchant_id) references merchants (id)
            on delete cascade
)
    comment '商品分类表' collate = utf8mb4_unicode_ci;

create index idx_categories_merchant_id
    on categories (merchant_id);

create index idx_categories_sort
    on categories (merchant_id, sort);

