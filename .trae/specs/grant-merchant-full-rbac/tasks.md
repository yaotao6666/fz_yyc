# Tasks

- [x] Task 1: 新增幂等迁移 SQL `server/migrations/20260818000200_grant_admin_role_to_merchant_staff.sql`
  - [x] 确保「超级管理员」角色（`sys_roles.id=1`）存在且启用
  - [x] 将角色 1 绑定全部 `sys_menus`（`INSERT IGNORE ... SELECT 1, id FROM sys_menus`）
  - [x] 将全部 `merchant_staffs` 绑定到角色 1（`INSERT IGNORE ... SELECT id, 1 FROM merchant_staffs`）
- [x] Task 2: 在开发库执行迁移
  - [x] 确认 RBAC 基础迁移已执行（`sys_menus/sys_roles/sys_role_menus/sys_departments/merchant_staff_roles` 表存在）
  - [x] 执行授权迁移：41 菜单、角色 1 绑定 41 菜单、20 个 staff 账号全部绑定角色 1
  - [x] 重跑验证幂等：无重复绑定
- [x] Task 3: 前端会话恢复逻辑（`web-admin/src/router/guards.ts`）
  - [x] 已登录但本地 `menus/permissions` 为空时，页面加载内自动调用一次 `fetchPermissions()` 刷新权限
- [x] Task 4: 重建并重启服务端
  - [x] `go build ./...` 本地编译通过
  - [x] `docker compose build api` 重建镜像
  - [x] `docker compose up -d api` 重启容器（healthy）
- [x] Task 5: 验证 merchant 账号登录可见全部菜单与权限
  - [x] owner（merchant/merchant123）：7 顶级菜单、39 权限码
  - [x] staff（staff2/merchant123）：7 顶级菜单、39 权限码
  - [x] 前端 `vue-tsc --noEmit` 类型检查通过

# Task Dependencies

- [Task 2] 依赖 [Task 1]（先建迁移再执行）
- [Task 5] 依赖 [Task 2]、[Task 4]（数据授权 + 服务端逻辑生效）
