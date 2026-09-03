# Tasks

> 说明：本 spec 不新建健康评估/照护/随访/体征模块，仅做「接单 + 用户信息」优化闭环。后端、数据库通知表、服务人员小程序、PRD 文档需同步（见 .trae/rules/快速迭代.md）。

## 阶段一：后端「人」的信息增强（服务对象/健康警示/下单人）

- [x] Task 1: 模型与组装服务
  - [x] `models.go` 扩展 Order 瞬态展示结构：`Record *HealthRecord`（经 `record_id`）与下单人 `User`；定义服务对象/健康警示/下单人 DTO
  - [x] 复用或扩展 `services/orderquery/query.go` 的 `FillOrderRecordInfo`，补充服务对象（姓名/性别/出生日期/评估等级）、健康警示（过敏史/慢病标签/紧急联系人）、下单人（昵称/头像/累计订单/累计消费）
- [x] Task 2: 工单接口回填并支持排序（`workorder.go`）
  - [x] `OrderDetail`：返回 record（服务对象）、健康警示、下单人画像
  - [x] `PendingOrders`：每条订单回填服务对象/下单人/预约时间/区域，支持 `?sort=scheduled_at`（升序）
  - [x] `AcceptedOrders`、`GetTodoList`：回填服务对象行数据
- [x] Task 3: 新增「放弃已接工单」
  - [x] `GiveUpOrder` handler：仅 `assigned_staff_id=本人` 且 `biz_status=2` 时原子回退 `biz_status→1`、`assigned_staff_id→NULL`、`actual_started_at→NULL`
  - [x] `main.go` staffAuthedGroup 注册 `POST /service-staff/orders/:id/give-up`

## 阶段二：接单后通知客户

- [x] Task 4: 通知基础设施
  - [x] 迁移脚本新增 `order_notify_logs` 表（order_id、user_id、channel、success、message、created_at）+ 模型
  - [x] 新增 `services/notify/notify.go`：best-effort 微信订阅消息双通道（模板未配置则跳过），幂等写日志
  - [x] 模板 ID 走 `WECHAT_STAFF_ACCEPT_TEMPLATE_ID` 环境变量（未配置跳过）
- [x] Task 5: 接入接单通知
  - [x] `AcceptOrder` 成功后触发通知下单用户（服务人员姓名/电话/预约时间），失败不阻断接单主流程

## 阶段三：服务人员小程序

- [x] Task 6: API 与类型
  - [x] `types/index.ts`：工单/服务对象/健康警示/下单人类型
  - [x] `api/index.ts`：新增 `giveUpOrder(id)`
- [x] Task 7: `workorder/index.vue` 待接单卡片增强
  - [x] 展示服务对象、预约时间、区域、实付金额、下单人；后端默认按预约时间排序
- [x] Task 8: `workorder/detail.vue` 用户信息与操作增强
  - [x] 新增「服务对象」卡（姓名/性别/年龄/评估等级，与下单人区分）
  - [x] 新增「健康警示」卡（过敏史/慢病标签/紧急联系人，折叠展示）
  - [x] 新增「下单人」卡（昵称/头像/累计订单/累计消费）
  - [x] 一键拨打（联系人/服务对象/紧急联系人）与一键导航（`uni.openLocation`，无坐标时回退复制地址）
  - [x] 新增「放弃工单」按钮（仅 `biz_status=2` 待出发）
  - [x] 移除「录入照护记录」/`goRecordVisit`/`canRecordVisit` 空入口与签退后询问弹窗
- [x] Task 9: `todo/index.vue` 待办行展示服务对象

## 阶段四：验证 + PRD 同步

- [x] Task 10: 构建与回归
  - [x] `go build ./...` 通过
  - [x] `staff-miniprogram` mp-weixin 构建通过
  - [x] 回归：give-up 仅待出发可放弃（核验 handler 原子条件）；主链路接口未改动破坏（构建通过 + handler 逻辑核验）
- [x] Task 11: PRD 文档同步（PRD.md / docs/prd 功能与接口）与 mermaid 工单链路更新
  - [x] PRD-接口文档.md 2.10 服务人员接口：补 sort/customer/give-up/todo/接单通知
  - [x] PRD-功能说明.md：服务人员端职责/页面/接单流程/链路图同步
  - [x] PRD.md 3.2 服务人员接单接口：新增放弃工单/待办/接单通知/customer，含签名勘误
  - [x] docs/uml/02-服务人员工单链路.md mermaid：补接单通知、放弃工单分支、order_notify_logs 表

# Task Dependencies

- Task 1 → Task 2（接口回填依赖组装服务）→ Task 3 可并行
- Task 4 → Task 5（接单通知依赖通知基础设施）；Task 5 依赖 Task 2 的接单流程（互不阻塞，可并行）
- Task 6 → Task 7/8/9（前端依赖 API 与新类型）
- Task 10 依赖所有后端与前端 Task；Task 11 依赖全部功能稳定后同步文档