package handlers

import "github.com/gin-gonic/gin"

// LocationHandlerInterface 位置处理器接口
type LocationHandlerInterface interface {
	CreateTrajectory(c *gin.Context)
	GetTrajectories(c *gin.Context)
	GetTrajectory(c *gin.Context)
	UpdateTrajectory(c *gin.Context)
	DeleteTrajectory(c *gin.Context)
	AddLocationPoint(c *gin.Context)
	AddLocationPointsBatch(c *gin.Context)
	GetLocationPoints(c *gin.Context)
	GetTrajectoryStats(c *gin.Context)
	SyncData(c *gin.Context)
	HealthCheck(c *gin.Context)
	UploadLocationPoints(c *gin.Context)
	DeleteLocationPoint(c *gin.Context)
	BatchUploadPoints(c *gin.Context)
	FinishTrajectory(c *gin.Context)
	GetTrajectoryPoints(c *gin.Context)
	GetSyncStatus(c *gin.Context)
}