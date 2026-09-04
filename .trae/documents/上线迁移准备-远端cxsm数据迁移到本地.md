# 上线迁移准备：远端 cxsm 业务数据 → 本地新架构库

## 一、目标与约束

### 目标

把远端生产库 `113.44.15.165:3306 / cxsm`（旧版 8 张核心表）中现有的**商家、分类、商品、商品规格、用户、收货地址、订单、订单明细**，按当前项目（本地 `fz_yyc_api`）的新表结构做字段转换与 ID 重映射后，迁移到本地库。本地全量校验通过后，再将本地结构+数据全量发布会线上。

### 硬约束（用户明确）

1. **绝不改动本地任何表结构**（不增删列、不改类型、不动 FK/索引）。迁移只能增删行。
2. **清空业务表并用远端全量重建**（not 合并）。
3. **丢弃本地分类树**，用远端扁平分类重建为一级分类。
4. **远端订单统一按零售** **`order_type=1`** **处理**；旧 status 1/3/4/6 直接 1:1 映射到本地同号状态。
5. **一并清空引用用户/商品/订单的本地 mock 表**，避免替换后产生孤儿数据，保持库纯净。
6. **保留本地菜单/系统配置/商家(id=1)/角色权限/协议/宣教/评估量表等主数据**。
7. **保留本地自营商户** **`merchants.id=1`（sub\_mch\_id=1112979963）**，商家不参与重建。
8. 远端个别数据若无法迁移（丢失/映射不上）可接受，不允许为迁移改动本地结构。

### 整体流程

```
远端 cxsm(旧结构)  ──ETL+转换──▶  本地 fz_yyc_api(新结构)  ──本地校验──▶  发布上线
     8 张核心表                       本地保留配置/菜单                        全量推送本地结构+数据到线上
```

## 二、现状分析（已探明）

### 2.1 两端连接方式

- 本地：`127.0.0.1:3306`，`fz_yyc` / `fz_yyc123`，库 `fz_yyc_api`（utf8mb4\_unicode\_ci）。MySQL 8.0 Docker 容器 `fz_yyc_mysql` 运行中。

- 远端：`113.44.15.165:3306`，`cxsm` / `3Fr3jEhemnZMfmyE`，库 `cxsm`（utf8mb4\_general\_ci）。`SHOW TABLE→` 仅 8 张：`categories, merchants, order_items, orders, product_specs, products, user_addresses, users`。

- 本地 mysql 客户端：`/opt/homebrew/opt/mysql@5.7/bin/mysql`、`mysqldump`。

- Go 工具链：`go1.21.1`（与 go.mod `go 1.21` 一致），可用于编译迁移工具。

- 本地 43 张表（新架构），远端 8 张表（旧架构）。

### 2.2 关键差异

| 表               | 远端(旧)                                                                               | 本地(新)                                                                                                                | 处理                                          |
| --------------- | ----------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| merchants       | 含 `service_provider_id`、`profit_sharing_ratio`；只有 id=1=财旭商贸，sub\_mch\_id=1112979963 | 去 merchant 化单商家；id=1=乐享辅具，sub\_mch\_id=1112979963                                                                    | **保留本地 id=1，不迁移**                           |
| categories      | 扁平：`merchant_id,name,sort,status`（31 行）                                             | 三级树：`category_type,parent_id,level,name,sort,status`（30 行 mock）                                                      | 清空本地 → 远端重建为 `level=1, parent_id=NULL` 一级分类 |
| products        | 有 `merchant_id`；无 product\_type/service\_content/sale\_type/租赁字段（307 行）             | 去 merchant；新增 `product_type,service_content,sale_type,rental_unit,rental_price,deposit,max_rental_duration`（45 mock） | 清空本地 → 远端导入，按分类推导新字段                        |
| product\_specs  | 结构一致（66 行）                                                                          | 一致                                                                                                                   | 清空本地，product\_id 重映射                        |
| orders          | 旧结构（64 行；status=1/3/4/6）                                                            | 新增 `order_type/total_deposit/押金/biz_status/service排期/pickup` 等                                                       | 清空本地 → 全部 order\_type=1，新字段默认值，status 1:1   |
| order\_items    | 有 `merchant_id`（84 行）                                                               | 新增 `sale_type/rental_*`（33 mock）                                                                                     | 清空本地，order\_id/product\_id 重映射，租赁字段默认       |
| users           | 结构一致（269 行）                                                                         | 一致（64 mock）                                                                                                          | 清空本地 → 远端全量导入，重建 id 映射                      |
| user\_addresses | 结构一致（11 行）                                                                          | 一致（41 mock）                                                                                                          | 清空本地，user\_id 重映射                           |

### 2.3 本地订单状态语义（constants）

- 本地 `orders.status`：1=待支付 2=已支付 3=已完成 4=已取消 5=退款中 6=已退款。

- 远端 status 分布 1(38)/3(3)/4(9)/6(14) → **与本地同号直接映射**，无需状态转换。

- 本地 `orders.order_type=1`=零售（实物订单主状态走 `status`），符合全部按零售处理。

## 三、采用的字段映射规则（ETL）

每张远端表按 `SELECT 列 …` 读取，写入本地时做转换。目标列名以本地为准，缺失/多余交由转换逻辑处理。

### 3.1 merchants —— 不迁移，保留本地 id=1

仅做一致性断言：本地 id=1 存在，且 `sub_mch_id=1112979963` 不变（hard constraint）。

### 3.2 categories（清空本地重建，31 行）

本地插入列：`name, category_type, parent_id, level, sort, status, created_at, updated_at`。

- 每条远端分类 → 新一级分类 `level=1, parent_id=NULL`。

- `category_type`：`租赁类` / `服务类` → `2`（服务/租赁分类）；其余 → `1`（商品分类）。

- `name/sort/status/created_at/updated_at` 直接搬。

- 建立映射 `mt_cat[远端id]=本地新id`。

### 3.3 products（清空本地重建，307 行）

本地插入列：`category_id,name,description,images,price,original_price,stock,unit,product_type,service_content,sale_type,rental_unit,rental_price,deposit,max_rental_duration,sales,sort,status,deleted_at,created_at,updated_at`。

- `category_id` = `mt_cat[远端category_id]`（无对应→NULL）。

- 按所属分类推导（结合 3.2 的 category\_type 判定）：

  - 属 `租赁类(17)`：`product_type=2, sale_type=2, category_id 分配到租赁分类`；`rental_unit=1, rental_price=price, deposit=10.00, max_rental_duration=0`。

  - 属 `服务类(18)`：`product_type=3, sale_type=1`；`service_content` 置为空 JSON 结构 `{"cycle":"","target_audience":"","services":[],"remark":""}`（保证 C 端解析不崩）。

  - 其余：`product_type=1, sale_type=1`。

  - 租赁/服务以外字段默认：`service_content=NULL,num`（零售）；`rental_unit=0,rental_price=0.00,deposit=0.00,max_rental_duration=0`。

- 其余字段直接搬。

- 建立映射 `mt_prod[远端id]=本地新id`。

### 3.4 product\_specs（清空本地重导入，66 行）

`product_id=mt_prod[远端product_id]、name、options、created_at、updated_at`。

> 校验：若 `mt_prod[远端product_id]` 缺失（商品未导入），跳过该条并在日志记录（可接受丢失）。

### 3.5 users（清空本地重导入，269 行）

本地/users 结构两边一致，逐列搬 `openid,union_id,nickname,avatar,phone,status,created_at,updated_at,first_visit_at,last_visit_at,visit_count,has_ordered,total_orders,total_spent,has_paid,first_paid_at`。
建立映射 `mt_user[远端id]=本地新id`。

### 3.6 user\_addresses（清空本地重导入，11 行）

`user_id=mt_user[远端user_id]、name,phone,province,city,district,address,lat,lng,is_default,created_at,updated_at`。user\_id 缺失则跳过并记录。

### 3.7 orders（清空本地重导入，64 行）

本地插入：`order_no,user_id,order_type,total_amount,delivery_fee,discount_amount,pay_amount,total_deposit,deposit_status,deposit_refund_amount,deposit_deduct_amount,deposit_refunded_at,rental_returned_at,rental_return_remark,rental_end_at,parent_order_id,renew_flag,delivery_type,delivery_distance,delivery_address,contact_name,contact_phone,status,biz_status,remark,verify_code,transaction_id,paid_at,pay_notify_payload,profit_sharing_status,profit_sharing_amount,profit_sharing_order_no,profit_sharing_at,profit_sharing_error,completed_at,completed_by_name,cancelled_at,refunded_at,scheduled_at,assigned_staff_id,record_id,delivery_district,actual_started_at,actual_ended_at,created_at,updated_at,pickup_point_id,pickup_point_name,pickup_point_address,pickup_point_lat,pickup_point_lng,assigned_at`。

- `user_id=mt_user[远端user_id]`（缺失跳过并记录）。

- `order_type=1`（全部零售）；`status` 直接搬远端值(1/3/4/6)；`biz_status=0`。

- 押金/租赁/服务/排期新字段默认：`total_deposit=0.00,deposit_status=0,deposit_refund_amount=0.00,deposit_deduct_amount=0.00,deposit_refunded_at=NULL,rental_returned_at=NULL,rental_return_remark=NULL,rental_end_at=NULL,parent_order_id=NULL,renew_flag=0,scheduled_at=NULL,assigned_staff_id=NULL,record_id=NULL,delivery_district=NULL,actual_started_at=NULL,actual_ended_at=NULL,assigned_at=NULL`。

- `delivery_type,delivery_distance,delivery_address,contact_name,contact_phone,remark,verify_code,transaction_id,paid_at,pay_notify_payload,profit_sharing_*,completed_at,completed_by_name,cancelled_at,refunded_at,delivery_type,pickup_point_*,created_at,updated_at` 直接搬。

- 建立映射 `mt_order[远端id]=本地新id`。

### 3.8 order\_items（清空本地重导入，84 行）

`order_id=mt_order[远端order_id]、product_id=mt_prod[远端product_id]、product_name,image,price,quantity,spec_info,subtotal,created_at`，新字段默认：`sale_type=1,rental_unit=0,rental_duration=0,unit_rental_price=0.00,rental_subtotal=0.00,deposit=0.00,deposit_deduct=0.00`。order\_id/product\_id 缺失则跳过并记录。

### 3.9 本地被清空的引用型 mock 表（user 确认：一并清空）

`user_behavior_events, user_visits, user_coupons, refunds, health_records, health_assessments, fitting_recommendations, agreement_consents, service_records, service_reviews, service_alert_events, service_location_tracks, profit_sharing_records, profit_sharing_record_receivers, store_home_recommends`。

### 3.10 保留（不清不迁移）

`sys_*（menus/roles/role_menus/departments）`、`system_configs`、`merchants`（id=1）、`merchant_delivery_settings`、`merchant_fees`、`merchant_rates`、`merchant_staffs`、`merchant_staff_roles`、`service_staffs`、`service_staff_roles`、`service_staff_audit_records`、`agreements`、`coupon_templates`、`mini_program_banners`、`health_education_categories`、`health_education_articles`、`health_assessment_forms`、`profit_sharing_receivers`、`printers`。

## 四、落地：迁移工具（唯一新增代码）

### 4.1 新增 `server/cmd/migrate_cxsm/main.go`

用标准库 `database/sql` + 既有依赖 `github.com/go-sql-driver/mysql`（已在 go.mod）同时连两个库，做 ETL：

- **源库（远端 cxsm）**：通过专用前缀环境变量注入，**不入 .env、不入 git**，命令行/临时环境提供：
  `CXSM_SRC_HOST/CXSM_SRC_PORT/CXSM_SRC_USER/CXSM_SRC_PASSWORD/CXSM_SRC_DB`。

- **目标库（本地）**：复用 `internal/config` 读 `.env`（`DB_*`=本地 fz\_yyc\_api）。

- 步骤（全部在目标库一个事务内，`SET FOREIGN_KEY_CHECKS=0` 开始、结束时恢复 `=1`）：

  1. **结构自检**：对 8 张目标表执行 `SHOW COLUMNS`，比对工具内置期望列；不一致立即 `FATAL` 且不写任何数据（保证“绝不动结构 / 结构不符即停”）。同时断言 `merchants.id=1 且 sub_mch_id=1112979963`。
  2. **清空**：按 3.9 + 8 张核心表，用 `DELETE FROM` 逐表清空（先子表后父表，配 FOREIGN\_KEY\_CHECKS=0 兜底）。
  3. **重建并重映射**：按 3.2→3.8 顺序插入并维护 `mt_cat/mt_prod/mt_user/mt_order` 映射。
  4. **重置自增**：对 8 张核心表 `ALTER TABLE … AUTO_INCREMENT=1`（非结构变更）。
  5. **校验与输出**：

     - 每张表写入行数 == 读取行数（**必须一致**）。

     - 孤儿检查：orders.user\_id、order\_items.order\_id/product\_id、orders.category\_id(products)、order\_items.order\_id 等均可 resolve。

     - 输出汇总（读取/写入/跳过/失败）+ 关键样例（如 orders 按 status 分布）。

- 工具带 `--dry-run`（只读远端+目标、打印要执行的清空/插入计划，不写库）用于先审阅。

- 使用 `--default-character-set=utf8mb4` 语义：Go 驱动 DSN 里 `charset=utf8mb4&parseTime=True&loc=Local`（与项目一致），避免中文乱码（遵循项目 hard constraint）。

### 4.2 运行与产物

- 编译：`cd server && go build -o dist/migrate_cxsm ./cmd/migrate_cxsm`。

- 先 `--dry-run` 查看计划，再正式执行（源库凭据从内存/临时环境传入，不落盘）。

- 迁移成功后生成一份**本地全量 SQL 快照**（供“发布线上”用）：
  `mysqldump -h127.0.0.1 -ufz_yyc -pfz_yyc123 fz_yyc_api --default-character-set=utf8mb4 > server/backups/fz_yyc_online_<yyyyMMdd_HHmmss>.sql`

- 同时生成一份**业务校验 SQL**：满足“行数与远端一致、无孤儿引用”的核对脚本。

## 五、发布上线步骤（本地校验通过后执行，本轮不实际写生产库）

1. 用户确认本地数据符合预期（跑 3 张校验 SQL + 登录后台人工抽查）。
2. 在线上目标机停服 → 备份线上旧库 → 执行 `本地 mysqldump 到线上` 导入（即把本地新结构+数据全量推上）→ 启动新版服务。

   - 说明：实际“推送到线上”属上线动作，需在本地全量确认后、且获授权后单独执行；本轮只准备迁移脚本 + 校验脚本 + 快照命令。

## 六、假设与决策（已与用户确认）

- 清空业务表 & 引用型 mock 表，用远端全量重建（A1/A2/关联表三项均选此）。

- 远端分类重建为一级扁平分类，丢弃本地三级树。

- 远端订单统一 `order_type=1`，status 1/3/4/6 直接映射。

- 保留本地表结构 & 系统配置/菜单/商家(id=1)/主数据；不动本地结构是本迁移的第一优先级。

- 远端数据个别丢失可接受，不由结构妥协。

## 七、验证清单

1. `go build` 通过，`--dry-run` 输出与预期一致。
2. 迁移执行后：`users=269, categories=31, products=307, product_specs=66, orders=64, order_items=84, user_addresses=11`（与远端完全一致）。

   > 注：products 会导入 307 条（含 3 条 deleted\_at 非空的下架软删，一并搬入）。
3. 孤儿校验全 0：`orders.user_id∉users`、`order_items.order_id∉orders`、`order_items.product_id∉products`、`products.category_id∉categories`、`user_addresses.user_id∉users` 均为空。
4. 保留表校验：`merchants.id=1` 且 sub\_mch\_id 未变；`sys_menus/sys_roles/system_configs/mini_program_banners 等行数不变`。
5. 本地后台(web-admin) + 两个小程序登录后：首页商品/分类、订单列表可正常展示。
6. 本地全量 SQL 快照成功生成（`server/backups/fz_yyc_online_*.sql`）。

## 八、涉及文件（仅新增，不改现有）

- 新增：`server/cmd/migrate_cxsm/main.go`（迁移工具）

- 可选新增：`server/backups/` 下的迁移校验脚本/README 说明（新 ETL 用法）

- 不改：现有服务/迁移/表结构

