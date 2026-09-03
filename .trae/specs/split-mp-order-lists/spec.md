# 拆分用户端/服务端小程序订单为实物/服务，下单区分计费 Spec

## Why

PC 端订单已拆分为「实物订单/服务订单」。用户端（C端）与服务端（服务人员）小程序仍然是混合订单，卡片固定展示配送费/收货地址，服务订单也会被要求选配送档位、扣除配送费，且与实物订单共用一套展示与下单逻辑，语义混杂。需要将两类小程序订单按实物/服务拆分独立呈现，并在下单时对服务订单正确区分（无配送费、不受商家配送参数约束）。

## What Changes

### 1. 用户端（C端）订单拆为两个独立页面

- 在 `pages.json` 新增两个固定分类的独立订单列表页：
  - **实物订单** `/pages/store/my-orders-goods`（固定 `category=1`）
  - **服务订单** `/pages/store/my-orders-service`（固定 `category=2`）
- 原 `my-orders.vue` 抽出共用列表逻辑/样式；两个页面各自固定 `category` 并做差异化卡片渲染。
- **实物卡片**：保留收货地址摘要、配送费明细行。
- **服务卡片**：展示「服务对象」（健康档案）与「服务地址」（原收货地址，服务订单下改名为『服务地址』），不显示配送费行、不显示配送距离；状态筛选 tab 可微调（服务类不展示纯配送相关状态）。
- 各入口跳转改为按分类进入对应页：首页「我的订单」、订单详情返回、支付成功/取消重定向等。

### 2. 服务端（服务人员）工单按实物/服务分类

- `staff-miniprogram/src/pages/workorder/index.vue`：顶部新增「实物/服务」分类切换（服务人员同时承接实物配送与上门服务）。
- `staffWorkorderApi.getPendingOrders / getAcceptedOrders` 增加 `category` 参数。
- 后端 `PendingOrders / AcceptedOrders`（`server/internal/handlers/service_staff/workorder.go`）支持 `category` 过滤（1=实物,2=服务，经 order_type 范围判定），缺省不过滤。

### 3. 下单区分：服务订单不计配送费、不受配送参数约束

- **C端确认订单页** `miniprogram/src/pages/store/confirm.vue`：
  - 当订单含服务商品（`hasServiceItems`，product_type ∈ 3/4）时：隐藏「配送费」明细行、不加载/不要求配送距离档位、跳过配送距离与商家配送范围（`takeout_enabled`/`max_distance`）校验；仍要求选择**服务地址**（原收货地址，本页面服务订单下文案改名为「服务地址」，用于区域匹配）与**服务对象**健康档案。
  - 实物订单保持现有配送费/距离/取件判断不变。
- **后端 `CreateOrder`**（`server/internal/handlers/user/handler.go`）：
  - 服务订单（`utils.OrderCategory(orderType)==OrderCategoryService`）：强制 `deliveryFee=0`，忽略配送距离参数，不受 `takeout_enabled`/配送规则约束，`payAmount` 不计配送费。
  - 实物订单：保持现有配送费计算与配送参数校验。
  - 后端补充说明/注释明确服务订单口径。

### 4. 禁止实物与服务混单支付

- 同一订单内的商品分类必须一致（全部为实物 order_type∈1/2，或全部为服务 order_type∈3/4），**禁止**同一订单同时含实物商品与服务商品一起支付。
- **后端 `CreateOrder`**：遍历 `req.Items` 判定商品所属分类，若同时存在实物与服务商品则拒绝下单（返回参数错误）。
- **C端确认订单页 `confirm.vue`**：下单前置校验，若购物车/本次下单同时含实物与服务商品，则拦截并提示拆分后分别下单。

## Impact

- Affected specs / modules：用户端小程序、服务端小程序、后端订单创建/工单列表接口。
- Affected code：
  - `miniprogram/src/pages.json`、`miniprogram/src/pages/store/my-orders.vue`（抽共用 + 两个固定分类页面）、`confirm.vue`、相关入口。
  - `staff-miniprogram/src/pages/workorder/index.vue`、`staff-miniprogram/src/api/index.ts`。
  - `server/internal/handlers/user/handler.go`（CreateOrder）、`server/internal/handlers/service_staff/workorder.go`（PendingOrders/AcceptedOrders 增加 category）。

## ADDED Requirements

### Requirement: C端实物订单独立页面

系统 SHALL 提供用户端「实物订单」独立列表页 `/pages/store/my-orders-goods`，请求固定 `category=1`。

#### Scenario: 进入实物订单列表
- **WHEN** 用户从「我的订单」进入实物订单分类
- **THEN** 只展示实物订单，卡片含收货地址摘要与配送费行，并可按订单状态筛选

### Requirement: C端服务订单独立页面

系统 SHALL 提供用户端「服务订单」独立列表页 `/pages/store/my-orders-service`，请求固定 `category=2`。

#### Scenario: 进入服务订单列表
- **WHEN** 用户从「我的订单」进入服务订单分类
- **THEN** 只展示服务订单，卡片展示服务对象（健康档案）与服务地址（文案为「服务地址」），不显示配送费行

### Requirement: 服务订单地址文案为「服务地址」

系统 SHALL 在服务订单相关界面（下单页、订单列表/详情）将收货地址字段命名为「服务地址」，仅文案不同，字段复用 `delivery_address`。

### Requirement: 禁止实物与服务混单支付

系统 SHALL 阻止同一订单同时包含实物商品与服务商品并一起支付，约束下单时必须保证订单内商品分类一致。

### Requirement: 服务端工单按实物/服务分类

系统 SHALL 使服务人员小程序工单页支持「实物/服务」分类切换，并按分类过滤列表。

#### Scenario: 分类切换工单
- **WHEN** 服务人员在工单页切换「实物/服务」分类
- **THEN** 列表仅展示对应分类的订单/工单（实物配送 vs 上门服务），其余工单状态 tab 逻辑不变

### Requirement: 服务订单下单不计配送费

系统 SHALL 确保服务订单在下单时不产生配送费，且不受商家配送参数（`takeout_enabled`、配送距离/规则）约束，同时仍需选择收货地址与服务对象。

#### Scenario: 服务订单下单
- **WHEN** 用户提交含服务商品（product_type 3/4）的订单
- **THEN** 金额明细不出现配送费，无需选择配送距离/档位，不因商家配送开关或范围受限；仍校验收货地址与服务对象健康档案

## MODIFIED Requirements

### Requirement: C端混合订单单页 `my-orders.vue`

由单一混合列表（全部/实物/服务 tab）改为「实物订单」「服务订单」两个独立固定分类页面；`category` 过滤语义、后端 `GetMyOrders` 不变。

### Requirement: 服务端工单列表 `workorder/index.vue`

在原有工单状态 tab 之上增加「实物/服务」分类维度，后端列表接口补充 `category` 过滤。

### Requirement: 下单计费 `CreateOrder` / 确认订单页

实物订单保留现有配送费与配送参数逻辑；服务订单分支强制 `deliveryFee=0` 并跳过配送档位/开关约束。

## REMOVED Requirements

### Requirement: 服务订单在下单/列表中展示配送费并要求配送档位

**Reason**：服务订单本质为上门服务/区域内作业，不涉及实物配送费用，也不应计入配送费。
**Migration**：服务订单分类下隐藏配送费行与配送距离选择，金额不计配送费；收货地址保留但文案改为「服务地址」，仍用于区域匹配。