# 服务商分账功能实施计划（自动分账 + 后台可视化接收方/记录管理 + 功能权限）

## 一、概要

基于微信服务商分账能力（`/v3/profitsharing/*`），在现有单商户（`merchants.id=1`，`sub_mch_id=1112979963`）架构上实现：
1. **分账接收方管理**：后台可视化维护多个分账接收方，支持「商户号 `MERCHANT_ID`」与「个人微信 `PERSONAL_OPENID`」两类账户；新增时同步到微信建立分账关系，删除时解绑关系。
2. **自动按配置分账**：当某笔订单支付成功后，若分账开关开启且有启用中的接收方，系统按各接收方的默认分账比例自动计算并调用微信发起分账单（每个支付单一个分账单、分给多方）。
3. **分账记录查看**：后台独立「分账管理」页集中查看分账记录（列表/明细/重试），同时在订单详情内联展示该订单的分账状态与金额。
4. **功能权限**：全部接入现有 RBAC（`sys_menus` 菜单 + `permission` 权限标识 + `middleware.RBAC`），仅授权员工可用。

> 注：本项目历史上曾实现过「支付成功自动按单一比例分账给服务商」的版本（commit `3c2ed42`/`647f713` 中 `services/wechatpay/service_provider.go`、`handlers/sp/payment.go`），后在未提交工作区中删除。本计划**复用其微信 API 交互与自动触发思路**，但按新需求重构为「多接收方 + 可视化管理 + RBAC」。

---

## 二、现状分析

- **WeChat 客户端**：[client.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/services/wechatpay/client.go) 使用 SDK `wxpay.Client`（证书/私钥已配好），已有配套的 `jsapi`、`refunddomestic` 服务调用方式；`NewServiceProviderClient()` 统一初始化。SDK 版本 `github.com/wechatpay-apiv3/wechatpay-go v0.2.21` 自带 `services/profitsharing` 全量包（`OrdersApiService`/`ReceiversApiService`/`MerchantsApiService`/`TransactionsApiService`/`ReturnOrdersApiService`）。
- **应用身份**：[app_identity.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/services/wechatpay/app_identity.go) 当前仍含 `GetActiveAppIdentity()` 及 `sp_app`/`sub_app` 模式切换。**本项目现固定为 `sub_app`，需在实施时移除该模式切换的配置参数与 if 判断（见 3.0）**，收单与分账统一使用特约商户小程序 appid。
- **支付回调**：[wechatpay.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/callbacks/wechatpay.go) 在 `markOrderPaid` 中处理支付成功，可拿到 `TransactionID`，是自动分账的唯一可靠触发点。
- **订单模型**：[models.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/models/models.go) `Order` 已无 `profit_sharing_*` 字段（曾在 `20260513120000` 加入、`20260815000000_remove_profit_sharing.sql` 移除）。
- **RBAC**：[rbac.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/middleware/rbac.go) 通过 `middleware.RBAC(code)` 控制；菜单在 [20260818000100_rbac.sql](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/migrations/20260818000100_rbac.sql) 中种入 `sys_menus`/`sys_roles`/`sys_role_menus`（admin 角色 id=1 全量绑定）。
- **路由**：[main.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/cmd/server/main.go) 中 `merchantOnlyGroup`（`JWTAuth`+`MerchantAuth`+`RBAC`）集中注册商家后端接口。
- **前端**：web-admin 采用 TS + Vue，API 封装于 [sp.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/api/sp.ts)，类型在 [sp.ts 类型](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/types/sp.ts)，路由在 [router/index.ts](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/web-admin/src/router/index.ts)。

---

## 三、设计方案

### 3.0 前置清理：移除 `sp_app`/`sub_app` 双模式（固定 `sub_app`）

系统现固定为 `sub_app` 模式，需删除模式切换相关的配置参数与判断代码，分账固定基于 `sub_app`（特约商户小程序 `appid`）实现：

- **[config.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/config/config.go)**：移除 `WechatPay.AppMode` 字段、`WECHAT_PAY_APP_MODE` 的 viper 与环境变量读取（`mergeConfig`/`mergeEnvConfig` 两处）、默认值 `AppMode: "sp_app"`。
- **[app_identity.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/services/wechatpay/app_identity.go)**：删除 `AppModeSPApp` 常量、`normalizeAppMode()` 及 `GetActiveAppIdentity()` 内的 `switch` 分支，固定返回 `sub_app` 身份（`AppID=SubAppID`、`AppSecret=SubAppSecret`、`Mode="sub_app"`、`OpenIDFieldName="sub_openid"`）。`Wechat` 结构里 `SubAppID/SubAppSecret` 继续保留（其余 `sp_app` 用到的 `AppID/AppSecret` 若仅服务于 `sp_app` 不再需要，可在清点引用后移除）。
- **[client.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/services/wechatpay/client.go)**：`JSAPIPayRequest` 删除 `AppMode` 字段；`CreatePartnerJSAPIPayOrder` 删除 `strings.EqualFold(req.AppMode, AppModeSubApp)` 判断与 `subAppidPayerOpenID()`，恒以 `config.Config.Wechat.SubAppID` 作为 `SubAppid`、`req.OpenID` 作为 `Payer.SubOpenid` 传入；`spAppid` 仍取服务商 `config.Config.Wechat.AppID`。删除 `AppModeSubApp` 常量引用。
- **[user/handler.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/user/handler.go)**：`JSAPIPayRequest` 字面量删除 `AppMode: appIdentity.Mode`。
- **[.env.example](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/.env.example) / [.env.production.example](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/.env.production.example)**：移除 `WECHAT_PAY_APP_MODE`。

> 分账固定使用 `sub_app`：`/v3/profitsharing/orders` 与 `/v3/profitsharing/receivers/add` 中 `appid` 一律取 `GetActiveAppIdentity().AppID`（= `SubAppID`），`sub_mchid` 取特约商户 `sub_mchid`。

### 3.1 数据模型（新增迁移 `20260822000000_profit_sharing.sql`）

**表 `profit_sharing_receivers`（分账接收方）**
| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | |
| merchant_id | BIGINT | 恒为 `1`（单商户） |
| receiver_type | TINYINT | 1=`MERCHANT_ID` 商户号 2=`PERSONAL_OPENID` 个人 |
| name | VARCHAR(64) | 显示名称 |
| account | VARCHAR(64) | 商户号 或 个人 openid |
| personal_name | VARCHAR(64) NULL | 个人接收方需真实姓名（微信校验） |
| relation_type | VARCHAR(32) | 默认 `SERVICE_PROVIDER`（服务商分润）|
| default_ratio | DECIMAL(5,2) | 自动分账默认比例(%)，0-100 |
| wechat_bound | TINYINT | 是否已在微信建立接收方关系 0/1 |
| wechat_error | VARCHAR(256) | 微信建关系失败原因 |
| status | TINYINT | 1=启用 0=停用 |
| sort/remark | | 排序/备注 |
| created_at/updated_at | | |

**表 `profit_sharing_records`（分账单，一个支付单一个）**
| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | |
| merchant_id | BIGINT | 恒 `1` |
| order_id / order_no | | 关联订单 |
| sp_mchid / sub_mchid | VARCHAR | 服务商/特约商户号 |
| appid | VARCHAR | 分账请求使用的 appid |
| transaction_id | VARCHAR | 微信支付交易单号 |
| out_order_no | VARCHAR(64) | `ps_{order_no}_{ts}` 微信分账单号 |
| total_amount | DECIMAL(10,2) | 订单实付（生猪用于校验） |
| total_share_amount | DECIMAL(10,2) | 本次分账总额 |
| status | TINYINT | 0=待分账 1=分账中 2=成功 3=失败 4=已跳过 |
| share_time | DATETIME | 分账成功时间 |
| error_message | VARCHAR(512) | 失败原因 |
| created_at/updated_at | | |

**表 `profit_sharing_record_receivers`（分账单内各方明细）**
| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | |
| record_id | BIGINT | 关联分账单 |
| receiver_id | BIGINT | 关联接收方 |
| receiver_type / receiver_name / account | | 快照 |
| amount | DECIMAL(10,2) | 分账金额 |
| result_status | VARCHAR(32) | 微信方状态：`PROCESSING`/`SUCCESS`/`FAILED`/`FINISHED` |
| fail_reason / finish_time | | 微信返回 |
| created_at/updated_at | | |

**`merchants`、`orders` 增量字段**
- `merchants`：`ADD COLUMN profit_sharing_enabled TINYINT NOT NULL DEFAULT 0`（自动分账总开关）。
- `orders`：`ADD COLUMN` `profit_sharing_status TINYINT`(0未分 1分账中 2成功 3失败 4跳过)、`profit_sharing_amount DECIMAL(10,2)`、`profit_sharing_order_no VARCHAR(64)`、`profit_sharing_at DATETIME`、`profit_sharing_error VARCHAR(256)`。

> 迁移文件头部加 `SET NAMES utf8mb4;`（避免中文乱码，见经验教训）。

---

### 3.2 后端代码

**A. 微信 API 层 — [client.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/services/wechatpay/client.go) 新增（复用 SDK `profitsharing`）**

新增请求/响应结构与方法（风格对齐现有 `jsapi`/`refunddomestic` 用法）：
- `ReceiverItem{ Type, Account, Name, Amount }`
- `ProfitSharingRequest{ AppID, SubMchID, TransactionID, OutOrderNo, Receivers []*profitsharing.CreateOrderReceiver }`
- `ProfitSharingResult{ OutOrderNo, Status, Receivers []... }`
- 方法：
  - `(c *ServiceProviderClient) AddProfitSharingReceiver(ctx, AddProfitSharingReceiverRequest) error` → `ReceiversApiService.AddReceiver`
  - `(c *ServiceProviderClient) DeleteProfitSharingReceiver(ctx, req) error` → `ReceiversApiService.DeleteReceiver`
  - `(c *ServiceProviderClient) CreateProfitSharingOrder(ctx, ProfitSharingRequest) (*ProfitSharingResult, error)` → `OrdersApiService.CreateOrder`
  - `(c *ServiceProviderClient) QueryProfitSharingOrder(ctx, subMchID, outOrderNo) (*ProfitSharingResult, error)` → `OrdersApiService.QueryOrder`
  - `(c *ServiceProviderClient) QueryProfitSharingMerchantRatio(ctx, subMchID) (int64, error)` → `MerchantsApiService.QueryMerchantRatio`
- 错误归类辅助（等价旧版逻辑）：`IsProfitSharingReceiverAlreadyExists`、`IsProfitSharingReceiverRelationNotExist`（通过 API 错误 `Code`/`Message` 判断，处理 `PARAM_ERROR`+「已存在」「关系不存在」）。

**B. 自动分账触发 — [callbacks/wechatpay.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/callbacks/wechatpay.go)**

在 `markOrderPaid` 支付成功后，调 `services/profitsharing/auto.go` 新增的函数 `AutoProfitShareForPaidOrder(ctx, order, transactionID)`：
1. 查询 `merchants` 取 `profit_sharing_enabled`、`sub_mchid`（缺失则跳过并置「已跳过」）；
2. 查询启用中的 `profit_sharing_receivers`（`status=1`），无则跳过；
3. 调 `QueryProfitSharingMerchantRatio` 取上限（默认 30%），校验总比例 ≤ 上限 ≤ 100；
4. 计算各方金额 `amount = floor(pay_amount * default_ratio / 100)`，汇总成 `total_share_amount`；金额为 0 的接收方剔除；
5. `out_order_no = fmt("ps_%s_%d", order.OrderNo, now.Unix())`；
6. 写 `profit_sharing_records` + `profit_sharing_record_receivers`，更新 `orders.profit_sharing_status=1`；
7. 若接收方关系未建，先 `AddProfitSharingReceiver` 再传；`CreateProfitSharingOrder`（`unfreeze_unsplit=false`）；
8. 异步拉 `QueryProfitSharingOrder` 回写明细结果，成功置 `orders.profit_sharing_status=2`，失败置 `3` 并记录 `profit_sharing_error`。
- 封装成 `go func()` 内执行 + 简易重试（失败保留记录供后台手动重试），不阻塞回调返回。

**C. 业务接口 — 新文件 [merchant/profit_sharing.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/handlers/merchant/profit_sharing.go)**

注册到 [main.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/cmd/server/main.go) 的 `merchantOnlyGroup`：
- 接收方：
  - `GET  /merchant/profit-sharing/receivers`（`RBAC("profit:view")` 列表）
  - `POST /merchant/profit-sharing/receivers`（`profit:receiver:manage` 创建+建微信关系）
  - `PUT  /merchant/profit-sharing/receivers/:id`（`profit:receiver:manage` 改比例/启停/备注）
  - `DELETE /merchant/profit-sharing/receivers/:id`（`profit:receiver:manage` 删除＋解绑微信关系）
  - `POST /merchant/profit-sharing/receivers/:id/sync`（`profit:receiver:manage` 重试建关系）
- 配置：
  - `GET  /merchant/profit-sharing/config`（`profit:view` 开关+微信允许比例）
  - `PUT  /merchant/profit-sharing/config`（`profit:config` 开关）
- 记录：
  - `GET  /merchant/profit-sharing/records`（`profit:view` 列表/筛选）
  - `GET  /merchant/profit-sharing/records/:id`（`profit:view` 明细含各方）
  - `POST /merchant/profit-sharing/records/:id/retry`（`profit:share` 重试失败单）
- 订单内联：`Order` 模型补回分账字段，`GetOrderDetail` 返回分账状态；订单详情响应附带该订单最近一条 `profit_sharing_records`。

**D. 模型 — [models.go](file:///Users/yaotao/Documents/trae_projects/fz_yyc_cxsm/server/internal/models/models.go)**
补回 `Order` 的分账字段，新增 `ProfitSharingReceiver` / `ProfitSharingRecord` / `ProfitSharingRecordReceiver` 结构（含 `TableName`），并给 `Merchant` 补 `ProfitSharingEnabled`。

---

### 3.3 RBAC 权限与种子

在迁移中追加 `sys_menus` 与 `sys_role_menus`（admin 角色全量绑定）：
| id | parent_id | 类型 | name | path | permission |
|---|---|---|---|---|---|
| 8 | 0 | 1 | 分账管理 | /settlement | NULL |
| 81 | 8 | 1 | 分账接收方 | /settlement/receivers | profit:view |
| 82 | 8 | 1 | 分账记录 | /settlement/records | profit:view |
| 811 | 81 | 2 | 新增/编辑/删除/同步 | – | profit:receiver:manage |
| 821 | 82 | 2 | 分账/重试 | – | profit:share |
| 822 | 82 | 2 | 分账配置 | – | profit:config |

> 权限标识即路由 `RBAC()` 所用 code。admin（id=1）自动全量；普通角色需在「角色管理→分配权限」勾选后可见。新增菜单后调用 `middleware.ClearRBACCache()`（或重启缓存 60s 自动失效）。

---

### 3.4 前端 web-admin

- **类型**：`types/sp.ts` 新增 `ProfitSharingReceiver/Record/RecordReceiver/Config` 接口。
- **API**：`api/sp.ts` 新增上述接口的请求函数（复用 `unwrapApiResponse`）。
- **路由**：`router/index.ts` 新增 `/settlement/receivers`、`/settlement/records`（懒加载 + 继承权限守卫，`visible` 由菜单权限驱动）。
- **视图**：
  - `views/settlement/ProfitSharingReceiversView.vue`：接收方表格 + 新增/编辑对话框（类型选择商户号/个人微信、账号、真实姓名、默认比例%、启停、关系状态徽标、删除/同步）。
  - `views/settlement/ProfitSharingRecordsView.vue`：分账记录表格（筛选状态/日期/单号）+ 明细抽屉（含各方金额/微信状态）+「重试」按钮。
  - 订单 `/order/OrderDetailView.vue`：展示该订单分账状态徽标、分账金额，跳转对应记录。

---

## 四、假设与决策

1. **触发点固定为支付回调**（`markOrderPaid`）——用户已确认「自动按配置分账」。
2. **比例逐接收方配置**存于 `profit_sharing_receivers.default_ratio`，自动分账下发时汇总校验 ≤ 微信允许最大比例（`QueryMerchantRatio`，兜底 30%），总额 ≤ 订单实付。
3. **接收方两类**：商户号 + 个人微信 openid（个人需真实姓名，微信建关系时校验）。
4. **多接收方共用一个分账单**，与微信 `orders.receivers[]` 模型一致。
5. **固定 `sub_app` 模式**：本项目不再支持 `sp_app`/`sub_app` 切换，本计划会一并移除相关配置参数与判断代码（见 3.0），分账使用特约商户 appid。
6. **失败可后台重试**（`records/:id/retry`），不自动无限重试。
7. **分账回退（退款后 ReturnOrder）不在本期**：已分账订单后续退款需调用 `/v3/profitsharing/return-orders` 回退，本计划作为已知边界标注，不在首版实现（避免过度设计）；如后续需要可增量补充 `ReturnOrdersApiService`。
8. 沿用现有 SDK `profitsharing` 包与 `wxpay.Client`，不自造 HTTP 签名（优于旧版手动 `doJSONRequest`）。
9. 分账金额单位一律整数分（`roundAmount`/`amountToCents` 已有约定参考旧版）。
10. **路由文件可改**：main.go 的「不修改路由」约束仅针对当时 RBAC 中间件重构；本功能为新增路由，允许在 `merchantOnlyGroup` 内追加。

---

## 五、验证步骤

1. **构建**：`cd server && go build ./...` 编译通过。
2. **迁移**：执行 `20260822000000_profit_sharing.sql`（需符合「纯 SQL 单条执行」约束，无存储过程），确认建表成功、数据可正常读写中文。
3. **RBAC**：admin 登录 → 侧边栏出现「分账管理」；新建普通角色勾选分账权限后可访问；未授权员工访问被 `Forbidden` 拦截。
4. **接收方**：后台新增一个商户号接收方和一个个人微信接收方，触发微信 `AddReceiver`（沙箱/正式按需），关系状态正确同步。
5. **自动分账**：完成一笔支付，观察回调触发自动分账，`profit_sharing_records` 落库、`orders.profit_sharing_status` 更新为成功（或失败可重试）。
6. **前端**：`cd web-admin && npm run build` 通过；「分账管理」两页可增删改查、记录明细正确；订单详情内联分账信息展示正常。
7. **配置开关**：关闭 `profit_sharing_enabled` 后新支付不再自动分账并置「已跳过」。

---

## 六、主要改动文件清单

**后端**
- `server/migrations/20260822000000_profit_sharing.sql`（新增：建表/加列/菜单种子，`SET NAMES utf8mb4`）
- `server/internal/config/config.go`（移除 `AppMode` 配置与 `WECHAT_PAY_APP_MODE` 读取）
- `server/internal/services/wechatpay/app_identity.go`（固定 `sub_app` 身份，移除模式切换）
- `server/internal/services/wechatpay/client.go`（移除 `AppMode` 判断；新增分账方法）
- `server/internal/handlers/user/handler.go`（移除 `AppMode` 传参）
- `server/internal/services/profitsharing/auto.go`（新增：自动分账逻辑）
- `server/internal/handlers/callbacks/wechatpay.go`（支付回调接线）
- `server/internal/handlers/merchant/profit_sharing.go`（新增业务接口）
- `server/internal/models/models.go`（模型补字段/新表结构）
- `server/cmd/server/main.go`（注册新路由）
- `server/.env.example`、`server/.env.production.example`（移除 `WECHAT_PAY_APP_MODE`）

**前端**
- `web-admin/src/types/sp.ts`（新增类型）
- `web-admin/src/api/sp.ts`（新增 API）
- `web-admin/src/router/index.ts`（新增路由）
- `web-admin/src/views/settlement/ProfitSharingReceiversView.vue`（新增）
- `web-admin/src/views/settlement/ProfitSharingRecordsView.vue`（新增）
- `web-admin/src/views/order/OrderDetailView.vue`（订单内联分账展示）