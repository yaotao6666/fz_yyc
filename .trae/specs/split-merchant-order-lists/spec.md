# 拆分 PC 端实物订单与服务订单为独立列表 Spec

## Why

PC 端「订单管理」目前把实物订单和服务订单混在一个页面，共用一套列与搜索条件。两类订单虽然都涉及派单/工单状态，但字段与操作仍有差异（服务订单没有租赁到期，实物订单多为配送取送），混合列表导致列冗余、搜索语义混杂、权限无法细分。

## What Changes

- 将「订单管理」拆成**两个独立菜单与路由页面**：

  - **实物订单** `/orders/goods`

  - **服务订单** `/orders/service`

- 每个独立页面持有**各自独立的搜索区、表格列、操作区**，请求各自固定传 `category=1` / `category=2`。

- **实物订单列表**（category=1，order\_type ∈ 1,2）：

  - 搜索：订单号、订单状态、订单类型（零售/租赁）、下单日期。

  - 列：订单号、用户、服务对象、商品、订单金额、订单状态、**工单状态**、**服务人员**、**租赁到期**、下单时间。

  - 操作：查看详情、**派单（配送）**、续租（仅租赁）。

  - 说明：实物订单的工单状态/服务对象/服务人员用于**服务人员配送/取送**场景，派单走现有 `order:dispatch`。

- **服务订单列表**（category=2，order\_type ∈ 3,4）：

  - 搜索：订单号、订单状态、工单状态、下单日期。

  - 列：订单号、用户、服务对象、订单金额、订单状态、工单状态、服务人员、下单时间。

  - 操作：查看详情、派单。

  - **不含**「租赁到期」列与「续租」操作。

- **权限拆分**：新增两个页面级权限码 `order:goods`、`order:service`，分别挂「实物订单」「服务订单」两个菜单节点；原 `orders:view` 保留作为兜底/详情。RBAC 走规范四步：接口→权限码→迁移脚本菜单节点→PRD。

- 后端复用现有 `GET /api/v1/merchant/orders` 的 `category=1/2` 与 `order:dispatch`、`order:renew`，字段已完整（`record_name / biz_status / assigned_staff / rental_end_at / order_type`），**无需 schema 或接口破坏性变更**。

## Impact

- Affected specs / modules：PC 后台（web-admin）订单管理、RBAC 菜单/权限。

- Affected code：

  - `web-admin/src/router/index.ts`：新增 `/orders/goods`、`/orders/service` 两个路由。

  - 新建 `web-admin/src/views/order/GoodsOrderListView.vue`、`ServiceOrderListView.vue`；`OrderListView.vue` 转为总览或弃用。

  - 迁移脚本：`server/migrations/` 新增菜单节点 `order:goods` / `order:service` 并授权 role\_id=1。

  - 后端 `GetOrders` / `DispatchOrder`：无功能性改动（`category` 过滤、派单已具备）。

## ADDED Requirements

### Requirement: 实物订单独立菜单+路由

系统 SHALL 在订单管理下提供独立「实物订单」菜单与路由 `/orders/goods`，权限码 `order:goods`。

#### Scenario: 进入实物订单列表

- **WHEN** 拥有 `order:goods` 权限的用户打开实物订单菜单

- **THEN** 请求固定 `category=1`，展示含「工单状态/服务人员/服务对象/租赁到期」列的实物订单

#### Scenario: 实物订单配送派单

- **WHEN** 对已支付实物订单执行派单

- **THEN** 通过 `order:dispatch` 指派服务人员配送，订单进入工单流转并可查看工单状态/服务人员

#### Scenario: 实物租赁订单续租

- **WHEN** 对已支付租赁实物订单操作续租

- **THEN** 展示「续租」操作并调用 `order:renew` 生成续租单

### Requirement: 服务订单独立菜单+路由

系统 SHALL 在订单管理下提供独立「服务订单」菜单与路由 `/orders/service`，权限码 `order:service`。

#### Scenario: 进入服务订单列表

- **WHEN** 拥有 `order:service` 权限的用户打开服务订单菜单

- **THEN** 请求固定 `category=2`，展示「服务对象/工单状态/服务人员」列，不含「租赁到期」列与「续租」操作，操作含「查看详情」「派单」

### Requirement: 两类列表派单能力

系统 SHALL 使实物订单与服务订单均可派单（`order:dispatch`），实物订单的工单状态/服务对象/服务人员用于配送场景，服务订单用于上门服务场景。

### Requirement: RBAC 独立权限（`order:goods` / `order:service`）

系统 SHALL 提供两个独立页面权限码与菜单节点：`order:goods`（实物订单）、`order:service`（服务订单），经前端按钮/接口/MCU 菜单三端一致落库。

## MODIFIED Requirements

### Requirement: 订单分类入口（原「订单管理 /orders」）

由单一 `OrderListView` 改为「实物订单」「服务订单」两个独立路由/菜单页。`category` 语义、分页、后端过滤逻辑不变；原 `/orders` 保留为总览或重定向，`orders:view` 保留兜底。

## REMOVED Requirements

### Requirement: 混合「全部/实物/服务」三 tab 单页列表

**Reason**：实物与服务订单在多列/操作/权限上差异大，混合页导致列冗余、搜索语义混杂、权限无法细分。
**Migration**：原三 tab 功能由「实物订单」「服务订单」两个独立页面覆盖（分别固定 `category=1/2`）；后端 `GET /merchant/orders` 在未传 `category` 时全量行为保留（供需要全部订单的其它入口使用）。
