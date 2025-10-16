package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/taishanglaojun/health-management/internal/domain"
)

// PostgreSQLHealthDataRepository PostgreSQL
type PostgreSQLHealthDataRepository struct {
	db *gorm.DB
}

// NewPostgreSQLHealthDataRepository PostgreSQL
func NewPostgreSQLHealthDataRepository(db *gorm.DB) domain.HealthDataRepository {
	return &PostgreSQLHealthDataRepository{
		db: db,
	}
}

// Save 潡
func (r *PostgreSQLHealthDataRepository) Save(ctx context.Context, healthData *domain.HealthData) error {
	return r.db.WithContext(ctx).Create(healthData).Error
}

// FindByID ID
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

// Update 
func (r *PostgreSQLHealthDataRepository) Update(ctx context.Context, healthData *domain.HealthData) error {
	return r.db.WithContext(ctx).Save(healthData).Error
}

// Delete 
func (r *PostgreSQLHealthDataRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.HealthData{}).Error
}

// FindByUserID ID
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

// FindByUserIDAndType ID?
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

// FindByUserIDAndTimeRange ID?
func (r *PostgreSQLHealthDataRepository) FindByUserIDAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND recorded_at >= ? AND recorded_at <= ?", userID, startTime, endTime).
		Order("recorded_at DESC").
		Find(&healthDataList).Error
	return healthDataList, err
}

// FindByUserIDTypeAndTimeRange ID
func (r *PostgreSQLHealthDataRepository) FindByUserIDTypeAndTimeRange(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType, startTime, endTime time.Time) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND data_type = ? AND recorded_at >= ? AND recorded_at <= ?", userID, dataType, startTime, endTime).
		Order("recorded_at DESC").
		Find(&healthDataList).Error
	return healthDataList, err
}

// CountByUserID 
func (r *PostgreSQLHealthDataRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthData{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByUserIDAndType 
func (r *PostgreSQLHealthDataRepository) CountByUserIDAndType(ctx context.Context, userID uuid.UUID, dataType domain.HealthDataType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthData{}).
		Where("user_id = ? AND data_type = ?", userID, dataType).
		Count(&count).Error
	return count, err
}

// GetLatestByUserIDAndType ?
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

// GetAverageByUserIDTypeAndTimeRange ?
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

// GetMinMaxByUserIDTypeAndTimeRange ?
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

// FindAbnormalDataByUserID ?
func (r *PostgreSQLHealthDataRepository) FindAbnormalDataByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	
	// 
	// 
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("recorded_at DESC").
		Limit(limit * 2). // 㹻?
		Find(&healthDataList).Error
	
	if err != nil {
		return nil, err
	}
	
	// 
	var abnormalData []*domain.HealthData
	for _, data := range healthDataList {
		if data.IsAbnormal() {
			abnormalData = append(abnormalData, data)
			if len(abnormalData) >= limit {
				break
			}
		}
	}
	
	// ?
	if offset >= len(abnormalData) {
		return []*domain.HealthData{}, nil
	}
	
	end := offset + limit
	if end > len(abnormalData) {
		end = len(abnormalData)
	}
	
	return abnormalData[offset:end], nil
}

// FindAbnormalDataByUserIDAndTimeRange ?
func (r *PostgreSQLHealthDataRepository) FindAbnormalDataByUserIDAndTimeRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*domain.HealthData, error) {
	var healthDataList []*domain.HealthData
	
	// ?
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND recorded_at >= ? AND recorded_at <= ?", userID, startTime, endTime).
		Order("recorded_at DESC").
		Find(&healthDataList).Error
	
	if err != nil {
		return nil, err
	}
	
	// 
	var abnormalData []*domain.HealthData
	for _, data := range healthDataList {
		if data.IsAbnormal() {
			abnormalData = append(abnormalData, data)
		}
	}
	
	return abnormalData, nil
}

// 汾?
// 

// FindAbnormalHeartRateByUserID ?
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

// FindAbnormalBloodPressureByUserID ?
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

// FindAbnormalBloodSugarByUserID ?
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

// FindAbnormalBodyTemperatureByUserID ?
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

// FindAbnormalBMIByUserID BMI
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

