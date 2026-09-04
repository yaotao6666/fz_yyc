create table service_staffs
(
    id             bigint unsigned auto_increment comment '服务人员ID'
        primary key,
    username       varchar(64)                  not null comment '登录用户名',
    password       varchar(128)                 not null comment '加密密码(不返回)',
    name           varchar(64)                  null comment '姓名',
    phone          varchar(20)                  null comment '手机号',
    openid         varchar(64)                  null comment '微信OpenID(用于快捷登录)',
    avatar         varchar(512)                 null comment '头像URL',
    qualifications json                         null comment '资质材料列表JSON[{type,name,url}]',
    service_region varchar(256)                 null comment '服务区域(区县,逗号分隔,空=不限)',
    quality_score  decimal(3, 1)    default 5.0 not null comment '服务质量分(阶段四写入)',
    status         tinyint unsigned default '0' not null comment '状态: 0=待审核 1=启用 2=禁用',
    audit_status   tinyint unsigned default '0' not null comment '审核状态: 0=无/已通过 1=待审核(与status解耦)',
    pending_fields json                         null comment '审核中待变更字段快照',
    last_login_at  datetime(3)                  null comment '最后登录时间',
    created_at     datetime(3)                  null comment '创建时间',
    updated_at     datetime(3)                  null comment '更新时间'
);

create index idx_service_staffs_open_id
    on service_staffs (openid);

