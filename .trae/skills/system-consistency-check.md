# 系统一致性检测技能

## 概述

对"适老化项目助手"项目进行全链路一致性检测，覆盖 PRD 文档、数据库表结构、后端接口模型、前端页面接入四个层面，输出差异清单并生成修复脚本。

**适用场景**：
- 新项目启动前的完整检测
- 系统重构后的全面验证
- 部署上线前的质量保障
- 接口或表结构变更后的连锁检查

## 前置条件

### 必需文件
- `PRD.md` — 产品需求文档（数据模型定义在第4章）
- `server/internal/models/models.go` — 数据库模型定义
- `server/cmd/server/main.go` — 路由配置
- `server/internal/handlers/` — 所有 handler 实现
- `server/migrations/` — 迁移脚本
- `miniprogram/src/api/index.ts` — 前端 API 封装
- `miniprogram/src/pages/` — 所有前端页面
- `miniprogram/src/types/index.ts` — 前端类型定义

### 必需环境
- Docker（用于启动数据库验证）
- curl（用于接口自测）

### 环境验证
```bash
# 确认后端服务运行中
docker ps --filter "name=fz_yyc" --format "{{.Names}}: {{.Status}}"

# 确认数据库可连接
docker exec fz_yyc_mysql mysql -uroot -pfz_yyc_2024 fz_yyc_api -e "SHOW TABLES;"
```

## 执行流程

### 阶段一：并行检测（任务1/2/3 同步执行）

---

### 任务1：PRD与数据库表结构一致性检测

**目标**：以 PRD 文档需求为准，排查已连的数据库表结构是否一致。

**执行步骤**：

1. 读取 `PRD.md` 第4章"数据模型设计"，提取所有表名和字段定义
2. 读取 `server/internal/models/models.go`，提取所有 GORM 模型及其字段、标签
3. 读取 `server/migrations/` 下所有 `.sql` 文件，提取实际建表语句
4. 逐表对比，按以下维度检查：
   - 表是否存在
   - 字段是否完整
   - 字段类型是否一致（如 varchar 长度、int 精度）
   - 索引是否完整
   - 约束条件是否一致（NOT NULL、DEFAULT、UNIQUE、外键）
5. 对于 PRD 未定义但数据库中存在的表/字段，标记为"PRD未定义"，提示确认

**PRD 中定义的数据表清单**（共 21 张）：

| 表名 | 说明 |
|------|------|
| merchants | 商家 |
| merchant_delivery_settings | 商家配送设置 |
| merchant_staffs | 商家员工 |
| categories | 商品分类 |
| products | 商品 |
| product_specs | 商品规格 |
| users | 用户 |
| user_visits | 用户访问记录 |
| user_behavior_events | 用户行为事件 |
| orders | 订单 |
| order_items | 订单商品 |
| refunds | 退款 |
| activities | 活动 |
| announcements | 公告 |
| merchant_fees | 商家年费 |
| merchant_rates | 商家费率 |
| user_addresses | 用户地址 |
| coupons | 优惠券 |
| coupon_records | 优惠券记录 |
| cloud_printers | 云打印机 |
| print_logs | 打印日志 |

**差异清单输出格式**：

```markdown
## 任务1：PRD与数据库差异清单

### 缺失的表
| 表名 | PRD章节 | 说明 | 建议方案 |
|------|---------|------|----------|

### 缺失的字段
| 表名 | 字段名 | PRD要求类型 | 实际类型 | 建议方案 |
|------|--------|------------|---------|----------|

### 类型不一致
| 表名 | 字段名 | PRD要求 | 实际定义 | 建议方案 |
|------|--------|---------|---------|----------|

### PRD未定义的扩展
| 表名/字段 | 类型 | 说明 | 建议 |
|-----------|------|------|------|
```

---

### 任务2：Gin框架接口与模型一致性检测

**目标**：全面扫描接口对应的模型和数据库是否一致。

**执行步骤**：

1. 读取 `server/cmd/server/main.go`，提取所有路由注册（HTTP方法 + 路径 + Handler函数）
2. 逐个读取 `server/internal/handlers/` 下各子目录的 handler 文件：
   - `admin/` — 管理员登录、支付回调
   - `merchant/` — 商家端所有业务（商品、订单、分类、员工、分析、配送、打印等）
   - `user/` — C端用户业务（店铺、订单、地址等）
   - `upload/` — 文件上传
   - `ws/` — WebSocket
3. 对每个 handler 检查：
   - 请求结构体（binding 标签）的字段是否与 models.go 中对应模型的字段一致
   - 响应数据是否正确映射模型字段
   - 数据库查询是否使用了正确的模型关联和预加载
   - 中间件认证是否正确（merchant/user 角色区分）
4. 验证路由与 handler 的映射完整性：
   - main.go 注册的路由是否有对应 handler 实现
   - handler 中导出的函数是否都在 main.go 中注册了路由

**路由分组清单**：

| 路由组 | 前缀 | 认证 | Handler包 |
|--------|------|------|-----------|
| 认证 | /api/v1/auth | 无 | admin, merchant, user |
| 文件上传 | /api/v1/upload | JWT | upload |
| C端店铺 | /api/v1/store/:merchant_id | 无 | user |
| C端用户 | /api/v1/user | JWT | user |
| 商家管理 | /api/v1/merchant | JWT | merchant |
| WebSocket | /api/v1/ws | JWT | ws |
| 支付回调 | /api/v1/notify | 无 | admin |

**差异清单输出格式**：

```markdown
## 任务2：接口与模型差异清单

### 请求结构体与模型不一致
| 接口路径 | Handler | 字段名 | 请求定义 | 模型定义 | 建议方案 |
|----------|---------|--------|---------|---------|----------|

### 响应结构与模型不一致
| 接口路径 | Handler | 问题描述 | 建议方案 |
|----------|---------|----------|----------|

### 路由映射问题
| 类型 | 路由/Handler | 问题描述 | 建议方案 |
|------|-------------|----------|----------|

### 中间件认证问题
| 接口路径 | 当前认证 | 期望认证 | 建议方案 |
|----------|---------|---------|----------|
```

---

### 任务3：小程序页面接口接入一致性检测

**目标**：全面扫描两种身份（商家、用户）的所有页面，验证是否正确接入接口。

**执行步骤**：

1. 读取 `miniprogram/src/api/index.ts`，提取所有 API 函数及其调用的接口路径
2. 扫描 `miniprogram/src/pages/` 下所有 `.vue` 文件，提取每个页面中调用的 API 函数
3. 对比前端 API 路径与后端实际路由
4. 检查参数命名和类型是否与后端 handler 的请求结构体一致
5. 验证不同身份页面的认证方式

**页面扫描范围**：

| 身份 | 页面目录 | 页面数 | 关键页面 |
|------|----------|--------|----------|
| 商家端 | pages/merchant/ | 11 | 工作台、商品列表/编辑、订单列表/详情、分类管理、数据分析、设置、员工、配送、通知 |
| 用户端 | pages/store/ | 7 | 店铺首页、商品详情、购物车、确认订单、我的订单、声音测试 |
| 认证页 | pages/auth/ | 1 | 商家登录 |

**检测维度**：

| 检测项 | 说明 |
|--------|------|
| 接口路径 | 前端调用路径是否与后端路由完全匹配 |
| HTTP方法 | GET/POST/PUT/DELETE 是否一致 |
| 请求参数 | 参数名、类型、是否必填是否匹配 |
| 响应解析 | 前端是否正确解析后端返回的数据结构 |
| 认证方式 | 商家端/用户端是否使用正确的 token |
| 错误处理 | 前端是否处理了后端可能返回的错误码 |

**差异清单输出格式**：

```markdown
## 任务3：前端接口差异清单

### 接口路径不匹配
| 页面 | 文件 | 前端路径 | 后端路径 | 建议方案 |
|------|------|---------|---------|----------|

### 请求参数不一致
| 页面 | 文件 | 接口 | 参数名 | 前端类型 | 后端类型 | 建议方案 |
|------|------|------|--------|---------|---------|----------|

### 响应解析问题
| 页面 | 文件 | 接口 | 问题描述 | 建议方案 |
|------|------|------|----------|----------|

### 认证方式问题
| 页面 | 文件 | 当前方式 | 正确方式 | 建议方案 |
|------|------|---------|---------|----------|

### 接口缺失（页面需要但未调用）
| 页面 | 文件 | 需要的接口 | 建议方案 |
|------|------|-----------|----------|

### 多余调用（调用了不存在的接口）
| 页面 | 文件 | 调用的接口 | 建议方案 |
|------|------|-----------|----------|
```

---

### 阶段二：自测与脚本生成（依赖阶段一完成）

---

### 任务4：自测验证与迁移脚本生成

**4.1 自测验证**

在差异清单确认并修复后，执行以下自测：

```bash
# 1. 重启后端服务
cd server && docker compose restart

# 2. 等待服务就绪
sleep 5 && curl -s http://localhost:8080/health

# 3. 测试认证接口
curl -s -X POST http://localhost:8080/api/v1/auth/merchant/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'

# 4. 测试核心CRUD接口（使用获取的token）
TOKEN="获取的JWT token"

# 商家信息
curl -s http://localhost:8080/api/v1/merchant/profile -H "Authorization: Bearer $TOKEN"

# 商品列表
curl -s http://localhost:8080/api/v1/merchant/products -H "Authorization: Bearer $TOKEN"

# 分类列表
curl -s http://localhost:8080/api/v1/merchant/categories -H "Authorization: Bearer $TOKEN"

# 订单列表
curl -s http://localhost:8080/api/v1/merchant/orders -H "Authorization: Bearer $TOKEN"
```

**4.2 迁移脚本生成规范**

当检测发现数据库结构需要变更时，在 `server/migrations/` 下生成新的迁移文件：

**命名格式**：`YYYYMMDDHHMMSS_description.sql`

**文件模板**：
```sql
-- 迁移描述
-- 创建时间: YYYY-MM-DD HH:MM:SS
-- 关联检测: system-consistency-check

-- ===== UP =====
BEGIN;

-- 表结构变更
ALTER TABLE `table_name` ADD COLUMN `new_column` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '说明';

-- 新建表
CREATE TABLE IF NOT EXISTS `new_table` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 初始数据
INSERT INTO `table_name` (`column1`, `column2`) VALUES ('value1', 'value2');

COMMIT;

-- ===== DOWN =====
-- BEGIN;
-- ALTER TABLE `table_name` DROP COLUMN `new_column`;
-- DROP TABLE IF EXISTS `new_table`;
-- COMMIT;
```

**4.3 初始化数据清单**

迁移脚本中应包含以下初始化数据：

| 数据类型 | 说明 | 示例 |
|----------|------|------|
| 系统配置 | 系统基础配置 | 商户号、API密钥占位 |
| 基础分类 | 通用商品分类模板 | 可选，按需创建 |
| 系统参数 | 全局配置项 | 配送费默认值、通知开关等 |

---

### 阶段三：汇总输出

将三个检测任务的差异清单合并，生成完整报告：

```markdown
# 系统一致性检测报告

## 检测概要
- 检测时间: YYYY-MM-DD HH:MM:SS
- PRD与数据库差异: X 项
- 接口与模型差异: X 项
- 前端接口差异: X 项

## 任务1：PRD与数据库差异清单
[合并任务1输出]

## 任务2：接口与模型差异清单
[合并任务2输出]

## 任务3：前端接口差异清单
[合并任务3输出]

## 修复建议优先级
| 优先级 | 类型 | 数量 | 说明 |
|--------|------|------|------|
| P0-紧急 | 数据结构缺失 | X | 影响核心功能 |
| P1-重要 | 接口不一致 | X | 影响数据正确性 |
| P2-一般 | 前端差异 | X | 影响用户体验 |
| P3-低 | PRD未定义扩展 | X | 需确认是否保留 |

## 生成的迁移脚本
- [ ] `server/migrations/YYYYMMDDHHMMSS_xxx.sql` — 说明
```

## 相关文件路径

```
fz_yyc/
├── PRD.md                                    # 产品需求文档
├── server/
│   ├── cmd/server/main.go                    # 路由配置
│   ├── internal/
│   │   ├── models/models.go                  # 数据库模型（19个模型）
│   │   ├── handlers/
│   │   │   ├── admin/                        # 管理员handler
│   │   │   │   ├── handler.go                # 登录
│   │   │   │   ├── announcement.go           # 公告（旧接口）
│   │   │   │   ├── invite.go                 # 邀请（旧接口）
│   │   │   │   └── payment.go                # 支付回调
│   │   │   ├── merchant/                     # 商家handler
│   │   │   │   ├── handler.go                # 登录、商家信息
│   │   │   │   ├── product.go                # 商品CRUD
│   │   │   │   ├── order.go                  # 订单管理
│   │   │   │   ├── analytics.go              # 数据分析
│   │   │   │   ├── staff.go                  # 员工管理
│   │   │   │   └── invite.go                 # 邀请
│   │   │   ├── user/                         # C端用户handler
│   │   │   │   └── handler.go                # 店铺、订单、地址
│   │   │   ├── upload/upload.go              # 文件上传
│   │   │   └── ws/                           # WebSocket
│   │   │       ├── hub.go                    # 连接管理
│   │   │       ├── merchant.go               # 商家WS
│   │   │       └── dev.go                    # 开发调试
│   │   └── middleware/auth.go                # JWT认证中间件
│   ├── migrations/                           # 迁移脚本
│   │   ├── 20240101000000_full_init.sql      # 全量初始化
│   │   └── 20260511120000_mock_seed.sql      # 模拟数据
│   └── docker-compose.yml                    # Docker配置
└── miniprogram/
    └── src/
        ├── api/index.ts                      # API封装（60+函数）
        ├── types/index.ts                    # 类型定义
        ├── utils/request.ts                  # 请求工具
        ├── pages/
        │   ├── auth/login.vue                # 商家登录
        │   │   ├── merchant/                     # 商家端（11页面）
        │   │   ├── home.vue                  # 工作台
        │   │   ├── categories.vue            # 分类管理
        │   │   ├── settings.vue              # 设置
        │   │   ├── delivery-settings.vue     # 配送设置
        │   │   ├── notification-settings.vue # 通知设置
        │   │   ├── staff.vue                 # 员工管理
        │   │   ├── products/list.vue         # 商品列表
        │   │   ├── products/edit.vue         # 商品编辑
        │   │   ├── orders/list.vue           # 订单列表
        │   │   ├── orders/detail.vue         # 订单详情
        │   │   └── analytics/index.vue       # 数据分析
        │   └── store/                        # 用户端（7页面）
        │       ├── home.vue                  # 店铺首页
        │       ├── product.vue               # 商品详情
        │       ├── cart.vue                  # 购物车
        │       ├── confirm.vue               # 确认订单
        │       ├── my-orders.vue             # 我的订单
        │       ├── order-sound-test.vue      # 声音测试
        │       └── test-entry.vue            # 测试入口
        └── stores/                           # Pinia状态管理
            ├── auth.ts                       # 认证状态
            └── merchant.ts                   # 商家状态
```

## 注意事项

1. **并发安全**：修改表结构时注意对生产数据的影响，避免锁表时间过长
2. **向后兼容**：接口变更需要考虑版本兼容，新增字段使用默认值，删除字段需评估影响
3. **事务完整**：迁移脚本必须使用事务（BEGIN/COMMIT），确保失败可回滚
4. **数据备份**：重大变更前执行 `docker exec fz_yyc_mysql mysqldump -uroot -pfz_yyc_2024 fz_yyc_api > backup.sql`
5. **文档同步**：修改产品功能或后端接口或数据表后，必须同步修改 PRD 文档
6. **以PRD为准**：当实现与PRD不一致时，以PRD文档需求为准，如不满足需给出建议方案
7. **两端隔离**：商家端与用户端的认证和权限完全隔离，检测时需分别验证
