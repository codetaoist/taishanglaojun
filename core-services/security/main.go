package main

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
	"github.com/codetaoist/taishanglaojun/core-services/security"
	"github.com/codetaoist/taishanglaojun/core-services/security/config"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建安全模块实例
	securityModule, err := security.NewSecurityModule(cfg)
	if err != nil {
		log.Fatalf("Failed to create security module: %v", err)
	}

	// 启动安全模块
	if err := securityModule.Start(); err != nil {
		log.Fatalf("Failed to start security module: %v", err)
	}

	// 创建HTTP服务�?
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      securityModule.Router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 启动服务�?
	go func() {
		log.Printf("Security module server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down security module...")

	// 创建关闭上下�?
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务�?
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// 停止安全模块
	if err := securityModule.Stop(); err != nil {
		log.Printf("Error stopping security module: %v", err)
	}

	log.Println("Security module stopped")
}
