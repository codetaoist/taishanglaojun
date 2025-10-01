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
	"github.com/taishang/sequence-zero/internal/permission"
	"github.com/taishang/sequence-zero/pkg/database"
)

func main() {
	// 初始化数据库连接
	db, err := database.NewConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 初始化权限服务
	permissionService := permission.NewService(db)

	// 设置Gin路由
	r := gin.Default()

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "taishang-permission-service"})
	})

	// Prometheus指标端点
	r.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "# HELP taishang_permission_requests_total Total number of permission requests\n# TYPE taishang_permission_requests_total counter\ntaishang_permission_requests_total 0\n")
	})

	// 权限管理路由
	api := r.Group("/api/v1")
	{
		// 权限检查
		api.POST("/check", permissionService.CheckPermission)
		api.POST("/batch-check", permissionService.BatchCheckPermissions)
		
		// 权限管理
		api.GET("/permissions", permissionService.ListPermissions)
		api.GET("/permissions/:id", permissionService.GetPermission)
		api.POST("/permissions", permissionService.CreatePermission)
		api.PUT("/permissions/:id", permissionService.UpdatePermission)
		api.DELETE("/permissions/:id", permissionService.DeletePermission)
		
		// 用户权限管理
		api.GET("/users/:id/permissions", permissionService.GetUserPermissions)
		api.POST("/users/:id/permissions", permissionService.GrantPermission)
		api.DELETE("/users/:id/permissions/:permission_id", permissionService.RevokePermission)
		
		// 权限等级管理
		api.GET("/users/:id/level", permissionService.GetUserLevel)
		api.PUT("/users/:id/level", permissionService.UpdateUserLevel)
		
		// 权限审计
		api.GET("/audit/permissions", permissionService.GetPermissionAudit)
		api.GET("/audit/users/:id", permissionService.GetUserAudit)
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
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

	fmt.Printf("太上老君权限服务启动在端口 %s\n", port)

	// 等待中断信号来优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("正在关闭权限服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("权限服务强制关闭:", err)
	}

	fmt.Println("权限服务已退出")
}