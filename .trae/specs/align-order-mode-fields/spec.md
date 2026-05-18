# 下单方式字段一一对应 Spec

## Why
当前下单方式有配送、堂食、自提三种，但存储与前端展示口径不完全一致，容易出现页面只显示部分方式、保存后状态不一致、或接口语义混乱的问题。需要将三种下单方式明确映射到商家表中的三个独立字段，并让客户端下单方式与后端字段一一对应。

## What Changes
- 商家表使用三个独立字段控制三种下单方式：
  - `takeout_enabled` 对应配送
  - `dine_in_enabled` 对应堂食
  - `pickup_enabled` 对应自提
- 商家设置接口需完整返回并保存这三个字段。
- 客户端确认页需根据这三个字段动态展示可选下单方式。
- 下单接口需根据这三个字段对 `delivery_type` 做一一对应校验。
- 数据初始化与迁移需保证 `pickup_enabled` 有明确默认值，避免历史商家缺失字段导致行为异常。

## Impact
- Affected specs:
  - 商家设置
  - C 端确认订单
  - C 端创建订单
  - 商家数据模型与迁移
- Affected code:
  - `server/internal/models/models.go`
  - `server/internal/handlers/merchant/handler.go`
  - `server/internal/handlers/user/handler.go`
  - `server/pkg/database/mysql.go`
  - `server/migrations/*.sql`
  - `miniprogram/src/pages/merchant/delivery-settings.vue`
  - `miniprogram/src/pages/store/confirm.vue`
  - `miniprogram/src/types/index.ts`
  - `miniprogram/src/api/index.ts`
  - `miniprogram/src/api/store.ts`

## ADDED Requirements
### Requirement: 商家表存储三种下单方式
系统 SHALL 在商家表层面提供三个独立字段来控制配送、堂食、自提是否可用，其中自提必须有独立字段 `pickup_enabled`。

#### Scenario: 历史商家自动补齐字段
- **WHEN** 旧商家数据中尚未存在 `pickup_enabled`
- **THEN** 系统自动补齐该字段并赋默认值
- **THEN** 历史商家不会因为字段缺失而导致客户端下单方式显示异常

### Requirement: 商家设置接口完整读写三种方式
系统 SHALL 通过商家设置接口完整返回并保存 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled`。

#### Scenario: 商家保存三种方式
- **WHEN** 商家在设置页修改配送、堂食、自提开关并保存
- **THEN** 接口持久化三个字段的最新值
- **THEN** 商家再次进入设置页时看到的开关状态与保存值一致

### Requirement: 客户端下单方式与字段一一对应
系统 SHALL 在客户端按商家表三个字段一一展示下单方式，确保配送、堂食、自提分别对应后端三个独立字段。

#### Scenario: 三种方式全部开启
- **WHEN** 商家配置 `takeout_enabled=true`、`dine_in_enabled=true`、`pickup_enabled=true`
- **THEN** 客户端确认页显示配送、堂食、自提三个选项

#### Scenario: 仅开启部分方式
- **WHEN** 商家只开启其中一个或两个字段
- **THEN** 客户端只显示对应开启的下单方式
- **THEN** 不显示未开启的下单方式

### Requirement: 创建订单校验与下单方式一致
系统 SHALL 在创建订单时按照 `delivery_type` 与三个字段做一一对应校验。

#### Scenario: 配送下单
- **WHEN** 用户提交 `delivery_type=1`
- **THEN** 系统校验 `takeout_enabled=true`
- **THEN** 若未开启则返回“商家暂未开启配送”

#### Scenario: 堂食下单
- **WHEN** 用户提交 `delivery_type=2`
- **THEN** 系统校验 `dine_in_enabled=true`
- **THEN** 若未开启则返回“商家暂未开启堂食”

#### Scenario: 自提下单
- **WHEN** 用户提交 `delivery_type=3`
- **THEN** 系统校验 `pickup_enabled=true`
- **THEN** 若未开启则返回“商家暂未开启自提”

## MODIFIED Requirements
### Requirement: 配送设置与下单方式控制
系统 SHALL 将配送费、配送距离规则继续保留在配送设置结构中，但“是否允许选择配送下单”由商家表字段 `takeout_enabled` 明确控制，而不是由客户端自行推断。

#### Scenario: 配送规则存在但配送开关关闭
- **WHEN** 商家已配置配送费规则，但 `takeout_enabled=false`
- **THEN** 客户端不展示配送选项
- **THEN** 创建订单接口拒绝 `delivery_type=1`

### Requirement: 商家配送设置页
系统 SHALL 在商家配送设置页中同时展示配送、堂食、自提三个开关，并将三者与后端字段一一绑定。

#### Scenario: 页面回显
- **WHEN** 商家进入配送设置页
- **THEN** 页面根据 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled` 正确回显三个开关状态

## REMOVED Requirements
### Requirement: 客户端根据不完整字段集合推断下单方式
**Reason**: 旧方案会导致客户端展示与商家真实配置不一致。  
**Migration**: 统一改为以后端三个独立字段作为唯一来源，客户端只做直连映射，不再做模糊兜底推断。
