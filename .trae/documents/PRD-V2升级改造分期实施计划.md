# PRD V2.0 升级改造分期实施计划

> 依据《银发智慧养老服务平台 PRD V2.0》与现有产品现状的差距分析制定。
> 第一期范围（用户已确认）：订单口径统一与商品/服务管理拆分、优惠券营销闭环、服务过程安全、服务完结触发评价、健康服务板块重构（抹除照护/随访/体征、多档案、宣教独立）。
> 结算提现体系（T+1 提现/财务审核/售后冲抵）、亏损订单预警、操作日志、区域数据范围列为第二期展望。

---

## 1. 现状盘点结论

| 模块 | 现状 | 结论 |
|------|------|------|
| 健康评估与适配（评估量表→评估记录→辅具适配建议） | 已具备 | 保留，挂接多档案 |
| 健康档案 | 单账号单档案（user_id 唯一索引） | 需改造为多档案 |
| 照护计划 / 随访任务 / 生命体征 | 已实现（care_plans / follow_up_tasks / health_monitoring 等 4 张表） | 按需求抹除 |
| 健康宣教 | 文章表内置字符串分类，按慢病标签推荐 | 重构为独立板块（分类+内容） |
| 商品租售（售卖/租赁/押金/归还） | 全链路已通 | 已具备 |
| 微信服务商支付+自动分账 | 已通 | 已具备 |
| RBAC（菜单/角色/部门/员工，三层控制） | 已具备 | 已具备 |
| 服务人员审核（双通道准入+留痕） | 已具备 | 已具备 |
| 订单类型口径 | `order_type` 1-6 混用，前后端无统一"实物/服务"二分口径 | 需统一 |
| PC 商品/服务管理 | 单一「商品管理」混合 5 类 `product_type` | 需拆分 |
| 优惠券营销 | 无表、无接口、无页面 | 全新建 |
| 服务过程安全（GPS轨迹/录音/SOS/超时预警） | 仅签到定位 | 全新建 |
| 评价与服务质量分 | 无 | 全新建 |

## 2. 核心口径定义（全阶段共享）

### 2.1 订单二分口径（不加字段，代码层约定）

```
实物订单 = order_type ∈ {1 普通商品, 2 租赁商品}
  主状态走 orders.status（待支付/已支付/已完成/已取消/退款中/已退款）
  携带押金、租赁归还、物流配送语义

服务订单 = order_type ∈ {3 即时服务, 4 预约服务, 5 上门服务, 6 到店服务}
  工单状态走 orders.biz_status（0无 1待接单 2已接单 3服务中 4待支付尾款 5已完成 6已取消）
  携带指派人员、签到/签退、轨迹、录音、评价语义
  **必须绑定 1 个健康档案（record_id）**：下单时指定档案人员，无档案不可下服务单
```

- 后端提供统一 helper（如 `OrderCategory(orderType)`），前后端共享常量定义。
- 管理后台订单列表、C 端订单列表均按"实物订单/服务订单"两个 Tab 切换。

### 2.2 商品/服务管理拆分口径（表结构复用，不拆表）

```
「商品管理」（实物） = product_type ∈ {1 辅具零售, 2 辅具租赁}
「服务管理」（服务） = product_type ∈ {3 康养套餐, 4 陪诊服务, 5 科普资讯}
```

- 底层继续复用 `products` 表与现有商品接口（接口已支持 `product_type` 筛选）。
- `web-admin` 将现有 `MerchantProductsTab` 组件参数化（传入类型过滤），分别挂载到「商品管理」「服务管理」两个菜单视图，编辑器按类型显隐字段。

## 3. 分期总览与依赖

```
阶段一 地基：订单口径统一 + 商品/服务管理拆分
   ↓
阶段二 营销：优惠券全链路（依赖：订单金额口径 DiscountAmount/PayAmount）
   ↓
阶段三 安全：服务区域 + 轨迹/录音/SOS/超时预警（依赖：服务订单口径）
   ↓
阶段四 评价：完结推送 → C端评价 → 服务质量分（依赖：阶段三的服务记录表）
   ↓
阶段五 健康重构：抹除照护/随访/体征 + 多档案 + 宣教独立（独立于前四阶段，可并行）
```

每阶段固定执行四步流程：**接口开发 → 按钮权限码 → 迁移脚本（含 sys_menus 菜单节点，SET NAMES utf8mb4）→ PRD 文档同步**。

---

## 4. 阶段一：订单口径统一与商品/服务管理拆分

### 4.1 后端
- `server/internal/utils/constants.go` 新增订单分类常量与 `OrderCategory(orderType)` helper（实物=1/服务=2）。
- 订单列表接口支持 `category` 参数过滤（实物/服务），映射为 order_type 范围查询。
- 服务订单下单时从所选收货地址写入 `orders.delivery_district`（为阶段三区域匹配做地基）。
- **服务订单绑定档案（硬性规则）**：`orders` 新增 `record_id`（健康档案ID）；服务订单下单必填 `record_id` 并校验档案归属下单用户，档案不存在或不属于本人 → 拒绝下单；实物订单不要求。用户无任何档案时下服务单 → 引导先建档（返回明确错误码，C 端跳建档页）。

### 4.2 web-admin
- 「订单管理」列表页拆分为"实物订单/服务订单"两个 Tab（组件复用，传 category）。
- 新增「服务管理」菜单（视图复用 `MerchantProductsTab`，固定 `product_type ∈ {3,4,5}`），原「商品管理」视图固定 `product_type ∈ {1,2}`。
- `MerchantProductEditorDialog` 按所在视图限定可选商品类型。

### 4.3 迁移脚本 `006_order_category_split.sql`
- `ALTER TABLE orders ADD COLUMN delivery_district`（varchar(32)，服务订单收货区县）。
- `ALTER TABLE orders ADD COLUMN record_id`（bigint，index，服务订单绑定的健康档案ID）。
- `sys_menus`：新增「服务管理」目录及列表菜单节点、按钮权限节点。
- 权限码：`service-products:view` / `service-products:create` / `service-products:update` / `service-products:delete`。
- 脚本头部 `SET NAMES utf8mb4;`。

### 4.4 验收要点
- 后台商品管理看不到服务类商品，服务管理看不到实物商品；商品编辑互不串类型。
- 订单列表两个 Tab 数据互斥且并集完整；C 端订单状态展示不受影响。
- RBAC：仅授予商品权限的角色看不到服务管理菜单。
- 服务订单不传 `record_id` / 传他人档案 → 下单被拒；无档案用户下服务单 → 收到"请先建档"引导；实物订单不受影响。

### 4.5 C 端交互（服务订单选档案）
- `confirm.vue` 结算页：购物车含服务类商品时展示"服务对象"栏，必选 1 个本人账号下的健康档案（现阶段单账号单档案即默认选中，阶段五 8.2 多档案后变为从档案列表选择）。
- 服务订单详情展示服务对象（姓名/性别/年龄），后台服务订单详情同样展示关联档案。

---

## 5. 阶段二：优惠券营销闭环

### 5.1 数据表

**coupon_templates 券模板表**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | |
| name | varchar(64) | 券名称 |
| type | tinyint | 1=满减券 2=折扣券 |
| threshold_amount | decimal(10,2) | 使用门槛（0=无门槛） |
| discount_amount | decimal(10,2) | 满减面值（满减券） |
| discount_rate | decimal(3,2) | 折扣率如 0.90（折扣券，最大优惠可另设封顶） |
| total_count | int | 发行总量（0=不限） |
| received_count | int | 已领取数（乐观锁扣减） |
| per_user_limit | int | 每人限领 |
| valid_type | tinyint | 1=固定期限 2=领取后N天有效 |
| valid_start_at / valid_end_at | datetime | 固定期限时使用 |
| valid_days | int | 领取后N天有效 |
| apply_scope | tinyint | 1=全场通用 2=指定分类 3=指定商品 |
| scope_ids | json | 适用范围ID列表 |
| status | tinyint | 1=启用 0=停用 |
| remark / created_at / updated_at | | |

**user_coupons 用户券表**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | |
| user_id | bigint index | |
| template_id | bigint index | |
| status | tinyint | 1=未使用 2=已使用 3=已过期 4=已作废 |
| source | tinyint | 1=自主领取 2=系统发放(30天唤回) 3=运营手动发放 |
| received_at | datetime | 领取时间 |
| expired_at | datetime | 过期时间（领取时计算落库） |
| used_at / order_id / order_no | | 核销信息 |
| 唯一约束 | | (user_id, template_id, 领取批次) 控制限领 |

### 5.2 接口清单

**后台（RBAC 保护）**

| 接口 | 说明 | 权限码 |
|------|------|--------|
| GET/POST /api/v1/merchant/coupon-templates | 列表/创建 | coupon-templates:view / create |
| PUT/DELETE /api/v1/merchant/coupon-templates/:id | 更新/删除 | coupon-templates:update / delete |
| POST /api/v1/merchant/coupon-templates/:id/status | 启停 | coupon-templates:update |
| POST /api/v1/merchant/coupon-templates/:id/grant | 手动发放给指定用户 | coupon-templates:create |
| GET /api/v1/merchant/user-coupons | 领取/使用记录 | user-coupons:view |

**C 端（store）**

| 接口 | 说明 |
|------|------|
| GET /api/v1/store/coupons/available | 可领券列表（首页弹窗/领券入口） |
| POST /api/v1/store/coupons/:template_id/receive | 领取（校验限领/总量/启用状态） |
| GET /api/v1/store/my-coupons?status=1 | 我的券（未使用/已使用/已过期） |
| 下单预览/提交接口 | 新增 `user_coupon_id` 参数，服务端校验并计算 `DiscountAmount`；不合规返回明确错误 |

### 5.3 业务规则
- 抵扣计算：`PayAmount = TotalAmount + DeliveryFee + 押金 - DiscountAmount`；**分账基数 = 微信实付金额**，优惠券抵扣部分天然不参与分润，`profitsharing/auto.go` 无需改动。
- 优惠券与订单退款：退款成功后已核销券**不返还**（作废，status=4 由核销后退款场景写入）。
- 定时任务（每日凌晨）：30 天未下单用户自动发券（source=2，模板取系统配置）；用户券过期状态刷新。
- C 端触达：进入小程序时检查"有新发放的券"→ 首页弹窗提醒领取（简单实现：返回待查看数量红点）。

### 5.4 前端改造
- web-admin：新增「优惠券管理」（模板 CRUD + 发放记录 Tab）。
- C 端：领券入口（首页活动位）、`confirm.vue` 结算页选券（展示最优可用券）、`my.vue` 我的券列表、订单详情显示优惠明细。

### 5.5 迁移脚本 `007_coupon_system.sql`
- 建两张表 + `sys_menus`「优惠券管理」菜单与按钮权限节点。

### 5.6 验收要点
- 满减/折扣两种券的下单抵扣金额计算正确（含门槛、押金不参与抵扣）。
- 限领/总量/停用券不可领取；过期券不可使用。
- 使用券后支付成功的订单，分账基数 = 实付金额。
- 30 天未下单任务只对目标用户发放一次（活动周期内去重）。

### 5.7 实施记录（2026-08-31）
- 后端：`internal/services/coupon`（领取/限领/总量乐观锁/有效期计算/核销/适用范围校验）、merchant 端模板 CRUD + 手动发放、C 端可领列表/领取/我的券/下单可用券预览、CreateOrder 集成 `user_coupon_id` 抵扣、定时任务（每小时过期刷新 + 每 24h 30天唤回发券）。
- 迁移 `007_coupon_system.sql` 已执行：coupon_templates / user_coupons 两表 + sys_menus 节点 96/960-965。
- web-admin：「优惠券管理」视图（模板 CRUD + 启停 + 手动发放 + 领取/使用记录 Tab）。
- C 端：首页领券中心（横滑可领券 + 一键领取）、`confirm.vue` 结算选券弹窗（金额明细含券抵扣行，提交携带 `user_coupon_id`）、`my-coupons.vue` 我的券（未使用/已使用/已过期 Tab）。
- 说明：30 天唤回发券当前使用默认模板策略（取启用中可领的满减券），后续如需指定唤回模板可加配置项。

---

## 6. 阶段三：服务过程安全

### 6.1 数据表

**service_staffs 增列**
- `service_region` varchar(256)：服务区域（区县，逗号分隔），接单池按 `orders.delivery_district` 匹配过滤。
- `quality_score` decimal(3,1)：服务质量分（阶段四写入）。

**service_records 服务记录表（按 PRD 5.4）**

| 字段 | 说明 |
|------|------|
| id, order_id(unique), staff_id | |
| start_time / end_time | 签到/签退时间（与 orders 实际字段冗余落一份快照） |
| gps_track_url | 轨迹聚合文件URL（可选，明细走轨迹点表） |
| audio_url | 服务录音URL（七牛） |
| audio_uploaded_at | 录音上传时间（自动删除任务的判断依据） |
| sos_triggered | 是否触发SOS 0/1 |
| status | 1=正常 2=异常 |

**service_location_tracks 轨迹点表（高频写）**
- id, order_id(index), staff_id(index), lat decimal(10,6), lng decimal(10,6), reported_at(index)

**service_alert_events 预警事件表（统一 SOS 与超时预警）**

| 字段 | 说明 |
|------|------|
| id, order_id(nullable), staff_id | |
| alert_type | 1=SOS求助 2=服务超时未结束 |
| lat / lng / address | 触发位置 |
| status | 1=待处理 2=处理中 3=已处理 |
| handler_id / handle_remark / handled_at | 处理留痕 |

### 6.2 接口清单

**服务人员端**

| 接口 | 说明 |
|------|------|
| POST /api/v1/service-staff/workorders/:id/location | 服务中定时上报定位（前端 60s 一次） |
| POST /api/v1/service-staff/workorders/:id/audio | 上传录音文件（复用七牛上传） |
| POST /api/v1/service-staff/sos | 一键SOS（携带当前定位，写预警事件） |
| GET /api/v1/service-staff/my-region | 查看/申请变更服务区域（变更走审核，复用现有审核通道） |

**后台（RBAC 保护）**

| 接口 | 说明 | 权限码 |
|------|------|--------|
| GET /api/v1/merchant/alert-events | 预警事件列表（按类型/状态筛选） | alert-events:view |
| POST /api/v1/merchant/alert-events/:id/handle | 处理预警（状态+备注留痕） | alert-events:update |
| 服务人员管理接口扩展 | 服务区域字段维护 | 沿用 service-staff 权限码 |

### 6.3 业务规则
- **接单池区域过滤**：`getPendingOrders` 仅返回 `delivery_district` 落在人员 `service_region` 内的订单；区域为空视为不限。
- **录音授权与生命周期**：服务开始前弹隐私授权；录音文件完结时随 `checkOut` 一并提交（签名/上传复用现有七牛通道，`audio_url` 由 service_records 承载）。**录音仅保留 1 个月**：每日定时任务清理 `audio_uploaded_at` 超过 30 天的录音——先调七牛删除文件，成功后清空 `audio_url` 并写删除标记，删除失败次日重试。
- **超时预警定时任务**（每 10 分钟）：`ActualStartedAt` 超过阈值（系统配置，默认 4 小时）且未签退的服务订单 → 写入 alert_type=2 预警事件。
- SOS 事件同时推送后台（复用现有 WS 消息通道，若可用则加红点+声音提醒）。

### 6.4 前端改造
- staff-miniprogram：工单详情（服务中）增加 SOS 悬浮按钮、录音开始/停止、后台定时上报定位；个人页展示服务区域。
- web-admin：新增「预警中心」视图（事件列表+处理）；服务订单详情增加服务记录（录音播放、轨迹时间轴列表展示）。

### 6.5 协议管理（后台新增，支撑录音/定位授权）

**数据表 `agreements` 协议表**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | |
| type | tinyint | 协议类型：1=用户协议 2=隐私政策 3=录音/定位授权协议（后续可扩展服务人员协议等） |
| title | varchar(128) | 协议标题 |
| content | text | 协议正文（富文本） |
| version | varchar(32) | 版本号（如 v1.2，同类型递增） |
| status | tinyint | 1=已发布（当前生效） 0=草稿/停用 |
| published_at | datetime | 发布时间 |
| created_at / updated_at | | |

**业务规则**
- 同一 `type` 仅一份 `status=1` 的协议生效；发布新版本时旧版本自动置为停用（保留历史版本可查）。
- C 端与服务人员端授权弹窗（录音、定位、隐私政策、用户协议）统一从协议接口拉取当前生效版本展示，不再硬编码文案。
- 用户首次同意时记录授权留痕（user_id、协议版本、同意时间，轻量表 `agreement_consents`），后续版本更新后需重新确认。

**接口清单**

| 端 | 接口 | 说明 |
|----|------|------|
| 后台 | GET/POST /api/v1/merchant/agreements | 协议列表/新建（agreements:view / create） |
| 后台 | PUT /api/v1/merchant/agreements/:id | 编辑草稿（agreements:update） |
| 后台 | POST /api/v1/merchant/agreements/:id/publish | 发布版本（旧版自动停用）（agreements:update） |
| C 端 | GET /api/v1/store/agreements/:type/current | 按类型取当前生效协议 |
| C 端 | POST /api/v1/store/agreements/:id/consent | 记录用户同意留痕 |
| 服务人员端 | GET /api/v1/service-staff/agreements/:type/current | 按类型取当前生效协议 |

**前端改造**
- web-admin：系统设置下新增「协议管理」视图（类型筛选 + 版本列表 + 富文本编辑 + 发布）。
- C 端：登录页/授权弹窗协议链接改为动态拉取；录音、定位授权弹窗展示 type=3 协议并要求确认后才开始采集。
- 服务人员端：接单前隐私与录音授权弹窗同上动态拉取。

### 6.6 迁移脚本 `008_service_safety.sql`
- `service_staffs` 增列、新建三张表、`agreements`+`agreement_consents` 建表、`sys_menus`「预警中心」「协议管理」菜单与权限节点。

### 6.7 验收要点
- 服务中工单按分钟级上报轨迹，后台可见时间轴。
- SOS 点击后 5 秒内后台预警列表出现待处理事件。
- 超时未签退订单被定时任务捕获并生成预警。
- 区域外人员看不到该区县服务订单。
- 超过 30 天的录音被定时任务自动删除（七牛文件与 URL 同步清理），1 个月内的录音可正常播放。
- 后台可新建/发布用户协议、隐私政策、录音授权协议；发布新版本后旧版自动停用。
- C 端与服务人员端授权弹窗动态展示当前生效协议版本；未确认协议前录音/定位采集不启动；同意行为留痕可查。

### 6.8 实施记录（2026-08-31）
- 后端模型：`ServiceRecord` / `ServiceLocationTrack` / `ServiceAlertEvent` / `Agreement` / `AgreementConsent`，`ServiceStaff` 增 `ServiceRegion` / `QualityScore`。
- 服务人员端：工单维度定位上报（60s）、录音文件提交（七牛，MIME 校验）、一键 SOS（写预警事件+服务记录标记）、my-region 查询；接单池按 `service_region` ↔ `delivery_district` 过滤；checkOut 签退时写入/更新 service_records。
- 后台：预警事件列表（类型/状态/关键字筛选）+ 处理（状态流转+备注留痕）、协议 CRUD + 发布（同类型互斥，旧版自动下线）、服务人员服务区域维护、订单服务记录查询（签到签退/时长/录音播放/轨迹明细）。
- C 端：协议按类型取当前生效版本 + 同意留痕接口（C 端与服务人员端共用 agreement_consents）。
- 定时任务：超时预警（`ActualStartedAt` 超阈值未签退 → alert_type=2，每 10 分钟）、录音清理（30 天保留期，先删七牛文件再清 URL，失败次日重试）。
- 迁移 `008_service_safety.sql` 已执行：增列 + 5 张新表 + sys_menus 节点（97 预警中心 / 76 协议管理及按钮权限）+ 超管角色授权。
- web-admin：「预警中心」视图、「协议管理」（系统设置下，textarea 富文本）视图、订单详情服务记录卡片（录音播放/轨迹折叠表）、服务人员列表服务区域列 + 维护弹窗。
- staff-miniprogram：工单详情 SOS 悬浮按钮、录音开始/停止（`utils/safety` 定时上报定位 + 签退随 checkOut 提交录音）、个人页展示服务区域。
- 验证：server `go build` ✅、web-admin `vue-tsc + vite build` ✅、staff-miniprogram `build:mp-weixin` ✅、迁移执行与表/菜单/授权核验 ✅。

---

## 7. 阶段四：评价与服务质量分（完结触发）

### 7.1 触发流程（用户指定）
```
服务人员工单详情点「完结」(checkOut)
  → orders.biz_status=5 已完成，写 service_records
  → 服务端向下单用户推送下发消息（微信订阅消息，模板：服务已完成邀请评价）
  → C端同步出现"待评价"标记：订单列表/详情"待评价"角标 + 首页红点（不依赖消息触达）
  → 用户提交评价 → service_reviews 落库 → 重算人员质量分
```

> 推送即消息下发动作：本期先保证 C 端"待评价"标记与评价流程闭环，订阅消息模板 ID 后续申请配置后接入，不阻塞本阶段。

### 7.2 数据表

**service_reviews 评价表**

| 字段 | 说明 |
|------|------|
| id, order_id(unique), user_id, staff_id | 一单一评 |
| score | 总体评分 1-5 |
| attitude_score / professional_score / punctual_score | 态度/专业/准时分项 1-5 |
| content / images(json) | 文字与图片 |
| status | 1=正常展示 0=后台隐藏 |
| created_at | |

**service_staffs.quality_score**：近 100 条评价加权平均（总分 60% + 分项各 10%+10%+20%），评价提交后异步重算。

### 7.3 接口清单

| 端 | 接口 | 说明 |
|----|------|------|
| C端 | POST /api/v1/user/orders/:order_id/review | 提交评价（校验：本人+服务订单+已完成+未评过） |
| C端 | GET /api/v1/user/orders/:order_id/review | 查看评价（未评价返回 null） |
| 后台 | GET /api/v1/merchant/service-reviews | 评价列表（service-reviews:view） |
| 后台 | POST /api/v1/merchant/service-reviews/:id/hide | 隐藏违规评价（service-reviews:update） |
| 服务人员端 | GET /api/v1/service-staff/my-quality-score | 我的质量分与近期评价 |

### 7.4 前端改造
- C 端：评价提交页（星级+标签+文字+图片）、订单详情评价展示、订阅消息授权（下单时申请一次，模板接入后生效）。
- staff-miniprogram：个人页展示质量分与近期评价。
- web-admin：「服务评价」管理视图。

### 7.5 迁移脚本 `009_service_review.sql`
- `service_reviews` 建表 + `sys_menus`「服务评价」菜单节点。

### 7.6 验收要点
- 完结后 C 端可评价（不依赖消息触达，"待评价"标记可见即可评价）。
- 同一订单不可重复评价；非本人/未完成订单不可评价。
- 质量分随新评价实时更新；后台可隐藏违规评价。

### 7.7 实施记录（2026-08-31）
- 后端：`ServiceReview` 模型 + `Order.CanReview` 待评价标记字段；`services/review.RecalculateQualityScore` 质量分重算（近 100 条正常展示评价，总体 60% + 态度 10% + 专业 10% + 准时 20% 加权平均，保留 1 位小数）。
- 后端接口：C 端 `SubmitReview`（校验本人+服务订单+biz_status=5 已完成+一单一评+有服务人员，落库后异步重算质量分）与 `GetReview`（未评价返回 null）；merchant 端 `GetServiceReviews` 列表（状态/评分/关键字筛选）+ `HideServiceReview` 隐藏（隐藏后重算质量分）；service-staff 端 `GetMyQualityScore`（质量分 + 平均分项 + 近期评价）。
- 订单待评价标记：订单详情查询时服务订单 biz_status=5 且未评价 → `can_review=true`；首页接口返回 `pending_review_count`（已完成未评价服务订单数，供首页红点）。
- 迁移 `009_service_review.sql` 已执行：`service_reviews` 表（一单一评唯一索引）+ sys_menus 节点 98/981 + 超管角色授权。
- C 端小程序：`review-submit.vue` 评价提交页（4 维度星级 + 文字 + 图片上传）、`order-detail.vue` 评价展示与「去评价」按钮、订单列表「待评价」角标 + 首页红点。
- staff-miniprogram：`profile/index.vue` 个人页展示质量分与近期评价。
- web-admin：「服务评价」管理视图（列表 + 详情 + 隐藏评价）。
- 验证：server `go build ./...` ✅、web-admin `vue-tsc + vite build` ✅、miniprogram `build:mp-weixin` ✅、staff-miniprogram `build:mp-weixin` ✅、迁移执行与表/菜单/授权核验 ✅。

---

## 8. 阶段五：健康服务板块重构

> 三个子任务相互独立，可按 8.1 → 8.2 → 8.3 顺序实施；8.1 抹除操作有数据备份前置条件。

### 8.1 抹除照护计划、随访任务、生命体征

**抹除范围（全部功能与数据）**

| 层 | 对象 |
|----|------|
| 数据表 | `care_plans`（照护计划）、`care_visits`（照护记录）、`follow_up_tasks`（随访任务）、`health_monitoring`（生命体征） |
| 后端模型 | models.go 中 CarePlan / CareVisit / FollowUpTask / HealthMonitoring 结构体 |
| 后端接口 | `handlers/health/care.go`（照护）、`handlers/health/followup.go`（随访+体征）全部路由 |
| 后端服务 | `services/followup/auto.go`（服务完成/租赁归还/评估完成自动生成随访任务的钩子，从订单完结逻辑中移除调用） |
| C端页面 | `care-plan-detail.vue`、`my-care-plans.vue`、`my-follow-ups.vue`、`my-monitoring.vue` 删除；`my-health.vue` 入口清理；`pages.json` 路由移除 |
| 员工端页面 | `care-plans.vue`、`care-plan-detail.vue`、`care-visit-form.vue`、`follow-ups.vue`、`follow-up-detail.vue` 删除；工单详情「录入照护记录」入口清理；`pages.json` 路由移除 |
| 后台视图 | `CarePlansView.vue`、`FollowUpTasksView.vue`、`HealthMonitoringView.vue` 删除；路由移除 |
| RBAC | `sys_menus` 中照护计划/随访任务/生命体征相关菜单与按钮权限节点删除 |

**保留不动的健康功能**：健康档案（health_records）、评估量表（health_assessment_forms）、评估记录（health_assessments）、适配建议（fitting_recommendations）。

**迁移脚本 `010_health_module_cleanup.sql`**
- 先在备份库导出 4 张表全量数据（`server/backups/` 已有备份机制，删除前再做一次专项备份）。
- `DROP TABLE` 4 张表；`DELETE` sys_menus 相关节点。
- 脚本头部 `SET NAMES utf8mb4;`。

**验收要点**
- 三端均无照护/随访/体征任何入口与页面；相关接口返回 404。
- 服务完成、租赁归还、评估完成后不再自动生成随访任务（原有钩子调用点清理干净，编译无残留引用）。
- 评估、适配、档案功能回归正常不受影响。

### 8.2 多档案（一个账号绑定多位老人）

**数据模型改造（复用 health_records 表，直接改模型，不写迁移脚本）**

| 变更 | 说明 |
|------|------|
| `user_id` 唯一索引 → 普通索引 | 解除一账号一档案限制（模型层直接改，现有测试数据不保留） |
| 新增 `relation` tinyint | 与账号关系：1=本人 2=父母 3=其他亲属 |
| 评估/适配挂接 | `health_assessments`、`fitting_recommendations` 新增 `record_id`（档案ID）字段，查询按档案维度 |

**权限边界（按 PRD 3.1）**
- 用户本人：可编辑账号下所有档案的全部字段。
- 服务人员：仅可更新所服务档案的**健康数值/评估结果**，不可修改基础信息（姓名/身份证/联系方式只读）。

**接口调整**
- C 端：档案列表/新增/编辑/删除/切换（`/api/v1/user/health-records` 扩展为列表 CRUD，评估接口带 `record_id`）。
- 员工端：residents 相关接口按 `record_id` 维度查询与更新。
- 后台：健康档案视图按账号→档案两级展示。

**与服务订单的联动（承接阶段一 4.1 的 `orders.record_id`）**
- C 端 `confirm.vue` "服务对象"栏从"单档案默认选中"升级为**档案列表选择**（展示姓名+关系，如"李奶奶·父母"）。
- 员工端工单详情展示服务对象档案摘要（姓名/性别/年龄/紧急联系人/慢病标签/过敏史），接单前即可确认服务对象；服务中更新健康数值直接写入该档案（受权限边界约束）。
- 档案被服务订单引用后**不可删除**（软删除/归档需提示"存在关联服务订单"）。

**前端改造**
- C 端 `my-health.vue` 改造为档案列表+当前档案切换；`health-record-edit.vue` 支持新增档案（选择关系：本人/父母/其他亲属）；评估入口绑定当前档案。
- 员工端 `resident.vue` 支持在账号下多档案间选择。

**实施方式**
- 模型结构直接修改（models.go），表结构跟随新模型重建；开发期数据为测试数据，不做存量回填。
- 后续迁移脚本编号从宣教独立（原 012）顺延调整。

**验收要点**
- 一个账号可添加多位老人档案并互相切换；评估结果落在所选档案下。
- 服务订单下单时从档案列表选择服务对象（多档案时不可跳过）；订单详情与员工端工单展示所选档案。
- 被服务订单引用的档案删除被拒绝。
- 服务人员只能更新健康数值，改基础信息被拒绝。

### 8.3 健康宣教独立板块（分类+内容）

**数据模型**

`health_education_categories` 宣教分类表（新建）

| 字段 | 说明 |
|------|------|
| id, name varchar(64), sort, status(1启用0停用), created_at, updated_at | 支持两级分类（parent_id） |

`health_education_articles` 改造

| 变更 | 说明 |
|------|------|
| 新增 `category_id` bigint index | 关联宣教分类，替换原字符串 `category` 字段 |
| 存量迁移 | 按现有字符串 `category` 去重生成分类记录并回填 `category_id` |
| `tags` 慢病标签 | 保留字段，仅作为文章附加属性，不再承担板块组织职责 |

**业务规则**
- 宣教从"健康服务"链路中独立出来，成为独立板块：用户主动按分类浏览查看，不与评估/照护流程耦合。
- 后台「健康宣教」独立菜单：分类管理（树形 CRUD）+ 内容管理（文章 CRUD、挂分类、发布/下架）。
- C 端独立宣教入口（首页宫格入口 + 我的页面入口）：分类列表 → 分类下文章列表 → 文章详情（复用现有 `health-education.vue`、`education-detail.vue` 改造）。

**接口清单**

| 端 | 接口 | 说明 |
|----|------|------|
| 后台 | GET/POST /api/v1/merchant/education-categories | 分类列表/创建（education-categories:view/create） |
| 后台 | PUT/DELETE /api/v1/merchant/education-categories/:id | 更新/删除（education-categories:update/delete） |
| 后台 | 文章接口扩展 | 新增/编辑文章必填 `category_id`（沿用现有权限码） |
| C 端 | GET /api/v1/store/education/categories | 启用中的分类列表 |
| C 端 | GET /api/v1/store/education/articles?category_id= | 按分类查文章（不传=全部） |

**迁移脚本 `011_education_categories.sql`**
- 建分类表、文章表加 `category_id`、存量分类字符串回填、`sys_menus`「健康宣教」独立菜单与权限节点。

**验收要点**
- 后台可维护两级宣教分类；文章必须挂分类才能发布。
- C 端从独立入口进入，按分类浏览文章详情；原"我的健康"内的宣教入口移除。
- 宣教与评估/档案流程无任何耦合。

---

## 9. 第二期展望（本期不实施）
- 结算提现体系：T+0 冻结/T+1 可提现、财务角色提现审核、售后冲抵待结算余额、押金独立账户核算。
- 亏损订单预警（净利润<0 推送运营复核）、组合套餐真实子项（扣库存+等比分摊+子项分润）、手机号一键登录、操作日志与区域数据范围、订阅消息模板接入与消息中心。

## 10. 跨阶段工程约定
1. **四步流程**：接口开发 → 按钮权限码（前端按钮码=后端鉴权码=sys_menus.permission）→ 迁移脚本（含菜单节点，`SET NAMES utf8mb4;`，需在数据库执行）→ PRD 同步（`PRD.md` 摘要 + `docs/prd/PRD-功能说明.md` + `PRD-接口文档.md`）。
2. 权限码命名：页面级 `:view`，按钮级 `create/update/delete`，不聚合。
3. 服务人员相关表统一 `service_` 前缀。
4. 每阶段完成后小步提交版本控制；涉及超过 10 个文件时及时提交。
5. 不修改路由文件结构、不动自营商户数据（merchants.id=1）。
6. 抹除类操作（阶段五 8.1）必须先做数据库专项备份再执行 DROP，备份文件留存于 `server/backups/`。

## 11. 风险与依赖
- **数据抹除不可逆（阶段五 8.1）**：照护/随访/体征 4 张表 DROP 前必须专项备份并验证可恢复；同时需清理订单完结逻辑中的随访自动生成钩子，避免运行时报错。
- **多档案改造（阶段五 8.2）**：模型直接改、表结构跟随重建，不做存量数据迁移与回填（开发期数据为测试数据）。
- **订阅消息**：本期推送即消息下发动作，先保证 C 端"待评价"标记闭环；模板 ID 后续申请配置后接入。
- **轨迹表写入频率**（60s/单）在并发工单量大时需关注表容量，预留按月分表/归档策略。
- **录音合规**需用户端与服务人员端双向授权提示，协议文案统一由后台「协议管理」维护（agreements 表），发布新版本后需用户重新确认；录音保留 1 个月由定时任务自动删除。
- 优惠券抵扣后退款：已核销券一律**不支持退还**，规则直接固化，不做退券扩展。
