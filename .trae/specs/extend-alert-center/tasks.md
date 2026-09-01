# Tasks

- [ ] Task 1: 数据库迁移与模型层
  - [ ] SubTask 1.1: 编写 `server/migrations/016_alert_center_extend.sql`（幂等，含 `SET NAMES utf8mb4;`）：
    - `service_alert_events.staff_id` 改为可空（MODIFY BIGINT UNSIGNED DEFAULT NULL）

    - `service_alert_events` 新增 `summary VARCHAR(256)` 列（address 之后）

    - `orders` 新增 `assigned_at DATETIME` 列

    - 新建 `alert_settings` 单行表（enabled + 6 个阈值列，含默认值，插入 id=1 默认行）

    - `sys_menus` 新增按钮节点 id=972（parent\_id=97，menu\_type=2，名称'预警设置'，permission='alert-settings:update'）并授权 role\_id=1

  - [ ] SubTask 1.2: `models.go`：`ServiceAlertEvent.StaffID` 改为 `*uint64`、新增 `Summary` 字段；`Order` 新增 `AssignedAt *time.Time`；新增 `AlertSettings` 模型（TableName='alert\_settings'）

  - [ ] SubTask 1.3: 本地执行迁移脚本，确认无报错（可查询 information\_schema 复核列）

- [ ] Task 2: 后端常量与预警中心接口
  - [ ] SubTask 2.1: `utils/constants.go`：新增预警类型常量 `AlertTypeGoodsUnverified=3 / AlertTypeServiceUnassigned=4 / AlertTypeEscortUnfinished=5 / AlertTypeServiceUnstarted=6 / AlertTypeRentalOverdue=7 / AlertTypeRefundStuck=8`，以及各默认阈值常量

  - [ ] SubTask 2.2: `alert_center.go`：`alertTypeText` 补充 3\~8 文案；`fillAlertStaffInfo` 过滤 nil/0 的 staff\_id；新增 `GetAlertSettings`（表空时创建默认行并返回）、`UpdateAlertSettings`（数值范围校验 + 幂等 upsert）

  - [ ] SubTask 2.3: `main.go` 新增路由：`GET /merchant/alert-settings`（RBAC "alert-events:view"）、`PUT /merchant/alert-settings`（RBAC "alert-settings:update"）

  - [ ] SubTask 2.4: `order_dispatch.go` 的 DispatchOrder 与 `service_staff/workorder.go` 的 AcceptOrder 写入 `assigned_at=now`

- [ ] Task 3: 定时任务（订单级超时预警扫描）
  - [ ] SubTask 3.1: 新增 `server/internal/tasks/order_timeout.go`：`loadAlertSettings()`（默认值兜底）、幂等写入 helper `ensureAlert(orderID, staffID, alertType, summary)`、6 个扫描函数（实物未核销/服务未指派/陪诊未完成/指派未签到/租赁逾期/退款卡住，均读配置阈值，enabled=false 时跳过）

  - [ ] SubTask 3.2: `main.go` 注册 `go tasks.StartOrderTimeoutTasks()`（启动即扫一次 + 5 分钟 ticker）

- [ ] Task 4: Web-admin 预警中心扩展
  - [ ] SubTask 4.1: `web-admin/src/api/safety.ts`：新增 `AlertSettings` 类型与 `getAlertSettings` / `updateAlertSettings`

  - [ ] SubTask 4.2: `AlertEventsView.vue`：预警类型筛选补 3\~8 选项；关联服务人员列 staff 为空时显示"-"；新增 summary 列；页面头部新增"预警设置"按钮（v-permission="'alert-settings:update'"）打开设置弹窗（enabled 开关 + 6 个数字输入，含单位与范围校验）

- [ ] Task 5: 构建与验证
  - [ ] SubTask 5.1: `cd server && go build ./...` 通过

  - [ ] SubTask 5.2: `cd web-admin && npm run build` 通过

  - [ ] SubTask 5.3: 接口冒烟：GET/PUT alert-settings、预警列表按 alert\_type=3\~8 筛选、staff 空值展示正确

# Task Dependencies

- Task 1 是 Task 2/3 的前置（迁移与模型先行）

- Task 3 依赖 Task 1（模型/常量）与 Task 2（assigned\_at 写入点）

- Task 4 依赖 Task 2 的接口

- Task 5 依赖 Task 1\~4

