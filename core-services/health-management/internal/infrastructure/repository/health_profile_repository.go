package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/taishanglaojun/health-management/internal/domain"
)

// PostgreSQLHealthProfileRepository PostgreSQL健康档案仓储实现
type PostgreSQLHealthProfileRepository struct {
	db *gorm.DB
}

// NewPostgreSQLHealthProfileRepository 创建PostgreSQL健康档案仓储
func NewPostgreSQLHealthProfileRepository(db *gorm.DB) domain.HealthProfileRepository {
	return &PostgreSQLHealthProfileRepository{
		db: db,
	}
}

// Save 保存健康档案
func (r *PostgreSQLHealthProfileRepository) Save(ctx context.Context, profile *domain.HealthProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

// FindByID 根据ID查找健康档案
func (r *PostgreSQLHealthProfileRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.HealthProfile, error) {
	var profile domain.HealthProfile
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// FindByUserID 根据用户ID查找健康档案
func (r *PostgreSQLHealthProfileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.HealthProfile, error) {
	var profile domain.HealthProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// Update 更新健康档案
func (r *PostgreSQLHealthProfileRepository) Update(ctx context.Context, profile *domain.HealthProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// Delete 删除健康档案
func (r *PostgreSQLHealthProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.HealthProfile{}).Error
}

// List 分页查询健康档案列表
func (r *PostgreSQLHealthProfileRepository) List(ctx context.Context, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error
	return profiles, err
}

// Count 统计健康档案总数
func (r *PostgreSQLHealthProfileRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Count(&count).Error
	return count, err
}

// FindByGender 根据性别查找健康档案
func (r *PostgreSQLHealthProfileRepository) FindByGender(ctx context.Context, gender domain.Gender, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	err := r.db.WithContext(ctx).
		Where("gender = ?", gender).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error
	return profiles, err
}

// FindByAgeRange 根据年龄范围查找健康档案
func (r *PostgreSQLHealthProfileRepository) FindByAgeRange(ctx context.Context, minAge, maxAge int, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	
	// 计算出生日期范围
	// 这里简化处理，实际应该考虑更精确的年龄计算
	query := r.db.WithContext(ctx).
		Where("EXTRACT(YEAR FROM AGE(date_of_birth)) BETWEEN ? AND ?", minAge, maxAge).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)
	
	err := query.Find(&profiles).Error
	return profiles, err
}

// FindByBloodType 根据血型查找健康档案
func (r *PostgreSQLHealthProfileRepository) FindByBloodType(ctx context.Context, bloodType domain.BloodType, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	err := r.db.WithContext(ctx).
		Where("blood_type = ?", bloodType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error
	return profiles, err
}

// CountByGender 统计特定性别的健康档案数量
func (r *PostgreSQLHealthProfileRepository) CountByGender(ctx context.Context, gender domain.Gender) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("gender = ?", gender).
		Count(&count).Error
	return count, err
}

// CountByAgeRange 统计特定年龄范围的健康档案数量
func (r *PostgreSQLHealthProfileRepository) CountByAgeRange(ctx context.Context, minAge, maxAge int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("EXTRACT(YEAR FROM AGE(date_of_birth)) BETWEEN ? AND ?", minAge, maxAge).
		Count(&count).Error
	return count, err
}

// CountByBloodType 统计特定血型的健康档案数量
func (r *PostgreSQLHealthProfileRepository) CountByBloodType(ctx context.Context, bloodType domain.BloodType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("blood_type = ?", bloodType).
		Count(&count).Error
	return count, err
}

// FindWithAllergies 查找有过敏史的健康档案
func (r *PostgreSQLHealthProfileRepository) FindWithAllergies(ctx context.Context, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	err := r.db.WithContext(ctx).
		Where("allergies IS NOT NULL AND allergies != ''").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error
	return profiles, err
}

// FindWithMedications 查找正在服药的健康档案
func (r *PostgreSQLHealthProfileRepository) FindWithMedications(ctx context.Context, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	err := r.db.WithContext(ctx).
		Where("medications IS NOT NULL AND medications != ''").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error
	return profiles, err
}

// FindWithMedicalHistory 查找有病史的健康档案
func (r *PostgreSQLHealthProfileRepository) FindWithMedicalHistory(ctx context.Context, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	err := r.db.WithContext(ctx).
		Where("medical_history IS NOT NULL AND medical_history != ''").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error
	return profiles, err
}

// SearchByKeyword 根据关键词搜索健康档案（搜索病史、过敏史、药物等）
func (r *PostgreSQLHealthProfileRepository) SearchByKeyword(ctx context.Context, keyword string, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	searchPattern := "%" + keyword + "%"
	
	err := r.db.WithContext(ctx).
		Where("medical_history ILIKE ? OR allergies ILIKE ? OR medications ILIKE ?", 
			searchPattern, searchPattern, searchPattern).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error
	return profiles, err
}

// GetStatistics 获取健康档案统计信息
func (r *PostgreSQLHealthProfileRepository) GetStatistics(ctx context.Context) (*domain.HealthProfileStatistics, error) {
	var stats domain.HealthProfileStatistics
	
	// 总数统计
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Count(&stats.TotalProfiles).Error
	if err != nil {
		return nil, err
	}
	
	// 性别统计
	var genderStats []struct {
		Gender domain.Gender
		Count  int64
	}
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Select("gender, COUNT(*) as count").
		Group("gender").
		Scan(&genderStats).Error
	if err != nil {
		return nil, err
	}
	
	stats.GenderDistribution = make(map[domain.Gender]int64)
	for _, stat := range genderStats {
		stats.GenderDistribution[stat.Gender] = stat.Count
	}
	
	// 血型统计
	var bloodTypeStats []struct {
		BloodType domain.BloodType
		Count     int64
	}
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Select("blood_type, COUNT(*) as count").
		Group("blood_type").
		Scan(&bloodTypeStats).Error
	if err != nil {
		return nil, err
	}
	
	stats.BloodTypeDistribution = make(map[domain.BloodType]int64)
	for _, stat := range bloodTypeStats {
		stats.BloodTypeDistribution[stat.BloodType] = stat.Count
	}
	
	// 年龄分布统计
	var ageStats []struct {
		AgeGroup string
		Count    int64
	}
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Select(`
			CASE 
				WHEN EXTRACT(YEAR FROM AGE(date_of_birth)) < 18 THEN 'under_18'
				WHEN EXTRACT(YEAR FROM AGE(date_of_birth)) BETWEEN 18 AND 30 THEN '18_30'
				WHEN EXTRACT(YEAR FROM AGE(date_of_birth)) BETWEEN 31 AND 50 THEN '31_50'
				WHEN EXTRACT(YEAR FROM AGE(date_of_birth)) BETWEEN 51 AND 65 THEN '51_65'
				ELSE 'over_65'
			END as age_group,
			COUNT(*) as count
		`).
		Group("age_group").
		Scan(&ageStats).Error
	if err != nil {
		return nil, err
	}
	
	stats.AgeDistribution = make(map[string]int64)
	for _, stat := range ageStats {
		stats.AgeDistribution[stat.AgeGroup] = stat.Count
	}
	
	// 有过敏史的用户数量
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("allergies IS NOT NULL AND allergies != ''").
		Count(&stats.ProfilesWithAllergies).Error
	if err != nil {
		return nil, err
	}
	
	// 正在服药的用户数量
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("medications IS NOT NULL AND medications != ''").
		Count(&stats.ProfilesWithMedications).Error
	if err != nil {
		return nil, err
	}
	
	// 有病史的用户数量
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("medical_history IS NOT NULL AND medical_history != ''").
		Count(&stats.ProfilesWithMedicalHistory).Error
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}