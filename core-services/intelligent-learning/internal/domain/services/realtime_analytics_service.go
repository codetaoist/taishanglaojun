package services

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// RealtimeAnalyticsService 
type RealtimeAnalyticsService struct {
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	
	// ?
	eventStream        chan *LearningEvent
	subscribers        map[string]chan *AnalyticsUpdate
	subscribersMutex   sync.RWMutex
	
	// ?
	analyzers          map[string]*RealtimeAnalyzer
	analyzersMutex     sync.RWMutex
	
	// 
	realtimeCache      map[uuid.UUID]*RealtimeLearnerData
	cacheMutex         sync.RWMutex
	
	// 
	config             *RealtimeConfig
	
	// 
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
}

// LearningEvent 
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

// AnalyticsUpdate 
type AnalyticsUpdate struct {
	Type        string                 `json:"type"`
	LearnerID   uuid.UUID              `json:"learner_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        interface{}            `json:"data"`
	Confidence  float64                `json:"confidence"`
	Insights    []string               `json:"insights,omitempty"`
	Alerts      []Alert                `json:"alerts,omitempty"`
}

// Alert 
type Alert struct {
	Level       string    `json:"level"` // "info", "warning", "critical"
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	ActionItems []string  `json:"action_items,omitempty"`
}

// RealtimeAnalyzer ?
type RealtimeAnalyzer struct {
	ID              string
	Type            string // "performance", "engagement", "progress", "behavior", "prediction"
	LearnerID       uuid.UUID
	Config          map[string]interface{}
	LastUpdate      time.Time
	IsActive        bool
	
	// ?
	EventCount      int64
	ProcessingTime  time.Duration
	Accuracy        float64
	
	// 
	WindowSize      time.Duration
	EventBuffer     []*LearningEvent
	bufferMutex     sync.Mutex
}

// RealtimeLearnerData ?
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

// LearningSession 
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

// RealtimeMetrics 
type RealtimeMetrics struct {
	LearningVelocity    float64   `json:"learning_velocity"`
	EngagementTrend     string    `json:"engagement_trend"` // "increasing", "stable", "decreasing"
	FocusScore          float64   `json:"focus_score"`
	EfficiencyScore     float64   `json:"efficiency_score"`
	MotivationLevel     float64   `json:"motivation_level"`
	CognitiveLoad       float64   `json:"cognitive_load"`
	LastUpdated         time.Time `json:"last_updated"`
}

// BehaviorPatterns 
type BehaviorPatterns struct {
	InteractionFrequency map[string]float64 `json:"interaction_frequency"`
	NavigationPatterns   []string           `json:"navigation_patterns"`
	TimeDistribution     map[string]float64 `json:"time_distribution"`
	PreferredContentTypes []string          `json:"preferred_content_types"`
	LearningRhythm       string             `json:"learning_rhythm"` // "steady", "burst", "irregular"
	AttentionSpan        time.Duration      `json:"attention_span"`
}

// EngagementState ?
type EngagementState struct {
	Level               string    `json:"level"` // "high", "medium", "low", "disengaged"
	Score               float64   `json:"score"`
	Trend               string    `json:"trend"`
	LastInteraction     time.Time `json:"last_interaction"`
	InteractionQuality  float64   `json:"interaction_quality"`
	AttentionIndicators map[string]float64 `json:"attention_indicators"`
	RiskFactors         []string  `json:"risk_factors"`
}

// PerformanceState ?
type PerformanceState struct {
	CurrentLevel        string             `json:"current_level"` // "excellent", "good", "average", "below_average", "poor"
	Score               float64            `json:"score"`
	Trend               string             `json:"trend"`
	StrengthAreas       []string           `json:"strength_areas"`
	ImprovementAreas    []string           `json:"improvement_areas"`
	RecentAchievements  []string           `json:"recent_achievements"`
	PerformanceMetrics  map[string]float64 `json:"performance_metrics"`
}

// PredictiveInsights 
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

// RealtimeConfig 
type RealtimeConfig struct {
	EventBufferSize     int           `json:"event_buffer_size"`
	AnalysisInterval    time.Duration `json:"analysis_interval"`
	CacheExpiration     time.Duration `json:"cache_expiration"`
	MaxSubscribers      int           `json:"max_subscribers"`
	EnablePrediction    bool          `json:"enable_prediction"`
	AlertThresholds     map[string]float64 `json:"alert_thresholds"`
	AnalyzerConfigs     map[string]interface{} `json:"analyzer_configs"`
}

// NewRealtimeAnalyticsService 
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
	
	// 
	service.wg.Add(1)
	go service.processEventStream()
	
	service.wg.Add(1)
	go service.runPeriodicAnalysis()
	
	service.wg.Add(1)
	go service.cleanupExpiredData()
	
	return service
}

// ProcessEvent 
func (s *RealtimeAnalyticsService) ProcessEvent(event *LearningEvent) error {
	select {
	case s.eventStream <- event:
		return nil
	default:
		return fmt.Errorf("event stream buffer full")
	}
}

// Subscribe 
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

// Unsubscribe 
func (s *RealtimeAnalyticsService) Unsubscribe(subscriberID string) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	
	if ch, exists := s.subscribers[subscriberID]; exists {
		close(ch)
		delete(s.subscribers, subscriberID)
	}
}

// GetRealtimeData 
func (s *RealtimeAnalyticsService) GetRealtimeData(learnerID uuid.UUID) (*RealtimeLearnerData, error) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	if data, exists := s.realtimeCache[learnerID]; exists {
		return data, nil
	}
	
	return nil, fmt.Errorf("no realtime data found for learner %s", learnerID)
}

// CreateAnalyzer ?
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

// processEventStream ?
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

// handleEvent 
func (s *RealtimeAnalyticsService) handleEvent(event *LearningEvent) {
	// 
	s.updateRealtimeCache(event)
	
	// 
	s.distributeToAnalyzers(event)
	
	// 
	update := s.generateRealtimeAnalysis(event)
	if update != nil {
		s.broadcastUpdate(update)
	}
}

// updateRealtimeCache 
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
	
	// 
	data.LastActivity = event.Timestamp
	
	// 
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
	
	// 
	s.calculateRealtimeMetrics(data)
}

// Stop 
func (s *RealtimeAnalyticsService) Stop() {
	s.cancel()
	s.wg.Wait()
	
	// 
	s.subscribersMutex.Lock()
	for _, ch := range s.subscribers {
		close(ch)
	}
	s.subscribersMutex.Unlock()
}

// ...

// distributeToAnalyzers 
func (s *RealtimeAnalyticsService) distributeToAnalyzers(event *LearningEvent) {
	s.analyzersMutex.RLock()
	defer s.analyzersMutex.RUnlock()
	
	for _, analyzer := range s.analyzers {
		if analyzer.IsActive && analyzer.LearnerID == event.LearnerID {
			analyzer.bufferMutex.Lock()
			analyzer.EventBuffer = append(analyzer.EventBuffer, event)
			analyzer.EventCount++
			
			// 
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

// generateRealtimeAnalysis 
func (s *RealtimeAnalyticsService) generateRealtimeAnalysis(event *LearningEvent) *AnalyticsUpdate {
	data, exists := s.realtimeCache[event.LearnerID]
	if !exists {
		return nil
	}
	
	var insights []string
	var alerts []Alert
	
	// 
	insights = append(insights, s.generateEngagementInsights(data)...)
	insights = append(insights, s.generatePerformanceInsights(data)...)
	insights = append(insights, s.generateBehaviorInsights(data)...)
	
	// 龯?
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

// broadcastUpdate 㲥
func (s *RealtimeAnalyticsService) broadcastUpdate(update *AnalyticsUpdate) {
	s.subscribersMutex.RLock()
	defer s.subscribersMutex.RUnlock()
	
	for _, ch := range s.subscribers {
		select {
		case ch <- update:
		default:
			// ?
		}
	}
}

// updateProgressMetrics 
func (s *RealtimeAnalyticsService) updateProgressMetrics(data *RealtimeLearnerData, event *LearningEvent) {
	if progress, ok := event.Data["progress"].(float64); ok {
		if data.CurrentSession != nil {
			data.CurrentSession.ProgressMade += progress
		}
		
		// 
		if timeSpent, ok := event.Data["time_spent"].(float64); ok && timeSpent > 0 {
			velocity := progress / (timeSpent / 3600) // ?
			data.RealtimeMetrics.LearningVelocity = velocity
		}
	}
}

// updateInteractionMetrics 
func (s *RealtimeAnalyticsService) updateInteractionMetrics(data *RealtimeLearnerData, event *LearningEvent) {
	if data.CurrentSession != nil {
		data.CurrentSession.InteractionCount++
		data.CurrentSession.LastActivity = event.Timestamp
	}
	
	// 
	if interactionType, ok := event.Data["interaction_type"].(string); ok {
		data.BehaviorPatterns.InteractionFrequency[interactionType]++
	}
	
	// ?
	if duration, ok := event.Data["duration"].(float64); ok {
		data.RealtimeMetrics.FocusScore = s.calculateFocusScore(duration, data.BehaviorPatterns)
	}
}

// updateEngagementMetrics ?
func (s *RealtimeAnalyticsService) updateEngagementMetrics(data *RealtimeLearnerData, event *LearningEvent) {
	if engagementScore, ok := event.Data["engagement_score"].(float64); ok {
		data.EngagementState.Score = engagementScore
		data.EngagementState.LastInteraction = event.Timestamp
		
		// ?
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

// updatePerformanceMetrics 
func (s *RealtimeAnalyticsService) updatePerformanceMetrics(data *RealtimeLearnerData, event *LearningEvent) {
	if score, ok := event.Data["score"].(float64); ok {
		data.PerformanceState.Score = score
		
		// 
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

// calculateRealtimeMetrics 
func (s *RealtimeAnalyticsService) calculateRealtimeMetrics(data *RealtimeLearnerData) {
	metrics := data.RealtimeMetrics
	
	// 
	if data.CurrentSession != nil && data.CurrentSession.Duration > 0 {
		efficiency := data.CurrentSession.ProgressMade / data.CurrentSession.Duration.Hours()
		metrics.EfficiencyScore = math.Min(efficiency, 1.0)
	}
	
	// 
	metrics.CognitiveLoad = s.calculateCognitiveLoad(data)
	
	// 㶯
	metrics.MotivationLevel = s.calculateMotivationLevel(data)
	
	// 
	metrics.EngagementTrend = s.calculateEngagementTrend(data)
	
	metrics.LastUpdated = time.Now()
}

// generateEngagementInsights ?
func (s *RealtimeAnalyticsService) generateEngagementInsights(data *RealtimeLearnerData) []string {
	var insights []string
	
	engagement := data.EngagementState
	
	if engagement.Score > 0.8 {
		insights = append(insights, "?)
	} else if engagement.Score < 0.3 {
		insights = append(insights, "")
		
		// ?
		if data.RealtimeMetrics.CognitiveLoad > 0.8 {
			insights = append(insights, "")
		}
	}
	
	// 
	if len(data.BehaviorPatterns.InteractionFrequency) > 0 {
		maxInteraction := ""
		maxCount := 0.0
		for interaction, count := range data.BehaviorPatterns.InteractionFrequency {
			if count > maxCount {
				maxCount = count
				maxInteraction = interaction
			}
		}
		insights = append(insights, fmt.Sprintf(": %s", maxInteraction))
	}
	
	return insights
}

// generatePerformanceInsights 
func (s *RealtimeAnalyticsService) generatePerformanceInsights(data *RealtimeLearnerData) []string {
	var insights []string
	
	performance := data.PerformanceState
	
	switch performance.CurrentLevel {
	case "excellent":
		insights = append(insights, "")
	case "poor", "below_average":
		insights = append(insights, "?)
	}
	
	// 
	if data.RealtimeMetrics.EfficiencyScore > 0.8 {
		insights = append(insights, "?)
	} else if data.RealtimeMetrics.EfficiencyScore < 0.3 {
		insights = append(insights, "?)
	}
	
	return insights
}

// generateBehaviorInsights 
func (s *RealtimeAnalyticsService) generateBehaviorInsights(data *RealtimeLearnerData) []string {
	var insights []string
	
	// 
	switch data.BehaviorPatterns.LearningRhythm {
	case "burst":
		insights = append(insights, "")
	case "irregular":
		insights = append(insights, "?)
	case "steady":
		insights = append(insights, "")
	}
	
	// ?
	if data.BehaviorPatterns.AttentionSpan < 15*time.Minute {
		insights = append(insights, "?)
	} else if data.BehaviorPatterns.AttentionSpan > 2*time.Hour {
		insights = append(insights, "")
	}
	
	return insights
}

// checkEngagementAlerts 
func (s *RealtimeAnalyticsService) checkEngagementAlerts(data *RealtimeLearnerData) []Alert {
	var alerts []Alert
	
	if data.EngagementState.Score < s.config.AlertThresholds["low_engagement"] {
		alerts = append(alerts, Alert{
			Level:     "warning",
			Message:   "",
			Timestamp: time.Now(),
			ActionItems: []string{
				"",
				"",
				"",
			},
		})
	}
	
	// 鳤?
	if time.Since(data.EngagementState.LastInteraction) > 10*time.Minute {
		alerts = append(alerts, Alert{
			Level:     "info",
			Message:   "?,
			Timestamp: time.Now(),
			ActionItems: []string{
				"",
				"",
			},
		})
	}
	
	return alerts
}

// checkPerformanceAlerts ?
func (s *RealtimeAnalyticsService) checkPerformanceAlerts(data *RealtimeLearnerData) []Alert {
	var alerts []Alert
	
	if data.PerformanceState.Score < s.config.AlertThresholds["performance_drop"] {
		alerts = append(alerts, Alert{
			Level:     "warning",
			Message:   "",
			Timestamp: time.Now(),
			ActionItems: []string{
				"",
				"",
				"",
			},
		})
	}
	
	return alerts
}

// checkBehaviorAlerts ?
func (s *RealtimeAnalyticsService) checkBehaviorAlerts(data *RealtimeLearnerData) []Alert {
	var alerts []Alert
	
	if data.RealtimeMetrics.CognitiveLoad > s.config.AlertThresholds["high_cognitive_load"] {
		alerts = append(alerts, Alert{
			Level:     "warning",
			Message:   "",
			Timestamp: time.Now(),
			ActionItems: []string{
				"",
				"",
				"",
			},
		})
	}
	
	if data.PredictiveInsights.RiskOfDropout > s.config.AlertThresholds["dropout_risk"] {
		alerts = append(alerts, Alert{
			Level:     "critical",
			Message:   "?,
			Timestamp: time.Now(),
			ActionItems: []string{
				"",
				"",
				"",
			},
		})
	}
	
	return alerts
}

// 㷽
func (s *RealtimeAnalyticsService) calculateFocusScore(duration float64, patterns *BehaviorPatterns) float64 {
	// 
	baseScore := math.Min(duration/3600, 1.0) // 1
	
	// 
	var avgFreq float64
	for _, freq := range patterns.InteractionFrequency {
		avgFreq += freq
	}
	if len(patterns.InteractionFrequency) > 0 {
		avgFreq /= float64(len(patterns.InteractionFrequency))
	}
	
	// ?
	freqScore := 1.0 - math.Abs(avgFreq-0.5)*2
	
	return (baseScore + freqScore) / 2
}

func (s *RealtimeAnalyticsService) calculateCognitiveLoad(data *RealtimeLearnerData) float64 {
	// 
	load := 0.0
	
	// 
	if len(data.BehaviorPatterns.InteractionFrequency) > 0 {
		var totalInteractions float64
		for _, freq := range data.BehaviorPatterns.InteractionFrequency {
			totalInteractions += freq
		}
		// ?
		if totalInteractions > 100 {
			load += 0.3
		}
	}
	
	// 
	if data.RealtimeMetrics.EfficiencyScore < 0.3 {
		load += 0.4
	}
	
	// ?
	if data.EngagementState.Score < 0.4 {
		load += 0.3
	}
	
	return math.Min(load, 1.0)
}

func (s *RealtimeAnalyticsService) calculateMotivationLevel(data *RealtimeLearnerData) float64 {
	// 㶯
	motivation := 0.0
	
	// ?
	motivation += data.EngagementState.Score * 0.4
	
	// 
	motivation += data.PerformanceState.Score * 0.3
	
	// 
	motivation += data.RealtimeMetrics.EfficiencyScore * 0.3
	
	return math.Min(motivation, 1.0)
}

func (s *RealtimeAnalyticsService) calculateEngagementTrend(data *RealtimeLearnerData) string {
	// ?
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
	// 
	confidence := 0.5 // ?
	
	// ?
	if time.Since(data.LastActivity) < 5*time.Minute {
		confidence += 0.3
	}
	
	// ?
	if data.CurrentSession != nil && data.CurrentSession.InteractionCount > 5 {
		confidence += 0.2
	}
	
	return math.Min(confidence, 1.0)
}

// runPeriodicAnalysis ?
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

// performPeriodicAnalysis ?
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

// analyzelearnerPredictions ?
func (s *RealtimeAnalyticsService) analyzelearnerPredictions(learnerID uuid.UUID) {
	s.cacheMutex.Lock()
	data, exists := s.realtimeCache[learnerID]
	if !exists {
		s.cacheMutex.Unlock()
		return
	}
	
	// 
	data.PredictiveInsights.CompletionProbability = s.predictCompletionProbability(data)
	data.PredictiveInsights.RiskOfDropout = s.predictDropoutRisk(data)
	data.PredictiveInsights.EstimatedCompletionTime = s.estimateCompletionTime(data)
	data.PredictiveInsights.RecommendedActions = s.generateRecommendedActions(data)
	
	s.cacheMutex.Unlock()
	
	// 㲥
	update := &AnalyticsUpdate{
		Type:       "prediction_update",
		LearnerID:  learnerID,
		Timestamp:  time.Now(),
		Data:       data.PredictiveInsights,
		Confidence: s.calculateConfidence(data),
	}
	
	s.broadcastUpdate(update)
}

// 
func (s *RealtimeAnalyticsService) predictCompletionProbability(data *RealtimeLearnerData) float64 {
	// 
	score := 0.0
	
	// 
	score += data.PerformanceState.Score * 0.4
	
	// ?
	score += data.EngagementState.Score * 0.3
	
	// 
	score += data.RealtimeMetrics.EfficiencyScore * 0.3
	
	return math.Min(score, 1.0)
}

func (s *RealtimeAnalyticsService) predictDropoutRisk(data *RealtimeLearnerData) float64 {
	// 
	risk := 0.0
	
	// 
	if data.EngagementState.Score < 0.3 {
		risk += 0.4
	}
	
	// ?
	if data.PerformanceState.Score < 0.4 {
		risk += 0.3
	}
	
	// ?
	if data.RealtimeMetrics.CognitiveLoad > 0.8 {
		risk += 0.3
	}
	
	return math.Min(risk, 1.0)
}

func (s *RealtimeAnalyticsService) estimateCompletionTime(data *RealtimeLearnerData) time.Time {
	// 
	baseTime := time.Now().Add(30 * 24 * time.Hour) // 30?
	
	// 
	if data.RealtimeMetrics.LearningVelocity > 0.5 {
		baseTime = baseTime.Add(-10 * 24 * time.Hour) // 10?
	} else if data.RealtimeMetrics.LearningVelocity < 0.2 {
		baseTime = baseTime.Add(20 * 24 * time.Hour) // 20?
	}
	
	return baseTime
}

func (s *RealtimeAnalyticsService) generateRecommendedActions(data *RealtimeLearnerData) []string {
	var actions []string
	
	// ?
	if data.EngagementState.Score < 0.4 {
		actions = append(actions, "", "")
	}
	
	// 
	if data.PerformanceState.Score < 0.5 {
		actions = append(actions, "", "")
	}
	
	// 
	if data.RealtimeMetrics.CognitiveLoad > 0.7 {
		actions = append(actions, "", "")
	}
	
	return actions
}

// cleanupExpiredData 
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

// performCleanup 
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

