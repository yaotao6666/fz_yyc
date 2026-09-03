create table user_behavior_events
(
    id         bigint unsigned auto_increment
        primary key,
    user_id    bigint unsigned not null comment '用户ID',
    openid     varchar(64)     null comment '微信OpenID',
    event_type varchar(32)     not null comment '事件类型: page_view=页面浏览 product_view=商品查看 submit_order=提交订单 pay_success=支付成功',
    page       varchar(64)     null comment '页面标识(如store_home/store_product)',
    product_id bigint unsigned null comment '关联商品ID(商品查看事件)',
    order_id   bigint unsigned null comment '关联订单ID(下单/支付事件)',
    source     varchar(32)     null comment '事件来源: scan=扫码 direct=直接进入',
    payload    json            null comment '事件附加数据JSON',
    created_at datetime(3)     null comment '创建时间',
    open_id    varchar(64)     null
)
    comment '用户行为事件表';

create index idx_user_behavior_events_created_at
    on user_behavior_events (created_at);

create index idx_user_behavior_events_event_type
    on user_behavior_events (event_type);

create index idx_user_behavior_events_open_id
    on user_behavior_events (open_id);

create index idx_user_behavior_events_openid
    on user_behavior_events (openid);

create index idx_user_behavior_events_order_id
    on user_behavior_events (order_id);

create index idx_user_behavior_events_product_id
    on user_behavior_events (product_id);

create index idx_user_behavior_events_user_id
    on user_behavior_events (user_id);

