# 多阶段构建 - 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 编译应用
WORKDIR /app/services/third-party
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/third-party-service ./cmd

# 运行阶段
FROM alpine:latest

# 安装ca-certificates用于HTTPS调用
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/third-party-service .

# 复制配置文件
COPY --from=builder /app/services/third-party/config ./config
COPY --from=builder /app/shared/observability/config/observability.yaml ./

# 创建数据目录
RUN mkdir -p /data

# 暴露端口
EXPOSE 8084

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8084/health || exit 1

# 运行应用
CMD ["./third-party-service"]