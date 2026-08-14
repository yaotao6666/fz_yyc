create table merchants
(
    id                     bigint unsigned auto_increment
        primary key,
    name                   varchar(128)                  not null comment '商家名称',
    logo                   varchar(512)                  null comment '商家Logo图片地址(七牛私有路径)',
    contact_name           varchar(64)                   null comment '联系人姓名',
    contact_phone          varchar(20)                   null comment '联系电话(用于用户端拨打退款)',
    contact_email          varchar(128)                  null comment '联系邮箱',
    address                varchar(256)                  null comment '商家地址',
    lat                    decimal(10, 6)                null comment '纬度',
    lng                    decimal(10, 6)                null comment '经度',
    business_category      varchar(64)                   null comment '经营类目',
    business_hours         varchar(64)                   null comment '营业时间描述',
    announcement           text                          null comment '商家公告',
    sub_mch_id             varchar(32)                   null comment '微信支付子商户号(线下进件后回填)',
    payment_config_status  tinyint unsigned default '0'  not null comment '支付配置状态: 0=未完成配置 1=已完成配置(已回填sub_mch_id)',
    status                 tinyint unsigned default '1'  not null comment '营业状态: 1=营业中 0=休息中',
    rating                 decimal(2, 1)    default 5.0  not null comment '商家评分(1.0-5.0)',
    sales_count            bigint unsigned  default '0'  not null comment '累计销量',
    created_at             datetime(3)                   null comment '创建时间',
    updated_at             datetime(3)                   null comment '更新时间',
    cover_image            varchar(512)                  null comment '商家背景/封面图地址(七牛私有路径)'
)
    comment '商家表' collate = utf8mb4_unicode_ci;

create index idx_merchants_location
    on merchants (lat, lng);

create index idx_merchants_status
    on merchants (status);

create index idx_merchants_sub_mch_id
    on merchants (sub_mch_id);
