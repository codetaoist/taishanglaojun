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

// HealthAnalysisService 健康分析应用服务
type HealthAnalysisService struct {
	healthDataRepo    domain.HealthDataRepository
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthAnalysisService 创建健康分析服务
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

// HealthTrendAnalysisRequest 健康趋势分析请求
type HealthTrendAnalysisRequest struct {
	UserID    uuid.UUID             `json:"user_id" validate:"required"`
	DataType  domain.HealthDataType `json:"data_type" validate:"required"`
	StartTime time.Time             `json:"start_time" validate:"required"`
	EndTime   time.Time             `json:"end_time" validate:"required"`
	Period    string                `json:"period" validate:"required,oneof=daily weekly monthly"`
}

// HealthTrendAnalysisResponse 健康趋势分析响应
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

// TrendDataPoint 趋势数据�?
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Count     int       `json:"count"`
}

// HealthRiskAssessmentRequest 健康风险评估请求
type HealthRiskAssessmentRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// HealthRiskAssessmentResponse 健康风险评估响应
type HealthRiskAssessmentResponse struct {
	UserID           uuid.UUID    `json:"user_id"`
	OverallRiskLevel string       `json:"overall_risk_level"` // low, medium, high, critical
	RiskScore        float64      `json:"risk_score"` // 0-100
	RiskFactors      []RiskFactor `json:"risk_factors"`
	Recommendations  []string     `json:"recommendations"`
	NextCheckupDate  *time.Time   `json:"next_checkup_date,omitempty"`
	AssessedAt       time.Time    `json:"assessed_at"`
}

// RiskFactor 风险因素
type RiskFactor struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	RiskLevel   string  `json:"risk_level"`
	Impact      float64 `json:"impact"` // 0-1
	Suggestions []string `json:"suggestions"`
}

// HealthInsightsRequest 健康洞察请求
type HealthInsightsRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

// HealthInsightsResponse 健康洞察响应
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

// KeyMetric 关键指标
type KeyMetric struct {
	Name         string  `json:"name"`
	Value        float64 `json:"value"`
	Unit         string  `json:"unit"`
	Status       string  `json:"status"` // excellent, good, fair, poor
	Trend        string  `json:"trend"` // improving, stable, declining
	TargetValue  *float64 `json:"target_value,omitempty"`
	PercentChange float64 `json:"percent_change"`
}

// Achievement 成就
type Achievement struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	EarnedAt    time.Time `json:"earned_at"`
	Icon        string    `json:"icon"`
}

// ImprovementArea 改进领域
type ImprovementArea struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"` // high, medium, low
	Actions     []string `json:"actions"`
}

// PersonalizedTip 个性化建议
type PersonalizedTip struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Difficulty  string `json:"difficulty"` // easy, medium, hard
}

// AnalyzeHealthTrend 分析健康趋势
func (s *HealthAnalysisService) AnalyzeHealthTrend(ctx context.Context, req *HealthTrendAnalysisRequest) (*HealthTrendAnalysisResponse, error) {
	// 获取指定时间范围内的健康数据
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
			Insights: []string{"暂无足够数据进行趋势分析"},
		}, nil
	}

	// 按时间段聚合数据
	trendData := s.aggregateDataByPeriod(healthData, req.Period)
	
	// 计算统计指标
	stats := s.calculateStatistics(trendData)
	
	// 分析趋势类型和强�?
	trendType, trendStrength := s.analyzeTrendPattern(trendData)
	
	// 生成洞察和建�?
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

// AssessHealthRisk 评估健康风险
func (s *HealthAnalysisService) AssessHealthRisk(ctx context.Context, req *HealthRiskAssessmentRequest) (*HealthRiskAssessmentResponse, error) {
	// 获取用户健康档案
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch health profile: %w", err)
	}

	// 获取最近的健康数据
	endTime := time.Now()
	startTime := endTime.AddDate(0, -3, 0) // 最�?个月
	
	healthData, err := s.healthDataRepo.FindByUserIDAndTimeRange(ctx, req.UserID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent health data: %w", err)
	}

	// 评估各种风险因素
	riskFactors := s.assessRiskFactors(profile, healthData)
	
	// 计算总体风险分数
	riskScore := s.calculateOverallRiskScore(riskFactors)
	
	// 确定风险等级
	riskLevel := s.determineRiskLevel(riskScore)
	
	// 生成建议
	recommendations := s.generateRiskRecommendations(riskFactors, riskLevel)
	
	// 计算下次检查日�?
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

// GenerateHealthInsights 生成健康洞察
func (s *HealthAnalysisService) GenerateHealthInsights(ctx context.Context, req *HealthInsightsRequest) (*HealthInsightsResponse, error) {
	// 获取用户健康档案
	profile, err := s.healthProfileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch health profile: %w", err)
	}

	// 获取指定时间范围内的健康数据
	healthData, err := s.healthDataRepo.FindByUserIDAndTimeRange(ctx, req.UserID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch health data: %w", err)
	}

	// 计算整体健康分数
	healthScore := s.calculateOverallHealthScore(profile, healthData)
	
	// 确定健康状�?
	healthStatus := s.determineHealthStatus(healthScore)
	
	// 分析关键指标
	keyMetrics := s.analyzeKeyMetrics(healthData, req.StartTime, req.EndTime)
	
	// 识别成就
	achievements := s.identifyAchievements(healthData, profile)
	
	// 识别改进领域
	improvementAreas := s.identifyImprovementAreas(healthData, profile)
	
	// 生成个性化建议
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

// 辅助方法

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
			// 简化处理，使用周一作为代表
			timestamp, err = time.Parse("2006-W02", key)
		case "monthly":
			timestamp, err = time.Parse("2006-01", key)
		}
		
		if err != nil {
			continue
		}
		
		// 计算平均�?
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
	
	// 按时间排�?
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
	
	// 计算标准�?
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
	
	// 使用线性回归分析趋�?
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
	
	// 计算相关系数作为趋势强度
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
	
	// 确定趋势类型
	var trendType string
	if math.Abs(slope) < 0.1 {
		trendType = "stable"
	} else if slope > 0 {
		trendType = "increasing"
	} else {
		trendType = "decreasing"
	}
	
	// 如果变化很大但没有明显趋势，标记为波�?
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
			insights = append(insights, "您的心率呈上升趋势，建议关注心血管健�?)
		case "decreasing":
			insights = append(insights, "您的心率呈下降趋势，这通常是健康改善的好兆�?)
		case "stable":
			insights = append(insights, "您的心率保持稳定，继续保持良好的生活习惯")
		case "fluctuating":
			insights = append(insights, "您的心率波动较大，建议规律作息和适度运动")
		}
	case domain.BloodPressure:
		switch trendType {
		case "increasing":
			insights = append(insights, "血压有上升趋势，建议控制盐分摄入并增加运动")
		case "decreasing":
			insights = append(insights, "血压呈下降趋势，说明您的健康管理很有效�?)
		case "stable":
			insights = append(insights, "血压保持稳定，继续维持健康的生活方�?)
		}
	case domain.Steps:
		switch trendType {
		case "increasing":
			insights = append(insights, "您的运动量在增加，这对健康非常有�?)
		case "decreasing":
			insights = append(insights, "运动量有所减少，建议增加日常活�?)
		case "stable":
			insights = append(insights, "运动量保持稳定，继续保持活跃的生活方�?)
		}
	}
	
	return insights
}

func (s *HealthAnalysisService) generateTrendRecommendations(dataType domain.HealthDataType, trendType string, stats Statistics) []string {
	var recommendations []string
	
	switch dataType {
	case domain.HeartRate:
		if trendType == "increasing" {
			recommendations = append(recommendations, "建议进行有氧运动来改善心血管健�?)
			recommendations = append(recommendations, "考虑减少咖啡因摄�?)
		}
	case domain.BloodPressure:
		if trendType == "increasing" {
			recommendations = append(recommendations, "减少钠盐摄入，增加钾元素丰富的食�?)
			recommendations = append(recommendations, "保持规律的有氧运�?)
		}
	case domain.Steps:
		if trendType == "decreasing" {
			recommendations = append(recommendations, "设定每日步数目标，逐步增加运动�?)
			recommendations = append(recommendations, "尝试步行或骑自行车上下班")
		}
	}
	
	return recommendations
}

func (s *HealthAnalysisService) assessRiskFactors(profile *domain.HealthProfile, healthData []*domain.HealthData) []RiskFactor {
	var riskFactors []RiskFactor
	
	// 年龄风险评估
	age := profile.GetAge()
	if age > 65 {
		riskFactors = append(riskFactors, RiskFactor{
			Category:    "年龄",
			Description: "年龄超过65岁，需要更密切的健康监�?,
			RiskLevel:   "medium",
			Impact:      0.3,
			Suggestions: []string{"定期体检", "保持活跃的生活方�?},
		})
	}
	
	// BMI风险评估
	bmi := profile.GetBMI()
	if bmi != nil {
		if *bmi > 30 {
			riskFactors = append(riskFactors, RiskFactor{
				Category:    "体重",
				Description: "BMI超过30，属于肥胖范�?,
				RiskLevel:   "high",
				Impact:      0.4,
				Suggestions: []string{"控制饮食", "增加运动", "咨询营养�?},
			})
		} else if *bmi > 25 {
			riskFactors = append(riskFactors, RiskFactor{
				Category:    "体重",
				Description: "BMI超过25，属于超重范�?,
				RiskLevel:   "medium",
				Impact:      0.2,
				Suggestions: []string{"适度减重", "均衡饮食"},
			})
		}
	}
	
	// 分析健康数据中的异常�?
	for _, data := range healthData {
		if data.IsAbnormal() {
			riskLevel := string(data.GetRiskLevel())
			riskFactors = append(riskFactors, RiskFactor{
				Category:    string(data.DataType),
				Description: fmt.Sprintf("%s数值异�?, data.DataType),
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
		return 10.0 // 基础风险分数
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
		totalImpact += factor.Impact * multiplier * 20 // 转换�?-100分数
	}
	
	// 确保分数�?-100范围�?
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
	
	// 基于风险等级的通用建议
	switch riskLevel {
	case "low":
		recommendations = append(recommendations, "继续保持健康的生活方�?)
		recommendations = append(recommendations, "定期进行健康检�?)
	case "medium":
		recommendations = append(recommendations, "加强健康监测频率")
		recommendations = append(recommendations, "考虑咨询医生制定健康计划")
	case "high":
		recommendations = append(recommendations, "建议尽快咨询医生")
		recommendations = append(recommendations, "密切监测健康指标")
	case "critical":
		recommendations = append(recommendations, "立即寻求医疗帮助")
		recommendations = append(recommendations, "严格遵循医生建议")
	}
	
	// 基于具体风险因素的建�?
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
	
	// 考虑年龄因素
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
	score := 100.0 // 起始分数
	
	// 基于年龄调整
	age := profile.GetAge()
	if age > 65 {
		score -= 10
	} else if age > 50 {
		score -= 5
	}
	
	// 基于BMI调整
	bmi := profile.GetBMI()
	if bmi != nil {
		if *bmi > 30 || *bmi < 18.5 {
			score -= 15
		} else if *bmi > 25 || *bmi < 20 {
			score -= 5
		}
	}
	
	// 基于健康数据异常情况调整
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
	
	// 确保分数�?-100范围�?
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
	
	// 按数据类型分�?
	dataByType := make(map[domain.HealthDataType][]*domain.HealthData)
	for _, data := range healthData {
		dataByType[data.DataType] = append(dataByType[data.DataType], data)
	}
	
	// 分析每种数据类型
	for dataType, data := range dataByType {
		if len(data) == 0 {
			continue
		}
		
		// 计算平均�?
		sum := 0.0
		for _, d := range data {
			sum += d.Value
		}
		average := sum / float64(len(data))
		
		// 确定状态和趋势
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
	
	// 分析步数成就
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
				Title:       "万步达人",
				Description: "单日步数超过10,000�?,
				Category:    "运动",
				EarnedAt:    time.Now(),
				Icon:        "🚶‍♂�?,
			})
		}
	}
	
	// 分析连续记录成就
	if len(healthData) >= 30 {
		achievements = append(achievements, Achievement{
			Title:       "坚持记录",
			Description: "连续记录健康数据30�?,
			Category:    "习惯",
			EarnedAt:    time.Now(),
			Icon:        "📊",
		})
	}
	
	return achievements
}

func (s *HealthAnalysisService) identifyImprovementAreas(healthData []*domain.HealthData, profile *domain.HealthProfile) []ImprovementArea {
	var areas []ImprovementArea
	
	// 分析BMI
	bmi := profile.GetBMI()
	if bmi != nil && *bmi > 25 {
		areas = append(areas, ImprovementArea{
			Category:    "体重管理",
			Description: "BMI超出正常范围，建议控制体�?,
			Priority:    "high",
			Actions:     []string{"制定减重计划", "控制饮食", "增加运动"},
		})
	}
	
	// 分析运动�?
	stepData := s.filterDataByType(healthData, domain.Steps)
	if len(stepData) > 0 {
		avgSteps := 0.0
		for _, data := range stepData {
			avgSteps += data.Value
		}
		avgSteps /= float64(len(stepData))
		
		if avgSteps < 8000 {
			areas = append(areas, ImprovementArea{
				Category:    "运动�?,
				Description: "日均步数不足，建议增加运�?,
				Priority:    "medium",
				Actions:     []string{"设定步数目标", "增加户外活动", "使用楼梯代替电梯"},
			})
		}
	}
	
	return areas
}

func (s *HealthAnalysisService) generatePersonalizedTips(profile *domain.HealthProfile, healthData []*domain.HealthData, areas []ImprovementArea) []PersonalizedTip {
	var tips []PersonalizedTip
	
	age := profile.GetAge()
	
	// 基于年龄的建�?
	if age > 50 {
		tips = append(tips, PersonalizedTip{
			Category:    "健康检�?,
			Title:       "定期体检",
			Description: "建议每年进行全面体检，重点关注心血管和骨密�?,
			Priority:    "high",
			Difficulty:  "easy",
		})
	}
	
	// 基于改进领域的建�?
	for _, area := range areas {
		if area.Category == "体重管理" {
			tips = append(tips, PersonalizedTip{
				Category:    "饮食",
				Title:       "控制热量摄入",
				Description: "建议每日减少300-500卡路里摄入，配合适度运动",
				Priority:    "high",
				Difficulty:  "medium",
			})
		}
	}
	
	// 通用健康建议
	tips = append(tips, PersonalizedTip{
		Category:    "生活方式",
		Title:       "保持充足睡眠",
		Description: "每晚保证7-8小时优质睡眠，有助于身体恢复和健康维�?,
		Priority:    "medium",
		Difficulty:  "easy",
	})
	
	return tips
}

// 辅助方法

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
	// 根据数据类型和值确定状�?
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
	
	// 简单的趋势分析：比较最近的值和之前的平均�?
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
		return "�?
	case domain.SleepDuration:
		return "小时"
	case domain.StressLevel:
		return "�?
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
		return "本周"
	} else if days <= 31 {
		return "本月"
	} else if days <= 365 {
		return "近期"
	}
	return "长期"
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
		return []string{"进行有氧运动", "减少咖啡因摄�?, "保持充足睡眠"}
	case domain.BloodPressure:
		return []string{"减少盐分摄入", "增加钾元素食�?, "规律运动"}
	case domain.Steps:
		return []string{"增加日常活动", "设定步数目标", "选择步行出行"}
	case domain.SleepDuration:
		return []string{"建立规律作息", "创造良好睡眠环�?, "避免睡前使用电子设备"}
	case domain.StressLevel:
		return []string{"学习放松技�?, "进行冥想练习", "寻求专业帮助"}
	}
	return []string{"咨询医生", "保持健康生活方式"}
}
