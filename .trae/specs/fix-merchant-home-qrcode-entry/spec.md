# 商家主页二维码商家ID参数对齐 Spec

## Why
小程序已经发布成功，商家主页中的“店铺二维码”需要进入真实可用的店铺首页入口。当前二维码生成链路中的页面路径与 `scene` 参数命名没有和现有店铺页入口解析规则完全对齐，存在扫码后无法稳定带出商家 ID 的风险。

## What Changes
- 调整商家二维码生成接口，使生成的小程序码入口指向当前实际使用的店铺首页。
- 确保二维码携带的参数格式与 `pages/store/home` 现有的 `merchant_id` / `scene` 解析逻辑一致。
- 保留商家主页现有的二维码加载与展示方式，不新增前端交互入口。
- 保留现有二维码接口返回结构中的调试信息，便于确认 `scene` 与 `page` 是否正确。

## Impact
- Affected specs:
  - 商家主页二维码生成
  - C 端店铺首页扫码进入链路
- Affected code:
  - `server/internal/handlers/merchant/handler.go`
  - `miniprogram/src/pages/merchant/home.vue`
  - `miniprogram/src/api/index.ts`
  - `miniprogram/src/utils/storeEntry.ts`

## ADDED Requirements
### Requirement: 商家二维码必须对齐店铺首页入口参数
系统 SHALL 为商家后台生成可直接进入当前店铺首页的小程序二维码，并保证二维码中包含可解析出的商家 ID。

#### Scenario: 生成正式二维码
- **WHEN** 商家调用 `/api/v1/merchant/qrcode` 获取店铺二维码
- **THEN** 接口生成的小程序码页面路径应指向当前实际店铺首页
- **AND** 二维码携带的参数应能解析出当前登录商家的 `merchant_id`

#### Scenario: 扫码进入店铺首页
- **WHEN** 用户扫描商家主页展示的店铺二维码
- **THEN** 小程序应进入 `pages/store/home`
- **AND** 页面入口参数解析后得到的 `merchantId` 应等于二维码所属商家 ID

#### Scenario: 调试返回信息可核对
- **WHEN** 调用商家二维码接口成功返回
- **THEN** 返回数据中的 `scene` 与 `page` 字段应与实际二维码生成配置一致

## MODIFIED Requirements
### Requirement: 商家主页二维码展示
商家主页 SHALL 展示与当前商家绑定的店铺二维码，且二维码对应的小程序入口必须与现有店铺首页解析规则保持一致。

#### Scenario: 商家主页加载二维码
- **WHEN** 商家进入 `pages/merchant/home`
- **THEN** 页面调用 `/api/v1/merchant/qrcode`
- **AND** 页面展示的二维码应对应当前商家店铺首页，而不是旧页面或无商家 ID 的入口

### Requirement: 店铺首页扫码参数解析
`pages/store/home` SHALL 继续兼容通过 `merchant_id` 或 `scene` 进入，但商家二维码生成链路必须至少保证其中一种方式稳定携带当前商家 ID。

#### Scenario: scene 参数格式对齐
- **WHEN** 二维码通过 `scene` 传递参数
- **THEN** `scene` 中的键名必须与 `parseStoreEntryOptions()` 当前支持的字段一致
- **AND** 不能继续使用无法被该解析逻辑识别的旧键名

## REMOVED Requirements
### Requirement: 商家二维码可指向旧首页占位入口
**Reason**: 当前小程序已发布，二维码必须进入真实可用的 C 端店铺首页，不再允许继续沿用旧的占位页面路径。
**Migration**: 将二维码生成逻辑统一切换到 `pages/store/home`，并使用可被现有入口解析逻辑识别的商家 ID 参数格式。
