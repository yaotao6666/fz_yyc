create table profit_sharing_receivers
(
    id            bigint unsigned auto_increment
        primary key,
    merchant_id   bigint unsigned  default '1'                not null comment '商家ID(单商户恒为1)',
    receiver_type tinyint unsigned                            not null comment '接收方类型: 1=商户号 2=个人微信openid',
    name          varchar(64)                                 not null comment '显示名称',
    account       varchar(64)                                 not null comment '接收方账号(商户号 或 个人openid)',
    personal_name varchar(64)                                 null comment '个人真实姓名(个人类型时, 微信实名校验, 可空)',
    relation_type varchar(32)      default 'SERVICE_PROVIDER' not null comment '与特约商户关系, 默认SERVICE_PROVIDER',
    default_ratio decimal(5, 2)    default 0.00               not null comment '自动分账默认比例(%)',
    wechat_bound  tinyint unsigned default '0'                not null comment '微信接收方关系是否已建立: 0=未建立 1=已建立',
    wechat_error  varchar(256)                                null comment '微信建立接收方关系失败原因',
    status        tinyint unsigned default '1'                not null comment '状态: 1=启用 0=停用',
    sort          int unsigned     default '0'                not null comment '排序值(越小越靠前)',
    remark        varchar(256)                                null comment '备注',
    created_at    datetime         default CURRENT_TIMESTAMP  not null,
    updated_at    datetime         default CURRENT_TIMESTAMP  not null on update CURRENT_TIMESTAMP
)
    comment '分账接收方表';

create index idx_profit_sharing_receivers_account
    on profit_sharing_receivers (account);

create index idx_profit_sharing_receivers_merchant_id
    on profit_sharing_receivers (merchant_id);

