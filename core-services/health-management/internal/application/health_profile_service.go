package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthProfileService 健康档案应用服务
type HealthProfileService struct {
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthProfileService 创建健康档案服务
func NewHealthProfileService(healthProfileRepo domain.HealthProfileRepository, eventPublisher EventPublisher) *HealthProfileService {
	return &HealthProfileService{
		healthProfileRepo: healthProfileRepo,
		eventPublisher:    eventPublisher,
	}
}

// CreateHealthProfileRequest 创建健康档案请求
type CreateHealthProfileRequest struct {
	UserID            uuid.UUID          `json:"user_id" validate:"required"`
	Gender            domain.Gender      `json:"gender" validate:"required"`
	DateOfBirth       *time.Time         `json:"date_of_birth,omitempty"`
	Height            *float64           `json:"height,omitempty" validate:"omitempty,gt=0"`
	BloodType         *domain.BloodType  `json:"blood_type,omitempty"`
	EmergencyContact  string             `json:"emergency_contact,omitempty"`
	EmergencyName     string             `json:"emergency_name,omitempty"`
	MedicalHistory    []string           `json:"medical_history,omitempty"`
	Allergies         []string           `json:"allergies,omitempty"`
	Medications       []string           `json:"medications,omitempty"`
	HealthGoals       []string           `json:"health_goals,omitempty"`
	PreferredUnits    map[string]string  `json:"preferred_units,omitempty"`
	NotificationPrefs map[string]bool    `json:"notification_prefs,omitempty"`
}

// CreateHealthProfileResponse 创建健康档案响应
type CreateHealthProfileResponse struct {
	ID                uuid.UUID          `json:"id"`
	UserID            uuid.UUID          `json:"user_id"`
	Gender            domain.Gender      `json:"gender"`
	DateOfBirth       *time.Time         `json:"date_of_birth,omitempty"`
	Age               *int               `json:"age,omitempty"`
	Height            *float64           `json:"height,omitempty"`
	BloodType         *domain.BloodType  `json:"blood_type,omitempty"`
	EmergencyContact  string             `json:"emergency_contact,omitempty"`
	EmergencyName     string             `json:"emergency_name,omitempty"`
	MedicalHistory    []string           `json:"medical_history,omitempty"`
	Allergies         []string           `json:"allergies,omitempty"`
	Medications       []string           `json:"medications,omitempty"`
	HealthGoals       []string           `json:"health_goals,omitempty"`
	PreferredUnits    map[string]string  `json:"preferred_units,omitempty"`
	NotificationPrefs map[string]bool    `json:"notification_prefs,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// CreateHealthProfile 创建健康档案
func (s *HealthProfileService) CreateHealthProfile(ctx context.Context, req *CreateHealthProfileRequest) (*CreateHealthProfileResponse, error) {
	// 检查用户是否已有健康档案
	exists, err := s.healthProfileRepo.ExistsByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}
	
	if exists {
		return nil, fmt.Errorf("health profile already exists for user %s", req.UserID)
	}
	
	// 创建健康档案聚合根
	profile := domain.NewHealthProfile(req.UserID, req.Gender)
	
	// 设置基本信息
	if req.DateOfBirth != nil || req.Height != nil || req.BloodType != nil {
		profile.UpdateBasicInfo(req.Gender, req.DateOfBirth, req.Height, req.BloodType)
	}
	
	// 设置紧急联系人
	if req.EmergencyContact != "" && req.EmergencyName != "" {
		profile.UpdateEmergencyContact(req.EmergencyName, req.EmergencyContact)
	}
	
	// 设置病史
	if req.MedicalHistory != nil {
		for _, condition := range req.MedicalHistory {
			profile.AddMedicalHistory(condition)
		}
	}
	
	// 设置过敏史
	if req.Allergies != nil {
		for _, allergen := range req.Allergies {
			profile.AddAllergy(allergen)
		}
	}
	
	// 设置用药史
	if req.Medications != nil {
		for _, medication := range req.Medications {
			profile.AddMedication(medication)
		}
	}
	
	// 设置健康目标
	if req.HealthGoals != nil {
		profile.SetHealthGoals(req.HealthGoals)
	}
	
	// 设置偏好单位
	if req.PreferredUnits != nil {
		profile.UpdatePreferredUnits(req.PreferredUnits)
	}
	
	// 设置通知偏好
	if req.NotificationPrefs != nil {
		profile.UpdateNotificationPrefs(req.NotificationPrefs)
	}
	
	// 保存到仓储
	if err := s.healthProfileRepo.Save(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save health profile: %w", err)
	}
	
	// 发布领域事件
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// 记录日志但不影响主流程
		// TODO: 添加日志记录
	}
	
	// 清除事件
	profile.ClearEvents()
	
	return &CreateHealthProfileResponse{
		ID:                profile.ID,
		UserID:            profile.UserID,
		Gender:            profile.Gender,
		DateOfBirth:       profile.DateOfBirth,
		Age:               profile.GetAge(),
		Height:            profile.Height,
		BloodType:         profile.BloodType,
		EmergencyContact:  profile.EmergencyContact,
		EmergencyName:     profile.EmergencyName,
		MedicalHistory:    profile.MedicalHistory,
		Allergies:         profile.Allergies,
		Medications:       profile.Medications,
		HealthGoals:       profile.HealthGoals,
		PreferredUnits:    profile.PreferredUnits,
		NotificationPrefs: profile.NotificationPrefs,
		CreatedAt:         profile.CreatedAt,
		UpdatedAt:         profile.UpdatedAt,
	}, nil
}

// GetHealthProfileRequest 获取健康档案请求
type GetHealthProfileRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// GetHealthProfile 获取健康档案
func (s *HealthProfileService) GetHealthProfile(ctx context.Context, req *GetHealthProfileRequest) (*CreateHealthProfileResponse, error) {
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return nil, nil
	}
	
	return &CreateHealthProfileResponse{
		ID:                profile.ID,
		UserID:            profile.UserID,
		Gender:            profile.Gender,
		DateOfBirth:       profile.DateOfBirth,
		Age:               profile.GetAge(),
		Height:            profile.Height,
		BloodType:         profile.BloodType,
		EmergencyContact:  profile.EmergencyContact,
		EmergencyName:     profile.EmergencyName,
		MedicalHistory:    profile.MedicalHistory,
		Allergies:         profile.Allergies,
		Medications:       profile.Medications,
		HealthGoals:       profile.HealthGoals,
		PreferredUnits:    profile.PreferredUnits,
		NotificationPrefs: profile.NotificationPrefs,
		CreatedAt:         profile.CreatedAt,
		UpdatedAt:         profile.UpdatedAt,
	}, nil
}

// UpdateHealthProfileRequest 更新健康档案请求
type UpdateHealthProfileRequest struct {
	UserID            uuid.UUID          `json:"user_id" validate:"required"`
	Gender            *domain.Gender     `json:"gender,omitempty"`
	DateOfBirth       *time.Time         `json:"date_of_birth,omitempty"`
	Height            *float64           `json:"height,omitempty" validate:"omitempty,gt=0"`
	BloodType         *domain.BloodType  `json:"blood_type,omitempty"`
	EmergencyContact  *string            `json:"emergency_contact,omitempty"`
	EmergencyName     *string            `json:"emergency_name,omitempty"`
	PreferredUnits    map[string]string  `json:"preferred_units,omitempty"`
	NotificationPrefs map[string]bool    `json:"notification_prefs,omitempty"`
}

// UpdateHealthProfile 更新健康档案
func (s *HealthProfileService) UpdateHealthProfile(ctx context.Context, req *UpdateHealthProfileRequest) (*CreateHealthProfileResponse, error) {
	// 获取现有档案
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return nil, fmt.Errorf("health profile not found for user %s", req.UserID)
	}
	
	// 更新基本信息
	if req.Gender != nil || req.DateOfBirth != nil || req.Height != nil || req.BloodType != nil {
		gender := profile.Gender
		if req.Gender != nil {
			gender = *req.Gender
		}
		
		dateOfBirth := profile.DateOfBirth
		if req.DateOfBirth != nil {
			dateOfBirth = req.DateOfBirth
		}
		
		height := profile.Height
		if req.Height != nil {
			height = req.Height
		}
		
		bloodType := profile.BloodType
		if req.BloodType != nil {
			bloodType = req.BloodType
		}
		
		profile.UpdateBasicInfo(gender, dateOfBirth, height, bloodType)
	}
	
	// 更新紧急联系人
	if req.EmergencyContact != nil && req.EmergencyName != nil {
		profile.UpdateEmergencyContact(*req.EmergencyName, *req.EmergencyContact)
	}
	
	// 更新偏好单位
	if req.PreferredUnits != nil {
		profile.UpdatePreferredUnits(req.PreferredUnits)
	}
	
	// 更新通知偏好
	if req.NotificationPrefs != nil {
		profile.UpdateNotificationPrefs(req.NotificationPrefs)
	}
	
	// 保存更新
	if err := s.healthProfileRepo.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to update health profile: %w", err)
	}
	
	// 发布领域事件
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// 记录日志但不影响主流程
		// TODO: 添加日志记录
	}
	
	// 清除事件
	profile.ClearEvents()
	
	return &CreateHealthProfileResponse{
		ID:                profile.ID,
		UserID:            profile.UserID,
		Gender:            profile.Gender,
		DateOfBirth:       profile.DateOfBirth,
		Age:               profile.GetAge(),
		Height:            profile.Height,
		BloodType:         profile.BloodType,
		EmergencyContact:  profile.EmergencyContact,
		EmergencyName:     profile.EmergencyName,
		MedicalHistory:    profile.MedicalHistory,
		Allergies:         profile.Allergies,
		Medications:       profile.Medications,
		HealthGoals:       profile.HealthGoals,
		PreferredUnits:    profile.PreferredUnits,
		NotificationPrefs: profile.NotificationPrefs,
		CreatedAt:         profile.CreatedAt,
		UpdatedAt:         profile.UpdatedAt,
	}, nil
}

// AddMedicalHistoryRequest 添加病史请求
type AddMedicalHistoryRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Condition string    `json:"condition" validate:"required"`
}

// AddMedicalHistory 添加病史
func (s *HealthProfileService) AddMedicalHistory(ctx context.Context, req *AddMedicalHistoryRequest) error {
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return fmt.Errorf("health profile not found for user %s", req.UserID)
	}
	
	profile.AddMedicalHistory(req.Condition)
	
	if err := s.healthProfileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update health profile: %w", err)
	}
	
	// 发布领域事件
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// 记录日志但不影响主流程
		// TODO: 添加日志记录
	}
	
	profile.ClearEvents()
	
	return nil
}

// RemoveMedicalHistoryRequest 移除病史请求
type RemoveMedicalHistoryRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Condition string    `json:"condition" validate:"required"`
}

// RemoveMedicalHistory 移除病史
func (s *HealthProfileService) RemoveMedicalHistory(ctx context.Context, req *RemoveMedicalHistoryRequest) error {
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return fmt.Errorf("health profile not found for user %s", req.UserID)
	}
	
	profile.RemoveMedicalHistory(req.Condition)
	
	if err := s.healthProfileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update health profile: %w", err)
	}
	
	// 发布领域事件
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// 记录日志但不影响主流程
		// TODO: 添加日志记录
	}
	
	profile.ClearEvents()
	
	return nil
}

// AddAllergyRequest 添加过敏史请求
type AddAllergyRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Allergen string    `json:"allergen" validate:"required"`
}

// AddAllergy 添加过敏史
func (s *HealthProfileService) AddAllergy(ctx context.Context, req *AddAllergyRequest) error {
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return fmt.Errorf("health profile not found for user %s", req.UserID)
	}
	
	profile.AddAllergy(req.Allergen)
	
	if err := s.healthProfileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update health profile: %w", err)
	}
	
	// 发布领域事件
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// 记录日志但不影响主流程
		// TODO: 添加日志记录
	}
	
	profile.ClearEvents()
	
	return nil
}

// RemoveAllergyRequest 移除过敏史请求
type RemoveAllergyRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Allergen string    `json:"allergen" validate:"required"`
}

// RemoveAllergy 移除过敏史
func (s *HealthProfileService) RemoveAllergy(ctx context.Context, req *RemoveAllergyRequest) error {
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return fmt.Errorf("health profile not found for user %s", req.UserID)
	}
	
	profile.RemoveAllergy(req.Allergen)
	
	if err := s.healthProfileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update health profile: %w", err)
	}
	
	// 发布领域事件
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// 记录日志但不影响主流程
		// TODO: 添加日志记录
	}
	
	profile.ClearEvents()
	
	return nil
}

// SetHealthGoalsRequest 设置健康目标请求
type SetHealthGoalsRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Goals  []string  `json:"goals" validate:"required"`
}

// SetHealthGoals 设置健康目标
func (s *HealthProfileService) SetHealthGoals(ctx context.Context, req *SetHealthGoalsRequest) error {
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return fmt.Errorf("health profile not found for user %s", req.UserID)
	}
	
	profile.SetHealthGoals(req.Goals)
	
	if err := s.healthProfileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update health profile: %w", err)
	}
	
	// 发布领域事件
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// 记录日志但不影响主流程
		// TODO: 添加日志记录
	}
	
	profile.ClearEvents()
	
	return nil
}

// CalculateBMIRequest 计算BMI请求
type CalculateBMIRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Weight float64   `json:"weight" validate:"required,gt=0"`
}

// CalculateBMIResponse 计算BMI响应
type CalculateBMIResponse struct {
	BMI        *float64 `json:"bmi"`
	Category   string   `json:"category"`
	IsHealthy  bool     `json:"is_healthy"`
	Suggestion string   `json:"suggestion"`
}

// CalculateBMI 计算BMI
func (s *HealthProfileService) CalculateBMI(ctx context.Context, req *CalculateBMIRequest) (*CalculateBMIResponse, error) {
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return nil, fmt.Errorf("health profile not found for user %s", req.UserID)
	}
	
	bmi := profile.GetBMI(req.Weight)
	if bmi == nil {
		return &CalculateBMIResponse{
			BMI:        nil,
			Category:   "无法计算",
			IsHealthy:  false,
			Suggestion: "请先设置身高信息",
		}, nil
	}
	
	// 根据BMI值确定分类和建议
	var category, suggestion string
	var isHealthy bool
	
	switch {
	case *bmi < 18.5:
		category = "偏瘦"
		isHealthy = false
		suggestion = "建议增加营养摄入，适当增重"
	case *bmi >= 18.5 && *bmi < 24:
		category = "正常"
		isHealthy = true
		suggestion = "保持良好的生活习惯"
	case *bmi >= 24 && *bmi < 28:
		category = "超重"
		isHealthy = false
		suggestion = "建议控制饮食，增加运动"
	default:
		category = "肥胖"
		isHealthy = false
		suggestion = "建议咨询医生，制定减重计划"
	}
	
	return &CalculateBMIResponse{
		BMI:        bmi,
		Category:   category,
		IsHealthy:  isHealthy,
		Suggestion: suggestion,
	}, nil
}

// publishEvents 发布领域事件
func (s *HealthProfileService) publishEvents(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.GetEventType(), err)
		}
	}
	return nil
}