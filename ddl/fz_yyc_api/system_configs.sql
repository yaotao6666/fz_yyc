create table system_configs
(
    id           bigint unsigned auto_increment comment '配置ID'
        primary key,
    config_key   varchar(100)                       not null comment '配置键',
    config_value text                               not null comment '配置值(JSON)',
    remark       varchar(255)                       null comment '备注',
    created_at   datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at   datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_config_key
        unique (config_key)
)
    comment '通用系统配置表(key-value+JSON+备注)' charset = utf8mb4;

