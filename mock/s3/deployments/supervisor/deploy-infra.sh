#!/bin/bash
# 部署基础设施配置文件脚本

set -e

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

# 基础路径
BASE_DIR="/home/qboxserver/zeroops_compose"

# 创建目录结构
create_directories() {
    log_info "创建基础设施目录结构..."

    mkdir -p "$BASE_DIR"
    mkdir -p "$BASE_DIR/consul"
    mkdir -p "$BASE_DIR/postgres"
    mkdir -p "$BASE_DIR/redis"
    mkdir -p "$BASE_DIR/prometheus"
    mkdir -p "$BASE_DIR/grafana/provisioning/datasources"
    mkdir -p "$BASE_DIR/grafana/provisioning/dashboards"
    mkdir -p "$BASE_DIR/otel"

    log_info "目录结构创建完成"
}

# 复制配置文件
copy_configs() {
    log_info "部署配置文件..."

    # Docker Compose文件
    if [ -f "docker-compose.yml" ]; then
        cp docker-compose.yml "$BASE_DIR/"
        log_info "  复制: docker-compose.yml"
    fi

    # Consul配置
    if [ -f "consul/consul-config.json" ]; then
        cp consul/consul-config.json "$BASE_DIR/consul/"
        log_info "  复制: consul-config.json"
    fi

    # PostgreSQL初始化脚本
    if [ -f "postgres/init.sql" ]; then
        cp postgres/init.sql "$BASE_DIR/postgres/"
        log_info "  复制: init.sql"
    fi

    # Redis配置
    if [ -f "redis/redis.conf" ]; then
        cp redis/redis.conf "$BASE_DIR/redis/"
        log_info "  复制: redis.conf"
    fi

    # Prometheus配置
    if [ -f "../observability/prometheus.yml" ]; then
        cp ../observability/prometheus.yml "$BASE_DIR/prometheus/prometheus.yml"
        log_info "  复制: prometheus.yml"
    elif [ -f "observability/prometheus.yml" ]; then
        cp observability/prometheus.yml "$BASE_DIR/prometheus/prometheus.yml"
        log_info "  复制: prometheus.yml"
    fi

    # Grafana配置
    if [ -d "../observability/grafana" ]; then
        cp -r ../observability/grafana/* "$BASE_DIR/grafana/"
        log_info "  复制: Grafana配置"
    elif [ -d "observability/grafana" ]; then
        cp -r observability/grafana/* "$BASE_DIR/grafana/"
        log_info "  复制: Grafana配置"
    fi

    # OpenTelemetry配置
    if [ -f "../observability/otel-config.yaml" ]; then
        cp ../observability/otel-config.yaml "$BASE_DIR/otel/otel-config.yaml"
        log_info "  复制: otel-config.yaml"
    elif [ -f "observability/otel-config.yaml" ]; then
        cp observability/otel-config.yaml "$BASE_DIR/otel/otel-config.yaml"
        log_info "  复制: otel-config.yaml"
    fi
}

# 创建启动脚本
create_start_script() {
    cat > "$BASE_DIR/start-infra.sh" <<'EOF'
#!/bin/bash
# 启动基础设施服务
# 注意：需要sudo权限运行此脚本

cd /home/qboxserver/zeroops_compose

if [ "$EUID" -ne 0 ]; then
   echo "请使用sudo运行此脚本"
   echo "用法: sudo ./start-infra.sh"
   exit 1
fi

echo "启动基础设施服务..."
docker-compose -f docker-compose.yml up -d

echo "等待服务启动..."
sleep 10

echo "检查服务状态："
docker-compose -f docker-compose.yml ps

echo ""
echo "服务访问地址："
echo "  - Consul UI: http://localhost:8500"
echo "  - PostgreSQL: localhost:5532"
echo "  - Redis: localhost:16379"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (admin/admin123)"
echo "  - Elasticsearch: http://localhost:9200"
echo "  - Kibana: http://localhost:5601"
EOF
    chmod +x "$BASE_DIR/start-infra.sh"
    log_info "创建启动脚本: start-infra.sh"
}

# 创建停止脚本
create_stop_script() {
    cat > "$BASE_DIR/stop-infra.sh" <<'EOF'
#!/bin/bash
# 停止基础设施服务
# 注意：需要sudo权限运行此脚本

cd /home/qboxserver/zeroops_compose

if [ "$EUID" -ne 0 ]; then
   echo "请使用sudo运行此脚本"
   echo "用法: sudo ./stop-infra.sh"
   exit 1
fi

echo "停止基础设施服务..."
docker-compose -f docker-compose.yml down

echo "服务已停止"
EOF
    chmod +x "$BASE_DIR/stop-infra.sh"
    log_info "创建停止脚本: stop-infra.sh"
}

# 主函数
main() {
    log_info "开始部署基础设施配置..."

    create_directories
    copy_configs
    create_start_script
    create_stop_script

    log_info "部署完成！"
    echo ""
    log_info "后续步骤："
    echo ""
    log_info "方式1: 使用Supervisor管理（推荐，无需手动sudo）"
    echo "  1. 请联系管理员将Supervisor配置文件安装到系统："
    echo "     sudo cp /tmp/dingnanjia/mock-s3-infra.conf /etc/supervisord/"
    echo ""
    echo "  2. 通过Supervisor启动基础设施："
    echo "     sudo supervisorctl reread"
    echo "     sudo supervisorctl update"
    echo "     sudo supervisorctl start mock-s3-infra"
    echo ""
    log_info "方式2: 手动启动（需要sudo权限）"
    echo "     cd $BASE_DIR"
    echo "     sudo ./start-infra.sh"
    echo ""
    echo "  3. 等待基础设施服务启动后，再部署业务服务："
    echo "     ./deploy.sh"
    echo "     ./manual-start.sh"
}

main "$@"