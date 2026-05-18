# 用户端退款拨号与商家下单方式开关优化计划

## Summary

* 将用户端订单详情页“联系商家退款”从当前的确认弹窗后二次拨号，改成点击后直接拨打商家电话。

* 将商家端下单方式开关统一为三项：配送、堂食、自提，并让商家可分别控制。

* 补齐前后端与文档口径，使 C 端展示、下单校验、商家设置页和接口返回对三种方式保持一致。

## Current State Analysis

### 1. 用户端订单详情“联系商家退款”

* 页面实现位于 `miniprogram/src/pages/store/order-detail.vue`。

* 当前按钮文案已经是“联系商家退款”，触发函数为 `contactMerchantForRefund()`。

* 现状行为：

  * 有电话时：先弹 `uni.showModal()`，用户再点击“拨打电话”，最后执行 `uni.makePhoneCall()`。

  * 无电话时：仅弹提示“请联系商家协助处理退款”。

* 结论：当前不是“直接拨打”，而是“先确认再拨打”。

### 2. 商家端下单方式开关现状

* 商家模型 `server/internal/models/models.go` 中已有：

  * `takeout_enabled`

  * `dine_in_enabled`

* 目前没有独立的 `pickup_enabled` / `self_pickup_enabled` 字段。

* 商家设置接口 `server/internal/handlers/merchant/handler.go`：

  * `GetSettings()` 返回 `takeout_enabled`、`dine_in_enabled`

  * `UpdateSettings()` 只支持更新这两个字段

* 商家配送设置接口 `server/internal/handlers/merchant/handler.go`：

  * `GetDeliverySettings()` / `UpdateDeliverySettings()` 读写 `MerchantDeliverySettings.enabled`

  * 当前“配送是否开启”实际落在配送设置表，而不是与堂食/自提同一口径

* 商家前端设置页：

  * `miniprogram/src/pages/merchant/settings.vue` 仅展示“配送设置”入口和当前是否开启配送

  * `miniprogram/src/pages/merchant/delivery-settings.vue` 只管理配送开关、配送费和距离规则

* 用户端下单方式展示与校验：

  * `miniprogram/src/pages/store/confirm.vue` 固定展示三种方式：配送、堂食、自提

  * `server/internal/handlers/user/handler.go` 在店铺商品列表返回 `takeout_enabled`、`dine_in_enabled`

  * `CreateOrder()` 当前仅对 `delivery_type = 1` 做配送规则校验，没有看到对“堂食关闭/自提关闭”的服务端拦截

* 结论：

  * “配送/堂食/自提”三种方式没有统一管理口径

  * 自提缺少独立开关

  * 用户端当前仍可能展示商家未开放的方式

  * 服务端下单缺少对堂食、自提关闭场景的强校验

## Proposed Changes

### A. 用户端订单详情改为直接拨打商家电话

#### 目标

* 点击“联系商家退款”后，若商家电话存在，直接调用系统拨号能力，不再弹出二次确认框。

* 若商家电话为空，则保留兜底提示，明确当前无法直接拨打。

#### 修改文件

* `miniprogram/src/pages/store/order-detail.vue`

#### 方案

* 保留按钮文案“联系商家退款”不变。

* 将 `contactMerchantForRefund()` 调整为：

  * 读取 `order.merchant?.phone` 或商家联系电话实际字段

  * 有电话时直接执行 `uni.makePhoneCall({ phoneNumber })`

  * 无电话时弹一次不可拨打提示

* 不改动退款接口和订单状态逻辑，仅优化交互方式。

#### 影响范围

* 仅影响用户端订单详情页的退款联系动作。

* `pages/store/my-orders` 如存在同类“联系商家退款”行为，本次不默认扩展，除非执行时确认需与订单详情页保持完全一致。

### B. 商家端统一三种下单方式开关

#### 目标

* 商家可分别控制：

  * 配送

  * 堂食

  * 自提

* C 端页面只展示商家当前允许的方式。

* C 端创建订单时，服务端对关闭的方式明确拒绝。

#### 设计决策

* 新增商家主表字段：`pickup_enabled`，用于承载“自提开关”。

* 保持配送费用和距离规则仍放在 `merchant_delivery_settings`，但“配送是否支持”统一通过商家设置主口径暴露为 `takeout_enabled`。

* `merchant_delivery_settings.enabled` 继续表示“配送费规则是否启用/有效”，但用户端是否允许选择配送，最终需同时满足：

  * `merchant.takeout_enabled = true`

  * `delivery_settings.enabled = true`

#### 修改文件

* 后端模型与数据库：

  * `server/internal/models/models.go`

  * `server/pkg/database/mysql.go` 或对应迁移逻辑位置

* 后端商家接口：

  * `server/internal/handlers/merchant/handler.go`

* 后端用户端店铺/下单接口：

  * `server/internal/handlers/user/handler.go`

* 小程序类型与接口：

  * `miniprogram/src/types/index.ts`

  * `miniprogram/src/api/index.ts`

* 商家设置页面：

  * `miniprogram/src/pages/merchant/delivery-settings.vue`

  * 视实现需要，`miniprogram/src/pages/merchant/settings.vue`

* C 端确认页：

  * `miniprogram/src/pages/store/confirm.vue`

#### 后端改法

* 在 `Merchant` 模型新增 `pickup_enabled bool`，默认 `true`。

* 在数据库迁移/自检逻辑中新增 `pickup_enabled` 列，默认值为 `1`，避免影响现有商家。

* `GetSettings()` / `UpdateSettings()` 增加 `pickup_enabled` 返回和更新能力。

* 用户端店铺商品列表接口当前会返回 `merchant.min_order_amount / takeout_enabled / dine_in_enabled`，应补充 `pickup_enabled`。

* `CreateOrder()` 增加服务端强校验：

  * `delivery_type = 1`：必须满足 `takeout_enabled = true` 且 `delivery_settings.enabled = true`

  * `delivery_type = 2`：必须满足 `dine_in_enabled = true`

  * `delivery_type = 3`：必须满足 `pickup_enabled = true`

* 返回错误文案需明确区分“商家暂未开启配送 / 堂食 / 自提”。

#### 前端改法

* 扩展 `MerchantSettings`、店铺返回 merchant 信息类型，增加 `pickup_enabled`。

* 商家端在 `pages/merchant/delivery-settings.vue` 中增加三种方式开关区域：

  * 配送开关

  * 堂食开关

  * 自提开关

* 配送规则表单继续只在“配送开关开启”时展示。

* 保存时拆分为两部分数据提交：

  * 下单方式开关走 `updateMerchantSettings`

  * 配送费规则走 `updateDeliverySettings`

  * 或在现有接口层封装成一个页面级保存流程，顺序提交两个接口

* `pages/store/confirm.vue` 需根据商家开关动态展示可选方式，不再固定渲染三种：

  * 配送关闭时不显示配送项

  * 堂食关闭时不显示堂食项

  * 自提关闭时不显示自提项

* 若当前选中的方式在最新配置下不可用，应自动回退到第一个可用方式；若三种都关闭，应给出“商家暂未开放下单方式”的提示并阻止提交。

## Assumptions & Decisions

* “直接拨打商家电话”按字面执行，不保留二次确认弹窗。

* “商家自提”对应新增独立字段 `pickup_enabled`，不复用现有 `takeout_enabled` 语义。

* 配送相关仍保留现有 `merchant_delivery_settings` 表，不重构配送费规则模型。

* 商家设置页的三种方式开关优先放在 `pages/merchant/delivery-settings.vue`，避免新增新的设置页面和入口。

* 本次属于接口、数据模型和页面行为变更，执行时必须同步更新：

  * `PRD.md`

  * `docs/prd/PRD-功能说明.md`

  * `docs/prd/PRD-接口文档.md`

  * `docs/prd/PRD-测试与附录.md`

## Verification Steps

### 1. 用户端订单详情拨号

* 打开 `pages/store/order-detail`

* 点击“联系商家退款”

* 有电话时直接拉起系统拨号

* 无电话时只出现兜底提示，不报错

### 2. 商家设置三种开关

* 商家进入 `pages/merchant/delivery-settings`

* 可看到配送、堂食、自提三项开关

* 保存后重新进入页面，开关状态保持一致

### 3. C 端确认页可选方式

* 仅开启配送：确认页只显示配送

* 仅开启堂食：确认页只显示堂食

* 仅开启自提：确认页只显示自提

* 多项开启：确认页展示全部已开启方式

### 4. 服务端下单校验

* 对关闭的方式提交订单，请求被拒绝并返回对应提示

* 对开启的方式提交订单，请求正常创建

### 5. 基础验证

* 后端相关 handler / model / migration 改动通过诊断与测试

* 小程序相关页面和类型改动通过诊断

* 小程序验收以 `npm run build:mp-weixin` 产物为准

