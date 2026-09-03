create table service_staff_audit_records
(
    id             bigint unsigned auto_increment
        primary key,
    staff_id       bigint unsigned                            not null comment '服务人员ID(service_staffs.id)',
    audit_type     tinyint unsigned default '1'               not null comment '审核类型:1=注册申请 2=信息变更 3=资质提交 4=状态变更',
    apply_type     tinyint unsigned default '1'               not null comment '申请来源:1=自注册 2=PC添加 3=端上变更',
    before_data    json                                       null comment '变更前字段快照',
    after_data     json                                       null comment '变更后字段快照(审核通过后回写)',
    qualifications json                                       null comment '资质材料URL列表',
    status         tinyint unsigned default '0'               not null comment '审核状态:0=待审 1=通过 2=驳回',
    reviewer_id    bigint unsigned                            null comment '审核人工员ID(merchant_staffs.id)',
    review_remark  varchar(256)                               null comment '审核备注',
    review_at      datetime                                   null comment '审核时间',
    created_at     datetime         default CURRENT_TIMESTAMP not null,
    updated_at     datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '服务人员审核记录表';

create index idx_staff_audit_staff_id
    on service_staff_audit_records (staff_id);

create index idx_staff_audit_status
    on service_staff_audit_records (status);

create index idx_staff_audit_type
    on service_staff_audit_records (audit_type);

