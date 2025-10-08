package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/infrastructure/mock"
)

// @title 智能学习系统 API (Mock版本)
// @version 1.0
// @description 智能学习系统的模拟API，用于前端开发和测试
// @host localhost:8080
// @BasePath /api/v1

func main() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	
	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// 设置API路由
	api := r.Group("/api/v1")
	{
		// 学习者相关API
		learners := api.Group("/learners")
		{
			learners.GET("/profile", getLearnerProfile)
			learners.GET("/analytics", getLearningAnalytics)
		}

		// 学习数据API
		learning := api.Group("/learning")
		{
			learning.GET("/weekly-activity", getWeeklyActivity)
			learning.GET("/skill-progress", getSkillProgress)
			learning.GET("/recommendations", getRecommendations)
			learning.GET("/activities", getActivities)
			learning.GET("/achievements", getAchievements)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "intelligent-learning-mock",
		})
	})

	// 启动服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		log.Printf("智能学习服务 (Mock版本) 启动在端口 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 5秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	log.Println("服务器已退出")
}

// getLearnerProfile 获取学习者档案
func getLearnerProfile(c *gin.Context) {
	profile := mock.GetMockLearnerProfile()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// getLearningAnalytics 获取学习分析数据
func getLearningAnalytics(c *gin.Context) {
	analytics := mock.GetMockLearningAnalytics()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// getWeeklyActivity 获取周活动数据
func getWeeklyActivity(c *gin.Context) {
	activities := mock.GetMockWeeklyActivity()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

// getSkillProgress 获取技能进度
func getSkillProgress(c *gin.Context) {
	skills := mock.GetMockSkillProgress()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    skills,
	})
}

// getRecommendations 获取推荐
func getRecommendations(c *gin.Context) {
	recommendations := mock.GetMockRecommendations()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recommendations,
	})
}

// getActivities 获取活动记录
func getActivities(c *gin.Context) {
	activities := mock.GetMockActivities()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

// getAchievements 获取成就
func getAchievements(c *gin.Context) {
	achievements := mock.GetMockAchievements()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    achievements,
	})
}