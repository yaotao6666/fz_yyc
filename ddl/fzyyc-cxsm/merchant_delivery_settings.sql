create table merchant_delivery_settings
(
    id                   bigint unsigned auto_increment
        primary key,
    merchant_id          bigint unsigned                          not null,
    enabled              tinyint(1)     default 1                 not null,
    base_fee             decimal(10, 2) default 0.00              not null,
    free_delivery_amount decimal(10, 2) default 0.00              not null,
    max_distance         int unsigned   default '10'              not null,
    distance_rules       json                                     null,
    created_at           datetime       default CURRENT_TIMESTAMP not null,
    updated_at           datetime       default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint uk_merchant_delivery_settings_merchant_id
        unique (merchant_id),
    constraint fk_merchant_delivery_settings_merchant
        foreign key (merchant_id) references merchants (id)
            on delete cascade
)
    comment '商家配送设置表' collate = utf8mb4_unicode_ci;

