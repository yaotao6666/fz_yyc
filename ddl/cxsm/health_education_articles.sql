create table health_education_articles
(
    id          bigint unsigned auto_increment
        primary key,
    title       varchar(128)                               not null comment '标题',
    category_id bigint unsigned  default '0'               not null comment '分类ID（关联health_education_categories）',
    category    varchar(32)                                null comment '分类',
    cover       varchar(512)                               null comment '封面图URL',
    content     text                                       null comment '正文',
    tags        json                                       null comment '定向慢病标签数组',
    status      tinyint unsigned default '0'               not null comment '状态:0草稿1发布',
    publish_at  datetime                                   null comment '发布时间',
    views       int unsigned     default '0'               not null comment '浏览量',
    created_at  datetime         default CURRENT_TIMESTAMP not null,
    updated_at  datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '健康宣教内容表';

create index idx_hea_category_id
    on health_education_articles (category_id);

