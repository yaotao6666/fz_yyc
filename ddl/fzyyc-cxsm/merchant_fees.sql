create table merchant_fees
(
    id          bigint unsigned auto_increment
        primary key,
    merchant_id bigint unsigned                          not null,
    year        int unsigned                             not null,
    amount      decimal(10, 2) default 0.00              not null,
    status      varchar(16)                              not null,
    pay_time    datetime                                 null,
    free_reason varchar(256)                             null,
    created_at  datetime       default CURRENT_TIMESTAMP not null,
    updated_at  datetime       default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint fk_merchant_fees_merchant
        foreign key (merchant_id) references merchants (id)
            on delete cascade
)
    comment '商家年费表' collate = utf8mb4_unicode_ci;

create index idx_merchant_fees_merchant_id
    on merchant_fees (merchant_id);

