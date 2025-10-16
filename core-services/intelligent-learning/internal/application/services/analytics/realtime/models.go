package realtime

import (
	"time"
	"github.com/google/uuid"
)

// =============================================================================
// 
// =============================================================================

// RealtimeLearningState ?- 
type RealtimeLearningState struct {
	LearnerID           uuid.UUID                      `json:"learner_id"`           // ID
	CurrentSession      *LearningSession               `json:"current_session"`      // 
	EngagementLevel     float64                        `json:"engagement_level"`     // ?
	ComprehensionLevel  float64                        `json:"comprehension_level"`  // ?
	MotivationLevel     float64                        `json:"motivation_level"`     // 
	FatigueLevel        float64                        `json:"fatigue_level"`        // ?
	EmotionalState      string                         `json:"emotional_state"`      // ?
	LearningVelocity    float64                        `json:"learning_velocity"`    // 
	DifficultyPreference float64                       `json:"difficulty_preference"` // 
	AttentionSpan       time.Duration                  `json:"attention_span"`       // ?
	InteractionPatterns map[string]interface{}         `json:"interaction_patterns"` // 
	PerformanceMetrics  *RealtimePerformanceMetrics    `json:"performance_metrics"`  // 
	Timestamp           time.Time                      `json:"timestamp"`            // ?
}

// LearningSession  - 
type LearningSession struct {
	SessionID   uuid.UUID              `json:"session_id"`   // ID
	StartTime   time.Time              `json:"start_time"`   // ?
	Duration    time.Duration          `json:"duration"`     // 
	ContentID   uuid.UUID              `json:"content_id"`   // ID
	Progress    float64                `json:"progress"`     // 
	Interactions []InteractionEvent    `json:"interactions"` // 
	Metadata    map[string]interface{} `json:"metadata"`     // ?
}

// LearningInsight  - 
type LearningInsight struct {
	InsightID   uuid.UUID              `json:"insight_id"`   // ID
	Type        InsightType            `json:"type"`         // 
	Title       string                 `json:"title"`        // 
	Description string                 `json:"description"`  // 
	Confidence  float64                `json:"confidence"`   // ?
	Impact      ImpactLevel            `json:"impact"`       // 
	Evidence    []string               `json:"evidence"`     // 
	Timestamp   time.Time              `json:"timestamp"`    // ?
	Metadata    map[string]interface{} `json:"metadata"`     // actionable
}

// LearningPattern  - 
type LearningPattern struct {
	PatternID       uuid.UUID                  `json:"pattern_id"`       // ID
	LearnerID       uuid.UUID                  `json:"learner_id"`       // ID
	Type            LearningPatternType        `json:"type"`             // 
	Characteristics *PatternCharacteristics    `json:"characteristics"`  // 
	Frequency       float64                   `json:"frequency"`        // 
	Strength        float64                   `json:"strength"`         // 
	Stability       float64                   `json:"stability"`        // ?
	Adaptability    float64                   `json:"adaptability"`     // ?
	Effectiveness   float64                   `json:"effectiveness"`    // ?
	Evolution       []*PatternEvolution       `json:"evolution"`        // 
	Predictions     []*PatternPrediction      `json:"predictions"`      // 
	Recommendations []*PatternRecommendation  `json:"recommendations"`  // 
	LastUpdated     time.Time                 `json:"last_updated"`     // ?
	Metadata        map[string]interface{}     `json:"metadata"`         // ?
}

// PatternEvolution  - 
type PatternEvolution struct {
	Timestamp   time.Time                  `json:"timestamp"`   // ?
	Changes     []*PatternChange           `json:"changes"`     // 仯
	Triggers    []*EvolutionTrigger        `json:"triggers"`    // ?
	Impact      float64                   `json:"impact"`      // 
	Confidence  float64                   `json:"confidence"`  // ?
	Description string                     `json:"description"` // 
	Metadata    map[string]interface{}     `json:"metadata"`    // ?
}

// EmotionalProfile  - 
type EmotionalProfile struct {
	CurrentMood     string                 `json:"current_mood"`     // 
	FocusLevel      float64                `json:"focus_level"`      // ?
	StressLevel     float64                `json:"stress_level"`     // 
	MotivationLevel float64                `json:"motivation_level"` // 
	PreferredTone   string                 `json:"preferred_tone"`   // 
	EmotionalNeeds  []string               `json:"emotional_needs"`  // ?
	LastUpdated     time.Time              `json:"last_updated"`     // ?
}

// PredictionRecommendation  - 
type PredictionRecommendation struct {
	RecommendationID uuid.UUID              `json:"recommendation_id"` // ID
	Type            RecommendationType      `json:"type"`              // 
	Priority        PriorityLevel           `json:"priority"`          // ?
	Title           string                  `json:"title"`             // 
	Description     string                  `json:"description"`       // 
	Actions         []string                `json:"actions"`           // ?
	ExpectedOutcome string                  `json:"expected_outcome"`  // 
	Confidence      float64                 `json:"confidence"`        // ?
	Timestamp       time.Time               `json:"timestamp"`         // ?
	Metadata        map[string]interface{}  `json:"metadata"`          // Category, ExpectedImpact, Timeline, Status?
}

// PredictionValidation  - 
type PredictionValidation struct {
	ValidationID uuid.UUID              `json:"validation_id"` // ID
	Method       ValidationMethod       `json:"method"`        // 
	Score        float64                `json:"score"`         // 
	Metrics      map[string]float64     `json:"metrics"`       // 
	Timestamp    time.Time              `json:"timestamp"`     // ?
	Metadata     map[string]interface{} `json:"metadata"`      // is_valid, issues, suggestions?
}

// =============================================================================
// 
// =============================================================================

// InsightType 
type InsightType string

const (
	InsightTypePerformance InsightType = "performance" // 
	InsightTypeEngagement  InsightType = "engagement"  // ?
	InsightTypeBehavior    InsightType = "behavior"    // 
	InsightTypeEmotional   InsightType = "emotional"   // 
	InsightTypePredictive  InsightType = "predictive"  // 
)

// ImpactLevel 
type ImpactLevel string

const (
	ImpactLevelLow      ImpactLevel = "low"      // ?
	ImpactLevelMedium   ImpactLevel = "medium"   // 
	ImpactLevelHigh     ImpactLevel = "high"     // ?
	ImpactLevelCritical ImpactLevel = "critical" // 
)

// RecommendationType 
type RecommendationType string

const (
	RecommendationTypeContent     RecommendationType = "content"     // 
	RecommendationTypePacing      RecommendationType = "pacing"      // 
	RecommendationTypeDifficulty  RecommendationType = "difficulty"  // 
	RecommendationTypeMotivation  RecommendationType = "motivation"  // 
	RecommendationTypeIntervention RecommendationType = "intervention" // 
	RecommendationTypeStrategy    RecommendationType = "strategy"    // 
	RecommendationTypePath        RecommendationType = "path"        // 
	RecommendationTypeResource    RecommendationType = "resource"    // 
	RecommendationTypePeer        RecommendationType = "peer"        // 齨
	RecommendationTypeOptimization RecommendationType = "optimization" // 
)

// PriorityLevel ?
type PriorityLevel string

const (
	PriorityLevelLow      PriorityLevel = "low"      // 
	PriorityLevelMedium   PriorityLevel = "medium"   // ?
	PriorityLevelHigh     PriorityLevel = "high"     // 
	PriorityLevelUrgent   PriorityLevel = "urgent"   // ?
)

// ValidationMethod 
type ValidationMethod string

const (
	ValidationMethodCrossValidation ValidationMethod = "cross_validation" // 
	ValidationMethodHoldout        ValidationMethod = "holdout"          // 
	ValidationMethodBootstrap      ValidationMethod = "bootstrap"        // 
	ValidationMethodTimeSeriesSplit ValidationMethod = "time_series_split" // 
)

// =============================================================================
// ?
// =============================================================================

// InteractionEvent 
type InteractionEvent struct {
	EventID   uuid.UUID              `json:"event_id"`   // ID
	Type      InteractionType        `json:"type"`       // 
	Timestamp time.Time              `json:"timestamp"`  // ?
	Duration  time.Duration          `json:"duration"`   // 
	Context   map[string]interface{} `json:"context"`    // ?
}

// PatternChange 仯
type PatternChange struct {
	Aspect       string      `json:"aspect"`       // 
	OldValue     interface{} `json:"old_value"`    // ?
	NewValue     interface{} `json:"new_value"`    // ?
	Magnitude    float64     `json:"magnitude"`    // 仯
	Direction    string      `json:"direction"`    // 
	Significance float64     `json:"significance"` // ?
}

// EvolutionTrigger ?
type EvolutionTrigger struct {
	TriggerID   uuid.UUID              `json:"trigger_id"`   // ID
	Type        string                 `json:"type"`         // ?
	Description string                 `json:"description"`  // 
	Strength    float64                `json:"strength"`     // 
	Metadata    map[string]interface{} `json:"metadata"`     // ?
}

// PatternPrediction 
type PatternPrediction struct {
	PredictionID uuid.UUID              `json:"prediction_id"` // ID
	Horizon      time.Duration          `json:"horizon"`       // 
	Confidence   float64                `json:"confidence"`    // ?
	Outcome      interface{}            `json:"outcome"`       // 
	Metadata     map[string]interface{} `json:"metadata"`      // ?
}

// PatternRecommendation 
type PatternRecommendation struct {
	RecommendationID uuid.UUID              `json:"recommendation_id"` // ID
	Type            RecommendationType      `json:"type"`              // 
	Priority        PriorityLevel           `json:"priority"`          // ?
	Description     string                  `json:"description"`       // 
	Actions         []string                `json:"actions"`           // ?
	Confidence      float64                 `json:"confidence"`        // ?
	Metadata        map[string]interface{}  `json:"metadata"`          // ?
}

