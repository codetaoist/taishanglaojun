package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LocationService 位置服务
type LocationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewLocationService 创建位置服务实例
func NewLocationService(db *gorm.DB, logger *zap.Logger) *LocationService {
	return &LocationService{
		db:     db,
		logger: logger,
	}
}

// CreateTrajectory 创建轨迹
func (s *LocationService) CreateTrajectory(userID string, req models.TrajectoryRequest) (*models.Trajectory, error) {
	var description *string
	if req.Description != "" {
		description = &req.Description
	}
	
	trajectory := &models.Trajectory{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Description: description,
		StartTime:   req.StartTime,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(trajectory).Error; err != nil {
		s.logger.Error("Failed to create trajectory", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	s.logger.Info("Trajectory created", zap.String("trajectory_id", trajectory.ID), zap.String("user_id", userID))
	return trajectory, nil
}

// GetTrajectories 获取用户轨迹列表
func (s *LocationService) GetTrajectories(userID string, query models.TrajectoryQuery) ([]models.Trajectory, int64, error) {
	var trajectories []models.Trajectory
	var total int64

	db := s.db.Model(&models.Trajectory{}).Where("user_id = ?", userID)

	// 应用过滤条件
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}
	if query.StartTime != nil {
		db = db.Where("start_time >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("end_time <= ?", *query.EndTime)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count trajectories", zap.Error(err), zap.String("user_id", userID))
		return nil, 0, err
	}

	// 应用排序和分页
	orderBy := query.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	order := query.Order
	if order == "" {
		order = "desc"
	}

	if err := db.Order(fmt.Sprintf("%s %s", orderBy, order)).
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&trajectories).Error; err != nil {
		s.logger.Error("Failed to get trajectories", zap.Error(err), zap.String("user_id", userID))
		return nil, 0, err
	}

	return trajectories, total, nil
}

// GetTrajectory 获取轨迹详情
func (s *LocationService) GetTrajectory(userID, trajectoryID string, includePoints bool) (*models.Trajectory, error) {
	var trajectory models.Trajectory

	query := s.db.Where("id = ? AND user_id = ?", trajectoryID, userID)
	if includePoints {
		query = query.Preload("Points", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp ASC")
		})
	}

	if err := query.First(&trajectory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("trajectory not found")
		}
		s.logger.Error("Failed to get trajectory", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		return nil, err
	}

	return &trajectory, nil
}

// UpdateTrajectory 更新轨迹
func (s *LocationService) UpdateTrajectory(userID, trajectoryID string, req models.TrajectoryUpdateRequest) (*models.Trajectory, error) {
	var trajectory models.Trajectory

	if err := s.db.Where("id = ? AND user_id = ?", trajectoryID, userID).First(&trajectory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("trajectory not found")
		}
		return nil, err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
		// 如果设置了结束时间，计算持续时间
		if *req.EndTime > trajectory.StartTime {
			updates["duration"] = *req.EndTime - trajectory.StartTime
		}
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
		// 如果设置为非活跃状态且没有结束时间，设置当前时间为结束时间
		if !*req.IsActive && trajectory.EndTime == nil {
			now := time.Now().UnixMilli()
			updates["end_time"] = now
			updates["duration"] = now - trajectory.StartTime
		}
	}

	updates["updated_at"] = time.Now()

	if err := s.db.Model(&trajectory).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update trajectory", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		return nil, err
	}

	// 重新获取更新后的轨迹
	if err := s.db.Where("id = ?", trajectoryID).First(&trajectory).Error; err != nil {
		return nil, err
	}

	s.logger.Info("Trajectory updated", zap.String("trajectory_id", trajectoryID), zap.String("user_id", userID))
	return &trajectory, nil
}

// DeleteTrajectory 删除轨迹
func (s *LocationService) DeleteTrajectory(userID, trajectoryID string) error {
	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查轨迹是否存在
	var trajectory models.Trajectory
	if err := tx.Where("id = ? AND user_id = ?", trajectoryID, userID).First(&trajectory).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("trajectory not found")
		}
		return err
	}

	// 删除相关的位置点
	if err := tx.Where("trajectory_id = ?", trajectoryID).Delete(&models.LocationPoint{}).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete location points", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		return err
	}

	// 删除轨迹
	if err := tx.Delete(&trajectory).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete trajectory", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		return err
	}

	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		return err
	}

	s.logger.Info("Trajectory deleted", zap.String("trajectory_id", trajectoryID), zap.String("user_id", userID))
	return nil
}

// AddLocationPoint 添加位置点
func (s *LocationService) AddLocationPoint(userID string, req models.LocationPointRequest) (*models.LocationPoint, error) {
	// 验证轨迹是否存在且属于用户
	var trajectory models.Trajectory
	if err := s.db.Where("id = ? AND user_id = ?", req.TrajectoryID, userID).First(&trajectory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("trajectory not found")
		}
		return nil, err
	}

	point := &models.LocationPoint{
		ID:          uuid.New().String(),
		UserID:      userID,
		TrajectoryID: req.TrajectoryID,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Altitude:    req.Altitude,
		Accuracy:    req.Accuracy,
		Speed:       req.Speed,
		Bearing:     req.Bearing,
		Timestamp:   req.Timestamp,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(point).Error; err != nil {
		s.logger.Error("Failed to create location point", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	// 异步更新轨迹统计信息
	go s.updateTrajectoryStats(req.TrajectoryID)

	return point, nil
}

// AddLocationPointsBatch 批量添加位置点
func (s *LocationService) AddLocationPointsBatch(userID string, req models.LocationPointBatchRequest) ([]models.LocationPoint, error) {
	// 验证轨迹是否存在且属于用户
	var trajectory models.Trajectory
	if err := s.db.Where("id = ? AND user_id = ?", req.TrajectoryID, userID).First(&trajectory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("trajectory not found")
		}
		return nil, err
	}

	points := make([]models.LocationPoint, len(req.Points))
	now := time.Now()

	for i, pointReq := range req.Points {
		points[i] = models.LocationPoint{
			ID:          uuid.New().String(),
			UserID:      userID,
			TrajectoryID: req.TrajectoryID,
			Latitude:    pointReq.Latitude,
			Longitude:   pointReq.Longitude,
			Altitude:    pointReq.Altitude,
			Accuracy:    pointReq.Accuracy,
			Speed:       pointReq.Speed,
			Bearing:     pointReq.Bearing,
			Timestamp:   pointReq.Timestamp,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}

	if err := s.db.CreateInBatches(points, 100).Error; err != nil {
		s.logger.Error("Failed to create location points batch", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	// 异步更新轨迹统计信息
	go s.updateTrajectoryStats(req.TrajectoryID)

	s.logger.Info("Location points batch created", 
		zap.String("trajectory_id", req.TrajectoryID), 
		zap.String("user_id", userID),
		zap.Int("count", len(points)))

	return points, nil
}

// GetLocationPoints 获取位置点
func (s *LocationService) GetLocationPoints(userID string, query models.LocationPointQuery) ([]models.LocationPoint, int64, error) {
	var points []models.LocationPoint
	var total int64

	db := s.db.Model(&models.LocationPoint{}).Where("user_id = ?", userID)

	// 应用过滤条件
	if query.TrajectoryID != "" {
		// 验证轨迹是否属于用户
		var trajectory models.Trajectory
		if err := s.db.Where("id = ? AND user_id = ?", query.TrajectoryID, userID).First(&trajectory).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, errors.New("trajectory not found")
			}
			return nil, 0, err
		}
		db = db.Where("trajectory_id = ?", query.TrajectoryID)
	}

	if query.StartTime != nil {
		db = db.Where("timestamp >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("timestamp <= ?", *query.EndTime)
	}
	if query.MinLat != nil && query.MaxLat != nil {
		db = db.Where("latitude BETWEEN ? AND ?", *query.MinLat, *query.MaxLat)
	}
	if query.MinLng != nil && query.MaxLng != nil {
		db = db.Where("longitude BETWEEN ? AND ?", *query.MinLng, *query.MaxLng)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		s.logger.Error("Failed to count location points", zap.Error(err), zap.String("user_id", userID))
		return nil, 0, err
	}

	// 应用分页和排序
	if err := db.Order("timestamp ASC").
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&points).Error; err != nil {
		s.logger.Error("Failed to get location points", zap.Error(err), zap.String("user_id", userID))
		return nil, 0, err
	}

	return points, total, nil
}

// GetTrajectoryStats 获取轨迹统计信息
func (s *LocationService) GetTrajectoryStats(userID string) (*models.TrajectoryStats, error) {
	var stats models.TrajectoryStats

	// 获取轨迹统计
	if err := s.db.Model(&models.Trajectory{}).
		Where("user_id = ?", userID).
		Select("COUNT(*) as total_trajectories, SUM(CASE WHEN is_active THEN 1 ELSE 0 END) as active_trajectories, SUM(distance) as total_distance, SUM(duration) as total_duration, SUM(point_count) as total_points, AVG(avg_speed) as avg_speed, MAX(max_speed) as max_speed").
		Scan(&stats).Error; err != nil {
		s.logger.Error("Failed to get trajectory stats", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	return &stats, nil
}

// updateTrajectoryStats 更新轨迹统计信息
func (s *LocationService) updateTrajectoryStats(trajectoryID string) {
	var points []models.LocationPoint
	if err := s.db.Where("trajectory_id = ?", trajectoryID).Order("timestamp ASC").Find(&points).Error; err != nil {
		s.logger.Error("Failed to get points for stats update", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		return
	}

	if len(points) == 0 {
		return
	}

	// 计算统计信息
	var totalDistance float64
	var maxSpeed float64
	var speedSum float64
	var speedCount int
	var minLat, maxLat, minLng, maxLng float64

	minLat = points[0].Latitude
	maxLat = points[0].Latitude
	minLng = points[0].Longitude
	maxLng = points[0].Longitude

	for i, point := range points {
		// 更新边界
		if point.Latitude < minLat {
			minLat = point.Latitude
		}
		if point.Latitude > maxLat {
			maxLat = point.Latitude
		}
		if point.Longitude < minLng {
			minLng = point.Longitude
		}
		if point.Longitude > maxLng {
			maxLng = point.Longitude
		}

		// 计算距离（从第二个点开始）
		if i > 0 {
			distance := s.calculateDistance(
				points[i-1].Latitude, points[i-1].Longitude,
				point.Latitude, point.Longitude,
			)
			totalDistance += distance
		}

		// 处理速度
		if point.Speed != nil {
			speed := *point.Speed
			if speed > maxSpeed {
				maxSpeed = speed
			}
			speedSum += speed
			speedCount++
		}
	}

	avgSpeed := float64(0)
	if speedCount > 0 {
		avgSpeed = speedSum / float64(speedCount)
	}

	// 更新轨迹统计
	updates := map[string]interface{}{
		"distance":      totalDistance,
		"max_speed":     maxSpeed,
		"avg_speed":     avgSpeed,
		"point_count":   len(points),
		"min_latitude":  minLat,
		"max_latitude":  maxLat,
		"min_longitude": minLng,
		"max_longitude": maxLng,
		"updated_at":    time.Now(),
	}

	if err := s.db.Model(&models.Trajectory{}).Where("id = ?", trajectoryID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update trajectory stats", zap.Error(err), zap.String("trajectory_id", trajectoryID))
	}
}

// calculateDistance 计算两点间距离（米）
func (s *LocationService) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371000 // 地球半径（米）

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// SyncData 数据同步
func (s *LocationService) SyncData(userID string, req models.SyncRequest) (*models.SyncResponse, error) {
	var newTrajectories []models.Trajectory
	var updatedTrajectories []models.Trajectory
	var deletedTrajectoryIDs []string

	// 获取新增的轨迹
	if err := s.db.Where("user_id = ? AND created_at > ?", userID, time.UnixMilli(req.LastSyncTime)).
		Find(&newTrajectories).Error; err != nil {
		s.logger.Error("Failed to get new trajectories for sync", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	// 获取更新的轨迹
	if err := s.db.Where("user_id = ? AND updated_at > ? AND created_at <= ?", 
		userID, time.UnixMilli(req.LastSyncTime), time.UnixMilli(req.LastSyncTime)).
		Find(&updatedTrajectories).Error; err != nil {
		s.logger.Error("Failed to get updated trajectories for sync", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	// 检查删除的轨迹（简化实现，实际应该有软删除标记）
	if len(req.TrajectoryIDs) > 0 {
		var existingIDs []string
		if err := s.db.Model(&models.Trajectory{}).
			Where("user_id = ? AND id IN ?", userID, req.TrajectoryIDs).
			Pluck("id", &existingIDs).Error; err != nil {
			s.logger.Error("Failed to check existing trajectories", zap.Error(err), zap.String("user_id", userID))
			return nil, err
		}

		existingIDMap := make(map[string]bool)
		for _, id := range existingIDs {
			existingIDMap[id] = true
		}

		for _, id := range req.TrajectoryIDs {
			if !existingIDMap[id] {
				deletedTrajectoryIDs = append(deletedTrajectoryIDs, id)
			}
		}
	}

	response := &models.SyncResponse{
		NewTrajectories:      make([]models.TrajectoryResponse, len(newTrajectories)),
		UpdatedTrajectories:  make([]models.TrajectoryResponse, len(updatedTrajectories)),
		DeletedTrajectoryIDs: deletedTrajectoryIDs,
		SyncTime:            time.Now().UnixMilli(),
	}

	for i, traj := range newTrajectories {
		response.NewTrajectories[i] = traj.ToResponse()
	}

	for i, traj := range updatedTrajectories {
		response.UpdatedTrajectories[i] = traj.ToResponse()
	}

	s.logger.Info("Data sync completed", 
		zap.String("user_id", userID),
		zap.Int("new_count", len(newTrajectories)),
		zap.Int("updated_count", len(updatedTrajectories)),
		zap.Int("deleted_count", len(deletedTrajectoryIDs)))

	return response, nil
}