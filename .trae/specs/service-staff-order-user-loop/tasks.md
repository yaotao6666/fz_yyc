# Tasks

> 说明：本 spec 不新建健康评估/照护/随访/体征模块，仅做「接单 + 用户信息」优化闭环。后端、数据库通知表、服务人员小程序、PRD 文档需同步（见 .trae/rules/快速迭代.md）。

## 阶段一：后端「人」的信息增强（服务对象/健康警示/下单人）

- [ ] Task 1: 模型与组装服务
  - [ ] `models.go` 扩展 Order 瞬态展示结构：`Record *HealthRecord`（经 `record_id`）与下单人 `User`；定义服务对象/健康警示/下单人 DTO
  - [ ] 复用或扩展 `services/orderquery/query.go` 的 `FillOrderRecordInfo`，补充服务对象（姓名/性别/出生日期/评估等级）、健康警示（过敏史/慢病标签/紧急联系人）、下单人（昵称/头像/累计订单/累计消费）
- [ ] Task 2: 工单接口回填并支持排序（`workorder.go`）
  - [ ] `OrderDetail`：返回 record（服务对象）、健康警示、下单人画像
  - [ ] `PendingOrders`：每条订单回填服务对象/下单人/预约时间/区域，支持 `?sort=scheduled_at`（升序）
  - [ ] `AcceptedOrders`、`GetTodoList`：回填服务对象行数据
- [ ] Task 3: 新增「放弃已接工单」
  - [ ] `GiveUpOrder` handler：仅 `assigned_staff_id=本人` 且 `biz_status=2` 时原子回退 `biz_status→1`、`assigned_staff_id→NULL`、`actual_started_at→NULL`
  - [ ] `main.go` staffAuthedGroup 注册 `POST /service-staff/orders/:id/give-up`

## 阶段二：接单后通知客户

- [ ] Task 4: 通知基础设施
  - [ ] 迁移脚本新增 `order_notify_logs` 表（order_id、user_id、channel、success、message、created_at）+ 模型
  - [ ] 新增 `services/notify/notify.go`：best-effort 微信订阅消息 + 短信双通道（模板/密钥未配置则跳过），幂等写日志
  - [ ] config.go 增加可选订阅消息模板 ID 配置（无则跳过）
- [ ] Task 5: 接入接单通知
  - [ ] `AcceptOrder` 成功后触发通知下单用户（服务人员姓名/电话/预约时间），失败不阻断接单主流程

## 阶段三：服务人员小程序

- [ ] Task 6: API 与类型
  - [ ] `types/index.ts`：工单/服务对象/健康警示/下单人类型
  - [ ] `api/index.ts`：新增 `giveUpOrder(id)`
- [ ] Task 7: `workorder/index.vue` 待接单卡片增强
  - [ ] 展示服务对象、预约时间、区域、实付金额、下单人；后端默认按预约时间排序
- [ ] Task 8: `workorder/detail.vue` 用户信息与操作增强
  - [ ] 新增「服务对象」卡（姓名/性别/年龄/评估等级，与下单人区分）
  - [ ] 新增「健康警示」卡（过敏史/慢病标签/紧急联系人，折叠展示）
  - [ ] 新增「下单人」卡（昵称/头像/累计订单/累计消费）
  - [ ] 一键拨打（联系人/服务对象/紧急联系人）与一键导航（`uni.openLocation`，无坐标时回退复制地址）
  - [ ] 新增「放弃工单」按钮（仅 `biz_status=2` 待出发）
  - [ ] 移除「录入照护记录」/`goRecordVisit`/`canRecordVisit` 空入口与签退后询问弹窗
- [ ] Task 9: `todo/index.vue` 待办行展示服务对象

## 阶段四：验证 + PRD 同步

- [ ] Task 10: 构建与回归
  - [ ] `go build ./...` 通过
  - [ ] `staff-miniprogram` mp-weixin 构建通过
  - [ ] 回归：待接单 → 接单 → 待办 → 签到 → 签退 主链路未被破坏；give-up 仅待出发可放弃
- [ ] Task 11: PRD 文档同步（PRD.md / docs/prd 功能与接口）与 mermaid 工单链路更新

# Task Dependencies

- Task 1 → Task 2（接口回填依赖组装服务）→ Task 3 可并行
- Task 4 → Task 5（接单通知依赖通知基础设施）；Task 5 依赖 Task 2 的接单流程（互不阻塞，可并行）
- Task 6 → Task 7/8/9（前端依赖 API 与新类型）
- Task 10 依赖所有后端与前端 Task；Task 11 依赖全部功能稳定后同步文档