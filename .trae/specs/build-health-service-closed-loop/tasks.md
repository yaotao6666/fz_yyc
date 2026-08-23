# Tasks

> 说明：每个阶段独立可交付；后端接口、数据库表变更、PRD 文档必须同步（见 .trae/rules/快速迭代.md）。所有新增状态码写入 server/internal/utils/constants.go，与三端保持一致。

## 阶段一：居民健康档案 + 人群健康评估（数据底座）

- [x] Task 1: 数据库迁移 `20260818100000_health_records_assessments.sql`：新增 health_records / health_assessment_forms / health_assessments 三表与索引、状态/等级枚举；RBAC 菜单与权限码种子（health:view/update、assessment:view/create/config）
  - [x] health_records：user_id 唯一索引，JSON 字段（既往史/过敏史/家族史/用药/慢病标签）
  - [x] health_assessment_forms：questions(JSON)、score_rule(JSON)、status
  - [x] health_assessments：form_id、assessor_type、staff_id、answers、total_score、level、conclusion、suggestions
- [x] Task 2: 后端模型与常量：models.go 新增三个模型；constants.go 新增量表维度/档案状态/评估等级
- [x] Task 3: 后端 API（handlers/health 模块，main.go 注册路由）
  - [x] C 端（user auth）：GET/PUT `/api/v1/user/health-record`；GET/POST `/api/v1/user/assessments`（自助评估）
  - [x] 服务人员（service-staff auth）：GET 档案（仅本人服务过的客户）；POST/GET `/api/v1/service-staff/residents/:user_id/assessments`、GET `/api/v1/service-staff/residents/:user_id/health-record`
  - [x] web-admin（merchant auth + RBAC）：GET/PUT `/api/v1/merchant/health-records`；GET/POST/PUT/DELETE `/api/v1/merchant/assessment-forms`；GET `/api/v1/merchant/health-assessments`
- [x] Task 4: web-admin 页面：健康档案管理（列表/详情/编辑/慢病标签）、评估量表管理（含题目与评分规则配置）；新增"健康服务"菜单 + RBAC 权限码 + v-permission
- [x] Task 5: 服务人员小程序：健康评估页（选量表→答题→提交）、工单详情跳转查看客户档案
- [x] Task 6: C 端小程序："我的健康"入口与页面（健康档案查看/编辑、自助评估列表与答题）
- [x] Task 7: PRD 文档同步（PRD.md / PRD-功能说明.md / PRD-接口文档.md）+ 构建验证（go build、web-admin npm run build、两端小程序构建）

## 阶段二：康复辅具适配（评估→销售/租赁打通）

- [x] Task 8: 数据库迁移：新增 fitting_recommendations 表（含 recommended_products JSON、status、order_id）+ RBAC 权限码（fitting:view/update）
- [x] Task 9: 后端 API：适配建议 CRUD（C 端查看/确认；服务人员生成；web-admin 管理）；推荐商品关联现有商品数据
- [x] Task 10: 前端：C 端"我的健康-适配建议"展示与一键跳转商品页；服务人员端评估提交后生成适配建议；web-admin 适配管理页
- [x] Task 11: PRD 文档同步 + 构建验证

## 阶段三：居家康养照护（照护计划 + 上门记录）

- [x] Task 12: 数据库迁移：新增 care_plans / care_visits 两表 + RBAC 权限码（care:view/create）
- [x] Task 13: 后端 API：照护计划 CRUD（web-admin 创建/指派/生成预约服务订单）；上门照护记录创建（服务人员）；计划与记录查询（C 端/服务人员）；签退 CheckOut 联动
- [x] Task 14: 前端：web-admin 照护计划管理页（含执行记录）；服务人员端"我的照护计划"与照护记录录入（护理项/生命体征/照片）；C 端计划与记录查看
- [x] Task 15: PRD 文档同步 + 构建验证

## 阶段四：持续康复随访 + 慢病管理 + 健康宣教

- [x] Task 16: 数据库迁移：新增 follow_up_tasks / health_monitoring / health_education_articles 三表 + RBAC 权限码（followup:view/update、monitor:view/create、education:view/create）
- [x] Task 17: 后端 API：随访任务自动生成（服务完成/租赁归还/评估完成触发）与执行/跳过；监测记录 CRUD；宣教内容 CRUD 与按慢病标签定向列表；`/api/v1/merchant/follow-up-tasks` 等 web-admin 路由
- [x] Task 18: 前端：web-admin 随访任务管理/生命体征监测/健康宣教管理页；服务人员端随访执行（含选派宣教）与监测录入；C 端随访记录与健康宣教列表
- [x] Task 19: PRD 文档同步 + 构建验证（go build、web-admin build、两端小程序构建）

# Task Dependencies

- 阶段一（Task 1-7）顺序执行，是后续各阶段数据底座，必须先完成并验证
- Task 8 依赖 Task 3（适配建议基于健康档案与评估记录）
- Task 12 依赖 Task 3（照护计划关联健康档案）
- Task 16 依赖 Task 3 与现有工单签退逻辑（随访由评估/服务完成触发）
- 各阶段内：后端（迁移+模型+API）与前端页面可并行开发
