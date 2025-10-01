package main

import (
	"log"
	"os"

	"github.com/taishanglaojun/project-scaffold/internal/config"
	"github.com/taishanglaojun/project-scaffold/internal/logger"
	"github.com/taishanglaojun/project-scaffold/internal/server"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	defer logger.Sync()

	// 创建服务器
	srv := server.New(cfg, logger)

	// 启动服务器
	logger.Info("Starting server", 
		"port", cfg.Server.Port,
		"env", cfg.App.Environment,
	)

	if err := srv.Start(); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}