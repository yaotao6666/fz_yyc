# 服务人员小程序端功能整理与收敛 — 计划

## Summary
以 **PC 后台(web-admin)现有业务板块能力为准**，全端对照梳理服务人员小程序端（staff-miniprogram）功能：产出《服务人员小程序端功能整理与收敛规划》文档，梳理「服务人员注册→审核→能做什么」的完整链路，标注 PC↔小程序能力对应关系，并**以 PC 为准删除/收敛小程序中与 PC 不对应或多余的功能**。本次独立交付为功能梳理文档；文档中明确的"建议删除/收敛项"，在用户确认后再落地为代码调整。

## Current State Analysis（已探索）

### 服务人员小程序端现状（staff-miniprogram）
页面（pages.json）：登录、待办(tab)、工单(tab)+详情、健康(居民档案/评估/适配建议/照护计划+详情/照护记录/随访+详情)、排班(tab)、我的(tab)+资料变更+我的审核记录。

后端 API 能力（server/cmd/server/main.go 327-361）：
- 认证：register(带资质)/login/wechat-login
- 工单：todo / orders pending+accepted / orders:id / accept / check-in / check-out / statistics
- 资料：profile GET+PUT / audits
- 健康：assessment-forms / residents health-record+assessments / fitting-recommendations / care-plans+detail / care-visits / follow-up-tasks(+complete/skip) / health-education / residents monitoring

### PC 后台（web-admin）现有板块
工作台(dashboard)、订单管理(+订单详情+租赁到期提醒 rental-due)、商品、分类、服务人员(列表+审核)、商户资料、经营分析、系统管理(RBAC: 菜单/角色/部门/员工)、健康服务(档案/量表/评估/适配/照护/随访/体征/宣教)、结算分账。

### PC↔小程序对齐关系
- PC 有而小程序**天然不应有**（角色隔离）：数据分析、系统管理(RBAC)、结算分账、商户资料、商品/分类管理 → 这些不出现在小程序端，正确。
- PC 有而小程序**对应承接**（需保留）：工单执行(签到/签退)、客户健康档案查看/评估/适配/照护/随访/体征/宣教、服务人员资料与审核记录、接单统计。
- PC 有而小程序**应有入口但缺**：PC「订单派单」指派后，订单进小程序"已接/待办"（已支持）；PC「租赁到期提醒」为后台运营，服务人员端无对应——**小程序"验机归还"菜单与 PC 归还操作(merchant 端)不对应，属多余**。
- 小程序有而 PC **无对应来源**（多余/占位）：**排班**(tab)、**消息通知/设置/帮助反馈**(我的页占位)。

## Proposed Changes

### 交付物：新增功能梳理文档
路径：`docs/prd/staff-miniprogram-功能整理与收敛规划.md`（新建）

文档结构与结论：

**A. 服务人员角色链路图（概述）**
1. 注册申请(带资质) → 2. PC 审核通过 → 3. 登录(账号/微信) → 4. 查看待办/抢单/被派单 → 5. 签到服务 → 6. 签退 → 7. 复用健康服务(档案/评估/适配/照护/随访/体征/宣教) → 8. 接单统计与资料变更(待审核)。

**B. 全端功能对齐对照表**（板块 | PC 端 | 服务人员小程序端 | 处置）
- 工单执行 | 订单派单/核销/归还(merchant) | 待办/已接/签到/签退 | ✅保留
- 健康档案/评估/适配/照护/随访/体征/宣教 | 健康服务全板块 | 全部页面 | ✅保留
- 接单统计 | 经营分析(商家视角) | statistics(个人视角) | ✅保留
- 资料与审核 | 服务人员管理/审核列表 | 资料变更/我的审核记录 | ✅保留
- 排班 | 无排班管理 | 排班(tab) | ⚠️**删除或标注占位**（PC无对应，多余）
- 验机归还 | 租赁到期提醒/归还(PC后台操作) | "验机归还"菜单占位 | ⚠️**删除**（归还属PC后台运营，非服务人员职责）
- 消息通知/设置/帮助 | 无对应 | "我的"占位菜单 | ⚠️**删除或折叠**（无后端来源）

**C. 建议删除/收敛清单**（含对应代码调整，待确认后执行）
| # | 小程序页面/菜单 | 处置 | 落地点 |
|---|---|---|---|
| 1 | `pages/schedule/index`(排班 tab) | 删除 tab + 页面 + staffScheduleApi | pages.json、tabBar、schedule页、api/index.ts |
| 2 | profile「验机归还」菜单项 | 删除 | profile/index.vue menuGroups |
| 3 | profile「接单统计」占位菜单 | 指向已实现 statistics，或删除占位 | profile/index.vue |
| 4 | profile「消息通知/设置/帮助反馈」占位 | 删除或折叠为"开发中"提示 | profile/index.vue |
| 5 | workorder 详情「查看客户健康档案」入口 | 保留并补强（作为下单→客户档案主入口） | workorder/detail.vue（若已存在） |

**D. 保留功能清单**（小程序端应呈现的完整功能树）

**E. 待补充/缺口建议**（可选，非本次必做）
- 服务人员端"客户列表"：当前仅从工单进入客户健康档案，无独立客户列表（PC 可提供服务人员的服务客户台账接口）。标注为"后续增强"。

### 可选的代码落地（仅当用户批准"C"清单后再执行）
- 删除排班 tab：pages.json 移除 `pages/schedule/index` 与 tabBar 三项、删除 schedule 目录、api 移除 staffScheduleApi 及类型。
- 收敛"我的"页占位菜单：profile/index.vue 精简 menuGroups。

## Assumptions & Decisions
1. 本次交付核心为**功能梳理文档**（用户已确认），不直接改代码；文档中"建议删除/收敛项"待用户批准后落地。
2. **以 PC 为准**：小程序端仅保留 PC 后台具有对应治理/职责能力的功能；PC 没有对应运营/管理端的功能（排班、消息通知、验机归还等）视为多余，标记删除或折叠。
3. 删除"验机归还"：归还/退押金为 PC 后台(merchant)职责，服务人员端不留入口。
4. 删除"排班"：PC 无排班管理能力，小程序排班无数据来源，属占位多余。
5. 服务人员"接单统计"保留（有 statistics 接口，个人维度）。
6. 文档命名遵循 `docs/prd/` 下现有 PRD 文档风格。

## Verification
1. 文档在 `docs/prd/` 下生成，结构完整、表格齐全。
2. 对照表能一一对应 PC 后台全部板块与小程序全部页面，无遗漏。
3. 处置结论（保留/删除）与后端接口能力一致（删除项确实无后端 API 支撑或属 PC 职责）。
4. （若执行代码落地）staff-miniprogram `npm run build` 通过，pages.json/tabBar 无残留排班引用。

## File Path Reference
- 新增文档：`docs/prd/staff-miniprogram-功能整理与收敛规划.md`
- 小程序源码参考（落地时）：`staff-miniprogram/src/pages.json`、`pages/schedule/index.vue`、`pages/profile/index.vue`、`src/api/index.ts`、`src/types/index.ts`
- 后端对照参考：`server/cmd/server/main.go`(service-staff 路由)、`server/internal/handlers/merchant/`(order_dispatch/health)、`web-admin/src/router/index.ts`