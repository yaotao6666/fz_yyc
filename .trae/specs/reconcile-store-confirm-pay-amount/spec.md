# 修复确认页支付金额不一致 Spec

## Why
`pages/store/confirm` 当前同时存在前端预估金额和后端实际结算金额两套口径。用户在确认页看到的“实付金额”可能与下单后微信实际拉起支付的金额不一致，影响下单信任感。

## What Changes
- 调整 `pages/store/confirm` 的金额展示逻辑，明确区分“预估金额”和“最终支付金额”
- 新增以下约束：一旦创建订单成功，确认页后续展示和支付前提示必须以后端返回的 `order.discount_amount`、`order.delivery_fee`、`order.pay_amount` 为准
- 梳理满减、配送费、免配送门槛在确认页的展示口径，避免仅以前端本地推导作为最终支付依据
- 补充创建订单成功后前端状态更新要求，保证支付前后金额文案一致

## Impact
- Affected specs: 用户下单确认页、满减展示、配送费展示、订单创建与支付拉起链路
- Affected code: `miniprogram/src/pages/store/confirm.vue`、`miniprogram/src/api/store.ts`、`miniprogram/src/types/index.ts`、`server/internal/handlers/user/handler.go`

## ADDED Requirements
### Requirement: 确认页金额展示一致性
系统 SHALL 在 `pages/store/confirm` 中保证用户看到的最终支付金额与后端创建订单返回的实际支付金额一致。

#### Scenario: 创建订单前展示预估金额
- **WHEN** 用户尚未点击提交订单，且确认页仅基于购物车、配送档位、满减规则进行本地计算
- **THEN** 页面必须将该金额视为预估结果
- **AND** 页面不得将本地预估金额直接等同于最终支付金额

#### Scenario: 创建订单后以后端金额为准
- **WHEN** 用户点击提交订单且后端成功返回订单数据
- **THEN** 页面必须以后端返回的 `order.delivery_fee`、`order.discount_amount`、`order.pay_amount` 作为最终支付口径
- **AND** 支付前展示、支付埋点、成功页跳转前的金额说明必须使用同一套后端返回金额

#### Scenario: 前端预估与后端实际不一致
- **WHEN** 因配送费计算、满减命中、免配送门槛或后端结算规则导致前端预估金额与后端返回金额不同
- **THEN** 页面必须自动刷新为后端实际金额
- **AND** 用户不会继续看到旧的本地预估值

### Requirement: 金额明细口径统一
系统 SHALL 在确认页中统一“商品金额 / 配送费 / 优惠金额 / 实付金额”的展示来源与更新时机。

#### Scenario: 明细字段统一展示
- **WHEN** 确认页展示金额摘要
- **THEN** 商品金额、配送费、优惠金额、实付金额必须使用可追溯且一致的字段来源
- **AND** 不得出现顶部金额、底部金额、支付实际金额三者口径不同

## MODIFIED Requirements
### Requirement: 确认订单页满减与实付展示
系统 SHALL 在 `pages/store/confirm` 中展示满减规则与金额明细；在创建订单前可显示本地预估优惠，但创建订单成功后必须切换为后端返回的最终优惠和实付金额。

#### Scenario: 满减展示切换为最终口径
- **WHEN** 用户命中满减并成功创建订单
- **THEN** 确认页中的优惠金额和实付金额必须以后端订单返回值为准
- **AND** 满减提示文案不得继续沿用旧的本地预估金额

## REMOVED Requirements
### Requirement: 本地计算结果直接代表最终支付金额
**Reason**: 本地预估无法覆盖后端最终结算规则，容易导致确认页金额与实际支付金额不一致。
**Migration**: 保留本地预估仅用于提交前提示；创建订单成功后统一切换为后端返回金额。
