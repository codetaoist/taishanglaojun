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

// HealthDashboardService ?
type HealthDashboardService struct {
	healthDataRepo    domain.HealthDataRepository
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthDashboardService ?
func NewHealthDashboardService(
	healthDataRepo domain.HealthDataRepository,
	healthProfileRepo domain.HealthProfileRepository,
	eventPublisher EventPublisher,
) *HealthDashboardService {
	return &HealthDashboardService{
		healthDataRepo:    healthDataRepo,
		healthProfileRepo: healthProfileRepo,
		eventPublisher:    eventPublisher,
	}
}

// DashboardPeriod ?
type DashboardPeriod string

const (
	DashboardPeriodDay   DashboardPeriod = "day"   // ?
	DashboardPeriodWeek  DashboardPeriod = "week"  // ?
	DashboardPeriodMonth DashboardPeriod = "month" // ?
	DashboardPeriodYear  DashboardPeriod = "year"  // ?
)

// HealthDashboard ?
type HealthDashboard struct {
	UserID           uuid.UUID                  `json:"user_id"`
	Period           DashboardPeriod            `json:"period"`
	Overview         HealthOverview             `json:"overview"`
	KeyMetrics       []KeyMetric                `json:"key_metrics"`
	TrendCharts      []TrendChart               `json:"trend_charts"`
	HealthScore      HealthScore                `json:"health_score"`
	Recommendations  []DashboardRecommendation  `json:"recommendations"`
	Alerts           []DashboardAlert           `json:"alerts"`
	Achievements     []Achievement              `json:"achievements"`
	Goals            []HealthGoal               `json:"goals"`
	LastUpdated      time.Time                  `json:"last_updated"`
	Metadata         map[string]interface{}     `json:"metadata,omitempty"`
}

// HealthOverview 
type HealthOverview struct {
	TotalDataPoints    int                    `json:"total_data_points"`
	ActiveDays         int                    `json:"active_days"`
	HealthStatus       string                 `json:"health_status"`
	OverallScore       float64                `json:"overall_score"`
	LastCheckup        *time.Time             `json:"last_checkup,omitempty"`
	NextCheckup        *time.Time             `json:"next_checkup,omitempty"`
	RiskFactors        []string               `json:"risk_factors"`
	HealthCategories   map[string]float64     `json:"health_categories"`
	Summary            string                 `json:"summary"`
}

// KeyMetric 
type KeyMetric struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Change      float64                `json:"change"`
	ChangeType  string                 `json:"change_type"` // increase, decrease, stable
	Status      string                 `json:"status"`      // normal, warning, critical
	Target      *float64               `json:"target,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Color       string                 `json:"color,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TrendChart 
type TrendChart struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // line, bar, area, pie
	DataType    string                 `json:"data_type"`
	Data        []ChartDataPoint       `json:"data"`
	XAxis       ChartAxis              `json:"x_axis"`
	YAxis       ChartAxis              `json:"y_axis"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ChartDataPoint ?
type ChartDataPoint struct {
	X     interface{} `json:"x"`
	Y     float64     `json:"y"`
	Label string      `json:"label,omitempty"`
	Color string      `json:"color,omitempty"`
}

// ChartAxis ?
type ChartAxis struct {
	Label  string `json:"label"`
	Unit   string `json:"unit,omitempty"`
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Format string `json:"format,omitempty"`
}

// HealthScore 
type HealthScore struct {
	Overall      float64                `json:"overall"`
	Categories   map[string]float64     `json:"categories"`
	Trend        string                 `json:"trend"` // improving, declining, stable
	LastScore    float64                `json:"last_score"`
	ScoreHistory []ScoreHistoryPoint    `json:"score_history"`
	Factors      []ScoreFactor          `json:"factors"`
}

// ScoreHistoryPoint ?
type ScoreHistoryPoint struct {
	Date  time.Time `json:"date"`
	Score float64   `json:"score"`
}

// ScoreFactor 
type ScoreFactor struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Score       float64 `json:"score"`
	Impact      string  `json:"impact"` // positive, negative, neutral
	Description string  `json:"description,omitempty"`
}

// DashboardRecommendation 彨?
type DashboardRecommendation struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
	Icon        string    `json:"icon,omitempty"`
	Color       string    `json:"color,omitempty"`
}

// DashboardAlert 徯?
type DashboardAlert struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	IsRead      bool      `json:"is_read"`
	Icon        string    `json:"icon,omitempty"`
	Color       string    `json:"color,omitempty"`
}

// Achievement 
type Achievement struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Progress    float64   `json:"progress"`
	Target      float64   `json:"target"`
	IsCompleted bool      `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Badge       string    `json:"badge,omitempty"`
}

// HealthGoal 
type HealthGoal struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Target      float64   `json:"target"`
	Current     float64   `json:"current"`
	Progress    float64   `json:"progress"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Status      string    `json:"status"` // active, completed, paused, expired
	CreatedAt   time.Time `json:"created_at"`
}

// GetDashboardRequest ?
type GetDashboardRequest struct {
	UserID uuid.UUID       `json:"user_id" binding:"required"`
	Period DashboardPeriod `json:"period,omitempty"`
}

// GetDashboardResponse ?
type GetDashboardResponse struct {
	Dashboard HealthDashboard `json:"dashboard"`
	Timestamp time.Time       `json:"timestamp"`
}

// GetDashboardSummaryRequest ?
type GetDashboardSummaryRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// GetDashboardSummaryResponse ?
type GetDashboardSummaryResponse struct {
	Summary   HealthOverview `json:"summary"`
	Timestamp time.Time      `json:"timestamp"`
}

// GetDashboard ?
func (s *HealthDashboardService) GetDashboard(ctx context.Context, req *GetDashboardRequest) (*GetDashboardResponse, error) {
	// 
	if req.Period == "" {
		req.Period = DashboardPeriodMonth
	}

	// 
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// 
	endTime := time.Now()
	startTime := s.calculateStartTime(endTime, req.Period)

	// ?
	dashboard := HealthDashboard{
		UserID:      req.UserID,
		Period:      req.Period,
		LastUpdated: time.Now(),
	}

	// 
	dashboard.Overview = s.generateOverview(ctx, req.UserID, startTime, endTime, profile)
	dashboard.KeyMetrics = s.generateKeyMetrics(ctx, req.UserID, startTime, endTime)
	dashboard.TrendCharts = s.generateTrendCharts(ctx, req.UserID, startTime, endTime, req.Period)
	dashboard.HealthScore = s.generateHealthScore(ctx, req.UserID, startTime, endTime)
	dashboard.Recommendations = s.generateDashboardRecommendations(ctx, req.UserID, startTime, endTime)
	dashboard.Alerts = s.generateDashboardAlerts(ctx, req.UserID, startTime, endTime)
	dashboard.Achievements = s.generateAchievements(ctx, req.UserID, startTime, endTime)
	dashboard.Goals = s.generateHealthGoals(ctx, req.UserID)

	return &GetDashboardResponse{
		Dashboard: dashboard,
		Timestamp: time.Now(),
	}, nil
}

// GetDashboardSummary ?
func (s *HealthDashboardService) GetDashboardSummary(ctx context.Context, req *GetDashboardSummaryRequest) (*GetDashboardSummaryResponse, error) {
	// 
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// ?0
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	// 
	overview := s.generateOverview(ctx, req.UserID, startTime, endTime, profile)

	return &GetDashboardSummaryResponse{
		Summary:   overview,
		Timestamp: time.Now(),
	}, nil
}

// generateOverview 
func (s *HealthDashboardService) generateOverview(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) HealthOverview {
	// ?
	allData, _ := s.healthDataRepo.GetByUserID(ctx, userID, startTime, endTime)

	// 
	activeDays := s.calculateActiveDays(allData, startTime, endTime)

	// 
	overallScore := s.calculateOverallScore(allData)

	// 
	riskFactors := s.identifyRiskFactors(allData, profile)

	// 㽡
	healthCategories := s.calculateHealthCategories(allData)

	// 
	summary := s.generateHealthSummary(overallScore, riskFactors, activeDays)

	return HealthOverview{
		TotalDataPoints:  len(allData),
		ActiveDays:       activeDays,
		HealthStatus:     s.determineHealthStatus(overallScore),
		OverallScore:     overallScore,
		RiskFactors:      riskFactors,
		HealthCategories: healthCategories,
		Summary:          summary,
	}
}

// generateKeyMetrics 
func (s *HealthDashboardService) generateKeyMetrics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) []KeyMetric {
	var metrics []KeyMetric

	// 
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	if len(stepsData) > 0 {
		avgSteps := s.calculateAverage(stepsData)
		prevPeriodStart := startTime.AddDate(0, 0, -int(endTime.Sub(startTime).Hours()/24))
		prevStepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", prevPeriodStart, startTime)
		prevAvgSteps := s.calculateAverage(prevStepsData)
		
		change := avgSteps - prevAvgSteps
		changeType := "stable"
		if change > 0 {
			changeType = "increase"
		} else if change < 0 {
			changeType = "decrease"
		}

		status := "normal"
		if avgSteps < 5000 {
			status = "warning"
		} else if avgSteps < 3000 {
			status = "critical"
		}

		target := 10000.0
		metrics = append(metrics, KeyMetric{
			ID:          "steps",
			Name:        "",
			Value:       avgSteps,
			Unit:        "?,
			Change:      change,
			ChangeType:  changeType,
			Status:      status,
			Target:      &target,
			Icon:        "",
			Color:       s.getStatusColor(status),
			Description: "",
		})
	}

	// 
	heartRateData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "heart_rate", startTime, endTime)
	if len(heartRateData) > 0 {
		avgHeartRate := s.calculateAverage(heartRateData)
		status := "normal"
		if avgHeartRate > 100 || avgHeartRate < 60 {
			status = "warning"
		}
		if avgHeartRate > 120 || avgHeartRate < 50 {
			status = "critical"
		}

		metrics = append(metrics, KeyMetric{
			ID:          "heart_rate",
			Name:        "",
			Value:       avgHeartRate,
			Unit:        "bpm",
			Status:      status,
			Icon:        "",
			Color:       s.getStatusColor(status),
			Description: "?,
		})
	}

	// 
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)
	if len(sleepData) > 0 {
		avgSleep := s.calculateAverage(sleepData)
		status := "normal"
		if avgSleep < 7 || avgSleep > 9 {
			status = "warning"
		}
		if avgSleep < 5 || avgSleep > 10 {
			status = "critical"
		}

		target := 8.0
		metrics = append(metrics, KeyMetric{
			ID:          "sleep",
			Name:        "",
			Value:       avgSleep,
			Unit:        "",
			Status:      status,
			Target:      &target,
			Icon:        "",
			Color:       s.getStatusColor(status),
			Description: "",
		})
	}

	return metrics
}

// generateTrendCharts 
func (s *HealthDashboardService) generateTrendCharts(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, period DashboardPeriod) []TrendChart {
	var charts []TrendChart

	// ?
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	if len(stepsData) > 0 {
		chartData := s.convertToChartData(stepsData, period)
		charts = append(charts, TrendChart{
			ID:       "steps_trend",
			Title:    "",
			Type:     "line",
			DataType: "steps",
			Data:     chartData,
			XAxis:    ChartAxis{Label: "", Format: s.getTimeFormat(period)},
			YAxis:    ChartAxis{Label: "", Unit: "?},
		})
	}

	// ?
	heartRateData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "heart_rate", startTime, endTime)
	if len(heartRateData) > 0 {
		chartData := s.convertToChartData(heartRateData, period)
		charts = append(charts, TrendChart{
			ID:       "heart_rate_trend",
			Title:    "",
			Type:     "line",
			DataType: "heart_rate",
			Data:     chartData,
			XAxis:    ChartAxis{Label: "", Format: s.getTimeFormat(period)},
			YAxis:    ChartAxis{Label: "", Unit: "bpm"},
		})
	}

	// ?
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)
	if len(sleepData) > 0 {
		chartData := s.convertToChartData(sleepData, period)
		charts = append(charts, TrendChart{
			ID:       "sleep_trend",
			Title:    "",
			Type:     "bar",
			DataType: "sleep_duration",
			Data:     chartData,
			XAxis:    ChartAxis{Label: "", Format: s.getTimeFormat(period)},
			YAxis:    ChartAxis{Label: "", Unit: ""},
		})
	}

	return charts
}

// generateHealthScore 
func (s *HealthDashboardService) generateHealthScore(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) HealthScore {
	// 
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	heartRateData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "heart_rate", startTime, endTime)
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)

	// ?
	categories := make(map[string]float64)
	factors := []ScoreFactor{}

	// 
	if len(stepsData) > 0 {
		avgSteps := s.calculateAverage(stepsData)
		exerciseScore := math.Min(avgSteps/10000*100, 100)
		categories["exercise"] = exerciseScore
		factors = append(factors, ScoreFactor{
			Name:        "",
			Weight:      0.3,
			Score:       exerciseScore,
			Impact:      s.getScoreImpact(exerciseScore),
			Description: "?,
		})
	}

	// ?
	if len(heartRateData) > 0 {
		avgHeartRate := s.calculateAverage(heartRateData)
		cardioScore := s.calculateCardioScore(avgHeartRate)
		categories["cardiovascular"] = cardioScore
		factors = append(factors, ScoreFactor{
			Name:        "?,
			Weight:      0.25,
			Score:       cardioScore,
			Impact:      s.getScoreImpact(cardioScore),
			Description: "?,
		})
	}

	// 
	if len(sleepData) > 0 {
		avgSleep := s.calculateAverage(sleepData)
		sleepScore := s.calculateSleepScore(avgSleep)
		categories["sleep"] = sleepScore
		factors = append(factors, ScoreFactor{
			Name:        "",
			Weight:      0.25,
			Score:       sleepScore,
			Impact:      s.getScoreImpact(sleepScore),
			Description: "?,
		})
	}

	// 
	overall := s.calculateWeightedScore(factors)

	// 
	scoreHistory := []ScoreHistoryPoint{
		{Date: startTime, Score: overall - 5},
		{Date: endTime, Score: overall},
	}

	// 
	trend := "stable"
	if len(scoreHistory) > 1 {
		if scoreHistory[len(scoreHistory)-1].Score > scoreHistory[0].Score {
			trend = "improving"
		} else if scoreHistory[len(scoreHistory)-1].Score < scoreHistory[0].Score {
			trend = "declining"
		}
	}

	return HealthScore{
		Overall:      overall,
		Categories:   categories,
		Trend:        trend,
		LastScore:    overall - 2,
		ScoreHistory: scoreHistory,
		Factors:      factors,
	}
}

// generateDashboardRecommendations 彨?
func (s *HealthDashboardService) generateDashboardRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) []DashboardRecommendation {
	var recommendations []DashboardRecommendation

	// 
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	if len(stepsData) > 0 {
		avgSteps := s.calculateAverage(stepsData)
		if avgSteps < 5000 {
			recommendations = append(recommendations, DashboardRecommendation{
				ID:          uuid.New(),
				Type:        "exercise",
				Priority:    "high",
				Title:       "",
				Description: "",
				Action:      "趨8000?,
				Icon:        "",
				Color:       "#ff6b6b",
			})
		}
	}

	// 
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)
	if len(sleepData) > 0 {
		avgSleep := s.calculateAverage(sleepData)
		if avgSleep < 7 {
			recommendations = append(recommendations, DashboardRecommendation{
				ID:          uuid.New(),
				Type:        "sleep",
				Priority:    "medium",
				Title:       "",
				Description: "䲻㽨?,
				Action:      "",
				Icon:        "",
				Color:       "#4ecdc4",
			})
		}
	}

	return recommendations
}

// generateDashboardAlerts 徯?
func (s *HealthDashboardService) generateDashboardAlerts(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) []DashboardAlert {
	var alerts []DashboardAlert

	// ?
	heartRateData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "heart_rate", startTime, endTime)
	for _, data := range heartRateData {
		if data.Value > 120 {
			alerts = append(alerts, DashboardAlert{
				ID:        uuid.New(),
				Type:      "heart_rate",
				Severity:  "warning",
				Title:     "",
				Message:   fmt.Sprintf("?.0f bpm", data.Value),
				Timestamp: data.RecordedAt,
				IsRead:    false,
				Icon:      "",
				Color:     "#ff9f43",
			})
		}
	}

	return alerts
}

// generateAchievements 
func (s *HealthDashboardService) generateAchievements(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) []Achievement {
	var achievements []Achievement

	// 
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	if len(stepsData) > 0 {
		maxSteps := 0.0
		for _, data := range stepsData {
			if data.Value > maxSteps {
				maxSteps = data.Value
			}
		}

		if maxSteps >= 10000 {
			achievements = append(achievements, Achievement{
				ID:          uuid.New(),
				Name:        "",
				Description: "10000?,
				Type:        "exercise",
				Progress:    100,
				Target:      10000,
				IsCompleted: true,
				Icon:        "",
				Badge:       "gold",
			})
		}
	}

	return achievements
}

// generateHealthGoals 
func (s *HealthDashboardService) generateHealthGoals(ctx context.Context, userID uuid.UUID) []HealthGoal {
	var goals []HealthGoal

	// 
	goals = append(goals, HealthGoal{
		ID:          uuid.New(),
		Name:        "",
		Description: "10000?,
		Type:        "exercise",
		Target:      10000,
		Current:     7500, // ?
		Progress:    75,
		Status:      "active",
		CreatedAt:   time.Now().AddDate(0, 0, -30),
	})

	goals = append(goals, HealthGoal{
		ID:          uuid.New(),
		Name:        "",
		Description: "8",
		Type:        "sleep",
		Target:      8,
		Current:     7.2, // ?
		Progress:    90,
		Status:      "active",
		CreatedAt:   time.Now().AddDate(0, 0, -30),
	})

	return goals
}

// 
func (s *HealthDashboardService) calculateStartTime(endTime time.Time, period DashboardPeriod) time.Time {
	switch period {
	case DashboardPeriodDay:
		return endTime.AddDate(0, 0, -1)
	case DashboardPeriodWeek:
		return endTime.AddDate(0, 0, -7)
	case DashboardPeriodMonth:
		return endTime.AddDate(0, -1, 0)
	case DashboardPeriodYear:
		return endTime.AddDate(-1, 0, 0)
	default:
		return endTime.AddDate(0, -1, 0)
	}
}

func (s *HealthDashboardService) calculateActiveDays(data []*domain.HealthData, startTime, endTime time.Time) int {
	daySet := make(map[string]bool)
	for _, d := range data {
		day := d.RecordedAt.Format("2006-01-02")
		daySet[day] = true
	}
	return len(daySet)
}

func (s *HealthDashboardService) calculateOverallScore(data []*domain.HealthData) float64 {
	if len(data) == 0 {
		return 0
	}
	// 㷨
	return math.Min(float64(len(data))/10*100, 100)
}

func (s *HealthDashboardService) identifyRiskFactors(data []*domain.HealthData, profile *domain.HealthProfile) []string {
	var factors []string
	
	// 
	for _, d := range data {
		if d.DataType == "heart_rate" && d.Value > 100 {
			factors = append(factors, "")
			break
		}
	}

	// 
	if profile != nil {
		bmi := profile.GetBMI()
		if bmi > 25 {
			factors = append(factors, "")
		}
		if profile.GetAge() > 50 {
			factors = append(factors, "")
		}
	}

	return factors
}

func (s *HealthDashboardService) calculateHealthCategories(data []*domain.HealthData) map[string]float64 {
	categories := make(map[string]float64)
	
	// ?
	typeGroups := make(map[string][]*domain.HealthData)
	for _, d := range data {
		typeGroups[d.DataType] = append(typeGroups[d.DataType], d)
	}

	for dataType, typeData := range typeGroups {
		score := s.calculateTypeScore(dataType, typeData)
		categories[dataType] = score
	}

	return categories
}

func (s *HealthDashboardService) calculateTypeScore(dataType string, data []*domain.HealthData) float64 {
	if len(data) == 0 {
		return 0
	}

	avg := s.calculateAverage(data)
	
	switch dataType {
	case "steps":
		return math.Min(avg/10000*100, 100)
	case "heart_rate":
		return s.calculateCardioScore(avg)
	case "sleep_duration":
		return s.calculateSleepScore(avg)
	default:
		return 75 // 
	}
}

func (s *HealthDashboardService) calculateCardioScore(heartRate float64) float64 {
	if heartRate >= 60 && heartRate <= 100 {
		return 100
	} else if heartRate >= 50 && heartRate <= 120 {
		return 75
	} else {
		return 50
	}
}

func (s *HealthDashboardService) calculateSleepScore(sleepHours float64) float64 {
	if sleepHours >= 7 && sleepHours <= 9 {
		return 100
	} else if sleepHours >= 6 && sleepHours <= 10 {
		return 75
	} else {
		return 50
	}
}

func (s *HealthDashboardService) generateHealthSummary(score float64, riskFactors []string, activeDays int) string {
	if score >= 80 {
		return fmt.Sprintf("?d", activeDays)
	} else if score >= 60 {
		return fmt.Sprintf("㽨%d?, len(riskFactors))
	} else {
		return "?
	}
}

func (s *HealthDashboardService) determineHealthStatus(score float64) string {
	if score >= 80 {
		return "excellent"
	} else if score >= 60 {
		return "good"
	} else if score >= 40 {
		return "fair"
	} else {
		return "poor"
	}
}

func (s *HealthDashboardService) calculateAverage(data []*domain.HealthData) float64 {
	if len(data) == 0 {
		return 0
	}

	total := 0.0
	for _, d := range data {
		total += d.Value
	}
	return total / float64(len(data))
}

func (s *HealthDashboardService) getStatusColor(status string) string {
	switch status {
	case "normal":
		return "#27ae60"
	case "warning":
		return "#f39c12"
	case "critical":
		return "#e74c3c"
	default:
		return "#95a5a6"
	}
}

func (s *HealthDashboardService) convertToChartData(data []*domain.HealthData, period DashboardPeriod) []ChartDataPoint {
	var chartData []ChartDataPoint

	// ?
	sort.Slice(data, func(i, j int) bool {
		return data[i].RecordedAt.Before(data[j].RecordedAt)
	})

	for _, d := range data {
		var x interface{}
		switch period {
		case DashboardPeriodDay:
			x = d.RecordedAt.Format("15:04")
		case DashboardPeriodWeek:
			x = d.RecordedAt.Format("Mon")
		case DashboardPeriodMonth:
			x = d.RecordedAt.Format("01-02")
		case DashboardPeriodYear:
			x = d.RecordedAt.Format("01")
		default:
			x = d.RecordedAt.Format("01-02")
		}

		chartData = append(chartData, ChartDataPoint{
			X: x,
			Y: d.Value,
		})
	}

	return chartData
}

func (s *HealthDashboardService) getTimeFormat(period DashboardPeriod) string {
	switch period {
	case DashboardPeriodDay:
		return "HH:mm"
	case DashboardPeriodWeek:
		return "ddd"
	case DashboardPeriodMonth:
		return "MM-DD"
	case DashboardPeriodYear:
		return "MM"
	default:
		return "MM-DD"
	}
}

func (s *HealthDashboardService) getScoreImpact(score float64) string {
	if score >= 80 {
		return "positive"
	} else if score >= 60 {
		return "neutral"
	} else {
		return "negative"
	}
}

func (s *HealthDashboardService) calculateWeightedScore(factors []ScoreFactor) float64 {
	if len(factors) == 0 {
		return 0
	}

	totalWeight := 0.0
	weightedSum := 0.0

	for _, factor := range factors {
		totalWeight += factor.Weight
		weightedSum += factor.Score * factor.Weight
	}

	if totalWeight == 0 {
		return 0
	}

	return weightedSum / totalWeight
}

