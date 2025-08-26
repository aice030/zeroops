# 使用官方nginx基础镜像
FROM nginx:1.25-alpine

# 安装必要的工具和consul-template
RUN apk add --no-cache \
    curl \
    wget \
    bash \
    jq \
    unzip \
    && rm -rf /var/cache/apk/*

# 安装consul-template
ENV CONSUL_TEMPLATE_VERSION=0.33.0
RUN wget -q https://releases.hashicorp.com/consul-template/${CONSUL_TEMPLATE_VERSION}/consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip \
    && unzip consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip \
    && mv consul-template /usr/local/bin/consul-template \
    && chmod +x /usr/local/bin/consul-template \
    && rm consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip

# 设置工作目录
WORKDIR /etc/nginx

# 创建必要的目录
RUN mkdir -p /etc/nginx/conf.d \
    && mkdir -p /etc/nginx/templates \
    && mkdir -p /var/log/nginx \
    && mkdir -p /usr/local/bin

# 复制nginx配置文件
COPY nginx.conf /etc/nginx/nginx.conf

# 复制启动脚本和模板
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY templates/ /etc/nginx/templates/

# 设置脚本执行权限
RUN chmod +x /usr/local/bin/entrypoint.sh

# 创建nginx用户和组(如果不存在)
RUN addgroup -g 101 -S nginx || true \
    && adduser -S -D -H -u 101 -h /var/cache/nginx -s /sbin/nologin -G nginx -g nginx nginx || true

# 设置权限
RUN chown -R nginx:nginx /var/cache/nginx \
    && chown -R nginx:nginx /var/log/nginx \
    && chown -R nginx:nginx /etc/nginx/conf.d \
    && touch /var/run/nginx.pid \
    && chown nginx:nginx /var/run/nginx.pid

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV NGINX_ENVSUBST_OUTPUT_DIR=/etc/nginx/conf.d
ENV NGINX_WORKER_PROCESSES=auto
ENV NGINX_WORKER_CONNECTIONS=1024
ENV CONSUL_HTTP_ADDR=consul:8500
ENV CONSUL_TEMPLATE_LOG=INFO

# 使用自定义启动脚本
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# 默认命令
CMD ["nginx", "-g", "daemon off;"]