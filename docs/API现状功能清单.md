# API 现状功能清单

> 说明：本文档基于当前项目后端代码（`server/internal/handlers` 等）反向生成，非 PRD 文档。所有接口前缀为 `/api/v1`。

## 总览

| 模块              | 接口数 | 说明                        |
| --------------- | --- | ------------------------- |
| 认证登录            | 3   | 商家/用户/服务人员 3 端登录          |
| C端店铺(公开)        | 14  | 首页/商品/优惠券/埋点/健康宣教公开浏览     |
| C端用户中心          | 28  | 订单/地址/档案/评估/适配/协议         |
| 商家-基础账号         | 10  | 资料/设置/密码/状态/二维码/微信绑定/支付配置 |
| 商家-员工与人员        | 13  | 后台员工、服务人员、审核中心            |
| 商家-打印机          | 5   | 飞鹅打印机管理                   |
| 商家-订单           | 12  | 列表/派单/核销/退款/退租/续租/统计/服务记录 |
| 商家-商品/分类/规格     | 18  | 三级分类、商品、规格                |
| 商家-优惠券/轮播图      | 13  | 券模板、轮播图                   |
| 商家-运营分析         | 7   | 销售/商品/顾客分析                |
| 商家-预警/协议/评价/分账  | 21  | 预警中心、协议管理、评价、分账           |
| 健康服务-商家端        | 21  | 档案/评估/宣教/适配               |
| RBAC 系统管理       | 16  | 菜单/角色/部门/权限               |
| 服务人员接单小程序       | 31  | 注册/接单/工时/安全/健康            |
| 支付回调/上传/WS/开发推送 | 7   | 回调、上传、协议、微信登录、WS          |

***

## 1. 认证登录

| 接口路径                        | 方法   | 入参                                   | 出参                                         | 核心业务逻辑简述                                                   | 依赖的外部表                                                                         |
| --------------------------- | ---- | ------------------------------------ | ------------------------------------------ | ---------------------------------------------------------- | ------------------------------------------------------------------------------ |
| /auth/merchant/login        | POST | LoginRequest(username\*, password\*) | token、merchant\_id、staff、menus、permissions | 校验后台员工账号密码，bcrypt验证，签发商家JWT并加载RBAC菜单权限                     | merchant\_staffs、sys\_roles、merchant\_staff\_roles、sys\_role\_menus、sys\_menus |
| /auth/merchant/wechat-login | POST | WechatQuickLoginRequest(code\*)      | token、merchant\_id、staff、menus、permissions | code换openid，查后台员工签发JWT并加载RBAC权限                            | merchant\_staffs(关联RBAC表)                                                      |
| /auth/user/wechat-login     | POST | WechatLoginRequest(code\*)           | token、app\_mode、app\_id、user               | C端小程序jscode2session换openid(dev\_前缀mock)，无则新建用户，更新访问信息签发JWT | users                                                                          |

## 2. C端店铺(公开路由,可选JWT)

| 接口路径                                 | 方法   | 入参                                                                 | 出参                                                                                  | 核心业务逻辑简述                                                                        | 依赖的外部表                                                                                                                                             |
| ------------------------------------ | ---- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| /store/home                          | GET  | 可选JWT                                                              | merchant、categories、hot\_products、delivery\_settings、banners、pending\_review\_count | 汇总首页数据，图片转七牛私有URL，登录时统计待评价红点                                                    | merchants、categories、products、merchant\_delivery\_settings、mini\_program\_banners、orders、service\_reviews                                          |
| /store/delivery-rules                | GET  | 无                                                                  | enabled、base\_fee、free\_delivery\_amount、max\_distance、distance\_rules              | 返回配送费规则，未配置给默认值                                                                 | merchant\_delivery\_settings                                                                                                                       |
| /store/products                      | GET  | category\_id、keyword、page、page\_size、product\_types                | list、pagination                                                                     | 上架商品筛选，聚合子孙分类，多值类型过滤，分页                                                         | products、categories、product\_specs                                                                                                                 |
| /store/products/:product\_id         | GET  | 路径product\_id                                                      | StoreProductResponse(含specs)                                                        | 预加载分类规格，按id+上架查详情                                                               | products、categories、product\_specs                                                                                                                 |
| /store/coupons/available             | GET  | 可选JWT                                                              | list(券模板含total/received/remain/can\_receive)                                        | 查找启用且领取窗口内的券模板，附已领数                                                             | coupon\_templates、user\_coupons                                                                                                                    |
| /store/visit                         | POST | openid、source；可选JWT                                                | user\_id、visit\_count                                                               | 创建/更新访问用户，写访问记录与行为事件，WS广播                                                       | users、user\_visits、user\_behavior\_events                                                                                                          |
| /store/event                         | POST | openid、event\_type、page、product\_id、order\_id、source、payload       | message                                                                             | 校验埋点枚举后写行为事件                                                                    | users、user\_behavior\_events                                                                                                                       |
| /store/orders                        | POST | CreateOrderRequest(items、address\_id、record\_id、user\_coupon\_id等) | order、pay\_params、pay\_hint                                                         | 核心下单：校验库存/规格，按商品类型定order\_type/biz\_status，服务单强制本人档案，事务建单+核销券+原子扣库存，异步微信JSAPI支付 | merchants、users、products、product\_specs、orders、order\_items、user\_coupons、coupon\_templates、health\_records、user\_addresses、user\_behavior\_events |
| /store/coupons/:template\_id/receive | POST | 路径template\_id                                                     | user\_coupon\_id、expired\_at                                                        | 校验券模板并发放用户券                                                                     | user\_coupons、coupon\_templates                                                                                                                    |
| /store/coupons/usable                | GET  | total\_amount、product\_ids、category\_ids                           | list(最优在前)                                                                          | 按抵扣额排序返回下单可用券                                                                   | user\_coupons、coupon\_templates、products                                                                                                           |
| /store/my-coupons                    | GET  | status、page、page\_size                                             | list、total                                                                          | 惰性过期置过期，按状态分页                                                                   | user\_coupons、coupon\_templates                                                                                                                    |
| /store/education/categories          | GET  | 无                                                                  | HealthEducationCategory数组                                                           | 返回启用宣教分类(首页公开)                                                                  | health\_education\_categories                                                                                                                      |
| /store/education/articles            | GET  | page、page\_size、category\_id                                       | list、total                                                                          | 已发布文章按分类/关键词分页                                                                  | health\_education\_articles                                                                                                                        |
| /store/education/articles/:id        | GET  | 路径id                                                               | HealthEducationArticle                                                              | 已发布文章详情+浏览量                                                                     | health\_education\_articles                                                                                                                        |

## 3. C端用户中心(需登录)

| 接口路径                                      | 方法     | 入参                                            | 出参                                  | 核心业务逻辑简述                                   | 依赖的外部表                                                               |
| ----------------------------------------- | ------ | --------------------------------------------- | ----------------------------------- | ------------------------------------------ | -------------------------------------------------------------------- |
| /user/orders                              | GET    | status、category(1实物/2服务)、page、page\_size      | list、pagination                     | 按用户+分类过滤订单，填充服务对象与待评价标记                    | orders、order\_items、products、service\_staffs、service\_reviews        |
| /user/orders/:order\_id                   | GET    | 路径order\_id                                   | Order(含Items、纪录信息)                  | 填充健康档案服务对象、图片URL                           | orders、order\_items、health\_records、service\_staffs、service\_reviews |
| /user/orders/:order\_id/cancel            | POST   | 路径order\_id                                   | message                             | 仅未支付可取消，释放优惠券，置status=4                    | orders、user\_coupons                                                 |
| /user/orders/:order\_id/refund            | POST   | ApplyRefundRequest(reason\*)                  | refund                              | 校验已支付未退款，建退款记录置status=5，同步微信退款             | orders、refunds、merchants、user\_coupons                               |
| /user/orders/:order\_id/renew             | POST   | RenewOrderRequest(duration)                   | order、pay\_params、pay\_hint         | 租赁续租：重算租金复制生成新单(parent/renew\_flag)，生成支付参数 | orders、order\_items、users、merchants                                  |
| /user/orders/:order\_id/review            | GET    | 路径order\_id                                   | service\_review/null                | 查本人评价                                      | service\_reviews                                                     |
| /user/orders/:order\_id/review            | POST   | SubmitReviewRequest(score、维度分、content、images) | message、id                          | 校验服务订单已完成，一单一评，创建评价并异步重算质量分                | service\_reviews、orders、service\_staffs                              |
| /user/addresses                           | GET    | 无                                             | 地址列表(默认在前)                          | 按用户查地址                                     | user\_addresses                                                      |
| /user/addresses                           | POST   | addressRequest                                | 新地址                                 | 建地址，设默认则清其他默认                              | user\_addresses                                                      |
| /user/addresses/:id                       | PUT    | 路径id+addressRequest                           | 更新后地址                               | 校验归属后更新                                    | user\_addresses                                                      |
| /user/addresses/:id                       | DELETE | 路径id                                          | message                             | 归属校验后删除                                    | user\_addresses                                                      |
| /user/agreements                          | GET    | type(默认3)                                     | agreement、consented                 | 拉取最新已发布协议(用户协议/隐私/录音)并判断是否同意               | agreements、agreement\_consents                                       |
| /user/agreements/:id/consent              | POST   | 路径id                                          | agreement\_id、version、consented\_at | 幂等写协议同意留痕(user\_type=1)                    | agreements、agreement\_consents                                       |
| /user/health-record                       | GET    | 无                                             | HealthRecord                        | 取账号最近更新档案(兼容单档案)                           | health\_records                                                      |
| /user/health-record                       | PUT    | UpsertRequest                                 | HealthRecord                        | 单档案upsert                                  | health\_records                                                      |
| /user/health-records                      | GET    | 无                                             | HealthRecord数组                      | 多档案倒序                                      | health\_records                                                      |
| /user/health-records                      | POST   | UpsertRequest(relation等)                      | HealthRecord                        | 校验<5个，新建档案                                 | health\_records                                                      |
| /user/health-records/:id                  | PUT    | 路径id+UpsertRequest                            | HealthRecord                        | 归属校验后覆盖更新                                  | health\_records                                                      |
| /user/health-records/:id                  | DELETE | 路径id                                          | {id}                                | 被订单引用则禁止删除                                 | health\_records、orders                                               |
| /user/assessment-forms                    | GET    | 无                                             | 量表数组                                | 启用量表                                       | health\_assessment\_forms                                            |
| /user/assessments                         | GET    | page、page\_size、record\_id                    | list、total                          | 本人评估记录                                     | health\_assessments、health\_assessment\_forms                        |
| /user/assessments                         | POST   | CreateAssessmentRequest(form\_id\*)           | HealthAssessment                    | 算分/等级/结论，回写档案等级                            | health\_assessment\_forms、health\_records、health\_assessments        |
| /user/fitting-recommendations             | GET    | 分页                                            | list、total                          | 本人适配建议                                     | fitting\_recommendations                                             |
| /user/fitting-recommendations/:id         | GET    | 路径id                                          | VO                                  | 归属校验                                       | fitting\_recommendations                                             |
| /user/fitting-recommendations/:id/confirm | POST   | 路径id                                          | VO                                  | 草稿→已确认                                     | fitting\_recommendations                                             |
| /user/health-education                    | GET    | category\_id、category、keyword                 | 文章数组                                | 已发布文章筛选                                    | health\_education\_articles                                          |
| /user/health-education/:id                | GET    | 路径id                                          | 文章                                  | 详情+浏览量                                     | health\_education\_articles                                          |

## 4. 商家-基础账号/员工/服务人员/审核/打印机

| 接口路径                                       | 方法     | 入参                                   | 出参                                 | 核心业务逻辑简述                         | 依赖的外部表                                                              |
| ------------------------------------------ | ------ | ------------------------------------ | ---------------------------------- | -------------------------------- | ------------------------------------------------------------------- |
| /merchant/profile                          | GET    | 无                                    | staff、merchant                     | 单商户资料，图片私有URL                    | merchants、merchant\_staffs                                          |
| /merchant/profile                          | PUT    | UpdateProfileRequest                 | merchant                           | 非空字段更新单行                         | merchants                                                           |
| /merchant/settings                         | GET    | 无                                    | 经营开关/微信绑定/配送设置                     | 汇总返回                             | merchants、merchant\_delivery\_settings、merchant\_staffs             |
| /merchant/settings                         | PUT    | UpdateSettingsRequest(指针bool)        | message                            | 经营开关写merchants、提示音写员工            | merchants、merchant\_staffs                                          |
| /merchant/account/change-password          | POST   | old\_password\*、new\_password\*      | message                            | bcrypt校验并更新                      | merchant\_staffs                                                    |
| /merchant/account/wechat/bind              | POST   | BindWechatRequest(code\*)            | openid、unionid等                    | code换openid校验占用后回写               | merchant\_staffs                                                    |
| /merchant/account/wechat/bind              | DELETE | 无                                    | message                            | 清空绑定                             | merchant\_staffs                                                    |
| /merchant/status                           | POST   | StatusRequest(status\*指针)            | message                            | 更新营业状态(指针允许0)                    | merchants                                                           |
| /merchant/qrcode                           | GET    | 无                                    | qrcode\_url、scene、page、placeholder | 生成店铺小程序码                         | merchants                                                           |
| /merchant/payment-config                   | PUT    | PaymentConfigRequest(sub\_mch\_id\*) | message+merchant                   | 回填子商户号置payment\_config\_status=1 | merchants                                                           |
| /merchant/staff                            | GET    | page、page\_size、keyword              | list、pagination                    | 员工分页，预加载角色部门                     | merchant\_staffs、merchant\_staff\_roles、sys\_roles、sys\_departments |
| /merchant/staff                            | POST   | CreateStaffRequest                   | {id}、message                       | 唯一校验+bcrypt，事务建员工与角色，清RBAC缓存     | merchant\_staffs、merchant\_staff\_roles                             |
| /merchant/staff/:id                        | PUT    | 路径id+UpdateStaffRequest              | staff                              | 更新字段/角色，owner保留校验                | merchant\_staffs、merchant\_staff\_roles                             |
| /merchant/staff/:id                        | DELETE | 路径id                                 | message                            | 禁删自己/owner保留，事务删                 | merchant\_staffs、merchant\_staff\_roles                             |
| /merchant/staff/:id/reset-password         | POST   | new\_password\*                      | message                            | bcrypt重置                         | merchant\_staffs                                                    |
| /merchant/service-staff                    | GET    | status、keyword、page、page\_size       | list、pagination                    | 服务人员分页                           | service\_staffs                                                     |
| /merchant/service-staff                    | POST   | CreateServiceStaffRequest            | {id}                               | 校验唯一+默认启用，写注册审核留痕                | service\_staffs、service\_staff\_audit\_records                      |
| /merchant/service-staff/:id/status         | PUT    | 路径id+status                          | {id,status}                        | 启用/禁用+状态变更留痕                     | service\_staffs、service\_staff\_audit\_records                      |
| /merchant/service-staff/:id/reset-password | POST   | new\_password\*                      | {id}                               | bcrypt重置                         | service\_staffs                                                     |
| /merchant/service-staff/:id                | DELETE | 路径id                                 | {id}                               | 删除                               | service\_staffs                                                     |
| /merchant/service-staff/:id/service-region | PUT    | 路径id+service\_region\*               | {id,service\_region}               | 维护可服务区县                          | service\_staffs                                                     |
| /merchant/staff-audits                     | GET    | audit\_type、status、keyword、page      | list、total                         | 审核记录联表分页                         | service\_staff\_audit\_records、service\_staffs                      |
| /merchant/staff-audits/:id                 | GET    | 路径id                                 | record、staff、qualifications        | 审核详情+资质快照                        | service\_staff\_audit\_records、service\_staffs                      |
| /merchant/staff-audits/:id/approve         | POST   | 路径id、remark可选                        | {id,status}                        | 通过：注册启用账号/信息变更回写after\_data      | service\_staff\_audit\_records、service\_staffs                      |
| /merchant/staff-audits/:id/reject          | POST   | 路径id、remark可选                        | {id,status}                        | 驳回+清除待审核状态                       | service\_staff\_audit\_records、service\_staffs                      |
| /merchant/printers                         | GET    | 无                                    | list(含has\_feie\_ukey)             | 打印机列表默认优先                        | printers                                                            |
| /merchant/printers/:id                     | GET    | 路径id                                 | printer                            | 详情                               | printers                                                            |
| /merchant/printers                         | POST   | CreatePrinterRequest                 | {id}                               | 创建，首台设默认                         | printers                                                            |
| /merchant/printers/:id                     | PUT    | 路径id+UpdatePrinterRequest            | printer                            | 更新，设默认去重                         | printers                                                            |
| /merchant/printers/:id                     | DELETE | 路径id                                 | message                            | 删默认则移给剩余                         | printers                                                            |
| /merchant/printers/:id/test                | POST   | 路径id                                 | message                            | 调飞鹅云Open\_printMsg测试             | printers                                                            |

## 5. 商家-订单

| 接口路径                                       | 方法   | 入参                                          | 出参                                     | 核心业务逻辑简述                       | 依赖的外部表                                                            |
| ------------------------------------------ | ---- | ------------------------------------------- | -------------------------------------- | ------------------------------ | ----------------------------------------------------------------- |
| /merchant/orders                           | GET  | status、order\_type、category、keyword、日期、page | list、pagination                        | 订单列表(实物/服务Tab)，刷新退款中状态，图转私有URL | orders、order\_items、refunds、users、health\_records                 |
| /merchant/orders/dispatchable-staff        | GET  | 无                                           | \[{id,name,phone}]                     | 已审核服务人员列表供派单                   | service\_staffs                                                   |
| /merchant/orders/rental-due                | GET  | due\_range、keyword、page                     | list(含is\_overdue、days\_left)          | 租赁到期/逾期订单，升序分页                 | orders、order\_items、users                                         |
| /merchant/orders/:order\_id                | GET  | 路径order\_id                                 | Order                                  | 详情+健康档案填充，退款中查微信退款同步           | orders、order\_items、refunds、health\_records                       |
| /merchant/orders/quick-complete            | POST | order\_id\*                                 | Order                                  | 状态验证后置完成                       | orders                                                            |
| /merchant/orders/:order\_id/complete       | POST | 路径order\_id+verify\_code\*                  | Order                                  | 校验核销码，置status=3完成              | orders                                                            |
| /merchant/orders/:order\_id/refund         | POST | 路径order\_id+refund\_amount、reason           | Refund                                 | 校验状态，建退款记录，调微信服务商退款并同步状态       | orders、refunds、merchants                                          |
| /merchant/orders/:order\_id/return         | POST | 路径order\_id+deduct\_amount、remark           | Order                                  | 租赁退押金：押金-扣除，先建退款记录再微信退款        | orders、order\_items、refunds、merchants                             |
| /merchant/orders/:order\_id/dispatch       | POST | 路径order\_id+staff\_id\*                     | {id、assigned\_staff\_id、biz\_status:2} | 派单给已审核服务人员，biz\_status=2+派单留痕  | orders、service\_staffs、service\_staff\_audit\_records             |
| /merchant/orders/:order\_id/renew          | POST | 路径order\_id+duration                        | {order}                                | 续租复制生成新单+items                 | orders、order\_items                                               |
| /merchant/orders/statistics                | GET  | 无                                           | total/today/pending/completed/refunded | 订单/销售额/退款额统计                   | orders、refunds                                                    |
| /merchant/orders/:order\_id/service-record | GET  | 路径order\_id                                 | service\_record、tracks、track\_count    | 服务记录+轨迹+录音私有URL                | orders、service\_records、service\_location\_tracks、service\_staffs |

## 6. 商家-商品/分类/规格/优惠券/轮播图

| 接口路径                                     | 方法     | 入参                                                         | 出参                          | 核心业务逻辑简述                            | 依赖的外部表                                |
| ---------------------------------------- | ------ | ---------------------------------------------------------- | --------------------------- | ----------------------------------- | ------------------------------------- |
| /merchant/categories                     | GET    | parent\_id可选                                               | 树(含children、product\_count) | 三级分类树BuildTree+商品数                  | categories、products                   |
| /merchant/categories                     | POST   | CategoryRequest(name\*、parent\_id等)                        | Category                    | 校验父级、算level(≤3)、同级唯一                | categories                            |
| /merchant/categories/:category\_id       | PUT    | 路径id+CategoryRequest                                       | Category                    | 重名/移动父级/深度校验                        | categories                            |
| /merchant/categories/:category\_id       | DELETE | 路径id                                                       | message                     | 有子分类拒绝删除                            | categories                            |
| /merchant/categories/sort                | POST   | categories:\[{id,sort}]                                    | message                     | 事务批量更新sort                          | categories                            |
| /merchant/products                       | GET    | category\_id、status、keyword、sale\_type、product\_types、page | list、pagination             | 商品筛选分页，图私有URL                       | products、categories                   |
| /merchant/products/:product\_id          | GET    | 路径product\_id                                              | ProductResponse(含specs)     | 详情                                  | products、categories、product\_specs    |
| /merchant/products                       | POST   | ProductRequest                                             | ProductResponse             | 类型归一化(租赁sale\_type=2)，租赁校验，事务建商品+规格 | products、product\_specs               |
| /merchant/products/:product\_id          | PUT    | 路径id+ProductRequest                                        | ProductResponse             | 类型归一化+删旧规格重建                        | products、product\_specs               |
| /merchant/products/:product\_id/on-sale  | POST   | 路径id                                                       | message                     | 上架                                  | products                              |
| /merchant/products/:product\_id/off-sale | POST   | 路径id                                                       | message                     | 下架                                  | products                              |
| /merchant/products/batch-status          | POST   | product\_ids\*、status\*                                    | message                     | 批量更新状态                              | products                              |
| /merchant/products/:product\_id          | DELETE | 路径id                                                       | message                     | 软删除                                 | products                              |
| /merchant/products/:product\_id/stock    | PUT    | 路径id+stock\*                                               | message                     | 更新库存                                | products                              |
| /merchant/products/:product\_id/specs    | GET    | 路径id                                                       | specs、skus                  | 查规格                                 | products、product\_specs               |
| /merchant/products/:product\_id/specs    | PUT    | 路径id+specs、skus                                            | message                     | 删旧重建规格                              | products、product\_specs               |
| /merchant/products/:product\_id/specs    | DELETE | 路径id+spec\_ids\*                                           | message                     | 删指定规格                               | products、product\_specs               |
| /merchant/coupon-templates               | GET    | name、status、type、page                                      | list、total                  | 券模板分页                               | coupon\_templates                     |
| /merchant/coupon-templates/:id           | GET    | 路径id                                                       | 模板详情                        | 详情+受领数                              | coupon\_templates                     |
| /merchant/coupon-templates               | POST   | CouponTemplateRequest                                      | 模板                          | 满减折扣/有效期/范围校验                       | coupon\_templates                     |
| /merchant/coupon-templates/:id           | PUT    | 路径id+CouponTemplateRequest                                 | 模板                          | 校验+发行量≥已领数                          | coupon\_templates                     |
| /merchant/coupon-templates/:id/status    | POST   | 路径id+status                                                | 模板                          | 启停                                  | coupon\_templates                     |
| /merchant/coupon-templates/:id           | DELETE | 路径id                                                       | {id}                        | 有已领拒绝删除                             | coupon\_templates、user\_coupons       |
| /merchant/coupon-templates/:id/grant     | POST   | 路径id+user\_id                                              | {user\_coupon\_id}          | 手动发券(复用Receive source=3)            | coupon\_templates、user\_coupons、users |
| /merchant/user-coupons                   | GET    | user\_id、template\_id、status、source、page                   | list、total                  | 用户券记录分页                             | user\_coupons、coupon\_templates、users |
| /merchant/miniprogram-banners            | GET    | 无                                                          | list                        | 轮播图列表图私有URL                         | mini\_program\_banners                |
| /merchant/miniprogram-banners/:id        | GET    | 路径id                                                       | banner                      | 详情                                  | mini\_program\_banners                |
| /merchant/miniprogram-banners            | POST   | BannerCreateRequest(image\*、link\_type\*)                  | banner                      | 创建，默认启用                             | mini\_program\_banners                |
| /merchant/miniprogram-banners/:id        | PUT    | 路径id+BannerUpdateRequest(指针)                               | banner                      | 部分更新                                | mini\_program\_banners                |
| /merchant/miniprogram-banners/:id/status | PATCH  | 路径id+status                                                | nil                         | 启停                                  | mini\_program\_banners                |
| /merchant/miniprogram-banners/:id        | DELETE | 路径id                                                       | nil                         | 删除                                  | mini\_program\_banners                |

## 7. 商家-运营分析/预警/协议/评价/分账

| 接口路径                                        | 方法     | 入参                                               | 出参                       | 核心业务逻辑简述               | 依赖的外部表                                                             |
| ------------------------------------------- | ------ | ------------------------------------------------ | ------------------------ | ---------------------- | ------------------------------------------------------------------ |
| /merchant/analytics/overview                | GET    | period                                           | total\_sales等+增长率        | 按周期统计订单/客户/访问转化        | orders、user\_behavior\_events                                      |
| /merchant/analytics/sales-trend             | GET    | start\_date、end\_date、days                       | 逐日序列                     | 按日循环统计                 | orders、user\_behavior\_events                                      |
| /merchant/analytics/product-ranking         | GET    | limit、日期                                         | top N商品                  | 联表聚合销量额                | order\_items、orders                                                |
| /merchant/analytics/hourly                  | GET    | 无                                                | 24小时序列                   | 当日每小时统计                | orders                                                             |
| /merchant/analytics/stock-alert             | GET    | threshold                                        | 低库存商品                    | stock<=阈值且上架           | products                                                           |
| /merchant/analytics/customers               | GET    | 无                                                | 客户漏斗指标                   | 去重客户/新增/复购/转化          | orders、user\_behavior\_events                                      |
| /merchant/analytics/customer-trend          | GET    | days                                             | 逐日累计                     | 累计新用户/订单               | orders、users                                                       |
| /merchant/alert-events                      | GET    | alert\_type、status、keyword、page                  | list、total               | 预警事件过滤+填充              | service\_alert\_events、service\_staffs、orders                      |
| /merchant/alert-events/:id                  | GET    | 路径id                                             | 详情                       | 单条+关联                  | service\_alert\_events、service\_staffs、orders                      |
| /merchant/alert-events/:id/handle           | POST   | 路径id+status、remark                               | id、handler信息             | 状态流转留痕                 | service\_alert\_events                                             |
| /merchant/alert-settings                    | GET    | 无                                                | 预警配置                     | 单行配置(空则建)              | alert\_settings                                                    |
| /merchant/alert-settings                    | PUT    | 各阈值(指针)                                          | 配置                       | 指针允许零值，范围校验，upsert     | alert\_settings                                                    |
| /merchant/agreements                        | GET    | type、status、page                                 | list、total               | 协议列表(排除content)        | agreements                                                         |
| /merchant/agreements/:id                    | GET    | 路径id                                             | 详情(含content)             | 协议详情                   | agreements                                                         |
| /merchant/agreements                        | POST   | CreateAgreementRequest(type\*、title\*、version\*) | id                       | 草稿态创建+同类型版本唯一          | agreements                                                         |
| /merchant/agreements/:id                    | PUT    | 路径id+title、content                               | id                       | 仅草稿可编辑                 | agreements                                                         |
| /merchant/agreements/:id/publish            | POST   | 路径id                                             | id、version、published\_at | 发布并下线同类型其他版本(唯一生效)     | agreements                                                         |
| /merchant/service-reviews                   | GET    | status、score、keyword、page                        | list、total               | 评价列表填充                 | service\_reviews、service\_staffs、users、orders                      |
| /merchant/service-reviews/:id/hide          | POST   | 路径id                                             | id                       | 隐藏+异步重算质量分             | service\_reviews、service\_staffs                                   |
| /merchant/profit-sharing/receivers          | GET    | 无                                                | 接收方列表                    | 按商家查分账接收方              | profit\_sharing\_receivers                                         |
| /merchant/profit-sharing/receivers          | POST   | ReceiverRequest(type\*、name\*、account\*)         | receiver                 | 先微信建关系失败不阻断，再入库        | profit\_sharing\_receivers、merchants                               |
| /merchant/profit-sharing/receivers/:id      | PUT    | 路径id+字段                                          | receiver                 | 更新，变更账号重置wechat\_bound | profit\_sharing\_receivers                                         |
| /merchant/profit-sharing/receivers/:id      | DELETE | 路径id                                             | message                  | 微信解绑尽力+删除              | profit\_sharing\_receivers、merchants                               |
| /merchant/profit-sharing/receivers/:id/sync | POST   | 路径id                                             | receiver                 | 重试建立微信分账关系             | profit\_sharing\_receivers、merchants                               |
| /merchant/profit-sharing/config             | GET    | 无                                                | 开关、子商户号、max\_ratio       | 查询最大分账比例               | merchants                                                          |
| /merchant/profit-sharing/config             | PUT    | profit\_sharing\_enabled                         | enabled                  | 更新自动分账开关               | merchants                                                          |
| /merchant/profit-sharing/records            | GET    | status、order\_no、日期、page                         | list、pagination          | 分账记录分页                 | profit\_sharing\_records、orders                                    |
| /merchant/profit-sharing/records/:id        | GET    | 路径id                                             | 记录+Receivers             | 详情                     | profit\_sharing\_records、orders、profit\_sharing\_record\_receivers |
| /merchant/profit-sharing/records/:id/retry  | POST   | 路径id                                             | 记录                       | 对失败记录调微信分账并回写状态        | profit\_sharing\_records、profit\_sharing\_record\_receivers、orders |

## 8. RBAC 系统管理(商家端)

| 接口路径                           | 方法     | 入参                               | 出参                 | 核心业务逻辑简述              | 依赖的外部表                                             |
| ------------------------------ | ------ | -------------------------------- | ------------------ | --------------------- | -------------------------------------------------- |
| /merchant/rbac/permissions     | GET    | 无                                | menus树、permissions | owner返回全部，员工按角色推导权限过滤 | sys\_menus、merchant\_staffs、sys\_role\_menus       |
| /merchant/rbac/menus           | GET    | 无                                | 完整菜单树              | 菜单管理回显                | sys\_menus                                         |
| /merchant/rbac/menus           | POST   | MenuUpsertRequest                | SysMenu            | 校验父级后创建，清缓存           | sys\_menus                                         |
| /merchant/rbac/menus/:id       | PUT    | 路径id+MenuUpsertRequest           | SysMenu            | 非空字段局部更新，清缓存          | sys\_menus                                         |
| /merchant/rbac/menus/:id       | DELETE | 路径id                             | message            | 有子/被引用拒绝，删后清缓存        | sys\_menus、sys\_role\_menus                        |
| /merchant/rbac/roles           | GET    | keyword、status、page              | list(含menu\_count) | 角色分页+菜单数              | sys\_roles、sys\_role\_menus                        |
| /merchant/rbac/roles/all       | GET    | 无                                | SysRole\[]         | 启用角色供下拉               | sys\_roles                                         |
| /merchant/rbac/roles           | POST   | RoleUpsertRequest(name\*、code\*) | SysRole            | code唯一校验，建角色          | sys\_roles                                         |
| /merchant/rbac/roles/:id       | PUT    | 路径id+RoleUpsertRequest           | SysRole            | 更新+code唯一校验           | sys\_roles                                         |
| /merchant/rbac/roles/:id       | DELETE | 路径id                             | message            | 被绑定拒绝，事务删关联再删         | sys\_roles、sys\_role\_menus、merchant\_staff\_roles |
| /merchant/rbac/roles/:id/menus | GET    | 路径id                             | menu\_ids          | 角色已绑菜单ID              | sys\_roles、sys\_role\_menus                        |
| /merchant/rbac/roles/:id/menus | PUT    | 路径id+menu\_ids                   | message            | 先删后覆盖写(空则清空)          | sys\_roles、sys\_role\_menus                        |
| /merchant/rbac/departments     | GET    | 无                                | 部门树                | 机构树                   | sys\_departments                                   |
| /merchant/rbac/departments     | POST   | DepartmentUpsertRequest          | SysDepartment      | 建部门                   | sys\_departments                                   |
| /merchant/rbac/departments/:id | PUT    | 路径id+请求                          | 部门                 | 局部更新                  | sys\_departments                                   |
| /merchant/rbac/departments/:id | DELETE | 路径id                             | message            | 有子/有员工拒绝删除            | sys\_departments、merchant\_staffs                  |

## 9. 服务人员接单小程序

| 接口路径                                                                 | 方法   | 入参                                                       | 出参                                           | 核心业务逻辑简述                      | 依赖的外部表                                                               |
| -------------------------------------------------------------------- | ---- | -------------------------------------------------------- | -------------------------------------------- | ----------------------------- | -------------------------------------------------------------------- |
| /service-staff/register                                              | POST | RegisterRequest(username\*、password\*、qualifications\[]) | id、status                                    | 唯一校验+bcrypt，待审核注册+审核留痕        | service\_staffs、service\_staff\_audit\_records                       |
| /service-staff/login                                                 | POST | username、password                                        | token、staff                                  | 校验密码/状态，签发JWT                 | service\_staffs                                                      |
| /service-staff/wechat-login                                          | POST | code                                                     | token、staff                                  | jscode2session换openid查人员      | service\_staffs                                                      |
| /service-staff/profile                                               | GET  | 无                                                        | ServiceStaff                                 | 当前人员                          | service\_staffs                                                      |
| /service-staff/profile                                               | PUT  | RequestProfileChangeRequest                              | audit\_id                                    | 写变更审核(审计留痕)，暂不改正式字段           | service\_staff\_audit\_records、service\_staffs                       |
| /service-staff/audits                                                | GET  | 无                                                        | list、total                                   | 我的审核记录                        | service\_staff\_audit\_records                                       |
| /service-staff/todo                                                  | GET  | 无                                                        | list、stats                                   | 待出发+服务中工单+今日统计                | orders、order\_items、users、health\_records                            |
| /service-staff/orders/pending                                        | GET  | category、page、sort                                       | list、total                                   | 待接单池按预约时间排序+区域过滤              | orders、order\_items、users、health\_records                            |
| /service-staff/orders/accepted                                       | GET  | biz\_status、page                                         | list                                         | 已接单列表                         | orders、order\_items、users、health\_records                            |
| /service-staff/orders/:id                                            | GET  | 路径id                                                     | Order(三卡片)                                   | 服务对象/健康预警/下单人画像，归属校验          | orders、order\_items、users、health\_records                            |
| /service-staff/orders/:id/accept                                     | POST | 路径id                                                     | id、assigned\_staff\_id                       | 原子抢占接单，异步订阅消息通知用户(幂等)         | orders、order\_notify\_logs、users                                     |
| /service-staff/orders/:id/give-up                                    | POST | 路径id                                                     | id                                           | 仅待出发可放弃，原子退回待接池               | orders                                                               |
| /service-staff/orders/:id/check-in                                   | POST | 路径id+lat、lng                                             | id、actual\_started\_at                       | 签到置biz\_status=3，建/补服务记录      | orders、service\_records                                              |
| /service-staff/orders/:id/check-out                                  | POST | 路径id+remark、images、audio\_url                            | id、actual\_ended\_at                         | 签退置biz\_status=5，租赁写归还时间      | orders、service\_records                                              |
| /service-staff/orders/:id/location                                   | POST | 路径id+lat、lng\*                                           | reported\_at                                 | 仅服务中级，约60s/次定位轨迹              | orders、service\_location\_tracks                                     |
| /service-staff/orders/:id/audio                                      | POST | 路径id+audio\_url\*                                        | order\_id、audio\_url                         | 录音写入服务记录(已存在拒绝)               | orders、service\_records                                              |
| /service-staff/sos                                                   | POST | order\_id、lat\*、lng\*、address                            | alert\_event\_id、sos\_marked                 | 一键SOS+标记异常+写预警事件              | service\_alert\_events、orders、service\_records                       |
| /service-staff/my-region                                             | GET  | 无                                                        | service\_region、region\_list、pending\_fields | 服务区域+待变更快照                    | service\_staffs                                                      |
| /service-staff/my-quality-score                                      | GET  | 无                                                        | quality\_score、avg各维度、recent\_reviews        | 评价统计+近10条                     | service\_reviews、service\_staffs                                     |
| /service-staff/statistics                                            | GET  | 无                                                        | today/total接单与金额                             | 工时金额统计                        | orders                                                               |
| /service-staff/agreements                                            | GET  | type(默认3)                                                | agreement、consented                          | 拉取协议+本人同意态                    | agreements、agreement\_consents                                       |
| /service-staff/agreements/:id/consent                                | POST | 路径id                                                     | agreement\_id、version                        | 幂等留痕(user\_type=2)            | agreements、agreement\_consents                                       |
| /service-staff/assessment-forms                                      | GET  | 无                                                        | 量表数组                                         | 启用量表                          | health\_assessment\_forms                                            |
| /service-staff/residents/:user\_id/health-record                     | GET  | 路径user\_id                                               | HealthRecord                                 | 校验服务关系取最近档案                   | orders、health\_records                                               |
| /service-staff/residents/:user\_id/health-records                    | GET  | 路径user\_id                                               | HealthRecord\[]                              | 客户全部档案                        | orders、health\_records                                               |
| /service-staff/residents/:user\_id/health-records/:record\_id/values | PUT  | 路径user\_id/record\_id+请求                                 | HealthRecord                                 | 校验服务关系，仅白名单更新健康数值，禁改基础信息      | orders、health\_records                                               |
| /service-staff/residents/:user\_id/assessments                       | GET  | 路径user\_id+分页                                            | list、total                                   | 客户评估记录                        | orders、health\_assessments、health\_assessment\_forms                 |
| /service-staff/residents/:user\_id/assessments                       | POST | 路径user\_id+CreateAssessmentRequest                       | HealthAssessment                             | 校验服务关系、算分、assessor\_type=服务人员 | orders、health\_assessment\_forms、health\_records、health\_assessments |
| /service-staff/residents/:user\_id/fitting-recommendations           | GET  | 路径user\_id+分页                                            | list、total                                   | 客户适配建议                        | orders、fitting\_recommendations                                      |
| /service-staff/residents/:user\_id/fitting-recommendations           | POST | 路径user\_id+请求                                            | VO                                           | 校验商品上架生成快照，草稿态                | orders、health\_records、products、fitting\_recommendations             |
| /service-staff/health-education                                      | GET  | 无                                                        | 文章数组                                         | 启用文章                          | health\_education\_articles                                          |

## 10. 健康服务-商家端

| 接口路径                                  | 方法     | 入参                                  | 出参                 | 核心业务逻辑简述      | 依赖的外部表                                                    |
| ------------------------------------- | ------ | ----------------------------------- | ------------------ | ------------- | --------------------------------------------------------- |
| /merchant/health-records              | GET    | keyword、assessment\_level、page      | list、total         | 档案列表Preload用户 | health\_records、users                                     |
| /merchant/health-records/:id          | GET    | 路径id                                | record、assessments | 档案+评估记录       | health\_records、health\_assessments、users                 |
| /merchant/health-records/:id          | PUT    | 路径id+UpsertRequest                  | HealthRecord       | 管理端全字段覆盖更新    | health\_records                                           |
| /merchant/assessment-forms            | GET    | 无                                   | 量表数组               | 含草稿           | health\_assessment\_forms                                 |
| /merchant/assessment-forms            | POST   | name\*、questions、score\_rule、status | 量表                 | 新建量表版本=1      | health\_assessment\_forms                                 |
| /merchant/assessment-forms/:id        | PUT    | 路径id+请求                             | 量表                 | 全字段覆盖         | health\_assessment\_forms                                 |
| /merchant/assessment-forms/:id/status | PATCH  | 路径id+status                         | {id,status}        | 启停            | health\_assessment\_forms                                 |
| /merchant/assessment-forms/:id        | DELETE | 路径id                                | {id}               | 删除            | health\_assessment\_forms                                 |
| /merchant/health-assessments          | GET    | keyword、form\_id、page               | list、total         | 评估记录join用户    | health\_assessments、users、health\_assessment\_forms       |
| /merchant/fitting-recommendations     | GET    | keyword、status、page                 | list、total         | 建议列表join用户    | fitting\_recommendations、users                            |
| /merchant/fitting-recommendations/:id | GET    | 路径id                                | VO(含User)          | 详情            | fitting\_recommendations、users                            |
| /merchant/fitting-recommendations/:id | PUT    | 路径id+请求                             | VO                 | 部分更新+商品快照     | fitting\_recommendations、products                         |
| /merchant/fitting-recommendations/:id | DELETE | 路径id                                | {id}               | 删除            | fitting\_recommendations                                  |
| /merchant/health-education            | GET    | status、category\_id、keyword、page    | list、total         | 文章列表(全部状态)    | health\_education\_articles                               |
| /merchant/health-education            | POST   | title\*、content\*                   | 文章                 | 创建，发布须选分类     | health\_education\_articles、health\_education\_categories |
| /merchant/health-education/:id        | PUT    | 路径id+请求                             | 空                  | 更新，发布须分类      | health\_education\_articles、health\_education\_categories |
| /merchant/health-education/:id        | DELETE | 路径id                                | 空                  | 删除            | health\_education\_articles                               |
| /merchant/education-categories        | GET    | 无                                   | 分类数组               | 含停用           | health\_education\_categories                             |
| /merchant/education-categories        | POST   | name\*                              | 分类                 | 仅两级，父须一级      | health\_education\_categories                             |
| /merchant/education-categories/:id    | PUT    | 路径id+请求                             | 空                  | 父不能指向自身/二级    | health\_education\_categories                             |
| /merchant/education-categories/:id    | DELETE | 路径id                                | 空                  | 有子/关联文章拒绝     | health\_education\_categories、health\_education\_articles |

## 11. 支付回调/上传/WS/开发推送

| 接口路径                       | 方法   | 入参                     | 出参                              | 核心业务逻辑简述                                                 | 依赖的外部表                     |
| -------------------------- | ---- | ---------------------- | ------------------------------- | -------------------------------------------------------- | -------------------------- |
| /callback/wechatpay/pay    | POST | 微信平台加密报文+验签头           | {code}                          | 验签解密，仅SUCCESS更新订单已支付+transaction\_id，幂等，触发自动分账，算租赁到期     | orders、order\_items(分账相关表) |
| /callback/wechatpay/refund | POST | 微信加密报文+验签头             | {code}                          | 验签解密，按refund\_status更新refunds，无其他退款则订单置已退款               | refunds、orders             |
| /upload/token              | GET  | mime；需JWT              | token、domain、prefix、upload\_url | 按身份拼前缀，录音mimeLimit=audio/*;video/*，图片=image/\*，签发七牛token | 无                          |
| /upload/callback           | POST | 七牛回调数据                 | {message}                       | 七牛回调占位，直接成功                                              | 无                          |
| /ws/merchant               | GET  | JWT+MerchantAuth       | WS长连接                           | 商家端实时推送通道                                                | 无(内存Hub)                   |
| /dev/order-notify          | POST | order\_no              | delivered                       | 调试广播订单通知，Release禁用                                       | 无                          |
| /dev/store-visit-notify    | POST | visitor\_openid、source | delivered                       | 调试广播到店通知，Release禁用                                       | 无                          |

***

## 勘误与说明

- 本清单基于代码实际注册（`server/cmd/server/main.go`）与各 handler 文件反向生成，非 PRD。

- 个别接口 RBAC 权限码以实际中间件为准（如 `/rbac/menus/:id` PUT 用 `system:menu:create`、`/rbac/roles/:id/menus` GET 用 `system:role:view`）。

- 涉及的核心服务：`services/orderquery`(订单填充)、`services/category/tree.go`(分类树)、`services/coupon`(券)、`services/review`(质量分)、`services/notify`(订阅消息)、`services/profitsharing`(自动分账)、`services/wechatpay`(服务商支付/退款)。

