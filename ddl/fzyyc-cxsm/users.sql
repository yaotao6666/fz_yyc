create table users
(
    id             bigint unsigned auto_increment
        primary key,
    openid         varchar(64)                   null comment '微信OpenID(用户唯一标识)',
    union_id       varchar(64)                   null comment '微信UnionID(跨小程序唯一)',
    nickname       varchar(64)                   null comment '用户昵称(默认微信用户)',
    avatar         varchar(512)                  null comment '用户头像URL',
    phone          varchar(20)                   null comment '用户手机号',
    status         tinyint unsigned default '1'  not null comment '状态: 1=正常 0=禁用',
    created_at     datetime(3)                   null comment '创建时间',
    updated_at     datetime(3)                   null comment '更新时间',
    first_visit_at datetime(3)                   null comment '首次访问时间',
    last_visit_at  datetime(3)                   null comment '最后访问时间',
    visit_count    bigint unsigned  default '1'  not null comment '累计访问次数',
    has_ordered    tinyint(1)       default 0    not null comment '是否下过单: true=是 false=否',
    total_orders   bigint unsigned  default '0'  not null comment '累计订单数',
    total_spent    decimal(10, 2)   default 0.00 not null comment '累计消费金额(元)',
    has_paid       tinyint(1)       default 0    not null comment '是否完成过支付: true=是 false=否',
    first_paid_at  datetime(3)                   null comment '首次支付时间',
    constraint idx_users_open_id
        unique (openid),
    constraint openid
        unique (openid),
    constraint uk_users_openid
        unique (openid)
)
    comment 'C端用户表' collate = utf8mb4_unicode_ci;

create index idx_users_phone
    on users (phone);

create index idx_users_union_id
    on users (union_id);

