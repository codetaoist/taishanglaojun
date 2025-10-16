package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/taishanglaojun/health-management/internal/domain"
)

// PostgreSQLHealthProfileRepository PostgreSQL
type PostgreSQLHealthProfileRepository struct {
	db *gorm.DB
}

// NewPostgreSQLHealthProfileRepository PostgreSQL
func NewPostgreSQLHealthProfileRepository(db *gorm.DB) domain.HealthProfileRepository {
	return &PostgreSQLHealthProfileRepository{
		db: db,
	}
}

// Save 潡
func (r *PostgreSQLHealthProfileRepository) Save(ctx context.Context, profile *domain.HealthProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

// FindByID ID
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

// FindByUserID ID
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

// Update 
func (r *PostgreSQLHealthProfileRepository) Update(ctx context.Context, profile *domain.HealthProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// Delete 
func (r *PostgreSQLHealthProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.HealthProfile{}).Error
}

// List 
func (r *PostgreSQLHealthProfileRepository) List(ctx context.Context, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error
	return profiles, err
}

// Count 
func (r *PostgreSQLHealthProfileRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Count(&count).Error
	return count, err
}

// FindByGender 
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

// FindByAgeRange 
func (r *PostgreSQLHealthProfileRepository) FindByAgeRange(ctx context.Context, minAge, maxAge int, limit, offset int) ([]*domain.HealthProfile, error) {
	var profiles []*domain.HealthProfile
	
	// 
	// 
	query := r.db.WithContext(ctx).
		Where("EXTRACT(YEAR FROM AGE(date_of_birth)) BETWEEN ? AND ?", minAge, maxAge).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)
	
	err := query.Find(&profiles).Error
	return profiles, err
}

// FindByBloodType ?
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

// CountByGender ?
func (r *PostgreSQLHealthProfileRepository) CountByGender(ctx context.Context, gender domain.Gender) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("gender = ?", gender).
		Count(&count).Error
	return count, err
}

// CountByAgeRange ?
func (r *PostgreSQLHealthProfileRepository) CountByAgeRange(ctx context.Context, minAge, maxAge int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("EXTRACT(YEAR FROM AGE(date_of_birth)) BETWEEN ? AND ?", minAge, maxAge).
		Count(&count).Error
	return count, err
}

// CountByBloodType 
func (r *PostgreSQLHealthProfileRepository) CountByBloodType(ctx context.Context, bloodType domain.BloodType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("blood_type = ?", bloodType).
		Count(&count).Error
	return count, err
}

// FindWithAllergies ?
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

// FindWithMedications ?
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

// FindWithMedicalHistory 
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

// SearchByKeyword ?
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

// GetStatistics 
func (r *PostgreSQLHealthProfileRepository) GetStatistics(ctx context.Context) (*domain.HealthProfileStatistics, error) {
	var stats domain.HealthProfileStatistics
	
	// 
	err := r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Count(&stats.TotalProfiles).Error
	if err != nil {
		return nil, err
	}
	
	// 
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
	
	// ?
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
	
	// 
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
	
	// ?
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("allergies IS NOT NULL AND allergies != ''").
		Count(&stats.ProfilesWithAllergies).Error
	if err != nil {
		return nil, err
	}
	
	// ?
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("medications IS NOT NULL AND medications != ''").
		Count(&stats.ProfilesWithMedications).Error
	if err != nil {
		return nil, err
	}
	
	// 
	err = r.db.WithContext(ctx).
		Model(&domain.HealthProfile{}).
		Where("medical_history IS NOT NULL AND medical_history != ''").
		Count(&stats.ProfilesWithMedicalHistory).Error
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}

