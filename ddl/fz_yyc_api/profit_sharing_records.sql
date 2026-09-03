create table profit_sharing_records
(
    id                 bigint unsigned auto_increment
        primary key,
    merchant_id        bigint unsigned  default '1'               not null comment '商家ID(单商户恒为1)',
    order_id           bigint unsigned                            not null comment '订单ID',
    order_no           varchar(32)                                not null comment '订单编号',
    sp_mchid           varchar(32)                                null comment '服务商商户号',
    sub_mchid          varchar(32)                                null comment '特约商户号(分账出资方)',
    appid              varchar(64)                                null comment '分账请求使用的appid(特约商户主体小程序)',
    transaction_id     varchar(64)                                null comment '微信支付交易单号',
    out_order_no       varchar(64)                                not null comment '微信分账单号',
    total_amount       decimal(10, 2)   default 0.00              not null comment '订单实付金额(元)',
    total_share_amount decimal(10, 2)   default 0.00              not null comment '本次分账总额(元)',
    status             tinyint unsigned default '0'               not null comment '状态: 0=待分账 1=分账中 2=分账成功 3=分账失败 4=已跳过',
    share_time         datetime                                   null comment '分账完成时间',
    error_message      varchar(512)                               null comment '失败原因',
    created_at         datetime         default CURRENT_TIMESTAMP not null,
    updated_at         datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '分账单表';

create index idx_profit_sharing_records_merchant_id
    on profit_sharing_records (merchant_id);

create index idx_profit_sharing_records_order_id
    on profit_sharing_records (order_id);

create index idx_profit_sharing_records_order_no
    on profit_sharing_records (order_no);

create index idx_profit_sharing_records_out_order_no
    on profit_sharing_records (out_order_no);

create index idx_profit_sharing_records_status
    on profit_sharing_records (status);

