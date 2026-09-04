create table products
(
    id                  bigint unsigned auto_increment
        primary key,
    category_id         bigint unsigned                            null,
    name                varchar(128)                               not null,
    description         text                                       null,
    images              json                                       null,
    price               decimal(10, 2)                             not null,
    original_price      decimal(10, 2)                             null,
    stock               int unsigned     default '0'               not null,
    unit                varchar(16)      default '份'              not null,
    product_type        tinyint unsigned default '1'               not null comment '商品类型: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务 5=科普活动 6=长护险服务',
    service_content     json                                       null comment '服务型商品的内容描述JSON',
    sale_type           tinyint unsigned default '1'               not null comment '销售类型: 1=一口价 2=租赁',
    rental_unit         tinyint unsigned default '0'               not null comment '租赁计费周期: 0=非租赁 1=按天 2=按周 3=按月',
    rental_price        decimal(10, 2)   default 0.00              not null comment '单位租金(元)',
    deposit             decimal(10, 2)   default 0.00              not null comment '押金(元)',
    max_rental_duration int unsigned     default '0'               not null comment '最大租赁时长(0=不限)',
    sales               int unsigned     default '0'               not null,
    sort                int unsigned     default '0'               not null,
    status              tinyint unsigned default '1'               not null,
    deleted_at          datetime                                   null,
    created_at          datetime         default CURRENT_TIMESTAMP not null,
    updated_at          datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint fk_products_category
        foreign key (category_id) references categories (id)
            on delete set null
)
    comment '商品表';

create index idx_products_category_id
    on products (category_id);

create index idx_products_product_type
    on products (product_type, status);

create index idx_products_sale_type
    on products (sale_type);

create index idx_products_sales
    on products (sales);

create index idx_products_sort
    on products (sort);

create index idx_products_status
    on products (status);

