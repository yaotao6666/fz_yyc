create table mini_program_banners
(
    id          bigint unsigned auto_increment comment '轮播图ID'
        primary key,
    merchant_id bigint unsigned                            not null comment '所属商家ID',
    title       varchar(128)                               null comment '轮播图标题(仅后台识别)',
    image       varchar(512)                               not null comment '轮播图图片URL(七牛私有路径)',
    link_type   varchar(16)      default 'none'            not null comment '跳转类型: none=无跳转 product=商品详情 category=分类页 url=外部链接',
    link_value  varchar(256)                               null comment '跳转目标值: product=product_id category=空 url=链接',
    sort        int unsigned     default '0'               not null comment '排序值(越小越靠前)',
    status      tinyint unsigned default '1'               not null comment '状态: 1=启用 0=禁用',
    created_at  datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at  datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
)
    comment '小程序首页轮播图表' charset = utf8mb4;

create index idx_mpb_merchant
    on mini_program_banners (merchant_id);

create index idx_mpb_status_sort
    on mini_program_banners (status, sort);

