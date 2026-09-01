# 预警中心扩展（订单/业务超时预警）Spec

## Why
现有预警中心（service_alert_events）仅覆盖服务人员维度：SOS 求助（type=1）、服务超时未结束（type=2）。缺少订单交易链路关键节点的超时预警：实物已支付未核销、服务已支付未指派、陪诊已指派未完成、已指派未签到、租赁逾期未归还、退款卡在处理中。这些环节逾期直接影响营收与用户体验，需纳入统一预警中心并支持阈值可配置。

## What Changes
- 扩展现有预警中心（service_alert_events 表 + web-admin 预警中心页），新增 6 类订单/业务超时预警（type=3~8）
- 新建单行预警配置表 `alert_settings`，6 个阈值 + 总开关，在预警中心页"预警设置"弹窗维护
- `orders` 表新增 `assigned_at`（指派时间），支撑"已指派超时未签到"口径
- `service_alert_events.staff_id` 改为可空（订单级预警无服务人员）、新增 `summary` 摘要列
- 新增订单级超时扫描定时任务（5 分钟粒度，启动即扫），幂等：同一 order_id + alert_type 仅保留一条未闭环（status≠3）预警
- **BREAKING**: `service_alert_events.staff_id` 由 NOT NULL 改为可空；`orders` 新增 `assigned_at` 列

## Impact
- Affected specs: 服务过程安全（阶段三）、订单交易链路
- Affected code:
  - `server/migrations/016_alert_center_extend.sql`（新增）
  - `server/internal/models/models.go`（ServiceAlertEvent / Order / AlertSettings）
  - `server/internal/utils/constants.go`（预警类型 3~8 常量与默认阈值）
  - `server/internal/handlers/merchant/alert_center.go`（类型文案、staff 空值处理、预警设置接口）
  - `server/internal/handlers/merchant/order_dispatch.go`、`server/internal/handlers/service_staff/workorder.go`（派单/接单写入 assigned_at）
  - `server/internal/tasks/order_timeout.go`（新增）+ `server/cmd/server/main.go`（路由 + 任务注册）
  - `web-admin/src/api/safety.ts`、`web-admin/src/views/safety/AlertEventsView.vue`

## ADDED Requirements

### Requirement: 预警配置（alert_settings）
系统 SHALL 提供单行预警配置（默认值兜底，初始化时自动创建默认行），包含：
- `goods_unverified_hours` 实物超时未核销（小时，默认 24）
- `service_unassigned_hours` 服务超时未指派（小时，默认 2）
- `escort_unfinished_minutes` 陪诊超时未完成（分钟，默认 120）
- `service_unstarted_minutes` 指派超时未签到（分钟，默认 30）
- `rental_overdue_hours` 租赁逾期未归还（小时，默认 24）
- `refund_stuck_hours` 退款卡在处理中（小时，默认 24）
- `enabled` 总开关（默认开启；关闭时订单级扫描跳过）

管理员 SHALL 能在预警中心页"预警设置"弹窗读取/修改；修改时校验数值范围（分钟 1~1440、小时 1~720）。

#### Scenario: 读取配置
- **WHEN** 管理员打开"预警设置"
- **THEN** 返回当前生效阈值；表为空时自动创建默认行

#### Scenario: 修改配置
- **WHEN** 管理员保存新阈值
- **THEN** 阈值持久化并即时影响后续扫描

### Requirement: 实物订单超时未核销预警（type=3）
系统 SHALL 对辅具零售订单（order_type=1）、已支付（status=2）、未核销（completed_at IS NULL）、支付时间早于 now-goods_unverified_hours 的订单生成预警事件。
（租赁 order_type=2 走租期/归还生命周期，不参与本预警，由租赁逾期预警 type=7 覆盖。）

#### Scenario: 零售订单支付超时未核销
- **WHEN** 存在 order_type=1、status=2、completed_at IS NULL、paid_at 早于阈值
- **THEN** 生成一条 type=3 的待处理预警，staff_id 为空，summary 记录支付时间

### Requirement: 服务订单超时未指派预警（type=4）
系统 SHALL 对服务订单（order_type∈3~6）、已支付（status=2）、仍待接单（biz_status=1）、未指派（assigned_staff_id IS NULL）、支付时间早于 now-service_unassigned_hours 的订单生成预警。

#### Scenario: 服务订单支付超时未指派
- **WHEN** 存在服务订单支付后超过阈值仍无人接单/未指派
- **THEN** 生成一条 type=4 的待处理预警，staff_id 为空

### Requirement: 陪诊订单超时未完成预警（type=5）
系统 SHALL 对陪诊订单（order_type=4）、已指派（assigned_staff_id 非空）、已签到开始服务（actual_started_at 非空）、未签退（actual_ended_at IS NULL）、服务中超过 escort_unfinished_minutes 分钟未完成的订单生成预警，并记录关联服务人员。

#### Scenario: 陪诊服务超时未完成
- **WHEN** 陪诊订单已开始服务但超过阈值仍未签退
- **THEN** 生成一条 type=5 的待处理预警，staff_id=指派人员

### Requirement: 服务订单指派超时未签到预警（type=6）
系统 SHALL 对服务订单（order_type∈3~6）、已指派（assigned_staff_id 非空）、未签到（actual_started_at IS NULL）、指派时间早于 now-service_unstarted_minutes 的订单生成预警，并记录关联服务人员。
（orders.assigned_at 在管理员派单 DispatchOrder 与服务人员接单 AcceptOrder 时写入。）

#### Scenario: 已指派但超时未签到
- **WHEN** 服务订单已指派但超过阈值未签到开始服务
- **THEN** 生成一条 type=6 的待处理预警，staff_id=指派人员

### Requirement: 租赁订单逾期未归还预警（type=7）
系统 SHALL 对租赁订单（order_type=2）、未归还（deposit_status∈0,1）、到期时间（rental_end_at）早于 now-rental_overdue_hours 的订单生成预警。

#### Scenario: 租赁逾期未归还
- **WHEN** 租赁订单已逾期超过阈值且未归还
- **THEN** 生成一条 type=7 的待处理预警，staff_id 为空

### Requirement: 退款卡在处理中预警（type=8）
系统 SHALL 对退款单 status=0（处理中）创建时间早于 now-refund_stuck_hours 且订单未完成回调的退款生成预警，关联其订单。

#### Scenario: 退款长时间未回调
- **WHEN** 退款单 status=0 超过阈值小时未回调成功
- **THEN** 生成一条 type=8 的待处理预警，关联订单

### Requirement: 预警事件幂等
系统 SHALL 对同一 order_id + alert_type 仅保留一条未闭环（status≠3）预警，重复扫描不重复创建。

### Requirement: 预警中心展示扩展
系统 SHALL 在预警中心列表/详情展示新预警类型文案（type 3~8）、摘要（summary）；staff_id 为空时关联服务人员列展示"-"，不再回退显示 "ID: 0"。

## MODIFIED Requirements
### Requirement: 预警中心（服务过程安全）
扩展 service_alert_events 以承载订单级预警：staff_id 可空、新增 summary 列；预警类型文案覆盖 1~8；预警类型筛选支持新类型。

## REMOVED Requirements
（无）
