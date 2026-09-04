create table categories
(
    id            bigint unsigned auto_increment
        primary key,
    name          varchar(64)                                not null,
    category_type tinyint unsigned default '1'               not null comment '分类类型: 1=商品分类 2=服务分类',
    parent_id     bigint unsigned                            null comment '父分类ID(空=一级)',
    level         tinyint unsigned default '1'               not null comment '层级: 1=一级 2=二级 3=三级',
    sort          int unsigned     default '0'               not null,
    status        tinyint unsigned default '1'               not null,
    created_at    datetime         default CURRENT_TIMESTAMP not null,
    updated_at    datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '商品分类表';

create index idx_categories_parent
    on categories (parent_id);

create index idx_categories_sort
    on categories (sort);

