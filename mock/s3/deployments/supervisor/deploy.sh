#!/bin/bash
# MockS3 用户级部署脚本（无需sudo权限）

set -e

# 配置
DEPLOY_USER="qboxserver"
DEPLOY_BASE="/home/qboxserver"

# 服务列表和端口（使用 zeroops_ 前缀）
declare -A SERVICES=(
    ["zeroops_metadata_1"]="8182"
    ["zeroops_metadata_2"]="8183"
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
                # 修改服务端口号（只修改service部分的port，不修改database/redis等其他部分的port）
                port="${SERVICES[$service_dir]}"
                # 使用awk来只修改service部分的port
                awk -v port="$port" '
                    /^service:/ {in_service=1}
                    /^[^ ]/ && !/^  / && !/^service:/ {in_service=0}
                    in_service && /^  port:/ {sub(/port: [0-9]+/, "port: " port)}
                    {print}
                ' "$dst_config" > "$dst_config.tmp" && mv "$dst_config.tmp" "$dst_config"
                # 修复配置文件中的主机名和端口

                # 1. 替换 host 字段（database和redis的host）
                sed -i 's/host: "postgres"/host: "127.0.0.1"/g' "$dst_config" 2>/dev/null || true
                sed -i 's/host: "redis"/host: "127.0.0.1"/g' "$dst_config" 2>/dev/null || true

                # 2. 替换 Consul address
                sed -i 's/address: "consul:8500"/address: "127.0.0.1:8500"/g' "$dst_config" 2>/dev/null || true

                # 3. 替换 Redis 配置（必须在修改端口之前）
                # metadata服务的redis address格式
                sed -i 's/address: "redis:6379"/address: "127.0.0.1:16379"/g' "$dst_config" 2>/dev/null || true
                # queue服务的redis url格式
                sed -i 's|url: "redis://redis:6379"|url: "redis://127.0.0.1:16379"|g' "$dst_config" 2>/dev/null || true

                # 4. 修改PostgreSQL端口（metadata服务专用）
                if [[ "$config_name" == "metadata" ]]; then
                    # 只修改database部分的port
                    sed -i '/^database:/,/^[^ ]/ { s/port: 5432/port: 5532/; }' "$dst_config" 2>/dev/null || true
                fi

                # 5. 替换 PostgreSQL URL（如果存在）
                sed -i 's|postgres://\([^:]*\):\([^@]*\)@postgres:5432|postgres://\1:\2@127.0.0.1:5532|g' "$dst_config" 2>/dev/null || true

                # 6. 替换 OpenTelemetry endpoint
                sed -i 's/otlp_endpoint: "otel-collector:4318"/otlp_endpoint: "127.0.0.1:4318"/g' "$dst_config" 2>/dev/null || true
                log_info "  部署配置: ${config_name}-config.yaml (端口: $port)"

                # 显示关键配置（用于调试）
                if [[ "$config_name" == "metadata" ]]; then
                    pg_port=$(grep -A 5 "^database:" "$dst_config" | grep "port:" | awk '{print $2}')
                    log_info "    PostgreSQL端口: $pg_port"
                fi
                if [[ "$config_name" == "queue" ]]; then
                    redis_url=$(grep "url:" "$dst_config" | head -1 | awk '{print $2}')
                    log_info "    Redis URL: $redis_url"
                fi
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