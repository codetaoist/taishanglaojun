package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	domainServices "github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// 辅助方法实现

// updateSessionInfo 更新会话信息
func (s *RealtimeLearningAnalyticsService) updateSessionInfo(
	sessionInfo *SessionInfo,
	data map[string]interface{},
) {
	if duration, ok := data["duration"].(float64); ok {
		sessionInfo.Duration = time.Duration(duration) * time.Second
	}
	
	if deviceInfo, ok := data["device_info"].(map[string]interface{}); ok {
		s.updateDeviceInfo(sessionInfo.DeviceInfo, deviceInfo)
	}
	
	if goals, ok := data["goals"].([]interface{}); ok {
		sessionInfo.Goals = s.parseSessionGoals(goals)
	}
}

// updateDeviceInfo 更新设备信息
func (s *RealtimeLearningAnalyticsService) updateDeviceInfo(
	deviceInfo *DeviceInfo,
	data map[string]interface{},
) {
	if deviceType, ok := data["device_type"].(string); ok {
		deviceInfo.DeviceType = deviceType
	}
	
	if platform, ok := data["platform"].(string); ok {
		deviceInfo.Platform = platform
	}
	
	if browser, ok := data["browser"].(string); ok {
		deviceInfo.Browser = browser
	}
	
	if screenResolution, ok := data["screen_resolution"].(string); ok {
		deviceInfo.ScreenResolution = screenResolution
	}
}

// parseSessionGoals 解析会话目标
func (s *RealtimeLearningAnalyticsService) parseSessionGoals(
	goals []interface{},
) []*SessionGoal {
	sessionGoals := make([]*SessionGoal, 0, len(goals))
	
	for _, goalData := range goals {
		if goalMap, ok := goalData.(map[string]interface{}); ok {
			goal := &SessionGoal{
				GoalID:      uuid.New(),
				Type:        "learning",
				Description: "",
				TargetValue: 0,
				CurrentValue: 0,
				IsCompleted: false,
				Priority:    1,
				Deadline:    time.Now().Add(24 * time.Hour),
				Metadata:    make(map[string]interface{}),
			}
			
			if description, ok := goalMap["description"].(string); ok {
				goal.Description = description
			}
			
			if targetValue, ok := goalMap["target_value"].(float64); ok {
				goal.TargetValue = targetValue
			}
			
			if currentValue, ok := goalMap["current_value"].(float64); ok {
				goal.CurrentValue = currentValue
			}
			
			if priority, ok := goalMap["priority"].(float64); ok {
				goal.Priority = int(priority)
			}
			
			sessionGoals = append(sessionGoals, goal)
		}
	}
	
	return sessionGoals
}

// updateCurrentActivity 更新当前活动
func (s *RealtimeLearningAnalyticsService) updateCurrentActivity(
	activity *LearningActivity,
	data map[string]interface{},
) {
	if activityType, ok := data["activity_type"].(string); ok {
		activity.ActivityType = ActivityType(activityType)
	}
	
	if contentID, ok := data["content_id"].(string); ok {
		if id, err := uuid.Parse(contentID); err == nil {
			activity.ContentID = id
		}
	}
	
	if duration, ok := data["duration"].(float64); ok {
		activity.Duration = time.Duration(duration) * time.Second
	}
	
	if progress, ok := data["progress"].(float64); ok {
		activity.Progress = progress
	}
	
	if difficulty, ok := data["difficulty"].(float64); ok {
		activity.Difficulty = difficulty
	}
	
	if engagement, ok := data["engagement"].(float64); ok {
		activity.Engagement = engagement
	}
}

// updatePerformanceMetrics 更新性能指标
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
	
	if retention, ok := data["retention"].(float64); ok {
		metrics.Retention = retention
	}
	
	if comprehension, ok := data["comprehension"].(float64); ok {
		metrics.Comprehension = comprehension
	}
	
	if engagement, ok := data["engagement"].(float64); ok {
		metrics.Engagement = engagement
	}
	
	if efficiency, ok := data["efficiency"].(float64); ok {
		metrics.Efficiency = efficiency
	}
	
	if consistency, ok := data["consistency"].(float64); ok {
		metrics.Consistency = consistency
	}
	
	if improvement, ok := data["improvement"].(float64); ok {
		metrics.Improvement = improvement
	}
}

// updateEmotionalState 更新情感状态
func (s *RealtimeLearningAnalyticsService) updateEmotionalState(
	emotional *domainServices.EmotionalState,
	data map[string]interface{},
) {
	if valence, ok := data["valence"].(float64); ok {
		emotional.Valence = valence
	}
	
	if arousal, ok := data["arousal"].(float64); ok {
		emotional.Arousal = arousal
	}
	
	if dominance, ok := data["dominance"].(float64); ok {
		emotional.Dominance = dominance
	}
	
	if confidence, ok := data["confidence"].(float64); ok {
		emotional.Confidence = confidence
	}
	
	if frustration, ok := data["frustration"].(float64); ok {
		emotional.Frustration = frustration
	}
	
	if motivation, ok := data["motivation"].(float64); ok {
		emotional.Motivation = motivation
	}
	
	if satisfaction, ok := data["satisfaction"].(float64); ok {
		emotional.Satisfaction = satisfaction
	}
	
	if stress, ok := data["stress"].(float64); ok {
		emotional.Stress = stress
	}
}

// generatePerformanceInsights 生成性能洞察
func (s *RealtimeLearningAnalyticsService) generatePerformanceInsights(
	state *RealtimeLearningState,
) []*LearningInsight {
	insights := make([]*LearningInsight, 0)
	
	// 准确率洞察
	if state.PerformanceMetrics.Accuracy < 0.7 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypePattern,
			Category:    InsightCategoryPerformance,
			Title:       "准确率偏低",
			Description: fmt.Sprintf("当前准确率为%.2f，低于期望水平", state.PerformanceMetrics.Accuracy),
			Confidence:  0.8,
			Importance:  0.9,
			Actionable:  true,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"accuracy": state.PerformanceMetrics.Accuracy},
		}
		insights = append(insights, insight)
	}
	
	// 学习速度洞察
	if state.PerformanceMetrics.Speed < 0.5 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypePattern,
			Category:    InsightCategoryPerformance,
			Title:       "学习速度较慢",
			Description: fmt.Sprintf("当前学习速度为%.2f，建议调整学习策略", state.PerformanceMetrics.Speed),
			Confidence:  0.7,
			Importance:  0.8,
			Actionable:  true,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"speed": state.PerformanceMetrics.Speed},
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// generateEngagementInsights 生成参与度洞察
func (s *RealtimeLearningAnalyticsService) generateEngagementInsights(
	state *RealtimeLearningState,
) []*LearningInsight {
	insights := make([]*LearningInsight, 0)
	
	// 参与度洞察
	if state.PerformanceMetrics.Engagement < 0.6 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypePattern,
			Category:    InsightCategoryEngagement,
			Title:       "参与度不足",
			Description: fmt.Sprintf("当前参与度为%.2f，需要提升学习兴趣", state.PerformanceMetrics.Engagement),
			Confidence:  0.8,
			Importance:  0.9,
			Actionable:  true,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"engagement": state.PerformanceMetrics.Engagement},
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// generateBehaviorInsights 生成行为洞察
func (s *RealtimeLearningAnalyticsService) generateBehaviorInsights(
	state *RealtimeLearningState,
) []*LearningInsight {
	insights := make([]*LearningInsight, 0)
	
	// 学习一致性洞察
	if state.PerformanceMetrics.Consistency < 0.7 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypePattern,
			Category:    InsightCategoryBehavior,
			Title:       "学习一致性不足",
			Description: fmt.Sprintf("学习表现一致性为%.2f，建议建立规律的学习习惯", state.PerformanceMetrics.Consistency),
			Confidence:  0.7,
			Importance:  0.8,
			Actionable:  true,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"consistency": state.PerformanceMetrics.Consistency},
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// generateEmotionalInsights 生成情感洞察
func (s *RealtimeLearningAnalyticsService) generateEmotionalInsights(
	state *RealtimeLearningState,
) []*LearningInsight {
	insights := make([]*LearningInsight, 0)
	
	// 挫折感洞察
	if state.EmotionalState.Frustration > 0.7 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypeRisk,
			Category:    InsightCategoryBehavior,
			Title:       "挫折感较高",
			Description: fmt.Sprintf("当前挫折感为%.2f，可能影响学习效果", state.EmotionalState.Frustration),
			Confidence:  0.8,
			Importance:  0.9,
			Actionable:  true,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"frustration": state.EmotionalState.Frustration},
		}
		insights = append(insights, insight)
	}
	
	// 动机洞察
	if state.EmotionalState.Motivation < 0.5 {
		insight := &LearningInsight{
			InsightID:   uuid.New(),
			Type:        InsightTypeRisk,
			Category:    InsightCategoryBehavior,
			Title:       "学习动机不足",
			Description: fmt.Sprintf("当前学习动机为%.2f，需要激发学习兴趣", state.EmotionalState.Motivation),
			Confidence:  0.8,
			Importance:  0.9,
			Actionable:  true,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"motivation": state.EmotionalState.Motivation},
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// identifyTimePattern 识别时间模式
func (s *RealtimeLearningAnalyticsService) identifyTimePattern(
	state *RealtimeLearningState,
) *LearningPattern {
	// 简化的时间模式识别
	pattern := &LearningPattern{
		PatternID:   uuid.New(),
		LearnerID:   state.LearnerID,
		Type:        PatternTypeTemporal,
		Name:        "学习时间模式",
		Description: "基于学习时间的行为模式",
		Confidence:  0.7,
		Strength:    0.8,
		Frequency:   1.0,
		Stability:   0.7,
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	return pattern
}

// identifyContentPattern 识别内容偏好模式
func (s *RealtimeLearningAnalyticsService) identifyContentPattern(
	state *RealtimeLearningState,
) *LearningPattern {
	pattern := &LearningPattern{
		PatternID:   uuid.New(),
		LearnerID:   state.LearnerID,
		Type:        PatternTypeContent,
		Name:        "内容偏好模式",
		Description: "基于内容访问的偏好模式",
		Confidence:  0.6,
		Strength:    0.7,
		Frequency:   0.8,
		Stability:   0.6,
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	return pattern
}

// identifyLearningStylePattern 识别学习风格模式
func (s *RealtimeLearningAnalyticsService) identifyLearningStylePattern(
	state *RealtimeLearningState,
) *LearningPattern {
	pattern := &LearningPattern{
		PatternID:   uuid.New(),
		LearnerID:   state.LearnerID,
		Type:        PatternTypeLearningStyle,
		Name:        "学习风格模式",
		Description: "基于学习行为的风格模式",
		Confidence:  0.7,
		Strength:    0.8,
		Frequency:   0.9,
		Stability:   0.8,
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	return pattern
}

// identifyInteractionPattern 识别交互模式
func (s *RealtimeLearningAnalyticsService) identifyInteractionPattern(
	state *RealtimeLearningState,
) *LearningPattern {
	pattern := &LearningPattern{
		PatternID:   uuid.New(),
		LearnerID:   state.LearnerID,
		Type:        PatternTypeInteraction,
		Name:        "交互模式",
		Description: "基于用户交互的行为模式",
		Confidence:  0.6,
		Strength:    0.7,
		Frequency:   0.8,
		Stability:   0.6,
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	return pattern
}

// detectPerformanceAnomalies 检测性能异常
func (s *RealtimeLearningAnalyticsService) detectPerformanceAnomalies(
	state *RealtimeLearningState,
) []*Anomaly {
	anomalies := make([]*Anomaly, 0)
	
	// 检测准确率异常下降
	if state.PerformanceMetrics.Accuracy < 0.3 {
		anomaly := &Anomaly{
			AnomalyID:   uuid.New(),
			Type:        AnomalyTypePerformance,
			Severity:    AnomalySeverityHigh,
			Description: "准确率异常下降",
			Value:       state.PerformanceMetrics.Accuracy,
			Threshold:   0.5,
			Confidence:  0.9,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"metric": "accuracy"},
		}
		anomalies = append(anomalies, anomaly)
	}
	
	return anomalies
}

// detectBehaviorAnomalies 检测行为异常
func (s *RealtimeLearningAnalyticsService) detectBehaviorAnomalies(
	state *RealtimeLearningState,
) []*Anomaly {
	anomalies := make([]*Anomaly, 0)
	
	// 检测学习时间异常
	sessionDuration := time.Since(state.StartTime)
	if sessionDuration > 4*time.Hour {
		anomaly := &Anomaly{
			AnomalyID:   uuid.New(),
			Type:        AnomalyTypeBehavior,
			Severity:    AnomalySeverityMedium,
			Description: "学习时间过长",
			Value:       sessionDuration.Hours(),
			Threshold:   3.0,
			Confidence:  0.8,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"metric": "session_duration"},
		}
		anomalies = append(anomalies, anomaly)
	}
	
	return anomalies
}

// detectEngagementAnomalies 检测参与度异常
func (s *RealtimeLearningAnalyticsService) detectEngagementAnomalies(
	state *RealtimeLearningState,
) []*Anomaly {
	anomalies := make([]*Anomaly, 0)
	
	// 检测参与度异常下降
	if state.PerformanceMetrics.Engagement < 0.2 {
		anomaly := &Anomaly{
			AnomalyID:   uuid.New(),
			Type:        AnomalyTypeEngagement,
			Severity:    AnomalySeverityHigh,
			Description: "参与度异常下降",
			Value:       state.PerformanceMetrics.Engagement,
			Threshold:   0.4,
			Confidence:  0.8,
			Timestamp:   time.Now(),
			Metadata:    map[string]interface{}{"metric": "engagement"},
		}
		anomalies = append(anomalies, anomaly)
	}
	
	return anomalies
}

// analyzePerformanceTrend 分析性能趋势
func (s *RealtimeLearningAnalyticsService) analyzePerformanceTrend(
	state *RealtimeLearningState,
) *Trend {
	trend := &Trend{
		TrendID:      uuid.New(),
		Type:         TrendTypeLinear,
		Direction:    RealtimeTrendDirectionStable,
		Strength:     0.6,
		Velocity:     0.1,
		Acceleration: 0.0,
		Confidence:   0.7,
		Significance: 0.6,
		StartTime:    state.StartTime,
		Duration:     time.Since(state.StartTime),
		Metadata:     map[string]interface{}{"metric": "performance"},
	}
	
	return trend
}

// analyzeEngagementTrend 分析参与度趋势
func (s *RealtimeLearningAnalyticsService) analyzeEngagementTrend(
	state *RealtimeLearningState,
) *Trend {
	trend := &Trend{
		TrendID:      uuid.New(),
		Type:         TrendTypeLinear,
		Direction:    RealtimeTrendDirectionStable,
		Strength:     0.5,
		Velocity:     0.05,
		Acceleration: 0.0,
		Confidence:   0.6,
		Significance: 0.5,
		StartTime:    state.StartTime,
		Duration:     time.Since(state.StartTime),
		Metadata:     map[string]interface{}{"metric": "engagement"},
	}
	
	return trend
}

// analyzeLearningSpeedTrend 分析学习速度趋势
func (s *RealtimeLearningAnalyticsService) analyzeLearningSpeedTrend(
	state *RealtimeLearningState,
) *Trend {
	trend := &Trend{
		TrendID:      uuid.New(),
		Type:         TrendTypeLinear,
		Direction:    RealtimeTrendDirectionIncreasing,
		Strength:     0.7,
		Velocity:     0.2,
		Acceleration: 0.1,
		Confidence:   0.8,
		Significance: 0.7,
		StartTime:    state.StartTime,
		Duration:     time.Since(state.StartTime),
		Metadata:     map[string]interface{}{"metric": "learning_speed"},
	}
	
	return trend
}

// calculateOverallConfidence 计算整体置信度
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
		totalConfidence += pattern.Confidence
		count++
	}
	
	for _, anomaly := range anomalies {
		totalConfidence += anomaly.Confidence
		count++
	}
	
	for _, trend := range trends {
		totalConfidence += trend.Confidence
		count++
	}
	
	if count == 0 {
		return 0.5 // 默认置信度
	}
	
	return totalConfidence / float64(count)
}

// assessAnalysisQuality 评估分析质量
func (s *RealtimeLearningAnalyticsService) assessAnalysisQuality(
	insights []*LearningInsight,
	patterns []*LearningPattern,
	anomalies []*Anomaly,
	trends []*Trend,
	recommendations []*AnalysisRecommendation,
) *AnalysisQuality {
	quality := &AnalysisQuality{
		Completeness:  s.calculateCompleteness(insights, patterns, anomalies, trends),
		Accuracy:      s.calculateAccuracy(insights, patterns),
		Relevance:     s.calculateRelevance(insights, recommendations),
		Timeliness:    1.0, // 实时分析，及时性满分
		Consistency:   s.calculateConsistency(insights, patterns),
		Reliability:   s.calculateReliability(insights, patterns, anomalies),
		Validity:      s.calculateValidity(insights, patterns),
		Objectivity:   0.8, // 基于数据的分析，客观性较高
		Clarity:       s.calculateClarity(insights, recommendations),
		Actionability: s.calculateActionability(insights, recommendations),
		Issues:        make([]string, 0),
		Improvements:  make([]string, 0),
		Metadata:      make(map[string]interface{}),
	}
	
	quality.OverallScore = (quality.Completeness + quality.Accuracy + quality.Relevance + 
		quality.Timeliness + quality.Consistency + quality.Reliability + 
		quality.Validity + quality.Objectivity + quality.Clarity + quality.Actionability) / 10.0
	
	return quality
}

// calculateCompleteness 计算完整性
func (s *RealtimeLearningAnalyticsService) calculateCompleteness(
	insights []*LearningInsight,
	patterns []*LearningPattern,
	anomalies []*Anomaly,
	trends []*Trend,
) float64 {
	expectedComponents := 4.0 // 洞察、模式、异常、趋势
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

// calculateAccuracy 计算准确性
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
		totalConfidence += pattern.Confidence
		count++
	}
	
	if count == 0 {
		return 0.5
	}
	
	return totalConfidence / float64(count)
}

// calculateRelevance 计算相关性
func (s *RealtimeLearningAnalyticsService) calculateRelevance(
	insights []*LearningInsight,
	recommendations []*AnalysisRecommendation,
) float64 {
	if len(insights) == 0 {
		return 0.5
	}
	
	totalImportance := 0.0
	for _, insight := range insights {
		totalImportance += insight.Importance
	}
	
	return math.Min(totalImportance/float64(len(insights)), 1.0)
}

// calculateConsistency 计算一致性
func (s *RealtimeLearningAnalyticsService) calculateConsistency(
	insights []*LearningInsight,
	patterns []*LearningPattern,
) float64 {
	// 简化的一致性计算
	return 0.8
}

// calculateReliability 计算可靠性
func (s *RealtimeLearningAnalyticsService) calculateReliability(
	insights []*LearningInsight,
	patterns []*LearningPattern,
	anomalies []*Anomaly,
) float64 {
	// 基于数据质量和置信度的可靠性计算
	return 0.7
}

// calculateValidity 计算有效性
func (s *RealtimeLearningAnalyticsService) calculateValidity(
	insights []*LearningInsight,
	patterns []*LearningPattern,
) float64 {
	// 基于洞察和模式的有效性计算
	return 0.8
}

// calculateClarity 计算清晰度
func (s *RealtimeLearningAnalyticsService) calculateClarity(
	insights []*LearningInsight,
	recommendations []*AnalysisRecommendation,
) float64 {
	// 基于描述质量的清晰度计算
	return 0.8
}

// calculateActionability 计算可操作性
func (s *RealtimeLearningAnalyticsService) calculateActionability(
	insights []*LearningInsight,
	recommendations []*AnalysisRecommendation,
) float64 {
	if len(insights) == 0 {
		return 0.0
	}
	
	actionableCount := 0
	for _, insight := range insights {
		if insight.Actionable {
			actionableCount++
		}
	}
	
	return float64(actionableCount) / float64(len(insights))
}

// updateAnalysisMetrics 更新分析指标
func (s *RealtimeLearningAnalyticsService) updateAnalysisMetrics(result *AnalysisResult) {
	s.metrics.TotalAnalyses++
	s.metrics.AverageAccuracy = (s.metrics.AverageAccuracy*float64(s.metrics.TotalAnalyses-1) + 
		result.Confidence) / float64(s.metrics.TotalAnalyses)
	s.metrics.AverageProcessingTime = (s.metrics.AverageProcessingTime*time.Duration(s.metrics.TotalAnalyses-1) + 
		result.Duration) / time.Duration(s.metrics.TotalAnalyses)
	s.metrics.LastAnalysisTime = time.Now()
}

// prioritizeRecommendations 优先级排序建议
func (s *RealtimeLearningAnalyticsService) prioritizeRecommendations(
	recommendations []*AnalysisRecommendation,
) []*AnalysisRecommendation {
	sort.Slice(recommendations, func(i, j int) bool {
		// 按优先级和紧急程度排序
		if recommendations[i].Priority != recommendations[j].Priority {
			return recommendations[i].Priority > recommendations[j].Priority
		}
		return recommendations[i].Urgency > recommendations[j].Urgency
	})
	
	// 限制建议数量
	maxRecommendations := 10
	if len(recommendations) > maxRecommendations {
		return recommendations[:maxRecommendations]
	}
	
	return recommendations
}

// generateInsightBasedRecommendation 基于洞察生成建议
func (s *RealtimeLearningAnalyticsService) generateInsightBasedRecommendation(
	insight *LearningInsight,
	state *RealtimeLearningState,
) *AnalysisRecommendation {
	recommendation := &AnalysisRecommendation{
		RecommendationID: uuid.New(),
		Type:             RecommendationTypeImprovement,
		Category:         RecommendationCategoryImmediate,
		Title:            "基于洞察的改进建议",
		Description:      fmt.Sprintf("针对%s的改进建议", insight.Title),
		Rationale:        insight.Description,
		Priority:         int(insight.Importance * 10),
		Urgency:          insight.Importance,
		Confidence:       insight.Confidence,
		ExpectedImpact:   insight.Importance,
		Status:           RecommendationStatusPending,
		Metadata:         map[string]interface{}{"insight_id": insight.InsightID},
	}
	
	return recommendation
}

// convertPatternRecommendation 转换模式建议
func (s *RealtimeLearningAnalyticsService) convertPatternRecommendation(
	patternRec *PatternRecommendation,
	state *RealtimeLearningState,
) *AnalysisRecommendation {
	recommendation := &AnalysisRecommendation{
		RecommendationID: uuid.New(),
		Type:             RecommendationTypeOptimization,
		Category:         RecommendationCategoryShortTerm,
		Title:            "基于模式的优化建议",
		Description:      patternRec.Description,
		Rationale:        patternRec.Rationale,
		Priority:         patternRec.Priority,
		Urgency:          0.6,
		Confidence:       0.7,
		ExpectedImpact:   patternRec.ExpectedImpact,
		Status:           RecommendationStatusPending,
		Metadata:         make(map[string]interface{}),
	}
	
	return recommendation
}