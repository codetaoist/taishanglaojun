package realtime

import (
	"time"
	"github.com/google/uuid"
)

// =============================================================================
// 核心数据模型统一定义
// =============================================================================

// RealtimeLearningState 实时学习状态 - 标准定义
type RealtimeLearningState struct {
	LearnerID           uuid.UUID                      `json:"learner_id"`           // 学习者ID
	CurrentSession      *LearningSession               `json:"current_session"`      // 当前会话
	EngagementLevel     float64                        `json:"engagement_level"`     // 参与度
	ComprehensionLevel  float64                        `json:"comprehension_level"`  // 理解度
	MotivationLevel     float64                        `json:"motivation_level"`     // 动机水平
	FatigueLevel        float64                        `json:"fatigue_level"`        // 疲劳度
	EmotionalState      string                         `json:"emotional_state"`      // 情感状态
	LearningVelocity    float64                        `json:"learning_velocity"`    // 学习速度
	DifficultyPreference float64                       `json:"difficulty_preference"` // 难度偏好
	AttentionSpan       time.Duration                  `json:"attention_span"`       // 注意力持续时间
	InteractionPatterns map[string]interface{}         `json:"interaction_patterns"` // 交互模式
	PerformanceMetrics  *RealtimePerformanceMetrics    `json:"performance_metrics"`  // 性能指标
	Timestamp           time.Time                      `json:"timestamp"`            // 时间戳
}

// LearningSession 学习会话 - 标准定义
type LearningSession struct {
	SessionID   uuid.UUID              `json:"session_id"`   // 会话ID
	StartTime   time.Time              `json:"start_time"`   // 开始时间
	Duration    time.Duration          `json:"duration"`     // 持续时间
	ContentID   uuid.UUID              `json:"content_id"`   // 内容ID
	Progress    float64                `json:"progress"`     // 进度
	Interactions []InteractionEvent    `json:"interactions"` // 交互事件
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据
}

// LearningInsight 学习洞察 - 标准定义
type LearningInsight struct {
	InsightID   uuid.UUID              `json:"insight_id"`   // 洞察ID
	Type        InsightType            `json:"type"`         // 洞察类型
	Title       string                 `json:"title"`        // 标题
	Description string                 `json:"description"`  // 描述
	Confidence  float64                `json:"confidence"`   // 置信度
	Impact      ImpactLevel            `json:"impact"`       // 影响级别
	Evidence    []string               `json:"evidence"`     // 证据
	Timestamp   time.Time              `json:"timestamp"`    // 时间戳
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据（包含actionable等字段）
}

// LearningPattern 学习模式 - 标准定义
type LearningPattern struct {
	PatternID       uuid.UUID                  `json:"pattern_id"`       // 模式ID
	LearnerID       uuid.UUID                  `json:"learner_id"`       // 学习者ID
	Type            LearningPatternType        `json:"type"`             // 模式类型
	Characteristics *PatternCharacteristics    `json:"characteristics"`  // 特征
	Frequency       float64                   `json:"frequency"`        // 频率
	Strength        float64                   `json:"strength"`         // 强度
	Stability       float64                   `json:"stability"`        // 稳定性
	Adaptability    float64                   `json:"adaptability"`     // 适应性
	Effectiveness   float64                   `json:"effectiveness"`    // 有效性
	Evolution       []*PatternEvolution       `json:"evolution"`        // 演化
	Predictions     []*PatternPrediction      `json:"predictions"`      // 预测
	Recommendations []*PatternRecommendation  `json:"recommendations"`  // 建议
	LastUpdated     time.Time                 `json:"last_updated"`     // 最后更新
	Metadata        map[string]interface{}     `json:"metadata"`         // 元数据
}

// PatternEvolution 模式演化 - 标准定义
type PatternEvolution struct {
	Timestamp   time.Time                  `json:"timestamp"`   // 时间戳
	Changes     []*PatternChange           `json:"changes"`     // 变化
	Triggers    []*EvolutionTrigger        `json:"triggers"`    // 触发器
	Impact      float64                   `json:"impact"`      // 影响
	Confidence  float64                   `json:"confidence"`  // 置信度
	Description string                     `json:"description"` // 描述
	Metadata    map[string]interface{}     `json:"metadata"`    // 元数据
}

// EmotionalProfile 情感档案 - 标准定义
type EmotionalProfile struct {
	CurrentMood     string                 `json:"current_mood"`     // 当前情绪
	FocusLevel      float64                `json:"focus_level"`      // 专注度
	StressLevel     float64                `json:"stress_level"`     // 压力水平
	MotivationLevel float64                `json:"motivation_level"` // 动机水平
	PreferredTone   string                 `json:"preferred_tone"`   // 偏好语调
	EmotionalNeeds  []string               `json:"emotional_needs"`  // 情感需求
	LastUpdated     time.Time              `json:"last_updated"`     // 最后更新
}

// PredictionRecommendation 预测建议 - 标准定义
type PredictionRecommendation struct {
	RecommendationID uuid.UUID              `json:"recommendation_id"` // 建议ID
	Type            RecommendationType      `json:"type"`              // 建议类型
	Priority        PriorityLevel           `json:"priority"`          // 优先级
	Title           string                  `json:"title"`             // 标题
	Description     string                  `json:"description"`       // 描述
	Actions         []string                `json:"actions"`           // 行动项
	ExpectedOutcome string                  `json:"expected_outcome"`  // 预期结果
	Confidence      float64                 `json:"confidence"`        // 置信度
	Timestamp       time.Time               `json:"timestamp"`         // 时间戳
	Metadata        map[string]interface{}  `json:"metadata"`          // 元数据（包含Category, ExpectedImpact, Timeline, Status）
}

// PredictionValidation 预测验证 - 标准定义
type PredictionValidation struct {
	ValidationID uuid.UUID              `json:"validation_id"` // 验证ID
	Method       ValidationMethod       `json:"method"`        // 验证方法
	Score        float64                `json:"score"`         // 验证分数
	Metrics      map[string]float64     `json:"metrics"`       // 验证指标
	Timestamp    time.Time              `json:"timestamp"`     // 时间戳
	Metadata     map[string]interface{} `json:"metadata"`      // 元数据（包含is_valid, issues, suggestions）
}

// =============================================================================
// 枚举类型定义
// =============================================================================

// InsightType 洞察类型
type InsightType string

const (
	InsightTypePerformance InsightType = "performance" // 性能洞察
	InsightTypeEngagement  InsightType = "engagement"  // 参与度洞察
	InsightTypeBehavior    InsightType = "behavior"    // 行为洞察
	InsightTypeEmotional   InsightType = "emotional"   // 情感洞察
	InsightTypePredictive  InsightType = "predictive"  // 预测洞察
)

// ImpactLevel 影响级别
type ImpactLevel string

const (
	ImpactLevelLow      ImpactLevel = "low"      // 低影响
	ImpactLevelMedium   ImpactLevel = "medium"   // 中等影响
	ImpactLevelHigh     ImpactLevel = "high"     // 高影响
	ImpactLevelCritical ImpactLevel = "critical" // 关键影响
)

// RecommendationType 建议类型
type RecommendationType string

const (
	RecommendationTypeContent     RecommendationType = "content"     // 内容建议
	RecommendationTypePacing      RecommendationType = "pacing"      // 节奏建议
	RecommendationTypeDifficulty  RecommendationType = "difficulty"  // 难度建议
	RecommendationTypeMotivation  RecommendationType = "motivation"  // 动机建议
	RecommendationTypeIntervention RecommendationType = "intervention" // 干预建议
	RecommendationTypeStrategy    RecommendationType = "strategy"    // 策略建议
	RecommendationTypePath        RecommendationType = "path"        // 路径建议
	RecommendationTypeResource    RecommendationType = "resource"    // 资源建议
	RecommendationTypePeer        RecommendationType = "peer"        // 同伴建议
	RecommendationTypeOptimization RecommendationType = "optimization" // 优化建议
)

// PriorityLevel 优先级
type PriorityLevel string

const (
	PriorityLevelLow      PriorityLevel = "low"      // 低优先级
	PriorityLevelMedium   PriorityLevel = "medium"   // 中等优先级
	PriorityLevelHigh     PriorityLevel = "high"     // 高优先级
	PriorityLevelUrgent   PriorityLevel = "urgent"   // 紧急
)

// ValidationMethod 验证方法
type ValidationMethod string

const (
	ValidationMethodCrossValidation ValidationMethod = "cross_validation" // 交叉验证
	ValidationMethodHoldout        ValidationMethod = "holdout"          // 留出验证
	ValidationMethodBootstrap      ValidationMethod = "bootstrap"        // 自助验证
	ValidationMethodTimeSeriesSplit ValidationMethod = "time_series_split" // 时间序列分割
)

// =============================================================================
// 辅助结构体
// =============================================================================

// InteractionEvent 交互事件
type InteractionEvent struct {
	EventID   uuid.UUID              `json:"event_id"`   // 事件ID
	Type      InteractionType        `json:"type"`       // 交互类型
	Timestamp time.Time              `json:"timestamp"`  // 时间戳
	Duration  time.Duration          `json:"duration"`   // 持续时间
	Context   map[string]interface{} `json:"context"`    // 上下文
}

// PatternChange 模式变化
type PatternChange struct {
	Aspect       string      `json:"aspect"`       // 方面
	OldValue     interface{} `json:"old_value"`    // 旧值
	NewValue     interface{} `json:"new_value"`    // 新值
	Magnitude    float64     `json:"magnitude"`    // 变化幅度
	Direction    string      `json:"direction"`    // 方向
	Significance float64     `json:"significance"` // 显著性
}

// EvolutionTrigger 演化触发器
type EvolutionTrigger struct {
	TriggerID   uuid.UUID              `json:"trigger_id"`   // 触发器ID
	Type        string                 `json:"type"`         // 触发器类型
	Description string                 `json:"description"`  // 描述
	Strength    float64                `json:"strength"`     // 强度
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据
}

// PatternPrediction 模式预测
type PatternPrediction struct {
	PredictionID uuid.UUID              `json:"prediction_id"` // 预测ID
	Horizon      time.Duration          `json:"horizon"`       // 预测范围
	Confidence   float64                `json:"confidence"`    // 置信度
	Outcome      interface{}            `json:"outcome"`       // 预测结果
	Metadata     map[string]interface{} `json:"metadata"`      // 元数据
}

// PatternRecommendation 模式建议
type PatternRecommendation struct {
	RecommendationID uuid.UUID              `json:"recommendation_id"` // 建议ID
	Type            RecommendationType      `json:"type"`              // 建议类型
	Priority        PriorityLevel           `json:"priority"`          // 优先级
	Description     string                  `json:"description"`       // 描述
	Actions         []string                `json:"actions"`           // 行动项
	Confidence      float64                 `json:"confidence"`        // 置信度
	Metadata        map[string]interface{}  `json:"metadata"`          // 元数据
}