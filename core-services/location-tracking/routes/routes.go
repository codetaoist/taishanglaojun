package routes

import (
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/handlers"
	"github.com/gin-gonic/gin"
)

// SetupLocationRoutes 设置位置跟踪路由
func SetupLocationRoutes(router *gin.RouterGroup, handler *handlers.LocationHandler) {
	// 健康检查
	router.GET("/location/health", handler.HealthCheck)

	// 轨迹管理 - 直接使用传入的路由组，避免重复的v1路径
	trajectories := router.Group("/trajectories")
	{
		trajectories.POST("", handler.CreateTrajectory)           // 创建轨迹
		trajectories.GET("", handler.GetTrajectories)            // 获取轨迹列表
		trajectories.GET("/stats", handler.GetTrajectoryStats)   // 获取轨迹统计
		trajectories.GET("/:id", handler.GetTrajectory)          // 获取轨迹详情
		trajectories.PUT("/:id", handler.UpdateTrajectory)       // 更新轨迹
		trajectories.DELETE("/:id", handler.DeleteTrajectory)    // 删除轨迹
	}

	// 位置点管理
	locationPoints := router.Group("/location-points")
	{
		locationPoints.POST("", handler.AddLocationPoint)        // 添加位置点
		locationPoints.POST("/batch", handler.AddLocationPointsBatch) // 批量添加位置点
		locationPoints.GET("", handler.GetLocationPoints)        // 获取位置点
	}

	// 数据同步
	router.POST("/sync", handler.SyncData)                      // 数据同步
}