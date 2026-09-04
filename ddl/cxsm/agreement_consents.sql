create table agreement_consents
(
    id           bigint unsigned auto_increment comment '留痕ID'
        primary key,
    agreement_id bigint unsigned                            not null comment '协议ID',
    user_type    tinyint unsigned default '1'               not null comment '用户类型: 1=C端用户 2=服务人员',
    user_id      bigint unsigned                            not null comment '用户ID',
    version      varchar(32)                                not null comment '同意时协议版本号快照',
    created_at   datetime         default CURRENT_TIMESTAMP not null comment '同意时间'
)
    comment '协议同意留痕表' charset = utf8mb4;

create index idx_consents_agreement_id
    on agreement_consents (agreement_id);

create index idx_consents_user
    on agreement_consents (user_type, user_id);

