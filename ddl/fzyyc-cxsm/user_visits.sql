create table user_visits
(
    id          bigint unsigned auto_increment
        primary key,
    user_id     bigint unsigned not null comment '用户ID',
    merchant_id bigint unsigned not null comment '商家ID',
    openid      varchar(64)     null comment '微信OpenID',
    visit_time  datetime(3)     null comment '访问时间',
    source      varchar(32)     null comment '访问来源: scan=扫码 direct=直接进入',
    constraint fk_user_visits_merchant
        foreign key (merchant_id) references merchants (id)
            on delete cascade,
    constraint fk_user_visits_user
        foreign key (user_id) references users (id)
            on delete cascade
)
    comment '用户访问记录表' collate = utf8mb4_unicode_ci;

create index idx_user_visits_merchant_id
    on user_visits (merchant_id);

create index idx_user_visits_open_id
    on user_visits (openid);

create index idx_user_visits_openid
    on user_visits (openid);

create index idx_user_visits_user_id
    on user_visits (user_id);

