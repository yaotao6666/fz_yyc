# PRD-接口文档

## 1. 文档定位

本文件用于承接 `PRD.md` 中与“接口、返回结构、字段口径、空值约定、WebSocket 协议”相关的内容。

读取顺序：

1. 先读根目录 `PRD.md`
2. 再读本文件
3. 如需功能背景，联读 `docs/prd/PRD-功能说明.md`

## 2. 接口分组

### 2.1 认证相关接口

- 商户管理员登录（PC后台）
- 服务人员登录
- 服务人员注册申请
- C 端微信登录

关键约束：

- 登录态字段、缓存字段、角色标识必须在前后端统一。
- 商户管理员退出登录后，前端需要主动断开 WebSocket 连接。

#### 商户管理员登录

- **接口**：`POST /api/v1/auth/admin/login`
- **请求**：`{ username: string, password: string }`
- **鉴权使用**：
  - 登录成功后返回 PC 后台 token，用于调用 `/api/v1/merchant/*` 全部接口
  - `Authorization: Bearer {admin_token}` 用于 PC 后台所有请求

#### 服务人员登录

- **账号密码登录**：`POST /api/v1/service-staff/login`
  - 请求：`{ username: string, password: string }`
- **微信快捷登录**：`POST /api/v1/service-staff/wechat-login`
  - 请求：`{ code: string }`
- **鉴权使用**：
  - 登录成功后返回服务人员 token，用于调用 `/api/v1/service-staff/*` 全部接口
  - `Authorization: Bearer {staff_token}` 用于服务人员接单小程序所有请求

#### 服务人员注册申请

- **接口**：`POST /api/v1/service-staff/register`
- **请求**：包含服务人员基本信息（姓名、手机号、登录账号、密码等）
- **流程**：注册申请提交后进入待审核状态，由商户管理员在 PC 后台执行审核，审核通过后服务人员方可登录接单小程序。

#### C 端微信登录

- **接口**：`POST /api/v1/auth/user/wechat-login`
- **请求**：`{ code: string }`
- **成功响应**：
  - `code`: `0`
  - `message`: `success`
  - `data.token`: C 端用户 token
  - `data.app_mode`: 当前生效的小程序支付身份，取值为 `sp_app` 或 `sub_app`
  - `data.app_id`: 当前生效的小程序 `appid`
  - `data.user`：
    - `id`
    - `openid`
    - `nickname`
- **前端缓存字段**：
  - `user_token`：对应 `data.token`
  - `userInfo`：对应 `data.user`
  - `openid`：对应 `data.user.openid`
  - `user_login_app_mode`：对应 `data.app_mode`
  - `user_login_app_id`：对应 `data.app_id`
- **鉴权使用**：
  - `Authorization: Bearer {user_token}` 用于 `GET/POST /api/v1/user/*`
  - `Authorization: Bearer {user_token}` 用于 `POST /api/v1/store/orders`（下单）
  - `/api/v1/store/home`、`/products`、`/products/:product_id`、`/delivery-rules` 为公开接口，不要求登录
  - `pages/home/index`、`pages/store/product`、`pages/store/confirm` 统一解析店铺入口参数（单商户模式，入口固定单店）
  - `GET /api/v1/store/home` 与 `/api/v1/store/delivery-rules` 进入页面后可直接发起，不等待登录完成

#### C 端访问与埋点

- **访问埋点**：`POST /api/v1/store/visit`
  - 请求：`{ openid: string, source?: string }`
- **行为埋点**：`POST /api/v1/store/event`
  - 请求：
    - `openid: string`
    - `event_type: page_view | product_view | submit_order | pay_success`
    - `page?: string`
    - `product_id?: number`
    - `order_id?: number`
    - `source?: string`
    - `payload?: object`

#### 商户公告接口说明

- 商户公告管理由 PC 后台承接，使用 `/api/v1/merchant/announcements*` 接口完成公告列表、详情、新增、编辑和删除（停用）管理。
- C 端商城首页公告展示读取已发布公告。

### 2.2 商户PC后台接口

- 当前调用端：`web-admin/` 商户 PC 后台。
- 仪表盘
- 商户资料：`GET/PUT /api/v1/merchant/profile`、`PUT /api/v1/merchant/assets`
- 商户支付配置：`PUT /api/v1/merchant/payment-config`
- 订单列表：`GET /api/v1/merchant/orders`
- 订单详情：`GET /api/v1/merchant/orders/:order_id`
- 商户费率
- 商户二维码：`GET /api/v1/merchant/qrcode`
- 系统公告：`/api/v1/merchant/announcements*`
- 数据分析：`/api/v1/merchant/analytics/*`
- 服务人员管理：`/api/v1/merchant/service-staff*`
- 服务人员审核中心：`/api/v1/merchant/staff-audits*`
  - 审核列表：`GET /api/v1/merchant/staff-audits`（`staffaudit:view`）
  - 审核详情：`GET /api/v1/merchant/staff-audits/:id`（`staffaudit:view`，含资质材料 URL 与变更前后数据）
  - 审核通过：`POST /api/v1/merchant/staff-audits/:id/approve`（`staffaudit:approve`；注册申请通过则启用账号，信息变更通过则回写正式字段）
  - 审核驳回：`POST /api/v1/merchant/staff-audits/:id/reject`（`staffaudit:reject`）

当前重点约束：

- 商户数据分析接口需按以下维度输出：
  - 访问率
  - 下单率
  - 下单金额
  - 下单均价
  - 日 / 周 / 月 / 年订单量
  - 商品排行维度切换
- `GET /api/v1/merchant/orders` 需支持 `status`、`start_date`、`end_date`、`keyword`、`page`、`page_size` 筛选，返回当前商户名下订单。
- `GET /api/v1/merchant/orders/:order_id` 返回订单详情结构，包含商品明细、用户信息、配送信息、支付单号、核销码、核销时间与核销人。
- `GET /api/v1/merchant/orders/analytics` 返回 `day/week/month/year` 四组订单量桶。
- `/api/v1/merchant/announcements*` 商户公告接口由 PC 后台直接调用。
- 商品分类、商品、规格管理通过 `/api/v1/merchant/categories*`、`/api/v1/merchant/products*`、`/api/v1/merchant/products/:product_id/specs` 接口完成，由 PC 后台直接调用，无需透传 `merchant_id`。

### 2.3 商户管理接口（PC后台）

- 商户资料查询 / 更新
- 商户设置
- 商户满减规则：`GET/PUT /api/v1/merchant/full-reduction-rules`
- 商户营业状态
- 配送设置
- 商户二维码
- 商户打印机管理：`GET/POST/PUT/DELETE /api/v1/merchant/printers*`
- 打印测试：`POST /api/v1/merchant/printers/:printer_id/test`
- 商品分类：`GET/POST/PUT/DELETE /api/v1/merchant/categories*`
- 商品管理：`GET/POST/PUT/DELETE /api/v1/merchant/products*`
- 商品规格：`GET/PUT/DELETE /api/v1/merchant/products/:product_id/specs`
- 服务人员管理：
  - 服务人员列表：`GET /api/v1/merchant/service-staff`
  - 新增服务人员：`POST /api/v1/merchant/service-staff`
  - 更新服务人员：`PUT /api/v1/merchant/service-staff/:id`
  - 删除服务人员：`DELETE /api/v1/merchant/service-staff/:id`
  - 审核服务人员注册申请：`PUT /api/v1/merchant/service-staff/:id/audit`
- 订单核销：
  - 当前订单核销：`POST /api/v1/merchant/orders/:id/complete`
  - 快速核销：`POST /api/v1/merchant/orders/quick-complete`

当前重点约束：

- 商户资料更新需支持 `logo` 与背景图字段。
- `GET /api/v1/merchant/settings` 需返回 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled`，`PUT /api/v1/merchant/settings` 需支持更新这三个开关。
- `GET /api/v1/store/delivery-rules` 除配送费结构外，还需返回 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled` 供确认页动态展示下单方式。
- `GET /api/v1/merchant/full-reduction-rules` 返回 `rules` 与 `active_rules`，单档规则至少包含 `threshold_amount`、`discount_amount`、`status`、`sort`。
- `PUT /api/v1/merchant/full-reduction-rules` 最多支持 5 档规则，`discount_amount` 必须小于 `threshold_amount`。
- `GET /api/v1/store/full-reduction-rules` 为公开接口，仅返回当前商户启用中的满减规则。
- `delivery_settings.enabled` 只表示配送费规则是否生效，确认页是否展示“配送”必须以后端返回的 `takeout_enabled` 为准。
- `GET /api/v1/merchant/printers` 返回打印机列表时，需返回 `type`、`status`、`auto_print`、`is_default`、`print_count`、`last_print_at`、`has_api_key`、`has_feie_ukey`。
- 飞鹅打印机请求字段至少包含 `feie_user`、`feie_ukey`、`feie_sn`。
- `GET /api/v1/merchant/qrcode` 必须固定生成指向 `pages/home/index` 的小程序码，且 `scene` 固定为店铺标识（单商户模式，不包含 `merchant_id`）。
- 服务人员管理接口用于维护接单小程序的账号、姓名、手机号、状态等，注册申请审核通过后服务人员方可登录接单。
- 订单核销由商户管理员在 PC 后台执行，无需重复输入核销码。

### 2.4 商品与分类接口

- 商品分类列表 / 新增 / 修改 / 删除
- 商品列表 / 详情 / 新增 / 更新 / 删除
- 商品规格相关接口

当前重点约束：

- 商品图片字段以当前接口返回结构为准。
- 商品相关列表接口返回空值时，数组字段统一返回 `[]`。
- PC 后台可通过 `PUT /api/v1/merchant/products/:product_id` 手动修改商品销量（`sales` 字段），由商户管理员操作。
- **租赁商品字段**：商品新增/更新接口需支持以下字段，并在商品列表与详情接口中返回：
  - `sale_type` 销售类型：1 一口价（默认）/ 2 租赁
  - `rental_unit` 计费周期：1 按天 / 2 按周 / 3 按月（租赁商品必填）
  - `rental_price` 单位租金
  - `deposit` 押金
  - `max_rental_duration` 最大租赁时长（0=不限）
- 商品列表接口支持按 `sale_type` 筛选；租赁商品在列表中应展示 `rental_price`、`rental_unit`、`deposit`，便于前端按 `¥{rental_price}/{unit}` 显示。

### 2.5 订单接口

- 商户订单列表
- 商户订单详情
- 当前订单核销
- 快速核销：`POST /api/v1/merchant/orders/quick-complete`
- 商户退款
- 订单统计
- **归还租赁商品（退押金）**：`POST /api/v1/merchant/orders/{order_id}/return`
- **派单（指派服务人员）**：`GET /api/v1/merchant/orders/dispatchable-staff`（`order:dispatch`，返回已审核通过服务人员）、`POST /api/v1/merchant/orders/{order_id}/dispatch`（`order:dispatch`，入参 `staff_id`）
- **租赁到期提醒**：`GET /api/v1/merchant/orders/rental-due`（`orderrental:view`，`due_range`=soon/overdue，返回未归还租赁订单与剩余天数）
- **续租**：`POST /api/v1/merchant/orders/{order_id}/renew`（`order:renew`，生成 `parent_order_id=原单`、`renew_flag=1` 的新待支付订单，可选 `duration`）
- C 端订单列表 / 详情 / 取消 / 申请退款

当前重点约束：

- 派单仅可选择**已审核通过**的服务人员（`service_staffs.status=1 AND audit_status=0`）；指派后订单 `biz_status=2`（待出发），服务人员在工单「已接订单/待办」可见并可签到/签退。
- 租赁订单支付成功时按 `paid_at + 最长租赁时长`（天/周/月折算）写入 `rental_end_at`，供到期提醒使用。
- 订单详情应返回：
  - `verify_code`
  - `completed_at`
  - `completed_by_name`
- 商户订单列表接口支持：
  - `status`
  - `start_date`
  - `end_date`
- 商户管理员对当前订单执行核销时，在 PC 后台操作，不再要求重复输入核销码。
- 商户退款接口需与前端请求参数保持一致，至少明确：
  - `reason`
  - `refund_amount`
- 商户退款接口在未传 `refund_amount` 或传入 `<= 0` 时，默认按订单实付金额处理。
- C 端创建订单接口需按 `delivery_type` 与三字段一一校验：
  - `delivery_type=1` 校验 `takeout_enabled`
  - `delivery_type=2` 校验 `dine_in_enabled`
  - `delivery_type=3` 校验 `pickup_enabled`
- 当对应方式未开启时，接口分别返回“商户暂未开启配送 / 堂食 / 自提”。
- C 端确认订单页展示的满减优惠仅作提示，下单结果必须以后端返回的 `discount_amount`、`pay_amount` 为准。
- 创建订单接口需按商户启用中的满减规则重算：
  - `discount_amount`
  - `pay_amount`
- **租赁商品下单约束**：
  - `items[].rental_duration` 为租赁商品必填字段，按商品 `rental_unit` 计；一口价商品可不传或传 0
  - 服务端按 `rental_price × rental_duration × quantity` 计算租金、`deposit × quantity` 计算押金
  - 订单 `total_amount` 仅含租金（一口价为商品金额），`total_deposit` 为押金合计，`pay_amount = total_amount + total_deposit + delivery_fee - discount_amount`
  - 创建订单与订单详情接口的 `order` 对象需返回 `total_deposit`、`deposit_status`、`deposit_refund_amount`、`deposit_deduct_amount`、`deposit_refunded_at`、`rental_returned_at`、`rental_return_remark`
  - 订单详情的 `items` 需返回 `sale_type`、`rental_unit`、`rental_duration`、`unit_rental_price`、`rental_subtotal`、`deposit`、`deposit_deduct`
- **归还租赁商品接口约束**：
  - 仅 `status=paid` 且 `total_deposit > 0` 且 `deposit_status=1` 的订单可发起归还
  - 请求体：`deduct_amount`（扣除金额，0=全额退还）、`remark`（验机备注）
  - 调用后写入 `rental_returned_at`、`deposit_deduct_amount`、`deposit_refund_amount`、`deposit_status`（2 已退还 / 3 已扣除）、`deposit_refunded_at`
  - 系统通过微信支付原路退还 `deposit_refund_amount`，与订单退款接口相互独立
- **退款状态口径**：
  - 订单 `status=5`：退款中（已发起退款流程；接口会先主动同步一次微信退款状态，未拿到最终结果时继续等待微信退款结果）
  - 订单 `status=6`：已退款（接口主动同步或收到微信退款成功回调后写入）
  - `refunded_at`：在订单进入 `status=6` 时写入
- 支付成功后，订单需回写：
  - `transaction_id`
  - `pay_notify_payload`
- 支付成功并完成订单状态回写后，后端向 PC 后台 WebSocket 推送 `order_notify`，并向服务人员小程序推送待接订单提醒。
- 订单接单由服务人员通过接单小程序执行，订单记录写入 `accepted_by_staff_id` 与 `accepted_at`；订单核销由商户管理员在 PC 后台执行。

### 2.6 支付接口

- C 端下单支付由商户主体统一拉起。
- 全局配置 `WECHAT_PAY_APP_MODE` 控制当前部署版本的小程序支付身份：
  - `sp_app`：请求微信支付时使用 `sp_appid + payer.sp_openid`
  - `sub_app`：请求微信支付时使用 `sub_appid + payer.sub_openid`
- 商户支付配置字段：
  - `sub_mch_id`
  - `payment_config_status`
- 微信支付回调入口：
  - `/api/v1/notify/payment`
  - `/api/v1/callback/wechat`
  - 支付与退款均使用同一回调入口，按 `event_type` 区分支付（TRANSACTION）与退款（REFUND）

当前重点约束：

- 下单时必须按商户维度读取 `sub_mch_id`。
- 已进件商户通过 PC 后台回填 `sub_mch_id` 完成支付配置。
- 微信支付商户凭证为系统级配置，由系统统一维护，不在商户维度暴露。
- `WECHAT_PAY_APP_MODE` 仅切换登录与下单支付时使用的小程序身份，不改变支付回调与退款实现。

### 2.7 数据分析接口

#### PC后台

- 今日概览
- 订单趋势
- 客户分析
- 商品排行
- 库存预警

当前重点约束：

- 所有列表型分析接口必须保证空数组兜底。
- 所有对象型统计字段在无数据时应返回 `0` 或明确的空对象结构。
- `GET /api/v1/merchant/analytics/stock-alert` 默认使用阈值查询低库存商品，当前前端按 `threshold = 10` 使用。
- 库存预警接口返回结果按库存升序处理，前端据此做风险分组和看板展示。

### 2.8 上传接口

- 上传 token
- 七牛上传地址
- 图片前缀 / 域名

当前重点约束：

- 前端禁止写死上传地址。
- 图片上传成功后的持久化字段口径必须在前后端统一。

### 2.9 WebSocket 与联调接口

- PC 后台 WebSocket：`/api/v1/ws/merchant`
- 开发联调接口：
  - 订单提醒下发：`POST /api/v1/dev/order-notify`
  - 用户进店提醒下发：`POST /api/v1/dev/store-visit-notify`

当前重点约束：

- 商户管理员登录 PC 后台后建立连接。
- 商户管理员退出登录时主动断开。
- 登录失效触发 401 时，也要收口前端登录态并主动断开。
- `order_notify` 用于支付成功后的新订单提醒；`store_visit_notify` 用于顾客进店提醒。
- 支付成功后可向服务人员小程序推送待接订单提醒，用于服务人员接单。
- 非开发环境不应暴露不必要的联调接口。

### 2.10 服务人员接口

服务人员接单小程序专用接口，独立身份，与 C 端用户、商户管理员账号体系分离。

> **表命名规范**：服务人员相关业务表统一使用 `service_` 前缀（如 `service_staffs`、`service_staff_audit_records`）。后续新增服务人员相关的表/列/记录表时须保持一致前缀，避免混用 `staff_*` 等无前缀命名。

- 注册申请：`POST /api/v1/service-staff/register`
  - 请求：包含服务人员基本信息，可选资质材料 `qualifications:[{type,name,url}]`
  - 提交后进入待审核状态（`status=0`、`audit_status=1`），并生成注册申请审核记录，由商户管理员在 PC 后台审核
- 账号密码登录：`POST /api/v1/service-staff/login`
- 微信快捷登录：`POST /api/v1/service-staff/wechat-login`
- 资料变更申请：`PUT /api/v1/service-staff/profile`
  - 请求：`{name, phone, avatar?, qualifications?:[{type,name,url}]}`
  - 服务人员端上提交 → 生成信息变更审核记录（`audit_type=2`），置 `audit_status=1`，不直接改正式字段；审核通过后回写生效，且不影响启用/禁用状态
- 我的审核记录：`GET /api/v1/service-staff/audits`
- 待接订单列表：`GET /api/v1/service-staff/orders/pending`
- 接单：`POST /api/v1/service-staff/orders/:id/accept`
  - 接单成功后订单写入 `accepted_by_staff_id`、`accepted_at`
- 已接订单列表：`GET /api/v1/service-staff/orders/accepted`
- 订单详情：`GET /api/v1/service-staff/orders/:id`
- 接单统计：`GET /api/v1/service-staff/statistics`

当前重点约束：

- 服务人员仅能查看与接单，无核销权限；订单核销由商户管理员在 PC 后台执行。
- 待接订单列表与已接订单列表均需保证空数组兜底。
- 接单为原子操作，避免多服务人员重复接单。

### 2.11 健康服务接口（基层健康服务闭环）

基层健康服务闭环分四期落地：阶段一覆盖居民健康档案与人群健康评估，阶段二覆盖康复辅具适配，阶段三覆盖居家康养照护（照护计划制定与上门照护记录），阶段四覆盖持续康复随访、生命体征监测与健康宣教。接口按端分为三组，C 端与服务人员端需登录，管理端在商户 RBAC 基础上按权限码校验。

### 2.12 RBAC 权限码维护规范（接口 ↔ 按钮 ↔ 菜单节点三方一致）

> **硬性约束**：所有商户管理端接口必须接入 RBAC 鉴权，且保证 **前端按钮权限码 = 后端接口鉴权码 = `sys_menus.permission` 节点** 三方完全一致。新增接口时必须同步生成对应的菜单权限节点，禁止出现"前端有按钮但接口无鉴权码"或"接口用聚合码但前端用细分码"的分叉。

#### 2.12.1 权限码命名约定

- **页面级权限码**：仅读取能力，格式 `模块:view`（如 `health:view`、`assessment:view`、`care:view`），挂载于菜单目录（`menu_type=1`）节点，同时用作前端路由守卫。
- **按钮级权限码**：写操作，格式 `模块:create | update | delete`（可含 `status` / `config` 等动作语义），挂载于按钮节点（`menu_type=2`），用于 `v-permission` 按钮级控制。
- **禁止聚合混合**：同一个模块不得既用细分的 `assessment:create/update/delete` 又用聚合的 `assessment:config`，二者只选其一，且全局统一。

#### 2.12.2 新增接口的 RBAC 落地流程（自动插入到对应节点）

新增/修改商户管理端接口时，必须依次完成以下四步（缺一不可）：

1. **后端接入**：在 `server/cmd/server/main.go` 的路由注册处为接口添加 `middleware.RBAC("模块:动作")`；改动权限码时同步更新 `internal/utils/constants.go` 及 `internal/middleware/rbac.go`（如需）。
2. **前端按钮**：在对应页面按钮上使用 `v-permission="'模块:动作'"`，与后端接口鉴权码完全一致。
3. **数据库菜单节点**：编写幂等迁移脚本（`server/migrations/*.sql`），用 `INSERT IGNORE INTO sys_menus ...` 在对应父节点（`parent_id`）下新增/更新按钮节点，`permission` 与上两步一致；**再 `INSERT IGNORE INTO sys_role_menus (role_id, menu_id) SELECT 1, id FROM sys_menus WHERE id IN (...)` 绑定超管角色**，确保既有账号不影响。
4. **文档同步**：同步更新本文件各阶段"管理端接口-权限码"表、`PRD-功能说明.md` 的 RBAC 描述，以及执行迁移脚本。

> **检查点**：交接或合并前，用 SQL 核对 `sys_menus` 中没有一个 `permission` 值被接口使用却未在节点中登记（`SELECT permission FROM sys_menus WHERE permission<>''` 与后端全部 RBAC 权限码做差集校验）。

#### 2.12.3 健康服务权限码现状（细粒度，前/后端/DB 一致）

| 子模块 | 页面级 | 按钮级（create/update/delete 或动作） |
| --- | --- | --- |
| 健康档案 | `health:view`（id=81） | 编辑 `health:update`（id=811） |
| 评估量表 | `assessment:view`（id=82） | 新增 `assessment:create`（id=821）· 编辑/启停 `assessment:update`（id=822）· 删除 `assessment:delete`（id=823） |
| 评估记录 | `assessment:view`（id=83） | 手动登记 `assessment:create`（id=831） |
| 适配建议 | `fitting:view`（id=84） | 编辑/确认 `fitting:update`（id=841） |
| 照护计划 | `care:view`（id=85） | 新增 `care:create`（id=851）· 编辑 `care:update`（id=852）· 删除 `care:delete`（id=853） |
| 随访任务 | `followup:view`（id=86） | 执行/登记 `followup:update`（id=861） |
| 生命体征 | `monitor:view`（id=87） | 录入 `monitor:create`（id=871） |
| 健康宣教 | `education:view`（id=88） | 新增/编辑 `education:create`（id=881） |
| 服务人员 | `staff:view`（id=4/747） | 添加 `staff:create`（id=44）· 审核/启停 `staff:update`（id=41）· 重置密码 `staff:reset-password`（id=42）· 删除 `staff:delete`（id=43） |
| 服务人员审核 | `staffaudit:view`（id=75，菜单 `/staff/audits`） | 审核通过 `staffaudit:approve`（id=751）· 驳回 `staffaudit:reject`（id=752） |
| 订单（派单/续租） | `orders:view`（id=2/745） | 派单 `order:dispatch`（id=930）· 续租 `order:renew`（id=931） |
| 租赁到期提醒 | `orderrental:view`（id=93，菜单 `/orders/rental-due`） | 续租 `orderrental:renew`（id=932）· 归还 `orderrental:return`（id=933） |

#### 阶段一：居民健康档案与人群健康评估

##### C 端（前缀 /api/v1/user，需用户登录）

| 接口 | 说明 |
| --- | --- |
| `GET /health-record` | 我的健康档案，未建档返回 `data=null` |
| `PUT /health-record` | 有则更新、无则创建我的健康档案 |
| `GET /assessment-forms` | 仅返回已启用的评估量表列表 |
| `GET /assessments` | 我的评估记录（分页倒序） |
| `POST /assessments` | 自助评估提交 `{form_id, answers:{key:label}, symptom_desc}`，后端计分并返回 `total_score` / `level` / `conclusion` |

##### 服务人员端（前缀 /api/v1/service-staff，需服务人员登录）

| 接口 | 说明 |
| --- | --- |
| `GET /assessment-forms` | 仅返回已启用的评估量表列表 |
| `GET /residents/:user_id/health-record` | 查看客户健康档案（数据权限校验） |
| `GET /residents/:user_id/assessments` | 客户评估记录（分页，数据权限校验） |
| `POST /residents/:user_id/assessments` | 上门评估登记，`assessor_type=2`（数据权限校验） |

##### 管理端（前缀 /api/v1/merchant + RBAC）

| 接口 | 权限码 | 说明 |
| --- | --- | --- |
| `GET /health-records` | `health:view` | 健康档案列表，`keyword` / `assessment_level` 筛选，分页，含用户信息 |
| `GET /health-records/:id` | `health:view` | 健康档案详情，含该用户全部评估记录 |
| `PUT /health-records/:id` | `health:update` | 编辑健康档案 |
| `GET /assessment-forms` | `assessment:view` | 评估量表列表（含草稿） |
| `POST /assessment-forms` | `assessment:create` | 新增评估量表 |
| `PUT /assessment-forms/:id` | `assessment:update` | 编辑评估量表 |
| `PATCH /assessment-forms/:id/status` | `assessment:update` | 启用/停用评估量表（仅更新 status，1 启用 / 0 草稿） |
| `DELETE /assessment-forms/:id` | `assessment:delete` | 删除评估量表 |
| `GET /health-assessments` | `assessment:view` | 评估记录列表，`keyword` / `form_id` 筛选，分页 |

##### 公共约束

- **数据权限**：服务人员仅能查看/评估自己接单服务过的客户（按 `orders.assigned_staff_id` 关联校验），无服务关系返回 403。
- **计分逻辑**：总分 = 各题选中 option 的 `score` 之和；等级取 `score_rule` 中第一条命中 `min <= total <= max` 的 `level` / `conclusion`。
- **等级回写**：评估保存后，将 `level` 回写 `health_records.assessment_level`（仅档案已存在时生效）。
- **量表维度**：`adl` 日常生活能力 / `barthel` 巴氏指数 / `fall` 跌倒风险 / `nutrition` 营养评估 / `cognition` 认知评估 / `pressure` 压疮风险 / `weak` 衰弱筛查 / `geriatric` 老年综合 / `self` 通用自评。

#### 阶段二：康复辅具适配

阶段二在健康档案与评估基础上，由服务人员端基于评估结果为居民生成康复辅具适配建议，用户确认后进入选购/租赁下单，形成「评估→适配→销售/租赁」闭环。推荐商品保存快照，接口共 9 个，按端分为三组。

##### C 端（前缀 /api/v1/user，需用户登录）

| 接口 | 说明 |
| --- | --- |
| `GET /fitting-recommendations` | 我的适配建议（分页倒序） |
| `GET /fitting-recommendations/:id` | 我的适配建议详情（仅本人，他人返回 403） |
| `POST /fitting-recommendations/:id/confirm` | 确认适配建议（仅草稿可确认，确认后 `status=1`） |

##### 服务人员端（前缀 /api/v1/service-staff，需服务人员登录）

| 接口 | 说明 |
| --- | --- |
| `GET /residents/:user_id/fitting-recommendations` | 客户适配建议（分页倒序，数据权限校验） |
| `POST /residents/:user_id/fitting-recommendations` | 为客户生成适配建议，请求 `{assessment_id?, symptom_desc, fitting_result, recommended_products:[{product_id, reason}]}`，推荐商品逐一校验并快照 name/sale_type，落库 `status=0`（数据权限校验） |

##### 管理端（前缀 /api/v1/merchant + RBAC）

| 接口 | 权限码 | 说明 |
| --- | --- | --- |
| `GET /fitting-recommendations` | `fitting:view` | 适配建议列表，`keyword`（用户昵称/手机号）/ `status` 筛选，分页，含用户信息 |
| `GET /fitting-recommendations/:id` | `fitting:view` | 适配建议详情，含用户信息 |
| `PUT /fitting-recommendations/:id` | `fitting:update` | 编辑适配建议（结论/推荐商品/状态/order_id，仅更新传入字段，推荐商品重新校验并快照） |
| `DELETE /fitting-recommendations/:id` | `fitting:update` | 删除适配建议 |

##### 阶段二业务规则

- **状态流转**：`status`：0 草稿 / 1 已确认 / 2 已下单；服务人员端生成即草稿，C 端确认后置 1，关联订单后置 2。
- **快照规则**：推荐商品保存快照 `{product_id, name, reason, sale_type}`，生成/编辑时逐一校验商品存在且上架（`status=1`），不满足返回"推荐商品不可用"。
- **数据权限**：服务人员仅能查看/生成自己服务过客户的适配建议（按 `orders.assigned_staff_id` 关联校验），无服务关系返回 403；C 端仅能查看/操作本人建议；管理端按 RBAC 权限码 `fitting:view` / `fitting:update` 控制。
- **RBAC 菜单**：健康服务菜单（id=8）下新增子菜单「适配建议」`/health/fitting`（id=84，`fitting:view`）+ 按钮「编辑/确认」（id=841，`fitting:update`），迁移脚本自动绑定角色 1。

#### 阶段三：居家康养照护

阶段三在阶段一、阶段二基础上，由 web-admin 为居民制定照护计划并指派服务人员，服务人员上门服务后录入照护记录（护理项完成情况、生命体征、照片、下次随访建议），形成「评估→适配→照护」基层健康服务闭环。接口共 11 个（C 端 2 个、服务人员端 3 个、管理端 6 个），按端分为三组。

##### C 端（前缀 /api/v1/user，需用户登录）

| 接口 | 说明 |
| --- | --- |
| `GET /care-plans` | 我的照护计划（分页，仅本人） |
| `GET /care-plans/:id` | 我的照护计划详情，含该计划全部上门照护记录（仅本人，他人返回 403） |

##### 服务人员端（前缀 /api/v1/service-staff，需服务人员登录）

| 接口 | 说明 |
| --- | --- |
| `GET /care-plans` | 我的照护计划（仅 `assigned_staff_id` 为当前服务人员，分页） |
| `GET /care-plans/:id` | 我的照护计划详情，含照护记录（数据权限校验，非本人指派返回 403） |
| `POST /care-visits` | 录入上门照护记录，请求 `{plan_id?, order_id?, visit_at, nursing_items:[{name,done,remark}], vitals:{blood_pressure,blood_glucose,heart_rate,oxygen,weight}, photos:[], remark, follow_up_advice}`（数据权限校验） |

##### 管理端（前缀 /api/v1/merchant + RBAC）

| 接口 | 权限码 | 说明 |
| --- | --- | --- |
| `GET /care-plans` | `care:view` | 照护计划列表，`keyword` / `plan_type` / `status` 筛选，分页，含用户信息与 `visit_count` |
| `GET /care-plans/:id` | `care:view` | 照护计划详情，含用户信息与该计划全部照护记录 |
| `POST /care-plans` | `care:create` | 新增照护计划（居民、计划类型、起止日期、频次、目标、护理项、指派服务人员、关联订单、状态） |
| `PUT /care-plans/:id` | `care:update` | 编辑照护计划（仅更新传入字段） |
| `DELETE /care-plans/:id` | `care:delete` | 删除照护计划，存在照护记录时拒绝删除（保留历史溯源） |
| `GET /care-visits` | `care:view` | 上门照护记录列表，`plan_id` / `user_id` 筛选，分页，含用户信息与录入人员姓名 |

##### 阶段三业务规则

- **数据权限**：服务人员端仅能查看 `assigned_staff_id` 为当前的照护计划；录入照护记录时 `plan_id` / `order_id` 至少提供一个，且均须属于当前服务人员负责范围（按 `care_plans.assigned_staff_id` / `orders.assigned_staff_id` 校验），`user_id` 取 plan/order 对应用户且保持一致，无服务关系返回 403。
- **删除保护**：存在照护记录（care_visits）的照护计划不可删除，保留历史溯源。
- **RBAC 菜单**：健康服务菜单（id=8）下新增子菜单「照护计划」`/health/care-plans`（id=85，`care:view`）+ 操作按钮 `care:create`（id=851，新增）`care:update`（id=852，编辑）`care:delete`（id=853，删除），迁移脚本自动绑定角色 1。
- **数据表**：新增 `care_plans` 照护计划表、`care_visits` 上门照护记录表，迁移脚本 `server/migrations/20260818120000_care_plans_visits.sql`，字段详见 `PRD.md` 4.27 / 4.28。

#### 阶段四：持续康复随访与健康宣教

阶段四在阶段一至阶段三基础上，打通「服务/租赁/评估完成后 → 随访任务 → 生命体征监测 → 健康宣教触达」的持续康复闭环：服务完成、租赁归还、评估完成自动生成随访任务，服务人员执行随访并选用宣教文章推送，C 端可查看随访、体征记录与健康宣教。接口共 21 个（C 端 5 个、服务人员端 7 个、管理端 9 个），按端分为三组。

##### C 端（前缀 /api/v1/user，需用户登录）

| 接口 | 说明 |
| --- | --- |
| `GET /follow-ups` | 我的随访任务（分页倒序，可按 `status` 筛选，仅本人） |
| `GET /health-education` | 已发布宣教文章，按本人档案慢病标签与文章 `tags` 匹配度排序（交集多者优先，相同按发布时间倒序），支持 `category` 筛选 |
| `GET /health-education/:id` | 宣教文章详情（仅已发布可见，浏览量异步 +1） |
| `GET /monitoring` | 我的生命体征记录（分页倒序，可按 `record_type` 筛选） |
| `POST /monitoring` | 自助录入体征 `{record_type, value, unit, extra, recorded_at?, remark}`，`recorded_by=0` 表示本人 |

##### 服务人员端（前缀 /api/v1/service-staff，需服务人员登录）

| 接口 | 说明 |
| --- | --- |
| `GET /follow-up-tasks` | 随访任务列表（`staff_id` 为当前 或 待认领，分页倒序，可按 `status` 筛选） |
| `GET /follow-up-tasks/:id` | 随访任务详情（须属于当前服务人员或待认领，否则 403） |
| `POST /follow-up-tasks/:id/complete` | 执行随访，请求 `{result:{contact_method, content, education_article_ids, satisfaction, remark}}`；待认领任务自动认领为当前服务人员，`status`→已完成、`completed_at`=now、`contact_method` 回写列、`result` 存 JSON |
| `POST /follow-up-tasks/:id/skip` | 跳过随访（权限校验同上，`status`→已跳过，待认领任务同样自动认领） |
| `GET /health-education` | 已发布宣教文章（供随访执行时选用，按发布时间倒序） |
| `GET /residents/:user_id/monitoring` | 客户生命体征记录（分页倒序，数据权限校验） |
| `POST /residents/:user_id/monitoring` | 为客户录入体征（数据权限校验，`recorded_by`=当前服务人员） |

##### 管理端（前缀 /api/v1/merchant + RBAC）

| 接口 | 权限码 | 说明 |
| --- | --- | --- |
| `GET /follow-up-tasks` | `followup:view` | 随访任务列表，`keyword`（用户昵称/手机号）/ `status` / `task_type` 筛选，分页，含用户与服务人员信息 |
| `GET /follow-up-tasks/:id` | `followup:view` | 随访任务详情（含用户与服务人员信息） |
| `POST /follow-up-tasks` | `followup:update` | 手动登记随访任务（`source_type` 固定为手动），`plan_follow_time` 缺省为当前+72小时 |
| `POST /follow-up-tasks/:id/complete` | `followup:update` | 管理员代执行随访（请求结构同服务人员端，不做认领） |
| `GET /monitoring` | `monitor:view` | 生命体征记录列表，`keyword`（用户昵称/手机号）/ `record_type` 筛选，分页，含用户信息 |
| `GET /health-education` | `education:view` | 宣教文章列表（含草稿），`keyword` / `category` / `status` 筛选，分页倒序 |
| `POST /health-education` | `education:create` | 新增宣教文章（标题/分类/封面/正文/定向慢病标签/状态/发布时间） |
| `PUT /health-education/:id` | `education:create` | 编辑宣教文章（仅更新传入字段） |
| `DELETE /health-education/:id` | `education:create` | 删除宣教文章 |

##### 阶段四业务规则

- **随访任务自动生成（三处触发）**：
  - 服务工单签退（服务人员端 CheckOut）完成后：租赁订单生成「租后回访」、其他服务订单生成「康复随访」，执行人=当前服务人员；
  - 租赁归还（PC 后台归还退押金）完成后：生成「租后回访」；
  - 健康评估完成（自助 `assessor_type=1` 与服务人员 `assessor_type=2` 均触发）后：生成「评估回访」。
  - 统一经 `server/internal/services/followup` 包 `CreateFollowUpTask` 生成，计划随访时间=完成/评估时点+72 小时，生成失败仅记录日志、不影响业务主流程。
- **任务状态**：`task_type`：1 康复随访 / 2 租后回访 / 3 慢病随访 / 4 评估回访；`source_type`：1 服务完成 / 2 租赁归还 / 3 评估完成 / 4 手动；`status`：0 待执行 / 1 已完成 / 2 已跳过；`contact_method`：1 电话 / 2 上门 / 3 微信。
- **待认领与自动认领**：`staff_id` 为空表示待认领，服务人员端可见并可执行，执行/跳过时自动认领为当前服务人员；C 端与管理端按各自可见范围操作。
- **宣教定向**：慢病标签维护在健康档案 `health_records.chronic_tags`；C 端宣教列表按本人慢病标签与文章 `tags` 匹配度排序，仅已发布文章对 C 端/服务人员端可见，草稿仅管理端可见。
- **生命体征**：`record_type`：1 血压 / 2 血糖 / 3 心率 / 4 血氧 / 5 体重；录入支持用户自助（`recorded_by=0`）与服务人员上门（`recorded_by=服务人员ID`），记录时间 `recorded_at` 用 RFC3339 带时区。
- **数据权限**：服务人员仅能随访（执行本人或待认领任务）与查看/录入本人服务过客户的体征（按 `orders.assigned_staff_id` 关联校验），无服务关系返回 403。
- **RBAC 菜单**：健康服务菜单（id=8）下新增子菜单「随访任务」`/health/follow-ups`（id=86，`followup:view`）+ 按钮「执行/登记」（id=861，`followup:update`）；「生命体征」`/health/monitoring`（id=87，`monitor:view`）+ 按钮「录入」（id=871，`monitor:create`）；「健康宣教」`/health/education`（id=88，`education:view`）+ 按钮「新增/编辑」（id=881，`education:create`），迁移脚本自动绑定角色 1。
- **数据表**：新增 `follow_up_tasks` 随访任务表、`health_monitoring` 生命体征监测表、`health_education_articles` 健康宣教内容表，迁移脚本 `server/migrations/20260818130000_followup_monitoring_education.sql`，字段详见 `PRD.md` 4.29 / 4.30 / 4.31。

## 3. 返回结构与空值约定

### 3.1 列表返回

- 列表字段统一使用空数组 `[]` 兜底。
- 若带分页，至少保证：
  - `list: []`
  - `pagination.total`
  - `pagination.page`
  - `pagination.page_size`

### 3.2 统计返回

- 统计对象无数据时返回结构化对象，不直接返回 `null`。
- 数值字段默认返回 `0`。

### 3.3 图表返回

- 趋势、排行、分布类图表字段必须返回数组。
- 维度切换接口无数据时返回空数组并由前端展示空态。

## 4. 当前重点同步要求

1. 功能改动涉及接口时，必须同时更新：
   - 根 `PRD.md`
   - 本文件对应章节
2. 若字段口径调整，必须同步更新：
   - 请求参数
   - 响应示例
   - 空值约定
3. 若只是 UI 调整但不改协议，可只更新功能说明文档，不必改本文件。

## 5. 与根 PRD 的对应关系

- 认证相关接口：对应 `PRD.md` `3.1`
- 商户PC后台接口：对应 `PRD.md` `3.2`
- 商户管理接口：对应 `PRD.md` `3.3`
- 商品分类 / 商品接口：对应 `PRD.md` `3.4`、`3.5`
- 订单接口：对应 `PRD.md` `3.6`
- 数据分析接口：对应 `PRD.md` `3.7`
- C 端接口：对应 `PRD.md` `3.8`
- 微信支付接口：对应 `PRD.md` `3.9`
- 服务人员接口：对应 `PRD.md` 新增章节
- 健康服务接口：对应 `PRD.md` 新增 `3.11` 章节
