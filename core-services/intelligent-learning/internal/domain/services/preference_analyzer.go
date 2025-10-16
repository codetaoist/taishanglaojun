package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// PreferenceAnalyzer ?
type PreferenceAnalyzer struct {
	behaviorTracker *UserBehaviorTracker
	contentRepo     ContentRepository
	userRepo        UserRepository
}

// PreferenceRepository 
type PreferenceRepository interface {
	SaveUserPreferences(ctx context.Context, userID string, preferences *UserPreferences) error
	GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error)
	GetContentPreferences(ctx context.Context, contentID string) (*ContentPreferences, error)
	SavePreferenceHistory(ctx context.Context, history *PreferenceHistory) error
}

// ContentRepository 
type ContentRepository interface {
	GetContentByID(ctx context.Context, contentID string) (*Content, error)
	GetContentsByCategory(ctx context.Context, category string) ([]*Content, error)
	GetContentsByDifficulty(ctx context.Context, difficulty string) ([]*Content, error)
	GetContentsByTags(ctx context.Context, tags []string) ([]*Content, error)
}

// UserRepository 
type UserRepository interface {
	GetUserProfile(ctx context.Context, userID string) (*UserProfile, error)
	GetUserLearningHistory(ctx context.Context, userID string) ([]*LearningRecord, error)
	GetUserInteractions(ctx context.Context, userID string, limit int) ([]*UserInteraction, error)
}

// NewPreferenceAnalyzer ?
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

// UserPreferences 
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

// ContentPreferenceProfile 
type ContentPreferenceProfile struct {
	PreferredFormats    map[string]float64 `json:"preferred_formats"`    // 
	PreferredDuration   *DurationRange     `json:"preferred_duration"`   // ?
	PreferredComplexity string             `json:"preferred_complexity"` // ?
	PreferredTopics     map[string]float64 `json:"preferred_topics"`     // 
	AvoidedTopics       []string           `json:"avoided_topics"`       // ?
	QualityThreshold    float64            `json:"quality_threshold"`    // ?
}

// LearningPreferenceProfile 
type LearningPreferenceProfile struct {
	LearningStyle       string             `json:"learning_style"`       // 
	PacePreference      string             `json:"pace_preference"`      // ?
	StructurePreference string             `json:"structure_preference"` // 
	FeedbackPreference  string             `json:"feedback_preference"`  // 
	SocialPreference    string             `json:"social_preference"`    // ?
	ChallengeLevel      float64            `json:"challenge_level"`      //  0-1
	RepetitionTolerance float64            `json:"repetition_tolerance"` // ?0-1
	ExplorationTendency float64            `json:"exploration_tendency"` //  0-1
}

// InteractionPatterns 
type InteractionPatterns struct {
	ClickPatterns       map[string]float64 `json:"click_patterns"`       // 
	ScrollPatterns      map[string]float64 `json:"scroll_patterns"`      // 
	SearchPatterns      []string           `json:"search_patterns"`      // 
	NavigationPatterns  map[string]float64 `json:"navigation_patterns"`  // 
	CompletionPatterns  map[string]float64 `json:"completion_patterns"`  // 
	EngagementPatterns  map[string]float64 `json:"engagement_patterns"`  // 
	AttentionSpan       time.Duration      `json:"attention_span"`       // ?
	SessionDuration     time.Duration      `json:"session_duration"`     // 
}

// DifficultyPreference 
type DifficultyPreference struct {
	PreferredLevel    string             `json:"preferred_level"`    // 
	AdaptationRate    float64            `json:"adaptation_rate"`    // 
	ChallengeSeeker   bool               `json:"challenge_seeker"`   // 
	ComfortZone       *DifficultyRange   `json:"comfort_zone"`       // ?
	ProgressionSpeed  float64            `json:"progression_speed"`  // 
	DifficultyHistory map[string]float64 `json:"difficulty_history"` // 
}

// TimePreferences 
type TimePreferences struct {
	PreferredTimes    []PreferenceTimeSlot        `json:"preferred_times"`    // ?
	SessionLength     time.Duration     `json:"session_length"`     // 
	BreakFrequency    time.Duration     `json:"break_frequency"`    // 
	WeeklyPattern     map[string]float64 `json:"weekly_pattern"`     // ?
	SeasonalPattern   map[string]float64 `json:"seasonal_pattern"`   // 
	TimeZonePreference string           `json:"timezone_preference"` // 
}

// DevicePreferences 豸
type DevicePreferences struct {
	PreferredDevices map[string]float64 `json:"preferred_devices"` // 豸
	ScreenSize       string             `json:"screen_size"`       // 
	InputMethod      string             `json:"input_method"`      // 
	ConnectivityType string             `json:"connectivity_type"` // 
	PerformanceLevel string             `json:"performance_level"` // 
}

// DurationRange 
type DurationRange struct {
	Min time.Duration `json:"min"`
	Max time.Duration `json:"max"`
}

// DifficultyRange 
type DifficultyRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// TimeSlot ?
type PreferenceTimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Weight float64  `json:"weight"`
}

// ContentPreferences 
type ContentPreferences struct {
	ContentID       string             `json:"content_id"`
	UserRatings     map[string]float64 `json:"user_ratings"`     // 
	ViewCounts      int64              `json:"view_counts"`      // 
	CompletionRates map[string]float64 `json:"completion_rates"` // ?
	EngagementScore float64            `json:"engagement_score"` // ?
	PopularityScore float64            `json:"popularity_score"` // ?
	QualityScore    float64            `json:"quality_score"`    // 
}

// PreferenceHistory 
type PreferenceHistory struct {
	UserID      string                 `json:"user_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Action      string                 `json:"action"`
	ContentID   string                 `json:"content_id"`
	Context     map[string]interface{} `json:"context"`
	Preferences map[string]float64     `json:"preferences"`
	Confidence  float64                `json:"confidence"`
}

// Content 
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

// ContentSearchCriteria 
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

// LearningRecord 
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

// AnalyzeUserPreferences 
func (pa *PreferenceAnalyzer) AnalyzeUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	// 
	userProfile, err := pa.userRepo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// 
	learningHistory, err := pa.userRepo.GetUserLearningHistory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning history: %w", err)
	}

	// 
	interactions, err := pa.userRepo.GetUserInteractions(ctx, userID, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interactions: %w", err)
	}

	// 
	behaviorSummary, err := pa.behaviorTracker.GetBehaviorSummary(ctx, userID, time.Now().AddDate(0, -3, 0), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}

	// 
	contentPrefs := pa.analyzeContentPreferences(learningHistory, interactions)

	// 
	learningPrefs := pa.analyzeLearningPreferences(userProfile, behaviorSummary, learningHistory)

	// 
	interactionPatterns := pa.analyzeInteractionPatterns(interactions, behaviorSummary)

	// 
	topicInterests := pa.analyzeTopicInterests(learningHistory, interactions)

	// 
	difficultyPref := pa.analyzeDifficultyPreference(learningHistory, userProfile)

	// 
	timePrefs := pa.analyzeTimePreferences(behaviorSummary, interactions)

	// 豸
	devicePrefs := pa.analyzeDevicePreferences(interactions, behaviorSummary)

	// 
	languagePrefs := pa.analyzeLanguagePreferences(learningHistory, userProfile)

	// ?
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

// analyzeContentPreferences 
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

	// 
	for _, record := range learningHistory {
		if record.Completed {
			// ?
			// ?
			// formatPrefs[content.Format] += calculateWeight(record.Score)
		}
	}

	// 
	for _, record := range learningHistory {
		if record.Progress > 0.7 { // ?
			// ?
			// topicPrefs[content.Category] += record.Progress
		}
	}

	// 
	for _, record := range learningHistory {
		if record.Completed {
			duration := record.EndTime.Sub(record.StartTime)
			totalDuration += duration
			durationCount++
		}
	}

	// 
	var avgDuration time.Duration
	if durationCount > 0 {
		avgDuration = totalDuration / time.Duration(durationCount)
	}

	// ?
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
		QualityThreshold: math.Max(0.6, qualitySum*0.9), // 0.6?
	}
}

// analyzeLearningPreferences 
func (pa *PreferenceAnalyzer) analyzeLearningPreferences(
	userProfile *UserProfile,
	behaviorSummary *BehaviorSummary,
	learningHistory []*LearningRecord,
) *LearningPreferenceProfile {
	// 
	learningStyle := pa.inferLearningStyle(behaviorSummary, learningHistory)
	
	// 
	pacePreference := pa.inferPacePreference(learningHistory, behaviorSummary)
	
	// 
	structurePreference := pa.inferStructurePreference(behaviorSummary)
	
	// 
	feedbackPreference := pa.inferFeedbackPreference(behaviorSummary)
	
	// 罻
	socialPreference := pa.inferSocialPreference(behaviorSummary)
	
	// 
	challengeLevel := pa.calculateChallengeLevel(learningHistory)
	
	// ?
	repetitionTolerance := pa.calculateRepetitionTolerance(behaviorSummary)
	
	// 
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

// analyzeInteractionPatterns 
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

	// 
	for _, interaction := range interactions {
		if interaction.Interaction == "click" {
			clickPatterns[interaction.ContentID.String()] += 1.0
		}
	}

	// 
	for _, interaction := range interactions {
		if interaction.Interaction == "scroll" {
			scrollPatterns[interaction.ContentID.String()] += 1.0
		}
	}

	// 
	if behaviorSummary != nil {
		// ?
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
		AttentionSpan:       30 * time.Minute, // ?
		SessionDuration:     60 * time.Minute, // ?
	}
}

// analyzeTopicInterests 
func (pa *PreferenceAnalyzer) analyzeTopicInterests(
	learningHistory []*LearningRecord,
	interactions []*UserInteraction,
) map[string]float64 {
	topicInterests := make(map[string]float64)
	
	// 
	for _, record := range learningHistory {
		// ?
		// 
		weight := record.Progress * record.Score
		if record.Completed {
			weight *= 1.5 // 
		}
		
		// topicInterests[content.Category] += weight
	}
	
	// 
	for _, interaction := range interactions {
		if interaction.Interaction == "view" || interaction.Interaction == "click" {
			// ?
			// weight := float64(interaction.Duration) / 60.0 // ?
			// topicInterests[content.Category] += weight * 0.5 // 
		}
	}
	
	// ?
	return pa.normalizeScores(topicInterests)
}

// analyzeDifficultyPreference 
func (pa *PreferenceAnalyzer) analyzeDifficultyPreference(
	learningHistory []*LearningRecord,
	userProfile *UserProfile,
) *DifficultyPreference {
	difficultyHistory := make(map[string]float64)
	var totalScore float64
	var scoreCount int
	
	// 
	for _, record := range learningHistory {
		if record.Score > 0 {
			// ?
			// difficultyHistory[content.Difficulty] += record.Score
			totalScore += record.Score
			scoreCount++
		}
	}
	
	// 
	var avgScore float64
	if scoreCount > 0 {
		avgScore = totalScore / float64(scoreCount)
	}
	
	// 
	preferredLevel := "medium"
	if avgScore > 0.8 {
		preferredLevel = "hard"
	} else if avgScore < 0.6 {
		preferredLevel = "easy"
	}
	
	// 
	adaptationRate := pa.calculateAdaptationRate(learningHistory)
	
	// 
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

// analyzeTimePreferences 
func (pa *PreferenceAnalyzer) analyzeTimePreferences(
	behaviorSummary *BehaviorSummary,
	interactions []*UserInteraction,
) *TimePreferences {
	weeklyPattern := make(map[string]float64)
	var preferredTimes []PreferenceTimeSlot
	
	// ?
	if behaviorSummary != nil {
		// 
		for day := 0; day < 7; day++ {
			dayName := time.Weekday(day).String()
			// ?
			weeklyPattern[dayName] = 1.0 // ?
		}
	}
	
	// ?
	timeSlots := pa.extractTimeSlots(interactions)
	for _, slot := range timeSlots {
		preferredTimes = append(preferredTimes, slot)
	}
	
	// 
	sessionLength := 60 * time.Minute // ?
	if behaviorSummary != nil {
		sessionLength = behaviorSummary.AverageSessionTime
	}
	
	return &TimePreferences{
		PreferredTimes:  preferredTimes,
		SessionLength:   sessionLength,
		BreakFrequency:  15 * time.Minute, // ?
		WeeklyPattern:   weeklyPattern,
		SeasonalPattern: make(map[string]float64),
		TimeZonePreference: "UTC", // ?
	}
}

// analyzeDevicePreferences 豸
func (pa *PreferenceAnalyzer) analyzeDevicePreferences(
	interactions []*UserInteraction,
	behaviorSummary *BehaviorSummary,
) *DevicePreferences {
	deviceCounts := make(map[string]float64)
	
	// 豸
	for _, interaction := range interactions {
		if device, ok := interaction.Context["device"].(string); ok && device != "" {
			deviceCounts[device] += 1.0
		}
	}
	
	// 豸?
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
		ScreenSize:       "medium", // ?
		InputMethod:      "touch",  // ?
		ConnectivityType: "wifi",   // ?
		PerformanceLevel: "medium", // ?
	}
}

// analyzeLanguagePreferences 
func (pa *PreferenceAnalyzer) analyzeLanguagePreferences(
	learningHistory []*LearningRecord,
	userProfile *UserProfile,
) []string {
	languageCounts := make(map[string]int)
	
	// 
	for range learningHistory {
		// 
		// languageCounts[content.Language]++
	}
	
	// 
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
	
	// 
	if len(preferences) == 0 {
		preferences = []string{"zh-CN", "en-US"}
	}
	
	return preferences
}

// 

// inferLearningStyle 
func (pa *PreferenceAnalyzer) inferLearningStyle(
	behaviorSummary *BehaviorSummary,
	learningHistory []*LearningRecord,
) string {
	// 
	// 
	// 䳤 -> ?
	//       -> ?
	//      ?-> ?
	
	return "visual" // ?
}

// inferPacePreference 
func (pa *PreferenceAnalyzer) inferPacePreference(
	learningHistory []*LearningRecord,
	behaviorSummary *BehaviorSummary,
) string {
	// ?
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
	
	// 
	if avgDuration < 30*time.Minute {
		return "fast"
	} else if avgDuration > 90*time.Minute {
		return "slow"
	}
	
	return "normal"
}

// inferStructurePreference 
func (pa *PreferenceAnalyzer) inferStructurePreference(behaviorSummary *BehaviorSummary) string {
	// 
	// ?-> ?
	// ?-> ?
	
	return "structured" // ?
}

// inferFeedbackPreference 
func (pa *PreferenceAnalyzer) inferFeedbackPreference(behaviorSummary *BehaviorSummary) string {
	// 
	
	return "immediate" // ?
}

// inferSocialPreference 罻
func (pa *PreferenceAnalyzer) inferSocialPreference(behaviorSummary *BehaviorSummary) string {
	// 罻
	
	return "independent" // ?
}

// calculateChallengeLevel 
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
		return 0.5 // 
	}
	
	avgScore := totalScore / float64(count)
	
	// ?
	return math.Min(1.0, avgScore+0.2)
}

// calculateRepetitionTolerance ?
func (pa *PreferenceAnalyzer) calculateRepetitionTolerance(behaviorSummary *BehaviorSummary) float64 {
	// ?
	
	return 0.6 // ?
}

// calculateExplorationTendency 
func (pa *PreferenceAnalyzer) calculateExplorationTendency(
	behaviorSummary *BehaviorSummary,
	learningHistory []*LearningRecord,
) float64 {
	// 
	
	return 0.7 // ?
}

// calculateAdaptationRate 
func (pa *PreferenceAnalyzer) calculateAdaptationRate(learningHistory []*LearningRecord) float64 {
	// ?
	
	return 0.6 // ?
}

// extractTimeSlots ?
func (pa *PreferenceAnalyzer) extractTimeSlots(interactions []*UserInteraction) []PreferenceTimeSlot {
	// 
	
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

// normalizeScores ?
func (pa *PreferenceAnalyzer) normalizeScores(scores map[string]float64) map[string]float64 {
	// ?
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}
	
	// ?
	normalized := make(map[string]float64)
	if maxScore > 0 {
		for key, score := range scores {
			normalized[key] = score / maxScore
		}
	}
	
	return normalized
}

// calculatePreferenceConfidence ?
func (pa *PreferenceAnalyzer) calculatePreferenceConfidence(
	learningHistory []*LearningRecord,
	interactions []*UserInteraction,
	behaviorSummary *BehaviorSummary,
) float64 {
	// 
	
	dataPoints := float64(len(learningHistory) + len(interactions))
	
	// ?
	confidence := math.Min(1.0, dataPoints/100.0)
	
	// 
	return math.Max(0.1, confidence)
}

// GetPreferenceInsights 
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

// PreferenceInsights 
type PreferenceInsights struct {
	UserID          string                    `json:"user_id"`
	Preferences     *UserPreferences          `json:"preferences"`
	Insights        []string                  `json:"insights"`
	Recommendations []PreferenceRecommendation `json:"recommendations"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

// PreferenceRecommendation 
type PreferenceRecommendation struct {
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	Confidence  float64 `json:"confidence"`
}

// generateInsights 
func (pa *PreferenceAnalyzer) generateInsights(preferences *UserPreferences) []string {
	var insights []string
	
	// 
	if preferences.LearningPreferences.ChallengeLevel > 0.8 {
		insights = append(insights, "")
	}
	
	if preferences.LearningPreferences.ExplorationTendency > 0.7 {
		insights = append(insights, "")
	}
	
	if preferences.TimePreferences.SessionLength > 90*time.Minute {
		insights = append(insights, "?)
	}
	
	return insights
}

// generatePreferenceRecommendations 
func (pa *PreferenceAnalyzer) generatePreferenceRecommendations(preferences *UserPreferences) []PreferenceRecommendation {
	var recommendations []PreferenceRecommendation
	
	// 
	if preferences.LearningPreferences.ChallengeLevel < 0.5 {
		recommendations = append(recommendations, PreferenceRecommendation{
			Type:        "difficulty",
			Title:       "",
			Description: "?,
			Priority:    "medium",
			Confidence:  0.7,
		})
	}
	
	return recommendations
}

