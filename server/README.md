# 后端服务启动和停止命令

## 环境要求

- Go 1.21 或更高版本（仅开发环境需要）
- Docker 和 Docker Compose
- MySQL 8.0（Docker 容器）
- Redis 7（Docker 容器）

## 目录结构

- 后端代码：`server/`
- 数据库迁移：`server/migrations/`
- Dockerfile：`server/Dockerfile`
- Docker Compose 配置：`server/docker-compose.yml`
- 环境变量示例：`server/.env.example`

## 🐳 Docker Compose 统一管理（推荐方式）

现在可以使用 Docker Compose 统一管理所有服务：MySQL、Redis 和后端 API。

### 启动所有服务

```bash
# 进入后端目录
cd server

# 启动所有服务（MySQL、Redis、API）
docker compose up -d

# 查看构建日志（可选）
docker compose up --build -d
```

### 查看服务状态

```bash
# 查看所有容器状态
docker ps -a --filter "name=fz_yyc"

# 查看容器状态（详细信息）
docker compose ps
```

### 查看日志

```bash
# 查看所有服务日志
docker compose logs -f

# 只查看 API 服务日志
docker compose logs -f api

# 只查看 MySQL 日志
docker compose logs -f mysql

# 只查看 Redis 日志
docker compose logs -f redis

# 查看最近 50 行 API 日志
docker compose logs --tail 50 api
```

### 重启服务

```bash
# 重启所有服务
docker compose restart

# 只重启 API 服务
docker compose restart api

# 重启并重新构建
docker compose up --build -d
```

### 停止所有服务

```bash
# 停止所有容器（保留数据卷）
docker compose down

# 停止并删除容器和网络
docker compose down

# 停止并删除所有资源（包括数据卷 - 会删除所有数据！）
docker compose down -v
```

## 停止命令

### 停止单个服务

```bash
# 停止 API 服务
docker compose stop api

# 停止 MySQL
docker compose stop mysql

# 停止 Redis
docker compose stop redis
```

### 清理 Docker 资源

```bash
# 删除已停止的容器
docker compose rm

# 删除所有容器和网络（保留数据卷）
docker compose down

# 完全清理（包括数据卷）
docker compose down -v
```

## 重启服务

### 重启所有服务

```bash
# 重启所有容器
docker compose restart

# 重新构建并启动
docker compose up --build -d
```

### 重启单个服务

```bash
# 重启 API 服务
docker compose restart api

# 重启 MySQL
docker compose restart mysql

# 重启 Redis
docker compose restart redis
```

## 开发环境运行

如果需要在开发环境直接运行（不使用 Docker）：

### 方式一：直接运行

```bash
# 进入后端目录
cd server

# 运行主程序
go run cmd/server/main.go
```

### 方式二：编译后运行

```bash
# 进入后端目录
cd server

# 编译
go build -o fz_yyc_api cmd/server/main.go

# 运行编译后的程序
./fz_yyc_api
```

服务默认运行在 `http://localhost:8080`

## 配置说明

### Docker 环境配置

Docker Compose 会自动配置以下环境变量：

```yaml
environment:
  # 数据库配置
  DB_HOST: mysql           # Docker 内部网络主机名
  DB_PORT: 3306
  DB_USER: fz_yyc
  DB_PASSWORD: fz_yyc123
  DB_NAME: fz_yyc_api
  DB_CHARSET: utf8mb4

  # Redis 配置
  REDIS_HOST: redis        # Docker 内部网络主机名
  REDIS_PORT: 6379
  REDIS_PASSWORD: redis123
  REDIS_DB: 0

  # 应用配置
  JWT_SECRET: default-secret-change-in-production
  JWT_EXPIRE: 720
  APP_HOST: 0.0.0.0
  APP_PORT: 8080
  APP_ENV: development
  APP_DEBUG: "true"
```

## 生产环境发布配置（推荐）

### 关键原则

1. 不在仓库文件中写入任何真实密钥（JWT、微信、七牛、Redis 密码等），统一通过 `server/.env.production` 注入。
2. 生产环境必须设置 `APP_DEBUG=false`，让 Gin 进入 Release 模式（同时禁用 `/api/v1/dev/*` 联调接口）。
3. MySQL / Redis 不对公网暴露端口，只允许容器网络或内网访问。
4. 对外提供 HTTPS（建议用 Nginx/Caddy 反向代理到 `127.0.0.1:8080`），WebSocket 生产使用 `wss://`。

### 生产配置文件

- Docker Compose：`server/docker-compose.prod.yml`
- 环境变量模板：`server/.env.production.example`
- 你的生产环境变量文件：`server/.env.production`（已加入 `.gitignore`，不要提交）

### 发布步骤（Docker Compose）

```bash
cd server

# 1) 复制生产环境变量模板
cp .env.production.example .env.production

# 2) 编辑 .env.production（填写真实密钥与数据库/Redis 密码）

# 3) 构建并启动（生产）
docker compose -f docker-compose.prod.yml --env-file .env.production up -d --build

# 4) 健康检查
curl http://127.0.0.1:8080/health
```

### 发布步骤（单机/裸机，不使用 Docker）

适用于你希望 API 以 systemd 服务运行（MySQL/Redis 可以是本机安装，也可以是内网的托管服务）。

#### 1) 准备环境

- Linux 服务器（建议 Ubuntu/Debian/CentOS）
- MySQL 8.0 与 Redis 7（本机或内网可达）
- 域名与 HTTPS（建议 Nginx/Caddy 反代到 `127.0.0.1:8080`）

#### 2) 生成生产配置文件

在服务器上创建 `/etc/fz_yyc_api/.env.production`（不要放到代码仓库），参考模板 `server/.env.production.example`。

关键项：

- `APP_ENV=production`
- `APP_DEBUG=false`
- `JWT_SECRET` 必须改为强随机
- `WECHAT_PAY_CALLBACK_URL` 必须是生产 HTTPS 回调地址

#### 3) 编译并部署二进制

```bash
cd server
go build -o fz_yyc_api ./cmd/server

sudo install -m 0755 fz_yyc_api /usr/local/bin/fz_yyc_api
sudo mkdir -p /etc/fz_yyc_api
sudo mkdir -p /var/log/fz_yyc_api
```

#### 4) 配置 systemd 服务并启动

- systemd 模板：`server/deploy/fz_yyc_api.service.example`

```bash
sudo cp deploy/fz_yyc_api.service.example /etc/systemd/system/fz_yyc_api.service
sudo systemctl daemon-reload
sudo systemctl enable --now fz_yyc_api

sudo systemctl status fz_yyc_api --no-pager
curl http://127.0.0.1:8080/health
```

#### 5) 配置 Nginx 反向代理（可选但强烈建议）

- Nginx 配置模板：`server/deploy/nginx.fz_yyc_api.conf.example`
- 反代到 `127.0.0.1:8080`，并开启 HTTPS（WebSocket 走 `wss://`）

### 反向代理与域名（建议）

- 反向代理将 `https://api.example.com` 转发到 `http://127.0.0.1:8080`
- 小程序端接口基址配置为 `https://api.example.com`，WebSocket 基址配置为 `wss://api.example.com`
- 微信小程序后台需配置“服务器域名”（request/socket/upload/download）并开启 HTTPS

### 本地开发配置

如果需要在本地直接运行，需要编辑 `server/.env` 文件：

```bash
# 数据库配置 - 本地环境
DB_HOST=localhost
DB_PORT=3306

# Redis 配置 - 本地环境
REDIS_HOST=localhost
REDIS_PORT=6379
```

## 数据库连接信息

### Docker 环境（容器内部）

| 服务    | 主机（容器内） | 端口   | 用户      | 密码         | 数据库          |
| ----- | ------- | ---- | ------- | ---------- | ------------ |
| MySQL | mysql   | 3306 | fz\_yyc | fz\_yyc123 | fz\_yyc\_api |
| Redis | redis   | 6379 | -       | redis123   | -            |

### 本地环境（从宿主机访问）

| 服务    | 主机（宿主机）   | 端口   | 用户      | 密码         | 数据库          |
| ----- | --------- | ---- | ------- | ---------- | ------------ |
| MySQL | localhost | 3306 | fz\_yyc | fz\_yyc123 | fz\_yyc\_api |
| Redis | localhost | 6379 | -       | redis123   | -            |
| API   | localhost | 8080 | -       | -          | -            |

## API 服务信息

- **Docker 访问地址**：`http://localhost:8080`
- **健康检查**：`http://localhost:8080/health`
- **API 文档**：可通过微信开发者工具或浏览器访问

## 健康检查

### Docker 健康检查

```bash
# 检查 API 服务健康状态
docker inspect --format='{{.State.Health.Status}}' fz_yyc_api

# 查看所有服务健康状态
docker ps --filter "name=fz_yyc" --format "table {{.Names}}\t{{.Status}}"
```

### API 健康检查

```bash
# 测试健康端点
curl http://localhost:8080/health

# 预期响应
{"status":"ok"}
```

## 常用开发流程

### 使用 Docker（推荐）

1. **启动所有服务**：
   ```bash
   cd server
   docker compose up -d
   ```
2. **查看日志确认启动**：
   ```bash
   docker compose logs -f api
   ```
3. **测试 API**：
   ```bash
   curl http://localhost:8080/health
   ```
4. **修改代码后重新构建**：
   ```bash
   docker compose up --build -d
   ```
5. **停止所有服务**：
   ```bash
   docker compose down
   ```

### 使用本地 Go 环境

1. **启动 Docker 数据库**：
   ```bash
   cd server
   docker compose up -d mysql redis
   ```
2. **修改 .env 文件**：
   - 将 `DB_HOST=mysql` 改为 `DB_HOST=localhost`
   - 将 `REDIS_HOST=redis` 改为 `REDIS_HOST=localhost`
3. **启动后端服务**：
   ```bash
   go run cmd/server/main.go
   ```
4. **测试 API**：
   ```bash
   curl http://localhost:8080/health
   ```
5. **停止服务**：
   - 按 `Ctrl + C` 停止后端服务
   - 执行 `docker compose down` 停止数据库

## 故障排查

### 查看日志

```bash
# 查看 API 启动错误
docker compose logs api

# 查看数据库连接错误
docker compose logs api | grep "database"

# 实时查看所有日志
docker compose logs -f
```

### 进入容器调试

```bash
# 进入 API 容器
docker exec -it fz_yyc_api sh

# 进入 MySQL 容器
docker exec -it fz_yyc_mysql mysql -uroot -proot123456

# 进入 Redis 容器
docker exec -it fz_yyc_redis redis-cli -a redis123
```

### 检查网络连接

```bash
# 从 API 容器测试连接数据库
docker exec fz_yyc_api ping mysql

# 从 API 容器测试连接 Redis
docker exec fz_yyc_api ping redis

# 查看网络信息
docker network inspect server_fz_yyc_network
```

## 注意事项

- **首次启动**：需要等待 MySQL 完全初始化（约10-20秒），可以使用 `docker compose logs -f mysql` 观察进度
- **数据持久化**：MySQL 和 Redis 的数据存储在 Docker volumes 中，删除容器不会丢失数据
- **完全清理**：执行 `docker compose down -v` 会删除所有数据，请谨慎使用
- **生产环境**：需要配置反向代理（如 Nginx）和 HTTPS
- **备份数据**：定期备份 MySQL 数据卷，可使用 `docker run --rm -v server_mysql_data:/data -v $(pwd):/backup alpine tar czf /backup/mysql_backup.tar.gz /data`

