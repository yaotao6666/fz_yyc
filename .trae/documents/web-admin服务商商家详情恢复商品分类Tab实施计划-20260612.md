# web-admin 服务商商家详情恢复商品/分类 Tab 实施计划

## Summary

修复 `web-admin` 服务商后台 `商家详情` 页丢失 `商品管理`、`分类管理` 两个 Tab 的问题，并确保这两部分全部走现有服务商代理接口 `/api/v1/sp/merchants/:merchant_id/...`，让服务商可以在商家详情页内继续管理商品、分类与规格。

本次不新增后端接口，不新增独立商品路由页，采用“在详情页内恢复 Tab + 新增子组件”的方式完成修复，并在实施完成后重新构建 `web-admin/sp/` 产物。

## Current State Analysis

### 1. 当前详情页只保留基础资料，不再包含商品/分类管理

- 当前页面文件是 [MerchantDetailView.vue](file:///e:/yt/fz_yyc/web-admin/src/views/merchant/MerchantDetailView.vue)。
- 页面现状仅包含：
  - 商家基础指标卡片
  - 商家资料 `el-descriptions`
  - 图片资产上传
  - 二维码弹窗
  - 自提点管理入口
- 文件中没有 `el-tabs`、商品表格、分类表格，也没有任何商品/分类 CRUD 逻辑。
- 当前 `web-admin/src/views/merchant/` 目录下也只有：
  - `MerchantDetailView.vue`
  - `MerchantEditView.vue`
  - `MerchantListView.vue`
  - `MerchantPickupPointsView.vue`
- 这说明商品/分类能力不是“换了接口但没调通”，而是前端页面功能本身已经缺失。

### 2. 服务商代理后端接口已存在，可直接接入

- 服务商代理实现位于 [merchant_delegate.go](file:///e:/yt/fz_yyc/server/internal/handlers/sp/merchant_delegate.go)。
- 该文件通过 `runAsMerchant()` 将服务商身份代理成目标商家，再复用商家 handler，因此无需新增一套后端逻辑。
- 已存在的 SP 代理接口包括：
  - 分类：
    - `GET /api/v1/sp/merchants/:merchant_id/categories`
    - `POST /api/v1/sp/merchants/:merchant_id/categories`
    - `PUT /api/v1/sp/merchants/:merchant_id/categories/:category_id`
    - `DELETE /api/v1/sp/merchants/:merchant_id/categories/:category_id`
    - `POST /api/v1/sp/merchants/:merchant_id/categories/sort`
  - 商品：
    - `GET /api/v1/sp/merchants/:merchant_id/products`
    - `GET /api/v1/sp/merchants/:merchant_id/products/:product_id`
    - `POST /api/v1/sp/merchants/:merchant_id/products`
    - `PUT /api/v1/sp/merchants/:merchant_id/products/:product_id`
    - `POST /api/v1/sp/merchants/:merchant_id/products/:product_id/on-sale`
    - `POST /api/v1/sp/merchants/:merchant_id/products/:product_id/off-sale`
    - `POST /api/v1/sp/merchants/:merchant_id/products/batch-status`
    - `DELETE /api/v1/sp/merchants/:merchant_id/products/:product_id`
    - `PUT /api/v1/sp/merchants/:merchant_id/products/:product_id/stock`
  - 规格：
    - `GET /api/v1/sp/merchants/:merchant_id/products/:product_id/specs`
    - `PUT /api/v1/sp/merchants/:merchant_id/products/:product_id/specs`
    - `DELETE /api/v1/sp/merchants/:merchant_id/products/:product_id/specs`

### 3. `web-admin` 当前 `sp` API 封装缺少商品/分类/规格方法

- 现有文件 [sp.ts](file:///e:/yt/fz_yyc/web-admin/src/api/sp.ts) 已封装：
  - 商家详情
  - 商家编辑
  - 支付配置
  - 图片资产
  - 自提点
  - 订单
  - 公告
- 但没有任何商品、分类、规格相关的前端 API 方法。
- 这会直接导致详情页无法恢复商品/分类 Tab 逻辑。

### 4. `web-admin` 类型定义缺少商品/分类结构

- 当前类型文件是 [sp.ts](file:///e:/yt/fz_yyc/web-admin/src/types/sp.ts)。
- 已有类型仅覆盖商家、自提点、订单、公告等。
- 缺少：
  - 分类实体与表单载荷
  - 商品列表分页结构
  - 商品详情结构
  - 商品规格结构
  - 商品创建/更新载荷

### 5. 后端真实返回/入参结构已可确定

- 分类列表返回由 [product.go](file:///e:/yt/fz_yyc/server/internal/handlers/merchant/product.go) 的 `GetCategories()` 构造，核心字段包括：
  - `id`
  - `merchant_id`
  - `name`
  - `sort`
  - `status`
  - `created_at`
  - `updated_at`
  - `product_count`
- 分类排序接口 `SortCategories()` 的请求体为：
  - `{"categories":[{"id":1,"sort":10}]}`
- 商品返回结构由 `ProductResponse` 定义，核心字段包括：
  - `id`
  - `merchant_id`
  - `category_id`
  - `name`
  - `description`
  - `images`
  - `price`
  - `original_price`
  - `stock`
  - `unit`
  - `sales`
  - `sort`
  - `status`
  - `category_name`
  - `specs`
  - `created_at`
  - `updated_at`
- 商品创建/更新入参 `ProductRequest` 包括：
  - `category_id`
  - `name`
  - `description`
  - `images`
  - `price`
  - `original_price`
  - `stock`
  - `unit`
  - `sort`
  - `specs`
- 规格接口 [product_specs_endpoints.go](file:///e:/yt/fz_yyc/server/internal/handlers/merchant/product_specs_endpoints.go) 真实结构为：
  - 读取返回：`{ specs, skus }`
  - 写入请求：`{ specs: [{ id?, name, values }], skus: [] }`

### 6. 小程序商家端可作为交互和字段参考，但 PC 端需改为 SP 接口

- 参考文件：
  - [api/index.ts](file:///e:/yt/fz_yyc/miniprogram/src/api/index.ts#L349-L654)
  - [types/index.ts](file:///e:/yt/fz_yyc/miniprogram/src/types/index.ts)
- 小程序已具备分类 CRUD、商品 CRUD、批量上下架、库存更新等交互。
- 本次只复用其数据模型和交互思路，不直接复用请求地址。

### 7. 最终生效依赖 `web-admin/sp/` 构建产物

- [vite.config.ts](file:///e:/yt/fz_yyc/web-admin/vite.config.ts) 中 `build.outDir = 'sp'`。
- [package.json](file:///e:/yt/fz_yyc/web-admin/package.json) 中构建命令为 `npm run build`。
- 因此修复必须覆盖源码和构建产物两个层面，否则线上看到的仍会是旧页面。

## Proposed Changes

### A. 扩展 `web-admin/src/types/sp.ts`

#### 修改文件

- `web-admin/src/types/sp.ts`

#### 增加内容

- 分类类型：
  - `MerchantCategory`
  - `MerchantCategoryPayload`
  - `MerchantCategorySortItem`
- 商品列表查询类型：
  - `MerchantProductQuery`
- 商品与规格类型：
  - `MerchantProduct`
  - `MerchantProductListResponse`
  - `MerchantProductSpec`
  - `MerchantProductSpecOption`
  - `MerchantProductSpecsResponse`
  - `MerchantProductSpecsPayload`
- 商品编辑载荷：
  - `MerchantProductUpsertPayload`

#### 具体决策

- 商品状态按当前商家端后端语义处理：
  - `1` 视为上架
  - `0` 或 `2` 的展示文案在实现时以实际接口回包为准，但批量/单个上下架操作仍直接调用专用接口，不依赖前端自行推导状态写回。
- `MerchantProductUpsertPayload` 直接对齐后端 `ProductRequest`，保留 `specs` 字段，避免后续弹窗保存时再做临时拼装。
- `MerchantProductSpecsPayload` 额外对齐规格专用接口 `{ specs, skus }`，`skus` 固定传空数组。

### B. 扩展 `web-admin/src/api/sp.ts`

#### 修改文件

- `web-admin/src/api/sp.ts`

#### 增加内容

- 分类接口：
  - `getMerchantCategories(merchantId)`
  - `createMerchantCategory(merchantId, payload)`
  - `updateMerchantCategory(merchantId, categoryId, payload)`
  - `deleteMerchantCategory(merchantId, categoryId)`
  - `sortMerchantCategories(merchantId, categories)`
- 商品接口：
  - `getMerchantProducts(merchantId, params)`
  - `getMerchantProduct(merchantId, productId)`
  - `createMerchantProduct(merchantId, payload)`
  - `updateMerchantProduct(merchantId, productId, payload)`
  - `merchantProductOnSale(merchantId, productId)`
  - `merchantProductOffSale(merchantId, productId)`
  - `batchUpdateMerchantProductStatus(merchantId, productIds, status)`
  - `deleteMerchantProduct(merchantId, productId)`
  - `updateMerchantProductStock(merchantId, productId, stock)`
- 规格接口：
  - `getMerchantProductSpecs(merchantId, productId)`
  - `updateMerchantProductSpecs(merchantId, productId, payload)`

#### 具体决策

- 所有请求统一使用 `/api/v1/sp/merchants/${merchantId}/...`。
- 分类排序方法按后端真实请求体发送 `categories`，不沿用小程序里的 `orders`。
- 商品保存时图片字段直接传数组，具体 URL 清洗逻辑复用现有前端上传处理方式，不在 API 方法里做多余推断。

### C. 重构 `web-admin/src/views/merchant/MerchantDetailView.vue`

#### 修改文件

- `web-admin/src/views/merchant/MerchantDetailView.vue`

#### 修改方式

- 将当前单页布局改为三 Tab 结构：
  - `基础资料`
  - `商品管理`
  - `分类管理`
- `基础资料` Tab 中保留现有能力：
  - 指标卡片
  - 商家资料
  - 图片资产
  - 查看二维码
  - 自提点管理入口
- `商品管理` Tab 挂载新组件 `MerchantProductsTab`
- `分类管理` Tab 挂载新组件 `MerchantCategoriesTab`

#### 具体决策

- 页面继续以 `merchantId` 为核心上下文，不新增详情页子路由。
- 基础资料的数据请求仍由当前页面统一负责，商品/分类各自组件独立拉取各自数据，减少页面耦合。
- 如需保持 Tab 切换状态，可通过本地 `ref(activeTab)` 控制；当前不要求与 URL query 同步。

### D. 新增 `MerchantProductsTab.vue`

#### 新增文件

- `web-admin/src/views/merchant/components/MerchantProductsTab.vue`

#### 负责内容

- 拉取分类选项与商品列表
- 提供筛选条件：
  - 分类
  - 状态
  - 关键词
- 展示商品表格：
  - 图片
  - 名称
  - 分类
  - 价格
  - 库存
  - 销量
  - 状态
  - 创建时间或排序
  - 操作列
- 支持操作：
  - 新增商品
  - 编辑商品
  - 单个上架/下架
  - 批量上架/下架
  - 删除商品
  - 快速修改库存

#### 具体决策

- 批量操作通过表格勾选项 + 顶部操作按钮完成。
- 快速改库存采用轻量交互：
  - 首选 `ElMessageBox.prompt`
  - 如果实现复杂度更低，也可以在行内打开小弹窗输入，但不新增独立页面。
- 商品列表分页结构按后端列表响应实现；如果当前接口未返回完整 `pagination`，则按已有返回结构做兼容兜底，但计划默认以后端标准分页结构为准。

### E. 新增 `MerchantProductEditorDialog.vue`

#### 新增文件

- `web-admin/src/views/merchant/components/MerchantProductEditorDialog.vue`

#### 负责内容

- 商品新增/编辑弹窗
- 表单字段：
  - `name`
  - `category_id`
  - `description`
  - `images`
  - `price`
  - `original_price`
  - `stock`
  - `unit`
  - `sort`
- 规格编辑：
  - 规格名
  - 规格值列表
- 图片上传：
  - 复用 [qiniu.ts](file:///e:/yt/fz_yyc/web-admin/src/utils/qiniu.ts)

#### 具体决策

- 使用 `ElDialog`，不新增路由页。
- 编辑时加载顺序为：
  1. 读取商品详情 `getMerchantProduct`
  2. 读取规格 `getMerchantProductSpecs`
  3. 组合成弹窗表单模型
- 保存顺序固定为：
  1. 创建或更新商品主体
  2. 再调用规格保存接口
- 这样可以兼容当前后端“商品主体”和“规格专用接口”分离的实现。
- 若规格区为空，仍按空 `specs` 提交规格保存，保证删除规格的场景能落库。

### F. 新增 `MerchantCategoriesTab.vue`

#### 新增文件

- `web-admin/src/views/merchant/components/MerchantCategoriesTab.vue`

#### 负责内容

- 展示分类列表：
  - 名称
  - 商品数
  - 排序
  - 状态
  - 创建时间
  - 操作
- 支持：
  - 新增分类
  - 编辑分类
  - 删除分类
  - 调整排序并保存
  - 修改状态

#### 具体决策

- 新增/编辑分类使用同一个弹窗表单。
- 状态切换通过 `updateMerchantCategory()` 完成，不新增单独状态接口。
- 排序采用“表格内数字输入 + 点击保存排序”方案，调用 `sortMerchantCategories()`，请求体形如：

```json
{
  "categories": [
    { "id": 1, "sort": 10 },
    { "id": 2, "sort": 20 }
  ]
}
```

### G. 视需要补充格式化工具

#### 可能修改文件

- `web-admin/src/utils/format.ts`

#### 修改内容

- 若当前缺少商品状态、分类状态展示方法，则新增：
  - `getProductStatusText`
  - `getCategoryStatusText`

#### 具体决策

- 状态映射只做展示，不影响真正的状态写操作。
- 若后端只存在 `1=启用/上架` 与其他值统称停用/下架，则保持简单映射，不引入过度枚举。

### H. 构建与验收

#### 涉及目录

- `web-admin/src/**`
- `web-admin/sp/**`

#### 操作要求

- 修改源码后执行 `npm run build`
- 确认 `web-admin/sp/` 被重新生成

## Assumptions & Decisions

- 本次目标是“恢复详情页内原有的商品管理/分类管理能力”，不是新建一套独立商品后台。
- 后端 SP 代理接口已经具备，实施范围默认不改后端。
- 当前 `merchant` 目录下不存在可复用的商品/分类组件，因此直接在 `web-admin/src/views/merchant/components/` 新增组件。
- 商品编辑采用弹窗而不是新路由，以降低改造面并符合“详情页内管理”的现有交互预期。
- 图片上传继续复用现有七牛上传工具，不重新设计上传链路。
- 最终交付以源码修复 + `web-admin/sp` 构建完成为准。

## Verification Steps

### 1. 静态检查

- `web-admin` 相关新增类型无明显冲突。
- 新增组件与 `MerchantDetailView.vue` 的 props / emits 对接完整。

### 2. 构建验证

- 在 `web-admin` 目录执行 `npm run build` 成功。
- `web-admin/sp/` 产物更新时间被刷新。

### 3. 页面结构验证

- 打开 `商家详情` 页面后能看到：
  - `基础资料`
  - `商品管理`
  - `分类管理`
- `基础资料` Tab 的现有内容不回退：
  - 详情展示正常
  - 图片上传入口正常
  - 二维码弹窗正常
  - 自提点管理按钮正常

### 4. 分类管理验证

- 服务商进入商家详情后，分类列表能成功加载。
- 新增分类成功并刷新列表。
- 编辑分类成功并正确回显。
- 删除分类成功。
- 调整排序后保存成功。
- 状态切换成功。
- 抓包确认调用的是 `/api/v1/sp/merchants/:merchant_id/categories...`

### 5. 商品管理验证

- 服务商进入商家详情后，商品列表能成功加载。
- 分类筛选、状态筛选、关键词搜索可用。
- 新增商品成功。
- 编辑商品成功。
- 图片上传并回显成功。
- 规格保存成功，重新打开编辑时可回填。
- 单个商品上架/下架成功。
- 批量上架/批量下架成功。
- 删除商品成功。
- 库存修改成功。
- 抓包确认调用的是 `/api/v1/sp/merchants/:merchant_id/products...`

### 6. 回归验证

- 不再出现对以下商家身份接口的直接调用：
  - `/api/v1/merchant/categories`
  - `/api/v1/merchant/products`
