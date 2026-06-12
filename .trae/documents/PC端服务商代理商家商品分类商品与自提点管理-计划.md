# PC端服务商代理商家商品分类/商品 + 自提点管理：实施计划

## Summary

实现两件事：

1. **修复/补齐服务商身份（PC端 web-admin）无法管理商家“商品分类/商品”**的问题：通过新增 **SP 代理接口**，让服务商在选定某个商家后，可以用同一套请求/响应结构复用商家端接口能力（分类/商品/规格等），但由后端校验该商家归属当前服务商。
2. **PC 端支持修改商家的自提点配置**：在 web-admin 增加独立路由页面 `/merchants/:id/pickup-points`，支持自提点 **增删改 + 设默认 + 启用/停用**，调用同样的 SP 代理接口完成落库。

用户已确认的关键决策：
- 商品/分类：**只做后端兼容**（PC 暂不做完整商品管理 UI）
- 接口形态：**新增 SP 代理接口（推荐方案）**
- 自提点：PC 端做 **独立路由页面**，支持 **增删改 + 默认 + 启停**

## Current State Analysis

### 1) 后端权限模型导致 SP 无法调用商家接口

- 商家接口统一挂在 `merchantGroup`，并使用 `middleware.MerchantAuth()`（要求 `user_type == merchant`）：
  - 路由位置：[main.go](file:///e:/yt/fz_yyc/server/cmd/server/main.go#L139-L216)
- 商家接口中的 `middleware.GetMerchantID(c)` 本质上返回 `user_id`：
  - [auth.go](file:///e:/yt/fz_yyc/server/internal/middleware/auth.go#L150-L166)
- 服务商 token 的 `user_id` 是服务商后台账号（`service_provider_sps.id`），并不等于 `merchant_id`，因此无法直接调用 `/api/v1/merchant/...`。

### 2) 分类/商品/规格接口目前仅商家端可用

- 分类接口在商家 handler 内实现：
  - [GetCategories/CreateCategory/UpdateCategory/DeleteCategory/SortCategories](file:///e:/yt/fz_yyc/server/internal/handlers/merchant/product.go#L197-L355)
- 商品接口在商家 handler 内实现：
  - [GetProducts/CreateProduct/UpdateProduct/上下架/库存/删除](file:///e:/yt/fz_yyc/server/internal/handlers/merchant/product.go#L340-L657)
- 规格接口：
  - [product_specs_endpoints.go](file:///e:/yt/fz_yyc/server/internal/handlers/merchant/product_specs_endpoints.go)

### 3) 自提点接口已存在（商家端 + C端）

- 商家端自提点 CRUD：
  - [merchant/pickup_points.go](file:///e:/yt/fz_yyc/server/internal/handlers/merchant/pickup_points.go)
- C 端下单页自提点公开列表：
  - [user/pickup_points.go](file:///e:/yt/fz_yyc/server/internal/handlers/user/pickup_points.go)

### 4) web-admin（PC端）现状

- web-admin 当前仅有商家列表/详情/订单/公告等路由，没有自提点管理页：
  - 路由：[router/index.ts](file:///e:/yt/fz_yyc/web-admin/src/router/index.ts)
- web-admin API 只封装了 SP 相关接口，无自提点接口：
  - [api/sp.ts](file:///e:/yt/fz_yyc/web-admin/src/api/sp.ts)

## Proposed Changes

### A) 后端：新增 SP 代理商家接口（分类/商品/规格/自提点）

#### 目标

在不改动现有商家 handlers（避免大面积改造）的前提下，为服务商后台提供一套代理路由：

- 路径形态：`/api/v1/sp/merchants/:merchant_id/...`
- 访问控制：仅允许访问 **当前服务商名下** 的商家
- 行为：通过 **临时注入 `c.Set("user_id", merchantID)`** 的方式复用商家 handler 逻辑，使 `middleware.GetMerchantID()` 返回正确的商家 ID

#### 具体实现

1. 新增一个统一的“代理执行器”
   - 文件：`server/internal/handlers/sp/merchant_delegate.go`（新文件）
   - 核心逻辑：
     - 解析 `merchant_id`（path param）
     - 使用 `getCurrentServiceProviderID(c)` 获取当前服务商 ID（现有函数）
     - 校验 `merchants.id == merchant_id AND merchants.service_provider_id == 当前服务商`
     - 保存原始 context 的 `user_id/user_type/username`，然后：
       - `c.Set("user_id", merchantID)`
       - 可选：`c.Set("user_type", "merchant")`（降低未来 merchant handler 引入 user_type 判断的风险）
     - 调用目标商家 handler（直接调用函数）
     - `defer` 恢复原始 context 值

2. 在 `server/cmd/server/main.go` 增加 SP 路由（在 `spGroup` 内）
   - 分类代理（复用 merchant/product.go）：
     - `GET    /api/v1/sp/merchants/:merchant_id/categories` → `merchant.GetCategories`
     - `POST   /api/v1/sp/merchants/:merchant_id/categories` → `merchant.CreateCategory`
     - `PUT    /api/v1/sp/merchants/:merchant_id/categories/:category_id` → `merchant.UpdateCategory`
     - `DELETE /api/v1/sp/merchants/:merchant_id/categories/:category_id` → `merchant.DeleteCategory`
     - `POST   /api/v1/sp/merchants/:merchant_id/categories/sort` → `merchant.SortCategories`
   - 商品代理（复用 merchant/product.go）：
     - `GET    /api/v1/sp/merchants/:merchant_id/products`
     - `GET    /api/v1/sp/merchants/:merchant_id/products/:product_id`
     - `POST   /api/v1/sp/merchants/:merchant_id/products`
     - `PUT    /api/v1/sp/merchants/:merchant_id/products/:product_id`
     - `POST   /api/v1/sp/merchants/:merchant_id/products/:product_id/on-sale`
     - `POST   /api/v1/sp/merchants/:merchant_id/products/:product_id/off-sale`
     - `POST   /api/v1/sp/merchants/:merchant_id/products/batch-status`
     - `DELETE /api/v1/sp/merchants/:merchant_id/products/:product_id`
     - `PUT    /api/v1/sp/merchants/:merchant_id/products/:product_id/stock`
   - 商品规格代理（复用 merchant/product_specs_endpoints.go）：
     - `GET    /api/v1/sp/merchants/:merchant_id/products/:product_id/specs`
     - `PUT    /api/v1/sp/merchants/:merchant_id/products/:product_id/specs`
     - `DELETE /api/v1/sp/merchants/:merchant_id/products/:product_id/specs`
   - 自提点代理（复用 merchant/pickup_points.go）：
     - `GET    /api/v1/sp/merchants/:merchant_id/pickup-points`
     - `POST   /api/v1/sp/merchants/:merchant_id/pickup-points`
     - `PUT    /api/v1/sp/merchants/:merchant_id/pickup-points/:id`
     - `DELETE /api/v1/sp/merchants/:merchant_id/pickup-points/:id`

3. 错误语义
   - 商家不属于当前服务商：返回 `404 商家不存在`（与现有 SP 侧 merchant 查询一致）
   - 其余业务校验与返回结构：完全沿用商家 handler（确保“兼容使用商家管理的接口”的数据结构）

### B) web-admin：新增“商家自提点管理”页面（独立路由）

#### 目标

PC 端可在服务商后台对某个商家维护自提点列表，能力对齐小程序商家端：

- 自提点列表展示（名称/地址/经纬度/默认/状态）
- 新增/编辑（表单录入：名称、地址、lat、lng、默认、状态）
- 删除（二次确认）
- “设为默认”（本质走 update 接口 `is_default=true`）
- “启用/停用”（update `status=1/0`）

#### 具体改造点

1. 路由
   - 文件：[router/index.ts](file:///e:/yt/fz_yyc/web-admin/src/router/index.ts)
   - 新增路由：`/merchants/:id/pickup-points`
   - 页面组件：`web-admin/src/views/merchant/MerchantPickupPointsView.vue`（新文件）

2. API 封装
   - 文件：[api/sp.ts](file:///e:/yt/fz_yyc/web-admin/src/api/sp.ts)
   - 新增：
     - `getMerchantPickupPoints(merchantId)`
     - `createMerchantPickupPoint(merchantId, payload)`
     - `updateMerchantPickupPoint(merchantId, pointId, payload)`
     - `deleteMerchantPickupPoint(merchantId, pointId)`
   - 对应调用后端：`/api/v1/sp/merchants/:merchant_id/pickup-points...`

3. 类型定义
   - 文件：[types/sp.ts](file:///e:/yt/fz_yyc/web-admin/src/types/sp.ts)
   - 新增 `MerchantPickupPoint` / `MerchantPickupPointPayload` 类型（与后端字段一致）

4. 入口
   - 文件：[MerchantDetailView.vue](file:///e:/yt/fz_yyc/web-admin/src/views/merchant/MerchantDetailView.vue)
   - 在页面头部按钮区增加：`自提点管理` → 跳转 `/merchants/${merchantId}/pickup-points`
   - （可选）在商家列表操作列增加快捷入口（若你希望更顺手）

### C) Verification

1. 后端
   - `go test ./...` 通过
   - 用 SP token 调用：
     - `GET /api/v1/sp/merchants/:merchant_id/categories` 能返回分类列表
     - `GET /api/v1/sp/merchants/:merchant_id/products` 能返回商品列表
     - `GET /api/v1/sp/merchants/:merchant_id/pickup-points` 能返回自提点列表
   - 使用非归属商家 merchant_id：返回 “商家不存在”

2. web-admin
   - 新增路由可正常访问（已登录前提）
   - 自提点增删改/设默认/启停可联调成功
   - `npm run build`（或项目现有构建命令）通过

## Notes / Non-Goals

- 本次不在 PC 端实现完整商品管理 UI（用户已选择“只做后端兼容”）。
- 商家原有 `/api/v1/merchant/...` 权限逻辑不改变，避免扩大影响面。

