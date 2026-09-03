create table service_records
(
    id                bigint unsigned auto_increment comment '服务记录ID'
        primary key,
    order_id          bigint unsigned                            not null comment '订单ID',
    staff_id          bigint unsigned                            not null comment '服务人员ID',
    start_time        datetime                                   null comment '签到时间(orders快照)',
    end_time          datetime                                   null comment '签退时间(orders快照)',
    gps_track_url     varchar(512)                               null comment '轨迹聚合文件URL(可选)',
    audio_url         varchar(512)                               null comment '服务录音URL(七牛)',
    audio_uploaded_at datetime                                   null comment '录音上传时间(30天清理依据)',
    audio_deleted_at  datetime                                   null comment '录音删除标记时间',
    sos_triggered     tinyint unsigned default '0'               not null comment '是否触发SOS: 0=否 1=是',
    status            tinyint unsigned default '1'               not null comment '状态: 1=正常 2=异常',
    created_at        datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at        datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_service_records_order_id
        unique (order_id)
)
    comment '服务记录表' charset = utf8mb4;

create index idx_service_records_staff_id
    on service_records (staff_id);

