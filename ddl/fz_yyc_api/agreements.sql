create table agreements
(
    id           bigint unsigned auto_increment comment '协议ID'
        primary key,
    type         tinyint unsigned                           not null comment '协议类型: 1=用户协议 2=隐私政策 3=录音/定位授权协议',
    title        varchar(128)                               not null comment '协议标题',
    content      text                                       null comment '协议正文(富文本)',
    version      varchar(32)                                not null comment '版本号(同类型递增,如v1.2)',
    status       tinyint unsigned default '0'               not null comment '状态: 1=已发布(当前生效) 0=草稿/停用',
    published_at datetime                                   null comment '发布时间',
    created_at   datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at   datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
)
    comment '协议表' charset = utf8mb4;

create index idx_agreements_status
    on agreements (status);

create index idx_agreements_type
    on agreements (type);

