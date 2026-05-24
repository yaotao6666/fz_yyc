# 服务商小程序下线并迁移到 Web 后台计划

## Summary

- 目标：
  - 将小程序中“服务商登录后”的全部功能下线。
  - 在当前仓库内新增一个基于 `Vite + Vue3 + Element Plus` 的 Web 后台项目，用于完整承接现有服务商端能力。
  - 将系统文档口径从“用户 + 商家小程序 + 服务商小程序”切换为“用户 + 商家 + 服务商 PC 后台管理”的演进模式。
- 已确认的产品决策：
  - 首期 Web 后台需要完整承接当前小程序服务商功能，而不是只做骨架。
  - 小程序端服务商入口本轮直接删除，不保留登录页提示或隐藏式保留。
  - 本次不修改任何后端接口，仅复用现有 `/api/v1/sp/*` 接口体系。
  - 当前真正落地的 PC 端仅为“服务商后台”；用户端与商家端仍保持现状，但文档要切到“未来可演进为三角色 PC 后台”的模式。
- 核心实施策略：
  - 在仓库根目录新增 `web-admin/` 作为新的 PC 后台项目目录。
  - `web-admin/` 采用“单项目、可扩展角色”的目录结构，但本次只实现服务商角色与菜单。
  - 小程序端彻底移除 `pages/sp/*` 页面、服务商登录链路、服务商状态仓库及仅供服务商使用的前端 API 封装。

## Current State Analysis

### 1. 当前仓库没有独立 Web 后台项目

- 根目录当前只有：
  - `miniprogram/`：微信小程序前端
  - `server/`：Go 后端
- 未发现已有的 `web/`、`admin-web/`、`pc-admin/` 等独立后台项目目录。
- 这意味着本次需要从 0 到 1 新建一个新的前端项目，而不是在现有项目内补几个页面。

### 2. 当前服务商能力全部仍在小程序端

- 小程序服务商路由仍保留在 `miniprogram/src/pages.json`：
  - `pages/sp/login`
  - `pages/sp/home`
  - `pages/sp/merchants/list`
  - `pages/sp/merchants/detail`
  - `pages/sp/merchants/edit`
  - `pages/sp/settlements/history`
  - `pages/sp/analytics/merchant-stats`
  - `pages/sp/settings`
- 服务商首页 `miniprogram/src/pages/sp/home.vue` 当前仍承担完整的后台导航作用：
  - 经营总览
  - 新增商家
  - 商家列表
  - 分账历史
  - 数据分析
- 服务商页面当前已不再包含公告页，但整体服务商后台仍完整运行在小程序内。

### 3. 小程序前端已形成完整的服务商接口与状态体系

- 服务商登录与业务 API 主要集中在 `miniprogram/src/api/index.ts`：
  - `POST /api/v1/sp/auth/login`
  - `POST /api/v1/sp/auth/logout`
  - `GET /api/v1/sp/settings`
  - `POST /api/v1/sp/account/change-password`
  - 商家创建、编辑、支付配置、列表、详情
  - 分账历史
  - 商家分布、订单分析、金额分析、排行榜
- 服务商状态仓库位于 `miniprogram/src/stores/sp.ts`。
- 请求层 `miniprogram/src/utils/request.ts` 对 `/api/v1/sp/*` 有单独的 token 与登录态处理逻辑。
- 说明本次迁移不仅是页面搬家，还需要把服务商身份链路整体从小程序前端抽离到 Web 项目。

### 4. 后端接口已具备可复用条件

- 服务商接口当前统一挂在 `server/cmd/server/main.go` 的 `/api/v1/sp/*` 下。
- 现有服务商 handler 已覆盖首期后台所需能力：
  - 登录与设置：`server/internal/handlers/sp/handler.go`
  - 商家管理：`server/internal/handlers/sp/merchant_management.go`
  - 支付与分账：`server/internal/handlers/sp/payment.go`
- 用户要求本次不修改后端接口，因此 Web 后台应以“复用现有接口”为第一原则，不新增中转层，不改接口契约。

### 5. PRD 与产品口径仍将服务商后台定义在小程序端

- `docs/prd/PRD-功能说明.md` 当前仍保留：
  - `pages/sp/login`
  - `pages/sp/home`
  - 各服务商页面结构
- `PRD.md` 当前仍描述“服务商登录后进入服务商管理端”，且多处链路仍以小程序页路径为主。
- 这意味着本次除了代码迁移，还需要同步重写：
  - 页面结构
  - 角色入口
  - 测试链路
  - 交付形态说明

## Proposed Changes

### 1. 新建独立 Web 后台项目 `web-admin/`

- 文件/目录：
  - `web-admin/package.json`
  - `web-admin/vite.config.ts`
  - `web-admin/tsconfig*.json`
  - `web-admin/index.html`
  - `web-admin/src/*`
  - `web-admin/.env.development`
  - `web-admin/.env.production`
  - `web-admin/README.md`
- 变更：
  - 在仓库根目录新增 `web-admin/`。
  - 使用 `Vite + Vue3 + TypeScript + Vue Router + Pinia + Element Plus`。
  - 建立标准后台工程基础：
    - 登录页
    - 基础布局（侧边菜单 + 顶栏 + 内容区）
    - 请求封装
    - 角色态存储
    - 路由守卫
    - 公共表格/表单页面基座
- 原因：
  - 当前仓库没有现成 PC 后台项目。
  - 目录名使用 `web-admin/`，便于后续逐步扩展商家/用户后台，而不把项目名锁死在 `sp-admin`。
- 做法：
  - 本次先只实现“服务商角色”菜单与页面。
  - 目录结构按未来多角色扩展设计，例如：
    - `src/layouts/`
    - `src/router/`
    - `src/stores/`
    - `src/api/`
    - `src/modules/sp/`
    - `src/types/`
    - `src/utils/`

### 2. 在 Web 后台完整承接现有服务商功能

- 来源页面：
  - `miniprogram/src/pages/sp/login.vue`
  - `miniprogram/src/pages/sp/home.vue`
  - `miniprogram/src/pages/sp/merchants/list.vue`
  - `miniprogram/src/pages/sp/merchants/detail.vue`
  - `miniprogram/src/pages/sp/merchants/edit.vue`
  - `miniprogram/src/pages/sp/settlements/history.vue`
  - `miniprogram/src/pages/sp/analytics/merchant-stats.vue`
  - `miniprogram/src/pages/sp/settings.vue`
- Web 目标页面建议：
  - `web-admin/src/views/login/LoginView.vue`
  - `web-admin/src/views/dashboard/DashboardView.vue`
  - `web-admin/src/views/merchant/MerchantListView.vue`
  - `web-admin/src/views/merchant/MerchantDetailView.vue`
  - `web-admin/src/views/merchant/MerchantEditView.vue`
  - `web-admin/src/views/settlement/ProfitSharingHistoryView.vue`
  - `web-admin/src/views/analytics/MerchantStatsView.vue`
  - `web-admin/src/views/settings/SpSettingsView.vue`
- 变更：
  - 将现有小程序服务商功能逐页迁移为 PC 后台页面。
  - 表单、列表、详情、统计页按 Web 交互重排，但不改变现有接口字段和业务语义。
- 原因：
  - 用户已明确首期要“完整承接服务商功能”。
- 做法：
  - 优先以“接口一一映射”迁移：
    - 小程序现有 API 封装在 `miniprogram/src/api/index.ts` 可作为 Web API 适配的参考源。
  - PC 后台 UI 保持简单、标准化后台风格，不做复杂视觉设计。

### 3. 将服务商认证与状态管理迁移到 Web 项目

- 参考文件：
  - `miniprogram/src/stores/sp.ts`
  - `miniprogram/src/utils/request.ts`
  - `miniprogram/src/api/index.ts`
- Web 新增文件建议：
  - `web-admin/src/stores/auth.ts`
  - `web-admin/src/api/sp.ts`
  - `web-admin/src/utils/request.ts`
  - `web-admin/src/router/guards.ts`
- 变更：
  - 使用现有 `/api/v1/sp/auth/login`、`/api/v1/sp/auth/logout`。
  - 在 Web 端重新实现：
    - token 持久化
    - 登录态恢复
    - 401 跳转登录
    - 路由守卫
- 原因：
  - 小程序当前对 `/api/v1/sp/*` 的 token 处理是 `uni` 语境，不能直接复用于 Web。
- 做法：
  - Web 端单独实现一套浏览器环境下的认证封装。
  - 不复用小程序 `uni` 相关工具，不与小程序代码混放。

### 4. 小程序端彻底下线服务商入口与服务商页面

- 变更文件：
  - `miniprogram/src/pages.json`
  - `miniprogram/src/pages/sp/login.vue`
  - `miniprogram/src/pages/sp/home.vue`
  - `miniprogram/src/pages/sp/merchants/*`
  - `miniprogram/src/pages/sp/settlements/history.vue`
  - `miniprogram/src/pages/sp/analytics/merchant-stats.vue`
  - `miniprogram/src/pages/sp/settings.vue`
  - `miniprogram/src/stores/sp.ts`
  - `miniprogram/src/api/index.ts`
  - `miniprogram/src/types/index.ts`
  - `miniprogram/src/utils/request.ts`
- 变更：
  - 从小程序路由中移除全部 `pages/sp/*` 页面。
  - 删除全部服务商页面文件。
  - 删除服务商前端状态仓库。
  - 删除仅供小程序服务商页面使用的 API 封装与类型。
  - 清理调试入口 `condition` 中的“服务商登录（后台）”。
- 原因：
  - 用户要求本轮“直接删除入口和页面”。
  - 服务商能力将由新 Web 项目承接，不应在小程序中继续保留双入口。
- 做法：
  - 保留商家端与 C 端相关逻辑不动。
  - 保留后端 `/api/v1/sp/*` 接口不动。
  - 清理小程序中所有跳转到 `pages/sp/*` 的引用与登录态重定向逻辑。

### 5. 调整小程序请求层，移除服务商专属登录态逻辑

- 重点文件：
  - `miniprogram/src/utils/request.ts`
  - `miniprogram/src/api/index.ts`
  - `miniprogram/src/types/index.ts`
- 变更：
  - 删除 `sp_token`、`sp_info`、`sp_id` 的请求处理分支。
  - 删除针对 `/api/v1/sp/*` 的小程序端自动鉴权逻辑。
  - 删除服务商相关 TypeScript 类型与导出。
- 原因：
  - 如果只删页面不删请求层，项目内会残留大量死代码和错误重定向路径。
- 做法：
  - 以“仅保留商家/C 端实际仍在使用的逻辑”为原则做清理。
  - 保证小程序剩余角色仍只包含：
    - 用户/C 端
    - 商家端

### 6. 统一系统模式文档为“PC 后台演进模式”

- 文档文件：
  - `PRD.md`
  - `docs/prd/PRD-功能说明.md`
  - `docs/prd/PRD-接口文档.md`
  - `docs/prd/PRD-测试与附录.md`
  - `docs/prd/完整业务链路目录.md`
  - `docs/prd/完整链路测试报告-20260513.md`
  - `测试账号文档.md`
  - `miniprogram/README.md`
  - `web-admin/README.md`
- 变更：
  - 将“小程序服务商后台”改写为“服务商 Web/PC 后台”。
  - 将整体产品口径调整为：
    - 用户端：当前以 C 端小程序为主
    - 商家端：当前仍以小程序为主
    - 服务商端：迁移为 Web/PC 后台
  - 文档中新增新的后台项目结构、启动方式、构建方式、角色说明。
  - 说明后续可扩展为“用户 + 商家 + 服务商 PC 后台管理”模式，但本次真正交付的是服务商后台。
- 原因：
  - 仅改代码不改文档，会导致系统角色定义与交付事实不一致。

### 7. 为后续多角色 PC 后台预留扩展结构，但不超范围实现

- 目录建议：
  - `web-admin/src/modules/sp/`
  - 预留但不实现：
    - `web-admin/src/modules/merchant/`
    - `web-admin/src/modules/user/`
- 变更：
  - 路由与布局按“未来可扩展多角色后台”思路组织。
  - 本次不同时创建第二、第三个前端项目。
- 原因：
  - 用户已确认当前范围为“仅服务商迁到 PC”，但希望系统模式切到三角色 PC 管理的演进方向。
- 做法：
  - 使用单项目可扩展结构，而不是同时创建三个独立前端工程。
  - 避免本次工程量膨胀到三套后台并行启动。

## Assumptions & Decisions

- 决策 1：新后台项目目录定为 `web-admin/`。
  - 理由：比 `sp-admin/` 更适合未来承接商家/用户后台。
- 决策 2：前端技术栈采用 `Vite + Vue3 + Element Plus + Pinia + Vue Router + TypeScript`。
- 决策 3：本次 Web 后台首期完整承接当前服务商所有小程序功能，而不是只建骨架。
- 决策 4：本次直接删除小程序全部 `pages/sp/*` 页面与入口，不保留迁移提示页。
- 决策 5：本次不修改任何后端接口、数据模型、数据库结构，只做前端迁移与文档改写。
- 决策 6：当前实际落地的 PC 后台只有“服务商后台”。
  - 文档中说明整体架构可演进为“用户 + 商家 + 服务商 PC 后台管理”模式。
  - 不在本次实现用户/商家 PC 前端。
- 决策 7：小程序端最终只保留：
  - 商家端
  - 用户/C 端
- 假设：
  - 当前 `/api/v1/sp/*` 接口字段足够支撑 Web 后台迁移，不需要额外增加分页、筛选或聚合字段。
  - 商家端与 C 端不依赖 `pages/sp/*` 页面存在。
  - 现有品牌配置脚本 `miniprogram/scripts/apply-brand-config.mjs` 与新 Web 项目相互独立，不需要复用。

## Verification Steps

### 1. 新 Web 项目基础验证

- `web-admin/` 可独立安装依赖并启动开发环境。
- 路由、布局、登录页、菜单、请求封装可正常运行。
- Web 项目 README 说明完整，包含启动和构建命令。

### 2. 服务商功能迁移验证

- 使用现有服务商账号可通过 Web 登录。
- 以下页面在 Web 后台可正常访问并完成基础功能：
  - 工作台
  - 商家列表
  - 商家详情
  - 商家新增/编辑
  - 分账历史
  - 数据分析
  - 服务商设置
- 所有页面使用现有 `/api/v1/sp/*` 接口，无后端改动前提下可完成联调。

### 3. 小程序下线验证

- `miniprogram/src/pages.json` 中不再存在任何 `pages/sp/*` 路由。
- 小程序代码中不再有跳转到 `pages/sp/*` 的逻辑。
- 小程序内不再保留服务商登录态处理与服务商页面专属 API 封装。
- `npm run build:mp-weixin` 通过。

### 4. 文档一致性验证

- `PRD.md` 与拆分 PRD 文档对“服务商后台形态”的描述一致。
- `miniprogram/README.md` 不再把服务商作为小程序端能力说明。
- `web-admin/README.md` 能独立指导本地开发与构建。
- 测试账号文档、业务链路文档中，服务商入口已切换为 Web/PC 后台。

### 5. 回归边界验证

- 商家端小程序登录、工作台、订单、商品、设置等原功能不受影响。
- C 端店铺、下单、购物车、我的订单等原功能不受影响。
- 服务商后端接口在不改代码的前提下继续可用，仅调用端从小程序切换为 Web。
