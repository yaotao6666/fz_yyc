# Checklist

- [x] 幂等迁移 SQL 文件存在且可重复执行（重跑无报错、无重复绑定）
- [x] 开发库中全部 20 个 merchant_staffs 账号已绑定「超级管理员」角色（role 1）
- [x] 角色 1 绑定全部 41 个菜单（sys_role_menus）
- [x] owner 账号（merchant/merchant123）登录返回完整菜单树（7 顶级菜单）与全部权限码（39）
- [x] staff 账号（staff2/merchant123）登录返回与 owner 相同的完整菜单树与权限码
- [x] 前端 `guards.ts` 已登录但缓存为空时自动调用 `/api/v1/merchant/rbac/permissions` 刷新权限（页面加载内仅一次）
- [x] 服务端以最新构建重启（fz_yyc_api 容器 healthy）
- [x] 前端类型检查（vue-tsc --noEmit）通过
