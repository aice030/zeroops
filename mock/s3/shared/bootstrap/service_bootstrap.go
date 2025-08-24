package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"mocks3/shared/middleware"
	"mocks3/shared/observability"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	ServiceName             string
	ServicePort             int
	Version                 string
	Environment             string
	LogLevel                string
	ConsulAddr              string
	ObservabilityConfigPath string // 可观测性配置文件路径
}

// ServiceBootstrap 服务启动器
type ServiceBootstrap struct {
	config         ServiceConfig
	providers      *observability.Providers
	collector      *observability.MetricCollector
	httpMiddleware *observability.HTTPMiddleware
	consulManager  *middleware.ConsulManager
	router         *gin.Engine
	server         *http.Server
}

// NewServiceBootstrap 创建服务启动器
func NewServiceBootstrap(config ServiceConfig) *ServiceBootstrap {
	return &ServiceBootstrap{
		config: config,
	}
}

// Initialize 初始化所有组件
func (sb *ServiceBootstrap) Initialize() error {
	// 初始化可观测性组件
	providers, collector, httpMiddleware, err := observability.Setup(sb.config.ServiceName, sb.config.ObservabilityConfigPath)
	if err != nil {
		return fmt.Errorf("failed to initialize observability: %w", err)
	}
	sb.providers = providers
	sb.collector = collector
	sb.httpMiddleware = httpMiddleware

	// 初始化 Consul 配置
	if err := middleware.InitializeServiceConfig(sb.config.ConsulAddr, sb.config.ServiceName, sb.config.ServicePort); err != nil {
		log.Printf("Failed to initialize consul config: %v", err)
	}

	// 初始化Consul管理器
	consulManager, err := middleware.NewDefaultConsulManager(sb.config.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to initialize consul: %w", err)
	}
	sb.consulManager = consulManager

	return nil
}

// SetupRouter 设置路由器
func (sb *ServiceBootstrap) SetupRouter() *gin.Engine {
	// 设置Gin模式
	if sb.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由器
	sb.router = gin.New()

	// 添加基础中间件
	sb.router.Use(gin.Logger())
	sb.router.Use(middleware.GinRecoveryMiddleware(middleware.DefaultRecoveryConfig()))

	// 添加可观测性中间件
	observability.SetupGinMiddlewares(sb.router, sb.config.ServiceName, sb.httpMiddleware)

	// 添加健康检查路由
	sb.router.GET("/health", sb.healthCheckHandler)

	return sb.router
}

// RegisterService 注册服务到Consul
func (sb *ServiceBootstrap) RegisterService(ctx context.Context) error {
	// 从 Consul KV 加载配置
	consulConfig, err := middleware.LoadServiceConfigFromConsul(sb.config.ConsulAddr, sb.config.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to load consul config: %w", err)
	}

	// 注册服务
	if err := sb.consulManager.RegisterService(ctx, consulConfig); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 创建HTTP服务器
	addr := fmt.Sprintf(":%d", consulConfig.ServicePort)
	sb.server = &http.Server{
		Addr:         addr,
		Handler:      sb.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

// Start 启动服务
func (sb *ServiceBootstrap) Start() error {
	if sb.server == nil {
		return fmt.Errorf("server not initialized, call RegisterService first")
	}

	// 启动系统指标收集
	ctx := context.Background()
	observability.StartSystemMetrics(ctx, sb.collector, sb.providers.Logger)

	// 启动服务器
	go func() {
		sb.providers.Logger.Info(context.Background(), "Starting service",
			observability.String("service", sb.config.ServiceName),
			observability.String("address", sb.server.Addr))
		if err := sb.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return nil
}

// WaitForShutdown 等待关闭信号并优雅关闭
func (sb *ServiceBootstrap) WaitForShutdown() {
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sb.providers.Logger.Info(context.Background(), "Shutting down service",
		observability.String("service", sb.config.ServiceName))

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 注销服务
	if sb.consulManager != nil {
		if err := sb.consulManager.DeregisterService(ctx); err != nil {
			sb.providers.Logger.ErrorWithErr(ctx, err, "Failed to deregister service")
		}
	}

	// 关闭HTTP服务器
	if sb.server != nil {
		if err := sb.server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}

	sb.providers.Logger.Info(context.Background(), "Service stopped",
		observability.String("service", sb.config.ServiceName))
}

// Shutdown 手动关闭所有资源
func (sb *ServiceBootstrap) Shutdown(ctx context.Context) {
	// 关闭可观测性资源
	if err := observability.Shutdown(ctx, sb.providers); err != nil {
		log.Printf("Failed to shutdown observability: %v", err)
	}
}

// GetLogger 获取日志器实例
func (sb *ServiceBootstrap) GetLogger() *observability.Logger {
	return sb.providers.Logger
}

// GetRouter 获取路由器实例
func (sb *ServiceBootstrap) GetRouter() *gin.Engine {
	return sb.router
}

// healthCheckHandler 健康检查处理器
func (sb *ServiceBootstrap) healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   sb.config.ServiceName,
		"version":   sb.config.Version,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ServiceInitializer 服务初始化接口
type ServiceInitializer interface {
	Initialize(bootstrap *ServiceBootstrap) error
}

// RunService 运行服务的便利函数
func RunService(config ServiceConfig, initializer ServiceInitializer) {
	// 创建启动器
	bootstrap := NewServiceBootstrap(config)

	// 初始化基础组件
	if err := bootstrap.Initialize(); err != nil {
		log.Fatalf("Failed to initialize bootstrap: %v", err)
	}

	// 设置路由器
	bootstrap.SetupRouter()

	// 让服务初始化自己的组件
	if err := initializer.Initialize(bootstrap); err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// 注册服务到Consul
	ctx := context.Background()
	if err := bootstrap.RegisterService(ctx); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}
	defer bootstrap.Shutdown(ctx)

	// 启动服务
	if err := bootstrap.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	// 等待关闭
	bootstrap.WaitForShutdown()
}
