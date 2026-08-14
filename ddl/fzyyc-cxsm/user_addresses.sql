create table user_addresses
(
    id         bigint unsigned auto_increment
        primary key,
    user_id    bigint unsigned                      not null,
    name       varchar(64)                          not null,
    phone      varchar(20)                          not null,
    province   varchar(32)                          null,
    city       varchar(32)                          null,
    district   varchar(32)                          null,
    address    varchar(256)                         not null,
    lat        decimal(10, 6)                       null,
    lng        decimal(10, 6)                       null,
    is_default tinyint(1) default 0                 not null,
    created_at datetime   default CURRENT_TIMESTAMP not null,
    updated_at datetime   default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint fk_user_addresses_user
        foreign key (user_id) references users (id)
            on delete cascade
)
    comment '用户收货地址表' collate = utf8mb4_unicode_ci;

create index idx_user_addresses_user_id
    on user_addresses (user_id);

