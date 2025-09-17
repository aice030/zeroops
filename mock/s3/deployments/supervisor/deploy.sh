#!/bin/bash
# MockS3 用户级部署脚本（无需sudo权限）

set -e

# 配置
DEPLOY_USER="qboxserver"
DEPLOY_BASE="/home/qboxserver"

# 服务列表和端口（使用 zeroops_ 前缀）
declare -A SERVICES=(
    ["zeroops_metadata_1"]="8181"
    ["zeroops_metadata_2"]="8182"
    ["zeroops_metadata_3"]="8183"
    ["zeroops_storage_1"]="8191"
    ["zeroops_storage_2"]="8192"
    ["zeroops_queue_1"]="8201"
    ["zeroops_queue_2"]="8202"
    ["zeroops_third_party_1"]="8211"
    ["zeroops_mock_error_1"]="8221"
)

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
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
        # 获取服务基础名称（zeroops_metadata_1 -> metadata-service）
        base_name=$(echo "$service_dir" | sed 's/^zeroops_//' | sed 's/_[0-9]*$//')
        # 转换名称格式
        case "$base_name" in
            metadata) base_name="metadata-service" ;;
            storage) base_name="storage-service" ;;
            queue) base_name="queue-service" ;;
            third_party) base_name="third-party-service" ;;
            mock_error) base_name="mock-error-service" ;;
        esac

        src="bin/$base_name"
        dst="$DEPLOY_BASE/$service_dir/_package/$base_name"

        if [ -f "$src" ]; then
            cp "$src" "$dst"
            chmod +x "$dst"
            log_info "  部署: $base_name -> $service_dir"

            # 部署配置文件到 _package/config 目录
            mkdir -p "$DEPLOY_BASE/$service_dir/_package/config"
            # 调整配置文件名格式（去掉 -service 后缀）
            config_name=$(echo "$base_name" | sed 's/-service//')
            src_config="config/${config_name}-config.yaml"
            dst_config="$DEPLOY_BASE/$service_dir/_package/config/${config_name}-config.yaml"

            if [ -f "$src_config" ]; then
                cp "$src_config" "$dst_config"
                # 修改端口号
                port="${SERVICES[$service_dir]}"
                sed -i "s/port: [0-9]*/port: $port/g" "$dst_config" 2>/dev/null || true
                log_info "  部署配置: ${base_name}-config.yaml (端口: $port)"
            else
                log_warn "  配置文件不存在: $src_config"
            fi
        else
            log_error "  文件不存在: $src"
        fi
    done
}

# 创建启动脚本
create_start_scripts() {
    log_info "创建启动脚本..."

    for service_dir in "${!SERVICES[@]}"; do
        # 获取服务基础名称（zeroops_metadata_1 -> metadata-service）
        base_name=$(echo "$service_dir" | sed 's/^zeroops_//' | sed 's/_[0-9]*$//')
        case "$base_name" in
            metadata) base_name="metadata-service" ;;
            storage) base_name="storage-service" ;;
            queue) base_name="queue-service" ;;
            third_party) base_name="third-party-service" ;;
            mock_error) base_name="mock-error-service" ;;
        esac
        port="${SERVICES[$service_dir]}"

        # 创建启动脚本
        cat > "$DEPLOY_BASE/$service_dir/_package/start.sh" <<EOF
#!/bin/bash
export HOME="/home/qboxserver"
export SERVICE_PORT=$port
cd $DEPLOY_BASE/$service_dir/_package
exec ./$base_name --port=\$SERVICE_PORT
EOF
        chmod +x "$DEPLOY_BASE/$service_dir/_package/start.sh"

        # 创建停止脚本
        cat > "$DEPLOY_BASE/$service_dir/_package/stop.sh" <<EOF
#!/bin/bash
pkill -f "$DEPLOY_BASE/$service_dir/_package/$base_name"
EOF
        chmod +x "$DEPLOY_BASE/$service_dir/_package/stop.sh"

        log_info "  创建脚本: $service_dir"
    done
}

# 提示如何使用 supervisor
show_supervisor_usage() {
    log_info "部署完成！"
    echo ""
    log_info "Supervisor 配置文件已准备好: mock-s3-simple.conf"
    log_info "请联系管理员执行以下命令："
    echo ""
    echo "  # 复制配置文件到 supervisor 目录"
    echo "  sudo cp /tmp/dingnanjia/mock-s3-simple.conf /etc/supervisord/mock-s3.conf"
    echo ""
    echo "  # 重新加载并启动服务"
    echo "  sudo supervisorctl reread"
    echo "  sudo supervisorctl update"
    echo "  sudo supervisorctl start mock-s3-*"
    echo ""
    log_info "或者，你可以手动启动服务（无需 sudo）："
    echo ""
    echo "  # 启动单个服务"
    echo "  nohup /home/qboxserver/metadata-service_1/_package/start.sh > /home/qboxserver/metadata-service_1/logs/service.log 2>&1 &"
    echo ""
    echo "  # 或使用提供的启动脚本"
    echo "  ./manual-start.sh"
}

# 创建手动启动脚本
create_manual_start_script() {
    cat > manual-start.sh <<'EOF'
#!/bin/bash
# 手动启动所有服务（无需sudo）

GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}启动 MockS3 服务...${NC}"

# 启动服务函数
start_service() {
    local service_dir=$1

    echo "启动: $service_dir"
    cd /home/qboxserver/$service_dir/_package
    nohup ./start.sh > ../logs/service.log 2>&1 &
    echo "  PID: $!"
}

# 启动所有服务（使用 zeroops_ 前缀）
start_service zeroops_metadata_1
start_service zeroops_metadata_2
start_service zeroops_metadata_3
start_service zeroops_storage_1
start_service zeroops_storage_2
start_service zeroops_queue_1
start_service zeroops_queue_2
start_service zeroops_third_party_1
start_service zeroops_mock_error_1

echo ""
echo -e "${GREEN}所有服务已启动！${NC}"
echo "查看进程: ps aux | grep -E 'metadata-service|storage-service|queue-service|third-party-service|mock-error-service'"
EOF

    chmod +x manual-start.sh
    log_info "创建手动启动脚本: manual-start.sh"
}

# 创建停止脚本
create_manual_stop_script() {
    cat > manual-stop.sh <<'EOF'
#!/bin/bash
# 停止所有服务

RED='\033[0;31m'
NC='\033[0m'

echo -e "${RED}停止 MockS3 服务...${NC}"

pkill -f "metadata-service"
pkill -f "storage-service"
pkill -f "queue-service"
pkill -f "third-party-service"
pkill -f "mock-error-service"

echo "所有服务已停止"
EOF

    chmod +x manual-stop.sh
    log_info "创建停止脚本: manual-stop.sh"
}

# 主函数
main() {
    log_info "MockS3 用户级部署（无需 sudo）"

    # 执行部署步骤
    setup_directories
    deploy_binaries
    create_start_scripts
    create_manual_start_script
    create_manual_stop_script

    # 显示使用说明
    show_supervisor_usage
}

# 执行
main "$@"