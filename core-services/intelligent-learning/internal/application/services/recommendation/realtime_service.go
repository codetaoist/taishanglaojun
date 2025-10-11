package recommendation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// RealtimeRecommendationService 实时推荐服务
type RealtimeRecommendationService struct {
	personalizationEngine *domainServices.PersonalizationEngine
	userBehaviorTracker   *domainServices.UserBehaviorTracker
	preferenceAnalyzer    *domainServices.PreferenceAnalyzer
	contextAnalyzer       *domainServices.ContextAnalyzer
	
	// 实时数据缓存
	userSessions    map[string]*RealtimeUserSession
	sessionMutex    sync.RWMutex
	
	// 推荐缓存
	recommendationCache map[string]*CachedRecommendations
	cacheMutex         sync.RWMutex
	
	// 事件处理
	eventChannel    chan *RealtimeEvent
	subscribers     map[string]chan *RecommendationUpdate
	subscriberMutex sync.RWMutex
	
	// 配置
	config *RealtimeConfig
}

// RealtimeUserSession 用户会话
type RealtimeUserSession struct {
	UserID        string                 `json:"user_id"`
	SessionID     string                 `json:"session_id"`
	StartTime     time.Time              `json:"start_time"`
	LastActivity  time.Time              `json:"last_activity"`
	Events        []*RealtimeEvent       `json:"events"`
	Context       map[string]interface{} `json:"context"`
	CurrentState  *LearningState         `json:"current_state"`
	Preferences   *domainServices.UserPreferences       `json:"preferences"`
}

// LearningState 学习状�?
type LearningState struct {
	CurrentContent    string                 `json:"current_content"`
	Progress          float64                `json:"progress"`
	Engagement        float64                `json:"engagement"`
	Difficulty        string                 `json:"difficulty"`
	LearningStyle     string                 `json:"learning_style"`
	FocusLevel        float64                `json:"focus_level"`
	ComprehensionRate float64                `json:"comprehension_rate"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// RealtimeEvent 实时事件
type RealtimeEvent struct {
	EventID     string                 `json:"event_id"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	EventType   string                 `json:"event_type"`
	ContentID   string                 `json:"content_id,omitempty"`
	Action      string                 `json:"action"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    int64                  `json:"duration,omitempty"`
	Properties  map[string]interface{} `json:"properties"`
	Context     *EventContext          `json:"context"`
}

// EventContext 事件上下�?
type EventContext struct {
	DeviceType    string                 `json:"device_type"`
	Platform      string                 `json:"platform"`
	Location      string                 `json:"location,omitempty"`
	TimeOfDay     string                 `json:"time_of_day"`
	Environment   map[string]interface{} `json:"environment"`
}

// CachedRecommendations 缓存的推�?
type CachedRecommendations struct {
	UserID        string                      `json:"user_id"`
	Recommendations []*domainServices.PersonalizedRecommendation `json:"recommendations"`
	GeneratedAt   time.Time                   `json:"generated_at"`
	ExpiresAt     time.Time                   `json:"expires_at"`
	Version       int                         `json:"version"`
	Strategy      string                      `json:"strategy"`
}

// RecommendationUpdate 推荐更新
type RecommendationUpdate struct {
	UserID        string                      `json:"user_id"`
	UpdateType    string                      `json:"update_type"`
	Recommendations []*domainServices.PersonalizedRecommendation `json:"recommendations"`
	Reason        string                      `json:"reason"`
	Timestamp     time.Time                   `json:"timestamp"`
	Metadata      map[string]interface{}      `json:"metadata"`
}

// RealtimeConfig 实时配置
type RealtimeConfig struct {
	CacheExpiration     time.Duration `json:"cache_expiration"`
	SessionTimeout      time.Duration `json:"session_timeout"`
	EventBufferSize     int           `json:"event_buffer_size"`
	UpdateThreshold     float64       `json:"update_threshold"`
	MinUpdateInterval   time.Duration `json:"min_update_interval"`
	MaxRecommendations  int           `json:"max_recommendations"`
	EnablePredictive    bool          `json:"enable_predictive"`
	EnableAdaptive      bool          `json:"enable_adaptive"`
}

// NewRealtimeRecommendationService 创建实时推荐服务
func NewRealtimeRecommendationService(
	personalizationEngine *domainServices.PersonalizationEngine,
	userBehaviorTracker *domainServices.UserBehaviorTracker,
	preferenceAnalyzer *domainServices.PreferenceAnalyzer,
	contextAnalyzer *domainServices.ContextAnalyzer,
) *RealtimeRecommendationService {
	config := &RealtimeConfig{
		CacheExpiration:     30 * time.Minute,
		SessionTimeout:      2 * time.Hour,
		EventBufferSize:     1000,
		UpdateThreshold:     0.3,
		MinUpdateInterval:   5 * time.Minute,
		MaxRecommendations:  20,
		EnablePredictive:    true,
		EnableAdaptive:      true,
	}

	service := &RealtimeRecommendationService{
		personalizationEngine: personalizationEngine,
		userBehaviorTracker:   userBehaviorTracker,
		preferenceAnalyzer:    preferenceAnalyzer,
		contextAnalyzer:       contextAnalyzer,
		userSessions:          make(map[string]*RealtimeUserSession),
		recommendationCache:   make(map[string]*CachedRecommendations),
		eventChannel:          make(chan *RealtimeEvent, config.EventBufferSize),
		subscribers:           make(map[string]chan *RecommendationUpdate),
		config:                config,
	}

	// 启动事件处理�?
	go service.processEvents()
	
	// 启动会话清理�?
	go service.cleanupSessions()

	return service
}

// ProcessEvent 处理实时事件
func (s *RealtimeRecommendationService) ProcessEvent(ctx context.Context, event *RealtimeEvent) error {
	// 验证事件
	if err := s.validateEvent(event); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	// 设置事件ID和时间戳
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 发送到事件通道
	select {
	case s.eventChannel <- event:
		return nil
	default:
		return fmt.Errorf("event channel is full")
	}
}

// GetRealtimeRecommendations 获取实时推荐
func (s *RealtimeRecommendationService) GetRealtimeRecommendations(ctx context.Context, userID string) ([]*domainServices.PersonalizedRecommendation, error) {
	// 检查缓�?
	s.cacheMutex.RLock()
	cached, exists := s.recommendationCache[userID]
	s.cacheMutex.RUnlock()

	if exists && time.Now().Before(cached.ExpiresAt) {
		return cached.Recommendations, nil
	}

	// 生成新的推荐
	return s.generateRealtimeRecommendations(ctx, userID)
}

// SubscribeToUpdates 订阅推荐更新
func (s *RealtimeRecommendationService) SubscribeToUpdates(userID string) <-chan *RecommendationUpdate {
	s.subscriberMutex.Lock()
	defer s.subscriberMutex.Unlock()

	updateChannel := make(chan *RecommendationUpdate, 100)
	s.subscribers[userID] = updateChannel
	
	return updateChannel
}

// UnsubscribeFromUpdates 取消订阅推荐更新
func (s *RealtimeRecommendationService) UnsubscribeFromUpdates(userID string) {
	s.subscriberMutex.Lock()
	defer s.subscriberMutex.Unlock()

	if channel, exists := s.subscribers[userID]; exists {
		close(channel)
		delete(s.subscribers, userID)
	}
}

// GetUserSession 获取用户会话
func (s *RealtimeRecommendationService) GetUserSession(userID string) (*RealtimeUserSession, error) {
	s.sessionMutex.RLock()
	defer s.sessionMutex.RUnlock()

	session, exists := s.userSessions[userID]
	if !exists {
		return nil, fmt.Errorf("session not found for user: %s", userID)
	}

	return session, nil
}

// processEvents 处理事件
func (s *RealtimeRecommendationService) processEvents() {
	for event := range s.eventChannel {
		s.handleEvent(event)
	}
}

// handleEvent 处理单个事件
func (s *RealtimeRecommendationService) handleEvent(event *RealtimeEvent) {
	// 更新用户会话
	s.updateUserSession(event)

	// 记录行为
	behaviorEvent := &domainServices.BehaviorEvent{
		LearnerID:   uuid.MustParse(event.UserID),
		SessionID:   uuid.MustParse(event.SessionID),
		EventType:   event.EventType,
		Timestamp:   event.Timestamp,
		Duration:    time.Duration(event.Duration) * time.Millisecond,
		Properties:  event.Properties,
		Context:     &domainServices.EventContext{},
	}

	if event.Context != nil {
		behaviorEvent.Context = &domainServices.EventContext{
			Device:      event.Context.DeviceType,
			Platform:    event.Context.Platform,
			Location:    event.Context.Location,
			NetworkType: "",
			UserAgent:   "",
			Referrer:    "",
			TimeZone:    event.Context.TimeOfDay,
			Language:    "",
			ScreenSize:  "",
			Metadata:    event.Context.Environment,
		}
	}

	// 异步记录行为
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.userBehaviorTracker.TrackBehaviorEvent(ctx, behaviorEvent)
	}()

	// 检查是否需要更新推�?
	if s.shouldUpdateRecommendations(event) {
		go s.updateRecommendations(event.UserID, event)
	}
}

// updateUserSession 更新用户会话
func (s *RealtimeRecommendationService) updateUserSession(event *RealtimeEvent) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	session, exists := s.userSessions[event.UserID]
	if !exists {
		session = &RealtimeUserSession{
			UserID:       event.UserID,
			SessionID:    event.SessionID,
			StartTime:    event.Timestamp,
			LastActivity: event.Timestamp,
			Events:       []*RealtimeEvent{},
			Context:      make(map[string]interface{}),
			CurrentState: &LearningState{},
		}
		s.userSessions[event.UserID] = session
	}

	// 更新会话信息
	session.LastActivity = event.Timestamp
	session.Events = append(session.Events, event)

	// 限制事件数量
	if len(session.Events) > 100 {
		session.Events = session.Events[len(session.Events)-100:]
	}

	// 更新学习状�?
	s.updateLearningState(session, event)
}

// updateLearningState 更新学习状�?
func (s *RealtimeRecommendationService) updateLearningState(session *RealtimeUserSession, event *RealtimeEvent) {
	state := session.CurrentState

	switch event.EventType {
	case "content_view":
		state.CurrentContent = event.ContentID
		if progress, ok := event.Properties["progress"].(float64); ok {
			state.Progress = progress
		}
	case "engagement":
		if engagement, ok := event.Properties["engagement"].(float64); ok {
			state.Engagement = engagement
		}
	case "comprehension":
		if rate, ok := event.Properties["comprehension_rate"].(float64); ok {
			state.ComprehensionRate = rate
		}
	case "focus":
		if focus, ok := event.Properties["focus_level"].(float64); ok {
			state.FocusLevel = focus
		}
	}

	// 更新难度偏好
	if difficulty, ok := event.Properties["difficulty"].(string); ok {
		state.Difficulty = difficulty
	}

	// 更新学习风格
	if style, ok := event.Properties["learning_style"].(string); ok {
		state.LearningStyle = style
	}
}

// shouldUpdateRecommendations 检查是否应该更新推�?
func (s *RealtimeRecommendationService) shouldUpdateRecommendations(event *RealtimeEvent) bool {
	// 检查缓�?
	s.cacheMutex.RLock()
	cached, exists := s.recommendationCache[event.UserID]
	s.cacheMutex.RUnlock()

	if !exists {
		return true
	}

	// 检查最小更新间�?
	if time.Since(cached.GeneratedAt) < s.config.MinUpdateInterval {
		return false
	}

	// 基于事件类型决定
	switch event.EventType {
	case "content_complete", "quiz_complete", "skill_mastery":
		return true
	case "content_view":
		// 检查观看时�?
		if duration, ok := event.Properties["duration"].(int64); ok {
			return duration > 300000 // 5分钟以上
		}
	case "engagement":
		// 检查参与度变化
		if engagement, ok := event.Properties["engagement"].(float64); ok {
			return engagement < 0.3 || engagement > 0.8
		}
	}

	return false
}

// updateRecommendations 更新推荐
func (s *RealtimeRecommendationService) updateRecommendations(userID string, triggerEvent *RealtimeEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 生成新推�?
	recommendations, err := s.generateRealtimeRecommendations(ctx, userID)
	if err != nil {
		return
	}

	// 更新缓存
	s.cacheMutex.Lock()
	s.recommendationCache[userID] = &CachedRecommendations{
		UserID:          userID,
		Recommendations: recommendations,
		GeneratedAt:     time.Now(),
		ExpiresAt:       time.Now().Add(s.config.CacheExpiration),
		Version:         s.getNextVersion(userID),
		Strategy:        "realtime",
	}
	s.cacheMutex.Unlock()

	// 通知订阅�?
	s.notifySubscribers(userID, recommendations, triggerEvent)
}

// generateRealtimeRecommendations 生成实时推荐
func (s *RealtimeRecommendationService) generateRealtimeRecommendations(ctx context.Context, userID string) ([]*domainServices.PersonalizedRecommendation, error) {
	// 获取用户会话
	session, err := s.GetUserSession(userID)
	if err != nil {
		// 如果没有会话，使用标准推�?
		req := &domainServices.PersonalizationRequest{
			LearnerID:          uuid.MustParse(userID),
			MaxRecommendations: s.config.MaxRecommendations,
		}
		response, err := s.personalizationEngine.GeneratePersonalizedRecommendations(ctx, req)
		if err != nil {
			return nil, err
		}
		// 转换为指针切�?
		result := make([]*domainServices.PersonalizedRecommendation, len(response.Recommendations))
		for i := range response.Recommendations {
			result[i] = &response.Recommendations[i]
		}
		return result, nil
	}

	// 构建实时上下�?
	realtimeContext := s.buildRealtimeContext(session)

	// 生成个性化推荐
	req := &domainServices.PersonalizationRequest{
		LearnerID:          uuid.MustParse(userID),
		MaxRecommendations: s.config.MaxRecommendations,
		Context:            realtimeContext,
	}

	response, err := s.personalizationEngine.GeneratePersonalizedRecommendations(ctx, req)
	if err != nil {
		return nil, err
	}

	// 转换为指针切�?
	recommendationPtrs := make([]*domainServices.PersonalizedRecommendation, len(response.Recommendations))
	for i := range response.Recommendations {
		recommendationPtrs[i] = &response.Recommendations[i]
	}

	// 应用实时调整
	adjustedRecommendations := s.applyRealtimeAdjustments(recommendationPtrs, session)

	return adjustedRecommendations, nil
}

// buildRealtimeContext 构建实时上下�?
func (s *RealtimeRecommendationService) buildRealtimeContext(session *RealtimeUserSession) *domainServices.PersonalizationContext {
	metadata := make(map[string]interface{})

	// 会话信息
	metadata["session_duration"] = time.Since(session.StartTime).Minutes()
	metadata["activity_count"] = len(session.Events)
	metadata["last_activity"] = session.LastActivity

	// 学习状�?
	currentContent := ""
	if session.CurrentState != nil {
		currentContent = session.CurrentState.CurrentContent
		metadata["progress"] = session.CurrentState.Progress
		metadata["engagement"] = session.CurrentState.Engagement
		metadata["focus_level"] = session.CurrentState.FocusLevel
		metadata["comprehension_rate"] = session.CurrentState.ComprehensionRate
		metadata["difficulty"] = session.CurrentState.Difficulty
		metadata["learning_style"] = session.CurrentState.LearningStyle
	}

	// 最近事件分�?
	recentEvents := s.getRecentEvents(session, 10)
	recentActivity := s.analyzeRecentEvents(recentEvents)

	return &domainServices.PersonalizationContext{
		SessionID:      session.SessionID,
		Device:         "unknown", // 可以从session中获�?
		Location:       "unknown", // 可以从session中获�?
		TimeOfDay:      time.Now().Format("15:04"),
		AvailableTime:  30, // 默认30分钟，可以根据实际情况调�?
		EnergyLevel:    0.8, // 默认值，可以根据用户行为分析
		Goals:          []string{}, // 可以从用户配置中获取
		CurrentContent: currentContent,
		RecentActivity: recentActivity,
		SocialContext:  "individual", // 默认个人学习
		Metadata:       metadata,
	}

	return &domainServices.PersonalizationContext{
		SessionID:      session.SessionID,
		Device:         "unknown", // 可以从session中获�?
		Location:       "unknown", // 可以从session中获�?
		TimeOfDay:      time.Now().Format("15:04"),
		AvailableTime:  30, // 默认30分钟，可以根据实际情况调�?
		EnergyLevel:    0.8, // 默认值，可以根据用户行为分析
		Goals:          []string{}, // 可以从用户配置中获取
		CurrentContent: currentContent,
		RecentActivity: recentActivity,
		SocialContext:  "individual", // 默认个人学习
		Metadata:       metadata,
	}
}

// applyRealtimeAdjustments 应用实时调整
func (s *RealtimeRecommendationService) applyRealtimeAdjustments(recommendations []*domainServices.PersonalizedRecommendation, session *RealtimeUserSession) []*domainServices.PersonalizedRecommendation {
	if session.CurrentState == nil {
		return recommendations
	}

	state := session.CurrentState
	adjusted := make([]*domainServices.PersonalizedRecommendation, len(recommendations))
	copy(adjusted, recommendations)

	for _, rec := range adjusted {
		// 基于参与度调�?
		if state.Engagement < 0.3 {
			// 低参与度，推荐更有趣的内�?
			if rec.Type == "interactive" || rec.Type == "game" {
				rec.Score *= 1.3
			}
		} else if state.Engagement > 0.8 {
			// 高参与度，可以推荐更有挑战性的内容
			if rec.Difficulty == "advanced" {
				rec.Score *= 1.2
			}
		}

		// 基于理解率调�?
		if state.ComprehensionRate < 0.5 {
			// 理解率低，推荐基础内容
			if rec.Difficulty == "beginner" || rec.Difficulty == "intermediate" {
				rec.Score *= 1.2
			}
		}

		// 基于专注度调�?
		if state.FocusLevel < 0.4 {
			// 专注度低，推荐短时间内容
			if rec.EstimatedTime <= 15 {
				rec.Score *= 1.25
			}
		}

		// 基于学习风格调整
		if state.LearningStyle == "visual" && rec.Type == "video" {
			rec.Score *= 1.15
		} else if state.LearningStyle == "auditory" && rec.Type == "audio" {
			rec.Score *= 1.15
		} else if state.LearningStyle == "kinesthetic" && rec.Type == "interactive" {
			rec.Score *= 1.15
		}
	}

	return adjusted
}

// 辅助方法
func (s *RealtimeRecommendationService) validateEvent(event *RealtimeEvent) error {
	if event.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if event.Action == "" {
		return fmt.Errorf("action is required")
	}
	return nil
}

func (s *RealtimeRecommendationService) getNextVersion(userID string) int {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	if cached, exists := s.recommendationCache[userID]; exists {
		return cached.Version + 1
	}
	return 1
}

func (s *RealtimeRecommendationService) notifySubscribers(userID string, recommendations []*domainServices.PersonalizedRecommendation, triggerEvent *RealtimeEvent) {
	s.subscriberMutex.RLock()
	defer s.subscriberMutex.RUnlock()

	if channel, exists := s.subscribers[userID]; exists {
		update := &RecommendationUpdate{
			UserID:          userID,
			UpdateType:      "realtime_update",
			Recommendations: recommendations,
			Reason:          fmt.Sprintf("Triggered by %s event", triggerEvent.EventType),
			Timestamp:       time.Now(),
			Metadata: map[string]interface{}{
				"trigger_event": triggerEvent.EventType,
				"trigger_action": triggerEvent.Action,
			},
		}

		select {
		case channel <- update:
		default:
			// 通道已满，跳过此次更�?
		}
	}
}

func (s *RealtimeRecommendationService) getRecentEvents(session *RealtimeUserSession, count int) []*RealtimeEvent {
	events := session.Events
	if len(events) <= count {
		return events
	}
	return events[len(events)-count:]
}

func (s *RealtimeRecommendationService) analyzeRecentEvents(events []*RealtimeEvent) []string {
	activities := make([]string, 0)
	
	if len(events) == 0 {
		return activities
	}

	// 收集最近的活动描述
	for _, event := range events {
		activity := fmt.Sprintf("%s:%s", event.EventType, event.Action)
		if event.ContentID != "" {
			activity += fmt.Sprintf(":%s", event.ContentID)
		}
		activities = append(activities, activity)
	}

	return activities
}

func (s *RealtimeRecommendationService) getTimeOfDay(t time.Time) string {
	hour := t.Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func (s *RealtimeRecommendationService) cleanupSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.sessionMutex.Lock()
		now := time.Now()
		for userID, session := range s.userSessions {
			if now.Sub(session.LastActivity) > s.config.SessionTimeout {
				delete(s.userSessions, userID)
			}
		}
		s.sessionMutex.Unlock()

		// 清理过期缓存
		s.cacheMutex.Lock()
		for userID, cached := range s.recommendationCache {
			if now.After(cached.ExpiresAt) {
				delete(s.recommendationCache, userID)
			}
		}
		s.cacheMutex.Unlock()
	}
}
