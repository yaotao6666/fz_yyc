create table orders
(
    id                      bigint unsigned auto_increment
        primary key,
    order_no                varchar(32)                   not null comment '订单编号',
    user_id                 bigint unsigned               not null comment '下单用户ID',
    order_type              tinyint unsigned default '1'  not null comment '订单类型: 1=零售 2=租赁 3=康养上门 4=陪诊 5=科普体验 6=长护险服务',
    total_amount            decimal(10, 2)   default 0.00 not null comment '商品总金额(元)',
    delivery_fee            decimal(10, 2)   default 0.00 not null comment '配送费(元)',
    discount_amount         decimal(10, 2)   default 0.00 not null comment '优惠减免金额(元)',
    pay_amount              decimal(10, 2)   default 0.00 not null comment '实付金额(元)=总金额+配送费-优惠',
    total_deposit           decimal(10, 2)   default 0.00 not null comment '总押金(元)',
    deposit_status          tinyint unsigned default '0'  not null comment '押金状态: 0=无押金 1=已收 2=已退 3=部分扣除',
    deposit_refund_amount   decimal(10, 2)   default 0.00 not null comment '押金退还金额',
    deposit_deduct_amount   decimal(10, 2)   default 0.00 not null comment '押金扣除总额(损坏赔偿)',
    deposit_refunded_at     datetime                      null comment '押金退还时间',
    rental_returned_at      datetime                      null comment '租赁归还时间',
    rental_return_remark    varchar(256)                  null comment '归还备注(验机情况)',
    rental_end_at           datetime                      null comment '租赁到期时间(支付时=paid_at+租赁时长)',
    parent_order_id         bigint unsigned               null comment '续租关联原订单ID(0/空=普通订单)',
    renew_flag              tinyint unsigned default '0'  not null comment '是否续租单:1=续租 0=非',
    delivery_type           tinyint unsigned default '1'  not null comment '取餐方式: 1=配送 2=堂食 3=自提',
    delivery_distance       decimal(5, 2)                 null comment '配送距离(km,配送时用户选择)',
    delivery_address        varchar(256)                  null comment '收货地址(配送时填写)',
    contact_name            varchar(64)                   null comment '联系人姓名(配送时填写)',
    contact_phone           varchar(20)                   null comment '联系电话(配送时填写)',
    status                  tinyint unsigned default '1'  not null comment '订单状态: 1=待支付 2=已支付 3=已完成 4=已取消 5=退款中 6=已退款',
    biz_status              tinyint unsigned default '0'  not null comment '业务子状态(按order_type语义不同)',
    remark                  varchar(256)                  null comment '用户备注',
    verify_code             varchar(16)                   null comment '核销码(已支付订单出示给商家)',
    transaction_id          varchar(64)                   null comment '微信支付交易单号',
    paid_at                 datetime(3)                   null comment '支付完成时间',
    pay_notify_payload      json                          null comment '微信支付回调原始报文JSON',
    profit_sharing_status   tinyint unsigned default '0'  not null comment '分账状态:0未分账1分账中2分账成功3分账失败4已跳过',
    profit_sharing_amount   decimal(10, 2)   default 0.00 not null comment '分账总额(元)',
    profit_sharing_order_no varchar(64)                   null comment '微信分账单号',
    profit_sharing_at       datetime                      null comment '分账完成时间',
    profit_sharing_error    varchar(512)                  null comment '分账失败原因',
    completed_at            datetime(3)                   null comment '核销完成时间',
    completed_by_name       varchar(64)                   null comment '核销操作人姓名',
    cancelled_at            datetime(3)                   null comment '取消时间',
    refunded_at             datetime(3)                   null comment '退款完成时间(status=6时写入)',
    scheduled_at            datetime(3)                   null comment '预约服务开始时间(康养/陪诊/科普)',
    assigned_staff_id       bigint unsigned               null comment '指派的服务人员ID(service_staffs.id)',
    record_id               bigint unsigned               null comment '服务订单绑定的健康档案ID(服务单必填)',
    delivery_district       varchar(32)                   null comment '收货区县(服务订单区域匹配用)',
    actual_started_at       datetime(3)                   null comment '实际服务开始时间(签到时间)',
    actual_ended_at         datetime(3)                   null comment '实际服务结束时间(签退时间)',
    created_at              datetime(3)                   null comment '创建时间',
    updated_at              datetime(3)                   null comment '更新时间',
    pickup_point_id         bigint unsigned               null comment '自提点ID(自提订单)',
    pickup_point_name       varchar(64)                   null comment '自提点名称快照',
    pickup_point_address    varchar(256)                  null comment '自提点地址快照',
    pickup_point_lat        decimal(10, 6)                null comment '自提点纬度快照',
    pickup_point_lng        decimal(10, 6)                null comment '自提点经度快照',
    assigned_at             datetime                      null comment '指派/接单时间(超时未签到预警依据)',
    constraint idx_orders_order_no
        unique (order_no),
    constraint order_no
        unique (order_no),
    constraint uk_orders_order_no
        unique (order_no),
    constraint fk_orders_user
        foreign key (user_id) references users (id)
)
    comment '订单表';

create index idx_orders_assigned_staff
    on orders (assigned_staff_id, status);

create index idx_orders_created_at
    on orders (created_at);

create index idx_orders_deposit_status
    on orders (deposit_status);

create index idx_orders_order_type
    on orders (order_type, status);

create index idx_orders_paid_at
    on orders (paid_at);

create index idx_orders_pickup_point_id
    on orders (pickup_point_id);

create index idx_orders_record_id
    on orders (record_id);

create index idx_orders_rental_end_at
    on orders (rental_end_at);

create index idx_orders_scheduled_at
    on orders (scheduled_at);

create index idx_orders_status
    on orders (status);

create index idx_orders_user_id
    on orders (user_id);

create index idx_orders_verify_code
    on orders (verify_code);

