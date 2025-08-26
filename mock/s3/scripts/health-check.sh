#!/bin/bash
# Mock S3 系统整体健康检查脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
CONSUL_ADDR=${CONSUL_ADDR:-localhost:8500}
GATEWAY_URL=${GATEWAY_URL:-http://localhost:8080}
TIMEOUT=${TIMEOUT:-10}

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_debug() {
    if [ "$DEBUG" = "true" ]; then
        echo -e "${BLUE}[DEBUG]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
    fi
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 >/dev/null 2>&1; then
        log_error "命令 $1 不存在，请安装"
        exit 1
    fi
}

# 检查服务健康状态
check_service_health() {
    local service_name=$1
    local service_url=$2
    
    log_debug "检查服务: $service_name ($service_url)"
    
    if curl -f -s --connect-timeout $TIMEOUT --max-time $TIMEOUT "$service_url/health" > /dev/null 2>&1; then
        log_info "✅ $service_name 服务健康"
        return 0
    else
        log_error "❌ $service_name 服务异常"
        return 1
    fi
}

# 检查Consul连接
check_consul() {
    log_info "检查 Consul 连接..."
    
    if curl -f -s --connect-timeout $TIMEOUT --max-time $TIMEOUT "http://$CONSUL_ADDR/v1/status/leader" > /dev/null 2>&1; then
        log_info "✅ Consul 连接正常"
        return 0
    else
        log_error "❌ Consul 连接失败"
        return 1
    fi
}

# 检查注册的服务
check_consul_services() {
    log_info "检查 Consul 中注册的服务..."
    
    local services_response
    if services_response=$(curl -f -s --connect-timeout $TIMEOUT --max-time $TIMEOUT "http://$CONSUL_ADDR/v1/agent/services" 2>/dev/null); then
        local service_count=$(echo "$services_response" | jq '. | length' 2>/dev/null || echo "0")
        log_info "Consul 中注册了 $service_count 个服务"
        
        # 检查特定服务
        local expected_services=("metadata-service" "storage-service" "queue-service" "third-party-service" "mock-error-service")
        local healthy_count=0
        
        for service in "${expected_services[@]}"; do
            if echo "$services_response" | jq -r "keys[]" 2>/dev/null | grep -q "$service"; then
                log_info "✅ $service 已注册"
                ((healthy_count++))
            else
                log_warn "⚠️  $service 未注册"
            fi
        done
        
        log_info "健康服务数量: $healthy_count/${#expected_services[@]}"
        return 0
    else
        log_error "❌ 无法获取 Consul 服务列表"
        return 1
    fi
}

# 检查Gateway功能
check_gateway_functionality() {
    log_info "检查 Gateway 功能..."
    
    # 检查Gateway健康状态
    if ! check_service_health "Gateway" "$GATEWAY_URL"; then
        return 1
    fi
    
    # 测试S3 API基本功能
    log_debug "测试基本 S3 API..."
    
    # 测试上传小文件
    local test_bucket="test-bucket"
    local test_key="health-check-$(date +%s).txt"
    local test_content="Health check content at $(date)"
    
    # PUT请求测试
    if echo "$test_content" | curl -f -s -X PUT \
        --connect-timeout $TIMEOUT --max-time $TIMEOUT \
        --data-binary @- \
        -H "Content-Type: text/plain" \
        "$GATEWAY_URL/$test_bucket/$test_key" > /dev/null 2>&1; then
        log_info "✅ S3 PUT 请求正常"
        
        # GET请求测试
        local retrieved_content
        if retrieved_content=$(curl -f -s --connect-timeout $TIMEOUT --max-time $TIMEOUT "$GATEWAY_URL/$test_bucket/$test_key" 2>/dev/null); then
            if [ "$retrieved_content" = "$test_content" ]; then
                log_info "✅ S3 GET 请求正常"
                
                # DELETE请求测试
                if curl -f -s -X DELETE --connect-timeout $TIMEOUT --max-time $TIMEOUT "$GATEWAY_URL/$test_bucket/$test_key" > /dev/null 2>&1; then
                    log_info "✅ S3 DELETE 请求正常"
                    return 0
                else
                    log_warn "⚠️  S3 DELETE 请求异常"
                fi
            else
                log_error "❌ S3 GET 内容不匹配"
            fi
        else
            log_error "❌ S3 GET 请求失败"
        fi
    else
        log_error "❌ S3 PUT 请求失败"
    fi
    
    return 1
}

# 检查系统资源
check_system_resources() {
    log_info "检查系统资源..."
    
    # 检查内存使用率
    if command -v free >/dev/null 2>&1; then
        local memory_usage=$(free | awk '/^Mem:/{printf("%.1f"), $3/$2*100}')
        if (( $(echo "$memory_usage > 90" | bc -l 2>/dev/null || echo 0) )); then
            log_warn "⚠️  内存使用率较高: ${memory_usage}%"
        else
            log_info "✅ 内存使用率正常: ${memory_usage}%"
        fi
    fi
    
    # 检查磁盘使用率
    if command -v df >/dev/null 2>&1; then
        local disk_usage=$(df / | awk 'NR==2{printf("%.1f"), $5}' | sed 's/%//')
        if (( $(echo "$disk_usage > 90" | bc -l 2>/dev/null || echo 0) )); then
            log_warn "⚠️  磁盘使用率较高: ${disk_usage}%"
        else
            log_info "✅ 磁盘使用率正常: ${disk_usage}%"
        fi
    fi
}

# 生成健康检查报告
generate_health_report() {
    log_info "生成健康检查报告..."
    
    local report_file="/tmp/mock-s3-health-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "system": "Mock S3",
  "overall_status": "$1",
  "consul_addr": "$CONSUL_ADDR",
  "gateway_url": "$GATEWAY_URL",
  "checks_performed": [
    "consul_connection",
    "consul_services",
    "gateway_health",
    "s3_api_functionality",
    "system_resources"
  ],
  "details": {
    "consul_available": $2,
    "services_registered": $3,
    "gateway_functional": $4,
    "s3_api_working": $5
  }
}
EOF
    
    log_info "健康检查报告已生成: $report_file"
    
    if [ "$SHOW_REPORT" = "true" ]; then
        echo "=== 健康检查报告 ==="
        cat "$report_file" | jq . 2>/dev/null || cat "$report_file"
        echo "=================="
    fi
}

# 主函数
main() {
    log_info "开始 Mock S3 系统健康检查..."
    
    # 检查必要命令
    check_command curl
    
    local overall_healthy=true
    local consul_ok=false
    local services_ok=false
    local gateway_ok=false
    local s3_api_ok=false
    
    # 执行各项检查
    if check_consul; then
        consul_ok=true
        if check_consul_services; then
            services_ok=true
        else
            overall_healthy=false
        fi
    else
        overall_healthy=false
    fi
    
    if check_gateway_functionality; then
        gateway_ok=true
        s3_api_ok=true
    else
        overall_healthy=false
    fi
    
    check_system_resources
    
    # 生成报告
    local status="healthy"
    if [ "$overall_healthy" = false ]; then
        status="unhealthy"
    fi
    
    generate_health_report "$status" "$consul_ok" "$services_ok" "$gateway_ok" "$s3_api_ok"
    
    # 输出总结
    echo ""
    echo "================================"
    if [ "$overall_healthy" = true ]; then
        log_info "🎉 Mock S3 系统整体健康！"
        exit 0
    else
        log_error "💥 Mock S3 系统存在问题，请检查上述错误信息"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    cat <<EOF
Mock S3 健康检查脚本

用法: $0 [选项]

选项:
  -h, --help              显示帮助信息
  -d, --debug             启用调试模式
  -t, --timeout SECONDS  设置超时时间 (默认: 10秒)
  -c, --consul ADDR       Consul地址 (默认: localhost:8500)
  -g, --gateway URL       Gateway地址 (默认: http://localhost:8080)
  -r, --report            显示详细报告
  
环境变量:
  DEBUG=true              启用调试输出
  CONSUL_ADDR            Consul地址
  GATEWAY_URL            Gateway URL
  TIMEOUT                超时时间（秒）
  SHOW_REPORT=true       显示详细报告

示例:
  $0                                    # 使用默认配置
  $0 -d -t 30                          # 调试模式，30秒超时
  $0 -c consul.local:8500              # 指定Consul地址
  SHOW_REPORT=true $0                  # 显示详细报告
EOF
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -d|--debug)
            export DEBUG=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -c|--consul)
            CONSUL_ADDR="$2"
            shift 2
            ;;
        -g|--gateway)
            GATEWAY_URL="$2"
            shift 2
            ;;
        -r|--report)
            export SHOW_REPORT=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 执行主函数
main "$@"