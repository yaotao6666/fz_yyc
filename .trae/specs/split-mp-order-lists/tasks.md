# Tasks

- [x] Task 1: C端订单拆为两个独立页面
  - 抽 `my-orders.vue` 共用列表逻辑/样式（卡片渲染区分实物/服务）。

  - 新建「实物订单」`my-orders-goods`（category=1）与「服务订单」`my-orders-service`（category=2）两页，在 `pages.json` 注册。

  - 服务卡片：展示服务对象与「服务地址」（收货地址服务场景改名），隐藏配送费行/配送距离；实物卡片：保留收货地址与配送费。

  - 更新入口跳转（首页我的订单、订单详情返回、支付成功/取消重定向）到对应分类页。

  - 验证：C 端两页能分别拉取并展示分类订单。

- [x] Task 2: 服务端工单按实物/服务分类
  - `workorder/index.vue` 顶部新增「实物/服务」分类切换。

  - `staffWorkorderApi.getPendingOrders / getAcceptedOrders` 增加 `category` 参数。

  - 后端 `PendingOrders / AcceptedOrders`（workorder.go）支持 `category` 过滤（经 order\_type 范围判定），缺省不过滤。

  - 验证：服务端切换分类后列表按实物/服务正确过滤。

- [x] Task 3: 下单区分——服务订单不计配送费（C端 confirm + 后端）
  - `confirm.vue`：服务订单（`hasServiceItems`=true）隐藏配送费行、不加载/不校验配送档位与商家配送范围，仍校验「服务地址」文案的地址与健康档案；实物订单保持原配送逻辑。

  - 后端 `CreateOrder`：服务订单分支强制 `deliveryFee=0`、忽略配送距离、不受 `takeout_enabled`/配送规则约束；实物订单保持原逻辑。

  - 验证：服务订单金额无配送费且可正常下单；实物订单行为不变。

- [x] Task 4: 禁止实物与服务混单支付
  - 后端 `CreateOrder`：遍历 `req.Items` 判断商品分类，若同时含实物与服务商品则拒绝下单并返回参数错误。

  - `confirm.vue`：下单前置校验，购物车/本次下单同时含实物与服务商品时拦截并提示先拆分下单。

  - 验证：混合下单被后端拦截，前端有明确提示。

# Task Dependencies

- Task 1 / Task 2 / Task 3 / Task 4 相互独立，可并行（改动面不同：C端列表、服务端工单、下单接口、混单校验）。

- 验证依赖各任务自身实现完成。

