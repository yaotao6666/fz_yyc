# `pages/sp/home` 提示“需要管理员权限”排查与统一鉴权计划

## Summary

- 目标：统一后端服务商鉴权口径，将现有 `admin` 语义收口为 `sp`，修复服务商登录后进入 `pages/sp/home` 时提示“需要管理员权限”的问题。
- 业务口径：
  - 本系统没有平台。
  - 无需 PC 管理后台。
  - 由单服务商承担传统平台管理员职责。
- 成功标准：
  - 使用 `pages/sp/login` 的现有服务商账号登录后，可正常进入 `pages/sp/home`。
  - `GET /api/v1/sp/dashboard` 与其他 `/api/v1/sp/*` 接口全部按 `sp` 身份放行，不再出现“需要管理员权限”。
  - 后端不再保留 `AdminAuth` 这类平台管理员鉴权语义。
  - PRD、测试账号文档、回归脚本中的“admin/管理员后台/PC 管理后台”口径同步改为服务商口径。

## Current State Analysis

- 前端服务商登录页 `miniprogram/src/pages/sp/login.vue` 登录成功后，直接 `reLaunch('/pages/sp/home')`。
- 前端服务商状态仓库 `miniprogram/src/stores/sp.ts` 通过 `spLogin()` 调用 `POST /api/v1/sp/auth/login`，并将返回的 token 存入 `sp_token`。
- 前端 `miniprogram/src/pages/sp/home.vue` 在 `onShow` 中请求 `GET /api/v1/sp/dashboard`。
- 前端请求封装 `miniprogram/src/utils/request.ts` 对 `/api/v1/sp/*` 会自动携带 `sp_token` 到 `Authorization: Bearer ...`。
- 后端服务商登录接口 `server/internal/handlers/sp/handler.go` 中 `Login()` 当前生成 token 的 `user_type` 为 `"sp"`：
  - `utils.GenerateToken(admin.ID, "sp", admin.Username)`
- 后端服务商接口路由 `server/cmd/server/main.go` 中，`/api/v1/sp/*` 保护组当前使用：
  - `middleware.JWTAuth(), middleware.AdminAuth()`
- 后端 `server/internal/middleware/auth.go` 中 `AdminAuth()` 只接受 `user_type == "admin"`，因此服务商 token 调服务商接口必然失败。
- 后端仍保留独立的 `/api/v1/auth/admin/login`、`/api/v1/admin/*`、`internal/handlers/admin/*`，与当前“无平台、仅单服务商”的业务口径不一致。
- 上传接口 `server/internal/handlers/upload/upload.go` 也仍按 `user_type == "admin"` 选择上传目录前缀，属于同一套旧口径残留。
- 测试脚本与文档仍保留旧表述，例如：
  - `server/scripts/regression-test.sh`
  - `server/scripts/regression.sh`
  - `测试账号文档.md`
  - `docs/prd/PRD-测试与附录.md`
  - `PRD.md`
- 结论：当前问题不是单点 bug，而是后端身份模型仍混用“平台管理员(admin)”与“服务商(sp)”两套概念，需整体统一为 `sp`。

## Proposed Changes

### 1. 统一 JWT 身份口径，移除平台管理员鉴权语义

- 文件：`server/internal/middleware/auth.go`
- 变更：
  - 删除 `AdminAuth()` 的使用场景。
  - 保留并使用 `SpAuth()`，判断 `user_type == "sp"`。
  - 将 JWT Claims 注释中的用户类型口径从 `admin, merchant, user` 调整为 `sp, merchant, user`。
- 原因：
  - 用户已明确：系统没有平台管理员角色，只有单服务商角色。
  - 当前 `admin` 仅是历史遗留命名，不应继续作为独立鉴权类型存在。
- 做法：
  - 新增或保留 `SpAuth()` 并作为唯一服务商鉴权中间件。
  - 删除 `AdminAuth()` 定义及其引用，避免后续继续误用。

### 2. 将现有服务商路由全部收口到 `sp` 鉴权

- 文件：`server/cmd/server/main.go`
- 变更：
  - 将 `spGroup.Use(middleware.JWTAuth(), middleware.AdminAuth())` 改为 `middleware.SpAuth()`。
  - 评估并处理 `/api/v1/auth/admin/login` 与 `/api/v1/admin/*`：
    - 当前业务口径下，这两组不再代表独立平台后台。
    - 需要改为服务商语义，或下线路由并把能力并入 `/api/v1/sp/*`。
- 原因：
  - 服务商登录入口、服务商页面、服务商保护接口必须使用一致的 `sp` 身份。
  - 保留 `/api/v1/admin/*` 只会继续强化“平台后台”假象，与当前产品边界冲突。
- 做法：
  - 服务商小程序相关保护接口统一改走 `SpAuth()`。
  - `/api/v1/admin/*` 若仍有最小剩余能力，需要迁移到 `/api/v1/sp/*` 后再删除旧路由。

### 3. 统一登录接口与返回语义为服务商口径

- 文件：
  - `server/internal/handlers/sp/handler.go`
  - `server/internal/handlers/admin/handler.go`
  - `server/cmd/server/main.go`
  - `miniprogram/src/stores/sp.ts`
  - `miniprogram/src/api/index.ts`
- 变更：
  - 保留并继续使用 `POST /api/v1/sp/auth/login` 作为唯一服务商登录入口。
  - 清理或废弃 `POST /api/v1/auth/admin/login` 这类旧平台管理员登录入口。
  - 若后端仍保留 `internal/handlers/admin/*` 的少量能力，需要迁移到 `sp` 处理器或直接删除未使用代码。
- 原因：
  - 前端当前已经按 `sp` 工作，不应再向旧 `admin` 口径回退。
  - 登录入口与 JWT 用户类型要与服务商页面、服务商接口保持同一语义。
- 做法：
  - 前端原则上只需要做最小兼容校验，不新增第二套后台登录。
  - 如果代码里有 admin 类型命名残留，只在必要处同步重命名，避免语义继续混乱。

### 4. 收口上传与依赖 `admin` 的辅助逻辑

- 文件：
  - `server/internal/handlers/upload/upload.go`
  - 其他检索到的 `user_type == "admin"` 判断位置
- 变更：
  - 将上传前缀等依赖 `admin` 身份的逻辑统一改成 `sp`。
- 原因：
  - 如果只改主路由鉴权，辅助逻辑仍按 `admin` 分支处理，后续会继续出现隐性异常。

### 5. 同步更新文档、测试账号与回归脚本

- 文件：
  - `PRD.md`
  - `docs/prd/PRD-功能说明.md`
  - `docs/prd/PRD-接口文档.md`
  - `docs/prd/PRD-测试与附录.md`
  - `测试账号文档.md`
  - `server/scripts/regression-test.sh`
  - `server/scripts/regression.sh`
  - 其他直接写死 `/api/v1/auth/admin/login` 或 `admin/admin123` 的测试文件
- 变更：
  - 将“管理员/平台后台/PC 管理后台”统一改成“服务商/服务商后台”。
  - 将测试账号口径说明为“服务商账号 `admin / admin123`，仅作为单服务商管理员账号，不代表平台管理员角色”。
  - 将服务商接口登录入口统一为 `/api/v1/sp/auth/login`。
- 原因：
  - 文档与脚本仍在强化旧的 admin 概念，会导致后续联调、测试和二次开发继续误解系统边界。

## Assumptions & Decisions

- 决策：采用“全面从 `admin` 收口到 `sp`”方案。
- 不保留平台管理员独立身份模型，也不保留“PC 管理后台”产品口径。<mccoremem id="01KRP9JRBERQH84RCRF7GAHPJ3" />
- 不采用“继续保留 `admin`，仅对 `pages/sp/home` 做局部兼容”方案。
- 决策依据：
  - 用户已明确：系统只有单服务商，由该角色承担传统平台职责。<mccoremem id="01KRP9JRBERQH84RCRF7GAHPJ3" />
  - 当前 `admin` 在代码里本质上就是服务商管理员历史命名，不再符合产品定义。
  - 若继续保留 `admin`，即使当前页面修好，后续接口、脚本、文档仍会继续出现概念偏差。
- 假设：
  - 当前 `pages/sp/login` 使用的是服务商管理员账号，对应表 `service_provider_admins`。
  - `server/internal/handlers/sp/*` 中依赖的服务商 ID 解析逻辑是围绕 `service_provider_admins` 工作的，切换身份名为 `sp` 不改变业务数据归属。

## Verification Steps

- 后端接口验证：
  - 使用 `POST /api/v1/sp/auth/login` 登录，确认返回 token 且可访问 `GET /api/v1/sp/dashboard`。
  - 验证商家列表、商家详情、公告、设置等其余 `/api/v1/sp/*` 接口均正常。
  - 验证旧 `/api/v1/auth/admin/login`、`/api/v1/admin/*` 的处理结果符合新方案：要么已删除，要么已迁移且不再作为对外主入口。
- 小程序端验证：
  - 在 `pages/sp/login` 使用现有服务商管理员账号登录。
  - 登录成功后跳转 `pages/sp/home`。
  - 页面正常展示服务商名称、经营总览、快捷入口。
  - 不再弹出“需要管理员权限”或其他 `admin` 相关错误提示。
- 回归验证：
  - 服务商上传功能前缀正常改为服务商语义。
  - 商家后台 `/api/v1/merchant/*` 与 C 端 `/api/v1/store/*`、`/api/v1/user/*` 不受影响。
  - 文档、测试账号与脚本中不再出现“平台管理员/PC 管理后台”为当前系统事实的描述。
