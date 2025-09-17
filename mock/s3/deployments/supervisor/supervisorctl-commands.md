# Supervisor 常用命令

## 部署和使用

### 1. 上传配置文件到服务器
```bash
# 从本机上传
scp supervisor/mock-s3.conf qboxserver@jfcs1021:/tmp/
scp supervisor/deploy-supervisor.sh qboxserver@jfcs1021:/tmp/
```

### 2. 在服务器上部署
```bash
# 安装supervisor（如未安装）
sudo apt-get update && sudo apt-get install -y supervisor
# 或
sudo yum install -y supervisor

# 复制配置文件
sudo cp /tmp/mock-s3.conf /etc/supervisor/conf.d/

# 创建服务目录并部署
cd /tmp/dingnanjia
sudo ./deploy-supervisor.sh deploy
```

### 3. Supervisor 管理命令

#### 基础命令
```bash
# 重新加载配置
supervisorctl reread
supervisorctl update

# 查看所有服务状态
supervisorctl status

# 查看mock-s3组服务
supervisorctl status mock-s3:*
```

#### 启动/停止服务
```bash
# 启动所有mock-s3服务
supervisorctl start mock-s3:*

# 停止所有服务
supervisorctl stop mock-s3:*

# 重启所有服务
supervisorctl restart mock-s3:*

# 操作单个服务
supervisorctl start metadata-service-1
supervisorctl stop storage-service-2
supervisorctl restart queue-service-1
```

#### 分组操作
```bash
# 只启动metadata服务组
supervisorctl start metadata-services:*

# 只重启storage服务组
supervisorctl restart storage-services:*

# 查看特定组状态
supervisorctl status metadata-services:*
```

#### 查看日志
```bash
# 查看服务日志
supervisorctl tail metadata-service-1
supervisorctl tail -f metadata-service-1  # 实时查看

# 查看错误日志
supervisorctl tail metadata-service-1 stderr

# 直接查看日志文件
tail -f /home/qboxserver/metadata-service_1/logs/supervisor.log
tail -f /home/qboxserver/metadata-service_1/logs/supervisor.error.log
```

### 4. 服务端口分配

| 服务 | 实例 | 端口 | 健康检查 |
|-----|------|------|---------|
| metadata-service | 1 | 8181 | http://localhost:8181/health |
| metadata-service | 2 | 8182 | http://localhost:8182/health |
| metadata-service | 3 | 8183 | http://localhost:8183/health |
| storage-service | 1 | 8191 | http://localhost:8191/health |
| storage-service | 2 | 8192 | http://localhost:8192/health |
| queue-service | 1 | 8201 | http://localhost:8201/health |
| queue-service | 2 | 8202 | http://localhost:8202/health |
| third-party-service | 1 | 8211 | http://localhost:8211/health |
| mock-error-service | 1 | 8221 | http://localhost:8221/health |

### 5. 故障排查

#### 服务无法启动
```bash
# 查看详细错误
supervisorctl tail -1000 metadata-service-1 stderr

# 检查配置文件语法
supervisord -c /etc/supervisor/supervisord.conf -n

# 手动测试启动命令
su - qboxserver
cd /home/qboxserver/metadata-service_1/_package
./metadata-service --port=8181
```

#### 权限问题
```bash
# 确保目录权限正确
sudo chown -R qboxserver:qboxserver /home/qboxserver/metadata-service_*
sudo chown -R qboxserver:qboxserver /home/qboxserver/storage-service_*
# ... 其他服务
```

#### 端口冲突
```bash
# 检查端口占用
netstat -tlnp | grep -E '81[8-9][0-9]|82[0-3][0-9]'

# 修改配置中的端口
sudo vi /etc/supervisor/conf.d/mock-s3.conf
# 修改 command 中的 --port 参数
```

### 6. 监控和告警

#### 配置Web界面（可选）
```bash
# 编辑supervisor配置
sudo vi /etc/supervisor/supervisord.conf

# 添加或修改 [inet_http_server] 部分
[inet_http_server]
port = 0.0.0.0:9001
username = admin
password = admin123

# 重启supervisor
sudo systemctl restart supervisor

# 访问 http://server:9001
```

#### 集成监控
```bash
# 导出状态到Prometheus
supervisorctl status | awk '{print $1, $2}' > /var/lib/prometheus/node_exporter/supervisor.prom

# 使用crontab定期更新
*/1 * * * * supervisorctl status | awk '{print $1, $2}' > /var/lib/prometheus/node_exporter/supervisor.prom
```

### 7. 快速操作示例

```bash
# 完整部署流程
cd /tmp/dingnanjia
tar -xzf mock-s3-*.tar.gz
sudo cp supervisor/mock-s3.conf /etc/supervisor/conf.d/
sudo chmod +x deploy-supervisor.sh
sudo ./deploy-supervisor.sh deploy

# 重启所有服务
supervisorctl restart mock-s3:*

# 查看所有服务状态
supervisorctl status | grep mock-s3

# 批量健康检查
for port in 8181 8182 8183 8191 8192 8201 8202 8211 8221; do
    echo -n "Port $port: "
    curl -sf http://localhost:$port/health && echo "OK" || echo "FAIL"
done
```