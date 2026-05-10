# 寻梦私域管家 - 回归测试报告

## 测试时间

- 测试日期: 2026-05-11
- 测试环境: Docker (MySQL 8.0, Redis 7, API Server)
- API 服务: http://localhost:8080
- 初始化方式: `server/migrations/20240101000000_full_init.sql`

## 测试执行总结

### 1. 数据库初始化测试

- 测试结果: 通过
- 结论:
  - 单文件初始化 SQL 可成功导入
  - `api` 重建后可正常启动并通过健康检查
  - `merchant_staffs` 相关字段未再出现缺失报错

### 2. 数据验证测试

- 测试结果: 通过
- 实际结果:
  - 数据表数量: 28
  - 服务商管理员账号: 1
  - 商家员工账号: 1
  - 商品数量: 0
  - C 端用户数量: 0
  - 订单数量: 0
- 说明:
  - 当前初始化脚本只保留最小登录数据
  - 商品、订单、C 端用户等业务数据由联调或测试过程生成

### 3. API 接口测试

#### 3.1 服务商管理员登录

- 路径: `POST /api/v1/auth/admin/login`
- 结果: 通过
- 响应: 返回 `token` 与 `admin` 信息

#### 3.2 商家员工登录

- 路径: `POST /api/v1/auth/merchant/login`
- 结果: 通过
- 响应: 返回 `token`、`merchant_id` 与 `staff` 信息

#### 3.3 旧登录路径验证

- 路径: `POST /api/v1/sp/auth/login`
- 结果: 未作为初始登录入口使用
- 现象: 直接请求返回 `未提供认证令牌`
- 结论: 回归脚本与文档需统一使用 `/api/v1/auth/admin/login`

## 当前初始化账号

### 服务商管理员

| 字段 | 值 |
|------|-----|
| 用户名 | admin |
| 密码 | admin123 |
| 登录接口 | POST /api/v1/auth/admin/login |

### 商家员工

| 字段 | 值 |
|------|-----|
| 用户名 | merchant |
| 密码 | merchant123 |
| 登录接口 | POST /api/v1/auth/merchant/login |

## 关键验证命令

```bash
# 导入单文件初始化 SQL
cd server
docker exec -i fz_yyc_mysql mysql -uroot -proot123456 < migrations/20240101000000_full_init.sql

# 重建并启动 api
docker compose up --build -d api

# 验证服务商管理员登录
curl -X POST http://localhost:8080/api/v1/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 验证商家员工登录
curl -X POST http://localhost:8080/api/v1/auth/merchant/login \
  -H "Content-Type: application/json" \
  -d '{"username":"merchant","password":"merchant123"}'
```

## 结论

- 单文件迁移重建方案可用，已可替代旧拆分迁移口径。
- 当前仓库应以 `20240101000000_full_init.sql` 作为唯一初始化入口。
- 账号口径已收敛为 `admin / admin123` 与 `merchant / merchant123`。
- 回归脚本、迁移文档和 PRD 需要统一到当前真实登录接口，避免继续误报。
