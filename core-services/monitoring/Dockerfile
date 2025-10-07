# 太上老君监控系统 Dockerfile
# 多阶段构建，优化镜像大小

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o monitoring main.go

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata curl

# 创建非 root 用户
RUN addgroup -g 1001 -S monitoring && \
    adduser -u 1001 -S monitoring -G monitoring

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/monitoring .

# 复制配置文件
COPY --from=builder /app/config ./config

# 创建必要的目录
RUN mkdir -p /var/log/monitoring /var/lib/monitoring && \
    chown -R monitoring:monitoring /app /var/log/monitoring /var/lib/monitoring

# 切换到非 root 用户
USER monitoring

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 设置环境变量
ENV MONITORING_SERVICE_ENVIRONMENT=production
ENV MONITORING_SERVICE_LOG_LEVEL=info

# 启动命令
CMD ["./monitoring", "-config", "config/monitoring.yaml"]

# 元数据标签
LABEL maintainer="taishanglaojun@example.com"
LABEL version="1.0.0"
LABEL description="太上老君监控系统"
LABEL org.opencontainers.image.source="https://github.com/taishanglaojun/core-services"
LABEL org.opencontainers.image.documentation="https://github.com/taishanglaojun/core-services/tree/main/monitoring"
LABEL org.opencontainers.image.licenses="MIT"