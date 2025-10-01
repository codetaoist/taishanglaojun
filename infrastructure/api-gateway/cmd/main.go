package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/config"
	"api-gateway/internal/gateway"
	"api-gateway/internal/logger"
	"api-gateway/internal/monitoring"
	"api-gateway/internal/proxy"
	"api-gateway/internal/registry"
	
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	log := logger.New(cfg.Log)

	// 初始化监控
	metrics := monitoring.New(cfg.Monitoring)

	// 初始化服务注册中心
	serviceRegistry := registry.New(cfg.Registry, log)

	// 初始化代理管理器
	proxyManager := proxy.New(cfg.Proxy, serviceRegistry, metrics, log)

	// 初始化网关
	gw, err := gateway.NewGateway(cfg, log, metrics, serviceRegistry, proxyManager)
	if err != nil {
		logrus.Fatalf("Failed to create gateway: %v", err)
	}

	// 创建路由器
	r := gw.GetRouter()

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		log.Infof("Starting API Gateway on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 启动监控服务器
	if cfg.Monitoring.Enabled {
		go func() {
			log.Infof("Starting monitoring server on port %d", cfg.Monitoring.Port)
			if err := metrics.StartServer(cfg.Monitoring.Port); err != nil {
				log.Errorf("Failed to start monitoring server: %v", err)
			}
		}()
	}

	// 启动健康检查
	go func() {
		ticker := time.NewTicker(cfg.HealthCheck.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				proxyManager.HealthCheck()
			}
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down API Gateway...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	// 关闭其他组件
	proxyManager.Close()
	serviceRegistry.Close()

	log.Info("API Gateway stopped")
}