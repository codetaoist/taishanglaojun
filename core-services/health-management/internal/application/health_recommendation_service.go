package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthRecommendationService 
type HealthRecommendationService struct {
	healthDataRepo    domain.HealthDataRepository
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthRecommendationService 
func NewHealthRecommendationService(
	healthDataRepo domain.HealthDataRepository,
	healthProfileRepo domain.HealthProfileRepository,
	eventPublisher EventPublisher,
) *HealthRecommendationService {
	return &HealthRecommendationService{
		healthDataRepo:    healthDataRepo,
		healthProfileRepo: healthProfileRepo,
		eventPublisher:    eventPublisher,
	}
}

// RecommendationType 
type RecommendationType string

const (
	RecommendationTypeExercise   RecommendationType = "exercise"   // 
	RecommendationTypeDiet       RecommendationType = "diet"       // 
	RecommendationTypeSleep      RecommendationType = "sleep"      // 
	RecommendationTypeStress     RecommendationType = "stress"     // 
	RecommendationTypeMedical    RecommendationType = "medical"    // 
	RecommendationTypeLifestyle  RecommendationType = "lifestyle"  // 
	RecommendationTypePrevention RecommendationType = "prevention" // 
)

// RecommendationPriority ?
type RecommendationPriority string

const (
	RecommendationPriorityHigh   RecommendationPriority = "high"   // 
	RecommendationPriorityMedium RecommendationPriority = "medium" // 
	RecommendationPriorityLow    RecommendationPriority = "low"    // 
)

// HealthRecommendation 
type HealthRecommendation struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Type        RecommendationType     `json:"type"`
	Priority    RecommendationPriority `json:"priority"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Actions     []string               `json:"actions"`
	Benefits    []string               `json:"benefits"`
	Duration    string                 `json:"duration,omitempty"`
	Frequency   string                 `json:"frequency,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateRecommendationsRequest 
type GenerateRecommendationsRequest struct {
	UserID uuid.UUID            `json:"user_id" binding:"required"`
	Types  []RecommendationType `json:"types,omitempty"`
	Days   int                  `json:"days,omitempty"`
	Limit  int                  `json:"limit,omitempty"`
}

// GenerateRecommendationsResponse 
type GenerateRecommendationsResponse struct {
	Recommendations []HealthRecommendation `json:"recommendations"`
	Summary         string                 `json:"summary"`
	Count           int                    `json:"count"`
	Timestamp       time.Time              `json:"timestamp"`
}

// GetPersonalizedTipsRequest 
type GetPersonalizedTipsRequest struct {
	UserID   uuid.UUID            `json:"user_id" binding:"required"`
	Category RecommendationType   `json:"category,omitempty"`
	Limit    int                  `json:"limit,omitempty"`
}

// GetPersonalizedTipsResponse 
type GetPersonalizedTipsResponse struct {
	Tips      []HealthTip `json:"tips"`
	Category  string      `json:"category"`
	Count     int         `json:"count"`
	Timestamp time.Time   `json:"timestamp"`
}

// HealthTip 
type HealthTip struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	Difficulty  string    `json:"difficulty"`
	EstimatedTime string  `json:"estimated_time,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

// GenerateRecommendations 
func (s *HealthRecommendationService) GenerateRecommendations(ctx context.Context, req *GenerateRecommendationsRequest) (*GenerateRecommendationsResponse, error) {
	// ?
	if req.Days == 0 {
		req.Days = 30
	}
	if req.Limit == 0 {
		req.Limit = 10
	}
	if len(req.Types) == 0 {
		req.Types = []RecommendationType{
			RecommendationTypeExercise,
			RecommendationTypeDiet,
			RecommendationTypeSleep,
			RecommendationTypeStress,
		}
	}

	// 
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// 
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -req.Days)
	
	var recommendations []HealthRecommendation

	// ?
	for _, recType := range req.Types {
		typeRecommendations, err := s.generateRecommendationsByType(ctx, req.UserID, recType, startTime, endTime, profile)
		if err != nil {
			continue // 
		}
		recommendations = append(recommendations, typeRecommendations...)
	}

	// 
	sort.Slice(recommendations, func(i, j int) bool {
		priorityOrder := map[RecommendationPriority]int{
			RecommendationPriorityHigh:   3,
			RecommendationPriorityMedium: 2,
			RecommendationPriorityLow:    1,
		}
		return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
	})

	// 
	if len(recommendations) > req.Limit {
		recommendations = recommendations[:req.Limit]
	}

	// 
	summary := s.generateRecommendationSummary(recommendations)

	return &GenerateRecommendationsResponse{
		Recommendations: recommendations,
		Summary:         summary,
		Count:           len(recommendations),
		Timestamp:       time.Now(),
	}, nil
}

// GetPersonalizedTips 
func (s *HealthRecommendationService) GetPersonalizedTips(ctx context.Context, req *GetPersonalizedTipsRequest) (*GetPersonalizedTipsResponse, error) {
	// ?
	if req.Limit == 0 {
		req.Limit = 5
	}

	// 
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// 
	tips := s.generatePersonalizedTips(req.UserID, req.Category, req.Limit, profile)

	category := "general"
	if req.Category != "" {
		category = string(req.Category)
	}

	return &GetPersonalizedTipsResponse{
		Tips:      tips,
		Category:  category,
		Count:     len(tips),
		Timestamp: time.Now(),
	}, nil
}

// generateRecommendationsByType 
func (s *HealthRecommendationService) generateRecommendationsByType(ctx context.Context, userID uuid.UUID, recType RecommendationType, startTime, endTime time.Time, profile *domain.HealthProfile) ([]HealthRecommendation, error) {
	var recommendations []HealthRecommendation

	switch recType {
	case RecommendationTypeExercise:
		recommendations = s.generateExerciseRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeDiet:
		recommendations = s.generateDietRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeSleep:
		recommendations = s.generateSleepRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeStress:
		recommendations = s.generateStressRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeMedical:
		recommendations = s.generateMedicalRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypeLifestyle:
		recommendations = s.generateLifestyleRecommendations(ctx, userID, startTime, endTime, profile)
	case RecommendationTypePrevention:
		recommendations = s.generatePreventionRecommendations(ctx, userID, startTime, endTime, profile)
	}

	return recommendations, nil
}

// generateExerciseRecommendations 
func (s *HealthRecommendationService) generateExerciseRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	// 
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	
	avgSteps := s.calculateAverageSteps(stepsData)
	age := 30
	if profile != nil {
		age = profile.GetAge()
	}

	// 
	if avgSteps < 5000 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeExercise,
			Priority:    RecommendationPriorityHigh,
			Title:       "?,
			Description: "?,
			Actions: []string{
				"30",
				"",
				"15-20",
				"?,
			},
			Benefits: []string{
				"?,
				"",
				"",
				"",
			},
			Duration:  "",
			Frequency: "",
			Tags:      []string{"", "", "?},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	} else if avgSteps < 8000 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeExercise,
			Priority:    RecommendationPriorityMedium,
			Title:       "",
			Description: "",
			Actions: []string{
				"3?0?,
				"?,
				"",
				"",
			},
			Benefits: []string{
				"",
				"",
				"?,
				"?,
			},
			Duration:  "30-45",
			Frequency: "3-4?,
			Tags:      []string{"", "", ""},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	// 
	if age > 50 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeExercise,
			Priority:    RecommendationPriorityMedium,
			Title:       "?,
			Description: "?,
			Actions: []string{
				"",
				"",
				"",
				"?,
			},
			Benefits: []string{
				"",
				"?,
				"?,
				"",
			},
			Duration:  "30-45",
			Frequency: "3-5?,
			Tags:      []string{"", "", "", ""},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	return recommendations
}

// generateDietRecommendations 
func (s *HealthRecommendationService) generateDietRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	// ?
	bloodSugarData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "blood_sugar", startTime, endTime)
	avgBloodSugar := s.calculateAverageValue(bloodSugarData)

	// ?
	if avgBloodSugar > 7.0 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeDiet,
			Priority:    RecommendationPriorityHigh,
			Title:       "?,
			Description: "?,
			Actions: []string{
				"GI",
				"",
				"",
				"?,
			},
			Benefits: []string{
				"?,
				"?,
				"?,
				"",
			},
			Duration:  "",
			Frequency: "",
			Tags:      []string{"?, "GI", ""},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	// BMI
	if profile != nil {
		bmi := profile.GetBMI()
		if bmi > 25 {
			recommendations = append(recommendations, HealthRecommendation{
				ID:          uuid.New(),
				UserID:      userID,
				Type:        RecommendationTypeDiet,
				Priority:    RecommendationPriorityHigh,
				Title:       "",
				Description: "BMI",
				Actions: []string{
					"?,
					"?,
					"?,
					"?,
				},
				Benefits: []string{
					"",
					"",
					"?,
					"",
				},
				Duration:  "3-6",
				Frequency: "",
				Tags:      []string{"", "", ""},
				CreatedAt: time.Now(),
				IsActive:  true,
			})
		}
	}

	return recommendations
}

// generateSleepRecommendations 
func (s *HealthRecommendationService) generateSleepRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	// 
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)
	avgSleep := s.calculateAverageValue(sleepData)

	// 
	if avgSleep < 7 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeSleep,
			Priority:    RecommendationPriorityHigh,
			Title:       "",
			Description: "䲻㽨?,
			Actions: []string{
				"?,
				"1豸",
				"",
				"",
			},
			Benefits: []string{
				"?,
				"?,
				"",
				"",
			},
			Duration:  "",
			Frequency: "",
			Tags:      []string{"", "", ""},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	return recommendations
}

// generateStressRecommendations 
func (s *HealthRecommendationService) generateStressRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	// 
	stressData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "stress_level", startTime, endTime)
	avgStress := s.calculateAverageValue(stressData)

	// 
	if avgStress > 7 {
		recommendations = append(recommendations, HealthRecommendation{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        RecommendationTypeStress,
			Priority:    RecommendationPriorityHigh,
			Title:       "",
			Description: "?,
			Actions: []string{
				"",
				"",
				"",
				"",
			},
			Benefits: []string{
				"",
				"",
				"",
				"",
			},
			Duration:  "15-30",
			Frequency: "",
			Tags:      []string{"", "", "", ""},
			CreatedAt: time.Now(),
			IsActive:  true,
		})
	}

	return recommendations
}

// generateMedicalRecommendations 
func (s *HealthRecommendationService) generateMedicalRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	if profile != nil {
		age := profile.GetAge()
		
		// 
		if age > 40 {
			recommendations = append(recommendations, HealthRecommendation{
				ID:          uuid.New(),
				UserID:      userID,
				Type:        RecommendationTypeMedical,
				Priority:    RecommendationPriorityMedium,
				Title:       "",
				Description: "鶨緢",
				Actions: []string{
					"",
					"?,
					"",
					"?,
				},
				Benefits: []string{
					"",
					"",
					"",
					"",
				},
				Duration:  "1-2",
				Frequency: "",
				Tags:      []string{"", "", "", ""},
				CreatedAt: time.Now(),
				IsActive:  true,
			})
		}
	}

	return recommendations
}

// generateLifestyleRecommendations 
func (s *HealthRecommendationService) generateLifestyleRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	recommendations = append(recommendations, HealthRecommendation{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        RecommendationTypeLifestyle,
		Priority:    RecommendationPriorityMedium,
		Title:       "",
		Description: "",
		Actions: []string{
			"",
			"?,
			"",
			"罻",
		},
		Benefits: []string{
			"",
			"?,
			"",
			"?,
		},
		Duration:  "",
		Frequency: "",
		Tags:      []string{"", "", "罻", ""},
		CreatedAt: time.Now(),
		IsActive:  true,
	})

	return recommendations
}

// generatePreventionRecommendations 
func (s *HealthRecommendationService) generatePreventionRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) []HealthRecommendation {
	var recommendations []HealthRecommendation

	recommendations = append(recommendations, HealthRecommendation{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        RecommendationTypePrevention,
		Priority:    RecommendationPriorityMedium,
		Title:       "",
		Description: "?,
		Actions: []string{
			"",
			"",
			"",
			"",
		},
		Benefits: []string{
			"",
			"",
			"?,
			"",
		},
		Duration:  "",
		Frequency: "?,
		Tags:      []string{"", "", "", ""},
		CreatedAt: time.Now(),
		IsActive:  true,
	})

	return recommendations
}

// generatePersonalizedTips 
func (s *HealthRecommendationService) generatePersonalizedTips(userID uuid.UUID, category RecommendationType, limit int, profile *domain.HealthProfile) []HealthTip {
	var tips []HealthTip

	// ?
	switch category {
	case RecommendationTypeExercise:
		tips = s.getExerciseTips(limit)
	case RecommendationTypeDiet:
		tips = s.getDietTips(limit)
	case RecommendationTypeSleep:
		tips = s.getSleepTips(limit)
	case RecommendationTypeStress:
		tips = s.getStressTips(limit)
	default:
		tips = s.getGeneralTips(limit)
	}

	return tips
}

// getExerciseTips 
func (s *HealthRecommendationService) getExerciseTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "10000?,
			Content:       "?0000?,
			Category:      "exercise",
			Difficulty:    "easy",
			EstimatedTime: "60-90",
			Tags:          []string{"", "?, ""},
		},
		{
			ID:            uuid.New(),
			Title:         "?,
			Content:       "2-3?,
			Category:      "exercise",
			Difficulty:    "medium",
			EstimatedTime: "30-45",
			Tags:          []string{"", "", "?},
		},
		{
			ID:            uuid.New(),
			Title:         "?,
			Content:       "?,
			Category:      "exercise",
			Difficulty:    "easy",
			EstimatedTime: "10-15",
			Tags:          []string{"", "?, ""},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// getDietTips 
func (s *HealthRecommendationService) getDietTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "",
			Content:       "5?,
			Category:      "diet",
			Difficulty:    "easy",
			EstimatedTime: "",
			Tags:          []string{"", "", ""},
		},
		{
			ID:            uuid.New(),
			Title:         "",
			Content:       "??,
			Category:      "diet",
			Difficulty:    "medium",
			EstimatedTime: "",
			Tags:          []string{"", "?, "?},
		},
		{
			ID:            uuid.New(),
			Title:         "",
			Content:       "1.5-2?,
			Category:      "diet",
			Difficulty:    "easy",
			EstimatedTime: "",
			Tags:          []string{"", "", ""},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// getSleepTips 
func (s *HealthRecommendationService) getSleepTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "",
			Content:       "?,
			Category:      "sleep",
			Difficulty:    "medium",
			EstimatedTime: "",
			Tags:          []string{"", "?, ""},
		},
		{
			ID:            uuid.New(),
			Title:         "",
			Content:       "1豸?,
			Category:      "sleep",
			Difficulty:    "easy",
			EstimatedTime: "1",
			Tags:          []string{"", "豸", ""},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// getStressTips 
func (s *HealthRecommendationService) getStressTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "?,
			Content:       "?,
			Category:      "stress",
			Difficulty:    "easy",
			EstimatedTime: "5-10",
			Tags:          []string{"?, "", ""},
		},
		{
			ID:            uuid.New(),
			Title:         "",
			Content:       "趨?,
			Category:      "stress",
			Difficulty:    "medium",
			EstimatedTime: "",
			Tags:          []string{"", "?, ""},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// getGeneralTips 
func (s *HealthRecommendationService) getGeneralTips(limit int) []HealthTip {
	allTips := []HealthTip{
		{
			ID:            uuid.New(),
			Title:         "?,
			Content:       "?,
			Category:      "general",
			Difficulty:    "medium",
			EstimatedTime: "",
			Tags:          []string{"?, "?, "?},
		},
		{
			ID:            uuid.New(),
			Title:         "",
			Content:       "緢?,
			Category:      "general",
			Difficulty:    "easy",
			EstimatedTime: "",
			Tags:          []string{"", "", ""},
		},
	}

	if len(allTips) > limit {
		return allTips[:limit]
	}
	return allTips
}

// 
func (s *HealthRecommendationService) calculateAverageSteps(data []*domain.HealthData) float64 {
	if len(data) == 0 {
		return 0
	}

	total := 0.0
	for _, d := range data {
		total += d.Value
	}
	return total / float64(len(data))
}

func (s *HealthRecommendationService) calculateAverageValue(data []*domain.HealthData) float64 {
	if len(data) == 0 {
		return 0
	}

	total := 0.0
	for _, d := range data {
		total += d.Value
	}
	return total / float64(len(data))
}

func (s *HealthRecommendationService) generateRecommendationSummary(recommendations []HealthRecommendation) string {
	if len(recommendations) == 0 {
		return ""
	}

	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, rec := range recommendations {
		switch rec.Priority {
		case RecommendationPriorityHigh:
			highCount++
		case RecommendationPriorityMedium:
			mediumCount++
		case RecommendationPriorityLow:
			lowCount++
		}
	}

	summary := fmt.Sprintf("?d?, len(recommendations))
	if highCount > 0 {
		summary += fmt.Sprintf("?d?, highCount)
	}
	if mediumCount > 0 {
		summary += fmt.Sprintf("?d?, mediumCount)
	}
	if lowCount > 0 {
		summary += fmt.Sprintf("?d?, lowCount)
	}

	return summary
}

