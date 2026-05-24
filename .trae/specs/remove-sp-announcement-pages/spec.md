# 删除服务商公告页面 Spec

## Why
当前小程序仍保留“服务商发布公告”页面与入口，但产品方向已明确要逐步将服务商能力迁移到 PC 端。为减少小程序端维护成本与功能分散，需要先下线服务商身份的公告发布页面，同时保留后端接口，避免影响后续 PC 端接入。

## What Changes
- 删除小程序服务商端“系统公告”入口与页面路由。
- 删除小程序内服务商公告列表页与编辑页实现文件。
- 保留服务商公告相关后端接口与数据模型，不做接口删除或行为修改。
- 保留商家端公告读取与展示能力，不影响商家工作台公告栏。

## Impact
- Affected specs: 服务商小程序后台导航、服务商公告发布入口、商家公告查看能力
- Affected code:
  - `miniprogram/src/pages/sp/home.vue`
  - `miniprogram/src/pages/sp/announcements/index.vue`
  - `miniprogram/src/pages/sp/announcements/edit.vue`
  - `miniprogram/src/pages.json`
  - `miniprogram/src/api/index.ts`
  - `docs/prd/PRD-功能说明.md`
  - `PRD.md`

## ADDED Requirements
### Requirement: 服务商公告能力迁移约束
系统 SHALL 在小程序端移除服务商公告发布页面，但继续保留后端公告接口，供后续 PC 端复用。

#### Scenario: 小程序服务商进入工作台
- **WHEN** 服务商登录并进入小程序工作台
- **THEN** 页面中不再出现“系统公告”菜单入口

#### Scenario: 小程序路由配置检查
- **WHEN** 检查小程序页面配置
- **THEN** 不再包含 `pages/sp/announcements/index` 与 `pages/sp/announcements/edit` 两个路由

#### Scenario: 后端公告接口可继续复用
- **WHEN** 检查服务商公告接口注册与处理器
- **THEN** `/api/v1/sp/announcements` 相关 GET/POST/PUT/DELETE 接口仍然保留

## MODIFIED Requirements
### Requirement: 服务商小程序后台能力范围
服务商小程序 SHALL 仅保留仍需在移动端使用的后台能力；已明确迁移到 PC 端的功能，不再在小程序中提供页面入口或页面实现。

#### Scenario: 服务商公告能力下线
- **WHEN** 服务商使用小程序后台
- **THEN** 无法再通过小程序进入公告列表页或公告编辑页
- **AND** 该变化不影响商家端公告获取与展示

## REMOVED Requirements
### Requirement: 服务商可在小程序发布系统公告
**Reason**: 服务商公告发布能力计划迁移到 PC 端，小程序不再承担该操作入口。

**Migration**: 保留现有后端公告接口与数据结构；后续 PC 端直接复用接口实现公告管理。
