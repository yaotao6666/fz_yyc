create table service_location_tracks
(
    id          bigint unsigned auto_increment comment '轨迹点ID'
        primary key,
    order_id    bigint unsigned                    not null comment '订单ID',
    staff_id    bigint unsigned                    not null comment '服务人员ID',
    lat         decimal(10, 6)                     not null comment '纬度',
    lng         decimal(10, 6)                     not null comment '经度',
    reported_at datetime                           not null comment '上报时间',
    created_at  datetime default CURRENT_TIMESTAMP not null comment '创建时间'
)
    comment '服务轨迹点表' charset = utf8mb4;

create index idx_tracks_order_id
    on service_location_tracks (order_id);

create index idx_tracks_reported_at
    on service_location_tracks (reported_at);

create index idx_tracks_staff_id
    on service_location_tracks (staff_id);

