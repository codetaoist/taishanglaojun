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
	"github.com/taishang/sequence-zero/internal/auth"
	"github.com/taishang/sequence-zero/pkg/database"
)

func main() {
	// 初始化数据库连接
	db, err := database.NewConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 初始化认证服务
	authService := auth.NewService(db)

	// 设置Gin路由
	r := gin.Default()

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "taishang-auth-service"})
	})

	// 认证相关路由
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", authService.Login)
		auth.POST("/logout", authService.Logout)
		auth.POST("/refresh", authService.RefreshToken)
		auth.GET("/verify", authService.VerifyToken)
		auth.POST("/register", authService.Register)
	}

	// 权限验证中间件保护的路由
	protected := r.Group("/api/v1/protected")
	protected.Use(authService.AuthMiddleware())
	{
		protected.GET("/profile", authService.GetProfile)
		protected.PUT("/profile", authService.UpdateProfile)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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

	fmt.Printf("太上老君认证服务启动在端口 %s\n", port)

	// 等待中断信号来优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	fmt.Println("服务器已退出")
}