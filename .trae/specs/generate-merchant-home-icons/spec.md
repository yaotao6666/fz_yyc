# 商家首页缺失图标生成 Spec

## Why
`pages/merchant/home` 仍有 5 个首页图标未完成生成，导致商家首页视觉不完整，快捷操作区与顶部操作区的图标风格无法统一。

## What Changes
- 为商家首页补齐 5 个已实际引用的图标资源
- 保持现有页面结构、交互和图标文件引用路径不变
- 统一这 5 个图标的视觉风格、尺寸适配与透明背景输出

## Impact
- Affected specs: 商家首页视觉资源、商家快捷功能入口
- Affected code: `miniprogram/src/pages/merchant/home.vue`, `miniprogram/src/static/icons/*`

## ADDED Requirements
### Requirement: 商家首页图标资源补齐
系统 SHALL 为商家首页当前已引用的 5 个图标位提供可直接使用的图标资源，且资源路径与页面现有引用保持一致。

#### Scenario: 顶部操作图标可正常显示
- **WHEN** 商家进入 `pages/merchant/home`
- **THEN** 顶部“暂停营业/开始营业”显示 `store.png`
- **THEN** 顶部“店铺二维码”显示 `qrcode.png`
- **THEN** 图标在浅色背景上清晰可见且无拉伸变形

#### Scenario: 快捷功能图标可正常显示
- **WHEN** 商家进入 `pages/merchant/home`
- **THEN** “分类管理”显示 `category.png`
- **THEN** “商品管理”显示 `product.png`
- **THEN** “快速核销”显示 `order.png`
- **THEN** 3 个图标风格一致、与各自语义匹配

### Requirement: 图标资源兼容现有页面样式
系统 SHALL 输出适配当前页面样式的图标资源，而不要求调整 `pages/merchant/home` 的布局结构。

#### Scenario: 复用现有引用路径
- **WHEN** 页面继续使用当前图标路径
- **THEN** 无需修改 `quickMenuItems` 的 `icon` 字段命名
- **THEN** 无需新增额外的页面逻辑分支

#### Scenario: 图标适配现有容器
- **WHEN** 图标被渲染在顶部 `action-icon` 与快捷入口 `menu-icon-image`
- **THEN** 图标在当前容器尺寸内清晰显示
- **THEN** 不需要额外放大、裁切或替换为文本图标

## MODIFIED Requirements
### Requirement: 商家首页视觉完整性
商家首页当前已引用的操作入口，必须全部使用正式图标资源，不允许保留临时占位、缺失图、不可识别图形或风格不一致的资源。

#### Scenario: 首页图标验收
- **WHEN** 打开 `pages/merchant/home`
- **THEN** 顶部 2 个操作图标和快捷功能 3 个图标全部正常显示
- **THEN** 不出现空白、破图、默认占位或明显风格割裂

## REMOVED Requirements
### Requirement: 首页图标使用临时资源
**Reason**: 当前商家首页已进入可用阶段，首页核心入口不应继续依赖临时图标或未最终生成的资源。
**Migration**: 将现有对应文件替换为正式图标资源，并保持原引用路径不变。
