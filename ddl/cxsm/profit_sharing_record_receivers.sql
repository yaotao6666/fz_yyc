create table profit_sharing_record_receivers
(
    id            bigint unsigned auto_increment
        primary key,
    record_id     bigint unsigned                          not null comment '分账单ID',
    receiver_id   bigint unsigned                          null comment '接收方ID',
    receiver_type tinyint unsigned                         not null comment '接收方类型快照: 1=商户号 2=个人微信openid',
    receiver_name varchar(64)                              not null comment '接收方名称快照',
    account       varchar(64)                              not null comment '接收方账号快照',
    amount        decimal(10, 2) default 0.00              not null comment '分账金额(元)',
    result_status varchar(32)                              null comment '微信分账结果: PROCESSING/SUCCESS/CLOSED/FAILED/FINISHED',
    detail_id     varchar(64)                              null comment '微信分账明细单号',
    fail_reason   varchar(128)                             null comment '分账失败原因',
    finish_time   datetime                                 null comment '分账完成时间',
    created_at    datetime       default CURRENT_TIMESTAMP not null,
    updated_at    datetime       default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '分账单明细表';

create index idx_profit_sharing_record_receivers_receiver_id
    on profit_sharing_record_receivers (receiver_id);

create index idx_profit_sharing_record_receivers_record_id
    on profit_sharing_record_receivers (record_id);

