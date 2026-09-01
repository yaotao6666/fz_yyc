# PC 端健康宣教分类管理缺失：排查与修复计划

## 1. 结论摘要（Root Cause）

「PC 端健康宣教分类没有地方维护」的根因是：**后端能力、接口、权限码、菜单节点、前端类型全部已就绪，唯独 web-admin 前端缺少「宣教分类管理」的**路由**与**视图组件**，且迁移 012 写入的菜单节点 `menu_type` 取值错误，导致该菜单即使有路由也无法在侧边栏渲染。**

证据链：
- 后端 `education_category.go` 已实现两级分类完整 CRUD（list/create/update/delete，删除带子分类/文章守卫）。
- `main.go` L350-353 已注册 `/api/v1/merchant/education-categories`，RBAC 权限码 `education-categories:view/create/update/delete`。
- web-admin `api/sp.ts` L619-640 已封装 4 个分类接口；`types/sp.ts` L671-679 已定义 `HealthEducationCategory`。
- 迁移 012 已插入「宣教分类管理」菜单及按钮权限，并授权超管角色。

## 2. 现状分析（Current State）

### 后端：已具备，无需改动
- **模型** [models.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/models/models.go)：`HealthEducationCategory`（parent_id / name / sort / status）。
- **处理** [education_category.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/health/education_category.go)：两级树；删除若存在子分类或关联文章则拒绝。
- **路由** [main.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/cmd/server/main.go#L350-L353)：`merchantOnlyGroup` 下 CRUD，已挂 RBAC。

### 前端 API/类型：已具备，无需改动
- [sp.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/api/sp.ts#L619-L640)：`getHealthEducationCategories` / `createHealthEducationCategory` / `updateHealthEducationCategory` / `deleteHealthEducationCategory`。
- [types/sp.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/types/sp.ts#L671-L679)：`HealthEducationCategory`。

### 前端路由/视图：缺失 ❌
- [router/index.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/router/index.ts#L196-L201)：只有 `/health/education`（文章），**没有 `/health/education-categories`**，也无 `HealthEducationCategoriesView.vue`。
- 文章页 [EducationArticlesView.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/views/health/EducationArticlesView.vue)：只能读取分类作勾勒选择/筛选，**无分类增删改启停入口**。

### 数据库菜单节点：menu_type 错误 ❌
- 迁移 [012](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/migrations/012_education_categories.sql#L95-L103) L95 将「宣教分类管理」写入 `menu_type=2`；按钮写入 `menu_type=3`。
- sys_menus schema 规定：**1=菜单/目录、2=按钮**（见 [002_seed.sql](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/migrations/002_seed.sql#L214) schema 注释 + L236 数据示例，如 `81/8/1,...menu_type=1,...'health:view'` 与 `811/81,2,...'health:update'`）。
- [AppLayout.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/layouts/AppLayout.vue#L22-L36) `buildMenuNodes` 只渲染 `node.menu_type === 1` 的节点 → 菜单被过滤，**侧边栏永远不显示。**

## 3. 修改方案（Proposed Changes）

### A. 数据库：新增迁移 `013_fix_education_category_menu.sql`
**不重跑 012**（其含 `DROP TABLE health_education_categories` 属破坏性操作）。用幂等 UPDATE 修正已写入节点：
- 「宣教分类管理」菜单：`menu_type 2 → 1`，使其可被侧边栏渲染。
- 三个分类按钮 + 文章 `education:update/delete` 按钮：`menu_type 3 → 2`，与 schema 统一（一致性；不影响权限判定，均不渲染）。
- 按 `permission` 定位，不依赖自增 id，天然幂等。

SQL 文件将置于 `server/migrations/`，执行命令写入文件头注释。

### B. web-admin 路由：新增 `/health/education-categories`
[router/index.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/router/index.ts#L196-L201) 在 `/health/education` 之后追加：

```ts
{
  path: '/health/education-categories',
  name: 'health-education-categories',
  component: () => import('@/views/health/HealthEducationCategoriesView.vue'),
  meta: { title: '宣教分类管理', requiresAuth: true, permission: 'education-categories:view' }
}
```

### C. web-admin 视图：新建 `HealthEducationCategoriesView.vue`
样式/交互沿用现有 Element Plus + `v-permission` 约定（参考文章页）。
功能：
- **列表**：两级分类，一级为父行、二级为子行（可用简单平铺表格 + 「所属父分类」列，或 el-table 树形）。展示 名称/父分类/排序/状态/操作。
- **新建**：对话框选父分类（级联/下拉，`parent_id=0` 表示一级），填 名称 + 排序 + 状态。
- **编辑**：同样对话框回填，可调整父分类（后端校验两级与自指）。
- **启停**：切换 `status`（1 启用 / 0 停用）。
- **删除**：`deleteHealthEducationCategory`，后端守卫拒绝“有子分类 / 有关联文章”；前端二次确认。
- **按钮权限码**：新建 `education-categories:create`、编辑 `education-categories:update`、删除 `education-categories:delete`、页面 `education-categories:view` —— 与后端 RBAC 及菜单节点**三位一体一致**。

## 4. 假设与决策（Assumptions & Decisions）
- 迁移 012 已被执行并写入上述菜单节点；假设当前 DB 中存在这些行。若不执行过，013 的 UPDATE 影响 0 行亦无害（幂等）。
- 分类排序仅通过编辑对话框维护 `sort`（后端无独立排序接口，与 product categories 不同；为最小改动不新增排序接口）。
- 不在本计划改动文章页按钮权限码（`education:create` vs `education:update/delete`），属既有一致性遗留，非本次范围。

## 5. 验证步骤（Verification）
1. 执行迁移：`mysql -h <host> -u <user> -p <db> < server/migrations/013_fix_education_category_menu.sql`，确认 UPDATE 命中菜单行。
2. `go build ./...`（server）：应通过（无后端代码改动，作回归）。
3. `npm run build` 或 `vue-tsc`（web-admin）：新增视图/路由类型检查通过。
4. 登录 web-admin（超管）：健康服务目录下出现「宣教分类管理」菜单；可新建一级/二级分类、编辑、启停、删除（含守卫提示）。
5. 返回「健康宣教管理」文章页：新增分类出现在分类级联选项中。