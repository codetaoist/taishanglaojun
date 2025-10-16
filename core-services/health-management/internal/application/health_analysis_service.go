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

// HealthAnalysisService 
type HealthAnalysisService struct {
	healthDataRepo    domain.HealthDataRepository
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthAnalysisService 
func NewHealthAnalysisService(
	healthDataRepo domain.HealthDataRepository,
	healthProfileRepo domain.HealthProfileRepository,
	eventPublisher EventPublisher,
) *HealthAnalysisService {
	return &HealthAnalysisService{
		healthDataRepo:    healthDataRepo,
		healthProfileRepo: healthProfileRepo,
		eventPublisher:    eventPublisher,
	}
}

// HealthTrendAnalysisRequest 
type HealthTrendAnalysisRequest struct {
	UserID    uuid.UUID             `json:"user_id" validate:"required"`
	DataType  domain.HealthDataType `json:"data_type" validate:"required"`
	StartTime time.Time             `json:"start_time" validate:"required"`
	EndTime   time.Time             `json:"end_time" validate:"required"`
	Period    string                `json:"period" validate:"required,oneof=daily weekly monthly"`
}

// HealthTrendAnalysisResponse 
type HealthTrendAnalysisResponse struct {
	UserID       uuid.UUID             `json:"user_id"`
	DataType     domain.HealthDataType `json:"data_type"`
	Period       string                `json:"period"`
	TrendData    []TrendDataPoint      `json:"trend_data"`
	TrendType    string                `json:"trend_type"` // increasing, decreasing, stable, fluctuating
	TrendStrength float64              `json:"trend_strength"` // 0-1
	Average      float64               `json:"average"`
	Min          float64               `json:"min"`
	Max          float64               `json:"max"`
	StdDeviation float64               `json:"std_deviation"`
	Insights     []string              `json:"insights"`
	Recommendations []string           `json:"recommendations"`
}

// TrendDataPoint ?
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Count     int       `json:"count"`
}

// HealthRiskAssessmentRequest 
type HealthRiskAssessmentRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// HealthRiskAssessmentResponse 
type HealthRiskAssessmentResponse struct {
	UserID           uuid.UUID    `json:"user_id"`
	OverallRiskLevel string       `json:"overall_risk_level"` // low, medium, high, critical
	RiskScore        float64      `json:"risk_score"` // 0-100
	RiskFactors      []RiskFactor `json:"risk_factors"`
	Recommendations  []string     `json:"recommendations"`
	NextCheckupDate  *time.Time   `json:"next_checkup_date,omitempty"`
	AssessedAt       time.Time    `json:"assessed_at"`
}

// RiskFactor 
type RiskFactor struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	RiskLevel   string  `json:"risk_level"`
	Impact      float64 `json:"impact"` // 0-1
	Suggestions []string `json:"suggestions"`
}

// HealthInsightsRequest 
type HealthInsightsRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

// HealthInsightsResponse 
type HealthInsightsResponse struct {
	UserID              uuid.UUID           `json:"user_id"`
	Period              string              `json:"period"`
	OverallHealthScore  float64             `json:"overall_health_score"` // 0-100
	HealthStatus        string              `json:"health_status"` // excellent, good, fair, poor
	KeyMetrics          []KeyMetric         `json:"key_metrics"`
	Achievements        []Achievement       `json:"achievements"`
	AreasForImprovement []ImprovementArea   `json:"areas_for_improvement"`
	PersonalizedTips    []PersonalizedTip   `json:"personalized_tips"`
	GeneratedAt         time.Time           `json:"generated_at"`
}

// KeyMetric 
type KeyMetric struct {
	Name         string  `json:"name"`
	Value        float64 `json:"value"`
	Unit         string  `json:"unit"`
	Status       string  `json:"status"` // excellent, good, fair, poor
	Trend        string  `json:"trend"` // improving, stable, declining
	TargetValue  *float64 `json:"target_value,omitempty"`
	PercentChange float64 `json:"percent_change"`
}

// Achievement 
type Achievement struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	EarnedAt    time.Time `json:"earned_at"`
	Icon        string    `json:"icon"`
}

// ImprovementArea 
type ImprovementArea struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"` // high, medium, low
	Actions     []string `json:"actions"`
}

// PersonalizedTip 
type PersonalizedTip struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Difficulty  string `json:"difficulty"` // easy, medium, hard
}

// AnalyzeHealthTrend 
func (s *HealthAnalysisService) AnalyzeHealthTrend(ctx context.Context, req *HealthTrendAnalysisRequest) (*HealthTrendAnalysisResponse, error) {
	// 
	healthData, err := s.healthDataRepo.FindByUserIDAndTypeAndTimeRange(
		ctx, req.UserID, req.DataType, req.StartTime, req.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch health data: %w", err)
	}

	if len(healthData) == 0 {
		return &HealthTrendAnalysisResponse{
			UserID:   req.UserID,
			DataType: req.DataType,
			Period:   req.Period,
			Insights: []string{"㹻"},
		}, nil
	}

	// 
	trendData := s.aggregateDataByPeriod(healthData, req.Period)
	
	// 
	stats := s.calculateStatistics(trendData)
	
	// ?
	trendType, trendStrength := s.analyzeTrendPattern(trendData)
	
	// ?
	insights := s.generateTrendInsights(req.DataType, trendType, stats)
	recommendations := s.generateTrendRecommendations(req.DataType, trendType, stats)

	return &HealthTrendAnalysisResponse{
		UserID:          req.UserID,
		DataType:        req.DataType,
		Period:          req.Period,
		TrendData:       trendData,
		TrendType:       trendType,
		TrendStrength:   trendStrength,
		Average:         stats.Average,
		Min:             stats.Min,
		Max:             stats.Max,
		StdDeviation:    stats.StdDeviation,
		Insights:        insights,
		Recommendations: recommendations,
	}, nil
}

// AssessHealthRisk 
func (s *HealthAnalysisService) AssessHealthRisk(ctx context.Context, req *HealthRiskAssessmentRequest) (*HealthRiskAssessmentResponse, error) {
	// 
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch health profile: %w", err)
	}

	// 
	endTime := time.Now()
	startTime := endTime.AddDate(0, -3, 0) // ?
	
	healthData, err := s.healthDataRepo.FindByUserIDAndTimeRange(ctx, req.UserID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent health data: %w", err)
	}

	// 
	riskFactors := s.assessRiskFactors(profile, healthData)
	
	// 
	riskScore := s.calculateOverallRiskScore(riskFactors)
	
	// 
	riskLevel := s.determineRiskLevel(riskScore)
	
	// 
	recommendations := s.generateRiskRecommendations(riskFactors, riskLevel)
	
	// ?
	nextCheckupDate := s.calculateNextCheckupDate(riskLevel, profile)

	return &HealthRiskAssessmentResponse{
		UserID:           req.UserID,
		OverallRiskLevel: riskLevel,
		RiskScore:        riskScore,
		RiskFactors:      riskFactors,
		Recommendations:  recommendations,
		NextCheckupDate:  nextCheckupDate,
		AssessedAt:       time.Now(),
	}, nil
}

// GenerateHealthInsights 
func (s *HealthAnalysisService) GenerateHealthInsights(ctx context.Context, req *HealthInsightsRequest) (*HealthInsightsResponse, error) {
	// 
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch health profile: %w", err)
	}

	// 
	healthData, err := s.healthDataRepo.FindByUserIDAndTimeRange(ctx, req.UserID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch health data: %w", err)
	}

	// 彡
	healthScore := s.calculateOverallHealthScore(profile, healthData)
	
	// ?
	healthStatus := s.determineHealthStatus(healthScore)
	
	// 
	keyMetrics := s.analyzeKeyMetrics(healthData, req.StartTime, req.EndTime)
	
	// 
	achievements := s.identifyAchievements(healthData, profile)
	
	// 
	improvementAreas := s.identifyImprovementAreas(healthData, profile)
	
	// 
	personalizedTips := s.generatePersonalizedTips(profile, healthData, improvementAreas)

	period := s.formatPeriod(req.StartTime, req.EndTime)

	return &HealthInsightsResponse{
		UserID:              req.UserID,
		Period:              period,
		OverallHealthScore:  healthScore,
		HealthStatus:        healthStatus,
		KeyMetrics:          keyMetrics,
		Achievements:        achievements,
		AreasForImprovement: improvementAreas,
		PersonalizedTips:    personalizedTips,
		GeneratedAt:         time.Now(),
	}, nil
}

// 

type Statistics struct {
	Average      float64
	Min          float64
	Max          float64
	StdDeviation float64
}

func (s *HealthAnalysisService) aggregateDataByPeriod(data []*domain.HealthData, period string) []TrendDataPoint {
	dataMap := make(map[string][]float64)
	
	for _, d := range data {
		var key string
		switch period {
		case "daily":
			key = d.RecordedAt.Format("2006-01-02")
		case "weekly":
			year, week := d.RecordedAt.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		case "monthly":
			key = d.RecordedAt.Format("2006-01")
		}
		
		dataMap[key] = append(dataMap[key], d.Value)
	}
	
	var trendData []TrendDataPoint
	for key, values := range dataMap {
		var timestamp time.Time
		var err error
		
		switch period {
		case "daily":
			timestamp, err = time.Parse("2006-01-02", key)
		case "weekly":
			// 
			timestamp, err = time.Parse("2006-W02", key)
		case "monthly":
			timestamp, err = time.Parse("2006-01", key)
		}
		
		if err != nil {
			continue
		}
		
		// ?
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		average := sum / float64(len(values))
		
		trendData = append(trendData, TrendDataPoint{
			Timestamp: timestamp,
			Value:     average,
			Count:     len(values),
		})
	}
	
	// ?
	sort.Slice(trendData, func(i, j int) bool {
		return trendData[i].Timestamp.Before(trendData[j].Timestamp)
	})
	
	return trendData
}

func (s *HealthAnalysisService) calculateStatistics(data []TrendDataPoint) Statistics {
	if len(data) == 0 {
		return Statistics{}
	}
	
	values := make([]float64, len(data))
	sum := 0.0
	min := data[0].Value
	max := data[0].Value
	
	for i, point := range data {
		values[i] = point.Value
		sum += point.Value
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}
	
	average := sum / float64(len(data))
	
	// ?
	variance := 0.0
	for _, value := range values {
		variance += math.Pow(value-average, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(data)))
	
	return Statistics{
		Average:      average,
		Min:          min,
		Max:          max,
		StdDeviation: stdDev,
	}
}

func (s *HealthAnalysisService) analyzeTrendPattern(data []TrendDataPoint) (string, float64) {
	if len(data) < 2 {
		return "stable", 0.0
	}
	
	// ?
	n := float64(len(data))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	
	for i, point := range data {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	// 
	avgX := sumX / n
	avgY := sumY / n
	
	numerator := 0.0
	denomX := 0.0
	denomY := 0.0
	
	for i, point := range data {
		x := float64(i)
		y := point.Value
		numerator += (x - avgX) * (y - avgY)
		denomX += math.Pow(x-avgX, 2)
		denomY += math.Pow(y-avgY, 2)
	}
	
	correlation := numerator / math.Sqrt(denomX*denomY)
	strength := math.Abs(correlation)
	
	// 
	var trendType string
	if math.Abs(slope) < 0.1 {
		trendType = "stable"
	} else if slope > 0 {
		trendType = "increasing"
	} else {
		trendType = "decreasing"
	}
	
	// 仯?
	if strength < 0.3 && len(data) > 5 {
		trendType = "fluctuating"
	}
	
	return trendType, strength
}

func (s *HealthAnalysisService) generateTrendInsights(dataType domain.HealthDataType, trendType string, stats Statistics) []string {
	var insights []string
	
	switch dataType {
	case domain.HeartRate:
		switch trendType {
		case "increasing":
			insights = append(insights, "?)
		case "decreasing":
			insights = append(insights, "?)
		case "stable":
			insights = append(insights, "")
		case "fluctuating":
			insights = append(insights, "")
		}
	case domain.BloodPressure:
		switch trendType {
		case "increasing":
			insights = append(insights, "")
		case "decreasing":
			insights = append(insights, "?)
		case "stable":
			insights = append(insights, "?)
		}
	case domain.Steps:
		switch trendType {
		case "increasing":
			insights = append(insights, "?)
		case "decreasing":
			insights = append(insights, "?)
		case "stable":
			insights = append(insights, "?)
		}
	}
	
	return insights
}

func (s *HealthAnalysisService) generateTrendRecommendations(dataType domain.HealthDataType, trendType string, stats Statistics) []string {
	var recommendations []string
	
	switch dataType {
	case domain.HeartRate:
		if trendType == "increasing" {
			recommendations = append(recommendations, "?)
			recommendations = append(recommendations, "?)
		}
	case domain.BloodPressure:
		if trendType == "increasing" {
			recommendations = append(recommendations, "?)
			recommendations = append(recommendations, "?)
		}
	case domain.Steps:
		if trendType == "decreasing" {
			recommendations = append(recommendations, "趨?)
			recommendations = append(recommendations, "")
		}
	}
	
	return recommendations
}

func (s *HealthAnalysisService) assessRiskFactors(profile *domain.HealthProfile, healthData []*domain.HealthData) []RiskFactor {
	var riskFactors []RiskFactor
	
	// 
	age := profile.GetAge()
	if age > 65 {
		riskFactors = append(riskFactors, RiskFactor{
			Category:    "",
			Description: "䳬65?,
			RiskLevel:   "medium",
			Impact:      0.3,
			Suggestions: []string{"", "?},
		})
	}
	
	// BMI
	bmi := profile.GetBMI()
	if bmi != nil {
		if *bmi > 30 {
			riskFactors = append(riskFactors, RiskFactor{
				Category:    "",
				Description: "BMI30?,
				RiskLevel:   "high",
				Impact:      0.4,
				Suggestions: []string{"", "", "?},
			})
		} else if *bmi > 25 {
			riskFactors = append(riskFactors, RiskFactor{
				Category:    "",
				Description: "BMI25?,
				RiskLevel:   "medium",
				Impact:      0.2,
				Suggestions: []string{"", ""},
			})
		}
	}
	
	// ?
	for _, data := range healthData {
		if data.IsAbnormal() {
			riskLevel := string(data.GetRiskLevel())
			riskFactors = append(riskFactors, RiskFactor{
				Category:    string(data.DataType),
				Description: fmt.Sprintf("%s?, data.DataType),
				RiskLevel:   riskLevel,
				Impact:      s.getRiskImpact(riskLevel),
				Suggestions: s.getDataTypeRecommendations(data.DataType),
			})
		}
	}
	
	return riskFactors
}

func (s *HealthAnalysisService) calculateOverallRiskScore(riskFactors []RiskFactor) float64 {
	if len(riskFactors) == 0 {
		return 10.0 // 
	}
	
	totalImpact := 0.0
	for _, factor := range riskFactors {
		multiplier := 1.0
		switch factor.RiskLevel {
		case "low":
			multiplier = 1.0
		case "medium":
			multiplier = 2.0
		case "high":
			multiplier = 3.0
		case "critical":
			multiplier = 4.0
		}
		totalImpact += factor.Impact * multiplier * 20 // ?-100
	}
	
	// ?-100?
	score := math.Min(100, math.Max(0, 10+totalImpact))
	return score
}

func (s *HealthAnalysisService) determineRiskLevel(score float64) string {
	if score < 25 {
		return "low"
	} else if score < 50 {
		return "medium"
	} else if score < 75 {
		return "high"
	}
	return "critical"
}

func (s *HealthAnalysisService) generateRiskRecommendations(riskFactors []RiskFactor, riskLevel string) []string {
	var recommendations []string
	
	// 
	switch riskLevel {
	case "low":
		recommendations = append(recommendations, "?)
		recommendations = append(recommendations, "?)
	case "medium":
		recommendations = append(recommendations, "")
		recommendations = append(recommendations, "")
	case "high":
		recommendations = append(recommendations, "龡")
		recommendations = append(recommendations, "")
	case "critical":
		recommendations = append(recommendations, "")
		recommendations = append(recommendations, "")
	}
	
	// ?
	for _, factor := range riskFactors {
		recommendations = append(recommendations, factor.Suggestions...)
	}
	
	return recommendations
}

func (s *HealthAnalysisService) calculateNextCheckupDate(riskLevel string, profile *domain.HealthProfile) *time.Time {
	var months int
	switch riskLevel {
	case "low":
		months = 12
	case "medium":
		months = 6
	case "high":
		months = 3
	case "critical":
		months = 1
	}
	
	// 
	age := profile.GetAge()
	if age > 65 {
		months = months / 2
		if months < 1 {
			months = 1
		}
	}
	
	nextDate := time.Now().AddDate(0, months, 0)
	return &nextDate
}

func (s *HealthAnalysisService) calculateOverallHealthScore(profile *domain.HealthProfile, healthData []*domain.HealthData) float64 {
	score := 100.0 // 
	
	// 
	age := profile.GetAge()
	if age > 65 {
		score -= 10
	} else if age > 50 {
		score -= 5
	}
	
	// BMI
	bmi := profile.GetBMI()
	if bmi != nil {
		if *bmi > 30 || *bmi < 18.5 {
			score -= 15
		} else if *bmi > 25 || *bmi < 20 {
			score -= 5
		}
	}
	
	// 
	abnormalCount := 0
	for _, data := range healthData {
		if data.IsAbnormal() {
			abnormalCount++
			switch data.GetRiskLevel() {
			case domain.LowRisk:
				score -= 2
			case domain.MediumRisk:
				score -= 5
			case domain.HighRisk:
				score -= 10
			case domain.CriticalRisk:
				score -= 20
			}
		}
	}
	
	// ?-100?
	return math.Max(0, math.Min(100, score))
}

func (s *HealthAnalysisService) determineHealthStatus(score float64) string {
	if score >= 85 {
		return "excellent"
	} else if score >= 70 {
		return "good"
	} else if score >= 50 {
		return "fair"
	}
	return "poor"
}

func (s *HealthAnalysisService) analyzeKeyMetrics(healthData []*domain.HealthData, startTime, endTime time.Time) []KeyMetric {
	var metrics []KeyMetric
	
	// ?
	dataByType := make(map[domain.HealthDataType][]*domain.HealthData)
	for _, data := range healthData {
		dataByType[data.DataType] = append(dataByType[data.DataType], data)
	}
	
	// 
	for dataType, data := range dataByType {
		if len(data) == 0 {
			continue
		}
		
		// ?
		sum := 0.0
		for _, d := range data {
			sum += d.Value
		}
		average := sum / float64(len(data))
		
		// 
		status := s.getMetricStatus(dataType, average)
		trend := s.getMetricTrend(data)
		
		metrics = append(metrics, KeyMetric{
			Name:         string(dataType),
			Value:        average,
			Unit:         s.getDefaultUnit(dataType),
			Status:       status,
			Trend:        trend,
			TargetValue:  s.getTargetValue(dataType),
			PercentChange: s.calculatePercentChange(data),
		})
	}
	
	return metrics
}

func (s *HealthAnalysisService) identifyAchievements(healthData []*domain.HealthData, profile *domain.HealthProfile) []Achievement {
	var achievements []Achievement
	
	// 
	stepData := s.filterDataByType(healthData, domain.Steps)
	if len(stepData) > 0 {
		maxSteps := 0.0
		for _, data := range stepData {
			if data.Value > maxSteps {
				maxSteps = data.Value
			}
		}
		
		if maxSteps >= 10000 {
			achievements = append(achievements, Achievement{
				Title:       "",
				Description: "10,000?,
				Category:    "",
				EarnedAt:    time.Now(),
				Icon:        "?,
			})
		}
	}
	
	// 
	if len(healthData) >= 30 {
		achievements = append(achievements, Achievement{
			Title:       "",
			Description: "30?,
			Category:    "",
			EarnedAt:    time.Now(),
			Icon:        "",
		})
	}
	
	return achievements
}

func (s *HealthAnalysisService) identifyImprovementAreas(healthData []*domain.HealthData, profile *domain.HealthProfile) []ImprovementArea {
	var areas []ImprovementArea
	
	// BMI
	bmi := profile.GetBMI()
	if bmi != nil && *bmi > 25 {
		areas = append(areas, ImprovementArea{
			Category:    "",
			Description: "BMI?,
			Priority:    "high",
			Actions:     []string{"", "", ""},
		})
	}
	
	// ?
	stepData := s.filterDataByType(healthData, domain.Steps)
	if len(stepData) > 0 {
		avgSteps := 0.0
		for _, data := range stepData {
			avgSteps += data.Value
		}
		avgSteps /= float64(len(stepData))
		
		if avgSteps < 8000 {
			areas = append(areas, ImprovementArea{
				Category:    "?,
				Description: "㽨?,
				Priority:    "medium",
				Actions:     []string{"趨", "", ""},
			})
		}
	}
	
	return areas
}

func (s *HealthAnalysisService) generatePersonalizedTips(profile *domain.HealthProfile, healthData []*domain.HealthData, areas []ImprovementArea) []PersonalizedTip {
	var tips []PersonalizedTip
	
	age := profile.GetAge()
	
	// ?
	if age > 50 {
		tips = append(tips, PersonalizedTip{
			Category:    "?,
			Title:       "",
			Description: "?,
			Priority:    "high",
			Difficulty:  "easy",
		})
	}
	
	// ?
	for _, area := range areas {
		if area.Category == "" {
			tips = append(tips, PersonalizedTip{
				Category:    "",
				Title:       "",
				Description: "300-500",
				Priority:    "high",
				Difficulty:  "medium",
			})
		}
	}
	
	// 
	tips = append(tips, PersonalizedTip{
		Category:    "",
		Title:       "",
		Description: "7-8?,
		Priority:    "medium",
		Difficulty:  "easy",
	})
	
	return tips
}

// 

func (s *HealthAnalysisService) filterDataByType(data []*domain.HealthData, dataType domain.HealthDataType) []*domain.HealthData {
	var filtered []*domain.HealthData
	for _, d := range data {
		if d.DataType == dataType {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func (s *HealthAnalysisService) getMetricStatus(dataType domain.HealthDataType, value float64) string {
	// ?
	switch dataType {
	case domain.HeartRate:
		if value >= 60 && value <= 100 {
			return "good"
		} else if value >= 50 && value <= 120 {
			return "fair"
		}
		return "poor"
	case domain.Steps:
		if value >= 10000 {
			return "excellent"
		} else if value >= 8000 {
			return "good"
		} else if value >= 5000 {
			return "fair"
		}
		return "poor"
	}
	return "fair"
}

func (s *HealthAnalysisService) getMetricTrend(data []*domain.HealthData) string {
	if len(data) < 2 {
		return "stable"
	}
	
	// ?
	recent := data[len(data)-1].Value
	sum := 0.0
	for i := 0; i < len(data)-1; i++ {
		sum += data[i].Value
	}
	previous := sum / float64(len(data)-1)
	
	change := (recent - previous) / previous
	if change > 0.05 {
		return "improving"
	} else if change < -0.05 {
		return "declining"
	}
	return "stable"
}

func (s *HealthAnalysisService) getDefaultUnit(dataType domain.HealthDataType) string {
	switch dataType {
	case domain.HeartRate:
		return "bpm"
	case domain.BloodPressure:
		return "mmHg"
	case domain.Steps:
		return "?
	case domain.SleepDuration:
		return ""
	case domain.StressLevel:
		return "?
	}
	return ""
}

func (s *HealthAnalysisService) getTargetValue(dataType domain.HealthDataType) *float64 {
	var target float64
	switch dataType {
	case domain.HeartRate:
		target = 75
	case domain.Steps:
		target = 10000
	case domain.SleepDuration:
		target = 8
	default:
		return nil
	}
	return &target
}

func (s *HealthAnalysisService) calculatePercentChange(data []*domain.HealthData) float64 {
	if len(data) < 2 {
		return 0
	}
	
	first := data[0].Value
	last := data[len(data)-1].Value
	
	if first == 0 {
		return 0
	}
	
	return ((last - first) / first) * 100
}

func (s *HealthAnalysisService) formatPeriod(start, end time.Time) string {
	days := int(end.Sub(start).Hours() / 24)
	if days <= 7 {
		return ""
	} else if days <= 31 {
		return ""
	} else if days <= 365 {
		return ""
	}
	return ""
}

func (s *HealthAnalysisService) getRiskImpact(riskLevel string) float64 {
	switch riskLevel {
	case "low":
		return 0.1
	case "medium":
		return 0.3
	case "high":
		return 0.6
	case "critical":
		return 1.0
	}
	return 0.1
}

func (s *HealthAnalysisService) getDataTypeRecommendations(dataType domain.HealthDataType) []string {
	switch dataType {
	case domain.HeartRate:
		return []string{"", "?, ""}
	case domain.BloodPressure:
		return []string{"", "?, ""}
	case domain.Steps:
		return []string{"", "趨", ""}
	case domain.SleepDuration:
		return []string{"", "?, "豸"}
	case domain.StressLevel:
		return []string{"?, "", ""}
	}
	return []string{"", ""}
}

