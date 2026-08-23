# 微信支付模式切换方案

## Summary

- 目标：在当前“微信支付服务商 + 商户主体小程序”实现上，新增 1 个 `.env` 配置项，用于全局切换当前部署版本的小程序支付身份。
- 两种模式定义：
  - `sp_app`：1 个微信支付服务商 + 该商户主体的小程序，支付时使用 `sp_appid + sp_openid`
  - `sub_app`：1 个微信支付服务商 + 1 个关联主体的小程序，支付时仍走系统支付体系，但使用 `sub_appid + sub_openid`
- 已确认决策：
  - 系统始终只有 1 个全局小程序身份，不存在每个商户一个不同小程序
  - 支付、退款、回调都继续走系统支付体系，不改成直连商户模式
  - 仅通过一个全局 `.env` 模式开关控制当前部署版本运行在哪种小程序支付身份下
  - 切到 `sub_app` 后，用户登录也要切到该全局关联主体小程序，拿到 `sub_openid`

## Current State Analysis

### 1. 配置层现状

- `server/internal/config/config.go`
  - 当前只有一套全局小程序配置：`WECHAT_APP_ID`、`WECHAT_APP_SECRET`
  - 当前只有一套全局系统支付配置：`WECHAT_PAY_SP_MCH_ID`、`WECHAT_PAY_SP_API_V3_KEY`、`WECHAT_PAY_SP_CERT_SERIAL_NO`、`WECHAT_PAY_SP_PRIVATE_KEY`、`WECHAT_PAY_SP_PUBLIC_KEY`、`WECHAT_PAY_SP_CALLBACK_URL`
  - 没有“当前运行模式”字段，也没有第二套小程序配置
- `server/.env.example`
  - 仅说明了商户小程序和系统支付配置，没有模式切换说明，也没有关联主体小程序配置

### 2. 用户登录现状

- `server/internal/handlers/user/handler.go`
  - `WechatLogin()` 固定走 `getWechatOpenID()`
  - `getWechatOpenID()` 固定使用全局 `config.Config.Wechat.AppID/AppSecret`
  - `users.openid` 当前默认就是这套全局小程序下的 `openid`
- `miniprogram/src/utils/useAuth.ts`
  - 前端登录缓存固定为一套全局 `user_token` / `userInfo` / `openid`
  - 未感知当前是哪种支付小程序模式
- 结论：
  - 当前系统只支持“单一全局小程序身份登录”
  - 如果切换到 `sub_app`，登录层只需要全局切换到第二套小程序配置，不需要商户级隔离

### 3. 支付下单现状

- `server/internal/handlers/user/handler.go`
  - `CreateOrder()` 创建支付单时固定调用 `wechatpay.NewServiceProviderClient()`
  - `createWechatPayOrder()` 固定把 `config.Config.Wechat.AppID` 作为支付请求中的 `AppID`
- `server/internal/services/wechatpay/service_provider.go`
  - `CreatePartnerJSAPIPayOrder()` 当前固定发送：
    - `sp_appid`
    - `sp_mchid`
    - `sub_mchid`
    - `payer.sp_openid`
  - 没有支持：
    - `sub_appid`
    - `payer.sub_openid`
- 结论：
  - 当前只支持“商户小程序身份发起 partner JSAPI 支付”

### 4. 回调 / 退款 现状

- `server/internal/handlers/sp/payment.go`
  - 支付回调和退款回调都通过 `wechatpay.NewServiceProviderClient()` 解析和处理
- `server/internal/handlers/merchant/order.go`
  - 退款继续调用支付客户端
- 结论：
  - 这几条链路都建立在“系统支付体系”之上
  - 只要新模式仍然走 partner JSAPI，并继续使用支付证书与API v3 Key，这些链路原则上无需改协议层

### 5. 商户数据现状

- `server/internal/models/models.go`
  - `Merchant` 已有 `sub_mch_id`
  - 没有任何商户级 `app_id/app_secret`
- 结论：
  - 在你修正后的目标里，这是合理的
  - 因为系统始终只有 1 个全局小程序身份，不需要给每个商户单独存小程序配置

## Proposed Changes

### 1. 新增全局支付模式开关与第二套小程序配置

- 修改文件：`server/internal/config/config.go`
- 修改内容：
  - 为配置新增支付模式字段，例如：
    - `WechatPay.AppMode string`
  - 保留现有商户主体小程序配置：
    - `WECHAT_APP_ID`
    - `WECHAT_APP_SECRET`
  - 新增关联主体小程序配置：
    - `WECHAT_SUB_APP_ID`
    - `WECHAT_SUB_APP_SECRET`
  - 新增模式变量：
    - `WECHAT_PAY_APP_MODE=sp_app|sub_app`
- 模式语义：
  - `sp_app`：当前默认逻辑
  - `sub_app`：登录与下单使用关联主体小程序身份，但支付、退款、回调仍走系统支付体系
- 同步文件：
  - `server/.env.example`
  - `server/.env.production.example`

### 2. 抽象“当前生效小程序身份”选择逻辑

- 新增建议文件：`server/internal/services/wechatpay/app_identity.go`
- 目标：
  - 统一根据 `WECHAT_PAY_APP_MODE` 选择当前生效的小程序配置
- 建议抽象结果：
  - `Mode`
  - `AppID`
  - `AppSecret`
  - `OpenIDFieldName`
- 逻辑定义：
  - `sp_app`
    - `AppID = WECHAT_APP_ID`
    - `AppSecret = WECHAT_APP_SECRET`
    - `OpenIDFieldName = sp_openid`
  - `sub_app`
    - `AppID = WECHAT_SUB_APP_ID`
    - `AppSecret = WECHAT_SUB_APP_SECRET`
    - `OpenIDFieldName = sub_openid`
- 这样可以避免模式判断散落在登录、下单、支付请求组装等多个位置

### 3. 重构用户登录逻辑，按全局模式切换 `openid` 来源

- 修改文件：`server/internal/handlers/user/handler.go`
- 需要改动的函数：
  - `WechatLogin()`
  - `getWechatOpenID()`
- 修改方向：
  - `getWechatOpenID()` 不再固定读 `config.Config.Wechat.AppID/AppSecret`
  - 改为读取“当前生效小程序身份”
- 实际行为：
  - `sp_app`：
    - 继续用 `WECHAT_APP_ID/WECHAT_APP_SECRET` 调 `jscode2session`
    - 返回商户主体小程序下的 `openid`
  - `sub_app`：
    - 改用 `WECHAT_SUB_APP_ID/WECHAT_SUB_APP_SECRET`
    - 返回关联主体小程序下的 `openid`
- 数据影响：
  - 本轮不引入商户级 `openid` 隔离
  - 仍以“当前部署实例只有 1 个全局小程序身份”为前提，沿用现有 `users.openid` 结构

### 4. 扩展系统支付下单请求，支持 `sub_appid/sub_openid`

- 修改文件：`server/internal/services/wechatpay/service_provider.go`
- 修改内容：
  - 扩展 `JSAPIPayRequest`
  - 调整 `CreatePartnerJSAPIPayOrder()` 的请求体组装逻辑
- 具体规则：
  - `sp_app` 模式：
    - 发送 `sp_appid`
    - `payer.sp_openid`
  - `sub_app` 模式：
    - 发送 `sub_appid`
    - `payer.sub_openid`
  - 两种模式都继续发送：
    - `sp_mchid`
    - `sub_mchid`
  - 统一继续调用 `/v3/pay/partner/transactions/jsapi`
- 说明：
  - 这一步是本次方案的核心改动点
  - 本质上是“partner JSAPI 下的小程序身份切换”，不是支付模式整体改成直连商户

### 5. 调整下单支付入口，按全局模式选择支付参数

- 修改文件：`server/internal/handlers/user/handler.go`
- 修改内容：
  - `createWechatPayOrder()` 不再直接使用 `config.Config.Wechat.AppID`
  - 改为读取“当前生效小程序身份”的 `AppID`
  - 当前用户 `OpenID` 也按该模式下的登录结果直接传入
- 校验建议：
  - `sp_app` 模式要求：
    - `WECHAT_APP_ID`
    - `WECHAT_APP_SECRET`
  - `sub_app` 模式要求：
    - `WECHAT_SUB_APP_ID`
    - `WECHAT_SUB_APP_SECRET`
  - 任一模式下都仍要求：
    - `WECHAT_PAY_SP_MCH_ID`
    - 商户 `sub_mch_id`
    - 支付证书/私钥/API V3 Key/回调地址

### 6. 前端用户态增加“模式切换后强制重登”的兼容策略

- 修改文件：
  - `miniprogram/src/utils/useAuth.ts`
  - 若有必要，补充 `miniprogram/src/types/index.ts`
- 当前情况：
  - 前端缓存只有一套全局 `openid`
- 本轮建议：
  - 不做商户级缓存拆分
  - 仅增加“当前登录 app 模式标记”，例如：
    - `user_login_app_mode`
  - 当发现本地缓存模式与服务端当前模式不一致时：
    - 清理旧 `user_token`
    - 清理旧 `openid`
    - 重新执行登录
- 这样可以避免从 `sp_app` 切到 `sub_app` 后继续错误复用旧 `openid`

### 7. 回调、退款保持支付链路不变

- 主要涉及文件：
  - `server/internal/handlers/sp/payment.go`
  - `server/internal/handlers/merchant/order.go`
  - `server/internal/services/orderquery/query.go`
- 本轮策略：
  - 不改支付客户端初始化
  - 不改回调解密配置来源
  - 不改退款调用链路
- 前提：
  - `sub_app` 仍是商户模式下的 partner JSAPI，只是变更小程序身份字段

### 8. 文档同步

- 修改文件：
  - `PRD.md`
  - `docs/prd/PRD-功能说明.md`
  - `docs/prd/PRD-接口文档.md`
  - `docs/prd/PRD-测试与附录.md`
- 需要补充：
  - 新增全局支付模式开关 `WECHAT_PAY_APP_MODE`
  - `sp_app` 与 `sub_app` 的业务定义
  - 登录接口与支付下单在两种模式下的行为差异
  - 测试验收中新增“模式切换后重新登录”和“sub_app 下仍可回调/退款”的验证项

## Assumptions & Decisions

- 决策：系统始终只有 1 个全局小程序身份，不做商户级小程序配置
- 决策：新增模式值建议仅支持：
  - `sp_app`
  - `sub_app`
- 决策：`sub_app` 不是直连商户支付，而是partner JSAPI 下切换为 `sub_appid/sub_openid`
- 决策：支付证书、回调解密、退款继续沿用现有配置
- 决策：本轮不新增数据库字段、不新增迁移、不改商户管理数据模型
- 决策：前端只做“模式变化后清理旧登录态并重登”，不做多身份并存
- 假设：当前部署实例只会服务于 1 个小程序版本，不会在同一实例内同时混跑 `sp_app` 和 `sub_app`
- 假设：关联主体小程序已在微信支付侧完成与系统支付体系所需的关联配置

## Verification Steps

- 配置验证
  - 检查 `server/internal/config/config.go` 可正确读取：
    - `WECHAT_PAY_APP_MODE`
    - `WECHAT_SUB_APP_ID`
    - `WECHAT_SUB_APP_SECRET`
  - 检查 `server/.env.example` 与 `server/.env.production.example` 已补充配置说明
- 登录验证
  - `WECHAT_PAY_APP_MODE=sp_app`
    - `WechatLogin()` 使用 `WECHAT_APP_ID`
    - 用户 `openid` 来源保持现状
  - `WECHAT_PAY_APP_MODE=sub_app`
    - `WechatLogin()` 使用 `WECHAT_SUB_APP_ID`
    - 模式变更后前端会清理旧登录态并重登
- 支付验证
  - `sp_app` 模式：
    - partner JSAPI 请求体包含 `sp_appid + payer.sp_openid`
  - `sub_app` 模式：
    - partner JSAPI 请求体包含 `sub_appid + payer.sub_openid`
  - 两种模式都继续包含：
    - `sp_mchid`
    - `sub_mchid`
- 回归验证
  - 支付成功后订单状态可正常落账
  - 支付回调可继续正确解密
  - 商家退款可继续走通
  - 原有 `sp_app` 模式无行为回归
