package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/taishanglaojun/health-management/internal/domain"
)

// PostgreSQLHealthDataRepository PostgreSQL健康数据仓储实现
type PostgreSQLHealthDataRepository struct {
	db *gorm.DB
}

// NewPostgreSQLHealthDataRepository 创建PostgreSQL健康数据仓储
func NewPostgreSQLHealthDataRepository(db *gorm.DB) domain.HealthDataRepository {
	return &PostgreSQLHealthDataRepository{
		db: db,
	}
}

// Save 保存健康数据
func (r *PostgreSQLHealthDataRepository) Save(ctx context.Context, healthData *domain.HealthData) error {
	return r.db.WithContext(ctx).Create(healthData).Error
}

// FindByID 根据ID查找健康数据
func (r *PostgreSQLHealthDataRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.HealthData, error) {
	var healthData domain.HealthData
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&healthData).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &healthData, nil
}

// Update 更新健康数据
func (r *PostgreSQLHealthDataRepository) Update(ctx context.Context, healthData *domain.HealthData) error {
	return r.db.WithContext(ctx).Save(healthData).Error
}

// Delete 删除健康数据
func (r *PostgreSQLHealthDataRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.HealthData{}).Error
}

// FindByUserID 根据用户ID查找健康数据
func (r *PostgreSQLHealthDataRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&healthDataList).Error
	return healthDataList, err
}

// FindByUserIDAndType 根据用户ID和数据类型查找健康数据
func (r *PostgreSQLHealthDataRepository) FindByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ?", userID, dataType).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&healthDataList).Error
	return healthDataList, err
}

// FindByUserIDAndTimeRange 根据用户ID和时间范围查找健康数据
func (r *PostgreSQLHealthDataRepository) FindByUserIDAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND recorded_at >= ? AND recorded_at <= ?", userID, startTime, endTime).
		Order("recorded_at DESC").
		Find(&healthDataList).Error
	return healthDataList, err
}

// FindByUserIDTypeAndTimeRange 根据用户ID、数据类型和时间范围查找健康数据
func (r *PostgreSQLHealthDataRepository) FindByUserIDTypeAndTimeRange(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType, startTime, endTime time.Time) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ? AND recorded_at >= ? AND recorded_at <= ?", userID, dataType, startTime, endTime).
		Order("recorded_at DESC").
		Find(&healthDataList).Error
	return healthDataList, err
}

// CountByUserID 统计用户的健康数据总数
func (r *PostgreSQLHealthDataRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthData{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByUserIDAndType 统计用户特定类型的健康数据总数
func (r *PostgreSQLHealthDataRepository) CountByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthData{}).
		Where("user_id = ? AND data_type = ?", userID, dataType).
		Count(&count).Error
	return count, err
}

// GetLatestByUserIDAndType 获取用户特定类型的最新健康数据
func (r *PostgreSQLHealthDataRepository) GetLatestByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType) (*domain.HealthData, error) {
	var healthData domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ?", userID, dataType).
		Order("recorded_at DESC").
		First(&healthData).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &healthData, nil
}

// GetAverageByUserIDTypeAndTimeRange 获取用户特定类型在时间范围内的平均值
func (r *PostgreSQLHealthDataRepository) GetAverageByUserIDTypeAndTimeRange(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType, startTime, endTime time.Time) (float64, error) {
	var result struct {
		Average float64
	}
	
	err := r.db.WithContext(ctx).
		Model(&domain.HealthData{}).
		Select("AVG(value) as average").
		Where("user_id = ? AND data_type = ? AND recorded_at >= ? AND recorded_at <= ?", userID, dataType, startTime, endTime).
		Scan(&result).Error
	
	return result.Average, err
}

// GetMinMaxByUserIDTypeAndTimeRange 获取用户特定类型在时间范围内的最小值和最大值
func (r *PostgreSQLHealthDataRepository) GetMinMaxByUserIDTypeAndTimeRange(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType, startTime, endTime time.Time) (min, max float64, err error) {
	var result struct {
		Min float64
		Max float64
	}
	
	err = r.db.WithContext(ctx).
		Model(&domain.HealthData{}).
		Select("MIN(value) as min, MAX(value) as max").
		Where("user_id = ? AND data_type = ? AND recorded_at >= ? AND recorded_at <= ?", userID, dataType, startTime, endTime).
		Scan(&result).Error
	
	return result.Min, result.Max, err
}

// FindAbnormalDataByUserID 查找用户的异常健康数据
func (r *PostgreSQLHealthDataRepository) FindAbnormalDataByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	
	// 查询所有数据，然后在应用层过滤异常数据
	// 这里简化处理，实际应该在数据库层面进行过滤以提高性能
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("recorded_at DESC").
		Limit(limit * 2). // 获取更多数据以确保有足够的异常数据
		Find(&healthDataList).Error
	
	if err != nil {
		return nil, err
	}
	
	// 过滤异常数据
	var abnormalData []*domain.HealthData
	for _, data := range healthDataList {
		if data.IsAbnormal() {
			abnormalData = append(abnormalData, data)
			if len(abnormalData) >= limit {
				break
			}
		}
	}
	
	// 应用偏移量
	if offset >= len(abnormalData) {
		return []*domain.HealthData{}, nil
	}
	
	end := offset + limit
	if end > len(abnormalData) {
		end = len(abnormalData)
	}
	
	return abnormalData[offset:end], nil
}

// FindAbnormalDataByUserIDAndTimeRange 查找用户在时间范围内的异常健康数据
func (r *PostgreSQLHealthDataRepository) FindAbnormalDataByUserIDAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	
	// 查询时间范围内的所有数据
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND recorded_at >= ? AND recorded_at <= ?", userID, startTime, endTime).
		Order("recorded_at DESC").
		Find(&healthDataList).Error
	
	if err != nil {
		return nil, err
	}
	
	// 过滤异常数据
	var abnormalData []*domain.HealthData
	for _, data := range healthDataList {
		if data.IsAbnormal() {
			abnormalData = append(abnormalData, data)
		}
	}
	
	return abnormalData, nil
}

// 优化版本的异常数据查询（使用数据库层面的条件过滤）
// 这个方法展示了如何在数据库层面进行异常数据的过滤，提高查询性能

// FindAbnormalHeartRateByUserID 查找用户的异常心率数据
func (r *PostgreSQLHealthDataRepository) FindAbnormalHeartRateByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ? AND (value < 60 OR value > 100)", userID, domain.HeartRate).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&healthDataList).Error
	return healthDataList, err
}

// FindAbnormalBloodPressureByUserID 查找用户的异常血压数据
func (r *PostgreSQLHealthDataRepository) FindAbnormalBloodPressureByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ? AND (value < 90 OR value > 140)", userID, domain.BloodPressure).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&healthDataList).Error
	return healthDataList, err
}

// FindAbnormalBloodSugarByUserID 查找用户的异常血糖数据
func (r *PostgreSQLHealthDataRepository) FindAbnormalBloodSugarByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ? AND (value < 3.9 OR value > 6.1)", userID, domain.BloodSugar).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&healthDataList).Error
	return healthDataList, err
}

// FindAbnormalBodyTemperatureByUserID 查找用户的异常体温数据
func (r *PostgreSQLHealthDataRepository) FindAbnormalBodyTemperatureByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ? AND (value < 36.1 OR value > 37.2)", userID, domain.BodyTemperature).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&healthDataList).Error
	return healthDataList, err
}

// FindAbnormalBMIByUserID 查找用户的异常BMI数据
func (r *PostgreSQLHealthDataRepository) FindAbnormalBMIByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ? AND (value < 18.5 OR value > 24)", userID, domain.BMI).
		Order("recorded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&healthDataList).Error
	return healthDataList, err
}