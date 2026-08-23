create table merchant_staffs
(
    id                    bigint unsigned auto_increment
        primary key,
    username              varchar(64)                                not null,
    password              varchar(128)                               not null,
    name                  varchar(64)                                null,
    phone                 varchar(20)                                null,
    openid                varchar(64)                                null,
    unionid               varchar(64)                                null comment '微信UnionID',
    wechat_bound_at       datetime                                   null,
    role                  varchar(32)      default 'staff'           not null,
    notify_enabled        tinyint(1)       default 1                 not null,
    browse_notify_enabled tinyint(1)       default 1                 not null,
    status                tinyint unsigned default '1'               not null,
    last_login_at         datetime                                   null,
    last_wechat_login_at  datetime                                   null,
    created_at            datetime         default CURRENT_TIMESTAMP not null,
    updated_at            datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint uk_merchant_staffs_username
        unique (username)
)
    comment '商家员工表' collate = utf8mb4_unicode_ci;

create index idx_merchant_staffs_openid
    on merchant_staffs (openid);

create index idx_merchant_staffs_phone
    on merchant_staffs (phone);

