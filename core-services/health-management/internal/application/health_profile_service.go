package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthProfileService 
type HealthProfileService struct {
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthProfileService 
func NewHealthProfileService(healthProfileRepo domain.HealthProfileRepository, eventPublisher EventPublisher) *HealthProfileService {
	return &HealthProfileService{
		healthProfileRepo: healthProfileRepo,
		eventPublisher:    eventPublisher,
	}
}

// CreateHealthProfileRequest 
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

// CreateHealthProfileResponse 
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

// CreateHealthProfile 
func (s *HealthProfileService) CreateHealthProfile(ctx context.Context, req *CreateHealthProfileRequest) (*CreateHealthProfileResponse, error) {
	// ?
	exists, err := s.healthProfileRepo.ExistsByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}
	
	if exists {
		return nil, fmt.Errorf("health profile already exists for user %s", req.UserID)
	}
	
	// ?
	profile := domain.NewHealthProfile(req.UserID, req.Gender)
	
	// 
	if req.DateOfBirth != nil || req.Height != nil || req.BloodType != nil {
		profile.UpdateBasicInfo(req.Gender, req.DateOfBirth, req.Height, req.BloodType)
	}
	
	// 
	if req.EmergencyContact != "" && req.EmergencyName != "" {
		profile.UpdateEmergencyContact(req.EmergencyName, req.EmergencyContact)
	}
	
	// 
	if req.MedicalHistory != nil {
		for _, condition := range req.MedicalHistory {
			profile.AddMedicalHistory(condition)
		}
	}
	
	// ?
	if req.Allergies != nil {
		for _, allergen := range req.Allergies {
			profile.AddAllergy(allergen)
		}
	}
	
	// ?
	if req.Medications != nil {
		for _, medication := range req.Medications {
			profile.AddMedication(medication)
		}
	}
	
	// 
	if req.HealthGoals != nil {
		profile.SetHealthGoals(req.HealthGoals)
	}
	
	// 
	if req.PreferredUnits != nil {
		profile.UpdatePreferredUnits(req.PreferredUnits)
	}
	
	// 
	if req.NotificationPrefs != nil {
		profile.UpdateNotificationPrefs(req.NotificationPrefs)
	}
	
	// 浽?
	if err := s.healthProfileRepo.Save(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save health profile: %w", err)
	}
	
	// 
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// ?
		// TODO: 
	}
	
	// 
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

// GetHealthProfileRequest 
type GetHealthProfileRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// GetHealthProfile 
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

// UpdateHealthProfileRequest 
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

// UpdateHealthProfile 
func (s *HealthProfileService) UpdateHealthProfile(ctx context.Context, req *UpdateHealthProfileRequest) (*CreateHealthProfileResponse, error) {
	// 
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}
	
	if profile == nil {
		return nil, fmt.Errorf("health profile not found for user %s", req.UserID)
	}
	
	// 
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
	
	// 
	if req.EmergencyContact != nil && req.EmergencyName != nil {
		profile.UpdateEmergencyContact(*req.EmergencyName, *req.EmergencyContact)
	}
	
	// 
	if req.PreferredUnits != nil {
		profile.UpdatePreferredUnits(req.PreferredUnits)
	}
	
	// 
	if req.NotificationPrefs != nil {
		profile.UpdateNotificationPrefs(req.NotificationPrefs)
	}
	
	// 
	if err := s.healthProfileRepo.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to update health profile: %w", err)
	}
	
	// 
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// ?
		// TODO: 
	}
	
	// 
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

// AddMedicalHistoryRequest 
type AddMedicalHistoryRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Condition string    `json:"condition" validate:"required"`
}

// AddMedicalHistory 
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
	
	// 
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// ?
		// TODO: 
	}
	
	profile.ClearEvents()
	
	return nil
}

// RemoveMedicalHistoryRequest 
type RemoveMedicalHistoryRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Condition string    `json:"condition" validate:"required"`
}

// RemoveMedicalHistory 
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
	
	// 
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// ?
		// TODO: 
	}
	
	profile.ClearEvents()
	
	return nil
}

// AddAllergyRequest ?
type AddAllergyRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Allergen string    `json:"allergen" validate:"required"`
}

// AddAllergy ?
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
	
	// 
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// ?
		// TODO: 
	}
	
	profile.ClearEvents()
	
	return nil
}

// RemoveAllergyRequest ?
type RemoveAllergyRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Allergen string    `json:"allergen" validate:"required"`
}

// RemoveAllergy ?
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
	
	// 
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// ?
		// TODO: 
	}
	
	profile.ClearEvents()
	
	return nil
}

// SetHealthGoalsRequest 
type SetHealthGoalsRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Goals  []string  `json:"goals" validate:"required"`
}

// SetHealthGoals 
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
	
	// 
	if err := s.publishEvents(ctx, profile.GetEvents()); err != nil {
		// ?
		// TODO: 
	}
	
	profile.ClearEvents()
	
	return nil
}

// CalculateBMIRequest BMI
type CalculateBMIRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Weight float64   `json:"weight" validate:"required,gt=0"`
}

// CalculateBMIResponse BMI
type CalculateBMIResponse struct {
	BMI        *float64 `json:"bmi"`
	Category   string   `json:"category"`
	IsHealthy  bool     `json:"is_healthy"`
	Suggestion string   `json:"suggestion"`
}

// CalculateBMI BMI
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
			Category:   "",
			IsHealthy:  false,
			Suggestion: "",
		}, nil
	}
	
	// BMI
	var category, suggestion string
	var isHealthy bool
	
	switch {
	case *bmi < 18.5:
		category = ""
		isHealthy = false
		suggestion = ""
	case *bmi >= 18.5 && *bmi < 24:
		category = ""
		isHealthy = true
		suggestion = "?
	case *bmi >= 24 && *bmi < 28:
		category = ""
		isHealthy = false
		suggestion = "?
	default:
		category = ""
		isHealthy = false
		suggestion = "?
	}
	
	return &CalculateBMIResponse{
		BMI:        bmi,
		Category:   category,
		IsHealthy:  isHealthy,
		Suggestion: suggestion,
	}, nil
}

// publishEvents 
func (s *HealthProfileService) publishEvents(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.GetEventType(), err)
		}
	}
	return nil
}

