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
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/infrastructure/mock"
)

// @title  API (Mock汾)
// @version 1.0
// @description API
// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Gin
	gin.SetMode(gin.ReleaseMode)
	
	// Gin
	r := gin.Default()

	// CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// API
	api := r.Group("/api/v1")
	{
		// API
		learners := api.Group("/learners")
		{
			learners.GET("/profile", getLearnerProfile)
			learners.GET("/analytics", getLearningAnalytics)
		}

		// API
		learning := api.Group("/learning")
		{
			learning.GET("/weekly-activity", getWeeklyActivity)
			learning.GET("/skill-progress", getSkillProgress)
			learning.GET("/recommendations", getRecommendations)
			learning.GET("/activities", getActivities)
			learning.GET("/achievements", getAchievements)
		}
	}

	// ?
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "intelligent-learning-mock",
		})
	})

	// ?
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// goroutine
	go func() {
		log.Printf(" (Mock汾) ?8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("? %v", err)
		}
	}()

	// ?
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("?..")

	// 5?
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("?", err)
	}

	log.Println("?)
}

// getLearnerProfile ?
func getLearnerProfile(c *gin.Context) {
	profile := mock.GetMockLearnerProfile()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// getLearningAnalytics 
func getLearningAnalytics(c *gin.Context) {
	analytics := mock.GetMockLearningAnalytics()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// getWeeklyActivity ?
func getWeeklyActivity(c *gin.Context) {
	activities := mock.GetMockWeeklyActivity()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

// getSkillProgress ?
func getSkillProgress(c *gin.Context) {
	skills := mock.GetMockSkillProgress()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    skills,
	})
}

// getRecommendations 
func getRecommendations(c *gin.Context) {
	recommendations := mock.GetMockRecommendations()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recommendations,
	})
}

// getActivities 
func getActivities(c *gin.Context) {
	activities := mock.GetMockActivities()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

// getAchievements 
func getAchievements(c *gin.Context) {
	achievements := mock.GetMockAchievements()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    achievements,
	})
}

