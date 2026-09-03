create table service_alert_events
(
    id            bigint unsigned auto_increment comment '预警事件ID'
        primary key,
    order_id      bigint unsigned                            null comment '关联订单ID(可空)',
    staff_id      bigint unsigned                            null comment '服务人员ID(订单级预警可空)',
    alert_type    tinyint unsigned                           not null comment '预警类型: 1=SOS求助 2=服务超时未结束',
    lat           decimal(10, 6)                             null comment '触发位置纬度',
    lng           decimal(10, 6)                             null comment '触发位置经度',
    address       varchar(256)                               null comment '触发位置地址',
    summary       varchar(256)                               null comment '预警摘要(订单级预警触发说明)',
    status        tinyint unsigned default '1'               not null comment '状态: 1=待处理 2=处理中 3=已处理',
    handler_id    bigint unsigned                            null comment '处理人ID(merchant_staffs.id)',
    handler_name  varchar(64)                                null comment '处理人姓名快照',
    handle_remark varchar(512)                               null comment '处理备注留痕',
    handled_at    datetime                                   null comment '处理时间',
    created_at    datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at    datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
)
    comment '服务预警事件表' charset = utf8mb4;

create index idx_alert_created_at
    on service_alert_events (created_at);

create index idx_alert_order_id
    on service_alert_events (order_id);

create index idx_alert_staff_id
    on service_alert_events (staff_id);

create index idx_alert_status
    on service_alert_events (status);

