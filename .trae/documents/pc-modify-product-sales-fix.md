# PC 端修改商品销量 - 实施计划

## Summary

修复 web-admin 商品编辑页销量输入框不显示的 bug，使商家可在 PC 端商品编辑页修改商品销量。后端、API、类型定义、表单数据流均已就绪，仅需修复前端 `isServiceProvider` 未定义问题。

## Current State Analysis

### 已就绪部分（无需改动）

1. **后端 `server/internal/handlers/merchant/product.go`**
   - `ProductRequest.Sales` 字段已存在（指针类型 `*uint`，第 440 行）
   - `UpdateProduct` 已处理 sales：`if req.Sales != nil { updates["sales"] = *req.Sales }`（第 546-548 行）
   - 鉴权通过 `MerchantOrSpAuth` 中间件 + `resolveTargetMerchantID`，商家可通过 `merchant_id` 指定目标商家

2. **前端 API `web-admin/src/api/sp.ts`**
   - `updateMerchantProduct` 已发送 sales 字段（第 115-119 行）

3. **前端类型 `web-admin/src/types/sp.ts`**
   - `MerchantProductUpsertPayload.sales?: number` 已定义（第 188 行）

4. **前端表单数据流 `web-admin/src/views/merchant/MerchantProductEditView.vue`**
   - `form.sales` 字段已初始化为 0（第 48 行）
   - 加载时回填 `form.sales = Number(product.sales || 0)`（第 87 行）
   - 提交 payload 包含 `sales: form.sales`（第 157 行）
   - UI 输入框已写好（第 241-243 行），但被 `v-if="isServiceProvider"` 屏蔽

### Bug：`isServiceProvider` 未定义

- `MerchantProductEditView.vue` 第 241 行 `v-if="isServiceProvider"` 引用了一个**从未定义**的变量
- 该文件的 `<script setup>`（第 1-197 行）没有 `isServiceProvider` 的定义
- `web-admin/src/stores/auth.ts` 的 `useAuthStore` 也**没有** `isServiceProvider` getter（只有 `isAuthenticated`、`serviceProviderName`、`operatorName`）
- 在 Vue 3 `<script setup>` 中模板引用未定义变量视为 `undefined`，`v-if` 永远为 falsy
- **后果**：销量输入框永远不显示，整个修改销量功能在前端实际不可用

### 上下文

web-admin 本身就是**商家后台**（登录者必为商家），不存在商家/商家角色切换，因此 `isServiceProvider` 在该后台内恒为 `true`。

## Proposed Changes

### 唯一改动：修复 `MerchantProductEditView.vue` 中的 `isServiceProvider` 定义

- **文件**：[web-admin/src/views/merchant/MerchantProductEditView.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc/web-admin/src/views/merchant/MerchantProductEditView.vue)
- **改动位置**：`<script setup>` 块内，建议放在 `productId` 计算属性之后（第 32 行后）
- **改动内容**：新增一行
  ```ts
  // web-admin 为商家后台，登录者必为商家，故恒为 true
  const isServiceProvider = true
  ```
- **理由**：
  - web-admin 不存在商家角色，无需运行时判断
  - 与项目最小改动原则一致，避免引入 auth store 不必要的 getter
  - 现有第 241 行 `v-if="isServiceProvider"` 逻辑可保持不变，向后兼容
- **后续行为**：
  - 商家进入商品编辑页时，销量输入框正常显示
  - 用户可修改销量值为任意非负整数（`el-input-number` 已设 `:min="0"`）
  - 保存时通过现有 `updateMerchantProduct` 提交，后端 `UpdateProduct` 更新 `sales` 字段
  - 返回列表页（`MerchantProductsTab.vue` 第 201-205 行）可看到销量已更新

## Assumptions & Decisions

- **决策**：采用 `const isServiceProvider = true` 而非在 auth store 增加 getter，因为 web-admin 内不存在角色切换，运行时判断无意义。
- **决策**：不在列表页加行内修改入口（已与用户确认采用"修复编辑页 bug"方案）。
- **决策**：不修改后端、API、类型定义，因为它们均已正确实现 sales 字段支持。
- **假设**：销量修改无需历史记录（与上一版计划一致）。
- **假设**：销量修改后立即生效，无需审核流程。

## Verification Steps

1. **静态检查**
   - 在 `web-admin` 目录运行 `npm run build`，确认 TypeScript 编译无 `isServiceProvider` 未定义错误。
   - 确认 `MerchantProductEditView.vue` 第 241 行 `v-if="isServiceProvider"` 不再报 lint 警告。

2. **运行时验证**
   - 商家登录 web-admin，进入某个商家详情页 → 商品 Tab → 点击某商品"编辑"。
   - 确认商品编辑页"商品基础信息"卡片中"排序值"下方出现"销量"输入框。
   - 修改销量值（例如从 0 改为 100），点击"保存商品"。
   - 返回商品列表，确认该商品"销量"列显示为 100。
   - 重新进入该商品编辑页，确认销量输入框回显为 100。

3. **回归验证**
   - 新建商品时不填销量（默认 0），保存后列表显示销量 0。
   - 修改商品其他字段（如名称、库存）但不改销量，保存后销量保持不变。
   - 销量输入框 `:min="0"` 限制生效，无法输入负数。

## Implementation Steps

1. **修改 `web-admin/src/views/merchant/MerchantProductEditView.vue`**
   - 在 `<script setup>` 第 32 行（`isEditMode` 计算属性）之后插入：
     ```ts
     // web-admin 为商家后台，登录者必为商家
     const isServiceProvider = true
     ```

2. **构建验证**
   - 在 `web-admin` 目录运行 `npm run build` 确认通过。
   - 启动 dev server，按 Verification Steps 中的运行时验证步骤手动测试。

3. **更新原有计划文档（可选）**
   - 原 `.trae/documents/web-admin-modify-product-sales-plan.md` 描述的方案已基本落地，本次只是修复前端 bug，可保留不动或追加一行说明。
