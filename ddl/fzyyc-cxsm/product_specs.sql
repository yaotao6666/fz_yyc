create table product_specs
(
    id         bigint unsigned auto_increment
        primary key,
    product_id bigint unsigned                    not null,
    name       varchar(64)                        not null,
    options    json                               null,
    created_at datetime default CURRENT_TIMESTAMP not null,
    updated_at datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint fk_product_specs_product
        foreign key (product_id) references products (id)
            on delete cascade
)
    comment '商品规格表' collate = utf8mb4_unicode_ci;

create index idx_product_specs_product_id
    on product_specs (product_id);

