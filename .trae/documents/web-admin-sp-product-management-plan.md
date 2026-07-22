# web-admin 服务商端商品管理实施计划

## Summary

- 目标：基于现有 `web-admin/` 服务商后台，在“商家详情”页内增加商品管理能力，支持查看不同商家的商品分类、商品、商品规格，以及商品新增、编辑、删除、上下架、批量上下架等操作。
- 已确认决策：
  - 页面入口放在商家详情页内 Tab，不新增一级侧边栏菜单。
  - 首版范围为完整 CRUD + 规格编辑，对齐现有小程序商家端核心商品能力。
  - 需要批量上下架。
  - 不新增服务商专用商品接口，沿用现有 `/api/v1/merchant/*` 商品/分类/规格接口。
  - 为支持服务商端调用，现有商家商品接口需要兼容 `sp` 身份，并通过 `merchant_id` 指定目标商家。
  - 商品规格在商品编辑页内完整编辑，不做列表弹窗版本。

## Current State Analysis

### 1. web-admin 现有结构

- 路由定义在 `web-admin/src/router/index.ts`
  - 当前已有：工作台、商家列表、商家创建/详情/编辑、订单列表/详情、公告、分账历史、数据分析、服务商设置。
  - 没有商品管理或分类管理路由。
- 侧边栏菜单在 `web-admin/src/layouts/AppLayout.vue`
  - 当前只有一级菜单：工作台、商家列表、订单管理、公告管理、分账历史、数据分析、服务商设置。
  - `activeMenu` 只处理 merchants / orders / announcements 三类高亮。
- 商家详情页在 `web-admin/src/views/merchant/MerchantDetailView.vue`
  - 当前展示基础资料、支付配置、经营数据、图片资产。
  - 这是最合适的承载入口，可扩展成 `基础资料 / 商品管理 / 分类管理` Tab。
- 服务商 API 与类型定义分别在：
  - `web-admin/src/api/sp.ts`
  - `web-admin/src/types/sp.ts`
  - 当前没有 Product / Category / ProductSpec 对应类型与请求封装。

### 2. 小程序商家端现有商品能力

- 商品列表页：`miniprogram/src/pages/merchant/products/list.vue`
  - 已支持：分类筛选、关键词搜索、上架/下架筛选、分页加载、单个上下架、删除、批量上下架。
- 商品编辑页：`miniprogram/src/pages/merchant/products/edit.vue`
  - 已支持：商品基础信息、分类、价格、库存、图片、规格、排序等编辑。
- 分类页：`miniprogram/src/pages/merchant/categories.vue`
  - 已支持：分类列表、创建、编辑、删除。
- 小程序商品 API 封装在 `miniprogram/src/api/index.ts`
  - 已具备商品、分类、规格相关调用逻辑和字段结构，可作为 web-admin 业务模型参考。

### 3. 后端现有商品接口能力

- 路由入口：`server/cmd/server/main.go`
- 现有商家商品域接口已完整存在，路径在 `/api/v1/merchant/*`：
  - 分类：
    - `GET /merchant/categories`
    - `POST /merchant/categories`
    - `PUT /merchant/categories/:category_id`
    - `DELETE /merchant/categories/:category_id`
    - `POST /merchant/categories/sort`
  - 商品：
    - `GET /merchant/products`
    - `GET /merchant/products/:product_id`
    - `POST /merchant/products`
    - `PUT /merchant/products/:product_id`
    - `POST /merchant/products/:product_id/on-sale`
    - `POST /merchant/products/:product_id/off-sale`
    - `POST /merchant/products/batch-status`
    - `DELETE /merchant/products/:product_id`
    - `PUT /merchant/products/:product_id/stock`
  - 规格：
    - `GET /merchant/products/:product_id/specs`
    - `PUT /merchant/products/:product_id/specs`
    - `DELETE /merchant/products/:product_id/specs`
- 主要实现文件：
  - `server/internal/handlers/merchant/product.go`
  - `server/internal/handlers/merchant/product_specs_endpoints.go`

### 4. 当前的关键限制

- 商家商品接口挂在 `merchantGroup.Use(JWTAuth(), MerchantAuth())` 下。
- `MerchantAuth()` 在 `server/internal/middleware/auth.go` 中只允许 `user_type == "merchant"`。
- `GetMerchantID()` 直接取 token 中的 `user_id` 作为商家 ID。
- 服务商 token 的 `user_type` 为 `sp`，因此 web-admin 不能直接调用现有商家商品接口。
- 结论：
  - “不新增接口”的前提下，只能改造现有 `/api/v1/merchant/*` 接口的鉴权与 merchant_id 解析逻辑，使其兼容服务商 token。

## Proposed Changes

### 1. 后端：让现有商家商品接口兼容服务商身份

- 修改文件：
  - `server/internal/middleware/auth.go`
  - `server/internal/handlers/sp/announcement.go`
  - `server/internal/handlers/merchant/product.go`
  - `server/internal/handlers/merchant/product_specs_endpoints.go`
  - 如需要可新增辅助文件：`server/internal/handlers/merchant/context.go`
- 方案：
  - 保持接口路径不变，继续使用 `/api/v1/merchant/*`。
  - 不新增 `/api/v1/sp/products*` 类服务商商品路由。
  - 将商品/分类/规格相关路由的权限从“仅 MerchantAuth”调整为“merchant 或 sp 均可访问”。
- 推荐实现方式：
  - 新增一个面向商品域的 merchant 解析辅助方法，例如：
    - `resolveTargetMerchantID(c *gin.Context) (merchantID uint64, ok bool, errMsg string)`
  - 规则：
    - 如果 `user_type == "merchant"`：
      - 沿用原逻辑，从 token `user_id` 直接取 merchant_id。
    - 如果 `user_type == "sp"`：
      - 从 query 或 body 中读取 `merchant_id`。
      - 通过 `getCurrentServiceProviderID` 查出当前服务商。
      - 校验目标商家 `id = merchant_id AND service_provider_id = current_sp_id`。
      - 校验失败则返回无权限或商家不存在。
- 覆盖范围：
  - 分类接口全部兼容 `sp`。
  - 商品接口全部兼容 `sp`。
  - 规格接口全部兼容 `sp`。
- 注意：
  - 只对商品域相关接口做兼容，不扩大到全部 `/api/v1/merchant/*`。
  - 不修改其它商家业务接口的权限模型。

### 2. 后端：兼容 `merchant_id` 透传方式

- 修改文件：
  - `server/internal/handlers/merchant/product.go`
  - `server/internal/handlers/merchant/product_specs_endpoints.go`
- 方案：
  - 对 GET 列表/详情接口，从 query 读取 `merchant_id`。
  - 对 POST / PUT / DELETE / batch 接口，从 body 优先读取 `merchant_id`，再回退 query。
  - 对路径中仅包含 `product_id` 的接口，仍需先解析目标 merchant，再做商品归属校验。
- 这样前端可以统一在服务商模式下传入 `merchant_id`，避免多条页面逻辑分叉。

### 3. web-admin：在商家详情页增加 Tab 容器

- 修改文件：`web-admin/src/views/merchant/MerchantDetailView.vue`
- 方案：
  - 将当前商家详情页改为基于 tab 的页面结构：
    - `基础资料`
    - `商品管理`
    - `分类管理`
  - 默认显示 `基础资料`。
  - 切到 `商品管理` / `分类管理` 时，仍保留当前商家上下文、返回按钮和页面头。
- 实现建议：
  - 页面内使用 `el-tabs`。
  - `基础资料` 复用现有内容。
  - `商品管理`、`分类管理` 作为子组件嵌入。
- 原因：
  - 与用户要求一致。
  - 不新增一级菜单，不打乱现有服务商后台结构。

### 4. web-admin：新增服务商商品列表组件

- 新增文件建议：
  - `web-admin/src/views/merchant/components/MerchantProductsTab.vue`
- 功能范围：
  - 基于当前商家 `merchantId` 加载商品列表。
  - 分类筛选、关键词搜索、状态筛选。
  - 表格展示：商品图、名称、分类、价格、库存、销量、状态、规格摘要、更新时间。
  - 单个上下架、删除。
  - 批量上下架。
  - “新增商品”“编辑商品”入口。
- API 调用：
  - 使用现有 `/api/v1/merchant/products` 等接口。
  - 每次请求附带 `merchant_id`。
- 交互建议：
  - `el-form` + `el-select` + `el-input` 作为筛选区。
  - `el-table` + `el-pagination` 作为列表区。
  - 顶部按钮：
    - 新增商品
    - 批量上架
    - 批量下架
  - 删除需要二次确认。

### 5. web-admin：新增服务商分类管理组件

- 新增文件建议：
  - `web-admin/src/views/merchant/components/MerchantCategoriesTab.vue`
- 功能范围：
  - 查看当前商家全部分类。
  - 新增分类。
  - 编辑分类名称。
  - 删除分类。
  - 显示商品数。
- API 调用：
  - 使用现有 `/api/v1/merchant/categories*` 接口。
  - 附带 `merchant_id`。
- 交互建议：
  - 列表表格或卡片均可，但推荐 `el-table`，和商品页保持一致。
  - 新增/编辑使用 `el-dialog`。
  - 删除前提示分类下商品影响说明。
- 排序：
  - 首版可以先不做拖拽排序 UI。
  - 仍需在计划中保留后端兼容可能，但本轮 UI 可不暴露排序入口。
  - 这样与用户“完整 CRUD+规格”目标不冲突，并避免 PC 拖拽实现膨胀。

### 6. web-admin：新增商品编辑页

- 新增文件建议：
  - `web-admin/src/views/merchant/MerchantProductEditView.vue`
- 路由建议新增：
  - `/merchants/:id/products/new`
  - `/merchants/:id/products/:productId/edit`
- 作用：
  - 承载商品新增与编辑。
  - 在该页内完整维护规格。
- 表单内容对齐小程序商家端：
  - 商品名称
  - 分类
  - 描述
  - 售价
  - 划线价
  - 库存
  - 单位
  - 图片数组
  - 规格组与规格项
  - 排序值
  - 上架状态或保存后操作
- 上传能力：
  - 复用 web-admin 现有七牛上传工具 `web-admin/src/utils/qiniu.ts`
  - 图片交互参考已存在的商家 Logo/背景图上传模式。
- 规格交互建议：
  - 每个规格组支持：
    - 规格名
    - 多个规格项
    - 每个规格项名称、价格
  - 支持新增/删除规格组、规格项。
- 保存策略：
  - 新增时先创建商品，再根据返回的 `product_id` 调规格接口保存规格。
  - 编辑时先更新商品，再调用规格整体覆盖接口 `PUT /merchant/products/:product_id/specs`。

### 7. web-admin：补齐服务商商品域 API 封装

- 修改文件：
  - `web-admin/src/api/sp.ts`
  - `web-admin/src/types/sp.ts`
- 说明：
  - 虽然不新增后端接口，但 web-admin 仍需要新增前端 API 封装函数。
- 新增内容：
  - 分类：
    - `getMerchantCategories(merchantId)`
    - `createMerchantCategory(merchantId, data)`
    - `updateMerchantCategory(merchantId, categoryId, data)`
    - `deleteMerchantCategory(merchantId, categoryId)`
  - 商品：
    - `getMerchantProducts(merchantId, params)`
    - `getMerchantProduct(merchantId, productId)`
    - `createMerchantProduct(merchantId, data)`
    - `updateMerchantProduct(merchantId, productId, data)`
    - `deleteMerchantProduct(merchantId, productId)`
    - `merchantProductOnSale(merchantId, productId)`
    - `merchantProductOffSale(merchantId, productId)`
    - `merchantBatchUpdateProductStatus(merchantId, data)`
  - 规格：
    - `getMerchantProductSpecs(merchantId, productId)`
    - `updateMerchantProductSpecs(merchantId, productId, data)`
    - `deleteMerchantProductSpecs(merchantId, productId, data)`
- 类型新增：
  - `MerchantCategory`
  - `MerchantProduct`
  - `MerchantProductSpec`
  - `MerchantProductListResponse`
  - `MerchantProductUpsertPayload`
  - 以及与小程序一致的规格表单类型

### 8. web-admin：补齐商品管理相关路由

- 修改文件：`web-admin/src/router/index.ts`
- 新增路由：
  - `/merchants/:id/products/new`
  - `/merchants/:id/products/:productId/edit`
- 说明：
  - Tab 内嵌适合列表和分类页。
  - 商品编辑页更适合使用独立路由，避免详情页内弹层过重。
- 标题建议：
  - `新增商品`
  - `编辑商品`

### 9. web-admin：是否需要侧边栏菜单改动

- 修改文件：`web-admin/src/layouts/AppLayout.vue`
- 本轮决策：
  - 不新增一级菜单。
  - 仅保证新路由 `/merchants/:id/products/*` 仍高亮到 `/merchants`。
- 修改点：
  - `activeMenu` 继续把 `/merchants` 下子路由都归到商家列表高亮。

### 10. 文档同步

- 修改文件：
  - `PRD.md`
  - `docs/prd/PRD-功能说明.md`
  - `docs/prd/PRD-接口文档.md`
  - `docs/prd/PRD-测试与附录.md`
- 需要补充：
  - web-admin 商家详情新增商品管理与分类管理 Tab。
  - 服务商端商品管理首版范围。
  - 现有 `/api/v1/merchant/*` 商品/分类/规格接口兼容服务商身份的口径。
  - 验收项新增：服务商可代管指定商家的分类、商品、规格、上下架与删除。

## Assumptions & Decisions

- 决策：不新增后端服务商商品接口路径，继续沿用 `/api/v1/merchant/*`。
- 决策：只对商品/分类/规格域接口兼容 `sp` 身份，不扩大其它商家接口。
- 决策：服务商端调用现有商家商品接口时，必须显式传入 `merchant_id`。
- 决策：服务商端页面入口放在商家详情内 Tab。
- 决策：商品新增/编辑使用独立路由页面，分类与商品列表留在详情页 Tab 内。
- 决策：首版包含完整商品规格编辑。
- 决策：首版包含批量上下架。
- 决策：首版分类 UI 不强制做排序拖拽入口。
- 假设：现有商家商品接口返回结构足以支撑 web-admin，不需要额外字段。
- 假设：服务商商品代管不需要细粒度角色权限，本轮沿用当前“服务商身份 + 商家归属校验”模型。

## Verification Steps

- 后端接口验证
  - 服务商 token 调用 `/api/v1/merchant/categories?merchant_id=xxx` 可成功返回当前服务商名下商家的分类。
  - 服务商 token 调用 `/api/v1/merchant/products?merchant_id=xxx` 可成功返回商品列表。
  - 服务商 token 指向非本服务商名下 merchant_id 时被拒绝。
  - 商家 token 调用相同接口行为保持不变，无回归。
- 商品操作验证
  - 服务商可新增商品并保存基础信息。
  - 服务商可编辑商品并修改规格。
  - 服务商可单个上架/下架商品。
  - 服务商可批量上架/下架商品。
  - 服务商可删除商品。
- 分类操作验证
  - 服务商可查看分类列表。
  - 服务商可新增、编辑、删除分类。
  - 删除被商品引用的分类时行为符合现有后端约束。
- 前端页面验证
  - 商家详情页 Tab 切换正常。
  - 商品管理列表的筛选、分页、批量操作可用。
  - 商品编辑页图片上传、规格编辑、保存成功。
  - 二级商品编辑路由返回商家详情后仍保留当前商家上下文。
- 回归验证
  - 小程序商家端原有商品管理功能保持可用。
  - web-admin 现有商家详情、订单、公告等功能不受影响。
  - 通过 `npm run build`（web-admin）与 `go test ./...`（server）验证无新增编译错误。
