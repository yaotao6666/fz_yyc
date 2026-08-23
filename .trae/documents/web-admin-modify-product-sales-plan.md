# web-admin 商品销量修改功能实施计划

## Summary

- 目标：在 PC 端 web-admin 的商品编辑页增加手动修改商品销量的能力，仅限商家访问。
- 已确认决策：
  - 修改入口：在商品编辑页（MerchantProductEditView）增加“销量”输入框。
  - 权限范围：仅商家可以通过 web-admin 修改任意商家的商品销量，商家端保持只读。
  - 接口方式：复用现有 PUT /api/v1/merchant/products/:id 接口，增加 sales 字段支持。
  - 列表页保持只读显示，不提供行内编辑入口。

## Current State Analysis

### 1. 后端现状

- 商品模型在 `server/internal/models/models.go` 中已有 `Sales` 字段（累计销量）：
  ```go
  Sales uint `gorm:"not null;default:0;comment:累计销量" json:"sales"`
  ```
- 现有更新接口 `UpdateProduct`（`server/internal/handlers/merchant/product.go`）不支持 sales 字段：
  ```go
  func UpdateProduct(c *gin.Context) {
      updates := map[string]interface{}{
          "category_id": req.CategoryID,
          "name": req.Name,
          "description": req.Description,
          "images": models.JSON(imagesJSON),
          "price": req.Price,
          "original_price": req.OriginalPrice,
          "stock": req.Stock,
          "unit": req.Unit,
          "sort": req.Sort,
          // 没有 "sales" 字段
      }
  }
  ```
- 商品更新接口当前使用 `MerchantAuth` 鉴权，仅允许商家访问。

### 2. 前端现状

- 商品列表页（`web-admin/src/views/merchant/components/MerchantProductsTab.vue`）已经正确显示销量列：
  ```vue
  <el-table-column label="销量" width="100">
    <template #default="scope">
      {{ scope.row.sales || 0 }}
    </template>
  </el-table-column>
  ```
- 商品编辑页（`web-admin/src/views/merchant/MerchantProductEditView.vue`）没有销量字段输入/显示。

### 3. 权限现状

- web-admin 目前使用商家 token，权限通过 `SpAuth` 中间件校验。
- 商品相关接口挂在 `merchantGroup` 下，当前只允许商家访问，商家无法直接调用。
- 上一轮已实现商品域接口兼容商家 token（通过 `resolveTargetMerchantID` 解析目标商家并校验归属），但更新接口本身没有 sales 字段。

## Proposed Changes

### 1. 后端：在 UpdateProduct 接口增加 sales 字段支持

- 修改文件：
  - `server/internal/handlers/merchant/product.go`
  - `server/internal/handlers/merchant/context.go`（已存在，确保 resolveTargetMerchantID 已支持 sp 身份）
- 改动点：
  - 在 `MerchantProductUpsertRequest` 结构体中增加 `Sales *uint` 字段。
  - 在 `UpdateProduct` 函数中，如果 `req.Sales != nil`，则把 `sales` 加入 updates map。
  - 保持现有业务逻辑：sales 可以为 0 或任意正整数。
- 说明：
  - 不新增接口，直接复用现有 PUT 接口。
  - 前端在编辑页提交时带上 `sales` 字段即可。

### 2. 前端：在商品编辑页增加销量输入框（仅商家可见）

- 修改文件：
  - `web-admin/src/views/merchant/MerchantProductEditView.vue`
- 改动点：
  - 在商品基础信息表单中，在“排序值”字段后面增加“销量”输入框。
  - 默认值为 `0`。
  - 字段类型为 `el-input-number`，限制最小值为 0。
  - 增加 `v-if="isServiceProvider"`，仅商家可见。
  - 字段校验：必须为非负整数。
- 说明：
  - 列表页保持只读显示，不增加编辑入口，避免交互复杂度。
  - 商家端小程序商品编辑页保持不变（不包含销量字段）。

### 3. 前端：更新 API 类型定义

- 修改文件：
  - `web-admin/src/types/sp.ts`
- 改动点：
  - 在 `MerchantProductUpsertRequest` 接口中增加 `sales?: number` 字段。
- 说明：
  - 确保 TypeScript 类型与后端请求结构一致。

### 4. 文档同步

- 修改文件：
  - `PRD.md`
  - `docs/prd/PRD-功能说明.md`
- 补充内容：
  - 说明 web-admin 支持商家手动修改商品销量，仅限 PC 端。
  - 说明商家端保持销量只读，由订单系统自动更新。
  - 说明修改权限仅限商家角色。

## Assumptions & Decisions

- 决策：复用现有 PUT 接口，不新增专用接口。
- 决策：仅商家可见，商家端不提供此能力。
- 决策：列表页只读显示，不在列表页增加行内编辑入口。
- 决策：商品编辑页的销量字段默认值为 0，用户可以修改为任意非负整数。
- 假设：商家修改销量是运营需求，不需要记录修改历史（如需可后续扩展）。

## Verification Steps

### 后端接口验证
- 商家 token 调用 `PUT /api/v1/merchant/products/:id` 时，带上 `sales` 字段可成功更新。
- 商家 token 调用时不带 `sales` 字段时，接口行为与之前保持一致（不更新销量）。
- 商家 token 调用接口时，可以正常更新商品信息（不包含 sales）。
- 商家 token 指向非本商家名下商家时被拒绝。

### 前端页面验证
- 商家登录 web-admin，进入商品编辑页，能看到“销量”输入框。
- 不带 `sales` 字段保存时，销量保持原值不变。
- 带上 `sales` 字段保存后，商品列表页的销量列正确更新。
- 商家登录小程序商家端，商品编辑页没有“销量”字段。
- 商家端商品列表页的销量列只读显示。

### 构建与诊断
- 后端 `go test ./...` 通过。
- `web-admin` `npm run build` 通过。
- 所有新增/修改文件诊断通过。

## Implementation Steps

1. **后端：在 UpdateProduct 中增加 sales 字段支持**
   - 修改 `MerchantProductUpsertRequest` 结构体，增加 `Sales *uint`。
   - 在 `UpdateProduct` 函数中，如果 `req.Sales != nil`，则加入 updates map。
   - 保持向后兼容，不传 sales 时行为不变。

2. **后端：确认 resolveTargetMerchantID 已支持 sp 身份**
   - 检查 `server/internal/handlers/merchant/context.go` 中的 `resolveTargetMerchantID` 函数。
   - 确认商家 token 可以通过 `merchant_id` 指定目标商家并校验归属。

3. **前端：更新 API 类型定义**
   - 在 `web-admin/src/types/sp.ts` 的 `MerchantProductUpsertRequest` 中增加 `sales?: number`。

4. **前端：在商品编辑页增加销量输入框**
   - 在 `MerchantProductEditView.vue` 中，在排序值字段后增加 `el-input-number`。
   - 添加 `v-if="isServiceProvider"` 限制仅商家可见。
   - 设置默认值为 0，最小值为 0。

5. **文档同步**
   - 在 `PRD.md` 中增加 web-admin 商品销量修改能力说明。
   - 在 `docs/prd/PRD-功能说明.md` 中补充商家端商品管理能力细节。

6. **验证与构建**
   - 运行后端测试和 `web-admin` 构建。
   - 检查所有新增/修改文件诊断。
