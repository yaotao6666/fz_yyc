# Tasks

- [x] Task 1: 新增两个独立路由与视图骨架
  - 新建 `GoodsOrderListView.vue`（实物，category=1）与 `ServiceOrderListView.vue`（服务，category=2）。

  - 在 `router/index.ts` 注册 `orders/goods`、`orders/service`；`orders` 保留兜底/重定向。

  - 每个视图：独立 `filters`、`columns`、操作区、分页、`loadOrders()` 固定传 `category`。

- [x] Task 2: 实现实物订单列表
  - 搜索：订单号、状态、订单类型(零售1/租赁2)、日期。

  - 列：订单号/用户/服务对象/商品/金额/状态/工单状态/服务人员/租赁到期/下单时间。

  - 操作：查看详情、派单(order:dispatch，配送)、续租(仅 order_type=2，order:renew)。

- [x] Task 3: 实现服务订单列表
  - 搜索：订单号、状态、工单状态、日期。

  - 列：订单号/用户/服务对象/金额/状态/工单状态/服务人员/下单时间。

  - 操作：查看详情、派单(order:dispatch)；**无**租赁到期列与续租。

- [x] Task 4: RBAC 菜单/权限落库
  - 新增迁移脚本，在现有订单菜单下加「实物订单」(order:goods)、「服务订单」(order:service) 两个页面节点，并授权 role_id=1；执行脚本验证落库。

  - 侧边栏菜单展示两个新入口。

- [x] Task 5: 构建验证
  - `npm run build` 通过（枚举移除 5/6 后与新增列/搜索无 TS 错误）。

# Task Dependencies

- Task 2 / Task 3 依赖 Task 1（两视图骨架）。

- Task 4 可与 Task 2/3 并行（不依赖页面实现，仅菜单/权限）。

- Task 5 依赖 Task 2/3/4。