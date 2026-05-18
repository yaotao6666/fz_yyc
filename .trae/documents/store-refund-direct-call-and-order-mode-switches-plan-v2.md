# 用户端退款直拨与商家下单方式开关优化计划

## Summary

* 将用户端 `pages/store/order-detail` 的“联系商家退款”改为点击后直接拉起系统拨号，不再二次确认。

* 将商家端下单方式统一为三项独立开关：配送、堂食、自提，其中配送继续沿用现有配送设置结构，堂食/自提走商家主表配置。

* 让商家设置接口、配送设置接口、C 端确认页、服务端下单校验、SQL 初始化数据与 PRD 文档对三种下单方式保持一致。

## Current State Analysis

### 1. 用户端退款联系行为

* `miniprogram/src/pages/store/order-detail.vue` 中的 `contactMerchantForRefund()` 已被部分改成直接拨号逻辑，当前优先读取 `order.value.merchant?.contact_phone`，再回退 `phone`。

* 但 `miniprogram/src/types/index.ts` 里的 `OrderMerchant` 仍只有 `phone?: string`，尚未补 `contact_phone?: string`，执行阶段会产生类型报错。

* 后端 `server/internal/handlers/user/handler.go` 的 `GetOrderDetail()` / `GetOrders()` 直接返回预加载后的 `Merchant`，运行时已经具备 `contact_phone` 字段，因此前端只需补类型与归一化口径即可。

### 2. 商家三种下单方式的当前分裂口径

* `server/internal/models/models.go` 中已有：

  * `takeout_enabled`

  * `dine_in_enabled`

  * 已新增但未打通的 `pickup_enabled`

* `server/internal/handlers/merchant/handler.go`

  * `GetSettings()` 只返回 `takeout_enabled`、`dine_in_enabled`

  * `UpdateSettings()` 只支持更新这两个字段

* `server/internal/handlers/merchant/handler.go` 的 `GetDeliverySettings()` / `UpdateDeliverySettings()` 仍只管理配送表 `merchant_delivery_settings` 的：

  * `enabled`

  * `base_fee`

  * `free_delivery_amount`

  * `max_distance`

  * `distance_rules`

* 这意味着“配送是否支持”目前依赖 `merchant_delivery_settings.enabled`，“堂食是否支持”依赖商家主表，自提还没有完整闭环。

### 3. C 端确认页的数据来源与缺口

* `miniprogram/src/pages/store/confirm.vue` 当前固定渲染三种方式：

  * 配送

  * 堂食

  * 自提

* 该页面当前只请求 `getStoreDeliveryRules()`，不会请求 `getStoreProducts()` 获取商家方式开关。

* `server/internal/handlers/user/handler.go` 的 `GetDeliveryRules()` 目前只返回配送规则，不返回 `dine_in_enabled` / `pickup_enabled`。

* 因此确认页若要按商家开关动态展示，下单方式元数据应直接并入 `GET /api/v1/store/:merchant_id/delivery-rules`，避免为 3 个布尔值额外增加一次请求。

### 4. 服务端下单校验缺口

* `server/internal/handlers/user/handler.go` 的 `CreateOrder()` 当前只对 `delivery_type = 1` 校验：

  * 配送地址

  * 联系人

  * 联系电话

  * 配送距离

  * 配送范围和配送配置

* 对 `delivery_type = 2`、`delivery_type = 3` 尚未校验商家是否已关闭堂食或自提。

* 这会导致前端即使隐藏了某些方式，服务端仍可能接受错误下单方式。

### 5. SQL 初始化数据当前不一致

* `server/migrations/20240101000000_full_init.sql` 已增加 `pickup_enabled` 字段与插入列名。

* 但初始化 `merchants` 示例数据的 `VALUES` 尚未补齐对应值，当前列和值顺序已错位，需要执行阶段修正。

* `server/migrations/20260511120000_mock_seed.sql` 的 `merchants` 插入语句仍未增加 `pickup_enabled` 列，也需要一并补齐。

## Assumptions & Decisions

* “直接拨打商家电话”按字面执行，不保留二次确认弹窗。

* 用户端订单详情优先使用 `merchant.contact_phone`，兼容回退 `merchant.phone`，避免历史数据或旧接口结构导致拨号失败。

* 三种方式的最终业务语义固定为：

  * `merchant_delivery_settings.enabled`: 是否支持配送下单

  * `dine_in_enabled`: 是否支持堂食下单

  * `pickup_enabled`: 是否支持自提下单

* 配送资费和距离规则仍保留在 `merchant_delivery_settings` 表中，不做表结构重构。

* `takeout_enabled` 视为遗留字段，本次不再扩展其业务语义，也不作为确认页和下单校验的判断来源。

* 确认页的下单方式元数据统一从 `GET /api/v1/store/:merchant_id/delivery-rules` 获取，不新增额外请求。

* 商家端三种方式开关继续集中放在 `pages/merchant/delivery-settings.vue`，不新增页面入口。

* `pages/merchant/settings.vue` 的“配送设置”摘要同步升级为三种方式摘要，避免入口文案与实际能力不一致。

* 本次是产品功能、接口和数据结构变更，必须同步更新 `PRD.md` 及 `docs/prd` 子文档。

## Proposed Changes

### A. 用户端订单详情改为直接拨号

#### 目标

* 点击“联系商家退款”时，若存在商家电话，直接拉起拨号。

* 若电话缺失或系统拨号失败，给出单次兜底提示，不影响页面其他行为。

#### 修改文件

* `miniprogram/src/pages/store/order-detail.vue`

* `miniprogram/src/types/index.ts`

* `miniprogram/src/api/index.ts`

#### 实施方式

* 在 `OrderMerchant` 中新增 `contact_phone?: string`。

* 保持 `normalizeOrder()` 对 `merchant` 的透传，不额外改字段名，只补类型兼容。

* `contactMerchantForRefund()` 使用：

  * `order.value.merchant?.contact_phone`

  * 回退 `order.value.merchant?.phone`

* 有号码时直接执行 `uni.makePhoneCall()`。

* 无号码时保留现有“请联系商家协助处理退款”兜底提示。

* `makePhoneCall` 失败时仅提示“请联系商家退款”，不再回退二次弹窗。

### B. 后端补齐自提开关与三种下单方式口径

#### 目标

* 后端对配送、堂食、自提三种方式形成统一的展示和校验口径，其中配送继续复用现有配送配置结构。

* 现有商家默认不受影响，升级后保留全开状态。

#### 修改文件

* `server/internal/models/models.go`

* `server/pkg/database/mysql.go`

* `server/internal/handlers/merchant/handler.go`

* `server/internal/handlers/user/handler.go`

* `server/migrations/20240101000000_full_init.sql`

* `server/migrations/20260511120000_mock_seed.sql`

#### 实施方式

* 保留 `models.Merchant.PickupEnabled` 当前新增结果，继续作为正式字段使用。

* `ensureMerchantColumns()` 保持补列逻辑，确保历史数据库自动补齐 `pickup_enabled`。

* `GetSettings()` 增加返回 `pickup_enabled`，保留现有 `dine_in_enabled`。

* `UpdateSettingsRequest` 增加 `PickupEnabled *bool`。

* `UpdateSettings()` 将 `pickup_enabled` 纳入 `merchantUpdates`，`dine_in_enabled` 维持现有更新逻辑。

* `GetProducts()` 的 `merchant` 元数据中补充 `pickup_enabled`，保留现有 `dine_in_enabled`；不再把配送能力绑定到 `takeout_enabled`。

* `GetDeliveryRules()` 响应新增三种方式元数据：

  * `dine_in_enabled`

  * `pickup_enabled`

* `GetDeliveryRules()` 在无配送设置记录时，仍返回完整结构，至少包含：

  * `enabled: false`

  * 堂食/自提布尔值

  * 空 `distance_rules`

  * `base_fee/free_delivery_amount/max_distance` 默认值

* `CreateOrder()` 增加方式校验分支：

  * `delivery_type = 1` 时，仅要求配送设置存在并 `Enabled = true`

  * `delivery_type = 2` 时，要求 `merchant.DineInEnabled = true`

  * `delivery_type = 3` 时，要求 `merchant.PickupEnabled = true`

* `CreateOrder()` 对关闭方式返回明确提示：

  * `商家暂未开启配送`

  * `商家暂未开启堂食`

  * `商家暂未开启自提`

* 对非配送方式继续将配送相关字段清空，保持已有订单数据结构不变。

* 修正 `20240101000000_full_init.sql` 的 `merchants` 初始化样例值顺序，给 `pickup_enabled` 补真实布尔值。

* 修正 `20260511120000_mock_seed.sql` 的 `merchants` 列表和每行数据，补齐 `pickup_enabled`。

### C. 前端类型与 API 口径补齐

#### 目标

* 小程序类型系统能完整表达“配送沿用原结构、堂食/自提由主表控制”的三种下单方式。

* 确认页只通过一条接口即可拿到配送规则和方式开关。

#### 修改文件

* `miniprogram/src/types/index.ts`

* `miniprogram/src/api/index.ts`

* `miniprogram/src/api/store.ts`

#### 实施方式

* `MerchantSettings` 增加 `pickup_enabled: boolean`，保留 `dine_in_enabled`。

* 为确认页新增明确的配送规则响应类型，例如：

  * `StoreDeliveryRules`

  * 字段包含 `enabled/base_fee/free_delivery_amount/max_distance/distance_rules/dine_in_enabled/pickup_enabled`

* `normalizeMerchantSettings()` 中对三个布尔值统一做 `!!` 归一化。

* `getStoreDeliveryRules()` 同步归一化三种方式布尔值，避免后端返回 `0/1` 时前端判断异常。

* 若 `miniprogram/src/api/index.ts` 与 `miniprogram/src/api/store.ts` 两处都保留了 `getStoreDeliveryRules()`，执行时一并同步，避免双实现漂移。

### D. 商家端“配送设置”页接入三种开关

#### 目标

* 商家在一个页面内完成：

  * 配送开关与配送资费

  * 堂食开关

  * 自提开关

#### 修改文件

* `miniprogram/src/pages/merchant/delivery-settings.vue`

* `miniprogram/src/pages/merchant/settings.vue`

#### 实施方式

* `delivery-settings.vue` 页面新增一个“下单方式”区域，包含三项 `switch`：

  * 配送

  * 堂食

  * 自提

* 页面加载时同时请求：

  * `getMerchantSettings()`

  * `getDeliverySettings()`

* 页面内分别维护：

  * 商家主表方式开关表单

  * 配送规则表单

* 保存时按固定顺序提交：

  1. `updateMerchantSettings({ dine_in_enabled, pickup_enabled })`
  2. `updateDeliverySettings({ enabled, base_fee, free_delivery_amount, max_distance, distance_rules })`

* 若第一步失败，则不提交第二步，避免页面误以为全部已保存。

* 若第二步失败，则提示“下单方式已保存，配送规则保存失败”或等价可理解文案，执行时根据现有 toast 风格保持简洁。

* 页面交互约束：

  * 配送规则区域仍只在配送开关打开时展示

  * 关闭配送时不强制删除距离规则数据，仅隐藏并保留，便于再次开启时恢复

* `pages/merchant/settings.vue` 中“配送设置”摘要从单一 `已开启配送 / 未开启配送` 改为三种方式摘要，例如：

  * `配送 / 堂食 / 自提`

  * 若全关则显示 `暂未开放下单方式`

### E. C 端确认页按开关动态展示并兜底

#### 目标

* 用户只看到商家当前开放的下单方式。

* 当前方式不可用时自动回退到第一个可用方式。

* 三种方式都关闭时，页面明确提示不可下单。

#### 修改文件

* `miniprogram/src/pages/store/confirm.vue`

#### 实施方式

* 将固定 `deliveryTypes` 改为“完整方式表 + 计算后的可用方式列表”。

* 页面加载配送规则后，根据返回值计算：

  * 配送可用：`enabled`

  * 堂食可用：`dine_in_enabled`

  * 自提可用：`pickup_enabled`

* 若当前 `deliveryType` 不在可用列表中，自动切换到第一个可用值。

* 若无任何可用方式：

  * 页面显示提示 `商家暂未开放下单方式`

  * 提交时直接拦截

* 表单校验保持现有逻辑，但只在“配送”方式下触发配送地址/电话/距离校验。

* 为避免方式切换后残留无效数据：

  * 从配送切到非配送时，只在提交 payload 时忽略配送字段，不强制清空用户输入

  * 这样用户切回配送时仍可保留已填信息

### F. PRD 同步更新

#### 目标

* 文档与代码口径一致，后续联调与验收不再依赖口头说明。

#### 修改文件

* `PRD.md`

* `docs/prd/PRD-功能说明.md`

* `docs/prd/PRD-接口文档.md`

* `docs/prd/PRD-测试与附录.md`

#### 实施方式

* 补充用户端订单详情“联系商家退款为直接拨号”的交互说明。

* 更新商家设置和配送设置说明，明确：

  * 配送是否支持及其费用结构继续由配送设置接口维护

  * 堂食、自提由商家主表配置维护

* 更新商家设置接口与店铺配送规则接口的返回/请求示例：

  * 增加 `pickup_enabled`

  * 明确 `delivery-rules` 返回配送开关本身以及堂食/自提开关

* 更新数据库字段说明，加入 `pickup_enabled`。

* 在测试与附录中补充三种方式开关组合的验收场景。

## Execution Order

1. 补齐前端类型与订单详情页拨号口径，消除当前 `contact_phone` 类型缺口。
2. 完成后端商家设置接口与用户侧 `delivery-rules` / `CreateOrder()` 的三种方式闭环。
3. 修正 `full_init.sql` 与 `mock_seed.sql` 的 `pickup_enabled` 列和值顺序。
4. 改商家端 `delivery-settings.vue` 和 `settings.vue` 摘要。
5. 改 C 端 `confirm.vue` 的动态方式展示与兜底逻辑。
6. 同步更新 PRD 文档。
7. 执行诊断、测试与构建验收。

## Verification Steps

### 1. 订单详情拨号

* 打开用户端 `pages/store/order-detail`。

* 点击“联系商家退款”。

* 有电话时直接拉起系统拨号。

* 无电话时仅出现一次兜底提示。

### 2. 商家三种方式开关

* 进入 `pages/merchant/delivery-settings`。

* 可以看到配送、堂食、自提三项开关。

* 分别切换并保存，重新进入页面后状态保持一致。

* 配送关闭时规则表单隐藏，再开启后原规则仍保留。

### 3. 确认页动态展示

* 仅开配送时，只显示配送。

* 仅开堂食时，只显示堂食。

* 仅开自提时，只显示自提。

* 多项开启时，只显示已开启的项。

* 三项全关时，页面提示不可下单且提交被拦截。

### 4. 服务端校验

* 对关闭的方式直接提交下单请求，接口返回对应错误提示。

* 对开启的方式提交订单，请求可正常创建。

* 配送方式仍校验地址、联系人、电话、距离与配送范围。

### 5. SQL 与数据初始化

* `full_init.sql` 中 `merchants` 初始化语句列和值顺序一致。

* `mock_seed.sql` 中 `merchants` 种子数据包含 `pickup_enabled`。

### 6. 基础验证

* 对修改过的前端文件执行 `GetDiagnostics` 检查。

* 运行后端相关测试或至少针对改动包编译验证。

* 小程序最终以 `npm run build:mp-weixin` 构建结果为验收准。

