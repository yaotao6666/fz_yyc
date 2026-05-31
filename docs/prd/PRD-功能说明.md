# PRD-功能说明

## 1. 文档定位

本文件用于承接 `PRD.md` 中与“产品功能、页面结构、角色职责、业务流程”相关的内容。

读取顺序：

1. 先读根目录 `PRD.md`
2. 再读本文件
3. 如需看接口细节，继续读 `docs/prd/PRD-接口文档.md`

## 2. 产品角色

### 2.1 商家端

- 默认入口，使用账号密码登录。
- 核心职责：
  - 店铺经营管理
  - 商品与分类管理
  - 订单查看、核销、退款处理
  - 数据分析
  - 设置与员工能力

### 2.2 用户端

- 通过扫码进入指定商家店铺。
- 核心职责：
  - 浏览商品
  - 加入购物车
  - 下单支付
  - 查看订单与申请退款

### 2.3 服务商端

- 独立 Web/PC 后台入口，负责统一管理多个商家。
- 核心职责：
  - 直接创建商家与维护商家资料
  - 配置商家 `sub_mch_id`、分账开关和分账比例
  - 查看服务商与商家分账历史
  - 商家经营看板与统计分析
  - 通过 `web-admin/` 公告管理页发布和维护系统公告

## 3. 页面结构

### 3.1 商家端页面

- `pages/auth/login`：商家登录页
- `pages/auth/agreement`：商家服务协议页，整合原隐私政策内容
- `pages/merchant/home`：商家工作台
- `pages/merchant/products/list`：商品管理
- `pages/merchant/products/edit`：商品编辑
- `pages/merchant/categories`：分类管理
- `pages/merchant/orders/list`：订单管理
- `pages/merchant/orders/detail`：订单详情
- `pages/merchant/analytics/index`：商家分析页
- `pages/merchant/settings`：商家设置首页
- `pages/merchant/marketing`：满减营销配置
- `pages/merchant/printers`：打印机管理
- `pages/merchant/settlements/history`：商家分账历史
- `pages/merchant/notification-settings`：消息提醒设置
- `pages/merchant/delivery-settings`：配送设置
- `pages/merchant/staff`：员工管理
- `pages/merchant/settings` 采用顶部信息卡 + 功能宫格，统一承接配送设置、满减营销、打印机管理、声音提醒、员工管理、分账历史等入口
- `pages/merchant/delivery-settings` 统一管理配送、堂食、自提三个下单方式开关；配送费与距离规则只服务于配送场景

### 3.2 用户端页面

- `pages/store/home`：店铺首页
- `pages/store/product`：商品详情
- `pages/store/cart`：购物车
- `pages/store/confirm`：确认订单
- `pages/store/my-orders`：我的订单
- `pages/store/order-detail`：订单详情
- `pages/store/order-sound-test`：商家 WebSocket 联调测试页
- `pages/store/home`、`pages/store/product`、`pages/store/confirm`：统一支持从 `merchant_id` 或 `scene` 解析商家入口参数
- `pages/store/confirm` 根据 `takeout_enabled`、`dine_in_enabled`、`pickup_enabled` 动态展示可选下单方式，未开启的方式不展示

### 3.3 服务商端页面（Web/PC 后台）

- `web-admin/src/views/login/LoginView.vue`：服务商登录页
- `web-admin/src/views/dashboard/DashboardView.vue`：服务商工作台
- `web-admin/src/views/merchant/MerchantListView.vue`：商家列表
- `web-admin/src/views/merchant/MerchantDetailView.vue`：商家详情
- `web-admin/src/views/merchant/MerchantEditView.vue`：商家创建与支付配置
- `web-admin/src/views/order/OrderListView.vue`：订单管理列表
- `web-admin/src/views/order/OrderDetailView.vue`：订单详情
- `web-admin/src/views/announcement/AnnouncementListView.vue`：公告列表与停用管理
- `web-admin/src/views/announcement/AnnouncementEditView.vue`：公告新增与编辑
- `web-admin/src/views/settlement/ProfitSharingHistoryView.vue`：服务商分账历史
- `web-admin/src/views/analytics/MerchantStatsView.vue`：服务商数据分析
- `web-admin/src/views/settings/SpSettingsView.vue`：服务商设置
- 服务商公告管理页已由 `web-admin/` 承接，继续复用后端 `/api/v1/sp/announcements*` 接口
- 服务商订单管理页由 `web-admin/` 承接，支持按商家、订单状态、日期区间和订单号查看订单，并提供独立详情页

## 4. 核心功能说明

### 4.1 商家登录与实时提醒

- 商家登录成功后进入工作台。
- 商家登录页需显式展示“同意商家服务协议”复选框，默认未勾选。
- 账号密码登录与微信快捷登录均需在勾选协议后才可继续。
- 商家登录页仅保留“商家服务协议”入口，点击后跳转至独立协议页，不再通过当前页弹窗展示协议内容。
- 原“隐私政策”内容已合并至“商家服务协议”中，由协议页统一承载。
- 工作台支持 WebSocket 连接状态展示。
- 工作台快捷功能需展示清晰图标、标题与辅助说明，避免仅靠文字识别。
- 工作台“今日概览”中：
  - 待处理订单按“待核销订单”展示，即当前已支付待核销订单数。
  - 商品数量按当前已上架商品总数展示。
- 工作台的待核销卡片与待办事项支持直接跳转订单列表对应状态。
- WebSocket 主要承载：
  - 新订单提醒
  - 用户进店提醒
- 用户完成支付后，后端向对应商家推送新订单提醒。
- 商家退出登录时必须立即断开 WebSocket，不再继续重连。

### 4.1.1 双小程序发布配置

- 小程序前端支持“一套源码发布两个微信小程序”。
- 差异配置统一收口到品牌配置文件与独立生产环境文件：
  - `miniprogram/config/brands/xunmeng.json`
  - `miniprogram/config/brands/caixu.json`
  - `miniprogram/.env.production.xunmeng`
  - `miniprogram/.env.production.caixu`
- 发布时通过预构建脚本 `miniprogram/scripts/apply-brand-config.mjs` 自动写入：
  - `src/manifest.json` 的 `name`
  - 顶层 `appid`
  - `mp-weixin.appid`
  - `src/pages.json` 的全局导航标题
  - `.env.production`
- 页面运行时品牌名称统一由 `src/config/env.ts` 中的 `APP_NAME` 提供，登录页与协议页不得再硬编码品牌名。
- 版本号仍统一维护在 `miniprogram/src/manifest.json` 的 `versionName` / `versionCode`。
- 标准发布命令：
  - `npm run build:mp-weixin:xunmeng`
  - `npm run build:mp-weixin:caixu`

### 4.2 商家接入与支付配置

- 商家由服务商直接创建，不再走小程序入驻、审核、进件页面。
- 服务商在商家管理中维护：
  - 商家基础资料
  - 商家账号
  - `sub_mch_id`
  - 分账开关
  - 分账比例
- 已进件商家只需要回填子商户号即可拉起支付，本轮联调样例为 `1112649854`。
- `sub_mch_id` 已配置且分账规则合法时，支付配置状态视为完成。

### 4.3 商品与分类管理

- 商家可以维护商品、分类、规格、图片与上下架状态。
- 用户端商品价格计算遵循“基础价 + 规格加价”。
- 店铺首页点击商品 `+` 支持弹出规格与数量选择。
- 确认订单页只展示商家已开启的下单方式：
  - 配送对应 `takeout_enabled`
  - 堂食对应 `dine_in_enabled`
  - 自提对应 `pickup_enabled`
- 用户选择配送时需填写地址、联系人、联系电话，并从商家配置的配送距离档位中选择范围；选择堂食或自提时不展示配送表单。
- 确认订单页展示商家当前启用的满减规则、已命中优惠和下一档提示，下单金额以服务端返回的 `discount_amount`、`pay_amount` 为准。

### 4.3.1 满减营销

- 商家可在 `pages/merchant/marketing` 配置最多 5 档满减规则。
- 单档规则至少包含：
  - 满减门槛金额
  - 减免金额
  - 启用状态
- 规则保存后整体覆盖当前商家的满减配置。
- 用户下单时按当前满足条件的最高优惠档位自动生效。

### 4.3.2 打印机管理

- 商家可在 `pages/merchant/printers` 管理打印机列表。
- 打印机管理至少支持：
  - 新增
  - 编辑
  - 删除
  - 启停开关
  - 自动打印开关
  - 默认打印机切换
  - 测试打印
- 打印机类型至少支持：
  - 通用云打印机
  - 飞鹅打印机
- 飞鹅打印机需维护：
  - 飞鹅账号
  - 飞鹅 UKey
  - 飞鹅终端号

### 4.4 订单与核销

- 商家订单详情可查看：
  - 订单基础信息
  - 支付时间
  - 支付单号
  - 核销码
  - 核销时间
  - 核销人
- 商家对“当前订单”执行核销时，不再要求再次输入核销码。
- 工作台保留“快速核销”入口，用于扫码或输入任意订单核销码快速核销。
- 商家订单列表支持时间段筛选：
  - 快捷范围：全部、今日、近7天、近30天、自定义
  - 自定义范围：开始日期、结束日期
  - 若开始日期晚于结束日期，前端需提示并阻止查询

### 4.5 退款处理

- 商家可对已支付、已完成订单发起退款处理。
- 退款处理需要填写退款原因。
- 退款金额默认按订单实付金额执行，具体以接口文档约定为准。
- 用户端 `pages/store/order-detail` 的“联系商家退款”点击后直接拉起商家电话拨号；无电话时只展示兜底提示。

### 4.6 服务商数据分析

- 服务商分析只保留与商家经营相关的核心分析能力：
  - 不同商家的访问率
  - 不同商家的下单率
  - 不同商家的下单金额
  - 不同商家的下单均价
  - 日 / 周 / 月 / 年订单量分析
  - 商家排行榜维度切换
- 页面必须兼容空数据返回，不能因 `null` 导致页面报错。
- 排行榜支持按访问率、下单率、下单金额、下单均价切换。
- 统计面板需展示汇总卡片、商家转化明细和周期订单桶数据。

### 4.7 商家分析页库存预警

- 商家分析页底部库存预警采用看板形态，而不是简单列表。
- 看板需包含：
  - 预警商品总数
  - 紧急补货数量
  - 建议关注数量
  - 风险分组商品列表
  - 空态说明
  - 去商品管理入口
- 风险分级口径：
  - 紧急补货：库存小于等于 5
  - 建议关注：库存 6 到 10
### 4.8 商家资料管理

- 服务商可在商家详情中维护：
  - 商家 Logo
  - 商家背景图
- 图片更新后需同步影响商家端与用户端展示口径。

### 4.9 分账历史

- 支付成功回调后，系统按商家配置即时发起分账。
- 分账记录单独沉淀为历史数据，至少包含：
  - 分账日期
  - 订单号
  - 支付金额
  - 抽佣比例
  - 抽佣金额
  - 商家实收
  - 状态
  - 失败原因
- 服务商端可查看全部商家分账历史。
- 商家端可查看本店分账历史。

## 5. 完整业务链路摘要

### 5.1 商家经营主链路

1. 商家登录
2. 配置商品与分类
3. 用户扫码进店
4. 店铺首页、商品详情、确认订单按统一入口参数加载当前商家数据
5. 浏览商品并加入购物车
6. 下单支付
7. 商家收到提醒
8. 商家核销
9. 服务商查看统计分析

### 5.2 服务商配置与分账主链路

1. 服务商创建商家
2. 配置商家商家账号
3. 回填 `sub_mch_id`
4. 配置分账开关与抽佣比例
5. 用户下单支付
6. 支付回调后自动分账
7. 服务商与商家查看分账历史

## 6. 与根 PRD 的对应关系

- 项目概述、角色与定位：对应 `PRD.md` 第 1 章
- 功能优先级与系统能力：对应 `PRD.md` 第 2 章
- 业务流程与状态约定：对应 `PRD.md` 第 1.8 节及相关业务章节
- 若功能实现有变化，必须同步更新本文件与根 `PRD.md` 索引摘要
