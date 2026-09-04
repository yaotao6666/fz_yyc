create table printers
(
    id            bigint unsigned auto_increment comment '打印机ID'
        primary key,
    name          varchar(64)                                not null comment '打印机名称',
    type          tinyint unsigned default '1'               not null comment '类型: 1=飞鹅 2=通用云打印',
    feie_user     varchar(128)                               null comment '飞鹅云账号',
    feie_ukey     varchar(128)                               null comment '飞鹅云UKey(API密钥)',
    feie_sn       varchar(32)                                null comment '飞鹅云打印机编号',
    status        tinyint unsigned default '1'               not null comment '状态: 1=启用 0=禁用',
    auto_print    tinyint(1)       default 0                 not null comment '是否自动打印小票: 0=否 1=是',
    is_default    tinyint(1)       default 0                 not null comment '是否为默认打印机: 0=否 1=是',
    print_count   int unsigned     default '0'               not null comment '累计打印次数',
    last_print_at datetime                                   null comment '最后打印时间',
    remark        varchar(256)                               null comment '备注',
    created_at    datetime         default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at    datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间'
)
    comment '打印机表';

create index idx_printers_default
    on printers (is_default);

