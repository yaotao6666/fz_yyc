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
    min_order_amount       decimal(10, 2)   default 0.00 not null comment '最低起送金额',
    takeout_enabled        tinyint(1)       default 1    not null comment '是否开启配送: true=开启 false=关闭',
    dine_in_enabled        tinyint(1)       default 1    not null comment '是否开启堂食: true=开启 false=关闭',
    sub_mch_id             varchar(32)                   null comment '微信支付子商户号(线下进件后回填)',
    sub_mch_status         tinyint unsigned default '0'  not null,
    applyment_status       tinyint unsigned default '0'  not null,
    audit_status           tinyint unsigned default '0'  not null,
    audit_remark           varchar(256)                  null,
    status                 tinyint unsigned default '1'  not null comment '营业状态: 1=营业中 0=休息中',
    rating                 decimal(2, 1)    default 5.0  not null comment '商家评分(1.0-5.0)',
    sales_count            bigint unsigned  default '0'  not null comment '累计销量',
    qrcode_url             varchar(512)                  null,
    created_at             datetime(3)                   null comment '创建时间',
    updated_at             datetime(3)                   null comment '更新时间',
    cover_image            varchar(512)                  null comment '商家背景/封面图地址(七牛私有路径)',
    payment_config_status  tinyint unsigned default '0'  not null comment '支付配置状态: 0=未完成配置 1=已完成配置(已回填sub_mch_id)',
    profit_sharing_enabled tinyint(1)       default 0    not null comment '是否开启自动分账',
    qr_code_url            varchar(512)                  null comment '商家小程序码图片地址',
    pickup_enabled_2       tinyint(1)       default 1    not null,
    pickup_enabled         tinyint(1)       default 1    not null comment '是否开启自提: true=开启 false=关闭'
)
    comment '商家表';

create index idx_merchants_audit_status
    on merchants (audit_status);

create index idx_merchants_location
    on merchants (lat, lng);

create index idx_merchants_status
    on merchants (status);

create index idx_merchants_sub_mch_id
    on merchants (sub_mch_id);

