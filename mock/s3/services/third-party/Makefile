.PHONY: build run test clean docker-build docker-run help

# 应用信息
APP_NAME := third-party-service
VERSION := v1.0.0
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# 构建参数
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
BINARY_PATH := bin/$(APP_NAME)

# Go相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

# Docker相关变量
DOCKER_IMAGE := mocks3/$(APP_NAME)
DOCKER_TAG := $(VERSION)

help: ## 显示帮助信息
	@echo "Third-Party Service 构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## 安装依赖
	$(GOMOD) download
	$(GOMOD) tidy

build: deps ## 构建应用
	@echo "构建 $(APP_NAME)..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/main.go
	@echo "✅ 构建完成: $(BINARY_PATH)"

run: build ## 运行应用
	@echo "启动 $(APP_NAME)..."
	./$(BINARY_PATH)

test: ## 运行测试
	@echo "运行测试..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "✅ 测试完成"

fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GOFMT) ./...
	@echo "✅ 代码格式化完成"

lint: ## 运行代码检查
	@echo "运行代码检查..."
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint" && exit 1)
	golangci-lint run ./...
	@echo "✅ 代码检查完成"

clean: ## 清理构建文件
	@echo "清理构建文件..."
	rm -rf bin/
	rm -f coverage.out
	@echo "✅ 清理完成"

docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .
	@echo "✅ Docker镜像构建完成"

docker-run: docker-build ## 运行Docker容器
	@echo "运行Docker容器..."
	docker run -p 8084:8084 --name $(APP_NAME) --rm $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	-docker stop $(APP_NAME)
	@echo "✅ Docker容器已停止"

# 多平台构建
build-all: ## 构建所有平台版本
	@echo "构建多平台版本..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 ./cmd/main.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 ./cmd/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/main.go
	@echo "✅ 多平台构建完成"

dev: ## 开发模式（带实时重载）
	@echo "启动开发模式..."
	@which air > /dev/null || (echo "请先安装 air: go install github.com/cosmtrek/air@latest" && exit 1)
	air -c .air.toml

# 健康检查
health: ## 检查服务健康状态
	@echo "检查服务健康状态..."
	curl -f http://localhost:8084/health || exit 1
	@echo "✅ 服务健康"

stats: ## 获取服务统计信息
	@echo "获取服务统计信息..."
	curl -s http://localhost:8084/api/v1/stats | jq '.' || curl -s http://localhost:8084/api/v1/stats

install: build ## 安装到系统路径
	@echo "安装 $(APP_NAME) 到系统路径..."
	sudo cp $(BINARY_PATH) /usr/local/bin/
	@echo "✅ 安装完成"