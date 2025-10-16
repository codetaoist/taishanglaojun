package main

import (
	"log"
	"os"

	httpServer "github.com/codetaoist/taishanglaojun/core-services/task-management/internal/interfaces/http"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 从环境变量获取配?
	config := httpServer.ConfigFromEnv()

	// 创建HTTP服务?
	server := httpServer.NewServer(config)

	// 启动服务器并支持优雅关闭
	if err := server.StartWithGracefulShutdown(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}

	log.Println("Server shutdown complete")
}

