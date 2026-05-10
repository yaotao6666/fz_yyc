# 数据库迁移与回归测试文档

## 当前迁移口径

当前仓库只保留一套数据库初始化入口：

- `server/migrations/20240101000000_full_init.sql`

该脚本负责：

1. 创建数据库
2. 创建当前代码所需的最新表结构
3. 写入最小初始化数据

当前不再使用旧拆分迁移脚本，也不再使用 `go run cmd/main.go migrate up` 这类不存在的旧命令入口。

## 初始化账号

初始化后仅保留以下账号，与 `PRD.md` 保持一致：

- 服务商管理员：`admin / admin123`
- 商家员工：`merchant / merchant123`

## 文件结构

```text
server/migrations/
├── 20240101000000_full_init.sql
├── MIGRATION_GUIDE.md
└── REGRESSION_TEST_REPORT.md

server/scripts/
├── regression-test.sh
└── regression.sh
```

## 执行方式

### 方式一：直接导入初始化 SQL

```bash
cd server
docker exec -i fz_yyc_mysql mysql -uroot -proot123456 < migrations/20240101000000_full_init.sql
```

### 方式二：使用回归脚本初始化

```bash
cd server
./scripts/regression-test.sh init
```

## 回归脚本功能

`./scripts/regression-test.sh` 支持以下命令：

| 命令 | 说明 |
|------|------|
| `init` | 初始化数据库 |
| `reset` | 重置数据库 |
| `test` | 执行回归测试 |
| `clean` | 清理测试数据 |
| `status` | 查看数据库状态 |

## 常用示例

```bash
cd server

# 初始化数据库
./scripts/regression-test.sh init

# 查看数据库状态
./scripts/regression-test.sh status

# 执行回归测试
./scripts/regression-test.sh test
```

## 注意事项

1. 当前初始化脚本适用于开发与测试环境，不应用于生产环境直接覆盖数据。
2. 如果通过 Docker 方式导入 SQL，请先确认 `fz_yyc_mysql` 容器正常运行。
3. 初始化完成后，如需验证接口，请重新启动或重建 `api` 容器。
4. 生产环境请替换默认账号密码，不要继续使用测试口令。
