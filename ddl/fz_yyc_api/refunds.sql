create table refunds
(
    id            bigint unsigned auto_increment
        primary key,
    order_id      bigint unsigned                            not null,
    refund_no     varchar(32)                                not null,
    refund_amount decimal(10, 2)                             not null,
    refund_reason varchar(256)                               null,
    status        tinyint unsigned default '0'               not null,
    refund_id     varchar(64)                                null,
    refunded_at   datetime                                   null,
    created_at    datetime         default CURRENT_TIMESTAMP not null,
    updated_at    datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint uk_refunds_refund_no
        unique (refund_no),
    constraint fk_refunds_order
        foreign key (order_id) references orders (id)
)
    comment '退款记录表';

create index idx_refunds_order_id
    on refunds (order_id);

create index idx_refunds_status
    on refunds (status);

