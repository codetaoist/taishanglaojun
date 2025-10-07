package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// PreferenceAnalyzer 偏好分析器
type PreferenceAnalyzer struct {
	behaviorTracker *UserBehaviorTracker
	contentRepo     ContentRepository
	userRepo        UserRepository
}

// PreferenceRepository 偏好数据仓库接口
type PreferenceRepository interface {
	SaveUserPreferences(ctx context.Context, userID string, preferences *UserPreferences) error
	GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error)
	GetContentPreferences(ctx context.Context, contentID string) (*ContentPreferences, error)
	SavePreferenceHistory(ctx context.Context, history *PreferenceHistory) error
}

// ContentRepository 内容仓库接口
type ContentRepository interface {
	GetContentByID(ctx context.Context, contentID string) (*Content, error)
	GetContentsByCategory(ctx context.Context, category string) ([]*Content, error)
	GetContentsByDifficulty(ctx context.Context, difficulty string) ([]*Content, error)
	GetContentsByTags(ctx context.Context, tags []string) ([]*Content, error)
}

// UserRepository 用户仓库接口
type UserRepository interface {
	GetUserProfile(ctx context.Context, userID string) (*UserProfile, error)
	GetUserLearningHistory(ctx context.Context, userID string) ([]*LearningRecord, error)
	GetUserInteractions(ctx context.Context, userID string, limit int) ([]*UserInteraction, error)
}

// NewPreferenceAnalyzer 创建偏好分析器
func NewPreferenceAnalyzer(
	behaviorTracker *UserBehaviorTracker,
	contentRepo ContentRepository,
	userRepo UserRepository,
) *PreferenceAnalyzer {
	return &PreferenceAnalyzer{
		behaviorTracker: behaviorTracker,
		contentRepo:     contentRepo,
		userRepo:        userRepo,
	}
}

// UserPreferences 用户偏好
type UserPreferences struct {
	UserID              string                    `json:"user_id"`
	ContentPreferences  *ContentPreferenceProfile `json:"content_preferences"`
	LearningPreferences *LearningPreferenceProfile `json:"learning_preferences"`
	InteractionPatterns *InteractionPatterns      `json:"interaction_patterns"`
	TopicInterests      map[string]float64        `json:"topic_interests"`
	DifficultyPreference *DifficultyPreference    `json:"difficulty_preference"`
	TimePreferences     *TimePreferences          `json:"time_preferences"`
	DevicePreferences   *DevicePreferences        `json:"device_preferences"`
	LanguagePreferences []string                  `json:"language_preferences"`
	UpdatedAt           time.Time                 `json:"updated_at"`
	Confidence          float64                   `json:"confidence"`
}

// ContentPreferenceProfile 内容偏好配置
type ContentPreferenceProfile struct {
	PreferredFormats    map[string]float64 `json:"preferred_formats"`    // 视频、文本、音频等
	PreferredDuration   *DurationRange     `json:"preferred_duration"`   // 偏好的内容时长
	PreferredComplexity string             `json:"preferred_complexity"` // 简单、中等、复杂
	PreferredTopics     map[string]float64 `json:"preferred_topics"`     // 主题偏好权重
	AvoidedTopics       []string           `json:"avoided_topics"`       // 避免的主题
	QualityThreshold    float64            `json:"quality_threshold"`    // 质量阈值
}

// LearningPreferenceProfile 学习偏好配置
type LearningPreferenceProfile struct {
	LearningStyle       string             `json:"learning_style"`       // 视觉、听觉、动手等
	PacePreference      string             `json:"pace_preference"`      // 快速、正常、慢速
	StructurePreference string             `json:"structure_preference"` // 结构化、自由式
	FeedbackPreference  string             `json:"feedback_preference"`  // 即时、延迟、无
	SocialPreference    string             `json:"social_preference"`    // 独立、协作、混合
	ChallengeLevel      float64            `json:"challenge_level"`      // 挑战程度偏好 0-1
	RepetitionTolerance float64            `json:"repetition_tolerance"` // 重复容忍度 0-1
	ExplorationTendency float64            `json:"exploration_tendency"` // 探索倾向 0-1
}

// InteractionPatterns 交互模式
type InteractionPatterns struct {
	ClickPatterns       map[string]float64 `json:"click_patterns"`       // 点击模式
	ScrollPatterns      map[string]float64 `json:"scroll_patterns"`      // 滚动模式
	SearchPatterns      []string           `json:"search_patterns"`      // 搜索模式
	NavigationPatterns  map[string]float64 `json:"navigation_patterns"`  // 导航模式
	CompletionPatterns  map[string]float64 `json:"completion_patterns"`  // 完成模式
	EngagementPatterns  map[string]float64 `json:"engagement_patterns"`  // 参与模式
	AttentionSpan       time.Duration      `json:"attention_span"`       // 注意力持续时间
	SessionDuration     time.Duration      `json:"session_duration"`     // 会话持续时间
}

// DifficultyPreference 难度偏好
type DifficultyPreference struct {
	PreferredLevel    string             `json:"preferred_level"`    // 偏好难度级别
	AdaptationRate    float64            `json:"adaptation_rate"`    // 适应速度
	ChallengeSeeker   bool               `json:"challenge_seeker"`   // 是否寻求挑战
	ComfortZone       *DifficultyRange   `json:"comfort_zone"`       // 舒适区间
	ProgressionSpeed  float64            `json:"progression_speed"`  // 进步速度
	DifficultyHistory map[string]float64 `json:"difficulty_history"` // 历史难度表现
}

// TimePreferences 时间偏好
type TimePreferences struct {
	PreferredTimes    []PreferenceTimeSlot        `json:"preferred_times"`    // 偏好时间段
	SessionLength     time.Duration     `json:"session_length"`     // 会话长度
	BreakFrequency    time.Duration     `json:"break_frequency"`    // 休息频率
	WeeklyPattern     map[string]float64 `json:"weekly_pattern"`     // 周模式
	SeasonalPattern   map[string]float64 `json:"seasonal_pattern"`   // 季节模式
	TimeZonePreference string           `json:"timezone_preference"` // 时区偏好
}

// DevicePreferences 设备偏好
type DevicePreferences struct {
	PreferredDevices map[string]float64 `json:"preferred_devices"` // 偏好设备
	ScreenSize       string             `json:"screen_size"`       // 屏幕尺寸偏好
	InputMethod      string             `json:"input_method"`      // 输入方式偏好
	ConnectivityType string             `json:"connectivity_type"` // 连接类型偏好
	PerformanceLevel string             `json:"performance_level"` // 性能级别偏好
}

// DurationRange 时长范围
type DurationRange struct {
	Min time.Duration `json:"min"`
	Max time.Duration `json:"max"`
}

// DifficultyRange 难度范围
type DifficultyRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// TimeSlot 时间段
type PreferenceTimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Weight float64  `json:"weight"`
}

// ContentPreferences 内容偏好
type ContentPreferences struct {
	ContentID       string             `json:"content_id"`
	UserRatings     map[string]float64 `json:"user_ratings"`     // 用户评分
	ViewCounts      int64              `json:"view_counts"`      // 观看次数
	CompletionRates map[string]float64 `json:"completion_rates"` // 完成率
	EngagementScore float64            `json:"engagement_score"` // 参与度分数
	PopularityScore float64            `json:"popularity_score"` // 流行度分数
	QualityScore    float64            `json:"quality_score"`    // 质量分数
}

// PreferenceHistory 偏好历史
type PreferenceHistory struct {
	UserID      string                 `json:"user_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Action      string                 `json:"action"`
	ContentID   string                 `json:"content_id"`
	Context     map[string]interface{} `json:"context"`
	Preferences map[string]float64     `json:"preferences"`
	Confidence  float64                `json:"confidence"`
}

// Content 内容结构
type Content struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Difficulty  string            `json:"difficulty"`
	Duration    time.Duration     `json:"duration"`
	Format      string            `json:"format"`
	Language    string            `json:"language"`
	Quality     float64           `json:"quality"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ContentSearchCriteria 内容搜索条件
type ContentSearchCriteria struct {
	Query       string        `json:"query"`
	Category    string        `json:"category"`
	Difficulty  string        `json:"difficulty"`
	Tags        []string      `json:"tags"`
	MinDuration time.Duration `json:"min_duration"`
	MaxDuration time.Duration `json:"max_duration"`
	Format      string        `json:"format"`
	Language    string        `json:"language"`
	Limit       int           `json:"limit"`
	Offset      int           `json:"offset"`
}

// LearningRecord 学习记录
type LearningRecord struct {
	UserID       string            `json:"user_id"`
	ContentID    string            `json:"content_id"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Progress     float64           `json:"progress"`
	Score        float64           `json:"score"`
	Completed    bool              `json:"completed"`
	Interactions int               `json:"interactions"`
	Context      map[string]interface{} `json:"context"`
}

// AnalyzeUserPreferences 分析用户偏好
func (pa *PreferenceAnalyzer) AnalyzeUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	// 获取用户基本信息
	userProfile, err := pa.userRepo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// 获取用户学习历史
	learningHistory, err := pa.userRepo.GetUserLearningHistory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning history: %w", err)
	}

	// 获取用户交互数据
	interactions, err := pa.userRepo.GetUserInteractions(ctx, userID, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interactions: %w", err)
	}

	// 获取行为数据
	behaviorSummary, err := pa.behaviorTracker.GetBehaviorSummary(ctx, userID, time.Now().AddDate(0, -3, 0), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}

	// 分析内容偏好
	contentPrefs := pa.analyzeContentPreferences(learningHistory, interactions)

	// 分析学习偏好
	learningPrefs := pa.analyzeLearningPreferences(userProfile, behaviorSummary, learningHistory)

	// 分析交互模式
	interactionPatterns := pa.analyzeInteractionPatterns(interactions, behaviorSummary)

	// 分析主题兴趣
	topicInterests := pa.analyzeTopicInterests(learningHistory, interactions)

	// 分析难度偏好
	difficultyPref := pa.analyzeDifficultyPreference(learningHistory, userProfile)

	// 分析时间偏好
	timePrefs := pa.analyzeTimePreferences(behaviorSummary, interactions)

	// 分析设备偏好
	devicePrefs := pa.analyzeDevicePreferences(interactions, behaviorSummary)

	// 分析语言偏好
	languagePrefs := pa.analyzeLanguagePreferences(learningHistory, userProfile)

	// 计算置信度
	confidence := pa.calculatePreferenceConfidence(learningHistory, interactions, behaviorSummary)

	preferences := &UserPreferences{
		UserID:              userID,
		ContentPreferences:  contentPrefs,
		LearningPreferences: learningPrefs,
		InteractionPatterns: interactionPatterns,
		TopicInterests:      topicInterests,
		DifficultyPreference: difficultyPref,
		TimePreferences:     timePrefs,
		DevicePreferences:   devicePrefs,
		LanguagePreferences: languagePrefs,
		UpdatedAt:           time.Now(),
		Confidence:          confidence,
	}

	return preferences, nil
}

// analyzeContentPreferences 分析内容偏好
func (pa *PreferenceAnalyzer) analyzeContentPreferences(
	learningHistory []*LearningRecord,
	interactions []*UserInteraction,
) *ContentPreferenceProfile {
	formatPrefs := make(map[string]float64)
	topicPrefs := make(map[string]float64)
	var totalDuration time.Duration
	var durationCount int
	var qualitySum float64
	var qualityCount int

	// 分析格式偏好
	for _, record := range learningHistory {
		if record.Completed {
			// 这里需要从内容中获取格式信息
			// 根据完成情况和分数调整偏好权重
			// formatPrefs[content.Format] += calculateWeight(record.Score)
		}
	}

	// 分析主题偏好
	for _, record := range learningHistory {
		if record.Progress > 0.7 { // 高完成度的内容
			// 这里需要从内容中获取主题信息
			// topicPrefs[content.Category] += record.Progress
		}
	}

	// 分析时长偏好
	for _, record := range learningHistory {
		if record.Completed {
			duration := record.EndTime.Sub(record.StartTime)
			totalDuration += duration
			durationCount++
		}
	}

	// 计算平均偏好时长
	var avgDuration time.Duration
	if durationCount > 0 {
		avgDuration = totalDuration / time.Duration(durationCount)
	}

	// 计算质量阈值
	if qualityCount > 0 {
		qualitySum = qualitySum / float64(qualityCount)
	}

	return &ContentPreferenceProfile{
		PreferredFormats: formatPrefs,
		PreferredDuration: &DurationRange{
			Min: time.Duration(float64(avgDuration) * 0.7),
			Max: time.Duration(float64(avgDuration) * 1.3),
		},
		PreferredTopics:  topicPrefs,
		QualityThreshold: math.Max(0.6, qualitySum*0.9), // 至少0.6的质量阈值
	}
}

// analyzeLearningPreferences 分析学习偏好
func (pa *PreferenceAnalyzer) analyzeLearningPreferences(
	userProfile *UserProfile,
	behaviorSummary *BehaviorSummary,
	learningHistory []*LearningRecord,
) *LearningPreferenceProfile {
	// 分析学习风格
	learningStyle := pa.inferLearningStyle(behaviorSummary, learningHistory)
	
	// 分析学习节奏
	pacePreference := pa.inferPacePreference(learningHistory, behaviorSummary)
	
	// 分析结构偏好
	structurePreference := pa.inferStructurePreference(behaviorSummary)
	
	// 分析反馈偏好
	feedbackPreference := pa.inferFeedbackPreference(behaviorSummary)
	
	// 分析社交偏好
	socialPreference := pa.inferSocialPreference(behaviorSummary)
	
	// 计算挑战级别偏好
	challengeLevel := pa.calculateChallengeLevel(learningHistory)
	
	// 计算重复容忍度
	repetitionTolerance := pa.calculateRepetitionTolerance(behaviorSummary)
	
	// 计算探索倾向
	explorationTendency := pa.calculateExplorationTendency(behaviorSummary, learningHistory)

	return &LearningPreferenceProfile{
		LearningStyle:       learningStyle,
		PacePreference:      pacePreference,
		StructurePreference: structurePreference,
		FeedbackPreference:  feedbackPreference,
		SocialPreference:    socialPreference,
		ChallengeLevel:      challengeLevel,
		RepetitionTolerance: repetitionTolerance,
		ExplorationTendency: explorationTendency,
	}
}

// analyzeInteractionPatterns 分析交互模式
func (pa *PreferenceAnalyzer) analyzeInteractionPatterns(
	interactions []*UserInteraction,
	behaviorSummary *BehaviorSummary,
) *InteractionPatterns {
	clickPatterns := make(map[string]float64)
	scrollPatterns := make(map[string]float64)
	navigationPatterns := make(map[string]float64)
	completionPatterns := make(map[string]float64)
	engagementPatterns := make(map[string]float64)
	var searchPatterns []string

	// 分析点击模式
	for _, interaction := range interactions {
		if interaction.Interaction == "click" {
			clickPatterns[interaction.ContentID.String()] += 1.0
		}
	}

	// 分析滚动模式
	for _, interaction := range interactions {
		if interaction.Interaction == "scroll" {
			scrollPatterns[interaction.ContentID.String()] += 1.0
		}
	}

	// 从行为摘要中提取模式
	if behaviorSummary != nil {
		// 计算平均注意力持续时间
		attentionSpan := behaviorSummary.AverageSessionTime
		sessionDuration := behaviorSummary.TotalTimeSpent
		
		return &InteractionPatterns{
			ClickPatterns:       clickPatterns,
			ScrollPatterns:      scrollPatterns,
			SearchPatterns:      searchPatterns,
			NavigationPatterns:  navigationPatterns,
			CompletionPatterns:  completionPatterns,
			EngagementPatterns:  engagementPatterns,
			AttentionSpan:       attentionSpan,
			SessionDuration:     sessionDuration,
		}
	}

	return &InteractionPatterns{
		ClickPatterns:       clickPatterns,
		ScrollPatterns:      scrollPatterns,
		SearchPatterns:      searchPatterns,
		NavigationPatterns:  navigationPatterns,
		CompletionPatterns:  completionPatterns,
		EngagementPatterns:  engagementPatterns,
		AttentionSpan:       30 * time.Minute, // 默认值
		SessionDuration:     60 * time.Minute, // 默认值
	}
}

// analyzeTopicInterests 分析主题兴趣
func (pa *PreferenceAnalyzer) analyzeTopicInterests(
	learningHistory []*LearningRecord,
	interactions []*UserInteraction,
) map[string]float64 {
	topicInterests := make(map[string]float64)
	
	// 基于学习历史分析主题兴趣
	for _, record := range learningHistory {
		// 这里需要从内容中获取主题信息
		// 根据完成度和分数计算兴趣权重
		weight := record.Progress * record.Score
		if record.Completed {
			weight *= 1.5 // 完成的内容给更高权重
		}
		
		// topicInterests[content.Category] += weight
	}
	
	// 基于交互数据分析主题兴趣
	for _, interaction := range interactions {
		if interaction.Interaction == "view" || interaction.Interaction == "click" {
			// 根据交互时长和类型计算兴趣
			// weight := float64(interaction.Duration) / 60.0 // 转换为分钟
			// topicInterests[content.Category] += weight * 0.5 // 交互权重较低
		}
	}
	
	// 归一化兴趣分数
	return pa.normalizeScores(topicInterests)
}

// analyzeDifficultyPreference 分析难度偏好
func (pa *PreferenceAnalyzer) analyzeDifficultyPreference(
	learningHistory []*LearningRecord,
	userProfile *UserProfile,
) *DifficultyPreference {
	difficultyHistory := make(map[string]float64)
	var totalScore float64
	var scoreCount int
	
	// 分析历史难度表现
	for _, record := range learningHistory {
		if record.Score > 0 {
			// 这里需要从内容中获取难度信息
			// difficultyHistory[content.Difficulty] += record.Score
			totalScore += record.Score
			scoreCount++
		}
	}
	
	// 计算平均表现
	var avgScore float64
	if scoreCount > 0 {
		avgScore = totalScore / float64(scoreCount)
	}
	
	// 推断偏好难度级别
	preferredLevel := "medium"
	if avgScore > 0.8 {
		preferredLevel = "hard"
	} else if avgScore < 0.6 {
		preferredLevel = "easy"
	}
	
	// 计算适应速度
	adaptationRate := pa.calculateAdaptationRate(learningHistory)
	
	// 判断是否寻求挑战
	challengeSeeker := avgScore > 0.7 && adaptationRate > 0.6
	
	return &DifficultyPreference{
		PreferredLevel:   preferredLevel,
		AdaptationRate:   adaptationRate,
		ChallengeSeeker:  challengeSeeker,
		ComfortZone: &DifficultyRange{
			Min: math.Max(0.0, avgScore-0.2),
			Max: math.Min(1.0, avgScore+0.2),
		},
		ProgressionSpeed:  adaptationRate,
		DifficultyHistory: difficultyHistory,
	}
}

// analyzeTimePreferences 分析时间偏好
func (pa *PreferenceAnalyzer) analyzeTimePreferences(
	behaviorSummary *BehaviorSummary,
	interactions []*UserInteraction,
) *TimePreferences {
	weeklyPattern := make(map[string]float64)
	var preferredTimes []PreferenceTimeSlot
	
	// 分析周模式
	if behaviorSummary != nil {
		// 从行为摘要中提取时间模式
		for day := 0; day < 7; day++ {
			dayName := time.Weekday(day).String()
			// 这里需要从行为数据中计算每天的活跃度
			weeklyPattern[dayName] = 1.0 // 默认值
		}
	}
	
	// 分析偏好时间段
	timeSlots := pa.extractTimeSlots(interactions)
	for _, slot := range timeSlots {
		preferredTimes = append(preferredTimes, slot)
	}
	
	// 计算平均会话长度
	sessionLength := 60 * time.Minute // 默认值
	if behaviorSummary != nil {
		sessionLength = behaviorSummary.AverageSessionTime
	}
	
	return &TimePreferences{
		PreferredTimes:  preferredTimes,
		SessionLength:   sessionLength,
		BreakFrequency:  15 * time.Minute, // 默认值
		WeeklyPattern:   weeklyPattern,
		SeasonalPattern: make(map[string]float64),
		TimeZonePreference: "UTC", // 默认值
	}
}

// analyzeDevicePreferences 分析设备偏好
func (pa *PreferenceAnalyzer) analyzeDevicePreferences(
	interactions []*UserInteraction,
	behaviorSummary *BehaviorSummary,
) *DevicePreferences {
	deviceCounts := make(map[string]float64)
	
	// 统计设备使用频率
	for _, interaction := range interactions {
		if device, ok := interaction.Context["device"].(string); ok && device != "" {
			deviceCounts[device] += 1.0
		}
	}
	
	// 归一化设备偏好
	totalCount := 0.0
	for _, count := range deviceCounts {
		totalCount += count
	}
	
	preferredDevices := make(map[string]float64)
	if totalCount > 0 {
		for device, count := range deviceCounts {
			preferredDevices[device] = count / totalCount
		}
	}
	
	return &DevicePreferences{
		PreferredDevices: preferredDevices,
		ScreenSize:       "medium", // 默认值
		InputMethod:      "touch",  // 默认值
		ConnectivityType: "wifi",   // 默认值
		PerformanceLevel: "medium", // 默认值
	}
}

// analyzeLanguagePreferences 分析语言偏好
func (pa *PreferenceAnalyzer) analyzeLanguagePreferences(
	learningHistory []*LearningRecord,
	userProfile *UserProfile,
) []string {
	languageCounts := make(map[string]int)
	
	// 从学习历史中统计语言使用
	for range learningHistory {
		// 这里需要从内容中获取语言信息
		// languageCounts[content.Language]++
	}
	
	// 排序并返回偏好语言
	type languageCount struct {
		language string
		count    int
	}
	
	var languages []languageCount
	for lang, count := range languageCounts {
		languages = append(languages, languageCount{lang, count})
	}
	
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].count > languages[j].count
	})
	
	var preferences []string
	for _, lang := range languages {
		preferences = append(preferences, lang.language)
	}
	
	// 如果没有历史数据，使用默认语言
	if len(preferences) == 0 {
		preferences = []string{"zh-CN", "en-US"}
	}
	
	return preferences
}

// 辅助方法

// inferLearningStyle 推断学习风格
func (pa *PreferenceAnalyzer) inferLearningStyle(
	behaviorSummary *BehaviorSummary,
	learningHistory []*LearningRecord,
) string {
	// 基于行为数据推断学习风格
	// 这里可以根据用户的交互模式来判断
	// 例如：视频观看时间长 -> 视觉学习者
	//      音频内容偏好 -> 听觉学习者
	//      交互式内容偏好 -> 动手学习者
	
	return "visual" // 默认值
}

// inferPacePreference 推断节奏偏好
func (pa *PreferenceAnalyzer) inferPacePreference(
	learningHistory []*LearningRecord,
	behaviorSummary *BehaviorSummary,
) string {
	// 基于学习完成速度和行为模式推断节奏偏好
	var totalDuration time.Duration
	var count int
	
	for _, record := range learningHistory {
		if record.Completed {
			duration := record.EndTime.Sub(record.StartTime)
			totalDuration += duration
			count++
		}
	}
	
	if count == 0 {
		return "normal"
	}
	
	avgDuration := totalDuration / time.Duration(count)
	
	// 根据平均学习时长判断节奏偏好
	if avgDuration < 30*time.Minute {
		return "fast"
	} else if avgDuration > 90*time.Minute {
		return "slow"
	}
	
	return "normal"
}

// inferStructurePreference 推断结构偏好
func (pa *PreferenceAnalyzer) inferStructurePreference(behaviorSummary *BehaviorSummary) string {
	// 基于导航模式推断结构偏好
	// 线性导航 -> 结构化
	// 跳跃式导航 -> 自由式
	
	return "structured" // 默认值
}

// inferFeedbackPreference 推断反馈偏好
func (pa *PreferenceAnalyzer) inferFeedbackPreference(behaviorSummary *BehaviorSummary) string {
	// 基于交互模式推断反馈偏好
	
	return "immediate" // 默认值
}

// inferSocialPreference 推断社交偏好
func (pa *PreferenceAnalyzer) inferSocialPreference(behaviorSummary *BehaviorSummary) string {
	// 基于社交行为推断偏好
	
	return "independent" // 默认值
}

// calculateChallengeLevel 计算挑战级别偏好
func (pa *PreferenceAnalyzer) calculateChallengeLevel(learningHistory []*LearningRecord) float64 {
	var totalScore float64
	var count int
	
	for _, record := range learningHistory {
		if record.Score > 0 {
			totalScore += record.Score
			count++
		}
	}
	
	if count == 0 {
		return 0.5 // 默认中等挑战
	}
	
	avgScore := totalScore / float64(count)
	
	// 分数越高，说明可以接受更高挑战
	return math.Min(1.0, avgScore+0.2)
}

// calculateRepetitionTolerance 计算重复容忍度
func (pa *PreferenceAnalyzer) calculateRepetitionTolerance(behaviorSummary *BehaviorSummary) float64 {
	// 基于重复行为模式计算容忍度
	
	return 0.6 // 默认值
}

// calculateExplorationTendency 计算探索倾向
func (pa *PreferenceAnalyzer) calculateExplorationTendency(
	behaviorSummary *BehaviorSummary,
	learningHistory []*LearningRecord,
) float64 {
	// 基于内容多样性和探索行为计算探索倾向
	
	return 0.7 // 默认值
}

// calculateAdaptationRate 计算适应速度
func (pa *PreferenceAnalyzer) calculateAdaptationRate(learningHistory []*LearningRecord) float64 {
	// 基于学习进步速度计算适应率
	
	return 0.6 // 默认值
}

// extractTimeSlots 提取时间段
func (pa *PreferenceAnalyzer) extractTimeSlots(interactions []*UserInteraction) []PreferenceTimeSlot {
	// 分析交互时间，提取偏好时间段
	
	return []PreferenceTimeSlot{
		{
			Start:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
			End:    time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
			Weight: 0.8,
		},
		{
			Start:  time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC),
			End:    time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC),
			Weight: 0.9,
		},
	}
}

// normalizeScores 归一化分数
func (pa *PreferenceAnalyzer) normalizeScores(scores map[string]float64) map[string]float64 {
	// 找到最大值
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}
	
	// 归一化
	normalized := make(map[string]float64)
	if maxScore > 0 {
		for key, score := range scores {
			normalized[key] = score / maxScore
		}
	}
	
	return normalized
}

// calculatePreferenceConfidence 计算偏好置信度
func (pa *PreferenceAnalyzer) calculatePreferenceConfidence(
	learningHistory []*LearningRecord,
	interactions []*UserInteraction,
	behaviorSummary *BehaviorSummary,
) float64 {
	// 基于数据量和一致性计算置信度
	
	dataPoints := float64(len(learningHistory) + len(interactions))
	
	// 数据点越多，置信度越高
	confidence := math.Min(1.0, dataPoints/100.0)
	
	// 确保最小置信度
	return math.Max(0.1, confidence)
}

// GetPreferenceInsights 获取偏好洞察
func (pa *PreferenceAnalyzer) GetPreferenceInsights(ctx context.Context, userID string) (*PreferenceInsights, error) {
	preferences, err := pa.AnalyzeUserPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze preferences: %w", err)
	}
	
	insights := &PreferenceInsights{
		UserID:      userID,
		Preferences: preferences,
		Insights:    pa.generateInsights(preferences),
		Recommendations: pa.generatePreferenceRecommendations(preferences),
		UpdatedAt:   time.Now(),
	}
	
	return insights, nil
}

// PreferenceInsights 偏好洞察
type PreferenceInsights struct {
	UserID          string                    `json:"user_id"`
	Preferences     *UserPreferences          `json:"preferences"`
	Insights        []string                  `json:"insights"`
	Recommendations []PreferenceRecommendation `json:"recommendations"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

// PreferenceRecommendation 偏好推荐
type PreferenceRecommendation struct {
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	Confidence  float64 `json:"confidence"`
}

// generateInsights 生成洞察
func (pa *PreferenceAnalyzer) generateInsights(preferences *UserPreferences) []string {
	var insights []string
	
	// 基于偏好生成洞察
	if preferences.LearningPreferences.ChallengeLevel > 0.8 {
		insights = append(insights, "用户喜欢具有挑战性的内容")
	}
	
	if preferences.LearningPreferences.ExplorationTendency > 0.7 {
		insights = append(insights, "用户具有较强的探索倾向，喜欢尝试新内容")
	}
	
	if preferences.TimePreferences.SessionLength > 90*time.Minute {
		insights = append(insights, "用户偏好长时间学习会话")
	}
	
	return insights
}

// generatePreferenceRecommendations 生成偏好推荐
func (pa *PreferenceAnalyzer) generatePreferenceRecommendations(preferences *UserPreferences) []PreferenceRecommendation {
	var recommendations []PreferenceRecommendation
	
	// 基于偏好生成推荐
	if preferences.LearningPreferences.ChallengeLevel < 0.5 {
		recommendations = append(recommendations, PreferenceRecommendation{
			Type:        "difficulty",
			Title:       "尝试更具挑战性的内容",
			Description: "基于您的学习表现，您可以尝试更高难度的内容来提升技能",
			Priority:    "medium",
			Confidence:  0.7,
		})
	}
	
	return recommendations
}