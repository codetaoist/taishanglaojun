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
	"github.com/taishang/sequence-zero/internal/cultural"
	"github.com/taishang/sequence-zero/pkg/database"
)

func main() {
	// 初始化数据库连接
	db, err := database.NewConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 初始化文化智慧服务
	culturalService := cultural.NewService(db)

	// 设置Gin路由
	r := gin.Default()

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "taishang-cultural-service"})
	})

	// 文化智慧相关路由
	api := r.Group("/api/v1/cultural")
	{
		// 智慧问答
		api.POST("/wisdom/ask", culturalService.AskWisdom)
		api.GET("/wisdom/daily", culturalService.GetDailyWisdom)
		api.GET("/wisdom/category/:category", culturalService.GetWisdomByCategory)
		
		// 文化知识
		api.GET("/knowledge/search", culturalService.SearchKnowledge)
		api.GET("/knowledge/classics", culturalService.GetClassics)
		api.GET("/knowledge/philosophy", culturalService.GetPhilosophy)
		
		// 个人修养
		api.POST("/cultivation/plan", culturalService.CreateCultivationPlan)
		api.GET("/cultivation/progress/:userId", culturalService.GetCultivationProgress)
		api.POST("/cultivation/practice", culturalService.RecordPractice)
		
		// 文化传承
		api.GET("/heritage/stories", culturalService.GetHeritageStories)
		api.GET("/heritage/traditions", culturalService.GetTraditions)
		api.POST("/heritage/share", culturalService.ShareCulturalExperience)
	}

	// 启动服务器
	port := os.Getenv("CULTURAL_SERVICE_PORT")
	if port == "" {
		port = "8083"
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

	fmt.Printf("太上序列零 - 文化智慧服务启动在端口 %s\n", port)

	// 等待中断信号来优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("正在关闭文化智慧服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("文化智慧服务强制关闭:", err)
	}

	fmt.Println("文化智慧服务已退出")
}