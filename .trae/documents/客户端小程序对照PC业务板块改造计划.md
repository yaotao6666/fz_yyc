# 客户端小程序对照 PC 业务板块改造计划

## Summary
以 PC 后台（web-admin，`/web-admin`）现有业务板块为基准，对用户端小程序（`/miniprogram`）做一次差距对照与改造，把 C 端升级为「单商户商城」形态：底部 tabBar（首页/分类/购物车/我的）、首页商城化（轮播图占位 + 多 icon 分类宫格 + 热销 + 分区）、订单运营对齐 PC（显示指派服务人员 / 租赁到期 / 续租）、健康服务闭环全面对齐、个人中心落地。轮播图 PC 端暂无配置入口，前端先以占位图承载，后续再补 PC 配置。

本次为**差距分析 + 实施计划**（phase 2 交付深度），不一次性全量落地，按优先级分阶段执行、每阶段可独立验证。

## Current State Analysis（已探索）

### C 端现状
- **结构**：无 tabBar，全部页面位于 `pages/store/*`（见 [pages.json](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages.json)）。首页入口 `pages/store/home` 为「店铺首页」：头部横幅 + 快捷入口 + 左侧分类侧边栏 + 右侧商品列表 + 底部购物车栏，已含按商品类型分区（零售/租赁/套餐/陪诊/资讯，见 [home.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages/store/home.vue) `PRODUCT_TYPE_SECTIONS`）。
- **购物车/下单**：`cart.vue`、`confirm.vue`、`my-orders.vue`、`order-detail.vue` 已完整；订单详情 [order-detail.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages/store/order-detail.vue) 仅展示基础状态/金额/押金，**未展示指派服务人员、租赁到期、续租入口**（仅押金状态）。
- **健康闭环**：[my-health.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages/store/my-health.vue) 已有 8 大入口（健康档案/自助评估/评估记录/适配建议/照护计划/康复随访/健康宣教/体征记录），整体覆盖 PC 健康模块。
- **资讯（科普资讯 product_type=5）**：正文存于 `service_content.content`，product.vue（[product.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages/store/product.vue#L93-L102)）已在商品详情内联展示，无专门资讯列表/图文阅读页。
- **类型定义**：[types/index.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/types/index.ts) `StoreHomeInfo` 无 `banners`；`Order` 无 `assigned_staff_name/biz_status/rental_end_at/parent_order_id/renew_flag`。

### 后端现状（已由探索确认）
- `GET /api/v1/store/home` → [handler.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/user/handler.go#L395-L441) `GetStoreHome`，返回 merchant/categories/hot_products/delivery_settings，**无 banners**（add 处）。
- `GET /api/v1/user/orders/:order_id` → [handler.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/user/handler.go#L1030-L1042) `GetOrderDetail`，经 `buildAccessibleOrder` 组装；[models.Order](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/models/models.go#L280-L307) 已有 `BizStatus/AssignedStaffID/RentalEndAt/ParentOrderID/RenewFlag`。
- 资讯商品数据归一路径见 [product.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/merchant/product.go#L632-L706)。
- PC 端派单/租赁到期/续租接口与 C 端无关（商户侧，见既有计划 [order-dispatch-rental-renewal-plan.md](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/.trae/documents/order-dispatch-rental-renewal-plan.md)）。

### 差距对照表（PC 业务板块 → C 端）
| PC 业务板块 | PC 页面 | C 端现状 | 差距 / 改造动作 |
|---|---|---|---|
| 商城交易（商品/分类） | ProductsView / CategoriesView | home 分类+商品+购物车 | 首页商城化、icon 分类宫格、轮播占位 |
| 订单运营（派单/到期/续租） | OrderList/Detail/RentalDue | my-orders/order-detail 无服务人员/到期/续租 | C 端详单补服务人员、租赁到期、续租 |
| 健康服务闭环 | health/* 8 子模块 | my-health 8 入口已覆盖 | 数据/交互对齐、与订单联动 |
| 数据分析 | MerchantStats | C 端无 | 可选（个人维度，低优先级） |
| 服务人员/员工 | staff/* | C 端无管理 | C 端订单/健康展示指派服务人员即可 |
| 系统 RBAC / 分账 | system/* , settlement/* | 与 C 端无关 | 不落地 C 端 |

### 用户明确新增的移动端基础能力
1. 轮播图（PC 暂无配置，前端先占位）
2. 查看资讯（科普资讯 product_type=5）
3. 首页改造为正常单商户商城、多个 icon 分类
4. 购物车（已具备，纳入 tabBar）
5. 预约订单（我的订单按业务/状态组织 + 服务预约信息展示）
6. 个人中心

## Proposed Changes（分阶段）

### 阶段 A：C 端架构升级（商城化 + 个人中心）— 最高优先级
**目标**：首页改成单商户商城；新增 tabBar（首页/分类/购物车/我的）；落地个人中心。
1. **pages.json**（[pages.json](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages.json)）：
   - 新增 `tabBar`：`pages/store/home`（首页）、`pages/store/category`（分类）、`pages/store/cart`（购物车）、`pages/store/me`（我的）。
   - 注册新页面 `pages/store/category`、`pages/store/me`。
   - tabbar 图标：复用 `static/tabbar` 现有（home/order/settings），补充 cart、me 图标（可用现有 settings/analytics 代替或新增 2 张，统一风格）。
2. **首页改造 [home.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages/store/home.vue)**：
   - 顶部新增**轮播图 `swiper`**：数据源读 `storeInfo.banners`（可选字段），为空时用 2-3 张占位图（占位图用商家封面 `merchant.cover_image` 或前端内置渐变占位，见假设 D1）。
   - 新增**icon 分类宫格**：把分类/商品类型映射为 icon 宫格（复用现有 `PRODUCT_TYPE_SECTIONS` 的 5 类配色图标，新增零售以外可点击跳商品），保持轻量、不引入复杂动画（遵守[快速迭代规则](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/.trae/rules/快速迭代.md)）。
   - 保留热销推荐、按类型分区、分类侧边栏（作为「分类」tab 的入口/兼容）。
3. **新增分类页 `category.vue`**：全部分类（多级）＋分类下商品，复用 `getStoreProducts`。
4. **新增个人中心 `me.vue`**：用户信息（头像/昵称）＋ 我的订单、我的健康、收货地址、联系商家、资讯入口。
5. **类型**[types/index.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/types/index.ts)：`StoreHomeInfo` 增可选 `banners?: {id;image;link_type?;link_value?}[]`；`categories` 增可选 `icon?`。

**说明**：`info-detail` 资讯与健康宣教合并入口放「我的健康 / 首页 icon」。

### 阶段 B：订单运营对齐（服务人员 / 租赁到期 / 续租）
**目标**：C 端订单详情展示 PC 已接入的字段，并提供续租。
1. **后端**：[handler.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/user/handler.go) `buildAccessibleOrder`/`GetOrderDetail` 增返回字段：
   - `assigned_staff_name`（由 `AssignedStaffID` JOIN `service_staffs` 取 name）
   - `biz_status`、`rental_end_at`、`parent_order_id`、`renew_flag`
   - 服务类(陪诊/康养)订单返回 `biz_status` 供 C 端显示待接单/已指派。
2. **后端新增 C 端续租接口** `POST /api/v1/user/orders/:order_id/renew`（用户侧，参照商户 `RenewOrder`）：校验本人、已支付租赁单未归还，生成 `parent_order_id=原单`、`renew_flag=1` 的新待支付订单，复用现有支付链路。路由注册见 [main.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/cmd/server/main.go)。
3. **前端 订单详情 [order-detail.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages/store/order-detail.vue)**：
   - 展示「指派服务人员」卡片（服务/租赁订单且已指派时）。
   - 租赁订单展示「到期时间 + 剩余天数/已逾期」；已支付未归还时显示「续租」按钮 → 调续租接口 → 提示新订单并跳支付/订单列表。
   - 更新 `Order` 类型字段。
4. **我的订单 [my-orders.vue](#)**：按业务/状态组织 tab（全部/待支付/进行中服务·租赁/已完成），服务类订单展示预约/服务状态。

### 阶段 C：健康服务闭环全面对齐
**目标**：系统比对 PC 健康 8 子模块与 C 端「我的健康」，补齐交互与数据联动。
1. 复核 [my-health.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/pages/store/my-health.vue) 8 入口各自列表页与后端 C 端接口一致（评估/适配/照护/随访/宣教/监测/档案）。
2. 健康项与订单联动：照护计划/随访展示关联订单态与服务人员（同阶段 B 返回字段）。
3. 资讯统一：`health-education`/`education-detail`（宣教文章）与 `product_type=5` 科普资讯入口互相可达，统一命名「科普资讯/健康宣教」。

### 阶段 D：PRD 同步（硬性约束）
遵守[快速迭代规则](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/.trae/rules/快速迭代.md)：改功能/接口/数据表必须同步 PRD。
- `PRD.md`：客户端小程序章节（tabBar/首页商城化/轮播占位/个人中心/订单对齐）、订单接口更新、`StoreHomeInfo.banners`、用户续租接口。
- `docs/prd/PRD-功能说明.md`、`PRD-接口文档.md`、`PRD-测试与附录.md`：同步页面结构、接口表、验收项。

## Assumptions & Decisions
- D1 **轮播图占位**：PC 暂无 banners 配置，`home.vue` 读 `banners` 字段但为空时回退占位图（商家封面或前端内置渐变占位 swiper），不做后端表新建；后续 PC 配置再加 `banners` 表与商户管理接口。
- D2 **预约订单**：解释为「我的订单按业务/状态分组 + 服务类订单展示预约/服务信息」，不在本期新造独立预约下单流程（避免超范围）。
- D3 **tabBar 四 tab**：首页/分类/购物车/我的；个人中心承担订单、健康、地址、联系商家聚合。既有非 tab 页面保留在 `pages/store/`。
- D4 **icon 分类宫格**：本期用现有 5 类商品类型配色图标映射，不新增分类 icon 数据字段（PC 分类 icon 后续加）。
- D5 **兼容性**：本期以新增/扩展为主，不改动订单核心状态机与支付链路语义；续租复用既有校验与支付。
- D6 每个阶段独立验证、分步提交（>10 文件提交版本控制）。

## Verification
1. 阶段 A：小程序 `npm run build`（或 `npx uni build`）通过；真机/模拟器首页显示轮播占位＋icon 宫格；tabBar 四 tab 可切换；个人中心可进订单/健康/地址。
2. 阶段 B：租赁订单支付后，C 端详单显示到期时间与剩余天数；指派服务人员的服务单显示服务人员姓名；点击续租生成 `parent_order_id=原单`、`renew_flag=1` 新单并进入支付。
3. 阶段 C：8 个健康子模块列表拉取正常，健康项与服务订单联动字段可见。
4. 阶段 D：PRD 与代码一致，`PRD.md` 章节无旧描述残留。
5. `go build ./...` 通过；新路由 `/api/v1/user/orders/:id/renew` 注册 `/health` 200。
6. 文档一致性脚本（若存在）通过。

## File Path Reference
- 小程序：`miniprogram/src/pages.json`、`pages/store/home.vue`、`pages/store/category.vue`(新)、`pages/store/me.vue`(新)、`pages/store/cart.vue`、`pages/store/my-orders.vue`、`pages/store/order-detail.vue`、`types/index.ts`、`api/store.ts`、`api/index.ts`
- 后端：`server/internal/handlers/user/handler.go`、`server/cmd/server/main.go`、`server/internal/models/models.go`
- 文档：`PRD.md`、`docs/prd/PRD-功能说明.md`、`docs/prd/PRD-接口文档.md`、`docs/prd/PRD-测试与附录.md`