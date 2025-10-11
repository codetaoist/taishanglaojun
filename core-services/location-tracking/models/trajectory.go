package models

import (
	"time"

	"gorm.io/gorm"
)

// Trajectory 轨迹模型
type Trajectory struct {
	ID           string          `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID       string          `gorm:"type:varchar(36);not null;index" json:"user_id"`
	Name         string          `gorm:"type:varchar(255);not null" json:"name"`
	Description  *string         `gorm:"type:text" json:"description"`
	StartTime    int64           `gorm:"not null;index" json:"start_time"`
	EndTime      *int64          `gorm:"index" json:"end_time"`
	Distance     float64         `gorm:"default:0" json:"distance"`
	Duration     int64           `gorm:"default:0" json:"duration"`
	MaxSpeed     float64         `gorm:"default:0" json:"max_speed"`
	AvgSpeed     float64         `gorm:"default:0" json:"avg_speed"`
	PointCount   int             `gorm:"default:0" json:"point_count"`
	MinLatitude  float64         `gorm:"default:0" json:"min_latitude"`
	MaxLatitude  float64         `gorm:"default:0" json:"max_latitude"`
	MinLongitude float64         `gorm:"default:0" json:"min_longitude"`
	MaxLongitude float64         `gorm:"default:0" json:"max_longitude"`
	IsActive     bool            `gorm:"default:true;index" json:"is_active"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    *gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Points []LocationPoint `gorm:"foreignKey:TrajectoryID" json:"points,omitempty"`
}

// TableName 指定表名
func (Trajectory) TableName() string {
	return "trajectories"
}

// TrajectoryRequest 轨迹创建请求结构
type TrajectoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
	StartTime   int64  `json:"start_time" binding:"required"`
}

// TrajectoryUpdateRequest 轨迹更新请求结构
type TrajectoryUpdateRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	EndTime     *int64  `json:"end_time,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// TrajectoryResponse 轨迹响应结构
type TrajectoryResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	StartTime    int64     `json:"start_time"`
	EndTime      *int64    `json:"end_time,omitempty"`
	Distance     float64   `json:"distance"`
	Duration     int64     `json:"duration"`
	MaxSpeed     float64   `json:"max_speed"`
	AvgSpeed     float64   `json:"avg_speed"`
	PointCount   int       `json:"point_count"`
	MinLatitude  float64   `json:"min_latitude"`
	MaxLatitude  float64   `json:"max_latitude"`
	MinLongitude float64   `json:"min_longitude"`
	MaxLongitude float64   `json:"max_longitude"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TrajectoryDetailResponse 轨迹详情响应结构（包含位置点�?
type TrajectoryDetailResponse struct {
	TrajectoryResponse
	Points []LocationPointResponse `json:"points,omitempty"`
}

// TrajectoryQuery 轨迹查询参数
type TrajectoryQuery struct {
	Name      string `form:"name"`
	IsActive  *bool  `form:"is_active"`
	StartTime *int64 `form:"start_time"`
	EndTime   *int64 `form:"end_time"`
	Limit     int    `form:"limit,default=50"`
	Offset    int    `form:"offset,default=0"`
	OrderBy   string `form:"order_by,default=created_at"`
	Order     string `form:"order,default=desc"`
}

// TrajectoryStats 轨迹统计信息
type TrajectoryStats struct {
	TotalTrajectories int     `json:"total_trajectories"`
	ActiveTrajectories int    `json:"active_trajectories"`
	TotalDistance     float64 `json:"total_distance"`
	TotalDuration     int64   `json:"total_duration"`
	TotalPoints       int     `json:"total_points"`
	AvgSpeed          float64 `json:"avg_speed"`
	MaxSpeed          float64 `json:"max_speed"`
}

// SyncRequest 数据同步请求结构
type SyncRequest struct {
	LastSyncTime   int64    `json:"last_sync_time" binding:"required"`
	DeviceID       string   `json:"device_id" binding:"required"`
	TrajectoryIDs  []string `json:"trajectory_ids,omitempty"`
}

// SyncResponse 同步响应
type SyncResponse struct {
	NewTrajectories      []TrajectoryResponse `json:"new_trajectories"`
	UpdatedTrajectories  []TrajectoryResponse `json:"updated_trajectories"`
	DeletedTrajectoryIDs []string             `json:"deleted_trajectory_ids"`
	SyncTime            int64                `json:"sync_time"`
}

// ToResponse 转换为响应格�?
func (t *Trajectory) ToResponse() TrajectoryResponse {
	description := ""
	if t.Description != nil {
		description = *t.Description
	}
	
	return TrajectoryResponse{
		ID:           t.ID,
		Name:         t.Name,
		Description:  description,
		StartTime:    t.StartTime,
		EndTime:      t.EndTime,
		Distance:     t.Distance,
		Duration:     t.Duration,
		MaxSpeed:     t.MaxSpeed,
		AvgSpeed:     t.AvgSpeed,
		PointCount:   t.PointCount,
		MinLatitude:  t.MinLatitude,
		MaxLatitude:  t.MaxLatitude,
		MinLongitude: t.MinLongitude,
		MaxLongitude: t.MaxLongitude,
		IsActive:     t.IsActive,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

// ToDetailResponse 转换为详细响应格�?
func (t *Trajectory) ToDetailResponse() TrajectoryDetailResponse {
	var points []LocationPointResponse
	if t.Points != nil {
		points = make([]LocationPointResponse, len(t.Points))
		for i, point := range t.Points {
			points[i] = point.ToResponse()
		}
	}

	return TrajectoryDetailResponse{
		TrajectoryResponse: t.ToResponse(),
		Points:            points,
	}
}
