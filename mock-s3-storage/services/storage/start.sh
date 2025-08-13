#!/bin/bash

# 文件存储服务启动脚本

echo "🚀 启动文件存储服务..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go 1.21或更高版本"
    exit 1
fi

# 启动PostgreSQL数据库
echo "📦 启动PostgreSQL数据库..."
docker-compose up -d postgres

# 等待数据库启动
echo "⏳ 等待数据库启动..."
sleep 10

# 检查数据库连接
echo "🔍 检查数据库连接..."
until docker exec postgres-file-storage pg_isready -U postgres; do
    echo "⏳ 等待数据库就绪..."
    sleep 2
done

echo "✅ 数据库已就绪"

# 安装Go依赖
echo "📥 安装Go依赖..."
go mod tidy

# 启动服务
echo "🌐 启动文件存储服务..."
echo "服务将在 http://localhost:8080 启动"
echo "按 Ctrl+C 停止服务"
echo ""

go run cmd/main.go
