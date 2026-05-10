# 数据库迁移与回归测试文档

## 📋 目录

- [迁移文件整理说明](#迁移文件整理说明)
- [回归测试命令](#回归测试命令)
- [使用示例](#使用示例)
- [注意事项](#注意事项)

---

## 迁移文件整理说明

### 📁 文件结构

```
server/migrations/
├── 20240101000000_full_init.sql          # ⭐ 完整初始化脚本（推荐）
├── 20240101000001_init.sql               # 旧版表结构定义
├── 20240101000002_seed_data.sql          # 旧版测试数据
├── 20240101000003_announcements.sql      # 旧版公告表（已废弃）
├── 20240101000004_merchant_audit_records.sql # 旧版审核记录表（已废弃）
├── README.md                             # 旧版说明文档
└── EXECUTION_REPORT.md                   # 执行报告

server/scripts/
└── regression-test.sh                     # ⭐ 回归测试命令脚本
```

### 📝 迁移文件合并说明

#### 合并原因

原有的迁移文件分散在多个文件中，存在以下问题：
1. **维护困难**：表结构和数据分离，容易遗漏
2. **版本不一致**：不同文件的编码和注释风格不统一
3. **执行繁琐**：需要按顺序执行多个文件
4. **冗余内容**：存在重复的表定义

#### 合并方案

创建了 `20240101000000_full_init.sql` 完整初始化脚本，包含：

**第一部分：数据库设置**
- 创建数据库
- 设置字符集
- 外键检查开关

**第二部分：表结构定义（16张表）**

| 序号 | 表名 | 说明 | 关联表 |
|------|------|------|--------|
| 1 | service_providers | 服务商表 | - |
| 2 | service_provider_admins | 服务商管理员表 | service_providers |
| 3 | merchants | 商家表 | service_providers |
| 4 | merchant_applications | 商家进件申请表 | merchants |
| 5 | merchant_delivery_settings | 商家配送设置表 | merchants |
| 6 | merchant_licenses | 商家营业执照表 | merchants |
| 7 | merchant_staffs | 商家员工表 | merchants |
| 8 | announcements | 系统公告表 | service_providers |
| 9 | merchant_audit_records | 商家审核记录表 | merchants |
| 10 | categories | 商品分类表 | merchants |
| 11 | products | 商品表 | merchants, categories |
| 12 | product_specs | 商品规格表 | products |
| 13 | users | C端用户表 | - |
| 14 | orders | 订单表 | users, merchants |
| 15 | order_items | 订单商品表 | orders, merchants, products |
| 16 | refunds | 退款记录表 | orders |

**第三部分：测试数据**

| 数据类型 | 数量 | 说明 |
|---------|------|------|
| 服务商 | 1 | 寻梦服务商 |
| 服务商管理员 | 1 | admin / admin123 |
| 商家 | 1 | 美味餐厅 |
| 商家员工 | 1 | merchant / merchant123 |
| 商品分类 | 5 | 热销推荐、招牌菜等 |
| 商品 | 10 | 招牌红烧肉、水煮鱼等 |
| 商品规格 | 3 | 部分商品的份量规格 |
| C端用户 | 5 | 小明、小红等 |
| 订单 | 5 | 待支付、已支付、已完成 |
| 订单商品 | 11 | 订单中的商品明细 |

**第四部分：环境清理**
- 恢复外键检查

### 🔄 迁移文件版本对比

| 特性 | 旧版（分散） | 新版（合并） |
|------|------------|-------------|
| 文件数量 | 4个 | 1个 |
| 表结构完整性 | 需额外检查 | 完整包含 |
| 测试数据完整性 | 需额外执行 | 完整包含 |
| 执行顺序 | 需手动排序 | 自动按顺序 |
| 维护便捷性 | 困难 | 简单 |
| 编码一致性 | 不一致 | 统一 |

---

## 回归测试命令

### 🚀 功能特性

`regression-test.sh` 脚本提供以下功能：

1. **数据库初始化**：创建表结构和测试数据
2. **数据库重置**：删除所有数据后重新初始化
3. **回归测试**：自动测试核心API接口
4. **数据清理**：仅清理数据，保留表结构
5. **状态查看**：查看数据库表和数据状态

### 📋 命令语法

```bash
./scripts/regression-test.sh <命令> [选项]
```

### 🎯 可用命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `init` | 初始化数据库 | `./regression-test.sh init` |
| `reset` | 重置数据库（需确认） | `./regression-test.sh reset` |
| `test` | 执行回归测试 | `./regression-test.sh test` |
| `clean` | 清理测试数据 | `./regression-test.sh clean` |
| `status` | 查看数据库状态 | `./regression-test.sh status` |
| `help` | 显示帮助信息 | `./regression-test.sh help` |

### ⚙️ 命令选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-h, --host` | 数据库主机 | localhost |
| `-p, --port` | 数据库端口 | 3306 |
| `-u, --user` | 数据库用户 | root |
| `-P, --password` | 数据库密码 | 空 |
| `-d, --database` | 数据库名称 | fz_yyc_api |

---

## 使用示例

### 示例1：首次使用

```bash
# 1. 进入项目目录
cd server

# 2. 初始化数据库
./scripts/regression-test.sh init

# 3. 查看数据库状态
./scripts/regression-test.sh status

# 4. 执行回归测试
./scripts/regression-test.sh test
```

### 示例2：使用自定义配置

```bash
# 使用指定用户和密码初始化
./scripts/regression-test.sh init \
  -u myuser \
  -P mypassword \
  -d fz_yyc_test

# 查看自定义数据库状态
./scripts/regression-test.sh status -u myuser -P mypassword
```

### 示例3：重置数据库

```bash
# 重置前会提示确认
./scripts/regression-test.sh reset

# 输入 y 确认重置
# 系统会：
# 1. 删除现有数据库
# 2. 重建数据库
# 3. 重新初始化表结构和数据
```

### 示例4：开发调试

```bash
# 1. 清理现有数据（保留表结构）
./scripts/regression-test.sh clean

# 2. 重新初始化
./scripts/regression-test.sh init

# 3. 测试特定接口
curl -X POST http://localhost:8080/api/v1/sp/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 4. 查看状态
./scripts/regression-test.sh status
```

### 示例5：集成到 CI/CD

```yaml
# .gitlab-ci.yml 示例
stages:
  - test

regression-test:
  stage: test
  script:
    - cd server
    - chmod +x scripts/regression-test.sh
    - ./scripts/regression-test.sh reset
    - ./scripts/regression-test.sh test
```

---

## 测试账号信息

初始化后自动创建以下测试账号：

### 🔐 服务商管理员

| 字段 | 值 |
|------|-----|
| 用户名 | admin |
| 密码 | admin123 |
| 角色 | 超级管理员 |
| 服务商 | 寻梦服务商（ID: 1） |

### 🏪 商家管理员

| 字段 | 值 |
|------|-----|
| 用户名 | merchant |
| 密码 | merchant123 |
| 角色 | 商家管理员 |
| 商家 | 美味餐厅（ID: 1） |

### 👥 C端用户

| 序号 | 昵称 | 电话 | OpenID |
|------|------|------|--------|
| 1 | 小明 | 13811112222 | mock_openid_001 |
| 2 | 小红 | 13811113333 | mock_openid_002 |
| 3 | 小张 | 13811114444 | mock_openid_003 |
| 4 | 小李 | 13811115555 | mock_openid_004 |
| 5 | 小王 | 13811116666 | mock_openid_005 |

---

## 回归测试内容

脚本会自动测试以下接口：

### 1. 服务商端测试

- ✅ 服务商管理员登录
- ✅ 获取商家列表
- ✅ 获取待审核商家
- ✅ 获取数据看板

### 2. 商家端测试

- ✅ 商家管理员登录
- ✅ 获取商品列表
- ✅ 获取订单列表
- ✅ 获取商家信息

### 3. C端测试

- ✅ 获取店铺信息
- ✅ 获取商品详情
- ✅ 创建订单

---

## 注意事项

### ⚠️ 重要提醒

1. **仅用于开发测试**
   - 所有数据仅供开发和测试环境使用
   - 生产环境请重置所有密码

2. **数据库依赖**
   - 需提前安装 MySQL 5.7+ 或 MySQL 8.0+
   - 确保 MySQL 服务正在运行

3. **权限要求**
   - 执行脚本的用户需要有 CREATE DATABASE 权限
   - 执行脚本的用户需要有所选数据库的读写权限

4. **环境变量**
   - 推荐在生产环境使用环境变量或配置文件管理数据库密码
   - 避免在命令行中明文传递密码

5. **数据安全**
   - `reset` 和 `clean` 命令会删除数据，执行前请确认
   - `reset` 命令需要手动输入 `y` 确认

### 🔧 故障排查

#### 问题1：mysql 命令未找到

```bash
# Ubuntu/Debian
sudo apt-get install mysql-client

# macOS
brew install mysql

# CentOS/RHEL
sudo yum install mysql
```

#### 问题2：连接被拒绝

```bash
# 检查 MySQL 服务状态
sudo systemctl status mysql

# 检查端口是否开放
netstat -tulpn | grep 3306

# 检查用户权限
mysql -u root -p -e "SHOW GRANTS FOR 'your_user'@'localhost';"
```

#### 问题3：权限不足

```bash
# 授予用户所有权限
mysql -u root -p -e "GRANT ALL PRIVILEGES ON *.* TO 'your_user'@'localhost' IDENTIFIED BY 'your_password'; FLUSH PRIVILEGES;"
```

#### 问题4：数据库已存在

```bash
# 使用 reset 命令重新初始化
./scripts/regression-test.sh reset

# 或手动删除后重新初始化
mysql -u root -p -e "DROP DATABASE IF EXISTS fz_yyc_api;"
./scripts/regression-test.sh init
```

---

## 更新日志

### 2026-05-10
- 创建完整的数据库迁移文档
- 合并所有迁移文件为一个完整脚本
- 创建回归测试命令脚本
- 添加详细的测试账号信息

### 2024-01-01
- 初始版本创建
- 分离表结构和数据

---

## 技术支持

如有问题，请检查：

1. **数据库日志**
   ```bash
   mysql -u root -p -e "SHOW VARIABLES LIKE 'log_error';"
   ```

2. **API 服务日志**
   - 查看 `server/logs/` 目录
   - 检查控制台输出

3. **脚本调试**
   ```bash
   # 使用 bash -x 调试脚本
   bash -x ./scripts/regression-test.sh init
   ```

4. **联系支持**
   - 项目 Issues: https://github.com/yaotao6666/fz_yyc/issues
