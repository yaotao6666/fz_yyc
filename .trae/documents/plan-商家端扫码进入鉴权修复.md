# 商家端扫码进入鉴权修复计划

## Summary

目标：修复“进入商家端页面后所有订单详情/退款接口提示未提供认证令牌”的问题；保证扫码/跳转进入订单详情、订单退款等链路在未登录或登录失效时能正确引导登录并恢复操作。

成功标准：

- `GET /api/v1/merchant/orders/:id` 在已登录时始终携带 `Authorization: Bearer {token}` 且不再出现“未提供认证令牌”。
- 商家端订单退款（列表/详情）在已登录时可正常请求；未登录时能跳转登录页，登录成功后回到原页面继续操作。
- 相关变更同步到 PRD（根 PRD + 接口文档中商家端鉴权约束/跳转规则）。

## Current State Analysis

### 1) 现象（用户反馈）

- 进入商家端后，“每次”查看订单详情都会报错：`message: 未提供认证令牌`。
- 商家端发起退款也提示：`未提供认证令牌`。

### 2) 后端鉴权现状（确定事实）

- 商家端订单相关接口属于 `merchantGroup.Use(middleware.JWTAuth(), middleware.MerchantAuth())` 管控，缺少 `Authorization` 时会返回 HTTP 401 + `code=1002`（未提供认证令牌）。
- 因此该类报错只可能由前端未携带 token，或 token 被清空/未初始化导致。

### 3) 前端 token 存储/读取现状（确定事实）

- 商家端请求封装位于 `miniprogram/src/utils/request.ts`，默认从 `uni.getStorageSync('token')` 读取并设置 `Authorization: Bearer ...`。
- 商家端登录状态由 Pinia `useAuthStore` 管理：`miniprogram/src/stores/auth.ts`。
  - `state.token` 初始化来自 `uni.getStorageSync('token')`。
  - `checkLogin()` 也以 `uni.getStorageSync('token')` 为主恢复登录态。
- App 启动时会调用 `authStore.checkLogin()`：`miniprogram/src/App.vue`。

### 4) 推断出的根因范围（需要用执行阶段验证）

由于“每次进入订单详情/退款都报未提供认证令牌”，最可能的问题是：

- 入口允许在 **未登录** 状态直接进入商家端订单页面（例如扫码/跳转进入非 Tab 页面），但页面自身未做登录态校验与跳转；
- 或者 `token` 在某些路径被清理（401/1002）后未能引导重新登录；
- 或者登录成功后未按预期写入 `token`（但当前代码显示账号登录与微信快捷登录均会写入）。

## Proposed Changes

### A. 增加“商家端页面鉴权守卫”并统一复用

目的：在进入商家端页面（尤其是非 Tab 页面，如订单详情）时，若未登录则跳转登录页，避免继续请求后端导致“未提供认证令牌”。

变更点：

1. 新增一个可复用的鉴权函数（建议放在 `miniprogram/src/utils/`）
   - 文件：`miniprogram/src/utils/merchant_auth_guard.ts`（新建）
   - 功能：
     - 读取 `uni.getStorageSync('token')` 判断是否有商家 token
     - 若无 token：
       - 计算当前页面路径 + query，拼成 `redirect` 参数
       - `uni.reLaunch({ url: '/pages/auth/login?redirect=...' })`
       - 返回 `false`
     - 若有 token：返回 `true`

2. 在以下商家端页面生命周期入口调用守卫（优先 `onLoad`/`onShow`）
   - `miniprogram/src/pages/merchant/orders/detail.vue`
   - `miniprogram/src/pages/merchant/orders/list.vue`
   - （可选）其它商家端非 Tab 页面：商品编辑、配送设置等（根据扫描入口是否可能直达决定）

### B. 登录页支持 “redirect 回跳”

目的：当用户被守卫带到登录页后，登录成功返回到原页面（例如订单详情），而不是固定跳到商家首页。

变更点：

- 文件：`miniprogram/src/pages/auth/login.vue`
- 改造：
  1. `onLoad(options)` 读取 `options.redirect`（URL 编码）
  2. 登录成功后：
     - 若存在 redirect：
       - 若 redirect 指向 tab 页（如 `/pages/merchant/home`），用 `uni.switchTab`
       - 否则用 `uni.redirectTo` 或 `uni.navigateTo`（根据是否需要保留登录页在栈中决定）
     - 若不存在 redirect：保持现有逻辑 `switchTab('/pages/merchant/home')`

### C. 商家端请求的 401/1002 兜底体验

目的：当 token 过期/无效导致 401 时，统一清理并跳转登录页，避免停留在当前页持续报错。

变更点：

- 文件：`miniprogram/src/utils/request.ts`
- 确认并完善：
  - 对 HTTP 401 且 body `code=1002` 的情况，按 merchantRequest 分支清理 `token/merchantId/staff/merchantInfo` 并 `uni.reLaunch('/pages/auth/login')`。
  - 避免只 toast 不跳转导致用户反复点击仍失败。

### D. 退款链路的前置校验（只做“防呆”，不改业务规则）

目的：减少“已知必失败请求”对用户的干扰，并提供可理解提示。

变更点：

- 文件：`miniprogram/src/pages/merchant/orders/detail.vue`、`list.vue`
- 改造：
  - 在发起退款前先调用鉴权守卫；未登录直接跳转登录页
  - 退款原因为空时可允许，但若后端要求可在前端提示“请输入退款原因”

### E. PRD 同步

目的：把“商家端页面鉴权与跳转规则”固化到 PRD，避免后续再次分叉。

变更点：

- `PRD.md`
  - 在“本轮已同步摘要”补充：商家端订单详情/退款等接口必须携带 merchant token；未登录/失效需跳转登录并支持回跳。
- `docs/prd/PRD-接口文档.md`
  - 在“认证相关接口/商家接口”中补充：
    - 商家端接口统一使用 `Authorization: Bearer {token}`（storage key：`token`）
    - 401/1002 的前端处理：清理登录态 + 跳转登录页
    - 订单详情与退款接口属于商家鉴权接口

## Assumptions & Decisions

- 已确认问题核心是 “商家端 token 未携带/未初始化” 导致的 401/1002。
- 选择“页面级守卫 + 登录回跳”作为最小改动方案，不引入全局路由系统或重构页面结构。
- 不在本次计划中新增复杂的扫码登录协议；只保证“无 token 进入商家端页面时能引导登录并恢复”。

## Verification Steps

1. 以未登录状态直接打开：
   - `/pages/merchant/orders/detail?id=xxx`
   - `/pages/merchant/orders/list`
   期望：自动跳转登录页；登录成功后回到原页面。
2. 已登录状态：
   - 打开订单详情：不出现“未提供认证令牌”
   - 发起退款：请求发出且后端不再返回“未提供认证令牌”
3. 人为模拟 token 失效（重启后端并更换 JWT_SECRET 或手动清理 storage）：
   - 调用任意 `/api/v1/merchant/*` 接口触发 401/1002
   - 期望：清理登录态并跳转登录页
4. 构建验证：
   - 小程序 `build:mp-weixin` 无编译错误

