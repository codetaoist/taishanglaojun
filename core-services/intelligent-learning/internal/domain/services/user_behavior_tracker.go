package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

// UserBehaviorTracker 用户行为跟踪�?
type UserBehaviorTracker struct {
	behaviorRepo BehaviorRepository
	eventStore   EventStore
}

// BehaviorRepository 行为数据仓库接口
type BehaviorRepository interface {
	SaveBehaviorEvent(ctx context.Context, event *BehaviorEvent) error
	GetBehaviorEvents(ctx context.Context, learnerID uuid.UUID, timeRange BehaviorTimeRange) ([]*BehaviorEvent, error)
	GetBehaviorSummary(ctx context.Context, learnerID uuid.UUID, timeRange BehaviorTimeRange) (*BehaviorSummary, error)
	GetEngagementMetrics(ctx context.Context, learnerID uuid.UUID, timeRange BehaviorTimeRange) (*EngagementMetrics, error)
}

// EventStore 事件存储接口
type EventStore interface {
	StoreEvent(ctx context.Context, event interface{}) error
	GetEvents(ctx context.Context, filter EventFilter) ([]interface{}, error)
}

// NewUserBehaviorTracker 创建用户行为跟踪�?
func NewUserBehaviorTracker(behaviorRepo BehaviorRepository, eventStore EventStore) *UserBehaviorTracker {
	return &UserBehaviorTracker{
		behaviorRepo: behaviorRepo,
		eventStore:   eventStore,
	}
}

// BehaviorEvent 行为事件
type BehaviorEvent struct {
	ID          uuid.UUID              `json:"id"`
	LearnerID   uuid.UUID              `json:"learner_id"`
	SessionID   uuid.UUID              `json:"session_id"`
	EventType   string                 `json:"event_type"`
	ContentID   *uuid.UUID             `json:"content_id,omitempty"`
	PathID      *uuid.UUID             `json:"path_id,omitempty"`
	ConceptID   *uuid.UUID             `json:"concept_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Properties  map[string]interface{} `json:"properties"`
	Context     *EventContext          `json:"context"`
	Performance *TrackerPerformanceData       `json:"performance,omitempty"`
}

// EventContext 事件上下�?
type EventContext struct {
	Device       string                 `json:"device"`
	Platform     string                 `json:"platform"`
	Location     string                 `json:"location"`
	NetworkType  string                 `json:"network_type"`
	UserAgent    string                 `json:"user_agent"`
	Referrer     string                 `json:"referrer"`
	TimeZone     string                 `json:"time_zone"`
	Language     string                 `json:"language"`
	ScreenSize   string                 `json:"screen_size"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// PerformanceData 性能数据
type TrackerPerformanceData struct {
	Score           float64   `json:"score"`
	Accuracy        float64   `json:"accuracy"`
	CompletionRate  float64   `json:"completion_rate"`
	TimeSpent       time.Duration `json:"time_spent"`
	AttemptsCount   int       `json:"attempts_count"`
	HintsUsed       int       `json:"hints_used"`
	ErrorsCount     int       `json:"errors_count"`
	SkipCount       int       `json:"skip_count"`
}

// BehaviorSummary 行为摘要
type BehaviorSummary struct {
	LearnerID           uuid.UUID              `json:"learner_id"`
	TimeRange           BehaviorTimeRange              `json:"time_range"`
	TotalSessions       int                    `json:"total_sessions"`
	TotalTimeSpent      time.Duration          `json:"total_time_spent"`
	AverageSessionTime  time.Duration          `json:"average_session_time"`
	ContentInteractions int                    `json:"content_interactions"`
	CompletionRate      float64                `json:"completion_rate"`
	EngagementScore     float64                `json:"engagement_score"`
	LearningVelocity    float64                `json:"learning_velocity"`
	PreferredTimes      []BehaviorTimeSlot             `json:"preferred_times"`
	PreferredDevices    map[string]int         `json:"preferred_devices"`
	TopicInteractions   map[string]int         `json:"topic_interactions"`
	BehaviorPatterns    []BehaviorPattern      `json:"behavior_patterns"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// EngagementMetrics 参与度指�?
type EngagementMetrics struct {
	OverallEngagement   float64                `json:"overall_engagement"`
	ContentEngagement   map[string]float64     `json:"content_engagement"`
	SessionEngagement   []SessionEngagement    `json:"session_engagement"`
	EngagementTrends    []EngagementTrend      `json:"engagement_trends"`
	DropoffPoints       []DropoffPoint         `json:"dropoff_points"`
	PeakEngagementTimes []time.Time            `json:"peak_engagement_times"`
	EngagementFactors   map[string]float64     `json:"engagement_factors"`
}

// SessionEngagement 会话参与�?
type SessionEngagement struct {
	SessionID       uuid.UUID     `json:"session_id"`
	StartTime       time.Time     `json:"start_time"`
	Duration        time.Duration `json:"duration"`
	EngagementScore float64       `json:"engagement_score"`
	InteractionCount int          `json:"interaction_count"`
	FocusScore      float64       `json:"focus_score"`
	CompletionRate  float64       `json:"completion_rate"`
}

// EngagementTrend 参与度趋�?
type EngagementTrend struct {
	Date            time.Time `json:"date"`
	EngagementScore float64   `json:"engagement_score"`
	SessionCount    int       `json:"session_count"`
	TotalTime       time.Duration `json:"total_time"`
}

// BehaviorTimeRange 行为时间范围
type BehaviorTimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// BehaviorTimeSlot 行为时间�?
type BehaviorTimeSlot struct {
	Hour      int     `json:"hour"`
	Frequency int     `json:"frequency"`
	Score     float64 `json:"score"`
}

// EventFilter 事件过滤�?
type EventFilter struct {
	LearnerID   *uuid.UUID `json:"learner_id,omitempty"`
	EventTypes  []string   `json:"event_types,omitempty"`
	TimeRange   *BehaviorTimeRange `json:"time_range,omitempty"`
	ContentID   *uuid.UUID `json:"content_id,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// TrackBehaviorEvent 跟踪行为事件
func (ubt *UserBehaviorTracker) TrackBehaviorEvent(ctx context.Context, event *BehaviorEvent) error {
	// 设置事件ID和时间戳
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 保存到行为仓�?
	if err := ubt.behaviorRepo.SaveBehaviorEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save behavior event: %w", err)
	}

	// 存储到事件存�?
	if err := ubt.eventStore.StoreEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

// AnalyzeBehaviorPattern 分析行为模式
func (ubt *UserBehaviorTracker) AnalyzeBehaviorPattern(ctx context.Context, learnerID uuid.UUID) (*BehaviorPattern, error) {
	// 获取最�?0天的行为数据
	timeRange := BehaviorTimeRange{
		Start: time.Now().AddDate(0, 0, -30),
		End:   time.Now(),
	}

	summary, err := ubt.behaviorRepo.GetBehaviorSummary(ctx, learnerID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}

	engagement, err := ubt.behaviorRepo.GetEngagementMetrics(ctx, learnerID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get engagement metrics: %w", err)
	}

	// 分析行为模式
	pattern := &BehaviorPattern{
		EngagementLevel:      engagement.OverallEngagement,
		CompletionRate:       summary.CompletionRate,
		SessionFrequency:     ubt.calculateSessionFrequency(summary),
		AverageSessionTime:   summary.AverageSessionTime,
		InteractionPatterns:  ubt.analyzeInteractionPatterns(ctx, learnerID, timeRange),
		DropoffPoints:        engagement.DropoffPoints,
		PeakPerformanceTimes: engagement.PeakEngagementTimes,
	}

	return pattern, nil
}

// GetLearningInsights 获取学习洞察
func (ubt *UserBehaviorTracker) GetLearningInsights(ctx context.Context, learnerID uuid.UUID, timeRange BehaviorTimeRange) (*LearningInsights, error) {
	events, err := ubt.behaviorRepo.GetBehaviorEvents(ctx, learnerID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior events: %w", err)
	}

	insights := &LearningInsights{
		LearnerID:         learnerID,
		TimeRange:         timeRange,
		GeneratedAt:       time.Now(),
		LearningPatterns:  ubt.identifyLearningPatterns(events),
		PerformanceTrends: ubt.analyzePerformanceTrends(events),
		EngagementInsights: ubt.analyzeEngagementInsights(events),
		RecommendedActions: ubt.generateRecommendedActions(events),
		PredictiveInsights: ubt.generatePredictiveInsights(events),
	}

	return insights, nil
}

// LearningInsights 学习洞察
type LearningInsights struct {
	LearnerID          uuid.UUID              `json:"learner_id"`
	TimeRange          BehaviorTimeRange              `json:"time_range"`
	GeneratedAt        time.Time              `json:"generated_at"`
	LearningPatterns   []LearningPattern      `json:"learning_patterns"`
	PerformanceTrends  []BehaviorPerformanceTrend     `json:"performance_trends"`
	EngagementInsights []EngagementInsight    `json:"engagement_insights"`
	RecommendedActions []BehaviorRecommendedAction    `json:"recommended_actions"`
	PredictiveInsights []PredictiveInsight    `json:"predictive_insights"`
}

// LearningPattern 学习模式
type LearningPattern struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Evidence    []string               `json:"evidence"`
	Impact      string                 `json:"impact"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BehaviorPerformanceTrend 行为性能趋势
type BehaviorPerformanceTrend struct {
	Metric      string    `json:"metric"`
	Trend       string    `json:"trend"` // increasing, decreasing, stable
	Value       float64   `json:"value"`
	Change      float64   `json:"change"`
	Period      BehaviorTimeRange `json:"period"`
	Confidence  float64   `json:"confidence"`
}

// EngagementInsight 参与度洞�?
type EngagementInsight struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Impact      float64                `json:"impact"`
	Suggestions []string               `json:"suggestions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BehaviorRecommendedAction 行为推荐行动
type BehaviorRecommendedAction struct {
	ID              uuid.UUID              `json:"id"`
	Type            string                 `json:"type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Priority        int                    `json:"priority"`
	ExpectedImpact  float64                `json:"expected_impact"`
	Timeline        string                 `json:"timeline"`
	Implementation  []string               `json:"implementation"`
	SuccessMetrics  []string               `json:"success_metrics"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PredictiveInsight 预测洞察
type PredictiveInsight struct {
	Type           string                 `json:"type"`
	Prediction     string                 `json:"prediction"`
	Confidence     float64                `json:"confidence"`
	TimeHorizon    time.Duration          `json:"time_horizon"`
	Factors        []string               `json:"factors"`
	Recommendations []string              `json:"recommendations"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// 辅助方法实现

func (ubt *UserBehaviorTracker) calculateSessionFrequency(summary *BehaviorSummary) float64 {
	if summary.TimeRange.End.Sub(summary.TimeRange.Start) == 0 {
		return 0
	}
	
	days := summary.TimeRange.End.Sub(summary.TimeRange.Start).Hours() / 24
	return float64(summary.TotalSessions) / days
}

func (ubt *UserBehaviorTracker) analyzeInteractionPatterns(ctx context.Context, learnerID uuid.UUID, timeRange BehaviorTimeRange) map[string]float64 {
	patterns := make(map[string]float64)
	
	// 简化实现，实际应该分析具体的交互模�?
	patterns["click_rate"] = 0.7
	patterns["scroll_depth"] = 0.8
	patterns["time_on_content"] = 0.6
	patterns["navigation_efficiency"] = 0.75
	
	return patterns
}

func (ubt *UserBehaviorTracker) identifyLearningPatterns(events []*BehaviorEvent) []LearningPattern {
	patterns := make([]LearningPattern, 0)
	
	// 分析学习时间模式
	timePattern := ubt.analyzeTimePatterns(events)
	if timePattern != nil {
		patterns = append(patterns, *timePattern)
	}
	
	// 分析内容偏好模式
	contentPattern := ubt.analyzeContentPatterns(events)
	if contentPattern != nil {
		patterns = append(patterns, *contentPattern)
	}
	
	// 分析学习节奏模式
	pacePattern := ubt.analyzePacePatterns(events)
	if pacePattern != nil {
		patterns = append(patterns, *pacePattern)
	}
	
	return patterns
}

func (ubt *UserBehaviorTracker) analyzeTimePatterns(events []*BehaviorEvent) *LearningPattern {
	if len(events) == 0 {
		return nil
	}
	
	// 分析学习时间分布
	hourCounts := make(map[int]int)
	for _, event := range events {
		hour := event.Timestamp.Hour()
		hourCounts[hour]++
	}
	
	// 找到最活跃的时间段
	maxCount := 0
	peakHour := 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}
	
	var timeDescription string
	if peakHour >= 6 && peakHour < 12 {
		timeDescription = "早晨学习�?
	} else if peakHour >= 12 && peakHour < 18 {
		timeDescription = "下午学习�?
	} else {
		timeDescription = "晚间学习�?
	}
	
	return &LearningPattern{
		Type:        "time_preference",
		Description: timeDescription,
		Confidence:  float64(maxCount) / float64(len(events)),
		Evidence:    []string{fmt.Sprintf("最活跃时间�? %d:00", peakHour)},
		Impact:      "影响学习效率和专注度",
	}
}

func (ubt *UserBehaviorTracker) analyzeContentPatterns(events []*BehaviorEvent) *LearningPattern {
	if len(events) == 0 {
		return nil
	}
	
	// 分析内容类型偏好
	contentTypes := make(map[string]int)
	for _, event := range events {
		if contentType, ok := event.Properties["content_type"].(string); ok {
			contentTypes[contentType]++
		}
	}
	
	if len(contentTypes) == 0 {
		return nil
	}
	
	// 找到最偏好的内容类�?
	maxCount := 0
	preferredType := ""
	for contentType, count := range contentTypes {
		if count > maxCount {
			maxCount = count
			preferredType = contentType
		}
	}
	
	return &LearningPattern{
		Type:        "content_preference",
		Description: fmt.Sprintf("偏好%s类型内容", preferredType),
		Confidence:  float64(maxCount) / float64(len(events)),
		Evidence:    []string{fmt.Sprintf("最常使用内容类�? %s", preferredType)},
		Impact:      "影响学习参与度和理解效果",
	}
}

func (ubt *UserBehaviorTracker) analyzePacePatterns(events []*BehaviorEvent) *LearningPattern {
	if len(events) < 2 {
		return nil
	}
	
	// 计算平均会话时长
	var totalDuration time.Duration
	sessionCount := 0
	
	for _, event := range events {
		if event.Duration > 0 {
			totalDuration += event.Duration
			sessionCount++
		}
	}
	
	if sessionCount == 0 {
		return nil
	}
	
	avgDuration := totalDuration / time.Duration(sessionCount)
	
	var paceDescription string
	if avgDuration < 15*time.Minute {
		paceDescription = "快节奏学习�?
	} else if avgDuration < 45*time.Minute {
		paceDescription = "中等节奏学习�?
	} else {
		paceDescription = "深度学习�?
	}
	
	return &LearningPattern{
		Type:        "learning_pace",
		Description: paceDescription,
		Confidence:  0.8,
		Evidence:    []string{fmt.Sprintf("平均会话时长: %v", avgDuration)},
		Impact:      "影响学习深度和知识保�?,
	}
}

func (ubt *UserBehaviorTracker) analyzePerformanceTrends(events []*BehaviorEvent) []BehaviorPerformanceTrend {
	trends := make([]BehaviorPerformanceTrend, 0)
	
	// 分析分数趋势
	scoreEvents := make([]*BehaviorEvent, 0)
	for _, event := range events {
		if event.Performance != nil && event.Performance.Score > 0 {
			scoreEvents = append(scoreEvents, event)
		}
	}
	
	if len(scoreEvents) >= 2 {
		// 按时间排�?
		sort.Slice(scoreEvents, func(i, j int) bool {
			return scoreEvents[i].Timestamp.Before(scoreEvents[j].Timestamp)
		})
		
		// 计算趋势
		firstHalf := scoreEvents[:len(scoreEvents)/2]
		secondHalf := scoreEvents[len(scoreEvents)/2:]
		
		firstAvg := ubt.calculateAverageScore(firstHalf)
		secondAvg := ubt.calculateAverageScore(secondHalf)
		
		change := secondAvg - firstAvg
		var trendType string
		if change > 0.05 {
			trendType = "increasing"
		} else if change < -0.05 {
			trendType = "decreasing"
		} else {
			trendType = "stable"
		}
		
		trends = append(trends, BehaviorPerformanceTrend{
			Metric:     "score",
			Trend:      trendType,
			Value:      secondAvg,
			Change:     change,
			Confidence: 0.8,
		})
	}
	
	return trends
}

func (ubt *UserBehaviorTracker) calculateAverageScore(events []*BehaviorEvent) float64 {
	if len(events) == 0 {
		return 0
	}
	
	var total float64
	for _, event := range events {
		if event.Performance != nil {
			total += event.Performance.Score
		}
	}
	
	return total / float64(len(events))
}

func (ubt *UserBehaviorTracker) analyzeEngagementInsights(events []*BehaviorEvent) []EngagementInsight {
	insights := make([]EngagementInsight, 0)
	
	// 分析会话长度
	if len(events) > 0 {
		var totalDuration time.Duration
		for _, event := range events {
			totalDuration += event.Duration
		}
		avgDuration := totalDuration / time.Duration(len(events))
		
		if avgDuration < 10*time.Minute {
			insights = append(insights, EngagementInsight{
				Type:        "session_length",
				Title:       "会话时间较短",
				Description: "平均学习会话时间较短，可能影响深度学�?,
				Severity:    "medium",
				Impact:      0.6,
				Suggestions: []string{"尝试延长学习会话", "设置学习目标", "减少干扰因素"},
			})
		}
	}
	
	return insights
}

// GetBehaviorSummary 获取行为摘要
func (ubt *UserBehaviorTracker) GetBehaviorSummary(ctx context.Context, userID string, startTime, endTime time.Time) (*BehaviorSummary, error) {
	// 将string类型的userID转换为uuid.UUID
	learnerID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// 创建时间范围
	timeRange := BehaviorTimeRange{
		Start: startTime,
		End:   endTime,
	}

	// 调用仓库方法获取行为摘要
	summary, err := ubt.behaviorRepo.GetBehaviorSummary(ctx, learnerID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}

	return summary, nil
}

func (ubt *UserBehaviorTracker) generateRecommendedActions(events []*BehaviorEvent) []BehaviorRecommendedAction {
	actions := make([]BehaviorRecommendedAction, 0)
	
	// 基于行为数据生成推荐行动
	if len(events) > 0 {
		// 分析学习频率
		timeSpan := events[len(events)-1].Timestamp.Sub(events[0].Timestamp)
		frequency := float64(len(events)) / timeSpan.Hours() * 24 // 每天的事件数
		
		if frequency < 1 {
			actions = append(actions, BehaviorRecommendedAction{
				ID:             uuid.New(),
				Type:           "frequency",
				Title:          "增加学习频率",
				Description:    "建议增加学习频率以提高学习效�?,
				Priority:       2,
				ExpectedImpact: 0.7,
				Timeline:       "1-2�?,
				Implementation: []string{"设置每日学习提醒", "制定学习计划", "设定小目�?},
				SuccessMetrics: []string{"学习天数增加", "完成率提�?},
			})
		}
	}
	
	return actions
}

func (ubt *UserBehaviorTracker) generatePredictiveInsights(events []*BehaviorEvent) []PredictiveInsight {
	insights := make([]PredictiveInsight, 0)
	
	// 基于历史数据预测学习趋势
	if len(events) >= 10 {
		// 简化的预测逻辑
		recentEvents := events[len(events)-5:]
		var recentEngagement float64
		
		for _, event := range recentEvents {
			if event.Duration > 30*time.Minute {
				recentEngagement += 1
			}
		}
		recentEngagement /= float64(len(recentEvents))
		
		if recentEngagement > 0.6 {
			insights = append(insights, PredictiveInsight{
				Type:           "engagement_trend",
				Prediction:     "学习参与度将继续保持较高水平",
				Confidence:     0.75,
				TimeHorizon:    7 * 24 * time.Hour, // 7�?
				Factors:        []string{"持续的学习习�?, "高参与度会话"},
				Recommendations: []string{"保持当前学习节奏", "适当增加挑战性内�?},
			})
		} else {
			insights = append(insights, PredictiveInsight{
				Type:           "engagement_risk",
				Prediction:     "学习参与度可能下�?,
				Confidence:     0.65,
				TimeHorizon:    3 * 24 * time.Hour, // 3�?
				Factors:        []string{"最近会话时间较�?, "参与度下降趋�?},
				Recommendations: []string{"调整学习内容", "增加互动元素", "设置激励机�?},
			})
		}
	}
	
	return insights
}
