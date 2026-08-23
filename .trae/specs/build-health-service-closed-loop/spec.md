# 基层健康服务闭环定制开发 Spec

## Why

当前系统是"商贸只卖器械"模式：用户购买/租赁辅具后服务即终止，缺少健康评估、适配、照护、随访的连续健康服务能力，无法承接老年慢病管理与居家长期照护需求。本 Spec 依托健康服务与管理核心专业能力（人群健康评估、老年慢病管理、康复辅具适配、居家长期照护、健康宣教、居民健康档案管理），把现有商城升级为"健康功能评估→辅具适配租赁销售→居家康养照护→持续康复随访"的基层健康服务闭环，补齐社区居家康复与老年照护资源缺口。

## What Changes

- **新增居民健康档案（数据底座）**：以 C 端用户为对象建立健康档案（基本信息/既往史/过敏史/慢病标签/体检指标），支撑后续评估、适配、照护、随访的关联。
- **新增人群健康评估**：评估量表模板（ADL/Barthel/跌倒风险/营养/认知等）+ 评估记录，支持服务人员上门评估与 C 端自助评估，输出等级结论与建议。
- **新增康复辅具适配**：由评估结论生成适配建议并推荐商品，联动现有购物车/下单链路，打通"评估→适配→销售/租赁"。
- **新增居家康养照护**：照护计划 + 上门照护记录（护理项/生命体征/照片），与现有服务工单（接单/签到/签退）流程打通。
- **新增持续康复随访**：随访任务自动生成（服务完成/租赁归还/评估完成后）与执行记录。
- **新增老年慢病管理与健康宣教**：慢病标签、生命体征监测记录、健康宣教内容与按标签定向推送。
- **三端联动落地**：web-admin（档案/量表/计划/宣教/随访管理与审核）、服务人员小程序（评估/照护/随访/监测执行录入）、C 端小程序（我的健康档案查看/自助评估/适配建议/随访与宣教）。
- **同步更新 PRD 文档**（修改产品功能/数据表必须同步 PRD，见 .trae/rules/快速迭代.md）。
- **BREAKING**：无。所有新增模块与现有订单/工单流程并行，不改变现有路由与数据结构。

## Impact

- Affected specs（能力）：
  - C 端商城下单与订单查询能力（新增"我的健康"入口与适配一键加购）
  - 服务人员工单流程（签退时可关联照护记录；服务/归还完成后自动生成随访任务）
  - web-admin RBAC（新增"健康服务"菜单与权限码，含按钮级 v-permission）
- Affected code：
  - 后端：[models.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/models/models.go)、新增 handlers/health 模块、[main.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/cmd/server/main.go) 路由、middleware/rbac 权限码、utils/constants 状态码
  - 数据库：新增 9 张业务表（health_records / health_assessment_forms / health_assessments / fitting_recommendations / care_plans / care_visits / follow_up_tasks / health_monitoring / health_education_articles），新增 RBAC 菜单/权限种子
  - web-admin：新增"健康服务"菜单与页面、api/sp.ts、types/sp.ts、router/guards
  - staff-miniprogram：新增健康档案/评估/照护/随访/监测录入页面、api/index.ts、types/index.ts
  - miniprogram：新增"我的健康"页面、api、types
  - 文档：PRD.md、docs/prd/PRD-功能说明.md、docs/prd/PRD-接口文档.md、测试账号文档.md

---

## ADDED Requirements

### Requirement: 居民健康档案（Phase 1）
系统 SHALL 以 C 端用户为单位建立居民健康档案，作为评估/适配/照护/随访的数据底座。

- 表 `health_records`：user_id(唯一)、real_name、gender(1男2女)、birth_date、id_card、phone、emergency_contact、emergency_phone、address、height_cm、weight_kg、blood_type、past_history(JSON)、allergy_history(JSON)、family_history(JSON)、surgery_history(JSON)、medication_list(JSON)、chronic_tags(JSON 慢病标签数组)、smoking、drinking、assessment_level、remark、status(0未建档 1正常 2已归档)。
- 每个用户至多一条档案；首次由用户或服务人员建档。

#### Scenario: 用户建档并查看
- **WHEN** C 端用户进入"我的健康"并提交健康档案
- **THEN** 系统创建/更新该用户唯一档案，用户可查看本人档案，其他人不可见

#### Scenario: 服务人员查看关联客户档案
- **WHEN** 服务人员打开自己接单/服务过的订单对应客户的档案
- **THEN** 系统仅返回该服务人员服务过的客户档案（数据权限隔离），未服务过的客户不可见

### Requirement: 人群健康评估（Phase 1）
系统 SHALL 提供评估量表模板与评估记录能力，支持服务人员上门评估与 C 端自助评估。

- 表 `health_assessment_forms`：name、dimension(adl/barthel/fall/nutrition/cognition/pressure/weak/geriatric/self)、description、questions(JSON 题目数组)、score_rule(JSON 等级规则)、version、status(0草稿 1启用)。
- 表 `health_assessments`：user_id、form_id、assessor_type(1自助 2服务人员)、staff_id、answers(JSON)、total_score、level、conclusion、suggestions(JSON)、symptom_desc(主诉/需求描述)。
- 评估结果同步回写 health_records.assessment_level 与 suggestions。

#### Scenario: 服务人员上门完成评估
- **WHEN** 服务人员在服务人员小程序选择量表→逐题作答→提交
- **THEN** 系统计算总分与等级结论，保存评估记录并回写档案评估等级，且该客户出现在 web-admin 评估记录中

#### Scenario: 用户自助评估
- **WHEN** C 端用户在"我的健康"选择启用的量表作答并提交
- **THEN** 系统保存自助评估记录（assessor_type=1）并回写档案

#### Scenario: web-admin 配置量表
- **WHEN** 管理员在 web-admin 创建/编辑量表（含题目与评分规则）并启用
- **THEN** 启用的量表对 C 端自助评估与服务人员端评估可见，草稿不可见

### Requirement: 康复辅具适配（Phase 2）
系统 SHALL 基于评估结论生成辅具适配建议，并联动商品购买/租赁。

- 表 `fitting_recommendations`：user_id、assessment_id(可空)、symptom_desc、fitting_result(适配结论)、recommended_products(JSON：[{product_id,name,reason,sale_type}])、staff_id、status(0草稿 1已确认 2已下单)、order_id(可空)。
- 适配建议可关联现有商品（零售/租赁均可），C 端可"一键查看/加购"跳转商品页。

#### Scenario: 评估后生成适配建议
- **WHEN** 评估完成且服务人员/用户确认生成适配建议
- **THEN** 系统保存适配建议并推荐匹配商品（如 ADL 低分→助行/轮椅/护理床）

#### Scenario: 用户按适配建议加购
- **WHEN** C 端用户查看适配建议并点击推荐商品
- **THEN** 跳转对应商品详情页，可加入购物车或直接下单（复用现有下单链路）

### Requirement: 居家康养照护（Phase 3）
系统 SHALL 提供照护计划与上门照护记录，与现有服务工单流程打通。

- 表 `care_plans`：user_id、name、plan_type(1生活照料 2基础护理 3康复训练 4综合康养)、start_date、end_date、frequency、goals、items(JSON 护理项配置)、assigned_staff_id、order_id(可空)、status(0草稿 1执行中 2已暂停 3已完成)。
- 表 `care_visits`：plan_id、order_id(可空)、user_id、staff_id、visit_at、nursing_items(JSON 完成护理项)、vitals(JSON 血压/血糖/心率/血氧)、photos(JSON)、remark、follow_up_advice。
- 照护计划可通过"预约服务"订单进入待接单池，由服务人员接单执行；服务签退时可同步录入照护记录。

#### Scenario: 管理员创建并指派照护计划
- **WHEN** web-admin 管理员创建照护计划并指派服务人员（或生成预约服务订单）
- **THEN** 计划进入执行中，服务人员端"我的照护计划"可见

#### Scenario: 服务人员录入上门照护记录
- **WHEN** 服务人员上门后录入护理项、生命体征与照片
- **THEN** 系统保存上门照护记录并关联到计划/订单，C 端可查看摘要

### Requirement: 持续康复随访（Phase 4）
系统 SHALL 在关键节点自动生成随访任务并支持执行记录。

- 表 `follow_up_tasks`：user_id、task_type(1康复随访 2租后回访 3慢病随访 4评估回访)、source_type(1服务完成 2租赁归还 3评估完成 4手动)、source_id、plan_follow_time、staff_id、contact_method(1电话 2上门 3微信)、status(0待执行 1已完成 2已跳过)、result(JSON)、completed_at、remark。
- 触发节点：服务工单签退完成、租赁订单归还完成、评估完成后，自动为对应客户生成随访任务。

#### Scenario: 服务完成后自动生成随访任务
- **WHEN** 服务人员签退完成服务（biz_status→5）或租赁归还完成
- **THEN** 系统自动生成随访任务，服务人员端"待办随访"可见

#### Scenario: 服务人员执行随访
- **WHEN** 服务人员按随访任务完成电话/上门/微信随访并录入结果（含宣教内容、满意度）
- **THEN** 任务状态置为已完成，web-admin 随访记录可见

### Requirement: 老年慢病管理与生命体征监测（Phase 4）
系统 SHALL 维护慢病标签并提供生命体征监测记录。

- 表 `health_monitoring`：user_id、record_type(1血压 2血糖 3心率 4血氧 5体重)、value、unit、extra(JSON)、recorded_by(服务人员ID或0=用户)、recorded_at、remark。
- 慢病标签维护在 health_records.chronic_tags，供宣教定向与随访计划使用。

#### Scenario: 录入监测数据
- **WHEN** 服务人员上门测量或 C 端用户录入血压/血糖/心率/血氧/体重
- **THEN** 系统保存监测记录并可在档案时间线中展示

### Requirement: 健康宣教（Phase 4）
系统 SHALL 提供健康宣教内容管理与按慢病标签定向推送。

- 表 `health_education_articles`：title、category、cover、content、tags(JSON 定向慢病标签)、status(0草稿 1发布)、publish_at、views。
- C 端"我的健康"展示与用户慢病标签匹配的已发布宣教内容；服务人员在随访时可选派宣教内容。

#### Scenario: 定向推送宣教内容
- **WHEN** C 端用户查看"健康宣教"列表
- **THEN** 系统按该用户慢病标签优先展示匹配的已发布内容

#### Scenario: web-admin 管理宣教内容
- **WHEN** 管理员创建/编辑/发布健康宣教文章并设置定向标签
- **THEN** 发布后 C 端可见，草稿仅 web-admin 可见

### Requirement: 三端联动与 RBAC 权限（全局）
系统 SHALL 在 web-admin 增加"健康服务"菜单与权限码，并保证三端数据联动一致。

- 新增 RBAC 权限码：health:view/update、assessment:view/create/config、fitting:view/update、care:view/create、followup:view/update、monitor:view/create、education:view/create。
- 新增菜单项（web-admin 路由 + 侧边栏 + v-permission 按钮级）。
- 数据权限：健康档案/评估/监测数据仅本人（C 端）、服务过的客户（服务人员）、管理员（web-admin）可见。

#### Scenario: RBAC 生效
- **WHEN** 管理员角色被授予健康服务相关菜单与权限码
- **THEN** 侧边栏出现健康服务菜单，无权限账号访问返回 403 且按钮隐藏

---

## MODIFIED Requirements

### Requirement: 服务人员工单流程扩展
现有工单（orders.biz_status 接单/签到/签退）保持不变，扩展：
- 签退（CheckOut）完成后，若订单为服务/租赁类型，自动生成随访任务；
- 服务人员工单详情页可跳转查看关联客户的健康档案（数据权限内）。

### Requirement: C 端商城"我的"能力扩展
现有"我的订单/地址"保持不变，新增"我的健康"入口（健康档案、自助评估、适配建议、康复随访、健康宣教），复用现有登录鉴权与分页规范。

## REMOVED Requirements

无。
