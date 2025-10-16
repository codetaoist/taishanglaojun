package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// ContextAnalyzer 
type ContextAnalyzer struct {
	behaviorTracker   *UserBehaviorTracker
	preferenceAnalyzer *PreferenceAnalyzer
	environmentRepo   EnvironmentRepository
	locationService   LocationService
	weatherService    WeatherService
}

// EnvironmentRepository 
type EnvironmentRepository interface {
	SaveEnvironmentData(ctx context.Context, data *EnvironmentData) error
	GetEnvironmentData(ctx context.Context, userID string, timeRange ContextTimeRange) ([]*EnvironmentData, error)
	GetContextHistory(ctx context.Context, userID string, limit int) ([]*ContextRecord, error)
}

// LocationService 
type LocationService interface {
	GetCurrentLocation(ctx context.Context, userID string) (*Location, error)
	GetLocationHistory(ctx context.Context, userID string, timeRange ContextTimeRange) ([]*LocationRecord, error)
	GetLocationContext(ctx context.Context, location *Location) (*LocationContext, error)
}

// WeatherService 
type WeatherService interface {
	GetCurrentWeather(ctx context.Context, location *Location) (*WeatherInfo, error)
	GetWeatherForecast(ctx context.Context, location *Location, days int) ([]*WeatherInfo, error)
}

// NewContextAnalyzer 
func NewContextAnalyzer(
	behaviorTracker *UserBehaviorTracker,
	preferenceAnalyzer *PreferenceAnalyzer,
	environmentRepo EnvironmentRepository,
	locationService LocationService,
	weatherService WeatherService,
) *ContextAnalyzer {
	return &ContextAnalyzer{
		behaviorTracker:    behaviorTracker,
		preferenceAnalyzer: preferenceAnalyzer,
		environmentRepo:    environmentRepo,
		locationService:    locationService,
		weatherService:     weatherService,
	}
}

// LearningContext ?
type LearningContext struct {
	UserID              string                 `json:"user_id"`
	Timestamp           time.Time              `json:"timestamp"`
	EnvironmentalContext *EnvironmentalContext `json:"environmental_context"`
	TemporalContext     *TemporalContext       `json:"temporal_context"`
	SocialContext       *SocialContext         `json:"social_context"`
	DeviceContext       *DeviceContext         `json:"device_context"`
	CognitiveContext    *CognitiveContext      `json:"cognitive_context"`
	MotivationalContext *MotivationalContext   `json:"motivational_context"`
	ContextualFactors   map[string]float64     `json:"contextual_factors"`
	ContextScore        float64                `json:"context_score"`
	Recommendations     []ContextualRecommendation `json:"recommendations"`
}

// EnvironmentalContext ?
type EnvironmentalContext struct {
	Location        *Location     `json:"location"`
	Weather         *WeatherInfo  `json:"weather"`
	NoiseLevel      float64       `json:"noise_level"`      //  0-1
	LightLevel      float64       `json:"light_level"`      //  0-1
	Temperature     float64       `json:"temperature"`      // 
	Humidity        float64       `json:"humidity"`         // 
	AirQuality      float64       `json:"air_quality"`      //  0-1
	Distractions    []Distraction `json:"distractions"`     // 
	Comfort         float64       `json:"comfort"`          //  0-1
	Productivity    float64       `json:"productivity"`     // ?0-1
}

// TemporalContext ?
type TemporalContext struct {
	TimeOfDay       string        `json:"time_of_day"`       // 
	DayOfWeek       string        `json:"day_of_week"`       // ?
	Season          string        `json:"season"`            // 
	Holiday         bool          `json:"holiday"`           // ?
	WorkingHours    bool          `json:"working_hours"`     // 
	PeakHours       bool          `json:"peak_hours"`        // 
	TimeZone        string        `json:"timezone"`          // 
	LocalTime       time.Time     `json:"local_time"`        // 
	Duration        time.Duration `json:"duration"`          // 
	Urgency         float64       `json:"urgency"`           // ?0-1
	EnergyLevel     float64       `json:"energy_level"`      //  0-1
	AttentionSpan   time.Duration `json:"attention_span"`    // ?
}

// SocialContext 罻?
type SocialContext struct {
	SocialSetting   string             `json:"social_setting"`   // 罻鹫?
	GroupSize       int                `json:"group_size"`       // 
	SocialPressure  float64            `json:"social_pressure"`  // 罻 0-1
	Collaboration   bool               `json:"collaboration"`    // 
	Competition     bool               `json:"competition"`      // 
	Support         float64            `json:"support"`          //  0-1
	Privacy         float64            `json:"privacy"`          //  0-1
	SocialInfluence map[string]float64 `json:"social_influence"` // 罻
}

// DeviceContext 豸?
type DeviceContext struct {
	DeviceType      string  `json:"device_type"`      // 豸
	ScreenSize      string  `json:"screen_size"`      // 
	BatteryLevel    float64 `json:"battery_level"`    //  0-1
	NetworkQuality  float64 `json:"network_quality"`  //  0-1
	StorageSpace    float64 `json:"storage_space"`    // 洢 0-1
	ProcessingPower float64 `json:"processing_power"` //  0-1
	InputMethod     string  `json:"input_method"`     // 
	Orientation     string  `json:"orientation"`      // 豸
	Accessibility   bool    `json:"accessibility"`    // ?
}

// CognitiveContext ?
type CognitiveContext struct {
	CognitiveLoad    float64           `json:"cognitive_load"`    //  0-1
	FocusLevel       float64           `json:"focus_level"`       //  0-1
	StressLevel      float64           `json:"stress_level"`      //  0-1
	FatigueLevel     float64           `json:"fatigue_level"`     //  0-1
	MoodState        string            `json:"mood_state"`        // ?
	ConfidenceLevel  float64           `json:"confidence_level"`  //  0-1
	LearningState    string            `json:"learning_state"`    // ?
	PriorKnowledge   float64           `json:"prior_knowledge"`   //  0-1
	MemoryCapacity   float64           `json:"memory_capacity"`   //  0-1
	ProcessingSpeed  float64           `json:"processing_speed"`  //  0-1
	MetacognitionLevel float64         `json:"metacognition_level"` // ?0-1
	CognitiveStyle   map[string]float64 `json:"cognitive_style"`   // 
}

// MotivationalContext ?
type MotivationalContext struct {
	MotivationLevel  float64            `json:"motivation_level"`  //  0-1
	GoalOrientation  string             `json:"goal_orientation"`  // 
	IntrinsicMotivation float64         `json:"intrinsic_motivation"` //  0-1
	ExtrinsicMotivation float64         `json:"extrinsic_motivation"` //  0-1
	SelfEfficacy     float64            `json:"self_efficacy"`     //  0-1
	Persistence      float64            `json:"persistence"`       // ?0-1
	CuriosityLevel   float64            `json:"curiosity_level"`   // ?0-1
	RewardSensitivity float64           `json:"reward_sensitivity"` // ?0-1
	MotivationalFactors map[string]float64 `json:"motivational_factors"` // 
}

// Location 
type Location struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	PlaceType   string  `json:"place_type"`   // ?
	Familiarity float64 `json:"familiarity"`  //  0-1
}

// WeatherInfo 
type WeatherInfo struct {
	Temperature   float64 `json:"temperature"`
	Humidity      float64 `json:"humidity"`
	Pressure      float64 `json:"pressure"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection string  `json:"wind_direction"`
	Visibility    float64 `json:"visibility"`
	CloudCover    float64 `json:"cloud_cover"`
	Precipitation float64 `json:"precipitation"`
	UVIndex       int     `json:"uv_index"`
	Condition     string  `json:"condition"`
	Description   string  `json:"description"`
	AirQuality    string  `json:"air_quality"`
	Sunrise       string  `json:"sunrise"`
	Sunset        string  `json:"sunset"`
	Comfort       string  `json:"comfort"`
	LearningIndex float64 `json:"learning_index"`
}

// Distraction 
type Distraction struct {
	Type      string  `json:"type"`      // 
	Intensity float64 `json:"intensity"` //  0-1
	Duration  time.Duration `json:"duration"` // 
	Source    string  `json:"source"`    // 
}

// ContextualRecommendation ?
type ContextualRecommendation struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Confidence  float64                `json:"confidence"`
	Context     map[string]interface{} `json:"context"`
	Actions     []RecommendedAction    `json:"actions"`
}

// RecommendedAction 
type RecommendedAction struct {
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Expected    string                 `json:"expected"`
}

// EnvironmentData 
type EnvironmentData struct {
	UserID      string                 `json:"user_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Location    *Location              `json:"location"`
	Weather     *WeatherInfo           `json:"weather"`
	DeviceInfo  *DeviceContext         `json:"device_info"`
	Environment map[string]interface{} `json:"environment"`
}

// ContextRecord ?
type ContextRecord struct {
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Context   *LearningContext       `json:"context"`
	Outcome   map[string]interface{} `json:"outcome"`
}

// LocationRecord 
type LocationRecord struct {
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Location  *Location `json:"location"`
	Duration  time.Duration `json:"duration"`
}

// LocationContext ?
type LocationContext struct {
	PlaceType     string  `json:"place_type"`
	NoiseLevel    float64 `json:"noise_level"`
	CrowdLevel    float64 `json:"crowd_level"`
	WiFiQuality   float64 `json:"wifi_quality"`
	Comfort       float64 `json:"comfort"`
	Productivity  float64 `json:"productivity"`
	Accessibility bool    `json:"accessibility"`
}

// ContextTimeRange ?
type ContextTimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AnalyzeLearningContext ?
func (ca *ContextAnalyzer) AnalyzeLearningContext(ctx context.Context, userID string) (*LearningContext, error) {
	timestamp := time.Now()
	
	// ?
	environmentalContext, err := ca.analyzeEnvironmentalContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze environmental context: %w", err)
	}
	
	// ?
	temporalContext := ca.analyzeTemporalContext(timestamp)
	
	// 罻?
	socialContext, err := ca.analyzeSocialContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze social context: %w", err)
	}
	
	// 豸?
	deviceContext, err := ca.analyzeDeviceContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze device context: %w", err)
	}
	
	// ?
	cognitiveContext, err := ca.analyzeCognitiveContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze cognitive context: %w", err)
	}
	
	// ?
	motivationalContext, err := ca.analyzeMotivationalContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze motivational context: %w", err)
	}
	
	// ?
	contextualFactors := ca.calculateContextualFactors(
		environmentalContext,
		temporalContext,
		socialContext,
		deviceContext,
		cognitiveContext,
		motivationalContext,
	)
	
	// ?
	contextScore := ca.calculateContextScore(contextualFactors)
	
	// ?
	recommendations := ca.generateContextualRecommendations(
		environmentalContext,
		temporalContext,
		socialContext,
		deviceContext,
		cognitiveContext,
		motivationalContext,
		contextualFactors,
	)
	
	learningContext := &LearningContext{
		UserID:              userID,
		Timestamp:           timestamp,
		EnvironmentalContext: environmentalContext,
		TemporalContext:     temporalContext,
		SocialContext:       socialContext,
		DeviceContext:       deviceContext,
		CognitiveContext:    cognitiveContext,
		MotivationalContext: motivationalContext,
		ContextualFactors:   contextualFactors,
		ContextScore:        contextScore,
		Recommendations:     recommendations,
	}
	
	return learningContext, nil
}

// analyzeEnvironmentalContext ?
func (ca *ContextAnalyzer) analyzeEnvironmentalContext(ctx context.Context, userID string) (*EnvironmentalContext, error) {
	// 
	location, err := ca.locationService.GetCurrentLocation(ctx, userID)
	if err != nil {
		// ?
		location = &Location{
			PlaceType:   "unknown",
			Familiarity: 0.5,
		}
	}
	
	// ?
	locationContext, err := ca.locationService.GetLocationContext(ctx, location)
	if err != nil {
		locationContext = &LocationContext{
			NoiseLevel:   0.5,
			CrowdLevel:   0.5,
			WiFiQuality:  0.7,
			Comfort:      0.6,
			Productivity: 0.6,
		}
	}
	
	// 
	weather, err := ca.weatherService.GetCurrentWeather(ctx, location)
	if err != nil {
		weather = &WeatherInfo{
			Temperature: 22.0,
			Humidity:    50.0,
			Condition:   "unknown",
			AirQuality:  "",
		}
	}
	
	// 
	distractions := ca.analyzeDistractions(locationContext, weather)
	
	// 
	comfort := ca.calculateComfort(location, weather, locationContext)
	productivity := ca.calculateProductivity(location, weather, locationContext, distractions)
	
	return &EnvironmentalContext{
		Location:     location,
		Weather:      weather,
		NoiseLevel:   locationContext.NoiseLevel,
		LightLevel:   ca.calculateLightLevel(weather),
		Temperature:  weather.Temperature,
		Humidity:     weather.Humidity,
		AirQuality:   ca.convertAirQualityToFloat(weather.AirQuality),
		Distractions: distractions,
		Comfort:      comfort,
		Productivity: productivity,
	}, nil
}

// analyzeTemporalContext ?
func (ca *ContextAnalyzer) analyzeTemporalContext(timestamp time.Time) *TemporalContext {
	// ?
	timeOfDay := ca.getTimeOfDay(timestamp)
	
	// ?
	dayOfWeek := timestamp.Weekday().String()
	
	// 
	season := ca.getSeason(timestamp)
	
	// ?
	holiday := ca.isHoliday(timestamp)
	
	// 
	workingHours := ca.isWorkingHours(timestamp)
	
	// 
	peakHours := ca.isPeakHours(timestamp)
	
	// 㾫
	energyLevel := ca.calculateEnergyLevel(timestamp)
	
	// ?
	attentionSpan := ca.estimateAttentionSpan(timestamp, energyLevel)
	
	return &TemporalContext{
		TimeOfDay:     timeOfDay,
		DayOfWeek:     dayOfWeek,
		Season:        season,
		Holiday:       holiday,
		WorkingHours:  workingHours,
		PeakHours:     peakHours,
		TimeZone:      timestamp.Location().String(),
		LocalTime:     timestamp,
		Duration:      2 * time.Hour, // 
		Urgency:       0.5,           // ?
		EnergyLevel:   energyLevel,
		AttentionSpan: attentionSpan,
	}
}

// analyzeSocialContext 罻?
func (ca *ContextAnalyzer) analyzeSocialContext(ctx context.Context, userID string) (*SocialContext, error) {
	// 罻
	// 罻
	
	socialSetting := "private" // 
	groupSize := 1             // 
	socialPressure := 0.3      // 罻?
	collaboration := false     // ?
	competition := false       // ?
	support := 0.6            // 
	privacy := 0.8            // ?
	
	socialInfluence := map[string]float64{
		"peer_pressure":    0.3,
		"family_support":   0.7,
		"teacher_guidance": 0.6,
		"social_media":     0.4,
	}
	
	return &SocialContext{
		SocialSetting:   socialSetting,
		GroupSize:       groupSize,
		SocialPressure:  socialPressure,
		Collaboration:   collaboration,
		Competition:     competition,
		Support:         support,
		Privacy:         privacy,
		SocialInfluence: socialInfluence,
	}, nil
}

// analyzeDeviceContext 豸?
func (ca *ContextAnalyzer) analyzeDeviceContext(ctx context.Context, userID string) (*DeviceContext, error) {
	// 豸
	behaviorSummary, err := ca.behaviorTracker.GetBehaviorSummary(ctx, userID, time.Now().AddDate(0, 0, -1), time.Now())
	if err != nil {
		// 豸?
		return &DeviceContext{
			DeviceType:      "desktop",
			ScreenSize:      "large",
			BatteryLevel:    0.8,
			NetworkQuality:  0.8,
			StorageSpace:    0.7,
			ProcessingPower: 0.8,
			InputMethod:     "keyboard",
			Orientation:     "landscape",
			Accessibility:   false,
		}, nil
	}
	
	// 豸
	deviceType := ca.inferDeviceType(behaviorSummary)
	screenSize := ca.inferScreenSize(deviceType)
	inputMethod := ca.inferInputMethod(deviceType)
	
	return &DeviceContext{
		DeviceType:      deviceType,
		ScreenSize:      screenSize,
		BatteryLevel:    0.8, // 豸API
		NetworkQuality:  0.8, // ?
		StorageSpace:    0.7, // 豸API
		ProcessingPower: 0.8, // 豸
		InputMethod:     inputMethod,
		Orientation:     "landscape", // ?
		Accessibility:   false,       // ?
	}, nil
}

// analyzeCognitiveContext ?
func (ca *ContextAnalyzer) analyzeCognitiveContext(ctx context.Context, userID string) (*CognitiveContext, error) {
	// ?
	behaviorSummary, err := ca.behaviorTracker.GetBehaviorSummary(ctx, userID, time.Now().AddDate(0, 0, -7), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}
	
	// 
	cognitiveLoad := ca.calculateCognitiveLoad(behaviorSummary)
	
	// 
	focusLevel := ca.calculateFocusLevel(behaviorSummary)
	
	// 
	stressLevel := ca.calculateStressLevel(behaviorSummary)
	
	// 
	fatigueLevel := ca.calculateFatigueLevel(behaviorSummary)
	
	// ?
	moodState := ca.inferMoodState(behaviorSummary)
	
	// 
	confidenceLevel := ca.calculateConfidenceLevel(behaviorSummary)
	
	// ?
	learningState := ca.inferLearningState(behaviorSummary)
	
	// 
	priorKnowledge := ca.assessPriorKnowledge(ctx, userID)
	
	// 
	memoryCapacity := ca.assessMemoryCapacity(behaviorSummary)
	
	// 
	processingSpeed := ca.assessProcessingSpeed(behaviorSummary)
	
	// ?
	metacognitionLevel := ca.assessMetacognitionLevel(behaviorSummary)
	
	// 
	cognitiveStyle := ca.analyzeCognitiveStyle(behaviorSummary)
	
	return &CognitiveContext{
		CognitiveLoad:      cognitiveLoad,
		FocusLevel:         focusLevel,
		StressLevel:        stressLevel,
		FatigueLevel:       fatigueLevel,
		MoodState:          moodState,
		ConfidenceLevel:    confidenceLevel,
		LearningState:      learningState,
		PriorKnowledge:     priorKnowledge,
		MemoryCapacity:     memoryCapacity,
		ProcessingSpeed:    processingSpeed,
		MetacognitionLevel: metacognitionLevel,
		CognitiveStyle:     cognitiveStyle,
	}, nil
}

// analyzeMotivationalContext ?
func (ca *ContextAnalyzer) analyzeMotivationalContext(ctx context.Context, userID string) (*MotivationalContext, error) {
	// ?
	preferences, err := ca.preferenceAnalyzer.AnalyzeUserPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze user preferences: %w", err)
	}
	
	behaviorSummary, err := ca.behaviorTracker.GetBehaviorSummary(ctx, userID, time.Now().AddDate(0, 0, -7), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}
	
	// 㶯
	motivationLevel := ca.calculateMotivationLevel(behaviorSummary, preferences)
	
	// 
	goalOrientation := ca.inferGoalOrientation(preferences)
	
	// 
	intrinsicMotivation := ca.calculateIntrinsicMotivation(behaviorSummary, preferences)
	
	// 
	extrinsicMotivation := ca.calculateExtrinsicMotivation(behaviorSummary, preferences)
	
	// 
	selfEfficacy := ca.assessSelfEfficacy(behaviorSummary)
	
	// ?
	persistence := ca.assessPersistence(behaviorSummary)
	
	// ?
	curiosityLevel := ca.assessCuriosityLevel(behaviorSummary, preferences)
	
	// ?
	rewardSensitivity := ca.assessRewardSensitivity(behaviorSummary)
	
	// 
	motivationalFactors := ca.analyzeMotivationalFactors(behaviorSummary, preferences)
	
	return &MotivationalContext{
		MotivationLevel:     motivationLevel,
		GoalOrientation:     goalOrientation,
		IntrinsicMotivation: intrinsicMotivation,
		ExtrinsicMotivation: extrinsicMotivation,
		SelfEfficacy:        selfEfficacy,
		Persistence:         persistence,
		CuriosityLevel:      curiosityLevel,
		RewardSensitivity:   rewardSensitivity,
		MotivationalFactors: motivationalFactors,
	}, nil
}

// 

// analyzeDistractions 
func (ca *ContextAnalyzer) analyzeDistractions(locationContext *LocationContext, weather *WeatherInfo) []Distraction {
	var distractions []Distraction
	
	// 
	if locationContext.NoiseLevel > 0.7 {
		distractions = append(distractions, Distraction{
			Type:      "noise",
			Intensity: locationContext.NoiseLevel,
			Duration:  time.Hour,
			Source:    "environment",
		})
	}
	
	// 
	if locationContext.CrowdLevel > 0.6 {
		distractions = append(distractions, Distraction{
			Type:      "crowd",
			Intensity: locationContext.CrowdLevel,
			Duration:  time.Hour,
			Source:    "social",
		})
	}
	
	// 
	if weather.Condition == "storm" || weather.WindSpeed > 20 {
		distractions = append(distractions, Distraction{
			Type:      "weather",
			Intensity: 0.8,
			Duration:  2 * time.Hour,
			Source:    "environmental",
		})
	}
	
	return distractions
}

// calculateComfort 
func (ca *ContextAnalyzer) calculateComfort(location *Location, weather *WeatherInfo, locationContext *LocationContext) float64 {
	comfort := 0.5 // 
	
	// 
	comfort += location.Familiarity * 0.3
	
	// 
	tempComfort := 1.0 - math.Abs(weather.Temperature-22.0)/20.0 // 22?
	comfort += tempComfort * 0.2
	
	// 
	humidityComfort := 1.0 - math.Abs(weather.Humidity-50.0)/50.0 // 50%?
	comfort += humidityComfort * 0.1
	
	// 
	comfort += locationContext.Comfort * 0.4
	
	return math.Max(0.0, math.Min(1.0, comfort))
}

// calculateProductivity ?
func (ca *ContextAnalyzer) calculateProductivity(location *Location, weather *WeatherInfo, locationContext *LocationContext, distractions []Distraction) float64 {
	productivity := 0.7 // ?
	
	// ?
	productivity += locationContext.Productivity * 0.3
	
	// ?
	productivity += locationContext.WiFiQuality * 0.2
	
	// ?
	for _, distraction := range distractions {
		productivity -= distraction.Intensity * 0.1
	}
	
	// ?
	productivity += ca.convertAirQualityToFloat(weather.AirQuality) * 0.1
	
	return math.Max(0.0, math.Min(1.0, productivity))
}

// calculateLightLevel 
func (ca *ContextAnalyzer) calculateLightLevel(weather *WeatherInfo) float64 {
	// 㸲?
	lightLevel := 1.0 - weather.CloudCover
	
	// 
	hour := time.Now().Hour()
	if hour < 6 || hour > 20 {
		lightLevel *= 0.1 // 
	} else if hour < 8 || hour > 18 {
		lightLevel *= 0.5 // 
	}
	
	return math.Max(0.0, math.Min(1.0, lightLevel))
}

// getTimeOfDay ?
func (ca *ContextAnalyzer) getTimeOfDay(timestamp time.Time) string {
	hour := timestamp.Hour()
	
	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 21:
		return "evening"
	default:
		return "night"
	}
}

// getSeason 
func (ca *ContextAnalyzer) getSeason(timestamp time.Time) string {
	month := timestamp.Month()
	
	switch {
	case month >= 3 && month <= 5:
		return "spring"
	case month >= 6 && month <= 8:
		return "summer"
	case month >= 9 && month <= 11:
		return "autumn"
	default:
		return "winter"
	}
}

// isHoliday ?
func (ca *ContextAnalyzer) isHoliday(timestamp time.Time) bool {
	// ?
	weekday := timestamp.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// isWorkingHours 
func (ca *ContextAnalyzer) isWorkingHours(timestamp time.Time) bool {
	hour := timestamp.Hour()
	weekday := timestamp.Weekday()
	
	return weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour < 18
}

// isPeakHours 
func (ca *ContextAnalyzer) isPeakHours(timestamp time.Time) bool {
	hour := timestamp.Hour()
	
	// ?
	return (hour >= 7 && hour <= 9) || (hour >= 17 && hour <= 19)
}

// calculateEnergyLevel 㾫
func (ca *ContextAnalyzer) calculateEnergyLevel(timestamp time.Time) float64 {
	hour := timestamp.Hour()
	
	// 
	switch {
	case hour >= 9 && hour <= 11:
		return 0.9 // 羫?
	case hour >= 14 && hour <= 16:
		return 0.8 // 羫
	case hour >= 19 && hour <= 21:
		return 0.7 // 
	case hour >= 6 && hour <= 8:
		return 0.6 // 糿?
	case hour >= 22 || hour <= 5:
		return 0.3 // 
	default:
		return 0.5 // 侫
	}
}

// estimateAttentionSpan ?
func (ca *ContextAnalyzer) estimateAttentionSpan(timestamp time.Time, energyLevel float64) time.Duration {
	// ?
	baseSpan := 25 * time.Minute // 
	
	// 
	adjustedSpan := time.Duration(float64(baseSpan) * energyLevel)
	
	// ?
	hour := timestamp.Hour()
	if hour >= 22 || hour <= 6 {
		adjustedSpan = adjustedSpan / 2 // ?
	}
	
	return adjustedSpan
}

// inferDeviceType 豸
func (ca *ContextAnalyzer) inferDeviceType(behaviorSummary *BehaviorSummary) string {
	// 豸
	// ?
	
	return "desktop" // ?
}

// inferScreenSize 
func (ca *ContextAnalyzer) inferScreenSize(deviceType string) string {
	switch deviceType {
	case "mobile":
		return "small"
	case "tablet":
		return "medium"
	case "desktop":
		return "large"
	default:
		return "medium"
	}
}

// inferInputMethod 
func (ca *ContextAnalyzer) inferInputMethod(deviceType string) string {
	switch deviceType {
	case "mobile", "tablet":
		return "touch"
	case "desktop":
		return "keyboard"
	default:
		return "touch"
	}
}

// calculateCognitiveLoad 
func (ca *ContextAnalyzer) calculateCognitiveLoad(behaviorSummary *BehaviorSummary) float64 {
	// 
	
	return 0.6 // ?
}

// calculateFocusLevel 
func (ca *ContextAnalyzer) calculateFocusLevel(behaviorSummary *BehaviorSummary) float64 {
	// ?
	
	return 0.7 // ?
}

// calculateStressLevel 
func (ca *ContextAnalyzer) calculateStressLevel(behaviorSummary *BehaviorSummary) float64 {
	// ?
	
	return 0.4 // ?
}

// calculateFatigueLevel 
func (ca *ContextAnalyzer) calculateFatigueLevel(behaviorSummary *BehaviorSummary) float64 {
	// 仯?
	
	return 0.3 // ?
}

// inferMoodState ?
func (ca *ContextAnalyzer) inferMoodState(behaviorSummary *BehaviorSummary) string {
	// ?
	
	return "neutral" // ?
}

// calculateConfidenceLevel 
func (ca *ContextAnalyzer) calculateConfidenceLevel(behaviorSummary *BehaviorSummary) float64 {
	// ?
	
	return 0.6 // ?
}

// inferLearningState ?
func (ca *ContextAnalyzer) inferLearningState(behaviorSummary *BehaviorSummary) string {
	// ?
	
	return "active" // ?
}

// assessPriorKnowledge 
func (ca *ContextAnalyzer) assessPriorKnowledge(ctx context.Context, userID string) float64 {
	// 
	
	return 0.5 // ?
}

// assessMemoryCapacity 
func (ca *ContextAnalyzer) assessMemoryCapacity(behaviorSummary *BehaviorSummary) float64 {
	// 
	
	return 0.7 // ?
}

// assessProcessingSpeed 
func (ca *ContextAnalyzer) assessProcessingSpeed(behaviorSummary *BehaviorSummary) float64 {
	// 
	
	return 0.6 // ?
}

// assessMetacognitionLevel ?
func (ca *ContextAnalyzer) assessMetacognitionLevel(behaviorSummary *BehaviorSummary) float64 {
	// ?
	
	return 0.5 // ?
}

// analyzeCognitiveStyle 
func (ca *ContextAnalyzer) analyzeCognitiveStyle(behaviorSummary *BehaviorSummary) map[string]float64 {
	// 
	
	return map[string]float64{
		"field_dependent":   0.4,
		"field_independent": 0.6,
		"sequential":        0.5,
		"random":           0.5,
		"concrete":         0.6,
		"abstract":         0.4,
	}
}

// calculateMotivationLevel 㶯
func (ca *ContextAnalyzer) calculateMotivationLevel(behaviorSummary *BehaviorSummary, preferences *UserPreferences) float64 {
	// 㶯?
	
	return 0.7 // ?
}

// inferGoalOrientation 
func (ca *ContextAnalyzer) inferGoalOrientation(preferences *UserPreferences) string {
	// 
	
	return "mastery" // 
}

// calculateIntrinsicMotivation 
func (ca *ContextAnalyzer) calculateIntrinsicMotivation(behaviorSummary *BehaviorSummary, preferences *UserPreferences) float64 {
	// 
	
	return 0.6 // ?
}

// calculateExtrinsicMotivation 
func (ca *ContextAnalyzer) calculateExtrinsicMotivation(behaviorSummary *BehaviorSummary, preferences *UserPreferences) float64 {
	// 
	
	return 0.4 // ?
}

// assessSelfEfficacy 
func (ca *ContextAnalyzer) assessSelfEfficacy(behaviorSummary *BehaviorSummary) float64 {
	// 
	
	return 0.6 // ?
}

// assessPersistence ?
func (ca *ContextAnalyzer) assessPersistence(behaviorSummary *BehaviorSummary) float64 {
	// ?
	
	return 0.7 // ?
}

// assessCuriosityLevel ?
func (ca *ContextAnalyzer) assessCuriosityLevel(behaviorSummary *BehaviorSummary, preferences *UserPreferences) float64 {
	// ?
	
	return preferences.LearningPreferences.ExplorationTendency
}

// assessRewardSensitivity ?
func (ca *ContextAnalyzer) assessRewardSensitivity(behaviorSummary *BehaviorSummary) float64 {
	// ?
	
	return 0.5 // ?
}

// analyzeMotivationalFactors 
func (ca *ContextAnalyzer) analyzeMotivationalFactors(behaviorSummary *BehaviorSummary, preferences *UserPreferences) map[string]float64 {
	// 
	
	return map[string]float64{
		"achievement":    0.7,
		"autonomy":       0.6,
		"competence":     0.8,
		"relatedness":    0.5,
		"curiosity":      preferences.LearningPreferences.ExplorationTendency,
		"challenge":      preferences.LearningPreferences.ChallengeLevel,
	}
}

// calculateContextualFactors ?
func (ca *ContextAnalyzer) calculateContextualFactors(
	environmental *EnvironmentalContext,
	temporal *TemporalContext,
	social *SocialContext,
	device *DeviceContext,
	cognitive *CognitiveContext,
	motivational *MotivationalContext,
) map[string]float64 {
	factors := make(map[string]float64)
	
	// 
	factors["comfort"] = environmental.Comfort
	factors["productivity"] = environmental.Productivity
	factors["noise_level"] = 1.0 - environmental.NoiseLevel // 
	factors["light_level"] = environmental.LightLevel
	
	// 
	factors["energy_level"] = temporal.EnergyLevel
	factors["time_availability"] = float64(temporal.Duration.Minutes()) / 120.0 // 2
	
	// 罻
	factors["privacy"] = social.Privacy
	factors["support"] = social.Support
	factors["social_pressure"] = 1.0 - social.SocialPressure // 
	
	// 豸
	factors["device_performance"] = device.ProcessingPower
	factors["network_quality"] = device.NetworkQuality
	factors["battery_level"] = device.BatteryLevel
	
	// 
	factors["focus_level"] = cognitive.FocusLevel
	factors["cognitive_load"] = 1.0 - cognitive.CognitiveLoad // 
	factors["stress_level"] = 1.0 - cognitive.StressLevel     // 
	factors["confidence"] = cognitive.ConfidenceLevel
	
	// 
	factors["motivation"] = motivational.MotivationLevel
	factors["intrinsic_motivation"] = motivational.IntrinsicMotivation
	factors["self_efficacy"] = motivational.SelfEfficacy
	
	return factors
}

// calculateContextScore ?
func (ca *ContextAnalyzer) calculateContextScore(factors map[string]float64) float64 {
	// 
	weights := map[string]float64{
		"comfort":             0.1,
		"productivity":        0.15,
		"noise_level":         0.1,
		"light_level":         0.05,
		"energy_level":        0.15,
		"time_availability":   0.1,
		"privacy":             0.05,
		"support":             0.05,
		"social_pressure":     0.05,
		"device_performance":  0.1,
		"network_quality":     0.1,
		"battery_level":       0.05,
		"focus_level":         0.2,
		"cognitive_load":      0.15,
		"stress_level":        0.15,
		"confidence":          0.1,
		"motivation":          0.2,
		"intrinsic_motivation": 0.15,
		"self_efficacy":       0.1,
	}
	
	var totalScore float64
	var totalWeight float64
	
	for factor, value := range factors {
		if weight, exists := weights[factor]; exists {
			totalScore += value * weight
			totalWeight += weight
		}
	}
	
	if totalWeight > 0 {
		return totalScore / totalWeight
	}
	
	return 0.5 // 
}

// generateContextualRecommendations ?
func (ca *ContextAnalyzer) generateContextualRecommendations(
	environmental *EnvironmentalContext,
	temporal *TemporalContext,
	social *SocialContext,
	device *DeviceContext,
	cognitive *CognitiveContext,
	motivational *MotivationalContext,
	factors map[string]float64,
) []ContextualRecommendation {
	var recommendations []ContextualRecommendation
	
	// ?
	if environmental.NoiseLevel > 0.7 {
		recommendations = append(recommendations, ContextualRecommendation{
			Type:        "environment",
			Title:       "",
			Description: "?,
			Priority:    "high",
			Confidence:  0.8,
			Actions: []RecommendedAction{
				{
					Action:      "use_headphones",
					Description: "",
					Parameters:  map[string]interface{}{"type": "noise_cancelling"},
					Expected:    "?,
				},
			},
		})
	}
	
	// ?
	if temporal.EnergyLevel < 0.5 {
		recommendations = append(recommendations, ContextualRecommendation{
			Type:        "temporal",
			Title:       "",
			Description: "?,
			Priority:    "medium",
			Confidence:  0.7,
			Actions: []RecommendedAction{
				{
					Action:      "take_break",
					Description: "15-20",
					Parameters:  map[string]interface{}{"duration": "15m"},
					Expected:    "",
				},
			},
		})
	}
	
	// ?
	if cognitive.CognitiveLoad > 0.8 {
		recommendations = append(recommendations, ContextualRecommendation{
			Type:        "cognitive",
			Title:       "",
			Description: "",
			Priority:    "high",
			Confidence:  0.8,
			Actions: []RecommendedAction{
				{
					Action:      "simplify_content",
					Description: "",
					Parameters:  map[string]interface{}{"difficulty": "easy"},
					Expected:    "",
				},
			},
		})
	}
	
	// ?
	if motivational.MotivationLevel < 0.5 {
		recommendations = append(recommendations, ContextualRecommendation{
			Type:        "motivational",
			Title:       "",
			Description: "趨",
			Priority:    "medium",
			Confidence:  0.6,
			Actions: []RecommendedAction{
				{
					Action:      "set_small_goals",
					Description: "趨",
					Parameters:  map[string]interface{}{"goal_size": "small"},
					Expected:    "?,
				},
			},
		})
	}
	
	// 
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Priority != recommendations[j].Priority {
			priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
			return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
		}
		return recommendations[i].Confidence > recommendations[j].Confidence
	})
	
	return recommendations
}

// convertAirQualityToFloat 
func (ca *ContextAnalyzer) convertAirQualityToFloat(airQuality string) float64 {
	switch airQuality {
	case "", "excellent":
		return 1.0
	case "", "good":
		return 0.8
	case "", "moderate":
		return 0.6
	case "?, "poor":
		return 0.4
	case "", "very poor":
		return 0.2
	default:
		return 0.6 // 
	}
}

