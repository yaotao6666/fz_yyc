# 支付、退款、分账全链路测试计划

## Summary

- 目标：验证当前项目中“支付、退款、分账”三条核心业务链路在真实联调环境下是否正常，并输出问题清单与定位依据。
- 交付口径：本轮以“先测后列问题”为主，不在计划阶段直接包含修复执行。
- 测试优先级：
  - P0：真实支付下单、支付回调、订单状态变更
  - P0：商家发起退款、退款回调、订单/退款状态变更
  - P0：支付成功后的自动分账、服务商与商家分账历史可查
  - P1：异常分支与跳过/失败场景验收

## Current State Analysis

### 1. 当前支付链路是“真实微信支付服务商模式”

- 下单支付入口在 `server/internal/handlers/user/handler.go`
  - 创建订单后，若 `payAmount > 0`，会调用 `wechatpay.NewServiceProviderClient()` 和 `CreatePartnerJSAPIPayOrder()`。
  - 关键校验：
    - 商家必须配置 `sub_mch_id`
    - `payment_config_status` 必须为 `1`
    - 用户必须存在 `openid`
    - 服务商支付回调地址 `config.Config.WechatPay.CallbackURL` 必须已配置

- 微信支付服务商客户端实现位于 `server/internal/services/wechatpay/service_provider.go`
  - 真实调用微信接口：
    - `POST /v3/pay/partner/transactions/jsapi`
  - 配置要求：
    - `SPMchID`
    - `APIV3Key`
    - `CertSerialNo`
    - `PrivateKey`
    - 可选的微信支付平台公钥 `PublicKey`

### 2. 当前支付回调已直接串到订单状态更新与自动分账

- 支付/退款回调统一入口：
  - `server/cmd/server/main.go`
    - `POST /api/v1/notify/payment`
    - `POST /api/v1/callbacks/payment/wechat`

- 支付回调处理在 `server/internal/handlers/sp/payment.go`
  - `PaymentNotify()` 解析并验签微信回调
  - `processPaymentSuccess()` 会：
    - 根据 `order_no` 查订单
    - 写入 `transaction_id`
    - 更新订单为 `status = 2`
    - 写入 `paid_at`
    - 自动进入分账逻辑

### 3. 当前分账逻辑是“支付成功后立即尝试分账或记录跳过/失败”

- 分账逻辑位于 `server/internal/handlers/sp/payment.go`
  - 会读取商家：
    - `profit_sharing_enabled`
    - `profit_sharing_ratio`
    - `sub_mch_id`
  - 分支行为：
    - 未开启分账或比例 <= 0：写入“跳过”记录
    - 分账金额四舍五入后为 0：写入“跳过”记录
    - 未配置 `sub_mch_id`：写入“跳过”记录
    - 正常情况：调用 `CreateProfitSharingOrder()`
      - 微信接口：`POST /v3/profitsharing/orders`
  - 会同时更新：
    - `orders.profit_sharing_status`
    - `orders.profit_sharing_amount`
    - `orders.profit_sharing_order_no`
    - `orders.profit_sharing_at`
    - `orders.profit_sharing_error`
    - `merchant_profit_sharing_records`

- 服务商与商家都已有查询入口：
  - 服务商：
    - `GET /api/v1/sp/profit-sharing-records`
    - `GET /api/v1/sp/orders/refunds`
  - 商家：
    - `GET /api/v1/merchant/profit-sharing-records`

### 4. 当前退款链路由商家端发起，用户端已改为“联系商家退款”

- 用户端接口虽然仍存在：
  - `POST /api/v1/user/orders/:order_id/refund`
  - 但小程序 C 端页面 `pages/store/my-orders.vue` 和 `pages/store/order-detail.vue` 已改为“联系商家退款”，不再作为本轮真实退款主链路入口。

- 商家退款主链路在 `server/internal/handlers/merchant/order.go`
  - `POST /api/v1/merchant/orders/:order_id/refund`
  - 当前关键行为：
    - 仅允许订单状态 `2/3/5`
    - 必须已有 `transaction_id`，否则拦截“订单尚未完成支付回调”
    - `refund_amount <= 0` 时回落到订单实付金额
    - 调用微信：
      - `POST /v3/refund/domestic/refunds`
    - 根据接口同步返回：
      - `SUCCESS` -> 订单直接更新为 `status = 6`
      - 其他处理中状态 -> 订单更新为 `status = 5`
      - `CLOSED/ABNORMAL` -> 退款记录失败

- 退款回调处理在 `server/internal/handlers/sp/payment.go`
  - `processRefundNotify()` 根据 `refund_no` 更新：
    - `refunds.status`
    - `refunds.refunded_at`
    - `orders.status`
    - `orders.refunded_at`

### 5. 现有脚本和文档只能覆盖部分前置验证，不能替代真实联调

- `server/scripts/regression-test.sh`
  - 当前只覆盖：
    - 服务商登录
    - 商家登录
    - 获取商家列表
    - 获取商品列表
  - 不能覆盖真实微信支付、退款、分账回调。

- `docs/prd/PRD-测试与附录.md`
  - 已明确真实联调前置条件：
    - 服务商 V3 Key、平台公钥证书、公钥、回调地址已配置
    - 至少存在一个已进件商家子商户号，例如 `1112649854`
  - 已明确验收目标：
    - 不同商家拉起各自 `sub_mch_id`
    - 未完成支付配置的商家被拦截
    - 支付回调后订单状态正确
    - 分账成功/失败/跳过三类记录都可落库

## Proposed Changes

### 1. 先做真实联调环境基线检查

#### 目标

- 在正式执行支付、退款、分账测试前，确认当前环境具备真实联调条件，避免把“环境未配齐”误判成业务 bug。

#### 检查范围

- 后端运行环境
  - `server/.env`
  - 实际 API 运行容器/进程读取到的微信支付配置
- 核心配置项
  - 服务商商户号
  - `APIV3Key`
  - 证书序列号
  - 服务商私钥
  - 微信支付平台公钥
  - 支付/退款回调地址
- 商家支付配置
  - 至少一个商家具备：
    - `sub_mch_id`
    - `payment_config_status = 1`
  - 推荐使用 PRD 约定样例：
    - `1112649854`

#### 输出

- 一份“联调基线通过/失败”结果表。
- 若失败，记录缺失项属于：
  - 环境配置问题
  - 商家配置问题
  - 回调地址不可达问题

### 2. 支付主链路测试

#### 覆盖目标

- 用户扫码进店 -> 下单 -> 拉起微信支付 -> 支付成功回调 -> 订单变为已支付。

#### 执行对象

- 小程序页面：
  - `miniprogram/src/pages/store/home.vue`
  - `miniprogram/src/pages/store/confirm.vue`
  - `miniprogram/src/pages/store/my-orders.vue`
  - `miniprogram/src/pages/store/order-detail.vue`

#### 关键验证点

- 使用“已完成支付配置”的商家：
  - 下单接口成功返回 `pay_params`
  - `pay_params.prepay_id` 非空
  - 真机/开发者工具能够拉起支付
- 支付成功后：
  - 支付回调接口返回成功
  - `orders.status` 从 `1` 变为 `2`
  - `orders.transaction_id` 写入
  - `orders.paid_at` 写入
  - 小程序订单列表与详情页显示“已支付”
- 使用“未完成支付配置”的商家：
  - 下单被拦截
  - 报错符合当前实现：`商家支付配置未完成，请联系服务商`

#### 证据来源

- 小程序页面实际表现
- `POST /api/v1/store/:merchant_id/orders` 返回
- API 日志
- `orders` 表字段变化
- 回调接口访问日志

### 3. 退款主链路测试

#### 覆盖目标

- 商家在订单管理中对已支付/已完成订单发起退款，验证同步返回、异步回调和状态流转。

#### 执行对象

- 小程序页面：
  - `miniprogram/src/pages/merchant/orders/list.vue`

#### 关键验证点

- 退款前置条件：
  - 订单必须已有 `transaction_id`
  - 订单状态必须为 `2/3/5`
- 商家发起退款后：
  - `refunds` 表新增或复用记录
  - `refund_amount` 正确
  - 未传金额时按 `order.pay_amount` 回落
- 微信退款接口同步返回分支：
  - `SUCCESS`：
    - `refunds.status = 2`
    - `orders.status = 6`
    - `refunded_at` 正确
  - 处理中：
    - `refunds.status = 1`
    - `orders.status = 5`
  - 失败：
    - `refunds.status = 3`
    - 页面提示错误
- 退款回调到达后：
  - `processRefundNotify()` 正常更新最终状态
  - 订单由 `5` 变 `6`

#### 证据来源

- 商家小程序页面状态
- `refunds` / `orders` 数据表
- 微信退款回调日志
- 服务商退款列表接口：
  - `GET /api/v1/sp/orders/refunds`

### 4. 分账主链路测试

#### 覆盖目标

- 支付成功后自动触发分账，验证成功、失败、跳过三类记录的完整闭环。

#### 执行对象

- 后端自动逻辑：
  - `server/internal/handlers/sp/payment.go`
- 服务商页面与商家页面：
  - 服务商分账记录页
  - 商家分账历史页

#### 关键验证点

- 成功场景
  - 已配置 `sub_mch_id`
  - 已开启分账
  - `profit_sharing_ratio > 0`
  - 支付成功后：
    - `orders.profit_sharing_status = 1`
    - `profit_sharing_amount` 正确
    - `profit_sharing_order_no` 回填
    - `merchant_profit_sharing_records.status = 1`
    - 服务商/商家端均可查询

- 跳过场景
  - 未开启分账
  - 比例 <= 0
  - 分账金额四舍五入后为 0
  - 无 `sub_mch_id`
  - 预期：
    - `orders.profit_sharing_status = 3`
    - `merchant_profit_sharing_records.status = 3`
    - `error_message` 为对应跳过原因

- 失败场景
  - 微信分账接口报错
  - 预期：
    - `orders.profit_sharing_status = 2`
    - `orders.profit_sharing_error` 写入
    - `merchant_profit_sharing_records.status = 2`
    - `error_message` 写入

#### 证据来源

- `merchant_profit_sharing_records` 表
- `orders` 表的分账字段
- 服务商分账记录接口
- 商家分账历史接口
- API 日志中分账调用与报错信息

### 5. 设计一套“真实联调优先”的测试顺序

#### 顺序

1. 环境基线检查
2. 服务商确认商家支付配置
3. 用户真实下单支付
4. 验证支付回调与订单状态
5. 验证支付成功后的分账结果
6. 商家对同一订单发起退款
7. 验证退款同步返回与异步回调
8. 验证退款后订单状态与退款列表
9. 汇总异常点并按链路归类

#### 归类方式

- 环境问题
- 支付下单问题
- 支付回调问题
- 退款调用问题
- 退款回调问题
- 分账计算/落库问题
- 页面展示问题

### 6. 形成“异常清单模板”，作为执行后输出

#### 每条异常需包含

- 场景名称
- 复现步骤
- 实际结果
- 预期结果
- 影响范围
- 初步根因位置
  - 文件
  - 接口
  - 表字段
- 是否阻塞后续链路
- 建议修复优先级

## Assumptions & Decisions

- 决策 1：本轮按“真实联调优先”规划，不以纯脚本或纯数据库验证替代真实支付链路。
- 决策 2：本轮交付目标是“验证 + 列问题”，不是直接实施修复。
- 决策 3：退款主验证入口按当前产品口径使用商家端发起退款；C 端“联系商家退款”不作为主测试入口。
- 决策 4：支付与退款都以当前微信服务商模式实现为准，不调整接口协议。
- 决策 5：分账以“支付成功后立即尝试”的现有实现为准，重点验证成功/失败/跳过三类落库结果。

## Verification Steps

### 1. 配置与数据前置检查

- 检查 API 实际运行环境中的微信支付配置是否完整。
- 检查至少一个商家 `sub_mch_id` 已配置且 `payment_config_status = 1`。
- 检查回调地址对微信侧可达。

### 2. 支付验证

- 用已配置商家完成一次真实下单支付。
- 验证：
  - 小程序拉起支付成功
  - 回调成功
  - `orders.status = 2`
  - `transaction_id`、`paid_at` 正确

### 3. 退款验证

- 用商家端对同一订单发起退款。
- 验证：
  - `refunds` 记录生成
  - 退款同步返回状态正确
  - 回调后 `orders.status = 6`
  - `refunded_at` 正确

### 4. 分账验证

- 对支付成功订单检查：
  - `orders.profit_sharing_*` 字段
  - `merchant_profit_sharing_records`
- 在服务商端与商家端确认分账历史均可见。

### 5. 失败与跳过验证

- 用未配置支付/分账的商家分别验证：
  - 支付被拦截
  - 分账跳过原因写入
- 如真实环境可制造分账失败，则验证失败记录写入。

### 6. 结果整理

- 按“支付 / 退款 / 分账 / 页面展示 / 环境配置”五类输出结果。
- 每类区分：
  - 通过
  - 阻塞
  - 有风险但不阻塞
