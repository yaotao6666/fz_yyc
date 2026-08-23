# PC 端服务人员审核机制 + 资质文件自注册审核 — 实施计划

## Summary
面向**服务人员端（接单小程序员工）**完善审核机制：
1. 新增**审核记录表**，注册申请、PC 添加、服务人员在端上变更资料，均生成待审核记录并留痕。
2. **服务人员信息变更走审核**：端上"我的资料"提交修改 → 进入待审核 → 管理员审核通过后才**使变更生效**，且**不影响启用/禁用状态**（审核状态与启用状态解耦）。
3. **自注册提交资质材料**：服务人员注册/变更时通过七牛上传资质文件（执业证/身份证/资质证书等）生成 URL 提交，后台可查看材料并审核。
4. **独立权限控制**：新增"审核中心/审核列表"菜单与**独立权限码**，与现有启停（`staff:update`）等权限区分。

## Current State Analysis
现状（已探索）：
- `service_staffs` 表字段：`username/password/name/phone/openid/avatar/status(0待审1启用2禁用)/last_login_at`。单一 `status` 字段混合"审核"与"启用"语义。
- 注册：`POST /api/v1/service-staff/register`（[auth.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/service_staff/auth.go)），创建 `status=0`。登录时若 `status==0` 报"待审核"、`==2` 报"已禁用"。
- PC 添加：`POST /api/v1/merchant/service-staff` 创建 `status=1`（本计划前刚实现）。
- 状态更新：`PUT /api/v1/merchant/service-staff/:id/status`（`staff:update`），仅 1/2。
- 无"信息变更"接口、无"资质材料"字段、无"审核记录表"、无独立审核权限。
- `sys_menus` 服务人员菜单：id=4 顶级（staff:view）、747（列表页 staff:view）、41 staff:update、42 staff:reset-password、43 staff:delete、44 staff:create。
- 上传走七牛：`GET /api/v1/upload/token`（[upload.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/upload/upload.go)），可复用。
- RBAC：`sys_menus`（菜单/按钮）、`sys_roles`、`sys_role_menus`、`merchant_staff_roles`；owner 直通，普通员工走权限码+60s 缓存。前端动态菜单 `AppLayout.vue` 按 `menu_type=1` 渲染，路由 `meta.permission` + `v-permission` 指令。
- 前端服务人员页 [ServiceStaffListView.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/views/staff/ServiceStaffListView.vue)；`sp.ts` 已有 staff API。
- 服务人员端 `staff-miniprogram`：注册页 [login/index.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/staff-miniprogram/src/pages/login/index.vue)，"我的"页 [profile/index.vue](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/staff-miniprogram/src/pages/profile/index.vue)；API 在 [index.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/staff-miniprogram/src/api/index.ts)。
- 迁移文件已整合为 `server/migrations/001_schema.sql` 与 `002_seed.sql`（原 health/RBAC 迁移已并入）。

## Proposed Changes

### 1. 数据模型与迁移

**A. 扩展 `service_staffs` 表**（新增资质与审核相关列，幂等迁移）：
- `qualifications` JSON：资质文件 URL 数组（`[{type,name,url}]`），注册/变更时提交的材料。
- `audit_status` tinyint：0=无/已通过 1=待审核（注册后默认待审核）。与 `status` 解耦。
- `pending_fields` JSON（可选）：正在审核中的待变更字段快照（用于后台对比），审核通过后回写正式字段。
- 说明：`status` 继续表示"启用/禁用"；新增的审核态由 `audit_status` 承接；二者不冲突。

**B. 新增 `staff_audit_records` 表**（审核记录留痕）：
- 字段：`id, staff_id, audit_type(1=注册申请 2=信息变更 3=资质提交), apply_type(1=自注册 2=PC添加 3=端上变更), before_data JSON, after_data JSON, qualifications JSON, status(0=待审 1=通过 2=驳回), reviewer_id, review_remark, review_at, created_at`。
- 每次注册/变更/资质提交都写一条待审记录；管理员通过/驳回时更新记录状态，同时回写 `service_staffs`（通过后）或仅留痕（驳回）。

**C. RBAC 菜单与权限**（`002_seed.sql` 或新增幂等 seed）：
- 服务人员顶级(4)下新增子菜单：**「审核列表」**（`/staff/audits`，`staffaudit:view`，`menu_type=1`）。
- 该页按钮：审核通过（`staffaudit:approve`）、驳回（`staffaudit:reject`）。
- 独立于 `staff:update/`：管理员有审核权限但不一定有点册启停权限，二者独立。绑定超管角色(id=1)。

### 2. 后端（Go）

**A. `internal/models/models.go`**：新增 `StaffAuditRecord` 结构体；扩展 `ServiceStaff` 增加 `Qualifications`（JSON）、`AuditStatus`、`PendingFields`（JSON）字段。

**B. `internal/handlers/service_staff/`**：
- 新增 `profile.go` 或扩展 `auth.go`：
  - `RequestProfileChange`（端上变更资料）：请求体含 name/phone/avatar/qualifications；校验变化 → 写入 `staff_audit_records`（audit_type=2，after_data 存新值、qualifications 存材料）→ 置 `audit_status=1`（不直接改正式字段）。
  - `GetMyAuditList`：服务人员端查看本人审核记录与状态。
- 修改 `Register`：接收可选 `qualifications`，注册即写审核记录（audit_type=1/apply_type=1），`status=0`+`audit_status=1`。

**C. `internal/handlers/merchant/`**：
- 新增 `audit.go`：
  - `ListStaffAudits`（PC 后台审核列表，分页/关键词/状态筛选）。
  - `StaffAuditDetail`（详情：申请人信息 vs 待变更字段、资质材料 URL 预览）。
  - `ApproveStaffAudit` / `RejectStaffAudit`：更新 `staff_audit_records.status`；**通过时**：
    - 若 audit_type=1（注册申请）：置 `service_staffs.status=1`（启用）+ `audit_status=0`。
    - 若 audit_type=2（信息变更）：将 after_data 回写 `name/phone/avatar/qualifications` 正式字段，置 `audit_status=0`，status 保持不变（不强改启用/禁用）。
    - 记录 reviewer/review_at/remark。
- 修改 `CreateServiceStaff`：PC 添加也写一条审核记录（audit_type=1/apply_type=2），但默认直接通过（或同样入库审核）。为"变更信息均需审核"，PC 添加直接启用并写一条已通过的留痕记录即可。
- 修改 `UpdateServiceStaffStatus`：保留，但该操作也应写一条审核留痕记录（audit_type 用 4=状态变更 或复用）供留存。

**D. 路由注册**（`cmd/server/main.go`）：
- 服务人员端（/service-staff 需登录）新增：
  - `GET /api/v1/service-staff/audits`（本人审核列表）
  - `PUT /api/v1/service-staff/profile` 或复用（资料变更，需登录）
- 商户端（/merchant，RBAC）新增：
  - `GET /api/v1/merchant/staff-audits` → `staffaudit:view`
  - `GET /api/v1/merchant/staff-audits/:id` → `staffaudit:view`
  - `POST /api/v1/merchant/staff-audits/:id/approve` → `staffaudit:approve`
  - `POST /api/v1/merchant/staff-audits/:id/reject` → `staffaudit:reject`

### 3. web-admin 前端
- `router/index.ts`：新增 `/staff/audits` 路由，`meta.permission='staffaudit:view'`（服务人员顶级下子菜单）。
- `views/staff/StaffAuditListView.vue`（新）：审核列表+筛选+详情弹窗（含资质材料图片预览）+通过/驳回操作。等 pull，`api/sp.ts` 新增 `listStaffAudits/getStaffAuditDetail/approveStaffAudit/rejectStaffAudit`；`types/sp.ts` 加类型。
- `ServiceStaffListView.vue`：可保留现有"审核通过/启用"（写员工状态），审核中心独立。

### 4. staff-miniprogram 前端
- `pages/login/index.vue`：注册表单增加**资质材料上传**（七牛 token → 选择图片/文件 → 上传获得 URL），随注册提交 `qualifications`。
- 新增 `pages/profile/edit.vue` 或改造 `profile/index.vue`：资料变更表单 + 资质材料上传，提交后走审核（提示"已提交审核"）。"我的"页新增"我的认证记录/审核进度"入口（`pages/profile/my-audits.vue`）。
- `pages.json`：注册这些新页面。
- `api/index.ts`：新增 `staffProfileApi.requestChange`、`getMyAuditList`、`staffAuthApi.register` 支持 `qualifications`；复用 `upload`（需确认 staff 端是否已有上传封装，无则加）。

### 5. 文档（PRD 同步，遵循 RBAC/文档规则）
- 更新 `PRD.md`、`docs/prd/PRD-接口文档.md`（新增接口表+权限码 `staffaudit:view/approve/reject`）、`PRD-功能说明.md`（服务人员审核机制、资质自注册流程）。
- 在 RBAC 权限码维护规范中补 `staffaudit:*`。

## Assumptions & Decisions
- 仅服务人员端（用户已确认）。
- 注册 + 变更均走审核；审核通过才生效，且审核与启用状态解耦（`audit_status` vs `status`）。
- 资质材料经七牛上传存 URL（用户已确认）。
- 审核为独立权限 `staffaudit:view/approve/reject`，与 `staff:update` 分离。
- 迁移采用幂等写法（`CREATE TABLE IF NOT EXISTS` / `ALTER TABLE ADD COLUMN IF NOT EXISTS` 用 INFORMATION_SCHEMA 判断 / `INSERT IGNORE` 菜单）。
- 因迁移已整合为 001/002，新增结构追加到这两个文件或新增 `003_*.sql`（倾向新增 `003_audit_qualification.sql + 004_seed.sql` 或并入现有，执行时以实际目录为准）。

## Verification
1. 后端 `go build ./...` 通过。
2. 迁移执行成功：`staff_audit_records` 表存在；`service_staffs` 新增列；`sys_menus` 出现审核菜单与权限，绑定角色1。
3. 手动流程：服务人员注册携带资质 → 生成待审记录 → PC 审核通过 → 状态置启用；服务人员在端上改手机号 → 生成待审（status 不变）→ 审核通过 → 手机号变更生效。
4. web-admin `npm run build` 通过，`/staff/audits` 页面可访问（有 `staffaudit:view` 权限账号）。
5. staff-miniprogram `npm run build` 通过。
6. 容器重建，新接口路由注册、`/health` 200。

## File Path Reference
- 迁移：`server/migrations/`（001_schema.sql / 002_seed.sql / 新增 003 或追加）
- 模型：`server/internal/models/models.go`
- 服务人员 handler：`server/internal/handlers/service_staff/auth.go`（+ 新增 profile）
- 商户 handler：`server/internal/handlers/merchant/service_staff.go`（+ 新增 audit.go）
- 路由：`server/cmd/server/main.go`
- 上传：`server/internal/handlers/upload/upload.go`（复用）
- web-admin：`router/index.ts`、`api/sp.ts`、`types/sp.ts`、`views/staff/`（新增）
- staff-miniprogram：`api/index.ts`、`pages.json`、`pages/login/index.vue`、`pages/profile/*`