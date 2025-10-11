package models

import (
	"time"

	"gorm.io/gorm"
)

// LocationPoint 位置点模�?
type LocationPoint struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID      string    `json:"user_id" gorm:"index;type:varchar(36);not null"`
	TrajectoryID string   `json:"trajectory_id" gorm:"index;type:varchar(36);not null"`
	Latitude    float64   `json:"latitude" gorm:"not null;index"`
	Longitude   float64   `json:"longitude" gorm:"not null;index"`
	Altitude    *float64  `json:"altitude,omitempty"`
	Accuracy    *float64  `json:"accuracy,omitempty"`
	Speed       *float64  `json:"speed,omitempty"`
	Bearing     *float64  `json:"bearing,omitempty"`
	Timestamp   int64     `json:"timestamp" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (LocationPoint) TableName() string {
	return "location_points"
}

// LocationPointRequest 位置点请求结�?
type LocationPointRequest struct {
	TrajectoryID string   `json:"trajectory_id" binding:"required"`
	Latitude     float64  `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude    float64  `json:"longitude" binding:"required,min=-180,max=180"`
	Altitude     *float64 `json:"altitude,omitempty"`
	Accuracy     *float64 `json:"accuracy,omitempty"`
	Speed        *float64 `json:"speed,omitempty"`
	Bearing      *float64 `json:"bearing,omitempty"`
	Timestamp    int64    `json:"timestamp" binding:"required"`
}

// LocationPointBatchRequest 批量位置点请求结�?
type LocationPointBatchRequest struct {
	TrajectoryID string                 `json:"trajectory_id" binding:"required"`
	Points       []LocationPointRequest `json:"points" binding:"required,min=1,max=1000"`
}

// LocationPointResponse 位置点响应结�?
type LocationPointResponse struct {
	ID          string    `json:"id"`
	TrajectoryID string   `json:"trajectory_id"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Altitude    *float64  `json:"altitude,omitempty"`
	Accuracy    *float64  `json:"accuracy,omitempty"`
	Speed       *float64  `json:"speed,omitempty"`
	Bearing     *float64  `json:"bearing,omitempty"`
	Timestamp   int64     `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LocationPointQuery 位置点查询参�?
type LocationPointQuery struct {
	TrajectoryID string   `form:"trajectory_id"`
	StartTime    *int64   `form:"start_time"`
	EndTime      *int64   `form:"end_time"`
	MinLat       *float64 `form:"min_lat"`
	MaxLat       *float64 `form:"max_lat"`
	MinLng       *float64 `form:"min_lng"`
	MaxLng       *float64 `form:"max_lng"`
	Limit        int      `form:"limit" binding:"min=1,max=1000"`
	Offset       int      `form:"offset" binding:"min=0"`
}

// ToResponse 转换为响应格�?
func (lp *LocationPoint) ToResponse() LocationPointResponse {
	return LocationPointResponse{
		ID:          lp.ID,
		TrajectoryID: lp.TrajectoryID,
		Latitude:    lp.Latitude,
		Longitude:   lp.Longitude,
		Altitude:    lp.Altitude,
		Accuracy:    lp.Accuracy,
		Speed:       lp.Speed,
		Bearing:     lp.Bearing,
		Timestamp:   lp.Timestamp,
		CreatedAt:   lp.CreatedAt,
	}
}
