# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 使用更稳定的 Go 模块代理，避免 Docker 构建阶段偶发拉取失败。
ENV GOPROXY=https://goproxy.cn,direct

# 安装构建依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fz_yyc_api ./cmd/server

# 运行阶段
FROM alpine:3.20

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 app && \
    adduser -u 1000 -G app -s /bin/sh -D app

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=builder /app/fz_yyc_api .

# 复制迁移文件
COPY migrations ./migrations

# 更改文件所有者
RUN chown -R app:app /app

# 切换到非root用户
USER app

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令
CMD ["./fz_yyc_api"]
