create table merchant_rates
(
    id             bigint unsigned auto_increment
        primary key,
    merchant_id    bigint unsigned                            not null,
    rate_type      varchar(32)                                not null,
    rate           decimal(5, 4)                              not null,
    effective_time datetime                                   not null,
    expire_time    datetime                                   null,
    remark         varchar(256)                               null,
    status         tinyint unsigned default '1'               not null,
    created_at     datetime         default CURRENT_TIMESTAMP not null,
    updated_at     datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint fk_merchant_rates_merchant
        foreign key (merchant_id) references merchants (id)
            on delete cascade
)
    comment '商家手续费率表' collate = utf8mb4_unicode_ci;

create index idx_merchant_rates_merchant_id
    on merchant_rates (merchant_id);

create index idx_merchant_rates_status
    on merchant_rates (status);

