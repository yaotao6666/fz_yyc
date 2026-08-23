# 为 merchant 账号授予全部菜单与功能权限 Spec

## Why

PC 管理端接入 RBAC 后，merchant 账号登录出现「无权限、看不到菜单」的问题。根因有三：
1. 开发库未执行 RBAC 基础迁移（`20260818000100_rbac.sql`），`sys_menus / sys_roles / sys_role_menus / sys_departments / merchant_staff_roles` 表不存在，菜单/权限查询失败。
2. 运行中的服务端为旧构建（未包含 RBAC 菜单/权限返回逻辑），登录响应 `menus/permissions` 为空。
3. `merchant_staff_roles` 无任何绑定：`role='staff'` 的账号（如 `staff2`）无角色即无任何菜单与功能权限；`role='owner'` 的账号依赖超管直通。

需要先把「超级管理员」的全部菜单与功能权限分配给现有 merchant 账号，恢复登录可见性，并让已登录会话无需重新登录即可恢复菜单。

## What Changes

- **新增幂等迁移 SQL** `server/migrations/20260818000200_grant_admin_role_to_merchant_staff.sql`：
  - 确保「超级管理员」角色（`sys_roles.id=1`）存在且绑定全部 `sys_menus`（`INSERT IGNORE ... SELECT 1, id FROM sys_menus`）。
  - 将**全部现有** `merchant_staffs` 账号绑定到「超级管理员」角色（`INSERT IGNORE ... SELECT id, 1 FROM merchant_staffs`），使所有 merchant 账号获得全部菜单与功能权限。
  - 全程幂等（`INSERT IGNORE` 依赖唯一索引 `uk_sys_role_menus` / `uk_merchant_staff_roles`），可重复执行。
- **开发库补执行基础迁移**：已手动执行 `20260818000100_rbac.sql`（建表 + 种子菜单/角色），新迁移在其之后按序追加。
- **前端会话恢复**：`web-admin/src/router/guards.ts` 在登录态存在但本地 `menus` 与 `permissions` 均为空时（旧会话/旧缓存），自动调用一次 `GET /api/v1/merchant/rbac/permissions` 刷新权限（页面加载内仅尝试一次），使已登录会话无需重登即可恢复菜单。
- **重建并重启服务端**：用最新代码 `go build` 并重启，使 owner 超管直通与 RBAC 菜单/权限返回逻辑生效。

## Impact

- Affected specs: RBAC 菜单/角色/权限加载、PC 后台登录态、动态菜单
- Affected code:
  - `server/migrations/20260818000200_grant_admin_role_to_merchant_staff.sql`（新增）
  - `web-admin/src/router/guards.ts`
  - `web-admin/src/stores/auth.ts`（复用已实现的 `fetchPermissions`）
- Affected data: `merchant_staff_roles`（全部 staff 绑定到角色 1）、`sys_role_menus`（角色 1 绑定全部菜单）

## ADDED Requirements

### Requirement: merchant 账号默认拥有全部菜单与功能权限
系统 SHALL 将全部现有 merchant 账号（含 `role='staff'`）绑定到「超级管理员」角色，使其登录后可见全部菜单并拥有全部功能权限码。

#### Scenario: staff 账号登录可见全部菜单
- **WHEN** 使用任一 staff 账号（如 `staff2`）登录 PC 后台
- **THEN** 登录响应 `menus` 返回完整菜单树，`permissions` 返回全部权限码，侧边栏完整显示

#### Scenario: 迁移幂等可重复执行
- **WHEN** 迁移 SQL 在已执行过的数据库上再次执行
- **THEN** 不报错、不产生重复绑定（依赖 `INSERT IGNORE` + 唯一索引）

### Requirement: 已登录会话自动恢复权限
系统 SHALL 在已登录但本地菜单/权限缓存为空时，自动向后端刷新一次权限，避免旧会话看不到菜单。

#### Scenario: 旧会话刷新后恢复菜单
- **WHEN** 浏览器持有有效 token、本地 `menus/permissions` 为空，进入任意受保护路由
- **THEN** 前端自动调用 `/api/v1/merchant/rbac/permissions`，刷新后侧边栏显示完整菜单，页面级权限校验按新权限放行

### Requirement: owner 超管登录返回完整菜单
系统 SHALL 保证 `role='owner'` 账号（如 `merchant / merchant123`）登录返回全部可见菜单树与权限码。

#### Scenario: owner 登录返回全部菜单
- **WHEN** 使用 `merchant / merchant123` 登录
- **THEN** 登录响应 `menus` 为完整树（含「商品管理」一级目录及其二级菜单、「系统管理」及其四子页），`permissions` 为全部权限码

## MODIFIED Requirements

### Requirement: 登录响应包含菜单与权限
登录接口（`/api/v1/auth/merchant/login`、微信快捷登录）SHALL 在响应中返回当前员工可见 `menus` 与 `permissions`；`GET /api/v1/merchant/rbac/permissions` 用于登录后刷新。现有实现保持，仅通过数据授权与前端恢复逻辑确保返回非空。

## REMOVED Requirements

（无移除项）
