create table health_records
(
    id                bigint unsigned auto_increment
        primary key,
    user_id           bigint unsigned                            not null comment '用户ID',
    relation          tinyint unsigned default '1'               not null comment '与账号关系:1本人2父母3其他亲属',
    real_name         varchar(64)                                null comment '真实姓名',
    gender            tinyint unsigned default '0'               not null comment '性别:1男2女',
    birth_date        varchar(16)                                null comment '出生日期',
    id_card           varchar(32)                                null comment '身份证号',
    phone             varchar(20)                                null comment '联系电话',
    emergency_contact varchar(64)                                null comment '紧急联系人',
    emergency_phone   varchar(20)                                null comment '紧急联系电话',
    address           varchar(256)                               null comment '常住地址',
    height_cm         decimal(5, 1)                              null comment '身高(cm)',
    weight_kg         decimal(5, 1)                              null comment '体重(kg)',
    blood_type        varchar(8)                                 null comment '血型',
    past_history      json                                       null comment '既往病史数组',
    allergy_history   json                                       null comment '过敏史数组',
    family_history    json                                       null comment '家族病史数组',
    surgery_history   json                                       null comment '手术史数组',
    medication_list   json                                       null comment '长期用药数组',
    chronic_tags      json                                       null comment '慢病标签数组',
    smoking           varchar(32)                                null comment '吸烟情况',
    drinking          varchar(32)                                null comment '饮酒情况',
    assessment_level  varchar(32)                                null comment '最近一次评估等级',
    remark            varchar(512)                               null comment '备注',
    status            tinyint unsigned default '1'               not null comment '0未建档1正常2已归档',
    created_at        datetime         default CURRENT_TIMESTAMP not null,
    updated_at        datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '居民健康档案表';

create index idx_health_records_user_id
    on health_records (user_id);

