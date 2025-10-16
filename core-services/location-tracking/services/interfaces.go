package services

import (
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/models"
)

// LocationServiceInterface 位置服务接口
type LocationServiceInterface interface {
	CreateTrajectory(userID string, req models.TrajectoryRequest) (*models.Trajectory, error)
	GetTrajectories(userID string, query models.TrajectoryQuery) ([]models.Trajectory, int64, error)
	GetTrajectory(userID, trajectoryID string, includePoints bool) (*models.Trajectory, error)
	UpdateTrajectory(userID, trajectoryID string, req models.TrajectoryUpdateRequest) (*models.Trajectory, error)
	DeleteTrajectory(userID, trajectoryID string) error
	AddLocationPoint(userID string, req models.LocationPointRequest) (*models.LocationPoint, error)
	AddLocationPointsBatch(userID string, req models.LocationPointBatchRequest) ([]models.LocationPoint, error)
	GetLocationPoints(userID string, query models.LocationPointQuery) ([]models.LocationPoint, int64, error)
	GetTrajectoryStats(userID string) (*models.TrajectoryStats, error)
	SyncData(userID string, req models.SyncRequest) (*models.SyncResponse, error)
	DeleteLocationPoint(userID, pointID string) error
	CleanupExpiredData(days int) error
	UpdateTrajectoryStats(trajectoryID string) error
}

// BatchUploadServiceInterface 批量上传服务接口
type BatchUploadServiceInterface interface {
	SubmitUploadTask(userID, trajectoryID string, points []models.LocationPoint) error
	GetQueueStatus() map[string]interface{}
	Stop()
}

