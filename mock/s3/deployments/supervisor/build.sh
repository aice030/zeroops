#!/bin/bash
# MockS3 生产环境构建脚本

set -e

# 配置
PROJECT_ROOT=$(dirname $(dirname $(realpath "$0")))
BUILD_DIR="$PROJECT_ROOT/build/output"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 打印函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 清理构建目录
clean() {
    log_info "清理构建目录..."
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"/{bin,config,scripts}
}

# 构建单个服务
build_service() {
    local service_name=$1
    local service_path=$2

    log_info "构建服务: $service_name"

    cd "$PROJECT_ROOT/$service_path"

    # 构建Linux二进制文件
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
        -o "$BUILD_DIR/bin/$service_name" \
        ./cmd

    # 复制配置文件
    if [ -d "config" ]; then
        cp -r config/* "$BUILD_DIR/config/" 2>/dev/null || true
    fi

    log_info "✓ $service_name 构建完成"
}

# 构建所有服务
build_all() {
    log_info "开始构建所有服务..."

    # 下载依赖
    cd "$PROJECT_ROOT"
    go mod download

    # 构建各个服务
    build_service "metadata-service" "services/metadata"
    build_service "storage-service" "services/storage"
    build_service "queue-service" "services/queue"
    build_service "third-party-service" "services/third-party"
    build_service "mock-error-service" "services/mock-error"

    # 复制共享配置
    cp -r "$PROJECT_ROOT/shared/observability/config/"* "$BUILD_DIR/config/" 2>/dev/null || true

    log_info "所有服务构建完成！"
}

# 创建部署包
create_package() {
    log_info "创建部署包..."

    cd "$BUILD_DIR"
    # 使用 COPYFILE_DISABLE=1 避免 macOS 扩展属性
    COPYFILE_DISABLE=1 tar -czf "mock-s3-$VERSION.tar.gz" bin/ config/ scripts/

    log_info "部署包创建完成: mock-s3-$VERSION.tar.gz"
    log_info "文件大小: $(du -h mock-s3-$VERSION.tar.gz | cut -f1)"
}

# 主流程
main() {
    log_info "MockS3 生产环境构建"
    log_info "版本: $VERSION"
    log_info "提交: $GIT_COMMIT"

    clean
    build_all
    create_package

    log_info "✅ 构建成功完成！"
    log_info "输出目录: $BUILD_DIR"
}

# 执行
main "$@"