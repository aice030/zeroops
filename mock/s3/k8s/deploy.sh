#!/bin/bash

# MockS3 Kubernetes 部署脚本
# 使用方法: ./deploy.sh [build|deploy|scale|clean|status]

set -e

# 配置变量
NAMESPACE="zeroops"
REGISTRY="d1ng404"  # Docker Hub用户名
VERSION="v1.0.0"
KUBECTL_CONTEXT=""  # 可选：指定kubectl context

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'  
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "docker 未安装，请先安装 docker"
        exit 1
    fi
    
    log_info "Docker 检查通过"
}

# 检查Kubernetes依赖
check_k8s_dependencies() {
    log_info "检查Kubernetes依赖..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装，请先安装 kubectl"
        exit 1
    fi
    
    # 检查kubectl连接（使用用户有权限的命名空间）
    if ! kubectl get pods -n zeroops &> /dev/null; then
        log_error "kubectl 无法连接到集群或没有zeroops命名空间权限，请检查配置"
        log_info "请先配置kubeconfig文件连接到KubeSphere集群"
        exit 1
    fi
    
    log_info "Kubernetes 依赖检查通过"
}

# 构建镜像
build_images() {
    log_info "重新标记并推送 Docker 镜像..."
    
    # 业务服务镜像
    services=("metadata-service" "storage-service" "queue-service" "third-party-service" "mock-error-service")
    
    for service in "${services[@]}"; do
        local_image="mock-s3/$service:latest"
        remote_image="${REGISTRY}/mock-s3-$service:${VERSION}"
        
        log_info "检查本地镜像 $local_image..."
        if docker images --format "{{.Repository}}:{{.Tag}}" | grep -q "^$local_image$"; then
            log_info "重新标记 $service 镜像: $remote_image"
            docker tag $local_image $remote_image
            
            log_info "推送 $service 镜像到仓库..."
            docker push $remote_image
        else
            log_warn "本地镜像 $local_image 不存在，跳过"
        fi
    done
    
    log_info "镜像推送完成"
}

# 更新镜像地址
update_image_references() {
    log_info "更新 Kubernetes 配置中的镜像地址..."
    
    # 在业务服务配置中替换镜像地址
    sed -i.bak "s|your-registry.com|${REGISTRY}|g" 06-business-services.yaml
    sed -i.bak "s|v1.0.0|${VERSION}|g" 06-business-services.yaml
    
    log_info "镜像地址更新完成"
}

# 部署到 Kubernetes
deploy_to_k8s() {
    log_info "开始部署到 Kubernetes..."
    
    # 设置 kubectl context (如果指定)
    if [ -n "$KUBECTL_CONTEXT" ]; then
        kubectl config use-context $KUBECTL_CONTEXT
    fi
    
    # 按顺序部署
    log_info "1. 检查命名空间（跳过创建，使用现有zeroops命名空间）..."
    kubectl get namespace $NAMESPACE || log_error "命名空间 $NAMESPACE 不存在"
    
    log_info "2. 创建配置和密钥..."
    kubectl apply -f 02-configmaps.yaml
    
    log_info "3. 创建持久化存储..."
    kubectl apply -f 03-storage.yaml
    
    log_info "4. 部署基础设施服务..."
    kubectl apply -f 04-infrastructure.yaml
    
    log_info "等待基础设施服务就绪..."
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=postgres --timeout=300s
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=redis --timeout=300s
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=consul --timeout=300s
    
    log_info "5. 部署监控组件..."
    kubectl apply -f 05-monitoring.yaml
    
    log_info "等待监控组件就绪..."
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=prometheus --timeout=300s
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=grafana --timeout=300s
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=otel-collector --timeout=300s
    
    log_info "6. 部署业务服务..."
    kubectl apply -f 06-business-services.yaml
    
    log_info "等待业务服务就绪..."
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=metadata-service --timeout=300s
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=storage-service --timeout=300s
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=queue-service --timeout=300s
    kubectl wait --namespace=$NAMESPACE --for=condition=ready pod -l app=third-party-service --timeout=300s
    
    log_info "7. 创建访问入口..."
    kubectl apply -f 07-ingress.yaml
    
    log_info "部署完成！"
}

# 扩缩容操作
scale_services() {
    local replicas=${2:-3}  # 默认3个副本
    
    log_info "扩缩容业务服务到 $replicas 个副本..."
    
    kubectl scale --namespace=$NAMESPACE deployment metadata-service --replicas=$replicas
    kubectl scale --namespace=$NAMESPACE deployment storage-service --replicas=$replicas
    kubectl scale --namespace=$NAMESPACE deployment queue-service --replicas=$replicas
    kubectl scale --namespace=$NAMESPACE deployment third-party-service --replicas=$replicas
    
    log_info "扩缩容操作完成"
}

# 查看状态
show_status() {
    log_info "查看部署状态..."
    
    echo "=== 命名空间 ==="
    kubectl get namespace $NAMESPACE
    
    echo "=== Pod 状态 ==="
    kubectl get pods --namespace=$NAMESPACE -o wide
    
    echo "=== Service 状态 ==="
    kubectl get services --namespace=$NAMESPACE
    
    echo "=== Ingress 状态 ==="
    kubectl get ingress --namespace=$NAMESPACE
    
    echo "=== PVC 状态 ==="
    kubectl get pvc --namespace=$NAMESPACE
    
    echo "=== 访问地址 ==="
    echo "NodePort 访问方式："
    echo "- Metadata Service: http://<节点IP>:30081"
    echo "- Storage Service: http://<节点IP>:30082" 
    echo "- Queue Service: http://<节点IP>:30083"
    echo "- Third-Party Service: http://<节点IP>:30084"
    echo "- Grafana: http://<节点IP>:30300 (admin/admin)"
    echo "- Prometheus: http://<节点IP>:30900"
    echo "- Consul: http://<节点IP>:30500"
}

# 清理资源
cleanup() {
    log_warn "清理所有资源..."
    
    read -p "确定要删除所有 MockS3 资源吗? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete -f 07-ingress.yaml --ignore-not-found=true
        kubectl delete -f 06-business-services.yaml --ignore-not-found=true
        kubectl delete -f 05-monitoring.yaml --ignore-not-found=true
        kubectl delete -f 04-infrastructure.yaml --ignore-not-found=true
        kubectl delete -f 03-storage.yaml --ignore-not-found=true
        kubectl delete -f 02-configmaps.yaml --ignore-not-found=true
        kubectl delete -f 01-namespace.yaml --ignore-not-found=true
        
        log_info "清理完成"
    else
        log_info "取消清理"
    fi
}

# 主函数
main() {
    case $1 in
        build)
            check_dependencies
            build_images
            ;;
        deploy)
            check_k8s_dependencies
            update_image_references
            deploy_to_k8s
            show_status
            ;;
        scale)
            check_k8s_dependencies
            scale_services $@
            show_status
            ;;
        status)
            check_k8s_dependencies
            show_status
            ;;
        clean)
            check_k8s_dependencies
            cleanup
            ;;
        *)
            echo "用法: $0 {build|deploy|scale|status|clean}"
            echo ""
            echo "命令说明:"
            echo "  build   - 构建并推送Docker镜像"
            echo "  deploy  - 部署到Kubernetes集群"
            echo "  scale   - 扩缩容服务 (用法: $0 scale <副本数>)"
            echo "  status  - 查看部署状态"
            echo "  clean   - 清理所有资源"
            echo ""
            echo "示例:"
            echo "  $0 build                 # 构建镜像"
            echo "  $0 deploy                # 部署到K8s"
            echo "  $0 scale 5               # 扩容到5个副本"
            echo "  $0 status                # 查看状态"
            echo "  $0 clean                 # 清理资源"
            exit 1
            ;;
    esac
}

main $@