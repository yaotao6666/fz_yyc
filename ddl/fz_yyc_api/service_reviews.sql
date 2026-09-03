create table service_reviews
(
    id                 bigint unsigned auto_increment comment '评价ID'
        primary key,
    order_id           bigint unsigned                            not null comment '订单ID(一单一评)',
    user_id            bigint unsigned                            not null comment '下单用户ID',
    staff_id           bigint unsigned                            not null comment '被评价服务人员ID',
    score              tinyint unsigned                           not null comment '总体评分 1-5',
    attitude_score     tinyint unsigned                           not null comment '服务态度分 1-5',
    professional_score tinyint unsigned                           not null comment '专业技能分 1-5',
    punctual_score     tinyint unsigned                           not null comment '准时守约分 1-5',
    content            varchar(512)                               null comment '评价内容',
    images             json                                       null comment '评价图片URL列表',
    status             tinyint unsigned default '1'               not null comment '状态: 1=正常展示 0=后台隐藏',
    created_at         datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at         datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_service_reviews_order_id
        unique (order_id)
)
    comment '服务评价表' charset = utf8mb4;

create index idx_service_reviews_created_at
    on service_reviews (created_at);

create index idx_service_reviews_staff_id
    on service_reviews (staff_id);

create index idx_service_reviews_status
    on service_reviews (status);

create index idx_service_reviews_user_id
    on service_reviews (user_id);

