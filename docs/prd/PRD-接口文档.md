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
  - `Authorization: Bearer {user_token}` 用于 `POST /api/v1/store/:merchant_id/orders`（下单）
  - `/api/v1/store/:merchant_id/home`、`/products`、`/products/:product_id`、`/delivery-rules` 为公开接口，不要求登录
  - `pages/home/index`、`pages/store/product`、`pages/store/confirm` 统一支持从 `merchant_id` 或 `scene` 解析商户入口参数
  - `GET /api/v1/store/:merchant_id/home` 与 `GET /api/v1/store/:merchant_id/delivery-rules` 进入页面后可直接发起，不等待登录完成

#### C 端访问与埋点

- **访问埋点**：`POST /api/v1/store/:merchant_id/visit`
  - 请求：`{ openid: string, source?: string }`
- **行为埋点**：`POST /api/v1/store/:merchant_id/event`
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
- `GET /api/v1/store/:merchant_id/delivery-rules` 除配送费结构外，还需返回 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled` 供确认页动态展示下单方式。
- `GET /api/v1/merchant/full-reduction-rules` 返回 `rules` 与 `active_rules`，单档规则至少包含 `threshold_amount`、`discount_amount`、`status`、`sort`。
- `PUT /api/v1/merchant/full-reduction-rules` 最多支持 5 档规则，`discount_amount` 必须小于 `threshold_amount`。
- `GET /api/v1/store/:merchant_id/full-reduction-rules` 为公开接口，仅返回当前商户启用中的满减规则。
- `delivery_settings.enabled` 只表示配送费规则是否生效，确认页是否展示“配送”必须以后端返回的 `takeout_enabled` 为准。
- `GET /api/v1/merchant/printers` 返回打印机列表时，需返回 `type`、`status`、`auto_print`、`is_default`、`print_count`、`last_print_at`、`has_api_key`、`has_feie_ukey`。
- 飞鹅打印机请求字段至少包含 `feie_user`、`feie_ukey`、`feie_sn`。
- `GET /api/v1/merchant/qrcode` 必须固定生成指向 `pages/home/index` 的小程序码，且 `scene` 需使用 `merchant_id={当前商户ID}` 以兼容现有商城入口解析逻辑。
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
- **归还租赁商品（退押金）**：`POST /api/v1/sp/merchants/{merchant_id}/orders/{order_id}/return`
- C 端订单列表 / 详情 / 取消 / 申请退款

当前重点约束：

- 商户订单详情应返回：
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
- 微信支付服务商凭证为系统级配置，由系统统一维护，不在商户维度暴露。
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

- 注册申请：`POST /api/v1/service-staff/register`
  - 请求：包含服务人员基本信息
  - 提交后进入待审核状态，由商户管理员在 PC 后台审核
- 账号密码登录：`POST /api/v1/service-staff/login`
- 微信快捷登录：`POST /api/v1/service-staff/wechat-login`
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
