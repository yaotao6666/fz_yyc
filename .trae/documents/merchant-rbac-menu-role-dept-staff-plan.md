# 阶段零：去除全链路 merchant_id 深度重构 + 阶段一：PC 管理端通用 RBAC 实施计划

## 一、摘要

分两阶段实施：

- **阶段零（前置深度重构）**：系统目前实际为单商户（merchants.id=1），但全链路仍按 `merchant_id` 做关联/过滤。本次将**所有表的 `merchant_id` 关联字段全部去掉**（含接口路径、请求参数、响应字段、过滤条件），**merchants 表保留为全局单例配置表**（仅 id=1 一行，不再被其他表外键引用），支付配置/联系方式/公告/营业状态等商家资料继续可用。范围覆盖**后端、C端小程序、PC 管理端、员工小程序**全链路。
- **阶段一（通用 RBAC）**：在无 merchant_id 的基础上，为 PC 管理端新增菜单/角色/部门/员工管理 + 通用 RBAC（多角色、owner 超管直通、三层精准控制、动态加载菜单、「商品管理」一级目录下沉二级菜单）。

## 二、现状分析（merchant_id 影响面）

- **后端 Go 源码**：约 730 处，涉及 models.go（27）、user/handler.go（37）、merchant/order.go（54）、merchant/product.go（35）、merchant/analytics.go（12）、merchant/staff.go（14）、middleware/auth.go、utils/merchant_qrcode.go、services/orderquery、ws/*（32）、service_staff/*、callbacks/wechatpay 等 31 个文件（含历史迁移，历史迁移**不改写**，仅新增迁移）。
- **C端小程序**：约 183 处 / 15 文件，核心为 `/api/v1/store/:merchant_id/*` 路由、[storeEntry.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/utils/storeEntry.ts)（scene 解析 merchant_id）、[storeCache.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/miniprogram/src/stores/storeCache.ts)、cart/analytics/各 store 页面。
- **PC 管理端**：5 处（types/sp.ts 类型字段）。
- **员工小程序**：1 处（types/index.ts）。

**关键约定（沿用）**：
- 保留 merchants.id=1、sub_mch_id=1112979963 数据（硬约束）。
- 后端业务侧用常量 `DefaultMerchantID = 1` 读取单例配置。
- 迁移遵循项目约束（纯 SQL、`SET @var + PREPARE/EXECUTE`、无循环、先删外键再删列）。

---

## 阶段零：去除 merchant_id 深度重构

### 0.1 数据库变更（新增 `server/migrations/20260818000000_remove_merchant_id.sql`）

1. `merchants` 表：保留（单例配置），删除其被其他表外键引用的约束不影响本身。
2. 依次删除以下表的 `merchant_id` 列（先删指向 merchants 的外键，再删索引，再删列；每步用 information_schema 探测 + PREPARE/EXECUTE 幂等处理）：
   - `categories`、`products`、`product_specs`（无此列则跳过）、`orders`、`order_items`、`user_visits`、`user_behavior_events`、`merchant_delivery_settings`、`merchant_fees`、`merchant_rates`、`merchant_staffs`、`service_staffs`
   - 若表上存在 `fk_xxx_merchant` 外键或 `idx_xxx_merchant_id` 索引，先删。
3. 同步更新参考快照目录（`ddl/fzyyc-cxsm/*.sql`、`data/*.sql`）中对应建表/种子语句，去掉 merchant_id 列与 INSERT 中的 merchant_id 值（**历史 migration 不改写**）。
4. 业务数据本就仅 merchant_id=1，删除列无数据丢失。

### 0.2 后端变更（server）

1. **常量**：`internal/utils`（或 `internal/config`）新增 `const DefaultMerchantID uint64 = 1`。
2. **middleware/auth.go**：
   - JWT `Claims` 增加 `StaffID`；`GetMerchantID(c)` 改为返回 `DefaultMerchantID`（调用点零改动）；`MerchantAuth` 仍校验 `user_type == "merchant"`。
   - `GenerateToken` 增加 staffID 参数（见阶段一，此处一并落地）。
3. **models.go**：删除 `Category / Product / Order / OrderItem / UserVisit / UserBehaviorEvent / MerchantDeliverySettings / MerchantFee / MerchantRate / MerchantStaff / ServiceStaff` 的 `MerchantID` 字段及对应 `Merchant *Merchant` 关联；`Merchant` 模型保留（单例配置，去掉多余外键关联）。
4. **merchant 处理器**：所有 `middleware.GetMerchantID(c)` 改为 `utils.DefaultMerchantID`；Create/Update 不再写 merchant_id；`GetStaffList`/`GetServiceStaffList` 等去掉 `merchant_id = ?` 过滤；`context.go` 的 `resolveTargetMerchantID` 简化。
5. **user 处理器（C端）**：store 路由去掉 `:merchant_id`；`GetStoreHome/GetProducts/GetProductDetail/GetDeliverySettings/GetCategories` 等按 `DefaultMerchantID` 查询；`CreateOrderRequest` 去掉 `merchant_id`；`GenerateOrderNo(merchantID)` 改为固定常量；订单响应保留 merchant 配置对象（商家展示信息），商品响应不再返回 merchant_id。
6. **service_staff 处理器**：去掉 `merchant_id` 过滤。
7. **ws/**：Hub 从「按 merchantID 分桶」改为单实例广播（去掉 merchantID 维度）；`BroadcastOrderNotify/BroadcastStoreVisitNotify` 签名去 merchantID；`dev.go` 去掉 merchant_id 参数。
8. **orderquery/query.go**：去掉 `MerchantID` 过滤参数。
9. **callbacks/wechatpay.go**：读取商家支付配置由 `order.MerchantID` 改为查 merchants 单行。
10. **utils/merchant_qrcode.go**：`GenerateMerchantStoreQRCode` 去掉 merchantID 参数，scene 不再含 `merchant_id=`（固定 `store=1` 或仅带 source）。
11. **main.go**：`/store/:merchant_id` → `/store`；`/dev/order-notify`、`/dev/store-visit-notify` 去掉 merchant_id 字段。
12. **测试文件**：`pagination_test.go`、`product_test.go` 等同步调整。
13. 清理 `merchant_id` 相关 JSON 标签与响应字段。

### 0.3 C端小程序变更（miniprogram）

1. `api/store.ts`、`api/index.ts`：`/api/v1/store/${merchantId}/...` → `/api/v1/store/...`，函数签名去掉 `merchantId`；`getMyOrders` 去掉 `merchant_id` 筛选参数。
2. `utils/storeEntry.ts`：去掉 scene 中 merchant_id 解析，入口固定单店；`parseStoreEntryOptions` 返回去掉 merchantId。
3. `stores/storeCache.ts`、`stores/cart.ts`：去掉 merchantId 状态/参数。
4. `utils/analytics.ts`：去掉 merchant_id 上报参数。
5. `pages/store/*`（home/product/confirm/cart/my-orders/order-detail/order-sound-test/test-entry/address-edit）：去掉 merchantId 传参。
6. `types/index.ts`：`Product` 等去掉 merchant_id 字段（或降级 optional 后删除）。

### 0.4 PC 管理端 / 员工小程序变更

- `web-admin/src/types/sp.ts`：去掉 merchant_id 相关字段（5 处）。
- `staff-miniprogram/src/types/index.ts`：去掉 merchant_id（1 处）。

### 0.5 文档与回归脚本

- 更新 `PRD.md` 及相关 PRD 文档：记录去 merchant_id 重构（数据模型、接口路径变更）。
- 同步更新 `test_*.ps1`、`test_api.py` 等联调脚本中的 merchant_id 参数。

---

## 阶段一：PC 管理端通用 RBAC（无 merchant_id 版）

> 基于阶段零结果：角色/部门/员工-角色关联表**不含 merchant_id**（全局），员工管理不再按商家过滤。其余设计同「四~九」但不引入 merchant_id。

### 1.1 数据模型（新增 `server/migrations/20260818000100_rbac.sql`）

| 表 | 说明 | 关键字段 |
|---|---|---|
| `sys_menus` | 菜单/按钮（全局） | `id, parent_id(0), menu_type(1=菜单 2=按钮), name, path, icon, sort, status, visible, permission, created_at, updated_at` |
| `sys_roles` | 角色（全局，无 merchant_id） | `id, name, code(唯一), remark, status, created_at, updated_at` |
| `sys_role_menus` | 角色-菜单关联 | `id, role_id, menu_id, UNIQUE(role_id, menu_id)` |
| `sys_departments` | 部门（全局，树形） | `id, parent_id(0), name, leader, phone, sort, status, created_at, updated_at` |
| `merchant_staff_roles` | 员工-角色关联 | `id, staff_id, role_id, UNIQUE(staff_id, role_id)` |
| `merchant_staffs` | ALTER 新增 | `ADD COLUMN department_id BIGINT UNSIGNED DEFAULT NULL` |

种子：`sys_menus` 按 1.4 权限清单初始化；`sys_roles` 预置「超级管理员」（code=admin，绑定全部菜单）作模板（owner 直通不强制绑定）。

### 1.2 JWT / 员工身份（阶段零已落地基础）

- JWT `Claims` 含 `staff_id`；merchant 登录 token 的 `user_id = staff_id`（取代原 merchant_id）。
- `Login`/`WechatQuickLogin` 响应追加 `menus`（当前员工可见菜单树）与 `permissions`（权限码数组）。
- 新增 `GET /api/v1/merchant/rbac/permissions` 供登录后刷新。

### 1.3 后端

1. **middleware/rbac.go**：`RBAC(permission)` 中间件 —— 解析当前员工（staff_id），`role == "owner"` 直通，否则加载其角色→菜单权限码集合判断；带 60s 内存缓存，角色/菜单/员工变更时清缓存。
2. **models.go**：新增 `SysMenu / SysRole / SysRoleMenu / SysDepartment / MerchantStaffRole`；`MerchantStaff` 增加 `DepartmentID`、`Roles []SysRole`、`Department *SysDepartment`。
3. **handlers/rbac/**：`menu.go`（树 CRUD）、`role.go`（CRUD + 分配菜单）、`department.go`（树 CRUD）、`permission.go`（我的菜单树 + 权限码）。
4. **merchant/staff.go 扩展**：`Create/UpdateStaff` 支持 `department_id`、`role_ids`（事务维护 `merchant_staff_roles`）；列表返回 roles/department；保留「至少一个 owner」校验。
5. **main.go**：注册 `/merchant/rbac/*` 路由；按 1.4 为既有 `/merchant/*` 路由逐个挂 `middleware.RBAC(...)`；注意 main.go 现有缩进错乱（L224/L248），修改保持可编译。

### 1.4 权限标识清单（种子菜单 + 接口映射）

```
工作台        /dashboard           dashboard:view
订单管理      /orders              orders:view
  核销/完成     按钮                  orders:complete
  退款          按钮                  orders:refund
  归还押金      按钮                  orders:return
商品管理      目录(/products)       （一级目录，无页面；有任一可见子菜单时显示）
  商品管理      /products            products:view        （真实商品管理页）
    新增/编辑     按钮                  products:create
    删除          按钮                  products:delete
    上/下架/批量  按钮                  products:status
    库存          按钮                  products:stock
    规格          按钮                  products:specs
  分类管理      /categories          categories:view
    新增/编辑     按钮                  categories:create
    删除          按钮                  categories:delete
    排序          按钮                  categories:sort
服务人员      /staff               staff:view
  审核/启停     按钮                  staff:update
  重置密码      按钮                  staff:reset-password
  删除          按钮                  staff:delete
数据分析      /analytics           analytics:view
商家资料      /profile             profile:view
  保存资料      按钮                  profile:update
  支付配置      按钮                  profile:payment
  修改密码      按钮                  profile:password
系统管理      目录(/system)         （目录，无页面）
  菜单管理      /system/menus       system:menu:view（+ create/delete 按钮）
  角色管理      /system/roles       system:role:view（+ create/delete/assign 按钮）
  部门管理      /system/departments system:dept:view（+ create/delete 按钮）
  员工管理      /system/staff       system:staff:view（+ create/update/delete/reset-password 按钮）
```

**既有接口映射**：profile/settings/qrcode(GET)→`profile:view`；profile/settings/status/wechat-bind→`profile:update`；change-password→`profile:password`；payment-config→`profile:payment`；orders GET/statistics→`orders:view`，complete/quick-complete→`orders:complete`，refund→`orders:refund`，return→`orders:return`；categories/products 按动作分别映射；service-staff→`staff:*`；merchant/staff→`system:staff:*`；announcements→`dashboard:view`。

### 1.5 前端（web-admin）

1. `types/sp.ts`：新增 `SysMenu/SysRole/SysDepartment/MenuNode/RbacPermissions` 及员工扩展类型；**基于阶段零删除 merchant_id 字段**。
2. `api/sp.ts`：新增菜单/角色/部门/员工 CRUD、`getRbacPermissions`、`assignRoleMenus/getRoleMenus`。
3. `stores/auth.ts`：state 增加 `menus`、`permissions`（持久化）；getter `hasPermission(code)`。
4. `router/index.ts`：新增 `/system/menus|roles|departments|staff`；所有受保护路由加 `meta.permission`（`/products`→`products:view`、`/categories`→`categories:view`，路由路径不变仅侧边层级调整）。
5. `router/guards.ts`：`beforeEach` 校验 `meta.permission`，无权限重定向 `/dashboard`。
6. `layouts/AppLayout.vue`：侧边菜单**按权限过滤后的菜单树动态加载**（递归渲染 `el-sub-menu`/`el-menu-item`，支持一级目录+二级子菜单；目录在存在任一可见子菜单时显示；`activeMenu` 按路由前缀匹配子页）。移除静态 `menuItems`。
7. 新增页面（`views/system/`）：`MenuManagementView`（树表格 CRUD）、`RoleManagementView`（列表 + 分配权限树多选回显）、`DepartmentManagementView`（树 CRUD）、`StaffManagementView`（员工列表 + 角色多选/部门下拉/重置密码/启停/删除）。
8. **按钮级精准控制**：`utils/permission.ts` + `main.ts` 注册 `v-permission` 指令；三层控制：接口级（后端 RBAC）→ 页面级（路由守卫）→ 按钮级（指令）。

### 1.6 假设与约定

- owner（role=owner）超管直通 RBAC；非 owner 员工无角色时默认无后台管理权限。
- 侧边菜单完全由后端权限过滤后的菜单树动态加载；目录无独立权限码，存在可见子菜单时展示。
- 不含数据权限；前端路由保持静态定义（`meta.permission`）。

---

## 七、验证步骤

1. 阶段零：后端 `cd server && go build ./...`；前端 `npm run build`（miniprogram `npm run build:mp-weixin`、web-admin、staff-miniprogram 均通过）；SQL 迁移在开发库可幂等执行；owner 登录全功能回归（下单→支付→工单→退款）、C端 store 首页/商品/购物车/下单/地址回归、WS 新订单提醒回归。
2. 阶段一：RBAC 构建通过；owner 登录见「商品管理」一级目录（含商品管理/分类管理二级）与「系统管理」四子页；新增角色/部门/员工并授权；新员工登录仅见被授权菜单（动态加载），直接访问无权限路由被拦截，直接调无权限接口返回 403；权限变更重登即时生效。
3. 同步更新 PRD.md。

## 八、实施步骤清单（含执行进度）

> 状态：✅ 已完成｜🟡 进行中｜⬜ 未开始（截至 2026-08-18）

**阶段零（merchant_id 深度重构）**
1. ✅ 编写迁移 `20260818000000_remove_merchant_id.sql` + 同步 ddl/data 快照。
2. ✅ 后端：常量 `DefaultMerchantID` / 中间件 / models 去 merchant_id。
3. ✅ 后端：merchant/user/service_staff/ws/orderquery/callbacks/qrcode/main 处理器与路由改造（`/store/:merchant_id`→`/store`）。
4. ✅ 后端：测试文件与清理；`go build ./...` 通过。
5. ✅ C端小程序 15 个文件去 merchant_id；`npm run build:mp-weixin` 通过。
6. ✅ web-admin / staff-miniprogram 类型清理；构建通过。
7. ✅ 更新 PRD.md / docs/prd/* 与回归脚本（test_api.py、test_api.ps1、test_api_en.ps1、test1.ps1）。

**阶段一（PC 管理端通用 RBAC）**
8. ✅ 迁移 `20260818000100_rbac.sql`：已建 `sys_menus / sys_roles / sys_role_menus / sys_departments / merchant_staff_roles` 表 + `merchant_staffs.department_id` 列 + 种子菜单（含「商品管理」一级目录→商品/分类二级）与超管角色。
9. ✅ 后端：models 新增 RBAC 模型；JWT `Claims.staff_id`；`middleware/rbac.go`（`RBAC`/`RBACAny` + 60s 缓存 + 变更清缓存）；`handlers/rbac/{menu,role,department,permission}.go`；`merchant/staff.go` 支持 department_id/role_ids（事务维护关联）；`main.go` 全部 `/merchant/*` 路由挂权限中间件 + `/rbac/*` 注册；Login/WechatQuickLogin 响应含 menus/permissions。
10. ✅ 前端（web-admin）：`types/sp.ts` 新增 RBAC 类型、`api/sp.ts` 封装、`stores/auth.ts` menus/permissions + `hasPermission`、`router` meta.permission + 守卫、`AppLayout.vue` 动态菜单（商品管理一级目录 + 系统管理四子页）、`v-permission` 指令、`views/system/{Menu,Role,Department,Staff}ManagementView`。
11. ✅ 构建验证（web-admin `npm run build` 通过、server `go build ./...` 通过）+ 回归脚本 + PRD.md / PRD-功能说明 / PRD-接口文档 更新。
