package realtime

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// 
const (
	TrendUp   = "up"
	TrendDown = "down"
	TrendFlat = "flat"
)

// Trend 
type Trend struct {
	Direction   string    `json:"direction"`
	Strength    float64   `json:"strength"`
	Confidence  float64   `json:"confidence"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
}

// 

// updateSessionInfo 
func (s *RealtimeLearningAnalyticsService) updateSessionInfo(
	sessionInfo *SessionInfo,
	data map[string]interface{},
) {
	if duration, ok := data["duration"].(float64); ok {
		sessionInfo.Duration = int64(duration)
	}
	
	// SessionInfoDeviceInfo豸InteractionContext?
	// device_info?
	
	// GoalsLearningSessionSessionInfo?
	// goalsupdateLearningSession?
}

// updateDeviceInfo 豸
func (s *RealtimeLearningAnalyticsService) updateDeviceInfo(
	deviceInfo *DeviceInfo,
	data map[string]interface{},
) {
	if deviceType, ok := data["device_type"].(string); ok {
		deviceInfo.Type = deviceType
	}
	
	if platform, ok := data["platform"].(string); ok {
		deviceInfo.OS = platform
	}
	
	if browser, ok := data["browser"].(string); ok {
		deviceInfo.Browser = browser
	}
	
	if screenResolution, ok := data["screen_resolution"].(string); ok {
		deviceInfo.ScreenSize = screenResolution
	}
}

// parseSessionGoals 
func (s *RealtimeLearningAnalyticsService) parseSessionGoals(
	goals []interface{},
) []*SessionGoal {
	sessionGoals := make([]*SessionGoal, 0, len(goals))
	
	for _, goalData := range goals {
		if goalMap, ok := goalData.(map[string]interface{}); ok {
			deadline := time.Now().Add(24 * time.Hour)
		goal := &SessionGoal{
			GoalID:      uuid.New(),
			Type:        GoalTypeCompletion,
			Description: "",
			Target:      0,
			Current:     0,
			Progress:    0.0,
			Priority:    1,
			Deadline:    &deadline,
			Status:      GoalStatusPending,
			Metadata:    make(map[string]interface{}),
		}
		
		if description, ok := goalMap["description"].(string); ok {
			goal.Description = description
		}
		
		if targetValue, ok := goalMap["target_value"].(float64); ok {
			goal.Target = targetValue
		}
		
		if currentValue, ok := goalMap["current_value"].(float64); ok {
			goal.Current = currentValue
		}
		
		if priority, ok := goalMap["priority"].(float64); ok {
			goal.Priority = int(priority)
		}
			
			sessionGoals = append(sessionGoals, goal)
		}
	}
	
	return sessionGoals
}

// updateCurrentActivity 
func (s *RealtimeLearningAnalyticsService) updateCurrentActivity(
	activity *LearningActivity,
	data map[string]interface{},
) {
	if activityType, ok := data["activity_type"].(string); ok {
		activity.Type = ActivityType(activityType)
	}
	
	if duration, ok := data["duration"].(float64); ok {
		activity.Duration = time.Duration(duration) * time.Second
	}
	
	if score, ok := data["score"].(float64); ok {
		activity.Score = &score
	}
	
	if attempts, ok := data["attempts"].(float64); ok {
		activity.Attempts = int(attempts)
	}
	
	if hints, ok := data["hints"].(float64); ok {
		activity.Hints = int(hints)
	}
	
	if success, ok := data["success"].(bool); ok {
		activity.Success = success
	}
	
	// 洢Metadata?
	if activity.Metadata == nil {
		activity.Metadata = make(map[string]interface{})
	}
	
	if contentID, ok := data["content_id"].(string); ok {
		activity.Metadata["content_id"] = contentID
	}
	
	if progress, ok := data["progress"].(float64); ok {
		activity.Metadata["progress"] = progress
	}
	
	if difficulty, ok := data["difficulty"].(float64); ok {
		activity.Metadata["difficulty"] = difficulty
	}
	
	if engagement, ok := data["engagement"].(float64); ok {
		activity.Metadata["engagement"] = engagement
	}
}

// updatePerformanceMetrics 
func (s *RealtimeLearningAnalyticsService) updatePerformanceMetrics(
	metrics *domainServices.PerformanceMetrics,
	data map[string]interface{},
) {
	if accuracy, ok := data["accuracy"].(float64); ok {
		metrics.Accuracy = accuracy
	}
	
	if speed, ok := data["speed"].(float64); ok {
		metrics.Speed = speed
	}
	
	if efficiency, ok := data["efficiency"].(float64); ok {
		metrics.Efficiency = efficiency
	}
	
	if completionRate, ok := data["completion_rate"].(float64); ok {
		metrics.CompletionRate = completionRate
	}
	
	if errorRate, ok := data["error_rate"].(float64); ok {
		metrics.ErrorRate = errorRate
	}
	
	if consistency, ok := data["consistency"].(float64); ok {
		metrics.Consistency = consistency
	}
	
	// retentioncomprehensionengagement洢ExpectedOutcome JSON
	expectedOutcome := make(map[string]interface{})
	
	// ExpectedOutcome
	if metrics.ExpectedOutcome != "" {
		json.Unmarshal([]byte(metrics.ExpectedOutcome), &expectedOutcome)
	}
	
	if retention, ok := data["retention"].(float64); ok {
		expectedOutcome["retention"] = retention
	}
	
	if comprehension, ok := data["comprehension"].(float64); ok {
		expectedOutcome["comprehension"] = comprehension
	}
	
	if engagement, ok := data["engagement"].(float64); ok {
		expectedOutcome["engagement"] = engagement
	}
	
	if improvement, ok := data["improvement"].(float64); ok {
		expectedOutcome["improvement"] = improvement
	}
	
	// mapJSON?
	if len(expectedOutcome) > 0 {
		if jsonBytes, err := json.Marshal(expectedOutcome); err == nil {
			metrics.ExpectedOutcome = string(jsonBytes)
		}
	}
}

// updateEmotionalState ?
func (s *RealtimeLearningAnalyticsService) updateEmotionalState(
	emotional *domainServices.EmotionalState,
	data map[string]interface{},
) {
	if mood, ok := data["mood"].(string); ok {
		emotional.Mood = mood
	}
	
	if stress, ok := data["stress"].(float64); ok {
		emotional.Stress = stress
	}
	
	if motivation, ok := data["motivation"].(float64); ok {
		emotional.Motivation = motivation
	}
	
	if confidence, ok := data["confidence"].(float64); ok {
		emotional.Confidence = confidence
	}
	
	if engagement, ok := data["engagement"].(float64); ok {
		emotional.Engagement = engagement
	}
	
	// valencearousaldominancefrustrationsatisfaction
	// EmotionalState洢?
}

// generatePerformanceInsights 
func (s *RealtimeLearningAnalyticsService) generatePerformanceInsights(
	state *RealtimeLearningState,
) []*LearningInsight {
	insights := make([]*LearningInsight, 0)
	
	// ?
	if state.PerformanceMetrics.Accuracy < 0.7 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypePerformance,
			Title:       "?,
			Description: fmt.Sprintf("%.2f?, state.PerformanceMetrics.Accuracy),
			Impact:      ImpactLevelHigh,
			Confidence:  0.8,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"accuracy": state.PerformanceMetrics.Accuracy, "importance": 0.9, "actionable": true, "category": "performance"},
		}
		insights = append(insights, insight)
	}
	
	// 
	if state.PerformanceMetrics.Speed < 0.5 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypePerformance,
			Title:       "",
			Description: fmt.Sprintf("?.2f?, state.PerformanceMetrics.Speed),
			Impact:      ImpactLevelMedium,
			Confidence:  0.7,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"speed": state.PerformanceMetrics.Speed, "importance": 0.8, "actionable": true, "category": "performance"},
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// generateEngagementInsights ?
func (s *RealtimeLearningAnalyticsService) generateEngagementInsights(
	state *RealtimeLearningState,
) []*LearningInsight {
	insights := make([]*LearningInsight, 0)
	
	// ?
	if state.PerformanceMetrics.Engagement < 0.6 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypeEngagement,
			Title:       "?,
			Description: fmt.Sprintf("%.2f?, state.PerformanceMetrics.Engagement),
			Impact:      ImpactLevelMedium,
			Confidence:  0.7,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"engagement": state.PerformanceMetrics.Engagement, "importance": 0.8, "actionable": true, "category": "engagement"},
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// generateBehaviorInsights 
func (s *RealtimeLearningAnalyticsService) generateBehaviorInsights(
	state *RealtimeLearningState,
) []*LearningInsight {
	insights := make([]*LearningInsight, 0)
	
	// ?
	if state.PerformanceMetrics.Consistency < 0.7 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypeBehavior,
			Title:       "?,
			Description: fmt.Sprintf("%.2f齨", state.PerformanceMetrics.Consistency),
			Impact:      ImpactLevelMedium,
			Confidence:  0.7,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"consistency": state.PerformanceMetrics.Consistency, "importance": 0.8, "actionable": true, "category": "behavior"},
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// generateEmotionalInsights 
func (s *RealtimeLearningAnalyticsService) generateEmotionalInsights(
	state *RealtimeLearningState,
) []*LearningInsight {
	insights := make([]*LearningInsight, 0)
	
	// ?
	if state.FatigueLevel > 0.7 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypeBehavior,
			Title:       "?,
			Description: fmt.Sprintf("%.2f", state.FatigueLevel),
			Impact:      ImpactLevelHigh,
			Confidence:  0.8,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"fatigue": state.FatigueLevel, "importance": 0.9, "actionable": true, "category": "behavior"},
		}
		insights = append(insights, insight)
	}
	
	// 
	if state.MotivationLevel < 0.5 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypeBehavior,
			Title:       "",
			Description: fmt.Sprintf("?.2f?, state.MotivationLevel),
			Impact:      ImpactLevelHigh,
			Confidence:  0.8,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"motivation": state.MotivationLevel, "importance": 0.9, "actionable": true, "category": "behavior"},
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// identifyTimePattern 
func (s *RealtimeLearningAnalyticsService) identifyTimePattern(
	state *RealtimeLearningState,
) *LearningPattern {
	// 
	pattern := &LearningPattern{
		PatternID:    uuid.New(),
		LearnerID:    state.LearnerID,
		Type:         LearningPatternTypeSequential,
		Strength:     0.8,
		Frequency:    1.0,
		Stability:    0.7,
		Adaptability: 0.6,
		Effectiveness: 0.7,
		LastUpdated:  time.Now(),
		Metadata: map[string]interface{}{
			"name":        "",
			"description": "?,
			"confidence":  0.7,
		},
	}
	
	return pattern
}

// identifyContentPattern 
func (s *RealtimeLearningAnalyticsService) identifyContentPattern(
	state *RealtimeLearningState,
) *LearningPattern {
	pattern := &LearningPattern{
		PatternID:     uuid.New(),
		LearnerID:     state.LearnerID,
		Type:          LearningPatternTypeStrategic,
		Strength:      0.7,
		Frequency:     0.8,
		Stability:     0.6,
		Adaptability:  0.7,
		Effectiveness: 0.6,
		LastUpdated:   time.Now(),
		Metadata: map[string]interface{}{
			"name":        "",
			"description": "?,
			"confidence":  0.6,
		},
	}
	
	return pattern
}

// identifyLearningStylePattern 
func (s *RealtimeLearningAnalyticsService) identifyLearningStylePattern(
	state *RealtimeLearningState,
) *LearningPattern {
	pattern := &LearningPattern{
		PatternID:     uuid.New(),
		LearnerID:     state.LearnerID,
		Type:          LearningPatternTypeDeep,
		Strength:      0.8,
		Frequency:     0.9,
		Stability:     0.8,
		Adaptability:  0.7,
		Effectiveness: 0.8,
		LastUpdated:   time.Now(),
		Metadata: map[string]interface{}{
			"name":        "",
			"description": "?,
			"confidence":  0.7,
		},
	}
	
	return pattern
}

// identifyInteractionPattern 
func (s *RealtimeLearningAnalyticsService) identifyInteractionPattern(
	state *RealtimeLearningState,
) *LearningPattern {
	pattern := &LearningPattern{
		PatternID:     uuid.New(),
		LearnerID:     state.LearnerID,
		Type:          LearningPatternTypeRandom,
		Strength:      0.7,
		Frequency:     0.8,
		Stability:     0.6,
		Adaptability:  0.6,
		Effectiveness: 0.7,
		LastUpdated:   time.Now(),
		Metadata: map[string]interface{}{
			"name":        "",
			"description": "?,
			"confidence":  0.6,
		},
	}
	
	return pattern
}

// detectPerformanceAnomalies 
func (s *RealtimeLearningAnalyticsService) detectPerformanceAnomalies(
	state *RealtimeLearningState,
) []*Anomaly {
	anomalies := make([]*Anomaly, 0)
	
	// 
	if state.PerformanceMetrics.Accuracy < 0.3 {
		anomaly := &Anomaly{
			AnomalyID:   uuid.New(),
			Type:        AnomalyTypeDrop,
			Severity:    0.9, // ?
			Description: "?,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"metric":    "accuracy",
				"value":     state.PerformanceMetrics.Accuracy,
				"threshold": 0.5,
				"confidence": 0.9,
			},
		}
		anomalies = append(anomalies, anomaly)
	}
	
	return anomalies
}

// detectBehaviorAnomalies ?
func (s *RealtimeLearningAnalyticsService) detectBehaviorAnomalies(
	state *RealtimeLearningState,
) []*Anomaly {
	anomalies := make([]*Anomaly, 0)
	
	// ?
	if state.CurrentSession != nil {
		sessionDuration := time.Since(state.CurrentSession.StartTime)
		if sessionDuration > 4*time.Hour {
			anomaly := &Anomaly{
				AnomalyID:   uuid.New(),
				Type:        AnomalyTypeSpike,
				Severity:    0.7, // 
				Description: "",
				Timestamp:   time.Now(),
				Metadata: map[string]interface{}{
					"metric":     "session_duration",
					"value":      sessionDuration.Hours(),
					"threshold":  3.0,
					"confidence": 0.8,
				},
			}
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies
}

// detectEngagementAnomalies 
func (s *RealtimeLearningAnalyticsService) detectEngagementAnomalies(
	state *RealtimeLearningState,
) []*Anomaly {
	anomalies := make([]*Anomaly, 0)
	
	// 
	if state.PerformanceMetrics.Engagement < 0.2 {
		anomaly := &Anomaly{
			AnomalyID:   uuid.New(),
			Type:        AnomalyTypeDrop,
			Severity:    0.9, // ?
			Description: "?,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"metric":     "engagement",
				"value":      state.PerformanceMetrics.Engagement,
				"threshold":  0.4,
				"confidence": 0.8,
			},
		}
		anomalies = append(anomalies, anomaly)
	}
	
	return anomalies
}

// analyzePerformanceTrend 
func (s *RealtimeLearningAnalyticsService) analyzePerformanceTrend(
	state *RealtimeLearningState,
) *Trend {
	trend := &Trend{
		Direction:   TrendFlat,
		Strength:    0.6,
		Confidence:  0.7,
		StartTime:   state.Timestamp,
		EndTime:     time.Now(),
		Description: "",
	}
	
	return trend
}

// analyzeEngagementTrend ?
func (s *RealtimeLearningAnalyticsService) analyzeEngagementTrend(
	state *RealtimeLearningState,
) *Trend {
	trend := &Trend{
		Direction:   TrendFlat,
		Strength:    0.5,
		Confidence:  0.6,
		StartTime:   state.Timestamp,
		EndTime:     time.Now(),
		Description: "?,
	}
	
	return trend
}

// analyzeLearningSpeedTrend 
func (s *RealtimeLearningAnalyticsService) analyzeLearningSpeedTrend(
	state *RealtimeLearningState,
) *Trend {
	trend := &Trend{
		Direction:   TrendUp,
		Strength:    0.7,
		Confidence:  0.8,
		StartTime:   state.Timestamp,
		EndTime:     time.Now(),
		Description: "",
	}
	
	return trend
}

// calculateOverallConfidence ?
func (s *RealtimeLearningAnalyticsService) calculateOverallConfidence(
	insights []*LearningInsight,
	patterns []*LearningPattern,
	anomalies []*Anomaly,
	trends []*Trend,
) float64 {
	totalConfidence := 0.0
	count := 0
	
	for _, insight := range insights {
		totalConfidence += insight.Confidence
		count++
	}
	
	for _, pattern := range patterns {
		if confidence, ok := pattern.Metadata["confidence"].(float64); ok {
			totalConfidence += confidence
		} else {
			totalConfidence += 0.5 // ?
		}
		count++
	}
	
	for _, anomaly := range anomalies {
		if confidence, ok := anomaly.Metadata["confidence"].(float64); ok {
			totalConfidence += confidence
		} else {
			totalConfidence += 0.5 // ?
		}
		count++
	}
	
	for _, trend := range trends {
		totalConfidence += trend.Confidence
		count++
	}
	
	if count == 0 {
		return 0.5 // ?
	}
	
	return totalConfidence / float64(count)
}

// assessAnalysisQuality 
func (s *RealtimeLearningAnalyticsService) assessAnalysisQuality(
	insights []*LearningInsight,
	patterns []*LearningPattern,
	anomalies []*Anomaly,
	trends []*Trend,
	recommendations []*AnalysisRecommendation,
) *AnalysisQuality {
	quality := &AnalysisQuality{
		QualityID:    uuid.New(),
		Completeness: s.calculateCompleteness(insights, patterns, anomalies, trends),
		Accuracy:     s.calculateAccuracy(insights, patterns),
		Reliability:  s.calculateReliability(insights, patterns, anomalies),
		Validity:     s.calculateValidity(insights, patterns),
		Confidence:   s.calculateOverallConfidence(insights, patterns, anomalies, trends),
		Timeliness:   1.0, // ?
		Issues:       make([]string, 0),
		Suggestions:  make([]string, 0),
		Timestamp:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}
	
	quality.Score = (quality.Completeness + quality.Accuracy + quality.Reliability + 
		quality.Validity + quality.Confidence + quality.Timeliness) / 6.0
	
	return quality
}

// calculateCompleteness ?
func (s *RealtimeLearningAnalyticsService) calculateCompleteness(
	insights []*LearningInsight,
	patterns []*LearningPattern,
	anomalies []*Anomaly,
	trends []*Trend,
) float64 {
	expectedComponents := 4.0 // ?
	actualComponents := 0.0
	
	if len(insights) > 0 {
		actualComponents++
	}
	if len(patterns) > 0 {
		actualComponents++
	}
	if len(anomalies) > 0 {
		actualComponents++
	}
	if len(trends) > 0 {
		actualComponents++
	}
	
	return actualComponents / expectedComponents
}

// calculateAccuracy ?
func (s *RealtimeLearningAnalyticsService) calculateAccuracy(
	insights []*LearningInsight,
	patterns []*LearningPattern,
) float64 {
	totalConfidence := 0.0
	count := 0
	
	for _, insight := range insights {
		totalConfidence += insight.Confidence
		count++
	}
	
	for _, pattern := range patterns {
		if confidence, ok := pattern.Metadata["confidence"].(float64); ok {
			totalConfidence += confidence
		} else {
			totalConfidence += 0.5 // ?
		}
		count++
	}
	
	if count == 0 {
		return 0.5
	}
	
	return totalConfidence / float64(count)
}

// calculateRelevance ?
func (s *RealtimeLearningAnalyticsService) calculateRelevance(
	insights []*LearningInsight,
	recommendations []*AnalysisRecommendation,
) float64 {
	if len(insights) == 0 {
		return 0.5
	}
	
	totalImportance := 0.0
	for _, insight := range insights {
		if importance, ok := insight.Metadata["importance"].(float64); ok {
			totalImportance += importance
		} else {
			totalImportance += 0.5 // ?
		}
	}
	
	return math.Min(totalImportance/float64(len(insights)), 1.0)
}

// calculateConsistency ?
func (s *RealtimeLearningAnalyticsService) calculateConsistency(
	insights []*LearningInsight,
	patterns []*LearningPattern,
) float64 {
	// ?
	return 0.8
}

// calculateReliability ?
func (s *RealtimeLearningAnalyticsService) calculateReliability(
	insights []*LearningInsight,
	patterns []*LearningPattern,
	anomalies []*Anomaly,
) float64 {
	// ?
	return 0.7
}

// calculateValidity ?
func (s *RealtimeLearningAnalyticsService) calculateValidity(
	insights []*LearningInsight,
	patterns []*LearningPattern,
) float64 {
	// ?
	return 0.8
}

// calculateClarity ?
func (s *RealtimeLearningAnalyticsService) calculateClarity(
	insights []*LearningInsight,
	recommendations []*AnalysisRecommendation,
) float64 {
	// 
	return 0.8
}

// calculateActionability ?
func (s *RealtimeLearningAnalyticsService) calculateActionability(
	insights []*LearningInsight,
	recommendations []*AnalysisRecommendation,
) float64 {
	if len(insights) == 0 {
		return 0.0
	}
	
	actionableCount := 0
	for _, insight := range insights {
		if actionable, ok := insight.Metadata["actionable"].(bool); ok && actionable {
			actionableCount++
		}
	}
	
	return float64(actionableCount) / float64(len(insights))
}

// updateAnalysisMetrics 
func (s *RealtimeLearningAnalyticsService) updateAnalysisMetrics(result *AnalysisResult) {
	s.metrics.TotalAnalyses++
	
	// ?
	accuracy := 0.5 // ?
	if result.Quality != nil {
		accuracy = result.Quality.Accuracy
	}
	
	s.metrics.AverageAccuracy = (s.metrics.AverageAccuracy*float64(s.metrics.TotalAnalyses-1) + 
		accuracy) / float64(s.metrics.TotalAnalyses)
	
	// 
	totalTime := int64(s.metrics.AverageProcessingTime) * int64(s.metrics.TotalAnalyses-1)
	s.metrics.AverageProcessingTime = int64((totalTime + int64(result.Duration)) / int64(s.metrics.TotalAnalyses))
	s.metrics.LastAnalysisTime = time.Now()
}

// prioritizeRecommendations ?
func (s *RealtimeLearningAnalyticsService) prioritizeRecommendations(
	recommendations []*AnalysisRecommendation,
) []*AnalysisRecommendation {
	sort.Slice(recommendations, func(i, j int) bool {
		// 
		if recommendations[i].Priority != recommendations[j].Priority {
			return recommendations[i].Priority > recommendations[j].Priority
		}
		
		// 
		return recommendations[i].Confidence > recommendations[j].Confidence
	})
	
	// 
	maxRecommendations := 10
	if len(recommendations) > maxRecommendations {
		return recommendations[:maxRecommendations]
	}
	
	return recommendations
}

// generateInsightBasedRecommendation 
func (s *RealtimeLearningAnalyticsService) generateInsightBasedRecommendation(
	insight *LearningInsight,
	state *RealtimeLearningState,
) *AnalysisRecommendation {
	// ?metadata ?
	importance := 0.5
	if imp, ok := insight.Metadata["importance"].(float64); ok {
		importance = imp
	}
	
	confidence := 0.5
	if conf, ok := insight.Metadata["confidence"].(float64); ok {
		confidence = conf
	}
	
	recommendation := &AnalysisRecommendation{
		RecommendationID: uuid.New(),
		Type:             "improvement",
		Category:         "immediate",
		Title:            "?,
		Description:      fmt.Sprintf("%s?, insight.Title),
		Priority:         int(importance * 10),
		Confidence:       confidence,
		ExpectedImpact:   importance,
		Status:           RecommendationStatusPending,
		Metadata:         map[string]interface{}{"insight_id": insight.InsightID},
	}
	
	return recommendation
}

// updateRealtimePerformanceMetrics 
func (s *RealtimeLearningAnalyticsService) updateRealtimePerformanceMetrics(
	metrics *RealtimePerformanceMetrics,
	data map[string]interface{},
) {
	if accuracy, ok := data["accuracy"].(float64); ok {
		metrics.Accuracy = accuracy
	}
	
	if speed, ok := data["speed"].(float64); ok {
		metrics.Speed = speed
	}
	
	if efficiency, ok := data["efficiency"].(float64); ok {
		metrics.Efficiency = efficiency
	}
	
	if retention, ok := data["retention"].(float64); ok {
		metrics.Retention = retention
	}
	
	if engagement, ok := data["engagement"].(float64); ok {
		metrics.Engagement = engagement
	}
	
	if satisfaction, ok := data["satisfaction"].(float64); ok {
		metrics.Satisfaction = satisfaction
	}
	
	if progress, ok := data["progress"].(float64); ok {
		metrics.Progress = progress
	}
	
	if mastery, ok := data["mastery"].(float64); ok {
		metrics.Mastery = mastery
	}
	
	if consistency, ok := data["consistency"].(float64); ok {
		metrics.Consistency = consistency
	}
	
	if improvement, ok := data["improvement"].(float64); ok {
		metrics.Improvement = improvement
	}
	
	metrics.LastUpdated = time.Now()
}

// convertPatternRecommendation 
func (s *RealtimeLearningAnalyticsService) convertPatternRecommendation(
	patternRec *PatternRecommendation,
	state *RealtimeLearningState,
) *AnalysisRecommendation {
	// ?PriorityLevel ?int
	var priority int
	switch patternRec.Priority {
	case PriorityLevelHigh:
		priority = 1
	case PriorityLevelMedium:
		priority = 2
	case PriorityLevelLow:
		priority = 3
	default:
		priority = 2
	}

	recommendation := &AnalysisRecommendation{
		RecommendationID: uuid.New(),
		Type:             "optimization",
		Category:         "short_term",
		Title:            "?,
		Description:      patternRec.Description,
		Priority:         priority,
		Confidence:       patternRec.Confidence,
		ExpectedImpact:   0.7, //  PatternRecommendation ?
		Status:           RecommendationStatusPending,
		Metadata:         make(map[string]interface{}),
	}
	
	return recommendation
}

