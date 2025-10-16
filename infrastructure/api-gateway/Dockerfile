# API Gateway Dockerfile

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o api-gateway ./cmd/main.go

# 运行阶段
FROM alpine:latest

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/api-gateway .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建日志目录
RUN mkdir -p /var/log/api-gateway && \
    chown -R appuser:appgroup /var/log/api-gateway

# 创建数据目录
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app/data

# 设置文件权限
RUN chmod +x api-gateway && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 设置环境变量
ENV GIN_MODE=release
ENV CONFIG_PATH=/app/configs/gateway.yaml

# 启动命令
CMD ["./api-gateway", "-config", "/app/configs/gateway.yaml"]

# 多阶段构建 - 开发版本
FROM golang:1.21-alpine AS dev

WORKDIR /app

# 安装开发工具
RUN apk add --no-cache git ca-certificates tzdata make curl

# 安装air用于热重载
RUN go install github.com/air-verse/air@latest

# 复制源代码
COPY . .

# 下载依赖
RUN go mod download

# 暴露端口
EXPOSE 8080 9090

# 开发模式启动命令
CMD ["air", "-c", ".air.toml"]