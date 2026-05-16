# `admin -> sp` 全量字面替换计划

## Summary

- 目标：按“全部字面替换”处理整个项目中服务商链路里的 `admin`，不仅替换鉴权角色，还替换数据库表名、模型名、接口返回字段、测试账号名/密码、PRD 与脚本文案。
- 直接问题：`pages/sp/home` 请求 `GET /api/v1/sp/dashboard` 报“需要管理员权限”，说明服务商链路曾长期混用 `sp` 与 `admin` 两套语义。
- 本次已确认的用户决策：
  - 数据库与接口层按纯字面替换执行；
  - `service_provider_admins` / `ServiceProviderAdmin` 这类名称也要改；
  - 服务商测试账号和密码也要一起从 `admin / admin123` 改为 `sp / sp123`。
- 本次范围：后端鉴权、数据库 schema 与种子数据、前后端服务商字段、PRD 与测试文档、测试脚本、校验脚本的全量同步。

## Current State Analysis

### 1. 已经切到 `sp` 的部分

- `server/internal/middleware/auth.go`
  - 已存在 `SpAuth()`。
  - `Claims.UserType` 注释已经是 `sp, merchant, user`。

- `server/cmd/server/main.go`
  - 服务商登录入口已是 `POST /api/v1/sp/auth/login`。
  - `/api/v1/sp/*` 已挂 `middleware.JWTAuth(), middleware.SpAuth()`。
  - 当前文件中已无 `/api/v1/auth/admin/login` 与 `/api/v1/admin/*` 路由注册。

- `miniprogram/src/stores/sp.ts`
  - 已使用 `sp_token`、`sp_id`、`sp_info`。
  - 登录接口已是 `/api/v1/sp/auth/login`。

- `server/internal/handlers/upload/upload.go`
  - 上传前缀已从 `uploads/admin` 改为 `uploads/sp`。

### 2. 仍残留 `admin` 的核心位置

#### 后端代码

- `server/internal/handlers/admin/handler.go`
  - 仍存在旧 package，且仍会签发 `GenerateToken(..., "admin", ...)`。

- `server/internal/handlers/sp/handler.go`
  - 仍使用 `models.ServiceProviderAdmin`。
  - 登录与设置接口仍返回 `admin_name`。
  - 局部变量仍大量使用 `admin` 命名。

- `server/internal/handlers/sp/announcement.go`
  - 仍使用 `models.ServiceProviderAdmin` 查询服务商上下文。

- `server/internal/models/models.go`
  - 模型仍为 `ServiceProviderAdmin`。
  - `TableName()` 仍映射到 `service_provider_admins`。

#### 前端类型与页面

- `miniprogram/src/types/index.ts`
  - 登录返回仍定义 `admin_name?: string`。
  - 类型仍定义 `ServiceProviderAdmin`。
  - `SpSettings` 仍定义 `admin_name: string`。

- `miniprogram/src/pages/sp/settings.vue`
  - 界面仍展示“管理员姓名”。
  - 数据字段仍取 `spInfo.admin_name`。

- `miniprogram/src/pages/sp/login.vue`
  - 副标题仍为“服务商管理平台”，与“无平台”新口径冲突。

#### 数据库与初始化数据

- `server/migrations/20240101000000_full_init.sql`
  - 仍创建 `service_provider_admins` 表。
  - 表注释仍为“服务商管理员表”。
  - 唯一键、索引、外键名称仍包含 `service_provider_admins`。
  - 种子数据用户名仍为 `admin`，姓名仍为“超级管理员”，角色值仍为 `admin`。

- 当前迁移目录存在多个 SQL 文件：
  - `20240101000000_full_init.sql`
  - `20260511120000_mock_seed.sql`
  - `20260513120000_service_provider_payment_refactor.sql`
  - `20260515000001_consistency_fix.sql`
  - 说明项目已有增量迁移机制，因此本次不应只改初始化 SQL，还需要补新增迁移处理已有库。

#### 脚本与校验程序

- `server/scripts/regression.sh`
  - 仍查询 `service_provider_admins`。
  - 仍使用 `admin / admin123`。

- `server/scripts/regression-test.sh`
  - 仍查询、清理、重置 `service_provider_admins`。
  - 仍使用 `admin / admin123`。

- `server/cmd/verify_seed/main.go`
  - 仍查询 `service_provider_admins`。
  - 仍验证用户名 `admin` 和密码 `admin123`。

- `test_api.py`
  - 仍使用 `username=admin`。
  - 服务商 token 变量仍是 `admin_token`。
  - 仍请求不存在的旧接口 `/admin/service-provider`。

- `test_api.ps1`、`test_api_en.ps1`
  - 同样仍调用 `/admin/service-provider`。

#### 文档与 PRD

- `PRD.md`
  - 已补充“单服务商模式 / 无 PC 管理后台”。
  - 但仍残留：
    - “服务商管理员登录”
    - `admin_name`
    - `service_provider_admins`
    - 测试账号 `admin / admin123`

- `docs/prd/PRD-功能说明.md`
  - 仍有“管理员账号”等旧表述。

- `docs/prd/PRD-测试与附录.md`
  - 仍记录 `admin / admin123`。

- `测试账号文档.md`
  - 服务商账号章节仍使用 `admin / admin123`。

## Proposed Changes

### 1. 数据库 schema 与现有库迁移

#### 新增迁移文件：`server/migrations/20260516000000_admin_to_sp_full_rename.sql`

- 新增一个增量迁移，专门处理已有数据库。
- 迁移内容按字面替换执行：
  - `RENAME TABLE service_provider_admins TO service_provider_sps`
  - 重建或重命名相关索引、唯一键、外键名称，使名称中不再出现 `admin`
  - 将服务商种子账号数据从 `admin / admin123` 迁移到 `sp / sp123`
  - 将姓名“超级管理员”改为“服务商”
  - 将角色值 `admin` 改为 `sp`
- 若数据库里存在依赖旧表名的约束或视图，迁移中同步修正。

#### 更新初始化脚本：`server/migrations/20240101000000_full_init.sql`

- 把建表语句整体改为：
  - 表名：`service_provider_sps`
  - 注释：服务商表或服务商账号表
  - 索引/约束名同步改为 `service_provider_sps` 口径
- 把初始化数据整体改为：
  - 用户名：`sp`
  - 密码明文口径：`sp123`
  - 姓名：`服务商`
  - 角色：`sp`
- 保持服务商 `service_provider_id` 与业务关联关系不变。

#### 复核其他 SQL 文件

- 逐个检查：
  - `20260511120000_mock_seed.sql`
  - `20260513120000_service_provider_payment_refactor.sql`
  - `20260515000001_consistency_fix.sql`
- 若这些文件中引用了旧表名、旧账号或旧角色值，一并替换。

### 2. 后端模型、handler 与鉴权语义全量替换

#### `server/internal/models/models.go`

- 将 `ServiceProviderAdmin` 改名为 `ServiceProviderSp`。
- 将 `TableName()` 从 `service_provider_admins` 改为 `service_provider_sps`。
- 注释、类型名、关联名同步去掉 `admin`。

#### `server/internal/middleware/auth.go`

- 保持 `SpAuth()` 为唯一服务商权限中间件。
- 彻底确认仓库内不再出现 `AdminAuth`、`user_type == "admin"`、`GenerateToken(..., "admin", ...)`。

#### `server/internal/handlers/sp/handler.go`

- 将服务商模型引用改为新的 `models.ServiceProviderSp`。
- 将服务商返回字段从 `admin_name` 改为 `sp_name`。
- 将局部变量命名从 `admin` 改为 `spUser` 或同类 `sp` 语义命名。
- 保持接口路径仍为 `/api/v1/sp/*`，不改变业务流程。

#### `server/internal/handlers/sp/announcement.go`

- 同步切换到新的服务商模型与表名。

#### `server/internal/handlers/admin/handler.go`

- 删除整个旧 `admin` handler 文件。
- 理由：
  - 当前主路由已不使用；
  - 文件仍签发 `admin` token；
  - 与“全部字面替换”目标冲突。

#### `server/cmd/server/main.go`

- 复核服务商路由只保留 `sp` 语义。
- 确认无任何 `admin` 路由恢复或遗留 import。

### 3. 前端类型、页面字段与文案同步

#### `miniprogram/src/types/index.ts`

- 将 `ServiceProviderAdmin` 改名为 `ServiceProviderSp`。
- 将服务商返回字段从 `admin_name` 改为 `sp_name`。
- 更新 `ServiceProviderLoginResponse`、`SpSettings` 等相关类型。

#### `miniprogram/src/pages/sp/settings.vue`

- 页面标签“管理员姓名”改为“服务商姓名”或“服务商账号姓名”。
- 绑定字段从 `admin_name` 改为 `sp_name`。

#### `miniprogram/src/pages/sp/login.vue`

- 将“服务商管理平台”改为“服务商管理端”。
- 只改文案，不改登录流程。

#### 其他前端引用

- 全局搜索并替换前端中所有服务商上下文里的 `admin_name`、`ServiceProviderAdmin`。
- 商家管理员相关文案不改，因为那是商家侧真实角色，不在本次替换范围。

### 4. 脚本、测试与校验程序同步

#### `server/scripts/regression.sh`

- 所有表查询从 `service_provider_admins` 改为 `service_provider_sps`。
- 所有服务商测试凭证从 `admin / admin123` 改为 `sp / sp123`。
- 所有输出文案改为“服务商”或“服务商账号”，不再出现服务商 `admin` 语义。

#### `server/scripts/regression-test.sh`

- 同步替换：
  - 表名
  - 删除/重置语句
  - 服务商登录凭证
  - 状态说明文案

#### `server/cmd/verify_seed/main.go`

- 查询表改为 `service_provider_sps`。
- 校验账号改为 `sp`。
- 密码校验改为 `sp123`。
- 统计标题改为“服务商账号数量”或同类 `sp` 口径。

#### `test_api.py`

- 服务商登录凭证改为 `sp / sp123`。
- token 变量 `admin_token` 改为 `sp_token`。
- 服务商配置接口改为当前真实存在的 `GET /api/v1/sp/settings`。

#### `test_api.ps1`

- 同步改为 `sp / sp123`。
- 把 `/admin/service-provider` 改为 `/sp/settings`。

#### `test_api_en.ps1`

- 同步改为 `sp / sp123`。
- 把 `/admin/service-provider` 改为 `/sp/settings`。

### 5. PRD 与文档全量同步

#### `PRD.md`

- 全面替换服务商上下文中的 `admin`：
  - “服务商管理员登录” -> “服务商登录”
  - `admin_name` -> `sp_name`
  - `service_provider_admins` -> `service_provider_sps`
  - `admin / admin123` -> `sp / sp123`
- 保留商家侧“管理员账号”描述，因为那是商家业务角色，不属于服务商 `admin -> sp` 替换。
- 统一服务商页面描述为“服务商管理端”，不再出现“平台管理员”“PC 管理后台”。

#### `docs/prd/PRD-功能说明.md`

- 服务商章节去掉 `admin` 表述，统一为 `sp` 口径。
- 商家管理员账号相关内容保留。

#### `docs/prd/PRD-接口文档.md`

- 服务商登录返回字段改成 `sp_name`。
- 服务商接口角色约定明确为 `sp`。
- 清理所有服务商上下文中的 `admin` 文案。

#### `docs/prd/PRD-测试与附录.md`

- 默认账号改为 `sp / sp123`。
- 与根 PRD 保持一致。

#### `测试账号文档.md`

- 服务商账号章节整体改为 `sp / sp123`。
- 角色说明改为服务商账号，不再出现服务商管理员的 `admin` 命名。

#### 其他 Markdown

- 复核并替换：
  - `docs/prd/完整链路测试报告-20260513.md`
  - `docs/prd/完整业务链路目录.md`
- 只替换服务商上下文中的 `admin`，不误伤商家管理员描述。

## Assumptions & Decisions

- 决策 1：服务商相关 `admin` 按纯字面替换执行：
  - `service_provider_admins` -> `service_provider_sps`
  - `ServiceProviderAdmin` -> `ServiceProviderSp`
  - `admin_name` -> `sp_name`
  - `admin` 账号 -> `sp`
  - `admin123` 密码 -> `sp123`
- 决策 2：服务商 JWT 角色统一为 `sp`，不再兼容 `admin`。
- 决策 3：旧 `server/internal/handlers/admin/handler.go` 直接删除，不保留兼容层。
- 决策 4：旧 `/api/v1/auth/admin/login` 与 `/api/v1/admin/*` 不恢复、不兼容。
- 决策 5：服务商现有信息读取接口统一以 `/api/v1/sp/settings` 为准，替代一切旧 `/admin/service-provider` 测试调用。
- 决策 6：本次“全量替换”仅针对服务商上下文，不替换商家侧真实存在的“管理员账号 / 商家管理员”业务表述。
- 决策 7：由于数据库物理表名要改，本次必须同时更新：
  - 增量迁移
  - 初始化 SQL
  - 所有依赖表名的脚本与校验程序

## Verification Steps

### 1. 全仓静态搜索

- 搜索并确认服务商上下文中以下内容已清理：
  - `AdminAuth`
  - `/api/v1/auth/admin/login`
  - `/api/v1/admin/`
  - `service_provider_admins`
  - `ServiceProviderAdmin`
  - `admin_name`
  - `GenerateToken(..., "admin", ...)`
  - `username":"admin"`
  - `admin123`

### 2. 数据库与迁移验证

- 执行新增迁移后确认：
  - 新表 `service_provider_sps` 存在；
  - 旧表 `service_provider_admins` 不再存在；
  - 索引与外键名已同步更新；
  - 种子账号已变为 `sp / sp123`；
  - 角色值已为 `sp`。

### 3. 后端链路验证

- 至少完成编译或测试验证，确保模型重命名和旧 handler 删除后无编译错误。
- 验证以下接口：
  - `POST /api/v1/sp/auth/login`
  - `GET /api/v1/sp/dashboard`
  - `GET /api/v1/sp/settings`
- 使用 `sp / sp123` 登录后访问服务商首页接口，不再返回“需要管理员权限”。

### 4. 前端链路验证

- `pages/sp/login` 使用 `sp / sp123` 登录成功。
- 跳转 `pages/sp/home` 后加载成功。
- `pages/sp/settings` 能正确展示 `sp_name`。
- 前端无 `admin_name` 字段残留导致的数据类型或渲染错误。

### 5. 文档与脚本一致性验证

- `PRD.md` 与拆分文档一致：
  - `docs/prd/PRD-功能说明.md`
  - `docs/prd/PRD-接口文档.md`
  - `docs/prd/PRD-测试与附录.md`
- `测试账号文档.md`、`server/scripts/regression*.sh`、`test_api*`、`server/cmd/verify_seed/main.go` 中的表名、账号名、密码、接口地址一致。
