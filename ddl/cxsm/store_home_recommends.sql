create table store_home_recommends
(
    id          bigint unsigned auto_increment comment '推荐ID'
        primary key,
    merchant_id bigint unsigned                            not null comment '所属商家ID',
    product_id  bigint unsigned                            not null comment '目标商品/服务ID(products.id)',
    target_type tinyint unsigned default '1'               not null comment '推荐对象类型: 1=实物商品(产品类型1/2) 2=服务(产品类型3/4)',
    title       varchar(128)                               null comment '展示标题(空则用商品名)',
    sort        int unsigned     default '0'               not null comment '排序值(越小越靠前)',
    status      tinyint unsigned default '1'               not null comment '状态: 1=启用 0=禁用',
    created_at  datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at  datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
)
    comment '小程序首页推荐商品/服务表' charset = utf8mb4;

create index idx_shr_merchant
    on store_home_recommends (merchant_id);

create index idx_shr_status_sort
    on store_home_recommends (status, sort);

