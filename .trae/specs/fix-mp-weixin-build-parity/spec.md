# 小程序 Build 与 Dev 一致性修复 Spec

## Why
当前 `npm run dev:mp-weixin` 运行正常，但 `npm run build:mp-weixin` 输出包存在运行异常，说明开发态与构建态行为不一致。该问题会直接影响真机发布与交付，因此必须以构建产物为验收基准，并补齐一致性保障。

## What Changes
- 明确 `build:mp-weixin` 为小程序交付与验收基准，`dev:mp-weixin` 仅作为开发预览。
- 修复 `dev` 与 `build` 模式下的模块加载、静态资源引用、页面入口与公开接口请求行为差异。
- 补充针对 C 端店铺页面的构建态回归检查，覆盖 `pages/store/home`、`pages/store/product`、`pages/store/confirm`。

## Impact
- Affected specs: 小程序构建发布、C 端店铺浏览与下单链路、静态资源加载、页面入口解析
- Affected code: `miniprogram/src/pages/store/*`、`miniprogram/src/api/*`、`miniprogram/src/utils/*`、构建产物验证流程

## ADDED Requirements
### Requirement: Build 产物作为验收基准
系统 SHALL 以 `npm run build:mp-weixin` 生成的小程序包作为发布前验收基准。

#### Scenario: 发布前验收
- **WHEN** 开发完成并准备交付小程序功能
- **THEN** 必须以 `build:mp-weixin` 产物验证页面与接口行为
- **THEN** `dev:mp-weixin` 结果不能替代构建产物验收

### Requirement: Dev 与 Build 页面行为一致
系统 SHALL 保证 `dev:mp-weixin` 与 `build:mp-weixin` 在同一页面入口参数下表现一致。

#### Scenario: 带 merchant_id 进入店铺首页
- **WHEN** 用户通过 `merchant_id` 或 `scene` 进入 `pages/store/home`
- **THEN** `dev` 与 `build` 模式都应请求对应商家首页接口
- **THEN** 页面不得因模块加载异常导致接口方法为 `undefined`

#### Scenario: 店铺商品与确认页加载
- **WHEN** 用户进入 `pages/store/product` 或 `pages/store/confirm`
- **THEN** `dev` 与 `build` 模式都应按相同入口参数解析商家数据
- **THEN** 商品详情与配送规则等公开接口在两种模式下都可正常请求

### Requirement: 构建态静态资源可用
系统 SHALL 保证 C 端店铺页面在 `build:mp-weixin` 模式下引用的静态资源实际存在且可加载。

#### Scenario: 默认封面与默认商品图展示
- **WHEN** 店铺封面图或商品图为空
- **THEN** 页面应展示可用的默认占位资源或占位样式
- **THEN** 构建态不得出现本地静态资源 500 或加载失败

## MODIFIED Requirements
### Requirement: C 端店铺页公开数据加载
`pages/store/home`、`pages/store/product`、`pages/store/confirm` 的公开数据加载必须在 `dev` 与 `build` 模式下保持同样的触发顺序与结果。

#### Scenario: 公开接口先于登录依赖
- **WHEN** 页面存在无需登录的公开店铺接口
- **THEN** 页面应先按入口参数发起公开接口请求
- **THEN** 登录、埋点等依赖不得阻塞公开接口加载

## REMOVED Requirements
### Requirement: 以开发态结果代替发布验收
**Reason**: `dev:mp-weixin` 与 `build:mp-weixin` 存在运行时差异，开发态正常不能代表发布包正常。
**Migration**: 后续所有小程序交付与问题验收统一改为优先校验 `build:mp-weixin` 产物，再以 `dev:mp-weixin` 辅助调试。
