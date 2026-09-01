# 服务人员小程序「接单 × 用户信息」优化闭环 Spec

## Why

服务人员小程序已具备「待接单 → 接单 → 待办 → 签到 → 签退」的执行骨架，但接单与履约过程对"人"的信息利用不足且存在误导：

1. **用户信息错位**：工单只展示**下单人**的 `contact_name/contact_phone`，看不到真正的**被服务人**。服务订单通过 `orders.record_id` 关联健康档案（多为下单人的父母/长辈），字段 `RecordName/RecordGender/RecordBirthDate` 已存在于模型但后端 `OrderDetail` 未回填、前端未展示——服务人员上门前无法确认服务对象、无法做准备。
2. **安全信息缺失**：过敏史、慢病标签、评估等级、紧急联系人等对上门服务安全至关重要的信息没有前置展示，也无一键联系/导航。
3. **接单池单薄**：待接单卡片无预约时间/服务对象/区域距离，无排序筛选；接单后客户不知道谁上门；已接工单没有放弃出口，遇突发情况会僵死。
4. **残留空入口**：健康评估/照护/随访/体征模块已按业务决策移除（保持移除），但工单详情仍残留跳转已删除页面 `/pages/health/care-visit-form` 的「录入照护记录」入口，点了会白屏。

本 spec 在**不新增健康模块**的前提下，把「接单 → 服务 → 签退」这一执行闭环从服务人员视角做扎实：正确区分下单人与服务对象、前置健康警示、补充下单人画像，提供一键联系/导航；同时增强接单池、接单后通知客户、支持放弃已接工单、移除空入口。

## What Changes

- **后端工单接口全面补充"人"的信息**：`OrderDetail/PendingOrders/AcceptedOrders/GetTodoList` 返回服务对象（姓名/性别/出生日期/评估等级）、健康警示（过敏史/慢病标签/紧急联系人，折叠或脱敏）、下单人画像（昵称/头像/累计订单/累计消费），并保留 `scheduled_at`/`delivery_district` 供排序与展示。
- **接单池增强**：待接单支持按预约时间（`scheduled_at`）升序排序的查询参数，卡片信息更丰富；区域过滤逻辑保留。
- **新增「放弃已接工单」**：`POST /service-staff/orders/:id/give-up`，仅当 `assigned_staff_id=本人` 且 `biz_status=2`（待出发）时，将工单退回到待接单池（`biz_status→1, assigned_staff_id→NULL, actual_started_at→NULL`）。
- **新增接单通知**：接单成功后 best-effort 通知下单用户（微信订阅消息 + 短信，双通道尽力而为，任一未配置则跳过），写入 `order_notify_logs` 表幂等去重。
- **服务人员小程序**：工单列表/待办/详情展示服务对象与健康警示与下单人画像，一键拨打、一键导航，新增「放弃工单」入口，移除「录入照护记录」空入口。
- **BREAKING**：无。均为新增响应字段/接口/可选配置，向后兼容；不删除或新建健康业务表（仅新增 `order_notify_logs` 一项通知留痕表）。

## Impact

- Affected specs（能力）：
  - 服务人员工单执行能力（待办/接单/签退）
  - 待接单池（排序/筛选/信息丰富）
  - 下单用户接单通知（新）
  - 服务人员小程序 UI 与 API 封装
- Affected code：
  - 后端：[workorder.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/service_staff/workorder.go)、[models.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/models/models.go)（Order 瞬态字段/通知模型）、新增 services/notify、[orderquery/query.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/services/orderquery/query.go)（复用/扩展 FillOrderRecordInfo）、[main.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/cmd/server/main.go)（staffAuthedGroup 注册 give-up）
  - 数据库：新增 `order_notify_logs` 表
  - staff-miniprogram：[api/index.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/staff-miniprogram/src/api/index.ts)、[types/index.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/staff-miniprogram/src/types/index.ts)、[workorder/index.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/staff-miniprogram/src/pages/workorder/index.vue)、[workorder/detail.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/staff-miniprogram/src/pages/workorder/detail.vue)、[todo/index.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/staff-miniprogram/src/pages/todo/index.vue)
  - 文档：PRD.md、docs/prd 相关文档（按 .trae/rules/快速迭代.md）

---

## ADDED Requirements

### Requirement: 服务对象与健康警示信息前置（用户信息核心）
系统 SHALL 在服务人员端工单详情返回并展示本单的**被服务人**信息与健康警示信息，与下单人区分。

- `OrderDetail` 返回 `record` 子对象：`{ id, real_name, gender, birth_date, assessment_level }`（来自 `orders.record_id` 关联的 health_records）。
- 健康警示字段：`allergy_history`、`chronic_tags`、`emergency_contact`、`emergency_phone`；敏感信息折叠展示，不因展示而外泄非本单数据。
- 下单人画像：昵称、头像、累计订单数、累计消费（来自 users）。
- 数据权限：服务对象信息仅取本单 `record_id` 指向的档案（该档案归属下单用户，服务人员因接本单获得查看权），不返回其他档案。

#### Scenario: 服务人员查看工单详情
- **WHEN** 服务人员进入待接或已接工单详情
- **THEN** 页面展示「服务对象」卡（姓名/性别/年龄/评估等级，"服务对象"与"下单人"不同时显著标识）、「健康警示」卡（过敏史/慢病标签/紧急联系人）、「下单人」卡，并可一键拨打各电话、一键导航到服务地址

### Requirement: 接单池信息与预约时间排序
系统 SHALL 在待接单列表返回增强字段并支持预约时间排序。
- `PendingOrders` 返回每条订单的 `record` 服务对象、`scheduled_at`、`delivery_district`、实付金额、下单人简称。
- 支持 `?sort=scheduled_at`（按预约时间升序）与基于 `staff.service_region` 的区域过滤（现有逻辑保留）。

#### Scenario: 浏览待接单
- **WHEN** 服务人员打开「待接单」Tab
- **THEN** 卡片展示服务对象、预约时间、区域、实付金额、下单人，并按预约时间升序排列（启用排序参数时）

### Requirement: 放弃已接工单
系统 SHALL 提供服务人员在待出发阶段放弃已接工单的能力。
- 新增 `POST /service-staff/orders/:id/give-up`。
- 仅当 `assigned_staff_id = 本人` 且 `biz_status = 2`（待出发）时放行，原子更新 `biz_status→1`、`assigned_staff_id→NULL`、`actual_started_at→NULL`，工单退回待接单池。

#### Scenario: 放弃待出发工单
- **WHEN** 服务人员在待接单/已接列表对 `biz_status=2` 工单点击「放弃工单」并确认
- **THEN** 工单回到待接单池，服务人员待办列表不再包含该工单

#### Scenario: 服务中不可放弃
- **WHEN** 工单已签到（`biz_status=3` 服务中）尝试放弃
- **THEN** 返回错误，服务中工单不可放弃

### Requirement: 接单后通知客户
系统 SHALL 在接单成功后尽力通知下单用户，且幂等不重复。
- `AcceptOrder` 成功触发通知：微信订阅消息（配置模板 ID 时）与短信（配置时），任一未配置则跳过该通道。
- 写 `order_notify_logs` 表（order_id、user_id、channel、success、message、created_at），接单成功仅发送一次。

#### Scenario: 接单后通知
- **WHEN** 服务人员接单成功
- **THEN** 尽力通知下单用户（内容含服务人员姓名/电话/预约时间）；重复抢单同一订单不重复发送

### Requirement: 移除残留空入口（修正）
系统 SHALL 移除工单详情中跳转已删除页面的「录入照护记录」入口，使健康模块保持移除状态。
- 删除 `workorder/detail.vue` 中的 `goRecordVisit`、`canRecordVisit`、底部「录入照护记录」按钮及签退后弹窗询问逻辑。

#### Scenario: 工单详情无空入口
- **WHEN** 服务人员查看已服务工单详情
- **THEN** 不再显示跳转到不存在页面 `/pages/health/care-visit-form` 的「录入照护记录」按钮或弹窗

## MODIFIED Requirements

### Requirement: 服务人员工单接口信息增强
现有 `PendingOrders`、`AcceptedOrders`、`OrderDetail`、`GetTodoList` 保持状态流转不变，扩展为返回服务对象、健康警示、下单人画像与预约时间等字段（见各 ADDED Requirement），并展示到服务人员小程序列表/详情。

## REMOVED Requirements

无。