# MockS3 Kubernetes 部署指南

MockS3故障模拟平台的Kubernetes多实例部署方案，支持在KubeSphere等K8s平台上进行高可用部署。

## 📋 部署概览

### 架构特点
- **多实例部署**: 每个业务服务支持2-N个副本，实现高可用
- **服务网格**: 使用Kubernetes Service实现服务发现和负载均衡
- **共享存储**: storage-service使用ReadWriteMany PVC支持多Pod访问，保持3倍冗余
- **监控完整**: 包含Prometheus、Grafana、OpenTelemetry全栈监控
- **访问便捷**: 提供Ingress和NodePort两种访问方式

### 服务拓扑
```
┌─────────────────────────────────────────┐
│              Load Balancer              │
│          (Ingress Controller)          │
└──────────────┬──────────────────────────┘
               │
┌─────────────────────────────────────────┐
│           Business Services             │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │metadata │ │ storage │ │  queue  │   │
│  │(2 pods) │ │(2 pods) │ │(2 pods) │   │
│  └─────────┘ └─────────┘ └─────────┘   │
│  ┌─────────┐                           │
│  │3rd-party│                           │
│  │(2 pods) │                           │
│  └─────────┘                           │
└─────────────┬───────────────────────────┘
              │
┌─────────────────────────────────────────┐
│          Infrastructure                 │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │PostgreSQL│ │  Redis  │ │ Consul  │   │
│  │(StatefulS)│ │(StatefulS)│(StatefulS)│   │
│  └─────────┘ └─────────┘ └─────────┘   │
└─────────────────────────────────────────┘
```

## 🚀 快速部署

### 前置要求

1. **Kubernetes集群**: 1.20+
2. **存储类**: 需要支持ReadWriteMany的StorageClass（如NFS、Ceph）
3. **镜像仓库**: 用于存储自定义镜像
4. **资源要求**: 至少8C16G可用资源

### 一键部署

```bash
cd k8s/

# 1. 构建并推送镜像 (首次部署)
./deploy.sh build

# 2. 部署到K8s集群
./deploy.sh deploy

# 3. 查看部署状态
./deploy.sh status
```

## 📊 详细部署步骤

### Step 1: 准备镜像

**方式1: 本地构建推送**
```bash
# 修改镜像仓库地址
export REGISTRY="d1ng404"
export VERSION="v1.0.0"

# 构建业务服务镜像
docker-compose build

# 标记并推送镜像
docker tag mock-s3/metadata-service:latest ${REGISTRY}/metadata-service:${VERSION}
docker push ${REGISTRY}/metadata-service:${VERSION}

docker tag mock-s3/storage-service:latest ${REGISTRY}/storage-service:${VERSION}
docker push ${REGISTRY}/storage-service:${VERSION}

docker tag mock-s3/queue-service:latest ${REGISTRY}/queue-service:${VERSION}
docker push ${REGISTRY}/queue-service:${VERSION}

docker tag mock-s3/third-party-service:latest ${REGISTRY}/third-party-service:${VERSION}
docker push ${REGISTRY}/third-party-service:${VERSION}
```

**方式2: 使用脚本自动构建**
```bash
# 修改deploy.sh中的REGISTRY变量
vim deploy.sh
# 设置: REGISTRY="your-registry.com"

./deploy.sh build
```

### Step 2: 配置调整

**镜像地址更新**
```bash
# 更新业务服务配置中的镜像地址
sed -i 's|your-registry.com|实际仓库地址|g' 06-business-services.yaml
```

**存储类配置** (根据集群情况调整)
```yaml
# 03-storage.yaml
spec:
  storageClassName: nfs-storage  # 改为实际的StorageClass名称
```

**资源限制调整** (可选)
```yaml
# 06-business-services.yaml
resources:
  requests:
    memory: "256Mi"  # 根据实际需要调整
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Step 3: 分步部署

```bash
# 1. 创建命名空间
kubectl apply -f 01-namespace.yaml

# 2. 创建配置和密钥
kubectl apply -f 02-configmaps.yaml

# 3. 创建持久化存储
kubectl apply -f 03-storage.yaml

# 等待PVC绑定
kubectl get pvc -n mock-s3

# 4. 部署基础设施服务
kubectl apply -f 04-infrastructure.yaml

# 等待基础设施就绪
kubectl wait --namespace=mock-s3 --for=condition=ready pod -l app=postgres --timeout=300s
kubectl wait --namespace=mock-s3 --for=condition=ready pod -l app=redis --timeout=300s
kubectl wait --namespace=mock-s3 --for=condition=ready pod -l app=consul --timeout=300s

# 5. 部署监控组件
kubectl apply -f 05-monitoring.yaml

# 等待监控组件就绪
kubectl wait --namespace=mock-s3 --for=condition=ready pod -l app=prometheus --timeout=300s

# 6. 部署业务服务
kubectl apply -f 06-business-services.yaml

# 7. 创建访问入口
kubectl apply -f 07-ingress.yaml
```

## 🔧 服务管理

### 🗂️ 存储架构说明 - 方案1实现

**共享存储 + 3倍冗余设计：**

**存储拓扑：**
```
Pod-1: storage-service-xxx-abc ─┐
Pod-2: storage-service-xxx-def ─┼── 共享PVC (/app/data/storage)
Pod-3: storage-service-xxx-ghi ─┘
                                │
                                ▼
                        /app/data/storage/
                        ├── replica1/ (冗余副本1)
                        ├── replica2/ (冗余副本2)
                        └── replica3/ (冗余副本3)
```

**一致性保证：**
- ✅ **写入**: 每个文件同时写入3个replica目录
- ✅ **读取**: 从任一replica目录读取（故障转移）
- ✅ **一致性**: 所有Pod操作相同的存储路径
- ✅ **冗余**: 保持原有的3倍数据冗余设计

**配置来源：**
- 配置文件通过ConfigMap动态挂载到 `/app/config/storage-config.yaml`
- 应用启动时读取配置，确定存储节点路径

### 扩缩容操作

```bash
# 使用脚本快速扩容到5个副本
./deploy.sh scale 5

# 手动扩缩容指定服务
kubectl scale deployment metadata-service -n mock-s3 --replicas=3
kubectl scale deployment storage-service -n mock-s3 --replicas=3
kubectl scale deployment queue-service -n mock-s3 --replicas=2
kubectl scale deployment third-party-service -n mock-s3 --replicas=2

# 查看扩容结果
kubectl get pods -n mock-s3 -l app=metadata-service
```

### 滚动更新

```bash
# 更新镜像版本
kubectl set image deployment/metadata-service metadata-service=your-registry.com/mock-s3/metadata-service:v1.1.0 -n mock-s3

# 查看更新状态
kubectl rollout status deployment/metadata-service -n mock-s3

# 回滚到上一版本
kubectl rollout undo deployment/metadata-service -n mock-s3
```

### 服务重启

```bash
# 重启指定服务的所有Pod
kubectl rollout restart deployment/metadata-service -n mock-s3
kubectl rollout restart deployment/storage-service -n mock-s3
kubectl rollout restart deployment/queue-service -n mock-s3
kubectl rollout restart deployment/third-party-service -n mock-s3
```

## 🌐 访问服务

### NodePort访问 (推荐)

服务通过NodePort直接访问，无需域名配置：

| 服务 | 访问地址 | 说明 |
|------|----------|------|
| Metadata Service | `http://节点IP:30081` | 元数据管理服务 |
| Storage Service | `http://节点IP:30082` | 存储服务 |
| Queue Service | `http://节点IP:30083` | 队列服务 |
| Third-Party Service | `http://节点IP:30084` | 第三方集成服务 |
| Grafana | `http://节点IP:30300` | 监控面板 (admin/admin) |
| Prometheus | `http://节点IP:30900` | 指标查询 |
| Consul UI | `http://节点IP:30500` | 服务发现管理 |

### Ingress访问

如果集群配置了Ingress Controller，可通过域名访问：

```bash
# 配置本地hosts文件 (/etc/hosts)
<节点IP> mock-s3-metadata.local
<节点IP> mock-s3-storage.local  
<节点IP> mock-s3-queue.local
<节点IP> mock-s3-thirdparty.local
<节点IP> mock-s3-grafana.local
<节点IP> mock-s3-prometheus.local
<节点IP> mock-s3-consul.local

# 通过域名访问
curl http://mock-s3-metadata.local/health
```

### KubeSphere访问

在KubeSphere控制台中：

1. **项目管理** → **mock-s3项目** → **工作负载**
2. **应用路由** → 查看Ingress配置  
3. **存储管理** → 查看PVC使用情况
4. **监控告警** → 查看资源使用监控

## 📊 监控验证

### 健康检查

```bash
# 检查所有Pod状态
kubectl get pods -n mock-s3

# 检查服务端点
kubectl get svc -n mock-s3

# 检查Ingress状态
kubectl get ingress -n mock-s3

# 查看Pod日志
kubectl logs -f deployment/metadata-service -n mock-s3
```

### 功能测试

**1. 服务健康检查**
```bash
# 测试各服务健康状态
curl http://节点IP:30081/health  # metadata-service
curl http://节点IP:30082/health  # storage-service  
curl http://节点IP:30083/health  # queue-service
curl http://节点IP:30084/health  # third-party-service
```

**2. 故障注入测试**
```bash
# 创建CPU峰值异常
curl -X POST http://节点IP:30085/api/v1/metric-anomaly \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "K8s CPU压力测试",
    "service": "storage-service", 
    "metric_name": "system_cpu_usage_percent",
    "anomaly_type": "cpu_spike",
    "target_value": 85.0,
    "duration": 120000000000,
    "enabled": true
  }'
```

**3. 监控面板访问**
- **Grafana**: http://节点IP:30300 (admin/admin)
- **Prometheus**: http://节点IP:30900
- **Consul**: http://节点IP:30500

## 🛠 故障排查

### 常见问题

**1. Pod启动失败**
```bash
# 查看Pod事件
kubectl describe pod <pod-name> -n mock-s3

# 查看Pod日志
kubectl logs <pod-name> -n mock-s3

# 常见原因: 镜像拉取失败、PVC绑定失败、资源不足
```

**2. 服务连接失败**
```bash
# 检查Service配置
kubectl get svc -n mock-s3

# 检查端口转发
kubectl port-forward svc/metadata-service 8081:8081 -n mock-s3

# 检查网络策略
kubectl get networkpolicy -n mock-s3
```

**3. 存储问题**
```bash  
# 查看PVC状态
kubectl get pvc -n mock-s3

# 查看StorageClass
kubectl get storageclass

# 如果使用ReadWriteMany，确保存储类支持
```

### 日志收集

```bash
# 收集所有服务日志
kubectl logs -l app=metadata-service -n mock-s3 --tail=100
kubectl logs -l app=storage-service -n mock-s3 --tail=100
kubectl logs -l app=queue-service -n mock-s3 --tail=100
kubectl logs -l app=third-party-service -n mock-s3 --tail=100

# 收集基础设施日志
kubectl logs -l app=postgres -n mock-s3 --tail=50
kubectl logs -l app=redis -n mock-s3 --tail=50
kubectl logs -l app=consul -n mock-s3 --tail=50
```

## 🧹 清理资源

### 完整清理

```bash
# 使用脚本清理（推荐）
./deploy.sh clean

# 手动清理
kubectl delete namespace mock-s3

# 清理持久化数据（慎重！）
kubectl delete pvc --all -n mock-s3
```

### 部分清理

```bash
# 只删除业务服务，保留基础设施
kubectl delete -f 06-business-services.yaml
kubectl delete -f 07-ingress.yaml

# 重新部署业务服务
kubectl apply -f 06-business-services.yaml
kubectl apply -f 07-ingress.yaml
```

## 📈 性能优化

### 资源调优

```yaml
# 根据实际负载调整资源配置
resources:
  requests:
    memory: "512Mi"  # 增加内存请求
    cpu: "200m"      # 增加CPU请求
  limits:
    memory: "2Gi"    # 增加内存限制
    cpu: "1000m"     # 增加CPU限制
```

### 存储优化

```yaml
# 使用高性能存储类
storageClassName: ssd-storage

# 增加存储空间
resources:
  requests:
    storage: 100Gi
```

### 网络优化

```yaml
# 启用Service网格加速
metadata:
  annotations:
    service.beta.kubernetes.io/external-traffic: OnlyLocal
```

## 📞 支持与维护

- **项目仓库**: [MockS3 GitHub](https://github.com/your-org/mock-s3)
- **问题反馈**: 通过GitHub Issues提交
- **文档更新**: 随版本更新维护

---

📝 **部署检查清单**

- [ ] Kubernetes集群就绪 (1.20+)
- [ ] 镜像仓库配置完成
- [ ] 存储类支持ReadWriteMany
- [ ] 镜像构建并推送成功
- [ ] 所有Pod处于Running状态
- [ ] 服务健康检查通过
- [ ] 监控面板可正常访问
- [ ] 故障注入功能验证成功