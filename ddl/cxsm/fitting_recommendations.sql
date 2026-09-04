create table fitting_recommendations
(
    id                   bigint unsigned auto_increment
        primary key,
    user_id              bigint unsigned                            not null comment '居民用户ID',
    record_id            bigint unsigned                            null comment '关联档案ID（多档案）',
    assessment_id        bigint unsigned                            null comment '关联评估记录ID',
    symptom_desc         varchar(512)                               null comment '症状/需求描述',
    fitting_result       varchar(512)                               null comment '适配结论',
    recommended_products json                                       null comment '推荐商品快照[{product_id,name,reason,sale_type}]',
    staff_id             bigint unsigned                            null comment '生成建议的服务人员ID',
    status               tinyint unsigned default '0'               not null comment '状态:0草稿1已确认2已下单',
    order_id             bigint unsigned                            null comment '关联订单ID',
    created_at           datetime         default CURRENT_TIMESTAMP not null,
    updated_at           datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '康复辅具适配建议表';

create index idx_fitting_recommendations_record_id
    on fitting_recommendations (record_id);

create index idx_fitting_recommendations_staff_id
    on fitting_recommendations (staff_id);

create index idx_fitting_recommendations_user_id
    on fitting_recommendations (user_id);

