package routes

import (
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/handlers"
	"github.com/gin-gonic/gin"
)

// SetupLocationRoutes 
func SetupLocationRoutes(router *gin.RouterGroup, handler *handlers.LocationHandler) {
	// ?
	router.GET("/location/health", handler.HealthCheck)

	//  - v1
	trajectories := router.Group("/trajectories")
	{
		trajectories.POST("", handler.CreateTrajectory)           // 
		trajectories.GET("", handler.GetTrajectories)            // 
		trajectories.GET("/stats", handler.GetTrajectoryStats)   // 
		trajectories.GET("/:id", handler.GetTrajectory)          // 
		trajectories.PUT("/:id", handler.UpdateTrajectory)       // 
		trajectories.DELETE("/:id", handler.DeleteTrajectory)    // 
	}

	// ?
	locationPoints := router.Group("/location-points")
	{
		locationPoints.POST("", handler.AddLocationPoint)        // ?
		locationPoints.POST("/batch", handler.AddLocationPointsBatch) // ?
		locationPoints.GET("", handler.GetLocationPoints)        // ?
	}

	// 
	router.POST("/sync", handler.SyncData)                      // 
}

