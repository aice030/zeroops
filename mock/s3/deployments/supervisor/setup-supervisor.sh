#!/bin/bash
# MockS3 Supervisor 快速部署脚本（符合公司规范）

set -e

# 配置
DEPLOY_USER="qboxserver"
DEPLOY_BASE="/home/qboxserver"
SUPERVISOR_CONF_DIR="/etc/supervisord"

# 服务列表和端口
declare -A SERVICES=(
    ["metadata-service_1"]="8181"
    ["metadata-service_2"]="8182"
    ["metadata-service_3"]="8183"
    ["storage-service_1"]="8191"
    ["storage-service_2"]="8192"
    ["queue-service_1"]="8201"
    ["queue-service_2"]="8202"
    ["third-party-service_1"]="8211"
    ["mock-error-service_1"]="8221"
)

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 创建目录结构
setup_directories() {
    log_info "创建服务目录结构..."

    for service_dir in "${!SERVICES[@]}"; do
        dir="$DEPLOY_BASE/$service_dir"

        # 创建目录
        mkdir -p "$dir/_package"
        mkdir -p "$dir/logs"
        mkdir -p "$dir/config"
        mkdir -p "$dir/data"

        # 设置权限
        chown -R "$DEPLOY_USER:$DEPLOY_USER" "$dir"

        log_info "  创建: $dir"
    done
}

# 部署二进制文件
deploy_binaries() {
    log_info "部署二进制文件..."

    # 检查bin目录
    if [ ! -d "bin" ]; then
        log_error "bin目录不存在，请先解压部署包"
        exit 1
    fi

    for service_dir in "${!SERVICES[@]}"; do
        # 获取服务基础名称（去掉_数字）
        base_name=$(echo "$service_dir" | sed 's/_[0-9]*$//')

        src="bin/$base_name"
        dst="$DEPLOY_BASE/$service_dir/_package/$base_name"

        if [ -f "$src" ]; then
            cp "$src" "$dst"
            chmod +x "$dst"
            chown "$DEPLOY_USER:$DEPLOY_USER" "$dst"
            log_info "  部署: $base_name -> $service_dir"
        else
            log_error "  文件不存在: $src"
        fi
    done
}

# 安装Supervisor配置
install_supervisor_conf() {
    log_info "安装Supervisor配置..."

    # 备份现有配置
    if [ -f "$SUPERVISOR_CONF_DIR/mock-s3.conf" ]; then
        mv "$SUPERVISOR_CONF_DIR/mock-s3.conf" "$SUPERVISOR_CONF_DIR/mock-s3.conf.bak.$(date +%Y%m%d%H%M%S)"
        log_info "备份现有配置"
    fi

    # 使用简化配置
    if [ -f "mock-s3-simple.conf" ]; then
        cp "mock-s3-simple.conf" "$SUPERVISOR_CONF_DIR/mock-s3.conf"
    else
        # 动态生成配置
        cat > "$SUPERVISOR_CONF_DIR/mock-s3.conf" <<EOF
# MockS3 Supervisor Configuration
# Generated at $(date)

EOF

        for service_dir in "${!SERVICES[@]}"; do
            base_name=$(echo "$service_dir" | sed 's/_[0-9]*$//')
            port="${SERVICES[$service_dir]}"
            program_name="mock-s3-$(echo $service_dir | tr '_' '-')"

            cat >> "$SUPERVISOR_CONF_DIR/mock-s3.conf" <<EOF
[program:$program_name]
environment=HOME="/home/qboxserver"
command=$DEPLOY_BASE/$service_dir/_package/$base_name --port=$port
directory=$DEPLOY_BASE/$service_dir/_package
priority=999
autostart=true
startsecs=1
autorestart=true
user=$DEPLOY_USER

EOF
        done
    fi

    log_info "配置已安装到: $SUPERVISOR_CONF_DIR/mock-s3.conf"
}

# 启动服务
start_services() {
    log_info "启动服务..."

    # 重新加载配置
    supervisorctl reread
    supervisorctl update

    # 启动所有mock-s3服务
    supervisorctl start 'mock-s3-*'

    # 等待服务启动
    sleep 3

    # 显示状态
    log_info "服务状态:"
    supervisorctl status | grep mock-s3
}

# 健康检查
health_check() {
    log_info "执行健康检查..."

    for service_dir in "${!SERVICES[@]}"; do
        port="${SERVICES[$service_dir]}"

        if curl -sf "http://localhost:$port/health" >/dev/null 2>&1; then
            log_info "  ✓ $service_dir (端口: $port) - 正常"
        else
            log_error "  ✗ $service_dir (端口: $port) - 异常"
        fi
    done
}

# 主函数
main() {
    log_info "MockS3 Supervisor 部署"

    # 检查权限
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用sudo执行此脚本"
        exit 1
    fi

    # 执行部署步骤
    setup_directories
    deploy_binaries
    install_supervisor_conf
    start_services
    health_check

    log_info "部署完成！"
    log_info "使用 'supervisorctl status | grep mock-s3' 查看服务状态"
}

# 执行
main "$@"