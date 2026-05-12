# 商家端 WebSocket 手动联调测试计划

## Summary

- 目标：扩展现有 `pages/store/order-sound-test` 为统一的 WebSocket 联调测试页，支持输入指定商家 ID，分别发送“顾客进店提醒”和“订单成功提醒”，用于验证商家端登录后是否成功建立 WebSocket 连接。
- 验收标准：
  - 测试页可手动输入商家 ID 并发送两类通知。
  - 商家端登录并进入页面后，能够看到 WebSocket 连接状态。
  - 商家端收到推送后，除现有 Toast / 提示音外，还能看到最近一条收到的消息及时间。
  - 继续复用后端已有 `/api/v1/dev/order-notify` 与 `/api/v1/dev/store-visit-notify`，不新增后端推送协议。

## Current State Analysis

### 现有后端能力

- `server/cmd/server/main.go`
  - 已注册开发联调接口：
    - `POST /api/v1/dev/order-notify`
    - `POST /api/v1/dev/store-visit-notify`
  - 商家 WebSocket 入口为 `GET /api/v1/ws/merchant`，走 `JWTAuth()` 鉴权。
- `server/internal/handlers/ws/dev.go`
  - 已支持按 `merchant_id` 发送订单提醒和进店提醒。
  - 返回体包含 `delivered`，可用于判断是否有已连接商家端收到推送。
- `server/internal/handlers/ws/events.go`
  - 已固定两类消息协议：
    - `order_notify`
    - `store_visit_notify`

### 现有前端能力

- `miniprogram/src/pages/store/order-sound-test.vue`
  - 当前仅支持输入商家 ID 后调用 `/api/v1/dev/order-notify` 发送订单提醒。
  - 不支持“顾客进店提醒”。
  - 不展示服务端 `delivered` 结果，也不提供更强的调试反馈。
- `miniprogram/src/stores/auth.ts`
  - 已负责商家端 `connectSocket()`、`onSocketOpen()`、`onSocketMessage()`、重连和提示音播放。
  - 当前未暴露连接状态，也未记录最近一条收到的消息。
- `miniprogram/src/pages/merchant/home.vue`
  - 是商家登录成功后的默认主页面，最适合展示“已连接 / 重连中 / 未连接”和最近消息，便于验证登录后 WebSocket 是否成功建立。
- `miniprogram/src/pages.json`
  - 已存在 `pages/store/order-sound-test` 页面和对应调试入口，无需新建独立测试页路由，优先复用现有入口。

### 已确认的实现偏好

- 页面位置：扩展现有测试页。
- 验证方式：商家端展示“连接状态 + 最近消息”。

## Proposed Changes

### 1. 扩展测试页为统一 WS 联调页

- 文件：`miniprogram/src/pages/store/order-sound-test.vue`
- 变更内容：
  - 保留“商家 ID”输入能力。
  - 将单一“发送提醒”改为两类独立操作：
    - 发送顾客进店提醒
    - 发送订单成功提醒
  - 为订单提醒保留 `order_no` 输入项。
  - 为进店提醒补充可选 `visitor_openid` / `source` 输入项，未填时沿用后端默认值。
  - 页面展示接口返回的 `delivered` 结果，明确提示：
    - 已推送给多少个在线连接
    - 若为 `0`，提示“商家端可能未连接 WebSocket 或未登录对应商家”
  - 优化文案，将页面用途明确为“商家端 WebSocket 联调测试”。
- 原因：
  - 该页已具备开发调试入口，改造成本最低。
  - 用户明确希望通过输入商家 ID 手动下发两类消息验证 WebSocket 链路。

### 2. 在认证 Store 中沉淀连接状态与最近消息

- 文件：`miniprogram/src/stores/auth.ts`
- 变更内容：
  - 为 `AuthState` 增加 WebSocket 调试态字段，例如：
    - `socketStatus`: `disconnected | connecting | connected | reconnecting`
    - `lastSocketMessage`: 最近一条收到的消息摘要对象
    - `lastSocketMessageAt`: 最近消息接收时间
  - 在以下节点更新状态：
    - `connectOrderSocket()` 调用前后标记 `connecting`
    - `uni.onSocketOpen()` 标记 `connected`
    - `uni.onSocketError()` 标记 `disconnected`
    - `uni.onSocketClose()` 在登录态下标记 `reconnecting`，否则标记 `disconnected`
    - `disconnectOrderSocket()` 主动断开时标记 `disconnected`
  - 在 `uni.onSocketMessage()` 中统一抽取最近消息记录逻辑：
    - 记录 `type`
    - 记录核心 payload 摘要，如 `merchant_id`、`order_no`、`visitor_openid`、`source`
    - 记录本地接收时间
  - 保留现有 Toast / 提示音逻辑，不改变推送消费行为。
- 原因：
  - 当前只能“听到声音/看到 Toast”，难以直观看到连接生命周期和最近一条消息。
  - 把状态沉淀到 Store，便于多个商家端页面复用，不重复造逻辑。

### 3. 在商家首页展示 WebSocket 连接反馈

- 文件：`miniprogram/src/pages/merchant/home.vue`
- 变更内容：
  - 在顶部卡片或“今日概览”附近增加一个轻量调试信息区，仅展示：
    - 当前 WebSocket 状态
    - 最近消息类型
    - 最近消息时间
    - 最近消息关键摘要
  - 状态文案建议：
    - 已连接
    - 连接中
    - 重连中
    - 未连接
  - 展示逻辑尽量轻量，不影响原有工作台主流程。
- 原因：
  - 这是商家登录成功后的主入口，最符合“验证登录后是否成功连接 WebSocket”的核心诉求。
  - 用户无需切换到设置页或其他页面，即可立即确认链路状态。

### 4. 更新调试入口与页面配置文案

- 文件：`miniprogram/src/pages.json`
- 变更内容：
  - 视页面文案修改情况，同步更新 `pages/store/order-sound-test` 的导航标题与 `condition.list` 中的入口名称。
  - 命名建议调整为“WebSocket 联调测试”或“商家 WS 测试”。
- 原因：
  - 现有“订单提醒测试”已不能准确覆盖页面用途。

### 5. 同步 PRD 中的开发联调说明

- 文件：`PRD.md`
- 变更内容：
  - 在开发联调 / 测试入口说明中补充：
    - 已提供前端联调页面，可输入商家 ID 发送进店提醒与订单提醒
    - 商家首页可查看 WebSocket 连接状态与最近消息
  - 保持与现有 `/api/v1/dev/order-notify`、`/api/v1/dev/store-visit-notify` 说明一致。
- 原因：
  - 本次属于测试能力增强，按项目规则需同步更新 PRD，避免文档与实际功能脱节。

## Assumptions & Decisions

- 决策：不新增后端接口，直接复用现有开发联调接口。
- 决策：不改动 WebSocket 消息协议结构，仍以 `order_notify` / `store_visit_notify` 为准。
- 决策：测试发送页放在 `pages/store/order-sound-test.vue`，不将“发送调试消息”能力直接加入正式商家首页。
- 决策：商家首页只展示轻量级连接反馈，不展示复杂调试面板，避免干扰经营主界面。
- 决策：最近消息仅保留最后一条，避免为调试历史记录引入额外状态复杂度。
- 假设：开发联调环境使用非 `release` 模式，因此后端 `/api/v1/dev/*` 可正常访问。
- 假设：商家端当前登录链路已经可以正常拿到 JWT 并建立 `/api/v1/ws/merchant` 连接。

## Verification Steps

### 手动验证

1. 使用商家账号登录，进入 `pages/merchant/home`。
2. 确认页面展示 WebSocket 状态为“已连接”或短暂“连接中”后进入“已连接”。
3. 打开 `pages/store/order-sound-test`（改造后页面）。
4. 输入与当前登录商家一致的 `merchant_id`。
5. 点击“发送顾客进店提醒”：
   - 页面提示发送成功，并展示 `delivered >= 1`
   - 商家首页最近消息更新为 `store_visit_notify`
   - 若开启浏览提示音，则播放提示音并出现 Toast。
6. 点击“发送订单成功提醒”：
   - 页面提示发送成功，并展示 `delivered >= 1`
   - 商家首页最近消息更新为 `order_notify`
   - 若开启订单提示音，则播放提示音并出现 Toast。
7. 切换为未登录或错误商家 ID 再次发送：
   - 页面展示 `delivered = 0`
   - 用于验证“未连上 / 非目标商家”时不会误推。

### 回归检查

- 商家端正常登录、退出登录不受影响。
- 现有订单提示音与浏览提示音设置开关保持原行为。
- WebSocket 断线重连逻辑保持原有行为，仅新增可视状态反馈。
- `GetDiagnostics` 检查新增与修改文件无明显诊断错误。
