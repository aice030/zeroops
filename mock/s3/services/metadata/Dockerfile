# Build stage
FROM golang:1.24-alpine AS builder

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 构建二进制文件
WORKDIR /app/services/metadata
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/metadata-service ./cmd

# Runtime stage
FROM alpine:3.18

# 安装必要的运行时包
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/metadata-service /app/
COPY --from=builder /app/services/metadata/config/ /app/config/
COPY --from=builder /app/shared/observability/config/observability.yaml /app/

# 设置权限
RUN chown -R appuser:appgroup /app
USER appuser

# 暴露端口
EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# 启动命令
CMD ["./metadata-service"]