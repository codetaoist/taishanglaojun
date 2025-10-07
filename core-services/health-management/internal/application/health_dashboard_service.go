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

// HealthDashboardService 健康仪表板服务
type HealthDashboardService struct {
	healthDataRepo    domain.HealthDataRepository
	healthProfileRepo domain.HealthProfileRepository
	eventPublisher    EventPublisher
}

// NewHealthDashboardService 创建健康仪表板服务
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

// DashboardPeriod 仪表板时间周期
type DashboardPeriod string

const (
	DashboardPeriodDay   DashboardPeriod = "day"   // 日
	DashboardPeriodWeek  DashboardPeriod = "week"  // 周
	DashboardPeriodMonth DashboardPeriod = "month" // 月
	DashboardPeriodYear  DashboardPeriod = "year"  // 年
)

// HealthDashboard 健康仪表板
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

// HealthOverview 健康概览
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

// KeyMetric 关键指标
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

// TrendChart 趋势图表
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

// ChartDataPoint 图表数据点
type ChartDataPoint struct {
	X     interface{} `json:"x"`
	Y     float64     `json:"y"`
	Label string      `json:"label,omitempty"`
	Color string      `json:"color,omitempty"`
}

// ChartAxis 图表轴
type ChartAxis struct {
	Label  string `json:"label"`
	Unit   string `json:"unit,omitempty"`
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Format string `json:"format,omitempty"`
}

// HealthScore 健康评分
type HealthScore struct {
	Overall      float64                `json:"overall"`
	Categories   map[string]float64     `json:"categories"`
	Trend        string                 `json:"trend"` // improving, declining, stable
	LastScore    float64                `json:"last_score"`
	ScoreHistory []ScoreHistoryPoint    `json:"score_history"`
	Factors      []ScoreFactor          `json:"factors"`
}

// ScoreHistoryPoint 评分历史点
type ScoreHistoryPoint struct {
	Date  time.Time `json:"date"`
	Score float64   `json:"score"`
}

// ScoreFactor 评分因子
type ScoreFactor struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Score       float64 `json:"score"`
	Impact      string  `json:"impact"` // positive, negative, neutral
	Description string  `json:"description,omitempty"`
}

// DashboardRecommendation 仪表板建议
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

// DashboardAlert 仪表板警报
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

// Achievement 成就
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

// HealthGoal 健康目标
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

// GetDashboardRequest 获取仪表板请求
type GetDashboardRequest struct {
	UserID uuid.UUID       `json:"user_id" binding:"required"`
	Period DashboardPeriod `json:"period,omitempty"`
}

// GetDashboardResponse 获取仪表板响应
type GetDashboardResponse struct {
	Dashboard HealthDashboard `json:"dashboard"`
	Timestamp time.Time       `json:"timestamp"`
}

// GetDashboardSummaryRequest 获取仪表板摘要请求
type GetDashboardSummaryRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// GetDashboardSummaryResponse 获取仪表板摘要响应
type GetDashboardSummaryResponse struct {
	Summary   HealthOverview `json:"summary"`
	Timestamp time.Time      `json:"timestamp"`
}

// GetDashboard 获取健康仪表板
func (s *HealthDashboardService) GetDashboard(ctx context.Context, req *GetDashboardRequest) (*GetDashboardResponse, error) {
	// 设置默认周期
	if req.Period == "" {
		req.Period = DashboardPeriodMonth
	}

	// 获取用户健康档案
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// 计算时间范围
	endTime := time.Now()
	startTime := s.calculateStartTime(endTime, req.Period)

	// 构建仪表板
	dashboard := HealthDashboard{
		UserID:      req.UserID,
		Period:      req.Period,
		LastUpdated: time.Now(),
	}

	// 生成各个组件
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

// GetDashboardSummary 获取仪表板摘要
func (s *HealthDashboardService) GetDashboardSummary(ctx context.Context, req *GetDashboardSummaryRequest) (*GetDashboardSummaryResponse, error) {
	// 获取用户健康档案
	profile, err := s.healthProfileRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// 计算时间范围（最近30天）
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	// 生成概览
	overview := s.generateOverview(ctx, req.UserID, startTime, endTime, profile)

	return &GetDashboardSummaryResponse{
		Summary:   overview,
		Timestamp: time.Now(),
	}, nil
}

// generateOverview 生成健康概览
func (s *HealthDashboardService) generateOverview(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, profile *domain.HealthProfile) HealthOverview {
	// 获取所有健康数据
	allData, _ := s.healthDataRepo.GetByUserID(ctx, userID, startTime, endTime)

	// 计算活跃天数
	activeDays := s.calculateActiveDays(allData, startTime, endTime)

	// 计算整体评分
	overallScore := s.calculateOverallScore(allData)

	// 识别风险因素
	riskFactors := s.identifyRiskFactors(allData, profile)

	// 计算健康类别评分
	healthCategories := s.calculateHealthCategories(allData)

	// 生成摘要
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

// generateKeyMetrics 生成关键指标
func (s *HealthDashboardService) generateKeyMetrics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) []KeyMetric {
	var metrics []KeyMetric

	// 步数指标
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
			Name:        "日均步数",
			Value:       avgSteps,
			Unit:        "步",
			Change:      change,
			ChangeType:  changeType,
			Status:      status,
			Target:      &target,
			Icon:        "👟",
			Color:       s.getStatusColor(status),
			Description: "每日平均步数统计",
		})
	}

	// 心率指标
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
			Name:        "平均心率",
			Value:       avgHeartRate,
			Unit:        "bpm",
			Status:      status,
			Icon:        "❤️",
			Color:       s.getStatusColor(status),
			Description: "静息心率平均值",
		})
	}

	// 睡眠指标
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
			Name:        "平均睡眠",
			Value:       avgSleep,
			Unit:        "小时",
			Status:      status,
			Target:      &target,
			Icon:        "😴",
			Color:       s.getStatusColor(status),
			Description: "每日平均睡眠时长",
		})
	}

	return metrics
}

// generateTrendCharts 生成趋势图表
func (s *HealthDashboardService) generateTrendCharts(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, period DashboardPeriod) []TrendChart {
	var charts []TrendChart

	// 步数趋势图
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	if len(stepsData) > 0 {
		chartData := s.convertToChartData(stepsData, period)
		charts = append(charts, TrendChart{
			ID:       "steps_trend",
			Title:    "步数趋势",
			Type:     "line",
			DataType: "steps",
			Data:     chartData,
			XAxis:    ChartAxis{Label: "时间", Format: s.getTimeFormat(period)},
			YAxis:    ChartAxis{Label: "步数", Unit: "步"},
		})
	}

	// 心率趋势图
	heartRateData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "heart_rate", startTime, endTime)
	if len(heartRateData) > 0 {
		chartData := s.convertToChartData(heartRateData, period)
		charts = append(charts, TrendChart{
			ID:       "heart_rate_trend",
			Title:    "心率趋势",
			Type:     "line",
			DataType: "heart_rate",
			Data:     chartData,
			XAxis:    ChartAxis{Label: "时间", Format: s.getTimeFormat(period)},
			YAxis:    ChartAxis{Label: "心率", Unit: "bpm"},
		})
	}

	// 睡眠趋势图
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)
	if len(sleepData) > 0 {
		chartData := s.convertToChartData(sleepData, period)
		charts = append(charts, TrendChart{
			ID:       "sleep_trend",
			Title:    "睡眠趋势",
			Type:     "bar",
			DataType: "sleep_duration",
			Data:     chartData,
			XAxis:    ChartAxis{Label: "时间", Format: s.getTimeFormat(period)},
			YAxis:    ChartAxis{Label: "睡眠时长", Unit: "小时"},
		})
	}

	return charts
}

// generateHealthScore 生成健康评分
func (s *HealthDashboardService) generateHealthScore(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) HealthScore {
	// 获取各类健康数据
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	heartRateData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "heart_rate", startTime, endTime)
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)

	// 计算各类别评分
	categories := make(map[string]float64)
	factors := []ScoreFactor{}

	// 运动评分
	if len(stepsData) > 0 {
		avgSteps := s.calculateAverage(stepsData)
		exerciseScore := math.Min(avgSteps/10000*100, 100)
		categories["exercise"] = exerciseScore
		factors = append(factors, ScoreFactor{
			Name:        "运动活动",
			Weight:      0.3,
			Score:       exerciseScore,
			Impact:      s.getScoreImpact(exerciseScore),
			Description: "基于日均步数的运动评分",
		})
	}

	// 心血管评分
	if len(heartRateData) > 0 {
		avgHeartRate := s.calculateAverage(heartRateData)
		cardioScore := s.calculateCardioScore(avgHeartRate)
		categories["cardiovascular"] = cardioScore
		factors = append(factors, ScoreFactor{
			Name:        "心血管健康",
			Weight:      0.25,
			Score:       cardioScore,
			Impact:      s.getScoreImpact(cardioScore),
			Description: "基于心率的心血管健康评分",
		})
	}

	// 睡眠评分
	if len(sleepData) > 0 {
		avgSleep := s.calculateAverage(sleepData)
		sleepScore := s.calculateSleepScore(avgSleep)
		categories["sleep"] = sleepScore
		factors = append(factors, ScoreFactor{
			Name:        "睡眠质量",
			Weight:      0.25,
			Score:       sleepScore,
			Impact:      s.getScoreImpact(sleepScore),
			Description: "基于睡眠时长的睡眠质量评分",
		})
	}

	// 计算总体评分
	overall := s.calculateWeightedScore(factors)

	// 获取历史评分（简化实现）
	scoreHistory := []ScoreHistoryPoint{
		{Date: startTime, Score: overall - 5},
		{Date: endTime, Score: overall},
	}

	// 确定趋势
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

// generateDashboardRecommendations 生成仪表板建议
func (s *HealthDashboardService) generateDashboardRecommendations(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) []DashboardRecommendation {
	var recommendations []DashboardRecommendation

	// 获取步数数据
	stepsData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "steps", startTime, endTime)
	if len(stepsData) > 0 {
		avgSteps := s.calculateAverage(stepsData)
		if avgSteps < 5000 {
			recommendations = append(recommendations, DashboardRecommendation{
				ID:          uuid.New(),
				Type:        "exercise",
				Priority:    "high",
				Title:       "增加日常活动",
				Description: "您的日均步数较低，建议增加日常活动量",
				Action:      "设定每日8000步目标",
				Icon:        "👟",
				Color:       "#ff6b6b",
			})
		}
	}

	// 获取睡眠数据
	sleepData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "sleep_duration", startTime, endTime)
	if len(sleepData) > 0 {
		avgSleep := s.calculateAverage(sleepData)
		if avgSleep < 7 {
			recommendations = append(recommendations, DashboardRecommendation{
				ID:          uuid.New(),
				Type:        "sleep",
				Priority:    "medium",
				Title:       "改善睡眠质量",
				Description: "您的睡眠时间不足，建议调整作息",
				Action:      "建立规律睡眠时间",
				Icon:        "😴",
				Color:       "#4ecdc4",
			})
		}
	}

	return recommendations
}

// generateDashboardAlerts 生成仪表板警报
func (s *HealthDashboardService) generateDashboardAlerts(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) []DashboardAlert {
	var alerts []DashboardAlert

	// 检查心率异常
	heartRateData, _ := s.healthDataRepo.GetByUserIDAndType(ctx, userID, "heart_rate", startTime, endTime)
	for _, data := range heartRateData {
		if data.Value > 120 {
			alerts = append(alerts, DashboardAlert{
				ID:        uuid.New(),
				Type:      "heart_rate",
				Severity:  "warning",
				Title:     "心率偏高",
				Message:   fmt.Sprintf("检测到心率异常：%.0f bpm", data.Value),
				Timestamp: data.RecordedAt,
				IsRead:    false,
				Icon:      "⚠️",
				Color:     "#ff9f43",
			})
		}
	}

	return alerts
}

// generateAchievements 生成成就
func (s *HealthDashboardService) generateAchievements(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) []Achievement {
	var achievements []Achievement

	// 步数成就
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
				Name:        "步行达人",
				Description: "单日步数达到10000步",
				Type:        "exercise",
				Progress:    100,
				Target:      10000,
				IsCompleted: true,
				Icon:        "🏆",
				Badge:       "gold",
			})
		}
	}

	return achievements
}

// generateHealthGoals 生成健康目标
func (s *HealthDashboardService) generateHealthGoals(ctx context.Context, userID uuid.UUID) []HealthGoal {
	var goals []HealthGoal

	// 默认目标
	goals = append(goals, HealthGoal{
		ID:          uuid.New(),
		Name:        "每日步数目标",
		Description: "每天走路10000步",
		Type:        "exercise",
		Target:      10000,
		Current:     7500, // 示例值
		Progress:    75,
		Status:      "active",
		CreatedAt:   time.Now().AddDate(0, 0, -30),
	})

	goals = append(goals, HealthGoal{
		ID:          uuid.New(),
		Name:        "睡眠时长目标",
		Description: "每天睡眠8小时",
		Type:        "sleep",
		Target:      8,
		Current:     7.2, // 示例值
		Progress:    90,
		Status:      "active",
		CreatedAt:   time.Now().AddDate(0, 0, -30),
	})

	return goals
}

// 辅助函数
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
	// 简化的评分算法
	return math.Min(float64(len(data))/10*100, 100)
}

func (s *HealthDashboardService) identifyRiskFactors(data []*domain.HealthData, profile *domain.HealthProfile) []string {
	var factors []string
	
	// 基于数据识别风险因素
	for _, d := range data {
		if d.DataType == "heart_rate" && d.Value > 100 {
			factors = append(factors, "心率偏高")
			break
		}
	}

	// 基于档案识别风险因素
	if profile != nil {
		bmi := profile.GetBMI()
		if bmi > 25 {
			factors = append(factors, "体重超标")
		}
		if profile.GetAge() > 50 {
			factors = append(factors, "年龄相关风险")
		}
	}

	return factors
}

func (s *HealthDashboardService) calculateHealthCategories(data []*domain.HealthData) map[string]float64 {
	categories := make(map[string]float64)
	
	// 按数据类型分组计算评分
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
		return 75 // 默认评分
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
		return fmt.Sprintf("您的健康状况良好，最近%d天保持了良好的活跃度", activeDays)
	} else if score >= 60 {
		return fmt.Sprintf("您的健康状况一般，建议关注%d个风险因素", len(riskFactors))
	} else {
		return "您的健康状况需要改善，建议咨询医生并调整生活方式"
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

	// 按时间排序
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