# 首页推荐商品/服务模块 Spec

## Why
用户端小程序首页 `pages/store/home` 目前顶部展示店铺「5星评分」，下方是完整的「左侧分类 + 右侧商品列表」浏览区（含热销推荐）。为让首页定位为「精选展示」而非「完整商品浏览」，需要：去掉 5 星评分，去掉下方的分类+商品浏览区，改为展示由 PC 端统一配置的「推荐商品/服务」模块。

## What Changes
- **C端 `pages/store/home`**：删除店铺头部的 5 星评分显示与相关样式；删除下方「分类和商品」整体浏览区（左侧分类栏 + 右侧商品列表 + 热销推荐），替换为「推荐商品/服务」模块，从后端拉取 PC 配置的推荐项并渲染。
- **PC 端**：新增「首页推荐」管理页，可配置推荐项（指向实物商品或服务）、设置展示标题、排序、启用/禁用。
- **后端**：新增 `store_home_recommends` 表；新增商家后台推荐管理接口（增/删/改/查/启停）；新增 C端 首页推荐查询接口（带商品/服务信息快照）。
- **RBAC**：新增「首页推荐」菜单节点及按钮权限码，绑定超管角色。
- **BREAKING**：C端首页不再具备内嵌的分类浏览能力；商品/服务挑选能力仍保留在金刚区 icon 跳转的内页（零售/租赁 → `product-list`，套餐 → `service-selection`）与对应详情页。

## Impact
- Affected code:
  - `miniprogram/src/pages/store/home.vue`（删除评分、删除分类+商品浏览、新增推荐模块）
  - 后端新文件：`server/internal/handlers/merchant/home_recommend.go`、`server/internal/handlers/user/handler.go`（或新增 C端推荐接口）
  - 后端模型：`server/internal/models/models.go`（新增 Recommend 模型）
  - 路由：`server/cmd/server/main.go`（商户端 + C端）
  - 迁移：`server/migrations/029_home_recommend.sql`
  - PC 端：`web-admin/src/views/miniprogram/HomeRecommendView.vue`、路由、API
- Affected specs: 商户订单列表拆分、小程序订单列表拆分（推荐商品与商品分类体系的隔离逻辑应保持一致：1/2=实物，3/4=服务）。

## ADDED Requirements
### Requirement: 首页推荐表与 PC 推荐管理
系统 SHALL 提供一张商家维度的首页推荐表 `store_home_recommends`，记录：推荐对象 `product_id`（引用 `products.id`）、`target_type`（1=实物商品=产品类型1/2，2=服务=产品类型3/4）、展示标题、排序、启停状态，并归属 `merchant_id`。

PC 后台 SHALL 提供「首页推荐」管理页与增删改查/启停接口，并按商家隔离数据。

#### Scenario: PC 配置推荐商品
- **WHEN** 管理员登录 PC，进入「首页推荐」，新增一个推荐项并选择某个已上架的实物商品（零售/租赁）或服务（康养套餐/陪诊服务），设置排序并启用
- **THEN** 该推荐项成功入库；再次打开列表可看到该推荐项及目标商品/服务信息；可编辑标题/排序/启停或删除。

### Requirement: C端首页推荐展示
C端 `pages/store/home` SHALL 在移除评分与分类浏览区后，渲染「推荐商品/服务」模块，数据来自 C端首页推荐接口（仅返回当前商家已启用、按排序升序的推荐项，并附带商品/服务名称、封面、价格、类型等快照）。

#### Scenario: 首页展示推荐项并跳转
- **WHEN** 用户进入首页，商家已配置启用若干推荐项
- **THEN** 首页下方展示「推荐」模块，卡片显示商品/服务信息；点击实物卡片跳转对应商品详情，点击服务卡片跳转对应服务详情/挑选页；无任何启用推荐项时展示空态「暂无可推荐商品」。

### Requirement: 首页推荐接口权限
C端推荐查询接口无需登录商户权限（用户可见）；PC 管理接口与「首页推荐」菜单应用 RBAC 权限码 `home-recommend:view/create/update/delete/status`。

## MODIFIED Requirements
### Requirement: C端首页布局
首页头部下方保留：金刚区（轮播图 + icon 分类宫格）、领券中心、健康宣教；移除店铺评分星星展示与本需求范围之外的分类+商品浏览区块（该区块由新「推荐商品/服务」模块取代）。门店名称、电话、地址、公告等基础信息与交互保持不变。

### Requirement: 商品分类体系隔离
推荐对象的复用遵循既有规则：实物商品（产品类型 1/2）与实物分类一致，服务（产品类型 3/4）与服务管理一致，推荐不跨类绑定。

## REMOVED Requirements
### Requirement: 首页分类商品浏览区
**Reason**：首页改版定位为精选展示，改由 PC 配置的推荐模块呈现商品/服务入口。
**Migration**：分类挑选能力保留在金刚区 icon 跳转的内页与商品/服务详情页；`products` 表结构不变，仅首页不再内嵌分类浏览。