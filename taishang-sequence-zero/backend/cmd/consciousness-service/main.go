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
	"github.com/taishang/sequence-zero/internal/consciousness"
	"github.com/taishang/sequence-zero/pkg/database"
)

func main() {
	// 初始化数据库连接
	db, err := database.NewConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 初始化意识融合服务
	consciousnessService := consciousness.NewService(db)

	// 设置Gin路由
	r := gin.Default()

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "taishang-consciousness-service"})
	})

	// 意识融合相关路由
	api := r.Group("/api/v1/consciousness")
	{
		api.POST("/analyze", consciousnessService.AnalyzeConsciousness)
		api.POST("/fuse", consciousnessService.FuseConsciousness)
		api.GET("/state/:userId", consciousnessService.GetConsciousnessState)
		api.POST("/adapt", consciousnessService.AdaptPersonality)
		api.GET("/history/:userId", consciousnessService.GetFusionHistory)
	}

	// 启动服务器
	port := os.Getenv("CONSCIOUSNESS_SERVICE_PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	fmt.Printf("太上序列零 - 意识融合服务启动在端口 %s\n", port)

	// 等待中断信号来优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("正在关闭意识融合服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("意识融合服务强制关闭:", err)
	}

	fmt.Println("意识融合服务已退出")
}