# 分账逻辑代码定位计划

## Summary

* 目标：梳理当前项目中“分账配置、自动分账执行、分账记录查询、前端展示”的完整代码入口，方便后续逐段审查逻辑是否符合预期。

* 本次不做代码修改，只输出实际文件位置、职责分工和建议检查顺序。

## Current State Analysis

### 1. 分账配置入口

* 后端服务商维护商家支付/分账配置：

  * `server/internal/handlers/sp/merchant_management.go`

  * 关键函数：

    * `CreateMerchant()`

    * `UpdateMerchantPaymentConfig()`

    * `buildPaymentConfigStatus()`

* 当前配置字段：

  * `sub_mch_id`

  * `profit_sharing_enabled`

  * `profit_sharing_ratio`

  * `payment_config_status`

* 这里的职责是：

  * 服务商创建商家时写入分账开关和比例

  * 服务商后续单独更新商家支付配置与分账比例

  * 根据 `sub_mch_id` 与分账比例计算 `payment_config_status`

### 2. 分账自动执行入口

* 核心自动分账逻辑在：

  * `server/internal/handlers/sp/payment.go`

* 关键调用链：

  * `PaymentNotify()`

  * `processPaymentSuccess()`

  * `createSkippedProfitSharingRecord()`

  * `syncOrderProfitSharingFromRecord()`

* 当前触发时机：

  * 微信支付回调到达

  * `trade_state == SUCCESS`

  * 订单支付状态写入成功后，立即进入分账逻辑

* 当前分账决策分支：

  * 已有成功/跳过状态：直接返回，避免重复分账

  * 已存在订单维度分账记录：复用记录，同步回订单字段

  * 未开启分账或比例 <= 0：写“已跳过”记录

  * 分账金额四舍五入后为 0：写“已跳过”记录

  * 未配置 `sub_mch_id`：写“已跳过”记录

  * 其余情况：调用微信分账接口发起真实分账

### 3. 微信分账接口封装

* 微信服务商分账客户端在：

  * `server/internal/services/wechatpay/service_provider.go`

* 关键类型与函数：

  * `type ProfitSharingReceiver`

  * `type ProfitSharingRequest`

  * `type ProfitSharingResponse`

  * `CreateProfitSharingOrder()`

* 当前真实调用微信接口：

  * `POST /v3/profitsharing/orders`

* 这里主要负责：

  * 组装微信分账请求

  * 发起请求

  * 返回微信分账结果中的 `order_id / state`

### 4. 当前分账接收方是谁

* 当前代码里分账接收方固定写在：

  * `server/internal/handlers/sp/payment.go`

* 实际接收方定义：

  * `Type = "MERCHANT_ID"`

  * `Account = client.GetSPMchID()`

  * `Description = "服务商抽佣"`

* `client.GetSPMchID()` 的来源：

  * `server/internal/services/wechatpay/service_provider.go`

  * 返回 `config.Config.WechatPay.SPMchID`

* 当前环境里的服务商商户号配置来源：

  * `server/.env`

  * `WECHAT_PAY_SP_MCH_ID=1745555658`

* 结论：

  * 当前项目的分账接收方不是商家 `sub_mch_id`

  * 也不是独立配置的分账接收账号

  * 而是服务商自己的微信商户号 `SPMchID`

### 5. 分账相关数据模型

* 数据模型定义在：

  * `server/internal/models/models.go`

* 商家配置相关字段：

  * `Merchant.ProfitSharingEnabled`

  * `Merchant.ProfitSharingRatio`

  * `Merchant.PaymentConfigStatus`

* 订单上的分账状态字段：

  * `Order.ProfitSharingStatus`

  * `Order.ProfitSharingAmount`

  * `Order.ProfitSharingOrderNo`

  * `Order.ProfitSharingAt`

  * `Order.ProfitSharingError`

* 分账历史表模型：

  * `type MerchantProfitSharingRecord`

### 6. 分账相关数据库表结构与初始化数据

* 表结构与字段定义在：

  * `server/migrations/20240101000000_full_init.sql`

* 重点表/字段：

  * `merchants.profit_sharing_enabled`

  * `merchants.profit_sharing_ratio`

  * `orders.profit_sharing_status`

  * `orders.profit_sharing_amount`

  * `orders.profit_sharing_order_no`

  * `orders.profit_sharing_at`

  * `orders.profit_sharing_error`

  * `merchant_profit_sharing_records`

* 模拟数据里已有成功/失败/跳过样例：

  * `server/migrations/20260511120000_mock_seed.sql`

### 7. 分账记录查询接口

* 路由入口在：

  * `server/cmd/server/main.go`

* 服务商查询分账记录：

  * `GET /api/v1/sp/profit-sharing-records`

  * 处理函数：`sp.GetProfitSharingRecords()`

* 商家查询本店分账记录：

  * `GET /api/v1/merchant/profit-sharing-records`

  * 处理函数：`sp.GetMerchantProfitSharingRecords()`

* 两个查询函数都在：

  * `server/internal/handlers/sp/merchant_management.go`

### 8. 前端分账页面与接口

* 前端 API 封装在：

  * `miniprogram/src/api/index.ts`

* 服务商分账记录查询：

  * `getSpProfitSharingRecords()`

* 商家分账记录查询：

  * `getMerchantProfitSharingRecords()`

* 服务商分账历史页：

  * `miniprogram/src/pages/sp/settlements/history.vue`

* 商家分账历史页：

  * `miniprogram/src/pages/merchant/settlements/history.vue`

* 服务商商家管理页也会展示分账配置：

  * `miniprogram/src/pages/sp/merchants/list.vue`

  * `miniprogram/src/pages/sp/merchants/detail.vue`

  * `miniprogram/src/pages/sp/merchants/edit.vue`

## Proposed Changes

### 1. 先按“配置 -> 执行 -> 记录 -> 展示”顺序检查

* 第一层：商家分账配置是否合法

  * `sp/merchant_management.go`

  * `models.go`

* 第二层：支付成功后是否真的触发分账

  * `sp/payment.go`

  * `service_provider.go`

* 第三层：订单字段与分账记录表是否一致

  * `models.go`

  * `20240101000000_full_init.sql`

  * `20260511120000_mock_seed.sql`

* 第四层：服务商/商家页面是否只是展示接口结果，还是又做了二次加工

  * `api/index.ts`

  * `sp/settlements/history.vue`

  * `merchant/settlements/history.vue`

### 2. 建议重点检查的逻辑点

* `profit_sharing_status` 的状态枚举含义是否统一

* 分账金额计算：

  * `order.PayAmount * ratio / 100`

  * `roundAmount()`

* 跳过原因是否符合业务预期：

  * 未开启分账

  * 分账金额为 0

  * 未配置子商户号

* 重复支付回调是否会重复发起分账

* 分账失败时：

  * 订单字段是否正确写入失败原因

  * 分账记录表是否和订单状态一致

* 服务商查询与商家查询是否读取的是同一张 `merchant_profit_sharing_records`

### 3. 如果后续要深入排查，建议直接看的代码片段

* 自动分账主逻辑：

  * `server/internal/handlers/sp/payment.go`

* 微信分账请求封装：

  * `server/internal/services/wechatpay/service_provider.go`

* 商家配置分账开关和比例：

  * `server/internal/handlers/sp/merchant_management.go`

* 分账记录与订单字段定义：

  * `server/internal/models/models.go`

* 服务商/商家分账历史展示：

  * `miniprogram/src/pages/sp/settlements/history.vue`

  * `miniprogram/src/pages/merchant/settlements/history.vue`

## Assumptions & Decisions

* 当前项目中的“分账”是支付成功回调后的自动处理逻辑，不是前端主动触发。

* 当前前端页面主要负责查询和展示分账结果，核心业务判断在后端。

* 当前服务商和商家分账历史查询复用同一张分账记录表，区别只在查询维度不同。

## Verification Steps

* 核对 `server/cmd/server/main.go` 中分账相关路由是否与实际页面/API 一致

* 核对 `sp/payment.go` 中支付成功后是否唯一触发分账

* 核对 `service_provider.go` 是否只封装了微信分账调用，不含额外业务判断

* 核对 `models.go`、SQL 初始化文件、模拟数据文件三处的分账字段是否一致

* 核对前端 `api/index.ts` 与两个分账历史页面是否只是透传并展示后端数据
