# 在 /docs/uml 目录生成业务板块流程图 — 实施计划

## Summary
在 `/docs/uml` 目录下生成一组 **Mermaid 业务流程图**，覆盖系统的四大业务板块：C端商城交易链路、服务人员工单链路、商户后台管理层、健康服务闭环。采用"端内纵向泳道级"粒度，图表以状态节点串联主流程，标注关键状态码、数据表与触发点。

## 用户已确认的决策
- **格式**：Mermaid 流程图（`.md` 文件内嵌 ` ```mermaid ` 代码块），与现有 docs 体系一致，GitHub/Trae/飞书可渲染。
- **范围**：四个板块全部覆盖（C端商城交易、服务人员工单、商户后台管理、健康服务闭环）。
- **粒度**：端内纵向泳道级——每个板块一图，以状态节点串联主流程，标注关键状态码/表/触发点。

## Current State Analysis
- `/docs/uml` 目录目前 **不存在**，需新建。
- 现有文档：`docs/prd/` 下有 `完整业务链路目录.md`（文字版四大链路）、`PRD-功能说明.md`、`PRD-接口文档.md`。
- 后端路由已在 `server/cmd/server/main.go` 完整注册，分组清晰（store / user / merchant / service-staff / health）。
- 已从代码确认的核心状态机（供流程图准确引用）：
  - `orders.status`：1待支付 → 2已支付 → 3已完成 / 5退款中 / 4已取消 → 6已退款
  - `orders.biz_status`（服务单工作流）：0无 → 1待接单 → 2已接单 → 3服务中 → 5已完成 / 6已取消
  - `orders.order_type`：1普通商品 2租赁商品 3即时服务 4预约服务 5上门服务 6到店服务
  - `orders.deposit_status`：0无押金 1已收 2已退 3部分扣除
  - 健康相关状态（`internal/utils/constants.go`）：HealthRecord（0/1/2）、Assessment（自助1/服务2）、Form（草稿0/启用1）、Fitting（0/1/2）、CarePlan（0草稿/1执行中/2暂停/3完成）、FollowUp（不限类型/状态0待执行/1完成/2跳过）、Monitoring 5类型、Education 草稿/发布
  - 随访任务自动生成触发点：服务签退（`workorder.go` CheckOut）、租赁归还、评估完成

## Proposed Changes

### 新建目录
`/Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/docs/uml/`

### 新建 4 个业务板块流程图文件（均用 Mermaid `flowchart TD`+`subgraph` 泳道）

#### 1. `docs/uml/01-商城交易链路.md`
C端用户完整交易主流程，纵向节点串联 + 按"端/域"泳道分组。核心节点：
- 扫码进入 → 自动登录（wechat-login）→ 首页/分类浏览 → 商品详情（选规格/租赁时长）
- 加购物车 → 购物车结算 → 确认订单（满减/押金明细/下单方式）
- 提交订单（CreateOrder 自动算 order_type/biz_status）→ 微信支付（JSAPI）
- 支付回调 → 订单状态 2已支付
- 订单查看 → 申请退款（退款中→已退款）/ 取消 / 租赁归还退押金
- 分泳道：客户端小程序动作 / 服务端处理 / 微信支付
- 标注关键表：orders, order_items, refunds, user_addresses

#### 2. `docs/uml/02-服务人员工单链路.md`
服务人员接单-履约-签退-随访主流程：
- 登录/注册审核 → 待接订单池（status=2已支付 + biz_status=1待接单 + 未指派）
- 接单（原子更新 assigned_staff_id、biz_status=2）
- 签到（biz_status=3服务中）→ 上门服务 → 签退（biz_status=5、actual_ended_at）
- 签退触发：租赁订单→租后回访随访任务；普通服务→康复随访（followup.CreateFollowUpTask）
- 查看已接/历史订单、待办（待出发+服务中）、统计
- 关联"我的照护计划/随访待办/监测录入"
- 标注：orders.biz_status 流转、follow_up_tasks 自动生成（source_type/service_id）

#### 3. `docs/uml/03-商户后台管理层.md`
商户后台经营与管理主流程，按域泳道：
- 管理员登录 → RBAC 鉴权（接口 middleware.RBAC + 菜单/角色/部门/员工管理）
- 商品/分类管理（三级分类、一口价/租赁销售类型）
- 订单管理（核销/退款/租赁归还退押金）+ 经营数据分析
- 服务人员管理（后台添加/注册审核/启停）
- 健康服务管理（档案/量表/评估/适配/照护/随访/监测/宣教）
- 标注：sys_menus/sys_roles/sys_role_menus 权限闭环，RBAC 60s 缓存

#### 4. `docs/uml/04-健康服务闭环.md`
"评估→适配→照护→随访"连续健康服务闭环，体现三端数据流转：
- 建档（health_records，用户/服务人员建档，数据权限隔离）
- 评估（health_assessment_forms 量表 + health_assessments 记录，自助 assessor_type=1 / 服务人员=2）→ 回写档案 assessment_level
- 适配（fitting_recommendations 推荐商品 → C端一键加购/下单，复用下单链路）
- 照护（care_plans 计划 + care_visits 上门照护，可关联预约服务订单进待接单池）
- 随访（follow_up_tasks 自动生成，健康宣教 articles 按慢病标签定向推送）
- 监测（health_monitoring 体征记录，服务人员/用户录入）
- 泳道：C端小程序 / 服务人员端 / web-admin / 服务端

### 可选：新建汇总索引 `docs/uml/README.md`
列出 4 张图清单 + 对应 PRD 文档链接 + 关键状态码速查表。**仅在用户需要时创建**（遵循"不主动建文档"原则，默认跳过，除非用户要求）。

## Assumptions & Decisions
1. 图粒度采用"端内纵向泳道级"，不画跨端时序（用户已选）。
2. 状态码/字段精确引用 `orders`、`biz_status`、健康状态常量，与代码一致，不做虚构。
3. 每图一个 `.md` 文件，内嵌单个 Mermaid `flowchart`；用 `subgraph` 表达泳道分组。
4. 不修改任何业务代码、路由、数据库；纯文档产出。
5. 遵循现有 docs 中文写作风格，注释简洁、可直接被 AI 工具读取使用。

## Verification
1. 检查 4 个 `.md` 文件已创建于 `docs/uml/`。
2. 逐个文件核对 Mermaid 语法（节点 id 唯一、箭头 `-->`、`subgraph` 闭合、状态码注释正确）。
3. 交叉核对：图内 `orders.status`/`biz_status`/`deposit_status`、健康状态码与 `internal/utils/constants.go`、`internal/models/models.go` 一致。
4. 核对触发点：签退→随访、归还→随访、评估→回写档案的描述与 `workorder.go`、`services/followup/auto.go` 一致。
5. 确认不产生任何代码/迁移/路由改动（`git status` 仅见 docs/uml 新增）。