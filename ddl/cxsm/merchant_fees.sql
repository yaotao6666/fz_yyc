create table merchant_fees
(
    id          bigint unsigned auto_increment
        primary key,
    year        int unsigned                             not null,
    amount      decimal(10, 2) default 0.00              not null,
    status      varchar(16)                              not null,
    pay_time    datetime                                 null,
    free_reason varchar(256)                             null,
    created_at  datetime       default CURRENT_TIMESTAMP not null,
    updated_at  datetime       default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '商家年费表';

