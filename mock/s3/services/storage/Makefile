.PHONY: build run clean test docker-build docker-run help

# 变量定义
SERVICE_NAME = storage-service
DOCKER_IMAGE = mocks3/$(SERVICE_NAME)
CONFIG_PATH = config/storage-config.yaml

# 默认目标
help:
	@echo "Available commands:"
	@echo "  build        - Build the service binary"
	@echo "  run          - Run the service locally"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  help         - Show this help message"

# 构建二进制文件
build:
	@echo "Building $(SERVICE_NAME)..."
	@cd ../.. && go build -o bin/$(SERVICE_NAME) ./services/storage/cmd
	@echo "Build completed: bin/$(SERVICE_NAME)"

# 运行服务
run: build
	@echo "Starting $(SERVICE_NAME)..."
	@cd ../.. && ./bin/$(SERVICE_NAME)

# 清理
clean:
	@echo "Cleaning build artifacts..."
	@rm -f ../../bin/$(SERVICE_NAME)
	@echo "Clean completed"

# 运行测试
test:
	@echo "Running tests..."
	@cd ../.. && go test ./services/storage/...
	@echo "Tests completed"

# 构建Docker镜像
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	@cd ../.. && docker build -f services/storage/Dockerfile -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

# 运行Docker容器
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run --rm -p 8082:8082 \
		--name $(SERVICE_NAME) \
		-e SERVICE_PORT=8082 \
		-e LOG_LEVEL=info \
		$(DOCKER_IMAGE)

# 开发环境运行
dev-run:
	@echo "Starting $(SERVICE_NAME) in development mode..."
	@cd ../.. && go run ./services/storage/cmd

# 检查配置文件
check-config:
	@echo "Checking configuration file..."
	@if [ -f $(CONFIG_PATH) ]; then \
		echo "✓ Configuration file exists: $(CONFIG_PATH)"; \
	else \
		echo "✗ Configuration file not found: $(CONFIG_PATH)"; \
		exit 1; \
	fi

# 格式化代码
fmt:
	@echo "Formatting code..."
	@cd ../.. && go fmt ./services/storage/...
	@echo "Code formatted"

# 静态检查
lint:
	@echo "Running static analysis..."
	@cd ../.. && go vet ./services/storage/...
	@echo "Static analysis completed"