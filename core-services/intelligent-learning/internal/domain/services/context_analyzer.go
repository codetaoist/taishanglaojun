package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// ContextAnalyzer 上下文分析器
type ContextAnalyzer struct {
	behaviorTracker   *UserBehaviorTracker
	preferenceAnalyzer *PreferenceAnalyzer
	environmentRepo   EnvironmentRepository
	locationService   LocationService
	weatherService    WeatherService
}

// EnvironmentRepository 环境数据仓库接口
type EnvironmentRepository interface {
	SaveEnvironmentData(ctx context.Context, data *EnvironmentData) error
	GetEnvironmentData(ctx context.Context, userID string, timeRange ContextTimeRange) ([]*EnvironmentData, error)
	GetContextHistory(ctx context.Context, userID string, limit int) ([]*ContextRecord, error)
}

// LocationService 位置服务接口
type LocationService interface {
	GetCurrentLocation(ctx context.Context, userID string) (*Location, error)
	GetLocationHistory(ctx context.Context, userID string, timeRange ContextTimeRange) ([]*LocationRecord, error)
	GetLocationContext(ctx context.Context, location *Location) (*LocationContext, error)
}

// WeatherService 天气服务接口
type WeatherService interface {
	GetCurrentWeather(ctx context.Context, location *Location) (*WeatherInfo, error)
	GetWeatherForecast(ctx context.Context, location *Location, days int) ([]*WeatherInfo, error)
}

// NewContextAnalyzer 创建上下文分析器
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

// LearningContext 学习上下文
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

// EnvironmentalContext 环境上下文
type EnvironmentalContext struct {
	Location        *Location     `json:"location"`
	Weather         *WeatherInfo  `json:"weather"`
	NoiseLevel      float64       `json:"noise_level"`      // 噪音水平 0-1
	LightLevel      float64       `json:"light_level"`      // 光照水平 0-1
	Temperature     float64       `json:"temperature"`      // 温度
	Humidity        float64       `json:"humidity"`         // 湿度
	AirQuality      float64       `json:"air_quality"`      // 空气质量 0-1
	Distractions    []Distraction `json:"distractions"`     // 干扰因素
	Comfort         float64       `json:"comfort"`          // 舒适度 0-1
	Productivity    float64       `json:"productivity"`     // 生产力指数 0-1
}

// TemporalContext 时间上下文
type TemporalContext struct {
	TimeOfDay       string        `json:"time_of_day"`       // 一天中的时间段
	DayOfWeek       string        `json:"day_of_week"`       // 星期几
	Season          string        `json:"season"`            // 季节
	Holiday         bool          `json:"holiday"`           // 是否节假日
	WorkingHours    bool          `json:"working_hours"`     // 是否工作时间
	PeakHours       bool          `json:"peak_hours"`        // 是否高峰时间
	TimeZone        string        `json:"timezone"`          // 时区
	LocalTime       time.Time     `json:"local_time"`        // 本地时间
	Duration        time.Duration `json:"duration"`          // 可用学习时长
	Urgency         float64       `json:"urgency"`           // 紧急程度 0-1
	EnergyLevel     float64       `json:"energy_level"`      // 精力水平 0-1
	AttentionSpan   time.Duration `json:"attention_span"`    // 注意力持续时间
}

// SocialContext 社交上下文
type SocialContext struct {
	SocialSetting   string             `json:"social_setting"`   // 社交环境：独处、小组、公共场所等
	GroupSize       int                `json:"group_size"`       // 群体大小
	SocialPressure  float64            `json:"social_pressure"`  // 社交压力 0-1
	Collaboration   bool               `json:"collaboration"`    // 是否协作学习
	Competition     bool               `json:"competition"`      // 是否竞争环境
	Support         float64            `json:"support"`          // 支持程度 0-1
	Privacy         float64            `json:"privacy"`          // 隐私程度 0-1
	SocialInfluence map[string]float64 `json:"social_influence"` // 社交影响因素
}

// DeviceContext 设备上下文
type DeviceContext struct {
	DeviceType      string  `json:"device_type"`      // 设备类型
	ScreenSize      string  `json:"screen_size"`      // 屏幕尺寸
	BatteryLevel    float64 `json:"battery_level"`    // 电池电量 0-1
	NetworkQuality  float64 `json:"network_quality"`  // 网络质量 0-1
	StorageSpace    float64 `json:"storage_space"`    // 存储空间 0-1
	ProcessingPower float64 `json:"processing_power"` // 处理能力 0-1
	InputMethod     string  `json:"input_method"`     // 输入方式
	Orientation     string  `json:"orientation"`      // 设备方向
	Accessibility   bool    `json:"accessibility"`    // 是否启用无障碍功能
}

// CognitiveContext 认知上下文
type CognitiveContext struct {
	CognitiveLoad    float64           `json:"cognitive_load"`    // 认知负荷 0-1
	FocusLevel       float64           `json:"focus_level"`       // 专注程度 0-1
	StressLevel      float64           `json:"stress_level"`      // 压力水平 0-1
	FatigueLevel     float64           `json:"fatigue_level"`     // 疲劳程度 0-1
	MoodState        string            `json:"mood_state"`        // 情绪状态
	ConfidenceLevel  float64           `json:"confidence_level"`  // 自信程度 0-1
	LearningState    string            `json:"learning_state"`    // 学习状态
	PriorKnowledge   float64           `json:"prior_knowledge"`   // 先验知识 0-1
	MemoryCapacity   float64           `json:"memory_capacity"`   // 记忆容量 0-1
	ProcessingSpeed  float64           `json:"processing_speed"`  // 处理速度 0-1
	MetacognitionLevel float64         `json:"metacognition_level"` // 元认知水平 0-1
	CognitiveStyle   map[string]float64 `json:"cognitive_style"`   // 认知风格
}

// MotivationalContext 动机上下文
type MotivationalContext struct {
	MotivationLevel  float64            `json:"motivation_level"`  // 动机水平 0-1
	GoalOrientation  string             `json:"goal_orientation"`  // 目标导向
	IntrinsicMotivation float64         `json:"intrinsic_motivation"` // 内在动机 0-1
	ExtrinsicMotivation float64         `json:"extrinsic_motivation"` // 外在动机 0-1
	SelfEfficacy     float64            `json:"self_efficacy"`     // 自我效能 0-1
	Persistence      float64            `json:"persistence"`       // 坚持性 0-1
	CuriosityLevel   float64            `json:"curiosity_level"`   // 好奇心水平 0-1
	RewardSensitivity float64           `json:"reward_sensitivity"` // 奖励敏感性 0-1
	MotivationalFactors map[string]float64 `json:"motivational_factors"` // 动机因素
}

// Location 位置信息
type Location struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	PlaceType   string  `json:"place_type"`   // 地点类型：家、办公室、咖啡厅等
	Familiarity float64 `json:"familiarity"`  // 熟悉程度 0-1
}

// WeatherInfo 天气信息
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

// Distraction 干扰因素
type Distraction struct {
	Type      string  `json:"type"`      // 干扰类型
	Intensity float64 `json:"intensity"` // 强度 0-1
	Duration  time.Duration `json:"duration"` // 持续时间
	Source    string  `json:"source"`    // 来源
}

// ContextualRecommendation 上下文推荐
type ContextualRecommendation struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Confidence  float64                `json:"confidence"`
	Context     map[string]interface{} `json:"context"`
	Actions     []RecommendedAction    `json:"actions"`
}

// RecommendedAction 推荐行动
type RecommendedAction struct {
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Expected    string                 `json:"expected"`
}

// EnvironmentData 环境数据
type EnvironmentData struct {
	UserID      string                 `json:"user_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Location    *Location              `json:"location"`
	Weather     *WeatherInfo           `json:"weather"`
	DeviceInfo  *DeviceContext         `json:"device_info"`
	Environment map[string]interface{} `json:"environment"`
}

// ContextRecord 上下文记录
type ContextRecord struct {
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Context   *LearningContext       `json:"context"`
	Outcome   map[string]interface{} `json:"outcome"`
}

// LocationRecord 位置记录
type LocationRecord struct {
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Location  *Location `json:"location"`
	Duration  time.Duration `json:"duration"`
}

// LocationContext 位置上下文
type LocationContext struct {
	PlaceType     string  `json:"place_type"`
	NoiseLevel    float64 `json:"noise_level"`
	CrowdLevel    float64 `json:"crowd_level"`
	WiFiQuality   float64 `json:"wifi_quality"`
	Comfort       float64 `json:"comfort"`
	Productivity  float64 `json:"productivity"`
	Accessibility bool    `json:"accessibility"`
}

// ContextTimeRange 上下文时间范围
type ContextTimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AnalyzeLearningContext 分析学习上下文
func (ca *ContextAnalyzer) AnalyzeLearningContext(ctx context.Context, userID string) (*LearningContext, error) {
	timestamp := time.Now()
	
	// 分析环境上下文
	environmentalContext, err := ca.analyzeEnvironmentalContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze environmental context: %w", err)
	}
	
	// 分析时间上下文
	temporalContext := ca.analyzeTemporalContext(timestamp)
	
	// 分析社交上下文
	socialContext, err := ca.analyzeSocialContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze social context: %w", err)
	}
	
	// 分析设备上下文
	deviceContext, err := ca.analyzeDeviceContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze device context: %w", err)
	}
	
	// 分析认知上下文
	cognitiveContext, err := ca.analyzeCognitiveContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze cognitive context: %w", err)
	}
	
	// 分析动机上下文
	motivationalContext, err := ca.analyzeMotivationalContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze motivational context: %w", err)
	}
	
	// 计算上下文因素
	contextualFactors := ca.calculateContextualFactors(
		environmentalContext,
		temporalContext,
		socialContext,
		deviceContext,
		cognitiveContext,
		motivationalContext,
	)
	
	// 计算上下文分数
	contextScore := ca.calculateContextScore(contextualFactors)
	
	// 生成上下文推荐
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

// analyzeEnvironmentalContext 分析环境上下文
func (ca *ContextAnalyzer) analyzeEnvironmentalContext(ctx context.Context, userID string) (*EnvironmentalContext, error) {
	// 获取当前位置
	location, err := ca.locationService.GetCurrentLocation(ctx, userID)
	if err != nil {
		// 使用默认位置或历史位置
		location = &Location{
			PlaceType:   "unknown",
			Familiarity: 0.5,
		}
	}
	
	// 获取位置上下文
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
	
	// 获取天气信息
	weather, err := ca.weatherService.GetCurrentWeather(ctx, location)
	if err != nil {
		weather = &WeatherInfo{
			Temperature: 22.0,
			Humidity:    50.0,
			Condition:   "unknown",
			AirQuality:  "良好",
		}
	}
	
	// 分析干扰因素
	distractions := ca.analyzeDistractions(locationContext, weather)
	
	// 计算舒适度和生产力
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

// analyzeTemporalContext 分析时间上下文
func (ca *ContextAnalyzer) analyzeTemporalContext(timestamp time.Time) *TemporalContext {
	// 分析时间段
	timeOfDay := ca.getTimeOfDay(timestamp)
	
	// 分析星期几
	dayOfWeek := timestamp.Weekday().String()
	
	// 分析季节
	season := ca.getSeason(timestamp)
	
	// 判断是否节假日
	holiday := ca.isHoliday(timestamp)
	
	// 判断是否工作时间
	workingHours := ca.isWorkingHours(timestamp)
	
	// 判断是否高峰时间
	peakHours := ca.isPeakHours(timestamp)
	
	// 计算精力水平
	energyLevel := ca.calculateEnergyLevel(timestamp)
	
	// 估算注意力持续时间
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
		Duration:      2 * time.Hour, // 默认可用时长
		Urgency:       0.5,           // 默认紧急程度
		EnergyLevel:   energyLevel,
		AttentionSpan: attentionSpan,
	}
}

// analyzeSocialContext 分析社交上下文
func (ca *ContextAnalyzer) analyzeSocialContext(ctx context.Context, userID string) (*SocialContext, error) {
	// 获取社交环境信息
	// 这里可以从用户行为、位置等推断社交环境
	
	socialSetting := "private" // 默认私人环境
	groupSize := 1             // 默认独自学习
	socialPressure := 0.3      // 默认低社交压力
	collaboration := false     // 默认非协作
	competition := false       // 默认非竞争
	support := 0.6            // 默认中等支持
	privacy := 0.8            // 默认高隐私
	
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

// analyzeDeviceContext 分析设备上下文
func (ca *ContextAnalyzer) analyzeDeviceContext(ctx context.Context, userID string) (*DeviceContext, error) {
	// 从用户行为数据中获取设备信息
	behaviorSummary, err := ca.behaviorTracker.GetBehaviorSummary(ctx, userID, time.Now().AddDate(0, 0, -1), time.Now())
	if err != nil {
		// 使用默认设备上下文
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
	
	// 从行为数据推断设备上下文
	deviceType := ca.inferDeviceType(behaviorSummary)
	screenSize := ca.inferScreenSize(deviceType)
	inputMethod := ca.inferInputMethod(deviceType)
	
	return &DeviceContext{
		DeviceType:      deviceType,
		ScreenSize:      screenSize,
		BatteryLevel:    0.8, // 默认值，实际应从设备API获取
		NetworkQuality:  0.8, // 默认值，实际应从网络状态获取
		StorageSpace:    0.7, // 默认值，实际应从设备API获取
		ProcessingPower: 0.8, // 默认值，实际应从设备性能获取
		InputMethod:     inputMethod,
		Orientation:     "landscape", // 默认值
		Accessibility:   false,       // 默认值
	}, nil
}

// analyzeCognitiveContext 分析认知上下文
func (ca *ContextAnalyzer) analyzeCognitiveContext(ctx context.Context, userID string) (*CognitiveContext, error) {
	// 获取用户学习历史和行为数据
	behaviorSummary, err := ca.behaviorTracker.GetBehaviorSummary(ctx, userID, time.Now().AddDate(0, 0, -7), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}
	
	// 分析认知负荷
	cognitiveLoad := ca.calculateCognitiveLoad(behaviorSummary)
	
	// 分析专注程度
	focusLevel := ca.calculateFocusLevel(behaviorSummary)
	
	// 分析压力水平
	stressLevel := ca.calculateStressLevel(behaviorSummary)
	
	// 分析疲劳程度
	fatigueLevel := ca.calculateFatigueLevel(behaviorSummary)
	
	// 推断情绪状态
	moodState := ca.inferMoodState(behaviorSummary)
	
	// 计算自信程度
	confidenceLevel := ca.calculateConfidenceLevel(behaviorSummary)
	
	// 推断学习状态
	learningState := ca.inferLearningState(behaviorSummary)
	
	// 评估先验知识
	priorKnowledge := ca.assessPriorKnowledge(ctx, userID)
	
	// 评估记忆容量
	memoryCapacity := ca.assessMemoryCapacity(behaviorSummary)
	
	// 评估处理速度
	processingSpeed := ca.assessProcessingSpeed(behaviorSummary)
	
	// 评估元认知水平
	metacognitionLevel := ca.assessMetacognitionLevel(behaviorSummary)
	
	// 分析认知风格
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

// analyzeMotivationalContext 分析动机上下文
func (ca *ContextAnalyzer) analyzeMotivationalContext(ctx context.Context, userID string) (*MotivationalContext, error) {
	// 获取用户偏好和行为数据
	preferences, err := ca.preferenceAnalyzer.AnalyzeUserPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze user preferences: %w", err)
	}
	
	behaviorSummary, err := ca.behaviorTracker.GetBehaviorSummary(ctx, userID, time.Now().AddDate(0, 0, -7), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}
	
	// 计算动机水平
	motivationLevel := ca.calculateMotivationLevel(behaviorSummary, preferences)
	
	// 推断目标导向
	goalOrientation := ca.inferGoalOrientation(preferences)
	
	// 分析内在动机
	intrinsicMotivation := ca.calculateIntrinsicMotivation(behaviorSummary, preferences)
	
	// 分析外在动机
	extrinsicMotivation := ca.calculateExtrinsicMotivation(behaviorSummary, preferences)
	
	// 评估自我效能
	selfEfficacy := ca.assessSelfEfficacy(behaviorSummary)
	
	// 评估坚持性
	persistence := ca.assessPersistence(behaviorSummary)
	
	// 评估好奇心水平
	curiosityLevel := ca.assessCuriosityLevel(behaviorSummary, preferences)
	
	// 评估奖励敏感性
	rewardSensitivity := ca.assessRewardSensitivity(behaviorSummary)
	
	// 分析动机因素
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

// 辅助方法

// analyzeDistractions 分析干扰因素
func (ca *ContextAnalyzer) analyzeDistractions(locationContext *LocationContext, weather *WeatherInfo) []Distraction {
	var distractions []Distraction
	
	// 基于噪音水平添加干扰
	if locationContext.NoiseLevel > 0.7 {
		distractions = append(distractions, Distraction{
			Type:      "noise",
			Intensity: locationContext.NoiseLevel,
			Duration:  time.Hour,
			Source:    "environment",
		})
	}
	
	// 基于人群水平添加干扰
	if locationContext.CrowdLevel > 0.6 {
		distractions = append(distractions, Distraction{
			Type:      "crowd",
			Intensity: locationContext.CrowdLevel,
			Duration:  time.Hour,
			Source:    "social",
		})
	}
	
	// 基于天气条件添加干扰
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

// calculateComfort 计算舒适度
func (ca *ContextAnalyzer) calculateComfort(location *Location, weather *WeatherInfo, locationContext *LocationContext) float64 {
	comfort := 0.5 // 基础舒适度
	
	// 位置熟悉度影响舒适度
	comfort += location.Familiarity * 0.3
	
	// 温度影响舒适度
	tempComfort := 1.0 - math.Abs(weather.Temperature-22.0)/20.0 // 22度为最舒适温度
	comfort += tempComfort * 0.2
	
	// 湿度影响舒适度
	humidityComfort := 1.0 - math.Abs(weather.Humidity-50.0)/50.0 // 50%为最舒适湿度
	comfort += humidityComfort * 0.1
	
	// 位置舒适度
	comfort += locationContext.Comfort * 0.4
	
	return math.Max(0.0, math.Min(1.0, comfort))
}

// calculateProductivity 计算生产力
func (ca *ContextAnalyzer) calculateProductivity(location *Location, weather *WeatherInfo, locationContext *LocationContext, distractions []Distraction) float64 {
	productivity := 0.7 // 基础生产力
	
	// 位置生产力
	productivity += locationContext.Productivity * 0.3
	
	// 网络质量影响生产力
	productivity += locationContext.WiFiQuality * 0.2
	
	// 干扰因素降低生产力
	for _, distraction := range distractions {
		productivity -= distraction.Intensity * 0.1
	}
	
	// 空气质量影响生产力
	productivity += ca.convertAirQualityToFloat(weather.AirQuality) * 0.1
	
	return math.Max(0.0, math.Min(1.0, productivity))
}

// calculateLightLevel 计算光照水平
func (ca *ContextAnalyzer) calculateLightLevel(weather *WeatherInfo) float64 {
	// 基于云层覆盖和时间计算光照水平
	lightLevel := 1.0 - weather.CloudCover
	
	// 根据时间调整
	hour := time.Now().Hour()
	if hour < 6 || hour > 20 {
		lightLevel *= 0.1 // 夜间光照很低
	} else if hour < 8 || hour > 18 {
		lightLevel *= 0.5 // 早晚光照较低
	}
	
	return math.Max(0.0, math.Min(1.0, lightLevel))
}

// getTimeOfDay 获取时间段
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

// getSeason 获取季节
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

// isHoliday 判断是否节假日
func (ca *ContextAnalyzer) isHoliday(timestamp time.Time) bool {
	// 简单实现，实际应查询节假日数据库
	weekday := timestamp.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// isWorkingHours 判断是否工作时间
func (ca *ContextAnalyzer) isWorkingHours(timestamp time.Time) bool {
	hour := timestamp.Hour()
	weekday := timestamp.Weekday()
	
	return weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour < 18
}

// isPeakHours 判断是否高峰时间
func (ca *ContextAnalyzer) isPeakHours(timestamp time.Time) bool {
	hour := timestamp.Hour()
	
	// 上下班高峰时间
	return (hour >= 7 && hour <= 9) || (hour >= 17 && hour <= 19)
}

// calculateEnergyLevel 计算精力水平
func (ca *ContextAnalyzer) calculateEnergyLevel(timestamp time.Time) float64 {
	hour := timestamp.Hour()
	
	// 基于生物钟的精力水平
	switch {
	case hour >= 9 && hour <= 11:
		return 0.9 // 上午精力最佳
	case hour >= 14 && hour <= 16:
		return 0.8 // 下午精力较好
	case hour >= 19 && hour <= 21:
		return 0.7 // 晚上精力中等
	case hour >= 6 && hour <= 8:
		return 0.6 // 早晨精力一般
	case hour >= 22 || hour <= 5:
		return 0.3 // 深夜精力很低
	default:
		return 0.5 // 其他时间精力中等
	}
}

// estimateAttentionSpan 估算注意力持续时间
func (ca *ContextAnalyzer) estimateAttentionSpan(timestamp time.Time, energyLevel float64) time.Duration {
	// 基础注意力持续时间
	baseSpan := 25 * time.Minute // 番茄工作法的基础时间
	
	// 根据精力水平调整
	adjustedSpan := time.Duration(float64(baseSpan) * energyLevel)
	
	// 根据时间段调整
	hour := timestamp.Hour()
	if hour >= 22 || hour <= 6 {
		adjustedSpan = adjustedSpan / 2 // 深夜注意力持续时间减半
	}
	
	return adjustedSpan
}

// inferDeviceType 推断设备类型
func (ca *ContextAnalyzer) inferDeviceType(behaviorSummary *BehaviorSummary) string {
	// 基于行为模式推断设备类型
	// 这里需要根据实际的行为数据来推断
	
	return "desktop" // 默认值
}

// inferScreenSize 推断屏幕尺寸
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

// inferInputMethod 推断输入方式
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

// calculateCognitiveLoad 计算认知负荷
func (ca *ContextAnalyzer) calculateCognitiveLoad(behaviorSummary *BehaviorSummary) float64 {
	// 基于学习活动的复杂性和频率计算认知负荷
	
	return 0.6 // 默认值
}

// calculateFocusLevel 计算专注程度
func (ca *ContextAnalyzer) calculateFocusLevel(behaviorSummary *BehaviorSummary) float64 {
	// 基于会话持续时间和交互模式计算专注程度
	
	return 0.7 // 默认值
}

// calculateStressLevel 计算压力水平
func (ca *ContextAnalyzer) calculateStressLevel(behaviorSummary *BehaviorSummary) float64 {
	// 基于学习表现和行为模式计算压力水平
	
	return 0.4 // 默认值
}

// calculateFatigueLevel 计算疲劳程度
func (ca *ContextAnalyzer) calculateFatigueLevel(behaviorSummary *BehaviorSummary) float64 {
	// 基于学习时长和表现变化计算疲劳程度
	
	return 0.3 // 默认值
}

// inferMoodState 推断情绪状态
func (ca *ContextAnalyzer) inferMoodState(behaviorSummary *BehaviorSummary) string {
	// 基于行为模式推断情绪状态
	
	return "neutral" // 默认值
}

// calculateConfidenceLevel 计算自信程度
func (ca *ContextAnalyzer) calculateConfidenceLevel(behaviorSummary *BehaviorSummary) float64 {
	// 基于学习表现和行为模式计算自信程度
	
	return 0.6 // 默认值
}

// inferLearningState 推断学习状态
func (ca *ContextAnalyzer) inferLearningState(behaviorSummary *BehaviorSummary) string {
	// 基于当前学习活动推断学习状态
	
	return "active" // 默认值
}

// assessPriorKnowledge 评估先验知识
func (ca *ContextAnalyzer) assessPriorKnowledge(ctx context.Context, userID string) float64 {
	// 基于学习历史评估先验知识水平
	
	return 0.5 // 默认值
}

// assessMemoryCapacity 评估记忆容量
func (ca *ContextAnalyzer) assessMemoryCapacity(behaviorSummary *BehaviorSummary) float64 {
	// 基于学习表现评估记忆容量
	
	return 0.7 // 默认值
}

// assessProcessingSpeed 评估处理速度
func (ca *ContextAnalyzer) assessProcessingSpeed(behaviorSummary *BehaviorSummary) float64 {
	// 基于学习速度评估处理速度
	
	return 0.6 // 默认值
}

// assessMetacognitionLevel 评估元认知水平
func (ca *ContextAnalyzer) assessMetacognitionLevel(behaviorSummary *BehaviorSummary) float64 {
	// 基于自我调节行为评估元认知水平
	
	return 0.5 // 默认值
}

// analyzeCognitiveStyle 分析认知风格
func (ca *ContextAnalyzer) analyzeCognitiveStyle(behaviorSummary *BehaviorSummary) map[string]float64 {
	// 分析认知风格特征
	
	return map[string]float64{
		"field_dependent":   0.4,
		"field_independent": 0.6,
		"sequential":        0.5,
		"random":           0.5,
		"concrete":         0.6,
		"abstract":         0.4,
	}
}

// calculateMotivationLevel 计算动机水平
func (ca *ContextAnalyzer) calculateMotivationLevel(behaviorSummary *BehaviorSummary, preferences *UserPreferences) float64 {
	// 基于学习行为和偏好计算动机水平
	
	return 0.7 // 默认值
}

// inferGoalOrientation 推断目标导向
func (ca *ContextAnalyzer) inferGoalOrientation(preferences *UserPreferences) string {
	// 基于偏好推断目标导向
	
	return "mastery" // 默认值：掌握导向
}

// calculateIntrinsicMotivation 计算内在动机
func (ca *ContextAnalyzer) calculateIntrinsicMotivation(behaviorSummary *BehaviorSummary, preferences *UserPreferences) float64 {
	// 基于自主学习行为计算内在动机
	
	return 0.6 // 默认值
}

// calculateExtrinsicMotivation 计算外在动机
func (ca *ContextAnalyzer) calculateExtrinsicMotivation(behaviorSummary *BehaviorSummary, preferences *UserPreferences) float64 {
	// 基于奖励驱动行为计算外在动机
	
	return 0.4 // 默认值
}

// assessSelfEfficacy 评估自我效能
func (ca *ContextAnalyzer) assessSelfEfficacy(behaviorSummary *BehaviorSummary) float64 {
	// 基于学习表现和挑战接受度评估自我效能
	
	return 0.6 // 默认值
}

// assessPersistence 评估坚持性
func (ca *ContextAnalyzer) assessPersistence(behaviorSummary *BehaviorSummary) float64 {
	// 基于学习持续性评估坚持性
	
	return 0.7 // 默认值
}

// assessCuriosityLevel 评估好奇心水平
func (ca *ContextAnalyzer) assessCuriosityLevel(behaviorSummary *BehaviorSummary, preferences *UserPreferences) float64 {
	// 基于探索行为评估好奇心水平
	
	return preferences.LearningPreferences.ExplorationTendency
}

// assessRewardSensitivity 评估奖励敏感性
func (ca *ContextAnalyzer) assessRewardSensitivity(behaviorSummary *BehaviorSummary) float64 {
	// 基于对奖励的反应评估奖励敏感性
	
	return 0.5 // 默认值
}

// analyzeMotivationalFactors 分析动机因素
func (ca *ContextAnalyzer) analyzeMotivationalFactors(behaviorSummary *BehaviorSummary, preferences *UserPreferences) map[string]float64 {
	// 分析各种动机因素
	
	return map[string]float64{
		"achievement":    0.7,
		"autonomy":       0.6,
		"competence":     0.8,
		"relatedness":    0.5,
		"curiosity":      preferences.LearningPreferences.ExplorationTendency,
		"challenge":      preferences.LearningPreferences.ChallengeLevel,
	}
}

// calculateContextualFactors 计算上下文因素
func (ca *ContextAnalyzer) calculateContextualFactors(
	environmental *EnvironmentalContext,
	temporal *TemporalContext,
	social *SocialContext,
	device *DeviceContext,
	cognitive *CognitiveContext,
	motivational *MotivationalContext,
) map[string]float64 {
	factors := make(map[string]float64)
	
	// 环境因素
	factors["comfort"] = environmental.Comfort
	factors["productivity"] = environmental.Productivity
	factors["noise_level"] = 1.0 - environmental.NoiseLevel // 噪音越低越好
	factors["light_level"] = environmental.LightLevel
	
	// 时间因素
	factors["energy_level"] = temporal.EnergyLevel
	factors["time_availability"] = float64(temporal.Duration.Minutes()) / 120.0 // 归一化到2小时
	
	// 社交因素
	factors["privacy"] = social.Privacy
	factors["support"] = social.Support
	factors["social_pressure"] = 1.0 - social.SocialPressure // 压力越低越好
	
	// 设备因素
	factors["device_performance"] = device.ProcessingPower
	factors["network_quality"] = device.NetworkQuality
	factors["battery_level"] = device.BatteryLevel
	
	// 认知因素
	factors["focus_level"] = cognitive.FocusLevel
	factors["cognitive_load"] = 1.0 - cognitive.CognitiveLoad // 负荷越低越好
	factors["stress_level"] = 1.0 - cognitive.StressLevel     // 压力越低越好
	factors["confidence"] = cognitive.ConfidenceLevel
	
	// 动机因素
	factors["motivation"] = motivational.MotivationLevel
	factors["intrinsic_motivation"] = motivational.IntrinsicMotivation
	factors["self_efficacy"] = motivational.SelfEfficacy
	
	return factors
}

// calculateContextScore 计算上下文分数
func (ca *ContextAnalyzer) calculateContextScore(factors map[string]float64) float64 {
	// 权重配置
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
	
	return 0.5 // 默认分数
}

// generateContextualRecommendations 生成上下文推荐
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
	
	// 基于环境因素的推荐
	if environmental.NoiseLevel > 0.7 {
		recommendations = append(recommendations, ContextualRecommendation{
			Type:        "environment",
			Title:       "降低环境噪音",
			Description: "当前环境噪音较高，建议使用耳机或寻找更安静的学习环境",
			Priority:    "high",
			Confidence:  0.8,
			Actions: []RecommendedAction{
				{
					Action:      "use_headphones",
					Description: "使用降噪耳机",
					Parameters:  map[string]interface{}{"type": "noise_cancelling"},
					Expected:    "提高专注度",
				},
			},
		})
	}
	
	// 基于时间因素的推荐
	if temporal.EnergyLevel < 0.5 {
		recommendations = append(recommendations, ContextualRecommendation{
			Type:        "temporal",
			Title:       "调整学习时间",
			Description: "当前精力水平较低，建议选择轻松的学习内容或休息一下",
			Priority:    "medium",
			Confidence:  0.7,
			Actions: []RecommendedAction{
				{
					Action:      "take_break",
					Description: "休息15-20分钟",
					Parameters:  map[string]interface{}{"duration": "15m"},
					Expected:    "恢复精力",
				},
			},
		})
	}
	
	// 基于认知因素的推荐
	if cognitive.CognitiveLoad > 0.8 {
		recommendations = append(recommendations, ContextualRecommendation{
			Type:        "cognitive",
			Title:       "降低认知负荷",
			Description: "当前认知负荷较高，建议简化学习内容或分解学习任务",
			Priority:    "high",
			Confidence:  0.8,
			Actions: []RecommendedAction{
				{
					Action:      "simplify_content",
					Description: "选择更简单的学习内容",
					Parameters:  map[string]interface{}{"difficulty": "easy"},
					Expected:    "减轻认知压力",
				},
			},
		})
	}
	
	// 基于动机因素的推荐
	if motivational.MotivationLevel < 0.5 {
		recommendations = append(recommendations, ContextualRecommendation{
			Type:        "motivational",
			Title:       "提升学习动机",
			Description: "当前学习动机较低，建议设定小目标或选择感兴趣的内容",
			Priority:    "medium",
			Confidence:  0.6,
			Actions: []RecommendedAction{
				{
					Action:      "set_small_goals",
					Description: "设定容易达成的小目标",
					Parameters:  map[string]interface{}{"goal_size": "small"},
					Expected:    "增强成就感",
				},
			},
		})
	}
	
	// 排序推荐（按优先级和置信度）
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Priority != recommendations[j].Priority {
			priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
			return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
		}
		return recommendations[i].Confidence > recommendations[j].Confidence
	})
	
	return recommendations
}

// convertAirQualityToFloat 将空气质量字符串转换为浮点数
func (ca *ContextAnalyzer) convertAirQualityToFloat(airQuality string) float64 {
	switch airQuality {
	case "优秀", "excellent":
		return 1.0
	case "良好", "good":
		return 0.8
	case "中等", "moderate":
		return 0.6
	case "差", "poor":
		return 0.4
	case "很差", "very poor":
		return 0.2
	default:
		return 0.6 // 默认中等
	}
}