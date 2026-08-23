# 商家支付提醒、订单号与支付页金额修复计划

## Summary
- 修复支付成功后商家端未收到 WebSocket 新订单提醒、未播放提示音的问题。
- 优化 C 端订单号生成规则，在现有时间戳基础上加入商家 ID 和随机数，降低并发重复风险。
- 修复小程序 `pages/store/confirm` 提交订单后拉起支付时页面金额变成 `0` 的问题，避免因购物车清空时机过早导致支付过程体验异常。
- 本次实现保持现有接口形状不变，优先做最小改动；如涉及接口返回说明变化，实施后同步更新 PRD。<mccoremem id="01KR9503FF55ZB6H0GKGSFSKAN|03g4hsie9a7ftzb2b6sbsmhqx|01KRF570PP87P9QM44KXRM780W" />

## Current State Analysis

### 1. 支付成功后商家端没有收到声音
- 商家端 WebSocket 事件定义在 `server/internal/handlers/ws/events.go`，已有：
  - `BroadcastOrderNotify(merchantID, orderNo)`
  - `BroadcastStoreVisitNotify(merchantID, visitorOpenID, source)`
- 顾客进店埋点 `server/internal/handlers/user/handler.go` 在 `TrackVisit` 路径中已经调用 `BroadcastStoreVisitNotify(...)`，因此“扫码进店提醒”链路是通的。
- 支付成功回调处理在 `server/internal/handlers/sp/payment.go` 的 `PaymentNotify()` / `processPaymentSuccess()`。
- 当前 `processPaymentSuccess()` 只更新订单支付状态和记录，没有调用 `BroadcastOrderNotify(...)`，因此商家端收不到“订单成功提醒”消息。
- 商家小程序播放声音逻辑在 `miniprogram/src/stores/auth.ts`：
  - 收到 `order_notify` 时播放 `/static/sounds/order.mp3`
  - 收到 `store_visit_notify` 时播放 `/static/sounds/browse.mp3`
  - 声音开关分别受 `notify_enabled` / `browse_notify_enabled` 控制
- 结论：第一项的核心缺口在后端支付成功回调未广播 `order_notify`，而不是前端播音能力缺失。

### 2. 订单号存在重复风险
- 用户下单接口在 `server/internal/handlers/user/handler.go` 的 `CreateOrder()` 中调用 `utils.GenerateOrderNo()`。
- 当前 `GenerateOrderNo()` 位于 `server/internal/utils/pagination.go`，规则为：
  - `yyyyMMddHHmmss` 14 位时间戳
  - `now.UnixNano()%10000` 4 位随机尾数
- 当前订单模型 `server/internal/models/models.go` 中 `orders.order_no` 长度为 `size:32`、唯一索引。
- 现有规则没有包含 `merchant_id`，并且随机段只有 4 位，在高并发或同秒下存在碰撞概率。

### 3. 提交订单后支付页显示 0 元
- 确认页逻辑在 `miniprogram/src/pages/store/confirm.vue`。
- 当前 `submitOrder()` 流程是：
  1. 调用 `createOrder(...)`
  2. 接口成功返回后立即执行 `cartStore.clearCart()`
  3. 再调用 `uni.requestPayment(...)`
- 金额展示依赖 `cartStore.totalAmount` 和基于其计算的 `deliveryFee`、`totalAmount`。
- 购物车状态在 `miniprogram/src/stores/cart.ts` 中持久化到本地存储；`clearCart()` 会清空内存和本地缓存。
- 结论：当前金额变 0 的直接原因，就是在支付真正完成前提前清空了购物车，页面仍停留在确认页上下文中时所有金额计算都归零。

## Proposed Changes

### A. 补齐支付成功后的商家 WebSocket 订单提醒

#### 目标
- 支付回调确认成功后，向对应商家广播一条 `order_notify` 消息。
- 商家端已在线并开启订单提示音时，能够像联调页 `pages/store/order-sound-test` 那样收到新订单提醒并播放 `order.mp3`。

#### 修改文件
- `server/internal/handlers/sp/payment.go`
- `server/internal/handlers/ws/events.go`（仅在需要补充 payload 字段时才改，默认优先复用现有事件）

#### 方案
- 在 `processPaymentSuccess()` 内，订单支付状态更新成功后、事务提交成功后再触发 `BroadcastOrderNotify(order.MerchantID, order.OrderNo)`。
- 采用“事务内只做数据更新，事务成功后做推送”的方式，避免数据库回滚时却提前发出成功提醒。
- 推送保持幂等友好：
  - 若订单已是已支付状态，仍允许在首次处理成功路径上只广播一次；
  - 若未来发现微信重复回调会导致重复提醒，再在当前方案上补充“仅从 `<2` 变为 `2` 时广播”的保护判断。

#### 关键注意点
- 不改前端消息协议，继续使用现有 `order_notify` 结构，最小化前端变更面。
- 如需减少重复提醒风险，可在回调逻辑中引入布尔标记，显式记录本次是否发生“未支付 -> 已支付”的状态跃迁，并据此决定是否广播。

### B. 调整订单号生成规则

#### 目标
- 新订单号同时包含时间戳、商家 ID、随机段，满足用户“增加商家ID和随机数，避免重复”的要求。
- 保持总长度不超过 `orders.order_no` 的 `32` 字符限制。

#### 修改文件
- `server/internal/utils/pagination.go`
- `server/internal/handlers/user/handler.go`
- 如需补充单元测试，则新增或扩展 `server/internal/utils/*_test.go`

#### 方案
- 将 `GenerateOrderNo()` 调整为接收 `merchantID uint64` 参数，例如：
  - `时间戳14位 + 商家ID + 6位随机数`
- 在 `CreateOrder()` 中改为 `utils.GenerateOrderNo(req.MerchantID)`。
- 实现时优先保证：
  - 仅使用数字，避免影响微信支付 `out_trade_no` 兼容性
  - 总长度稳定不超过 32 位
  - 随机段从当前 4 位提升到更稳妥的 6 位

#### 建议格式
- 默认采用：`yyyyMMddHHmmss + merchantID + 6位随机数`
- 说明：
  - 时间戳 14 位
  - 商家 ID 直接拼接，便于排查来源
  - 6 位随机数显著优于当前 4 位
- 若拼接后可能超 32 位，则在实现时对商家 ID 部分做上限裁剪或保留后若干位，但优先保持“可读地包含商家ID”的目标。

### C. 修复确认页支付过程中的 0 元显示

#### 目标
- 在拉起支付和支付结果返回前，确认页金额、商品列表、配送费保持稳定，不因购物车被提前清空而变为 0。
- 支付成功后再清空购物车；支付取消或支付失败时保留购物车，方便用户继续支付或修改订单。

#### 修改文件
- `miniprogram/src/pages/store/confirm.vue`
- `miniprogram/src/stores/cart.ts`（仅在需要增加“订单提交中快照/恢复能力”时才改，默认优先只改页面提交流程）

#### 方案
- 重排 `submitOrder()` 的购物车清理时机：
  - `createOrder()` 成功后，不立即执行 `cartStore.clearCart()`
  - 若 `res.pay_params` 存在：
    - `uni.requestPayment.success` 后再清空购物车
    - `fail/cancel` 时不清空购物车
  - 若无需支付（0 元订单）：
    - 订单创建成功后直接清空购物车，再跳转订单页
- 为避免支付中重复点击造成二次下单，补充一个页面级提交锁，例如 `submitting/ref`，并同步禁用提交按钮或忽略重复点击。
- 如支付过程中页面仍需展示稳定金额，可优先复用 `cartStore` 现有数据；若实际验证仍会因页面切换/重进导致状态抖动，再补一层提交快照状态，但第一版先不扩大复杂度。

#### 关键注意点
- 取消支付后的订单已经创建，用户会被引导到待支付订单页，因此保留购物车可能带来“重复提交同一购物车”的风险。
- 为平衡体验与最小改动，本次先按用户指出的问题处理为“取消/失败不清空购物车”，同时通过提交锁降低连续重复提交概率。
- 实施时需要额外检查：
  - 从确认页跳转到我的订单待支付列表后，返回首页是否会因购物车仍在而造成误导；
  - 如有明显体验问题，再在第二步增加“根据支付结果区分是保留购物车还是引导清空”的细化逻辑。

## Assumptions & Decisions
- 默认不改动现有 WebSocket 消息协议，直接复用 `order_notify`。
- 默认不新增数据库字段，订单提醒是否广播由当前回调状态变化判断解决。
- 订单号保持数字串，兼容当前支付链路和数据库长度限制。
- 本次先不重构购物车存储模型，优先通过调整 `clearCart()` 时机解决 0 元显示问题，保持改动面最小。
- 如实现后涉及订单号规则、支付提醒链路、确认页支付交互文档，应同步更新：
  - 根 `PRD.md`
  - `docs/prd/PRD-功能说明.md`
  - `docs/prd/PRD-接口文档.md`
  - `docs/prd/PRD-测试与附录.md` <mccoremem id="01KR9503FF55ZB6H0GKGSFSKAN|03g4hsie9a7ftzb2b6sbsmhqx|01KRF570PP87P9QM44KXRM780W" />

## Verification Steps

### 1. 支付成功提醒
- 商家端登录并保持首页 WebSocket 已连接。
- 用户端正常下单并完成微信支付。
- 验证商家端：
  - 首页最近消息收到 `order_notify`
  - 显示订单号
  - 订单提示音开关开启时播放声音
- 额外验证：
  - 重复支付回调不应造成明显重复提醒
  - 顾客进店提醒原链路不受影响

### 2. 订单号规则
- 创建多个不同商家的订单，核对 `order_no`：
  - 包含时间戳
  - 包含商家 ID
  - 包含随机段
  - 长度不超过 32
- 回归微信支付回调、订单查询、退款等所有依赖 `order_no` 的链路，确保兼容。

### 3. 确认页 0 元问题
- 在 `pages/store/confirm` 提交订单并拉起支付时，确认页金额不应归零。
- 支付成功后：
  - 购物车被清空
  - 跳转到订单列表/订单详情后状态正常
- 支付取消或失败后：
  - 购物车仍保留
  - 返回确认页或重新发起支付时金额仍正确

### 4. 基础检查
- 后端：
  - `go test ./...` 或至少覆盖 `internal/handlers/user/...`、`internal/handlers/sp/...`
- 小程序：
  - `GetDiagnostics` 检查改动文件
  - `npm run build:mp-weixin`

