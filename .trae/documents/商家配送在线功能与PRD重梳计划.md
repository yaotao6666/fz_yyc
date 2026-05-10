# 商家配送在线功能与PRD重梳计划

## Summary

本次计划聚焦 5 个方向，并以 `PRD.md` 同步为最终收口：

1. 修正商家配送设置与 C 端下单距离选择链路，明确为“商家配置距离档位 + 用户手动选择”，不接入地图定位。
2. 固化商家在线与营业状态口径，保证“登录即在线、休息时禁止新下单但继续接收浏览提醒”。
3. 落地声音提醒管理、修改密码、员工微信快捷登录与登录页快捷登录入口。
4. 以轻量方案补齐员工管理页面，首版仅店主管理。
5. 结合已接入的用户行为事件，修正商家数据看板，并把以上能力完整同步进 `PRD.md`。

补充约束：

1. 本轮取消商家发券、营销配置相关改动。
2. `mp3` 静态音频文件由用户单独提供，本次不包含音频文件生成方案。
3. 修改密码成功后立即退出并要求重新登录。
4. 员工管理首版不引入细粒度权限体系，仅按 `owner` 收口。

## Current State Analysis

### 1. 现有代码已部分推进，但实现与文档仍未完全对齐

- `miniprogram/src/pages/merchant/settings.vue` 已存在“修改密码”“微信快捷登录”“声音提醒管理”“员工管理”入口形态。
- `miniprogram/src/stores/auth.ts` 已能区分 `order_notify` 与 `store_visit_notify`，并准备播放 `/static/sounds/order.mp3`、`/static/sounds/browse.mp3`。
- `miniprogram/src/pages/merchant/notification-settings.vue`、`miniprogram/src/pages/merchant/staff/index.vue` 已存在，说明前端页面骨架已落地。
- `server/internal/handlers/merchant/account.go`、`server/internal/handlers/merchant/staff.go`、`server/internal/handlers/ws/events.go`、`server/internal/handlers/ws/dev.go` 已承接账号、员工、提醒相关后端能力。
- `PRD.md` 仍停留在旧口径，当前只描述了“新订单声音提醒”单一开关和旧版员工字段，未覆盖浏览提醒、微信快捷登录、修改密码、登录即在线等新规则。

### 2. 配送设置页面真实存在，但当前仍缺少完整闭环

- `miniprogram/src/pages/merchant/delivery-settings.vue` 已对接 `getDeliverySettings` / `updateDeliverySettings`，但页面保存前没有完整的前端区间校验。
- `PRD.md` 已有 `3.3.5 配送设置（获取/更新）` 和配送规则说明，但描述仍偏“按距离计费规则”，没有强调“前端手选档位、不做定位、不接地图”。
- 结合现有需求，配送距离只是商家配置的服务范围选项，不是实际定位距离，C 端只需在超出商家支持范围时给出提示。

### 3. 在线状态、浏览提醒与营业状态的业务口径已变更

- `PRD.md` 当前在 `3.3.12.1 商家 WebSocket 连接（新订单提醒）` 只描述新订单消息，没有覆盖浏览提醒事件。
- `miniprogram/src/stores/auth.ts` 已按“登录后保持 WebSocket 在线”实现方向推进，但 `PRD.md` 还未明确“在线状态与营业状态解耦”。
- 根据已确认需求，商家休息时只关闭接单，不关闭在线连接；用户扫码浏览店铺仍然推送浏览提醒。
- 已支付中的订单不应因为商家中途切换到休息而失败，文档也需要明确这一边界。

### 4. 账号安全、微信快捷登录与员工管理已有真实落点

- `miniprogram/src/pages/auth/login.vue`、`miniprogram/src/stores/auth.ts`、`server/internal/handlers/merchant/account.go` 已形成快捷登录实现入口。
- `PRD.md` 当前缺少商家员工维度的微信绑定、快捷登录异常提示、修改密码成功后退出重登的说明。
- `PRD.md` 员工表章节当前只列出了 `notify_enabled`，未覆盖 `browse_notify_enabled`、`wechat_bound_at`、`last_wechat_login_at` 等字段。
- `server/cmd/server/main.go`、`server/internal/handlers/merchant/staff.go` 已具备员工 CRUD 基础，适合在文档中明确首版能力边界。

### 5. 看板页面与埋点链路正在接通，但 `PRD` 指标口径未同步

- `miniprogram/src/pages/merchant/analytics/index.vue` 已展示销售额、订单数、客户数、访客数、访问次数、支付人数等指标。
- `miniprogram/src/utils/analytics.ts`、`server/internal/handlers/user/handler.go`、`server/internal/handlers/merchant/order.go` 已是行为事件与分析数据的主要落点。
- `server/internal/models/models.go`、`server/pkg/database/mysql.go`、`server/migrations/20240101000000_full_init.sql` 已引入 `user_behavior_events` 方向的数据结构。
- `PRD.md` 目前仍以旧版看板口径为主，尚未把访客、浏览、下单、支付成功等事件来源写清楚。

## Proposed Changes

### 1. 配送设置与下单链路

#### 目标文件

- `miniprogram/src/pages/merchant/delivery-settings.vue`
- `miniprogram/src/pages/store/confirm.vue`
- `miniprogram/src/api/index.ts`
- `miniprogram/src/types/index.ts`
- `server/internal/handlers/merchant/handler.go`
- `server/internal/handlers/user/handler.go`
- `PRD.md`

#### 实施内容

1. 在 `delivery-settings.vue` 中补充保存前校验，保证基础配送费、满免金额、最大距离、区间顺序、区间不重叠、区间不超过最大距离。
2. 在 `server/internal/handlers/merchant/handler.go` 中保持与前端一致的规则校验，避免前端绕过后写入非法配置。
3. 在 `store/confirm.vue` 中改为基于接口返回的 `distance_rules` 生成选择项，不再按固定 `0.5km` 步长和写死配送费计算。
4. 在 `server/internal/handlers/user/handler.go` 中确保：
   - 商家休息时拒绝新的创建订单请求。
   - 已创建订单的后续支付成功不受营业状态变化影响。
5. 在 `PRD.md` 中同步“商家自定义档位”“用户手动选择”“超出范围仅提示”“不接入第三方地图”的业务说明、接口说明和验收口径。

### 2. 在线状态、WebSocket 与声音提醒

#### 目标文件

- `miniprogram/src/stores/auth.ts`
- `miniprogram/src/pages/merchant/settings.vue`
- `miniprogram/src/pages/merchant/notification-settings.vue`
- `miniprogram/src/pages/merchant/home.vue`
- `miniprogram/src/pages.json`
- `miniprogram/src/api/index.ts`
- `miniprogram/src/types/index.ts`
- `server/internal/handlers/user/handler.go`
- `server/internal/handlers/ws/events.go`
- `server/internal/handlers/ws/dev.go`
- `PRD.md`

#### 实施内容

1. 固化 `auth.ts` 的连接策略：登录成功即连接、恢复登录态自动重连、仅退出登录时断开。
2. 继续区分两种提醒事件：
   - `order_notify`
   - `store_visit_notify`
3. 声音提醒只控制本地播放，不控制 WebSocket 在线状态。
4. 在 `notification-settings.vue` 中承接：
   - 下单提醒开关
   - 浏览提醒开关
   - 下单提醒测试按钮
   - 浏览提醒测试按钮
5. 在 `settings.vue` 中把原单一提示音入口统一收口为“声音提醒管理”。
6. 在 `server/internal/handlers/user/handler.go` 中确保用户进入店铺首页时可触发浏览提醒广播，并允许休息中的商家店铺被浏览。
7. 在 `PRD.md` 中重写相关章节，明确：
   - 登录即在线
   - 休息中不断开连接
   - 休息中继续接收浏览提醒
   - 休息中禁止新下单
   - `mp3` 资源由静态文件提供

### 3. 修改密码与微信快捷登录

#### 目标文件

- `miniprogram/src/pages/merchant/settings.vue`
- `miniprogram/src/pages/auth/login.vue`
- `miniprogram/src/stores/auth.ts`
- `miniprogram/src/api/index.ts`
- `miniprogram/src/types/index.ts`
- `server/internal/handlers/merchant/account.go`
- `server/internal/handlers/merchant/handler.go`
- `PRD.md`

#### 实施内容

1. 在 `settings.vue` 中完善修改密码弹窗校验与提交成功后的退出重登行为。
2. 在 `settings.vue` 中承接员工维度的微信绑定/解绑展示与操作。
3. 在 `login.vue` 中提供微信快捷登录入口，并复用 `auth.ts` 中的快捷登录逻辑。
4. 在 `server/internal/handlers/merchant/account.go` 中保持以下业务口径：
   - 绑定关系归属当前登录员工
   - 未绑定商家账号的微信快捷登录返回“您还不是商家，请注册后使用”
5. 在 `PRD.md` 中补充账号安全和快捷登录相关页面说明、接口说明、异常提示与验收标准。

### 4. 轻量员工管理

#### 目标文件

- `miniprogram/src/pages/merchant/settings.vue`
- `miniprogram/src/pages/merchant/staff/index.vue`
- `miniprogram/src/pages.json`
- `miniprogram/src/api/index.ts`
- `miniprogram/src/types/index.ts`
- `server/cmd/server/main.go`
- `server/internal/handlers/merchant/staff.go`
- `server/internal/models/models.go`
- `server/migrations/20240101000000_full_init.sql`
- `PRD.md`

#### 实施内容

1. 以 `pages/merchant/staff/index.vue` 为首版员工管理页，承接列表、新增、编辑、删除、重置密码。
2. 页面和操作入口统一按 `owner` 收口，非店主不展示入口且禁止直接进入。
3. 保持后端已有保护规则，并在必要处补齐：
   - 不能删除当前登录账号
   - 不能删除最后一个 `owner`
   - 不能跨商家操作
4. 在 `PRD.md` 中同步首版员工管理范围，并更新员工表字段与相关接口入参/出参示例。
5. 明确新增或已扩展字段：
   - `merchant_staffs.browse_notify_enabled`
   - `merchant_staffs.wechat_bound_at`
   - `merchant_staffs.last_wechat_login_at`
   - 继续复用 `merchant_staffs.openid` 作为微信绑定标识

### 5. 用户行为事件与数据看板

#### 目标文件

- `miniprogram/src/pages/merchant/analytics/index.vue`
- `miniprogram/src/pages/merchant/home.vue`
- `miniprogram/src/pages/store/home.vue`
- `miniprogram/src/pages/store/product.vue`
- `miniprogram/src/pages/store/confirm.vue`
- `miniprogram/src/utils/analytics.ts`
- `miniprogram/src/api/index.ts`
- `miniprogram/src/types/index.ts`
- `server/internal/handlers/user/handler.go`
- `server/internal/handlers/merchant/order.go`
- `server/internal/models/models.go`
- `server/pkg/database/mysql.go`
- `server/migrations/20240101000000_full_init.sql`
- `PRD.md`

#### 实施内容

1. 以 `user_behavior_events` 为最小行为事件表，统一承接：
   - `store_visit`
   - `page_view`
   - `product_view`
   - `submit_order`
   - `pay_success`
2. 让 `miniprogram/src/utils/analytics.ts` 统一走真实上报，不再保留日志占位路径。
3. 对齐 `analytics/index.vue` 与后端概览、趋势、排行、客户分析字段，保证页面展示可信。
4. 在 `merchant/home.vue` 中与分析页复用相同关键概览口径，避免首页卡片与看板口径不一致。
5. 在 `PRD.md` 中增加看板指标来源说明，写明本轮以订单数据和行为事件为主，不扩展复杂 BI 功能。

### 6. PRD 同步策略

#### 目标文件

- `PRD.md`

#### 实施内容

1. 更新商家设置相关章节，补入：
   - 声音提醒管理子页面
   - 修改密码弹窗
   - 微信快捷登录绑定/解绑
   - 员工管理入口
2. 更新 WebSocket 与提醒章节，补入浏览提醒事件和测试推送接口。
3. 更新员工接口、员工表示例和字段说明。
4. 更新配送设置与用户下单章节，写清前端选择距离档位的规则。
5. 更新数据看板与行为埋点章节，补齐访客和行为事件来源。
6. 明确本轮不包含优惠券、营销配置和地图定位。

## Assumptions & Decisions

### 已锁定决策

1. 商家登录即在线，WebSocket 在线状态与营业状态解耦。
2. 商家休息时仅关闭新下单，不影响浏览提醒和已进入支付流程的订单。
3. 声音提醒拆分为“用户下单提醒”“用户浏览提醒”两个独立开关。
4. `mp3` 文件由用户单独提供，本次不实现音频生成。
5. 微信快捷登录绑定在员工维度，不新增独立绑定关系表。
6. 员工管理首版仅店主可管理。
7. 修改密码成功后立即退出重新登录。
8. 本轮不做优惠券、分享券与其他营销能力。
9. 配送距离为商家自定义服务范围选项，不接入第三方地图。

### 本轮不纳入范围

1. 商家发券、优惠券分享、营销活动配置。
2. 员工细粒度权限系统。
3. 地图 SDK、真实距离计算、地址定位能力。
4. 高级数据分析、漏斗分析、留存分析。
5. `mp3` 音频文件制作或格式转换。

## Verification Steps

### 1. 配送链路

1. 商家可加载并保存配送规则，非法区间会被前后端一致拦截。
2. C 端确认页距离选项来自商家真实配置，不再是写死档位。
3. 超出配送支持范围时只提示并阻止提交，不触发地图定位逻辑。
4. 商家休息时创建订单被拒绝。
5. 支付中的订单在商家切休息后仍可支付成功并看到成功提示。

### 2. 在线与提醒

1. 商家登录后自动建立 WebSocket 连接，退出登录后断开。
2. 商家切换为休息时连接不断开。
3. 用户进入商家主页时，商家可收到浏览提醒。
4. 下单提醒与浏览提醒可独立开关并独立测试播放。
5. 运行时只依赖静态 `mp3` 文件，不再涉及生成逻辑。

### 3. 账号与员工

1. 修改密码成功后立即退出并返回登录页。
2. 已绑定员工可使用微信快捷登录，未绑定时提示“您还不是商家，请注册后使用”。
3. 设置页可查看绑定状态、绑定时间，并执行绑定/解绑。
4. 仅店主可见员工管理入口并执行员工管理操作。
5. 员工列表、新增、编辑、删除、重置密码链路完整可用。

### 4. 数据看板

1. 分析页概览、趋势、排行、客户与访客指标字段映射正确。
2. 访问、页面浏览、商品浏览、提交订单、支付成功等事件能落库。
3. 首页卡片与分析页关键指标口径一致。

### 5. 文档一致性

1. `PRD.md` 已同步本轮所有保留需求与表字段变化。
2. `PRD.md` 明确本轮不包含优惠券和营销配置。
3. `PRD.md` 明确 `mp3` 文件由静态资源提供，不包含生成方案。
