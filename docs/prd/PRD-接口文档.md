# PRD-接口文档

## 1. 文档定位

本文件用于承接 `PRD.md` 中与“接口、返回结构、字段口径、空值约定、WebSocket 协议”相关的内容。

读取顺序：

1. 先读根目录 `PRD.md`
2. 再读本文件
3. 如需功能背景，联读 `docs/prd/PRD-功能说明.md`

## 2. 接口分组

### 2.1 认证相关接口

- 商家账号密码登录
- 商家微信快捷登录
- 服务商登录 / 退出
- C 端微信登录

关键约束：

- 登录态字段、缓存字段、角色标识必须在前后端统一。
- 商家退出登录后，前端需要主动断开 WebSocket 连接。

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
  - `pages/store/home`、`pages/store/product`、`pages/store/confirm` 统一支持从 `merchant_id` 或 `scene` 解析商家入口参数
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

#### 服务商公告接口说明

- 服务商公告管理页面已从当前小程序下线，现阶段服务商能力由 `web-admin/` 承接。
- `web-admin/` 已接入以下 `/api/v1/sp/announcements*` 接口，用于公告列表、详情、新增、编辑和删除（停用）管理。
- 商家端首页公告展示仍可继续读取已发布公告。

### 2.2 服务商接口

- 当前调用端：`web-admin/` 服务商 Web/PC 后台。
- 仪表盘
- 创建商家：`POST /api/v1/sp/merchants`
- 商家列表
- 商家详情
- 更新商家资料：`PUT /api/v1/sp/merchants/:merchant_id`
- 更新支付配置：`PUT /api/v1/sp/merchants/:merchant_id/payment-config`
- 商家图片资产更新：`PUT /api/v1/sp/merchants/:merchant_id/assets`
- 订单列表：`GET /api/v1/sp/orders`
- 订单详情：`GET /api/v1/sp/orders/:order_id`
- 分账历史：`GET /api/v1/sp/profit-sharing-records`
- 商家费率
- 商家二维码
- 系统公告
- 服务商数据分析

当前重点约束：

- 服务商数据分析接口需按以下维度输出：
  - 商家访问率
  - 商家下单率
  - 商家下单金额
  - 商家下单均价
  - 日 / 周 / 月 / 年订单量
  - 商家排行维度切换
- `GET /api/v1/sp/merchants/analytics/distribution` 返回 `merchants + totals` 结构。
- `GET /api/v1/sp/orders` 需支持 `merchant_id`、`status`、`start_date`、`end_date`、`keyword`、`page`、`page_size` 筛选，并且仅返回当前服务商名下商家的订单。
- `GET /api/v1/sp/orders/:order_id` 需校验订单归属当前服务商，返回结构与商家侧订单详情保持一致，包含商品明细、用户信息、配送信息、支付单号、核销码、核销时间与核销人。
- `GET /api/v1/sp/orders/analytics` 返回 `day/week/month/year` 四组订单量桶。
- `GET /api/v1/sp/amount/top-merchants` 支持 `metric` 参数切换排行维度。
- `/api/v1/sp/announcements*` 服务商公告接口当前由 `web-admin/` 直接调用，小程序端不再承载该页面。

### 2.3 商家管理接口

- 商家资料查询 / 更新
- 商家设置
- 商家满减规则：`GET/PUT /api/v1/merchant/full-reduction-rules`
- 员工绑定微信 / 解绑微信
- 商家营业状态
- 配送设置
- 商家二维码
- 商家打印机管理：`GET/POST/PUT/DELETE /api/v1/merchant/printers*`
- 打印测试：`POST /api/v1/merchant/printers/:printer_id/test`
- 商家分账历史：`GET /api/v1/merchant/profit-sharing-records`

当前重点约束：

- 商家资料更新需支持 `logo` 与背景图字段。
- `GET /api/v1/merchant/settings` 需返回 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled`，`PUT /api/v1/merchant/settings` 需支持更新这三个开关。
- `GET /api/v1/store/:merchant_id/delivery-rules` 除配送费结构外，还需返回 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled` 供确认页动态展示下单方式。
- `GET /api/v1/merchant/full-reduction-rules` 返回 `rules` 与 `active_rules`，单档规则至少包含 `threshold_amount`、`discount_amount`、`status`、`sort`。
- `PUT /api/v1/merchant/full-reduction-rules` 最多支持 5 档规则，`discount_amount` 必须小于 `threshold_amount`。
- `GET /api/v1/store/:merchant_id/full-reduction-rules` 为公开接口，仅返回当前商家启用中的满减规则。
- `delivery_settings.enabled` 只表示配送费规则是否生效，确认页是否展示“配送”必须以后端返回的 `takeout_enabled` 为准。
- `GET /api/v1/merchant/printers` 返回打印机列表时，需返回 `type`、`status`、`auto_print`、`is_default`、`print_count`、`last_print_at`、`has_api_key`、`has_feie_ukey`。
- 飞鹅打印机请求字段至少包含 `feie_user`、`feie_ukey`、`feie_sn`。
- `GET /api/v1/merchant/qrcode` 必须固定生成指向 `pages/store/home` 的小程序码，且 `scene` 需使用 `merchant_id={当前商家ID}` 以兼容现有店铺入口解析逻辑。
- `GET /api/v1/sp/merchants/{id}/qrcode` 作为服务商代查看商家二维码接口，必须与商家侧二维码保持一致，通过微信小程序码接口生成真实二维码，并返回可直接展示的二维码图片、`pages/store/home` 页面路径以及 `merchant_id={商家ID}` 的 `scene` 参数；若微信生成失败，应直接返回错误，不能回退为系统自绘占位二维码。
- 服务商若代商家维护资料，必须通过明确的服务商侧管理接口或授权更新接口。
- 商家分账历史至少返回：
  - `profit_sharing_date`
  - `order_no`
  - `pay_amount`
  - `profit_sharing_ratio`
  - `profit_sharing_amount`
  - `merchant_received_amount`
  - `status`
  - `error_message`

### 2.4 商品与分类接口

- 商品分类列表 / 新增 / 修改 / 删除
- 商品列表 / 详情 / 新增 / 更新 / 删除
- 商品规格相关接口

当前重点约束：

- 商品图片字段以当前接口返回结构为准。
- 商品相关列表接口返回空值时，数组字段统一返回 `[]`。

### 2.5 订单接口

- 商家订单列表
- 商家订单详情
- 当前订单核销
- 快速核销：`POST /api/v1/merchant/orders/quick-complete`
- 商家退款
- 订单统计
- C 端订单列表 / 详情 / 取消 / 申请退款

当前重点约束：

- 商家订单详情应返回：
  - `verify_code`
  - `completed_at`
  - `completed_by_name`
- 商家订单列表接口支持：
  - `status`
  - `start_date`
  - `end_date`
- 商家对当前订单执行核销时，不再要求重复输入核销码。
- 商家退款接口需与前端请求参数保持一致，至少明确：
  - `reason`
  - `refund_amount`
- 商家退款接口在未传 `refund_amount` 或传入 `<= 0` 时，默认按订单实付金额处理。
- C 端创建订单接口需按 `delivery_type` 与三字段一一校验：
  - `delivery_type=1` 校验 `takeout_enabled`
  - `delivery_type=2` 校验 `dine_in_enabled`
  - `delivery_type=3` 校验 `pickup_enabled`
- 当对应方式未开启时，接口分别返回“商家暂未开启配送 / 堂食 / 自提”。
- C 端确认订单页展示的满减优惠仅作提示，下单结果必须以后端返回的 `discount_amount`、`pay_amount` 为准。
- 创建订单接口需按商家启用中的满减规则重算：
  - `discount_amount`
  - `pay_amount`
- **退款状态口径**：
  - 订单 `status=5`：退款中（已发起退款流程；接口会先主动同步一次微信退款状态，未拿到最终结果时继续等待微信退款结果）
  - 订单 `status=6`：已退款（接口主动同步或收到微信退款成功回调后写入）
  - `refunded_at`：在订单进入 `status=6` 时写入
- 支付成功后，订单需回写：
  - `transaction_id`
  - `pay_notify_payload`
  - `profit_sharing_status`
  - `profit_sharing_amount`
  - `profit_sharing_order_no`
  - `profit_sharing_at`
  - `profit_sharing_error`
- 支付成功并完成订单状态回写后，后端向商家 WebSocket 推送 `order_notify`。

### 2.6 支付与分账接口

- C 端下单支付由服务商模式统一拉起。
- 全局配置 `WECHAT_PAY_APP_MODE` 控制当前部署版本的小程序支付身份：
  - `sp_app`：请求微信支付时使用 `sp_appid + payer.sp_openid`
  - `sub_app`：请求微信支付时继续必传 `sp_appid`，同时使用 `sub_appid + payer.sub_openid`
- 商家支付配置字段：
  - `sub_mch_id`
  - `profit_sharing_enabled`
  - `profit_sharing_ratio`
  - `payment_config_status`
- 微信支付回调入口：
  - `/api/v1/notify/payment`
  - `/api/v1/callback/wechat`
  - 支付与退款均使用同一回调入口，按 `event_type` 区分支付（TRANSACTION）与退款（REFUND）

当前重点约束：

- 下单时必须按商家维度读取 `sub_mch_id`。
- 已进件商家仅通过服务商后台回填 `sub_mch_id` 完成支付配置，本轮联调样例为 `1112649854`。
- `WECHAT_PAY_APP_MODE` 仅切换登录与下单支付时使用的小程序身份，不改变服务商支付回调、退款与分账实现。
- 支付成功与分账成功不是同一状态，前后端需分别展示。
- 分账历史需覆盖成功 / 失败 / 跳过三类状态。

### 2.7 数据分析接口

#### 商家端

- 今日概览
- 订单趋势
- 客户分析
- 商品排行
- 库存预警

#### 服务商端

- 商家转化分析
- 商家排行榜
- 周期订单分析

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

- 商家 WebSocket：`/api/v1/ws/merchant`
- 开发联调接口：
  - 订单提醒下发：`POST /api/v1/dev/order-notify`
  - 用户进店提醒下发：`POST /api/v1/dev/store-visit-notify`

当前重点约束：

- 商家登录成功后建立连接。
- 商家退出登录时主动断开。
- 商家登录失效触发 401 时，也要收口前端登录态并主动断开。
- `order_notify` 用于商家支付成功后的新订单提醒；`store_visit_notify` 用于顾客进店提醒。
- 非开发环境不应暴露不必要的联调接口。

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
- 服务商接口：对应 `PRD.md` `3.2`
- 商家管理接口：对应 `PRD.md` `3.3`
- 商品分类 / 商品接口：对应 `PRD.md` `3.4`、`3.5`
- 订单接口：对应 `PRD.md` `3.6`
- 数据分析接口：对应 `PRD.md` `3.7`
- C 端接口：对应 `PRD.md` `3.8`
- 微信支付接口：对应 `PRD.md` `3.9`
