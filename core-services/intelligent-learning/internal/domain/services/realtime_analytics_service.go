package services

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// RealtimeAnalyticsService 实时学习分析服务
type RealtimeAnalyticsService struct {
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	
	// 实时数据流
	eventStream        chan *LearningEvent
	subscribers        map[string]chan *AnalyticsUpdate
	subscribersMutex   sync.RWMutex
	
	// 实时分析器
	analyzers          map[string]*RealtimeAnalyzer
	analyzersMutex     sync.RWMutex
	
	// 数据缓存
	realtimeCache      map[uuid.UUID]*RealtimeLearnerData
	cacheMutex         sync.RWMutex
	
	// 配置
	config             *RealtimeConfig
	
	// 控制
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
}

// LearningEvent 学习事件
type LearningEvent struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"` // "progress", "interaction", "completion", "error", "engagement"
	LearnerID   uuid.UUID              `json:"learner_id"`
	ContentID   uuid.UUID              `json:"content_id,omitempty"`
	SessionID   uuid.UUID              `json:"session_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsUpdate 分析更新
type AnalyticsUpdate struct {
	Type        string                 `json:"type"`
	LearnerID   uuid.UUID              `json:"learner_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        interface{}            `json:"data"`
	Confidence  float64                `json:"confidence"`
	Insights    []string               `json:"insights,omitempty"`
	Alerts      []Alert                `json:"alerts,omitempty"`
}

// Alert 警报
type Alert struct {
	Level       string    `json:"level"` // "info", "warning", "critical"
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	ActionItems []string  `json:"action_items,omitempty"`
}

// RealtimeAnalyzer 实时分析器
type RealtimeAnalyzer struct {
	ID              string
	Type            string // "performance", "engagement", "progress", "behavior", "prediction"
	LearnerID       uuid.UUID
	Config          map[string]interface{}
	LastUpdate      time.Time
	IsActive        bool
	
	// 分析状态
	EventCount      int64
	ProcessingTime  time.Duration
	Accuracy        float64
	
	// 数据窗口
	WindowSize      time.Duration
	EventBuffer     []*LearningEvent
	bufferMutex     sync.Mutex
}

// RealtimeLearnerData 实时学习者数据
type RealtimeLearnerData struct {
	LearnerID           uuid.UUID                  `json:"learner_id"`
	LastActivity        time.Time                  `json:"last_activity"`
	CurrentSession      *LearningSession           `json:"current_session"`
	RealtimeMetrics     *RealtimeMetrics           `json:"realtime_metrics"`
	BehaviorPatterns    *BehaviorPatterns          `json:"behavior_patterns"`
	EngagementState     *EngagementState           `json:"engagement_state"`
	PerformanceState    *PerformanceState          `json:"performance_state"`
	PredictiveInsights  *PredictiveInsights        `json:"predictive_insights"`
	Alerts              []Alert                    `json:"alerts"`
}

// LearningSession 学习会话
type LearningSession struct {
	ID              uuid.UUID              `json:"id"`
	StartTime       time.Time              `json:"start_time"`
	LastActivity    time.Time              `json:"last_activity"`
	Duration        time.Duration          `json:"duration"`
	ContentItems    []uuid.UUID            `json:"content_items"`
	InteractionCount int                   `json:"interaction_count"`
	ProgressMade    float64                `json:"progress_made"`
	EngagementScore float64                `json:"engagement_score"`
	FocusLevel      float64                `json:"focus_level"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// RealtimeMetrics 实时指标
type RealtimeMetrics struct {
	LearningVelocity    float64   `json:"learning_velocity"`
	EngagementTrend     string    `json:"engagement_trend"` // "increasing", "stable", "decreasing"
	FocusScore          float64   `json:"focus_score"`
	EfficiencyScore     float64   `json:"efficiency_score"`
	MotivationLevel     float64   `json:"motivation_level"`
	CognitiveLoad       float64   `json:"cognitive_load"`
	LastUpdated         time.Time `json:"last_updated"`
}

// BehaviorPatterns 行为模式
type BehaviorPatterns struct {
	InteractionFrequency map[string]float64 `json:"interaction_frequency"`
	NavigationPatterns   []string           `json:"navigation_patterns"`
	TimeDistribution     map[string]float64 `json:"time_distribution"`
	PreferredContentTypes []string          `json:"preferred_content_types"`
	LearningRhythm       string             `json:"learning_rhythm"` // "steady", "burst", "irregular"
	AttentionSpan        time.Duration      `json:"attention_span"`
}

// EngagementState 参与状态
type EngagementState struct {
	Level               string    `json:"level"` // "high", "medium", "low", "disengaged"
	Score               float64   `json:"score"`
	Trend               string    `json:"trend"`
	LastInteraction     time.Time `json:"last_interaction"`
	InteractionQuality  float64   `json:"interaction_quality"`
	AttentionIndicators map[string]float64 `json:"attention_indicators"`
	RiskFactors         []string  `json:"risk_factors"`
}

// PerformanceState 表现状态
type PerformanceState struct {
	CurrentLevel        string             `json:"current_level"` // "excellent", "good", "average", "below_average", "poor"
	Score               float64            `json:"score"`
	Trend               string             `json:"trend"`
	StrengthAreas       []string           `json:"strength_areas"`
	ImprovementAreas    []string           `json:"improvement_areas"`
	RecentAchievements  []string           `json:"recent_achievements"`
	PerformanceMetrics  map[string]float64 `json:"performance_metrics"`
}

// PredictiveInsights 预测洞察
type PredictiveInsights struct {
	CompletionProbability   float64            `json:"completion_probability"`
	EstimatedCompletionTime time.Time          `json:"estimated_completion_time"`
	RiskOfDropout           float64            `json:"risk_of_dropout"`
	RecommendedActions      []string           `json:"recommended_actions"`
	NextBestContent         []uuid.UUID        `json:"next_best_content"`
	OptimalStudyTime        time.Duration      `json:"optimal_study_time"`
	PredictedChallenges     []string           `json:"predicted_challenges"`
	SuccessFactors          map[string]float64 `json:"success_factors"`
}

// RealtimeConfig 实时配置
type RealtimeConfig struct {
	EventBufferSize     int           `json:"event_buffer_size"`
	AnalysisInterval    time.Duration `json:"analysis_interval"`
	CacheExpiration     time.Duration `json:"cache_expiration"`
	MaxSubscribers      int           `json:"max_subscribers"`
	EnablePrediction    bool          `json:"enable_prediction"`
	AlertThresholds     map[string]float64 `json:"alert_thresholds"`
	AnalyzerConfigs     map[string]interface{} `json:"analyzer_configs"`
}

// NewRealtimeAnalyticsService 创建新的实时分析服务
func NewRealtimeAnalyticsService(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
) *RealtimeAnalyticsService {
	ctx, cancel := context.WithCancel(context.Background())
	
	config := &RealtimeConfig{
		EventBufferSize:  1000,
		AnalysisInterval: 5 * time.Second,
		CacheExpiration:  30 * time.Minute,
		MaxSubscribers:   100,
		EnablePrediction: true,
		AlertThresholds: map[string]float64{
			"low_engagement":    0.3,
			"high_cognitive_load": 0.8,
			"dropout_risk":      0.7,
			"performance_drop":  0.4,
		},
	}
	
	service := &RealtimeAnalyticsService{
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
		eventStream:        make(chan *LearningEvent, config.EventBufferSize),
		subscribers:        make(map[string]chan *AnalyticsUpdate),
		analyzers:          make(map[string]*RealtimeAnalyzer),
		realtimeCache:      make(map[uuid.UUID]*RealtimeLearnerData),
		config:             config,
		ctx:                ctx,
		cancel:             cancel,
	}
	
	// 启动后台处理
	service.wg.Add(1)
	go service.processEventStream()
	
	service.wg.Add(1)
	go service.runPeriodicAnalysis()
	
	service.wg.Add(1)
	go service.cleanupExpiredData()
	
	return service
}

// ProcessEvent 处理学习事件
func (s *RealtimeAnalyticsService) ProcessEvent(event *LearningEvent) error {
	select {
	case s.eventStream <- event:
		return nil
	default:
		return fmt.Errorf("event stream buffer full")
	}
}

// Subscribe 订阅分析更新
func (s *RealtimeAnalyticsService) Subscribe(subscriberID string) (<-chan *AnalyticsUpdate, error) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	
	if len(s.subscribers) >= s.config.MaxSubscribers {
		return nil, fmt.Errorf("maximum subscribers reached")
	}
	
	ch := make(chan *AnalyticsUpdate, 100)
	s.subscribers[subscriberID] = ch
	
	return ch, nil
}

// Unsubscribe 取消订阅
func (s *RealtimeAnalyticsService) Unsubscribe(subscriberID string) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	
	if ch, exists := s.subscribers[subscriberID]; exists {
		close(ch)
		delete(s.subscribers, subscriberID)
	}
}

// GetRealtimeData 获取实时数据
func (s *RealtimeAnalyticsService) GetRealtimeData(learnerID uuid.UUID) (*RealtimeLearnerData, error) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	if data, exists := s.realtimeCache[learnerID]; exists {
		return data, nil
	}
	
	return nil, fmt.Errorf("no realtime data found for learner %s", learnerID)
}

// CreateAnalyzer 创建分析器
func (s *RealtimeAnalyticsService) CreateAnalyzer(analyzerType string, learnerID uuid.UUID, config map[string]interface{}) (*RealtimeAnalyzer, error) {
	s.analyzersMutex.Lock()
	defer s.analyzersMutex.Unlock()
	
	analyzerID := fmt.Sprintf("%s_%s_%s", analyzerType, learnerID.String(), uuid.New().String()[:8])
	
	analyzer := &RealtimeAnalyzer{
		ID:           analyzerID,
		Type:         analyzerType,
		LearnerID:    learnerID,
		Config:       config,
		LastUpdate:   time.Now(),
		IsActive:     true,
		WindowSize:   5 * time.Minute,
		EventBuffer:  make([]*LearningEvent, 0),
	}
	
	s.analyzers[analyzerID] = analyzer
	
	return analyzer, nil
}

// processEventStream 处理事件流
func (s *RealtimeAnalyticsService) processEventStream() {
	defer s.wg.Done()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case event := <-s.eventStream:
			s.handleEvent(event)
		}
	}
}

// handleEvent 处理单个事件
func (s *RealtimeAnalyticsService) handleEvent(event *LearningEvent) {
	// 更新实时数据缓存
	s.updateRealtimeCache(event)
	
	// 分发给相关分析器
	s.distributeToAnalyzers(event)
	
	// 生成实时分析
	update := s.generateRealtimeAnalysis(event)
	if update != nil {
		s.broadcastUpdate(update)
	}
}

// updateRealtimeCache 更新实时数据缓存
func (s *RealtimeAnalyticsService) updateRealtimeCache(event *LearningEvent) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	data, exists := s.realtimeCache[event.LearnerID]
	if !exists {
		data = &RealtimeLearnerData{
			LearnerID:        event.LearnerID,
			LastActivity:     event.Timestamp,
			RealtimeMetrics:  &RealtimeMetrics{},
			BehaviorPatterns: &BehaviorPatterns{
				InteractionFrequency: make(map[string]float64),
				TimeDistribution:     make(map[string]float64),
			},
			EngagementState:  &EngagementState{},
			PerformanceState: &PerformanceState{},
			PredictiveInsights: &PredictiveInsights{},
			Alerts:           make([]Alert, 0),
		}
		s.realtimeCache[event.LearnerID] = data
	}
	
	// 更新基本信息
	data.LastActivity = event.Timestamp
	
	// 根据事件类型更新相应数据
	switch event.Type {
	case "progress":
		s.updateProgressMetrics(data, event)
	case "interaction":
		s.updateInteractionMetrics(data, event)
	case "engagement":
		s.updateEngagementMetrics(data, event)
	case "performance":
		s.updatePerformanceMetrics(data, event)
	}
	
	// 更新实时指标
	s.calculateRealtimeMetrics(data)
}

// Stop 停止服务
func (s *RealtimeAnalyticsService) Stop() {
	s.cancel()
	s.wg.Wait()
	
	// 关闭所有订阅者通道
	s.subscribersMutex.Lock()
	for _, ch := range s.subscribers {
		close(ch)
	}
	s.subscribersMutex.Unlock()
}

// 其他辅助方法将在后续实现...

// distributeToAnalyzers 分发事件给分析器
func (s *RealtimeAnalyticsService) distributeToAnalyzers(event *LearningEvent) {
	s.analyzersMutex.RLock()
	defer s.analyzersMutex.RUnlock()
	
	for _, analyzer := range s.analyzers {
		if analyzer.IsActive && analyzer.LearnerID == event.LearnerID {
			analyzer.bufferMutex.Lock()
			analyzer.EventBuffer = append(analyzer.EventBuffer, event)
			analyzer.EventCount++
			
			// 保持窗口大小
			cutoff := time.Now().Add(-analyzer.WindowSize)
			var filtered []*LearningEvent
			for _, e := range analyzer.EventBuffer {
				if e.Timestamp.After(cutoff) {
					filtered = append(filtered, e)
				}
			}
			analyzer.EventBuffer = filtered
			analyzer.bufferMutex.Unlock()
		}
	}
}

// generateRealtimeAnalysis 生成实时分析
func (s *RealtimeAnalyticsService) generateRealtimeAnalysis(event *LearningEvent) *AnalyticsUpdate {
	data, exists := s.realtimeCache[event.LearnerID]
	if !exists {
		return nil
	}
	
	var insights []string
	var alerts []Alert
	
	// 生成洞察
	insights = append(insights, s.generateEngagementInsights(data)...)
	insights = append(insights, s.generatePerformanceInsights(data)...)
	insights = append(insights, s.generateBehaviorInsights(data)...)
	
	// 检查警报
	alerts = append(alerts, s.checkEngagementAlerts(data)...)
	alerts = append(alerts, s.checkPerformanceAlerts(data)...)
	alerts = append(alerts, s.checkBehaviorAlerts(data)...)
	
	return &AnalyticsUpdate{
		Type:       "realtime_analysis",
		LearnerID:  event.LearnerID,
		Timestamp:  time.Now(),
		Data:       data,
		Confidence: s.calculateConfidence(data),
		Insights:   insights,
		Alerts:     alerts,
	}
}

// broadcastUpdate 广播更新
func (s *RealtimeAnalyticsService) broadcastUpdate(update *AnalyticsUpdate) {
	s.subscribersMutex.RLock()
	defer s.subscribersMutex.RUnlock()
	
	for _, ch := range s.subscribers {
		select {
		case ch <- update:
		default:
			// 通道满了，跳过这个订阅者
		}
	}
}

// updateProgressMetrics 更新进度指标
func (s *RealtimeAnalyticsService) updateProgressMetrics(data *RealtimeLearnerData, event *LearningEvent) {
	if progress, ok := event.Data["progress"].(float64); ok {
		if data.CurrentSession != nil {
			data.CurrentSession.ProgressMade += progress
		}
		
		// 计算学习速度
		if timeSpent, ok := event.Data["time_spent"].(float64); ok && timeSpent > 0 {
			velocity := progress / (timeSpent / 3600) // 每小时进度
			data.RealtimeMetrics.LearningVelocity = velocity
		}
	}
}

// updateInteractionMetrics 更新交互指标
func (s *RealtimeAnalyticsService) updateInteractionMetrics(data *RealtimeLearnerData, event *LearningEvent) {
	if data.CurrentSession != nil {
		data.CurrentSession.InteractionCount++
		data.CurrentSession.LastActivity = event.Timestamp
	}
	
	// 更新交互频率
	if interactionType, ok := event.Data["interaction_type"].(string); ok {
		data.BehaviorPatterns.InteractionFrequency[interactionType]++
	}
	
	// 计算专注度
	if duration, ok := event.Data["duration"].(float64); ok {
		data.RealtimeMetrics.FocusScore = s.calculateFocusScore(duration, data.BehaviorPatterns)
	}
}

// updateEngagementMetrics 更新参与度指标
func (s *RealtimeAnalyticsService) updateEngagementMetrics(data *RealtimeLearnerData, event *LearningEvent) {
	if engagementScore, ok := event.Data["engagement_score"].(float64); ok {
		data.EngagementState.Score = engagementScore
		data.EngagementState.LastInteraction = event.Timestamp
		
		// 更新参与度等级
		if engagementScore >= 0.8 {
			data.EngagementState.Level = "high"
		} else if engagementScore >= 0.6 {
			data.EngagementState.Level = "medium"
		} else if engagementScore >= 0.3 {
			data.EngagementState.Level = "low"
		} else {
			data.EngagementState.Level = "disengaged"
		}
		
		if data.CurrentSession != nil {
			data.CurrentSession.EngagementScore = engagementScore
		}
	}
}

// updatePerformanceMetrics 更新表现指标
func (s *RealtimeAnalyticsService) updatePerformanceMetrics(data *RealtimeLearnerData, event *LearningEvent) {
	if score, ok := event.Data["score"].(float64); ok {
		data.PerformanceState.Score = score
		
		// 更新表现等级
		if score >= 0.9 {
			data.PerformanceState.CurrentLevel = "excellent"
		} else if score >= 0.8 {
			data.PerformanceState.CurrentLevel = "good"
		} else if score >= 0.6 {
			data.PerformanceState.CurrentLevel = "average"
		} else if score >= 0.4 {
			data.PerformanceState.CurrentLevel = "below_average"
		} else {
			data.PerformanceState.CurrentLevel = "poor"
		}
	}
}

// calculateRealtimeMetrics 计算实时指标
func (s *RealtimeAnalyticsService) calculateRealtimeMetrics(data *RealtimeLearnerData) {
	metrics := data.RealtimeMetrics
	
	// 计算效率分数
	if data.CurrentSession != nil && data.CurrentSession.Duration > 0 {
		efficiency := data.CurrentSession.ProgressMade / data.CurrentSession.Duration.Hours()
		metrics.EfficiencyScore = math.Min(efficiency, 1.0)
	}
	
	// 计算认知负荷
	metrics.CognitiveLoad = s.calculateCognitiveLoad(data)
	
	// 计算动机水平
	metrics.MotivationLevel = s.calculateMotivationLevel(data)
	
	// 更新趋势
	metrics.EngagementTrend = s.calculateEngagementTrend(data)
	
	metrics.LastUpdated = time.Now()
}

// generateEngagementInsights 生成参与度洞察
func (s *RealtimeAnalyticsService) generateEngagementInsights(data *RealtimeLearnerData) []string {
	var insights []string
	
	engagement := data.EngagementState
	
	if engagement.Score > 0.8 {
		insights = append(insights, "学习者表现出高度参与，建议继续当前学习策略")
	} else if engagement.Score < 0.3 {
		insights = append(insights, "学习者参与度较低，建议调整学习内容或方式")
		
		// 分析可能的原因
		if data.RealtimeMetrics.CognitiveLoad > 0.8 {
			insights = append(insights, "认知负荷过高可能是参与度低的原因")
		}
	}
	
	// 分析交互模式
	if len(data.BehaviorPatterns.InteractionFrequency) > 0 {
		maxInteraction := ""
		maxCount := 0.0
		for interaction, count := range data.BehaviorPatterns.InteractionFrequency {
			if count > maxCount {
				maxCount = count
				maxInteraction = interaction
			}
		}
		insights = append(insights, fmt.Sprintf("最常用的交互方式是: %s", maxInteraction))
	}
	
	return insights
}

// generatePerformanceInsights 生成表现洞察
func (s *RealtimeAnalyticsService) generatePerformanceInsights(data *RealtimeLearnerData) []string {
	var insights []string
	
	performance := data.PerformanceState
	
	switch performance.CurrentLevel {
	case "excellent":
		insights = append(insights, "表现优秀，可以尝试更具挑战性的内容")
	case "poor", "below_average":
		insights = append(insights, "表现需要改进，建议回顾基础知识或寻求帮助")
	}
	
	// 分析学习效率
	if data.RealtimeMetrics.EfficiencyScore > 0.8 {
		insights = append(insights, "学习效率很高，保持当前节奏")
	} else if data.RealtimeMetrics.EfficiencyScore < 0.3 {
		insights = append(insights, "学习效率较低，建议调整学习方法")
	}
	
	return insights
}

// generateBehaviorInsights 生成行为洞察
func (s *RealtimeAnalyticsService) generateBehaviorInsights(data *RealtimeLearnerData) []string {
	var insights []string
	
	// 分析学习节奏
	switch data.BehaviorPatterns.LearningRhythm {
	case "burst":
		insights = append(insights, "倾向于集中学习，建议适当休息避免疲劳")
	case "irregular":
		insights = append(insights, "学习节奏不规律，建议制定固定的学习计划")
	case "steady":
		insights = append(insights, "学习节奏稳定，这是一个很好的学习习惯")
	}
	
	// 分析注意力持续时间
	if data.BehaviorPatterns.AttentionSpan < 15*time.Minute {
		insights = append(insights, "注意力持续时间较短，建议采用短时间高频率的学习方式")
	} else if data.BehaviorPatterns.AttentionSpan > 2*time.Hour {
		insights = append(insights, "注意力持续时间很长，但要注意适当休息")
	}
	
	return insights
}

// checkEngagementAlerts 检查参与度警报
func (s *RealtimeAnalyticsService) checkEngagementAlerts(data *RealtimeLearnerData) []Alert {
	var alerts []Alert
	
	if data.EngagementState.Score < s.config.AlertThresholds["low_engagement"] {
		alerts = append(alerts, Alert{
			Level:     "warning",
			Message:   "学习者参与度过低",
			Timestamp: time.Now(),
			ActionItems: []string{
				"调整学习内容难度",
				"增加互动元素",
				"提供个性化推荐",
			},
		})
	}
	
	// 检查长时间无活动
	if time.Since(data.EngagementState.LastInteraction) > 10*time.Minute {
		alerts = append(alerts, Alert{
			Level:     "info",
			Message:   "学习者长时间无活动",
			Timestamp: time.Now(),
			ActionItems: []string{
				"发送提醒通知",
				"提供学习建议",
			},
		})
	}
	
	return alerts
}

// checkPerformanceAlerts 检查表现警报
func (s *RealtimeAnalyticsService) checkPerformanceAlerts(data *RealtimeLearnerData) []Alert {
	var alerts []Alert
	
	if data.PerformanceState.Score < s.config.AlertThresholds["performance_drop"] {
		alerts = append(alerts, Alert{
			Level:     "warning",
			Message:   "学习表现下降",
			Timestamp: time.Now(),
			ActionItems: []string{
				"提供额外支持",
				"调整学习策略",
				"安排复习内容",
			},
		})
	}
	
	return alerts
}

// checkBehaviorAlerts 检查行为警报
func (s *RealtimeAnalyticsService) checkBehaviorAlerts(data *RealtimeLearnerData) []Alert {
	var alerts []Alert
	
	if data.RealtimeMetrics.CognitiveLoad > s.config.AlertThresholds["high_cognitive_load"] {
		alerts = append(alerts, Alert{
			Level:     "warning",
			Message:   "认知负荷过高",
			Timestamp: time.Now(),
			ActionItems: []string{
				"建议休息",
				"降低内容难度",
				"分解学习任务",
			},
		})
	}
	
	if data.PredictiveInsights.RiskOfDropout > s.config.AlertThresholds["dropout_risk"] {
		alerts = append(alerts, Alert{
			Level:     "critical",
			Message:   "高辍学风险",
			Timestamp: time.Now(),
			ActionItems: []string{
				"立即干预",
				"提供个性化支持",
				"联系学习顾问",
			},
		})
	}
	
	return alerts
}

// 辅助计算方法
func (s *RealtimeAnalyticsService) calculateFocusScore(duration float64, patterns *BehaviorPatterns) float64 {
	// 基于交互持续时间和模式计算专注度
	baseScore := math.Min(duration/3600, 1.0) // 标准化到1小时
	
	// 考虑交互频率
	var avgFreq float64
	for _, freq := range patterns.InteractionFrequency {
		avgFreq += freq
	}
	if len(patterns.InteractionFrequency) > 0 {
		avgFreq /= float64(len(patterns.InteractionFrequency))
	}
	
	// 适中的交互频率表示良好的专注度
	freqScore := 1.0 - math.Abs(avgFreq-0.5)*2
	
	return (baseScore + freqScore) / 2
}

func (s *RealtimeAnalyticsService) calculateCognitiveLoad(data *RealtimeLearnerData) float64 {
	// 基于多个因素计算认知负荷
	load := 0.0
	
	// 基于交互频率
	if len(data.BehaviorPatterns.InteractionFrequency) > 0 {
		var totalInteractions float64
		for _, freq := range data.BehaviorPatterns.InteractionFrequency {
			totalInteractions += freq
		}
		// 交互过于频繁可能表示认知负荷高
		if totalInteractions > 100 {
			load += 0.3
		}
	}
	
	// 基于学习效率
	if data.RealtimeMetrics.EfficiencyScore < 0.3 {
		load += 0.4
	}
	
	// 基于参与度
	if data.EngagementState.Score < 0.4 {
		load += 0.3
	}
	
	return math.Min(load, 1.0)
}

func (s *RealtimeAnalyticsService) calculateMotivationLevel(data *RealtimeLearnerData) float64 {
	// 基于多个指标计算动机水平
	motivation := 0.0
	
	// 参与度贡献
	motivation += data.EngagementState.Score * 0.4
	
	// 表现贡献
	motivation += data.PerformanceState.Score * 0.3
	
	// 学习效率贡献
	motivation += data.RealtimeMetrics.EfficiencyScore * 0.3
	
	return math.Min(motivation, 1.0)
}

func (s *RealtimeAnalyticsService) calculateEngagementTrend(data *RealtimeLearnerData) string {
	// 简化的趋势计算，实际应该基于历史数据
	current := data.EngagementState.Score
	
	if current > 0.7 {
		return "increasing"
	} else if current > 0.4 {
		return "stable"
	} else {
		return "decreasing"
	}
}

func (s *RealtimeAnalyticsService) calculateConfidence(data *RealtimeLearnerData) float64 {
	// 基于数据完整性和时效性计算置信度
	confidence := 0.5 // 基础置信度
	
	// 数据新鲜度
	if time.Since(data.LastActivity) < 5*time.Minute {
		confidence += 0.3
	}
	
	// 数据完整性
	if data.CurrentSession != nil && data.CurrentSession.InteractionCount > 5 {
		confidence += 0.2
	}
	
	return math.Min(confidence, 1.0)
}

// runPeriodicAnalysis 运行周期性分析
func (s *RealtimeAnalyticsService) runPeriodicAnalysis() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(s.config.AnalysisInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performPeriodicAnalysis()
		}
	}
}

// performPeriodicAnalysis 执行周期性分析
func (s *RealtimeAnalyticsService) performPeriodicAnalysis() {
	s.cacheMutex.RLock()
	learners := make([]uuid.UUID, 0, len(s.realtimeCache))
	for learnerID := range s.realtimeCache {
		learners = append(learners, learnerID)
	}
	s.cacheMutex.RUnlock()
	
	for _, learnerID := range learners {
		s.analyzelearnerPredictions(learnerID)
	}
}

// analyzelearnerPredictions 分析学习者预测
func (s *RealtimeAnalyticsService) analyzelearnerPredictions(learnerID uuid.UUID) {
	s.cacheMutex.Lock()
	data, exists := s.realtimeCache[learnerID]
	if !exists {
		s.cacheMutex.Unlock()
		return
	}
	
	// 更新预测洞察
	data.PredictiveInsights.CompletionProbability = s.predictCompletionProbability(data)
	data.PredictiveInsights.RiskOfDropout = s.predictDropoutRisk(data)
	data.PredictiveInsights.EstimatedCompletionTime = s.estimateCompletionTime(data)
	data.PredictiveInsights.RecommendedActions = s.generateRecommendedActions(data)
	
	s.cacheMutex.Unlock()
	
	// 广播预测更新
	update := &AnalyticsUpdate{
		Type:       "prediction_update",
		LearnerID:  learnerID,
		Timestamp:  time.Now(),
		Data:       data.PredictiveInsights,
		Confidence: s.calculateConfidence(data),
	}
	
	s.broadcastUpdate(update)
}

// 预测方法
func (s *RealtimeAnalyticsService) predictCompletionProbability(data *RealtimeLearnerData) float64 {
	// 简化的完成概率预测
	score := 0.0
	
	// 基于当前表现
	score += data.PerformanceState.Score * 0.4
	
	// 基于参与度
	score += data.EngagementState.Score * 0.3
	
	// 基于学习效率
	score += data.RealtimeMetrics.EfficiencyScore * 0.3
	
	return math.Min(score, 1.0)
}

func (s *RealtimeAnalyticsService) predictDropoutRisk(data *RealtimeLearnerData) float64 {
	// 简化的辍学风险预测
	risk := 0.0
	
	// 低参与度增加风险
	if data.EngagementState.Score < 0.3 {
		risk += 0.4
	}
	
	// 低表现增加风险
	if data.PerformanceState.Score < 0.4 {
		risk += 0.3
	}
	
	// 高认知负荷增加风险
	if data.RealtimeMetrics.CognitiveLoad > 0.8 {
		risk += 0.3
	}
	
	return math.Min(risk, 1.0)
}

func (s *RealtimeAnalyticsService) estimateCompletionTime(data *RealtimeLearnerData) time.Time {
	// 简化的完成时间估算
	baseTime := time.Now().Add(30 * 24 * time.Hour) // 默认30天
	
	// 根据学习速度调整
	if data.RealtimeMetrics.LearningVelocity > 0.5 {
		baseTime = baseTime.Add(-10 * 24 * time.Hour) // 提前10天
	} else if data.RealtimeMetrics.LearningVelocity < 0.2 {
		baseTime = baseTime.Add(20 * 24 * time.Hour) // 延后20天
	}
	
	return baseTime
}

func (s *RealtimeAnalyticsService) generateRecommendedActions(data *RealtimeLearnerData) []string {
	var actions []string
	
	// 基于参与度
	if data.EngagementState.Score < 0.4 {
		actions = append(actions, "增加互动内容", "调整学习节奏")
	}
	
	// 基于表现
	if data.PerformanceState.Score < 0.5 {
		actions = append(actions, "复习基础知识", "寻求额外帮助")
	}
	
	// 基于认知负荷
	if data.RealtimeMetrics.CognitiveLoad > 0.7 {
		actions = append(actions, "适当休息", "分解学习任务")
	}
	
	return actions
}

// cleanupExpiredData 清理过期数据
func (s *RealtimeAnalyticsService) cleanupExpiredData() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performCleanup()
		}
	}
}

// performCleanup 执行清理
func (s *RealtimeAnalyticsService) performCleanup() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	cutoff := time.Now().Add(-s.config.CacheExpiration)
	
	for learnerID, data := range s.realtimeCache {
		if data.LastActivity.Before(cutoff) {
			delete(s.realtimeCache, learnerID)
		}
	}
}