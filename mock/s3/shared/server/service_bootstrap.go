package server

import (
	"context"
	"fmt"
	"mocks3/shared/observability"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"mocks3/shared/middleware/consul"
	"mocks3/shared/middleware/error_injection"
)

// ServiceConfig 服务配置接口
type ServiceConfig interface {
	GetServiceName() string
	GetHost() string
	GetPort() int
}

// ConsulServiceConfig 支持Consul的服务配置接口
type ConsulServiceConfig interface {
	ServiceConfig
	GetConsulAddress() string
}

// ServiceHandler 服务处理器接口
type ServiceHandler interface {
	SetupRoutes(router *gin.Engine)
}

// ServiceBootstrap 服务启动器
type ServiceBootstrap struct {
	ServiceName string
	Config      ServiceConfig
	Handler     ServiceHandler
	Logger      *observability.Logger

	// 可观测性组件
	Providers      *observability.Providers
	Collector      *observability.MetricCollector
	HTTPMiddleware *observability.HTTPMiddleware

	// 错误注入
	MetricInjector *error_injection.MetricInjector

	// Consul客户端
	ConsulClient consul.ConsulClient

	// 配置路径
	ObservabilityConfigPath  string
	MetricInjectorConfigPath string

	// 自定义初始化函数
	CustomInit func(ctx context.Context, logger *observability.Logger) error

	// 自定义清理函数
	CustomCleanup func(ctx context.Context, logger *observability.Logger) error
}

// NewServiceBootstrap 创建服务启动器
func NewServiceBootstrap(serviceName string, config ServiceConfig) *ServiceBootstrap {
	return &ServiceBootstrap{
		ServiceName: serviceName,
		Config:      config,
		// 设置默认配置路径
		ObservabilityConfigPath:  "observability.yaml",
		MetricInjectorConfigPath: "../config/metric-injector-config.yaml",
	}
}

// WithHandler 设置服务处理器
func (sb *ServiceBootstrap) WithHandler(handler ServiceHandler) *ServiceBootstrap {
	sb.Handler = handler
	return sb
}

// WithCustomInit 设置自定义初始化函数
func (sb *ServiceBootstrap) WithCustomInit(initFunc func(ctx context.Context, logger *observability.Logger) error) *ServiceBootstrap {
	sb.CustomInit = initFunc
	return sb
}

// WithCustomCleanup 设置自定义清理函数
func (sb *ServiceBootstrap) WithCustomCleanup(cleanupFunc func(ctx context.Context, logger *observability.Logger) error) *ServiceBootstrap {
	sb.CustomCleanup = cleanupFunc
	return sb
}

// WithObservabilityConfig 设置可观测性配置路径
func (sb *ServiceBootstrap) WithObservabilityConfig(configPath string) *ServiceBootstrap {
	sb.ObservabilityConfigPath = configPath
	return sb
}

// WithMetricInjectorConfig 设置错误注入配置路径
func (sb *ServiceBootstrap) WithMetricInjectorConfig(configPath string) *ServiceBootstrap {
	sb.MetricInjectorConfigPath = configPath
	return sb
}

// Start 启动服务
func (sb *ServiceBootstrap) Start() error {
	ctx := context.Background()

	// 1. 初始化可观测性组件
	if err := sb.setupObservability(); err != nil {
		return fmt.Errorf("failed to setup observability: %w", err)
	}
	defer observability.Shutdown(context.Background(), sb.Providers)

	sb.Logger.Info(ctx, fmt.Sprintf("Starting %s", sb.ServiceName),
		observability.String("service", sb.Config.GetServiceName()),
		observability.String("host", sb.Config.GetHost()),
		observability.Int("port", sb.Config.GetPort()))

	// 2. 初始化Consul服务注册
	if err := sb.setupConsulRegistration(); err != nil {
		sb.Logger.Warn(ctx, "Failed to setup Consul registration", observability.Error(err))
	}

	// 3. 初始化错误注入中间件
	if err := sb.setupErrorInjection(); err != nil {
		sb.Logger.Warn(ctx, "Failed to setup error injection", observability.Error(err))
	}

	// 4. 执行自定义初始化
	if sb.CustomInit != nil {
		if err := sb.CustomInit(ctx, sb.Logger); err != nil {
			return fmt.Errorf("custom initialization failed: %w", err)
		}
	}

	// 5. 设置HTTP服务器
	router := sb.setupRouter()

	// 6. 启动系统指标收集
	observability.StartSystemMetrics(ctx, sb.Collector, sb.Logger)

	// 7. 连接错误注入器到指标收集器
	sb.connectErrorInjection()

	// 8. 启动HTTP服务器
	server := sb.startHTTPServer(router)

	// 9. 等待关闭信号
	sb.waitForShutdown(server)

	return nil
}

// setupObservability 设置可观测性组件
func (sb *ServiceBootstrap) setupObservability() error {
	providers, collector, httpMiddleware, err := observability.Setup(
		sb.Config.GetServiceName(),
		sb.ObservabilityConfigPath,
	)
	if err != nil {
		return err
	}

	sb.Providers = providers
	sb.Collector = collector
	sb.HTTPMiddleware = httpMiddleware
	sb.Logger = observability.GetLogger(providers)

	return nil
}

// setupConsulRegistration 设置Consul服务注册
func (sb *ServiceBootstrap) setupConsulRegistration() error {
	ctx := context.Background()

	// 检查配置是否支持Consul
	consulConfig, ok := sb.Config.(ConsulServiceConfig)
	if !ok {
		sb.Logger.Warn(ctx, "Service config does not support Consul, skipping registration")
		return nil
	}

	// 创建Consul客户端
	consulClient, err := consul.CreateConsulClient(consulConfig.GetConsulAddress(), sb.Logger)
	if err != nil {
		return fmt.Errorf("failed to create Consul client: %w", err)
	}

	sb.ConsulClient = consulClient

	// 注册服务到Consul
	err = consul.RegisterService(ctx, consulClient,
		sb.Config.GetServiceName(),
		sb.Config.GetHost(),
		sb.Config.GetPort())
	if err != nil {
		return fmt.Errorf("failed to register service with Consul: %w", err)
	}

	sb.Logger.Info(ctx, "Service registered with Consul successfully",
		observability.String("consul_addr", consulConfig.GetConsulAddress()),
		observability.String("service_name", sb.Config.GetServiceName()))

	return nil
}

// setupErrorInjection 设置错误注入中间件
func (sb *ServiceBootstrap) setupErrorInjection() error {
	ctx := context.Background()

	// 尝试从配置文件加载
	metricInjector, err := error_injection.NewMetricInjector(
		sb.MetricInjectorConfigPath,
		sb.Config.GetServiceName(),
		sb.Logger,
	)

	if err != nil {
		sb.Logger.Warn(ctx, "Failed to load metric injector config, using defaults",
			observability.Error(err))
		// 使用默认配置创建
		sb.MetricInjector = error_injection.NewMetricInjectorWithDefaults(
			"http://localhost:8085",
			sb.Config.GetServiceName(),
			sb.Logger,
		)
	} else {
		sb.MetricInjector = metricInjector
	}

	if sb.MetricInjector != nil {
		sb.Logger.Info(ctx, "Metric injector initialized successfully")
	}

	return nil
}

// setupRouter 设置路由
func (sb *ServiceBootstrap) setupRouter() *gin.Engine {
	ctx := context.Background()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 使用标准可观测性中间件
	observability.SetupGinMiddlewares(router, sb.Config.GetServiceName(), sb.HTTPMiddleware)

	// 添加错误注入中间件
	if sb.MetricInjector != nil {
		httpMiddleware := sb.MetricInjector.HTTPMiddleware()
		router.Use(func(c *gin.Context) {
			httpMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Next()
			})).ServeHTTP(c.Writer, c.Request)
		})
		sb.Logger.Info(ctx, "Error injection middleware enabled")
	}

	// 设置业务路由
	if sb.Handler != nil {
		sb.Handler.SetupRoutes(router)
	}

	return router
}

// connectErrorInjection 连接错误注入器到指标收集器
func (sb *ServiceBootstrap) connectErrorInjection() {
	ctx := context.Background()

	if sb.MetricInjector != nil {
		metricCollector := observability.GetCollector(sb.Collector)
		if metricCollector != nil {
			metricCollector.SetMetricInjector(sb.MetricInjector)
			sb.Logger.Info(ctx, "Metric injector connected to metric collector")
		}
	}
}

// startHTTPServer 启动HTTP服务器
func (sb *ServiceBootstrap) startHTTPServer(router *gin.Engine) *http.Server {
	ctx := context.Background()

	addr := fmt.Sprintf("%s:%d", sb.Config.GetHost(), sb.Config.GetPort())
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 启动服务器
	go func() {
		sb.Logger.Info(ctx, "HTTP server starting", observability.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sb.Logger.Error(ctx, "HTTP server failed", observability.Error(err))
		}
	}()

	sb.Logger.Info(ctx, fmt.Sprintf("%s started successfully", sb.ServiceName),
		observability.String("addr", addr))

	return server
}

// waitForShutdown 等待关闭信号并优雅关闭
func (sb *ServiceBootstrap) waitForShutdown(server *http.Server) {
	ctx := context.Background()

	// 等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sb.Logger.Info(ctx, fmt.Sprintf("Shutting down %s...", sb.ServiceName))

	// 注销Consul服务
	sb.deregisterFromConsul()

	// 执行自定义清理
	if sb.CustomCleanup != nil {
		if err := sb.CustomCleanup(ctx, sb.Logger); err != nil {
			sb.Logger.Error(ctx, "Custom cleanup failed", observability.Error(err))
		}
	}

	// 清理错误注入器资源
	if sb.MetricInjector != nil {
		sb.MetricInjector.Cleanup()
		sb.Logger.Info(ctx, "Metric injector cleaned up")
	}

	// 关闭HTTP服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		sb.Logger.Error(ctx, "HTTP server shutdown failed", observability.Error(err))
	} else {
		sb.Logger.Info(ctx, "HTTP server stopped")
	}

	sb.Logger.Info(ctx, fmt.Sprintf("%s stopped", sb.ServiceName))
}

// deregisterFromConsul 从Consul注销服务
func (sb *ServiceBootstrap) deregisterFromConsul() {
	ctx := context.Background()

	if sb.ConsulClient == nil {
		return
	}

	// 生成服务ID (与注册时保持一致)
	serviceID := fmt.Sprintf("%s-%s-%d", 
		sb.Config.GetServiceName(), 
		sb.Config.GetHost(), 
		sb.Config.GetPort())

	if err := sb.ConsulClient.DeregisterService(ctx, serviceID); err != nil {
		sb.Logger.Error(ctx, "Failed to deregister service from Consul", observability.Error(err))
	} else {
		sb.Logger.Info(ctx, "Service deregistered from Consul successfully")
	}
}
