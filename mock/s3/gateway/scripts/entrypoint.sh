#!/bin/bash
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 函数：等待Consul可用
wait_for_consul() {
    local consul_addr=${CONSUL_HTTP_ADDR:-consul:8500}
    local max_attempts=${1:-30}
    local wait_time=${2:-2}
    
    log_info "等待 Consul 服务启动... ($consul_addr)"
    
    for i in $(seq 1 $max_attempts); do
        if curl -f -s --connect-timeout 1 --max-time 3 "http://$consul_addr/v1/status/leader" > /dev/null 2>&1; then
            log_info "Consul 服务已启动"
            return 0
        fi
        
        log_debug "尝试 $i/$max_attempts: Consul 服务尚未就绪"
        sleep $wait_time
    done
    
    log_error "Consul 服务在 ${max_attempts} 次尝试后仍未就绪"
    return 1
}

# 函数：启动consul-template
start_consul_template() {
    local consul_addr=${CONSUL_HTTP_ADDR:-consul:8500}
    
    log_info "启动 consul-template..."
    
    # 创建consul-template配置
    cat > /tmp/consul-template.hcl <<EOF
consul {
  address = "$consul_addr"
  retry {
    enabled = true
    attempts = 12
    backoff = "250ms"
    max_backoff = "1m"
  }
}

log_level = "${CONSUL_TEMPLATE_LOG:-INFO}"

template {
  source      = "/etc/nginx/templates/upstreams.conf.ctmpl"
  destination = "/etc/nginx/conf.d/upstreams.conf"
  command     = "nginx -t && nginx -s reload || echo 'Config validation failed, keeping current config'"
  command_timeout = "60s"
  perms = 0644
}
EOF
    
    # 启动consul-template作为后台进程
    consul-template -config=/tmp/consul-template.hcl &
    local consul_template_pid=$!
    
    log_info "consul-template 已启动 (PID: $consul_template_pid)"
    echo $consul_template_pid > /var/run/consul-template.pid
    
    return 0
}

# 函数：检查nginx配置
check_nginx_config() {
    log_info "检查 Nginx 配置..."
    
    if nginx -t; then
        log_info "Nginx 配置验证通过"
        return 0
    else
        log_error "Nginx 配置验证失败"
        return 1
    fi
}

# 函数：处理信号
handle_signal() {
    log_info "接收到停止信号，正在关闭服务..."
    
    # 停止consul-template
    if [ -f /var/run/consul-template.pid ]; then
        local ct_pid=$(cat /var/run/consul-template.pid)
        log_info "停止 consul-template (PID: $ct_pid)..."
        kill $ct_pid 2>/dev/null || true
        rm -f /var/run/consul-template.pid
    fi
    
    # 停止nginx
    log_info "停止 Nginx..."
    nginx -s quit
    wait $nginx_pid 2>/dev/null || true
    
    log_info "所有服务已停止"
    exit 0
}

# 主函数
main() {
    log_info "Mock S3 Gateway 正在启动..."
    
    # 设置Consul地址
    export CONSUL_HTTP_ADDR=${CONSUL_HTTP_ADDR:-consul:8500}
    
    log_info "配置信息："
    log_info "  Consul地址: $CONSUL_HTTP_ADDR"
    log_info "  日志级别: ${CONSUL_TEMPLATE_LOG:-INFO}"
    
    # 先创建默认的upstream配置，确保nginx能正常启动
    log_info "创建默认upstream配置..."
    cat > /etc/nginx/conf.d/upstreams.conf <<EOF
# 默认upstream配置
upstream storage_service {
    server storage-service:8082 max_fails=3 fail_timeout=30s;
    keepalive 32;
}
upstream metadata_service {
    server metadata-service:8081 max_fails=3 fail_timeout=30s;
    keepalive 32;
}
upstream queue_service {
    server queue-service:8083 max_fails=3 fail_timeout=30s;
    keepalive 32;
}
upstream third_party_service {
    server third-party-service:8084 max_fails=3 fail_timeout=30s;
    keepalive 32;
}
upstream mock_error_service {
    server mock-error-service:8085 max_fails=3 fail_timeout=30s;
    keepalive 32;
}
EOF

    # 等待Consul服务启动
    if ! wait_for_consul 60 2; then
        log_warn "Consul服务不可用，使用默认配置启动"
    else
        # 启动consul-template进行动态配置
        start_consul_template
        
        # 等待初始配置生成
        log_info "等待初始upstream配置生成..."
        for i in {1..30}; do
            if [ -f /etc/nginx/conf.d/upstreams.conf ]; then
                log_info "初始upstream配置已生成"
                break
            fi
            sleep 1
        done
        
        # 如果配置文件仍不存在，创建默认配置
        if [ ! -f /etc/nginx/conf.d/upstreams.conf ]; then
            log_warn "consul-template未能及时生成配置，使用默认配置"
            cat > /etc/nginx/conf.d/upstreams.conf <<EOF
# 临时默认配置
upstream storage_service { server 127.0.0.1:9999 down; keepalive 32; }
upstream metadata_service { server 127.0.0.1:9999 down; keepalive 32; }
upstream queue_service { server 127.0.0.1:9999 down; keepalive 32; }
upstream third_party_service { server 127.0.0.1:9999 down; keepalive 32; }
upstream mock_error_service { server 127.0.0.1:9999 down; keepalive 32; }
EOF
        fi
    fi
    
    # 处理模板文件
    if [ -d "/etc/nginx/templates" ] && [ "$(ls -A /etc/nginx/templates 2>/dev/null)" ]; then
        log_info "处理 Nginx 模板文件..."
        for template in /etc/nginx/templates/*.template; do
            if [ -f "$template" ]; then
                output_file="/etc/nginx/conf.d/$(basename "$template" .template)"
                log_debug "处理模板: $template -> $output_file"
                envsubst < "$template" > "$output_file"
            fi
        done
    fi
    
    # 检查配置文件
    if ! check_nginx_config; then
        log_error "Nginx 配置错误，退出"
        exit 1
    fi
    
    # 设置信号处理
    trap handle_signal TERM INT QUIT
    
    # 启动 Nginx
    log_info "启动 Nginx 服务..."
    
    # 如果有参数传入，执行传入的命令
    if [ $# -gt 0 ]; then
        log_info "执行命令: $*"
        exec "$@"
    else
        # 启动nginx并获取PID
        nginx -g "daemon off;" &
        nginx_pid=$!
        
        log_info "Mock S3 Gateway 已启动 (PID: $nginx_pid)"
        log_info "Gateway URL: http://0.0.0.0:8080"
        log_info "Health Check: http://0.0.0.0:8080/health"
        
        # 等待nginx进程
        wait $nginx_pid
    fi
}

# 执行主函数
main "$@"