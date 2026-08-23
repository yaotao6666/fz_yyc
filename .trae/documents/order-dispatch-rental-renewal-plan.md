# PC 后台订单派单 + 租赁到期提醒 + 续租 — 实施计划

## Summary
为 PC 后台（web-admin）增强订单运营能力：
1. **订单派单**：对已支付订单，选择**已审核通过的服务人员**进行指派；指派后订单进入该服务人员的工单列表，服务人员可在小程序查看并进行后续接单/签退交互。
2. **租赁到期提醒列表**：orders 新增 `rental_end_at` 到期时间字段（支付时按租赁时长自动计算），PC 后台新增"待归还/将到期/已逾期"租赁到期列表。
3. **续租**：对租赁订单点击"续租" → 生成一笔关联新订单（记录 `parent_order_id`），复用现有支付链路，用户侧直接对续租订单完成支付。
4. 三块功能均新增独立 RBAC 权限码，并同步 PRD 文档。

## Current State Analysis（已探索）
- **订单模型**（[models.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/models/models.go#L280-L325)）：`orders` 已有 `order_type`(2=租赁)、`biz_status`(1待接单/2已接单/3服务中/5已完成)、`assigned_staff_id`、`deposit_status`、`rental_returned_at`、`paid_at`、`completed_at` 等，**无 `rental_end_at`、无 `parent_order_id`**。
- **下单**（[handler.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/user/handler.go#L642)）：`CreateOrder` 已支持 `req.AssignedStaffID`；服务类订单(order_type 3/4/5)支付后自动 `biz_status=1` 进待接单池；租赁(order_type=2)不设 biz_status。
- **服务人员工单**（[workorder.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/service_staff/workorder.go)）：
  - `PendingOrders` = 已支付 + biz_status=1 + assigned_staff_id IS NULL（待接单池，全员可抢单）
  - `AcceptedOrders` / `TodoList` = `assigned_staff_id = 当前服务人员`（已指派/已接单即出现在此）
  - `AcceptOrder` = 仅 `biz_status=1 AND assigned_staff_id IS NULL` 才能接。**指派后应设 `biz_status=2`（已接单/待出发）**，使其直接进入服务人员"已接订单/待办"，无需再抢单。
- **服务人员列表**（[service_staff.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/merchant/service_staff.go)）：`GetServiceStaffList` 已支持 status 筛选；需选已审核通过（status=1 且 audit_status=0）的服务人员。
- **web-admin**：订单列表 [OrderListView.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/views/order/OrderListView.vue)、详情 [OrderDetailView.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/views/order/OrderDetailView.vue)；`sp.ts` 有 `getServiceStaffList`、`getOrders`；`SpOrder` 类型缺新字段。
- **权限码规范**：服务人员 `staff:view/update`；需为派单/到期/续租新增独立权限码并按 RBAC 三方一致(前端按钮=接口=sys_menus 节点)。

## Proposed Changes

### 1. 数据模型 / 迁移（新增 `004_order_dispatch_rental.sql`）
**orders 表新增列（幂等）**：
- `rental_end_at` DATETIME NULL COMMENT '租赁到期时间(支付时=paid_at+时长)'
- `parent_order_id` BIGINT UNSIGNED NULL COMMENT '续租关联原订单ID(续租订单记录,0/空=普通订单)'
- `renew_flag` TINYINT 默认0 COMMENT '是否续租单:1=续租 0=非'

**RBAC 菜单/权限**（新增到 `sys_menus`，绑定角色1）：
- 订单管理(id=2)下新增按钮：
  - 派单 `order:dispatch`（menu_type=2）
  - 续租 `order:renew`（menu_type=2）
- 新增顶级/子菜单「租赁到期提醒」(path `/orders/rental-due`)：`orderrental:view`（menu_type=1）
  - 其按钮：续租 `orderrental:renew`、归还 `orderrental:return`（可复用现有归还）

### 2. 后端（Go）
**A. `models.go`**：`Order` 结构体新增 `RentalEndAt`、`ParentOrderID`、`RenewFlag` 字段。

**B. 到期计算与下单写入**（[handler.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/user/handler.go#L642)）：
- `CreateOrder` 中：当商品含租赁项时，计算 `rental_end_at = 支付时点 / 当前时间 + 最长租赁时长`。由于支持多租赁商品，取 Max(rental_duration) 按对应 rental_unit 折算天数。**简化约定**：rental_unit 1=天/2=周/3=月，统一折算为天数后加到创建时间。支付回调成功时更新 `rental_end_at = paid_at + duration`。为正确，**在支付回调(pay)成功处基于 paid_at 重新计算**更可靠。

**C. 订单支付回调用 `rental_end_at`**：查看 `callbacks/wechatpay.go` 支付成功逻辑，更新租赁订单的 `rental_end_at`。计划在回调内新增：若 order_type=2 且 items 含租赁，按最长时长折算并写 `rental_end_at = paid_at + X`。

**D. 商户端新增 handler**（[order_dispatch.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/merchant/order_dispatch.go)，新建）：
- `GetDispatchableStaffList`：返回 `status=1 AND audit_status=0` 的已审核服务人员（id/name/phone）。
- `DispatchOrder`（POST /merchant/orders/:order_id/dispatch）：入参 staff_id。校验订单已支付(Status=2)，指派后设置 `assigned_staff_id`、`biz_status=2`（待出发/已接单），使服务人员在"已接/待办"可见。可留存指派日志（写入 service_staff_audit_records 或直接更新）。
- `ListRentalDueOrders`（GET /merchant/orders/rental-due）：按 `order_type=2 AND status IN(2,3) AND deposit_status IN(0,1)`（未归还）过滤，支持 `due_range`(即将到期7天/已逾期/全部) 与关键词、分页，返回订单+用户+items，附 `rental_end_at`、`is_overdue`、`days_left`。
- `RenewOrder`（POST /merchant/orders/:order_id/renew）：基于原租赁订单生成关联新订单（`parent_order_id=原ID`、`renew_flag=1`、复制商品 items/费用，`status=1` 待支付），返回新订单，前端随后走现有支付。续租金额按续租时长重算；续租时长默认取原最长租赁时长，可选入参 `duration`。

**E. 路由注册**（[main.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/cmd/server/main.go)）：
```
GET  /merchant/orders/dispatchable-staff   → order:dispatch（返回可选服务人员）
POST /merchant/orders/:order_id/dispatch   → order:dispatch
GET  /merchant/orders/rental-due           → orderrental:view
POST /merchant/orders/:order_id/renew      → order:renew
```

### 3. web-admin 前端
- **类型**[sp.ts]：`SpOrder` 增 `rental_end_at`、`parent_order_id`、`renew_flag`；`SpOrderItem` 已有租赁字段。
- **API**[sp.ts]：`getDispatchableStaffList`、`dispatchOrder`、`getRentalDueOrders`、`renewOrder`。
- **订单详情页** [OrderDetailView.vue]：
  - 已支付订单新增「派单」按钮（`v-permission="order:dispatch"`）→ 弹窗选择已审核服务人员 → 调用 dispatch。
  - 工单信息卡展示指派服务人员姓名。
  - 租赁订单新增「续租」按钮（`v-permission="order:renew"`）→ 确认后调用 renew，提示新订单号并跳转/展示支付。
- **订单列表页** [OrderListView.vue]：操作列对租赁已支付订单加「续租」「派单」入口；展示 `rental_end_at` 列（租赁时）。
- **新增页面** `views/order/RentalDueListView.vue`：租赁到期提醒列表（待归还/将到期/已逾期筛选，展示剩余天数，行内续租/归还）。路由 `/orders/rental-due`，meta.permission `orderrental:view`。

### 4. 服务人员小程序（staff-miniprogram）
- 现有 `AcceptedOrders`(已接) / `TodoList`(待办，biz_status∈{2,3}) 已覆盖"指派后查看"。**无需改逻辑**，指派后订单自动出现在服务人员"工单→已接/待办"。可选：派单时推送（现有 ws 机制）。
- 仅需确认指派(biz_status=2)后服务人员可见并交互（已在计划内通过 biz_status=2 保证）。文档/PRD 注明。

### 5. 文档（PRD 同步）
- [PRD-接口文档.md](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/docs/prd/PRD-接口文档.md)：新增派单/到期/续租接口表 + 权限码。
- [PRD-功能说明.md](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/docs/prd/PRD-功能说明.md)：
  - 订单章节补"派单"流程（指派→biz_status=2→服务人员工单可见）。
  - 租赁章节补"到期提醒 + 续租"流程。
  - RBAC 权限维护规范补 `order:dispatch/order:renew/orderrental:view`.
- [PRD.md](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/PRD.md)：同步摘要。

## Assumptions & Decisions
1. **到期时间**：新增 `rental_end_at` 字段（用户确认），支付成功时基于 `paid_at + 租赁时长` 计算；时长按 item 最长租赁时长、按 rental_unit(天/周/月) 折算天数。
2. **续租**：生成关联新订单（`parent_order_id` 指向原单）+ `renew_flag=1`（用户确认），复用现有支付；续租金额按续租时长重算。
3. **到期提醒**：独立"租赁到期提醒"列表（用户确认），按未归还(未归还=deposit_status 非 2/3 或 rental_returned_at 为空)筛选，标注将到期/已逾期/剩余天数。
4. **派单服务人员范围**：仅已审核通过（`status=1 AND audit_status=0`）（用户确认）。
5. **指派后状态**：`assigned_staff_id` 写入 + `biz_status=2`（已接单/待出发），使服务人员在工单"已接/待办"无需抢单即可看到并进入签到/签退流程。
6. 迁移幂等（`CREATE TABLE IF NOT EXISTS` 不适合加列，用 `information_schema` 判断 ADD COLUMN，参照项目既有 `003_*.sql` 写法）。
7. 涉及新表字段与接口，遵循"改接口/数据表必须同步 PRD"。服务人员相关新表统一 `service_` 前缀；此处无新增服务人员表，续租/到期复用 orders。

## Verification
1. 后端 `go build ./...` 通过；迁移执行成功（orders 新列、sys_menus 新权限、绑定角色1）。
2. 支付租赁订单 → `rental_end_at` 正确写入。
3. `POST /dispatch`：指派后 DB `assigned_staff_id`=选择值、`biz_status`=2；用该服务人员 token 调 `GET /service-staff/orders/accepted` 可见该订单；可继续签到/签退。
4. `GET /orders/rental-due`：仅返回未归还租赁订单，含到期/剩余天数；`POST /orders/:id/renew` 生成 parent_order_id=原单的新待支付订单。
5. web-admin `npm run build` 通过；订单详情/列表出现派单、续租、到期入口，权限码生效（无权限不显示）。
6. staff-miniprogram：指派订单在工单"已接/待办"可见（构建通过）。
7. 容器重建，新路由注册、`/health` 200。

## File Path Reference
- 迁移：`server/migrations/004_order_dispatch_rental.sql`（新增）
- 模型：`server/internal/models/models.go`
- 商户 handler：`server/internal/handlers/merchant/order_dispatch.go`（新增）
- 下单/支付：`server/internal/handlers/user/handler.go`、`server/internal/handlers/callbacks/wechatpay.go`
- 路由：`server/cmd/server/main.go`
- web-admin：`api/sp.ts`、`types/sp.ts`、`router/index.ts`、`views/order/OrderListView.vue`、`OrderDetailView.vue`、`views/order/RentalDueListView.vue`（新增）
- 文档：`PRD.md`、`docs/prd/PRD-接口文档.md`、`docs/prd/PRD-功能说明.md`