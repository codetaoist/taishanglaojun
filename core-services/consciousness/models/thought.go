package models

import (
	"time"
)

// ThoughtRequest T轴思维层请求
type ThoughtRequest struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	EntityID         string                 `json:"entity_id" gorm:"index"`
	RequestType      string                 `json:"request_type"`
	ThoughtContent   string                 `json:"thought_content"`
	Context          map[string]interface{} `json:"context" gorm:"type:json"`
	TargetDepth      int                    `json:"target_depth"`
	Requirements     []string               `json:"requirements" gorm:"type:json"`
	Constraints      []string               `json:"constraints" gorm:"type:json"`
	Priority         string                 `json:"priority"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// Thought 思维
type Thought struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	Content          string                 `json:"content"`
	Type             string                 `json:"type"`
	Depth            int                    `json:"depth"`
	Complexity       float64                `json:"complexity"`
	Clarity          float64                `json:"clarity"`
	Coherence        float64                `json:"coherence"`
	Relations        []ThoughtRelation      `json:"relations" gorm:"type:json"`
	Patterns         []string               `json:"patterns" gorm:"type:json"`
	Insights         []string               `json:"insights" gorm:"type:json"`
	Limitations      []string               `json:"limitations" gorm:"type:json"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ThoughtRelation 思维关系
type ThoughtRelation struct {
	RelationID   string                 `json:"relation_id"`
	SourceID     string                 `json:"source_id"`
	TargetID     string                 `json:"target_id"`
	Type         string                 `json:"type"`
	Strength     float64                `json:"strength"`
	Direction    string                 `json:"direction"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ThoughtDepth 思维深度
type ThoughtDepth struct {
	Level            int                    `json:"level"`
	Description      string                 `json:"description"`
	Characteristics  []string               `json:"characteristics"`
	RequiredSkills   []string               `json:"required_skills"`
	Complexity       float64                `json:"complexity"`
	TimeRequired     time.Duration          `json:"time_required"`
	Prerequisites    []string               `json:"prerequisites"`
	Outcomes         []string               `json:"outcomes"`
	Limitations      []string               `json:"limitations"`
	NextLevels       []int                  `json:"next_levels"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ThoughtDepthEvaluation 思维深度评估
type ThoughtDepthEvaluation struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	ThoughtID        string                 `json:"thought_id" gorm:"index"`
	CurrentDepth     int                    `json:"current_depth"`
	MaxPossibleDepth int                    `json:"max_possible_depth"`
	DepthScore       float64                `json:"depth_score"`
	Dimensions       []DepthDimension       `json:"dimensions" gorm:"type:json"`
	Barriers         []string               `json:"barriers" gorm:"type:json"`
	Opportunities    []string               `json:"opportunities" gorm:"type:json"`
	Recommendations  []string               `json:"recommendations" gorm:"type:json"`
	EvaluatedAt      time.Time              `json:"evaluated_at"`
}

// DepthDimension 深度维度
type DepthDimension struct {
	Name         string                 `json:"name"`
	Score        float64                `json:"score"`
	MaxScore     float64                `json:"max_score"`
	Description  string                 `json:"description"`
	Indicators   []string               `json:"indicators"`
	Improvements []string               `json:"improvements"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ThoughtPatternAnalysis 思维模式分析
type ThoughtPatternAnalysis struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	EntityID         string                 `json:"entity_id" gorm:"index"`
	AnalysisType     string                 `json:"analysis_type"`
	Patterns         []ThoughtPattern       `json:"patterns" gorm:"type:json"`
	Trends           []ThoughtTrend         `json:"trends" gorm:"type:json"`
	Anomalies        []ThoughtAnomaly       `json:"anomalies" gorm:"type:json"`
	Insights         []string               `json:"insights" gorm:"type:json"`
	Recommendations  []string               `json:"recommendations" gorm:"type:json"`
	AnalyzedAt       time.Time              `json:"analyzed_at"`
}

// ThoughtPattern 思维模式
type ThoughtPattern struct {
	PatternID    string                 `json:"pattern_id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Frequency    float64                `json:"frequency"`
	Strength     float64                `json:"strength"`
	Triggers     []string               `json:"triggers"`
	Outcomes     []string               `json:"outcomes"`
	Variations   []string               `json:"variations"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ThoughtTrend 思维趋势
type ThoughtTrend struct {
	TrendID      string                 `json:"trend_id"`
	Name         string                 `json:"name"`
	Direction    string                 `json:"direction"`
	Magnitude    float64                `json:"magnitude"`
	Duration     time.Duration          `json:"duration"`
	Confidence   float64                `json:"confidence"`
	Factors      []string               `json:"factors"`
	Predictions  []string               `json:"predictions"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ThoughtAnomaly 思维异常
type ThoughtAnomaly struct {
	AnomalyID    string                 `json:"anomaly_id"`
	Type         string                 `json:"type"`
	Severity     string                 `json:"severity"`
	Description  string                 `json:"description"`
	Indicators   []string               `json:"indicators"`
	PossibleCauses []string             `json:"possible_causes"`
	Impact       string                 `json:"impact"`
	Recommendations []string            `json:"recommendations"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ThoughtLimitation 思维限制
type ThoughtLimitation struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	ThoughtID        string                 `json:"thought_id" gorm:"index"`
	LimitationType   string                 `json:"limitation_type"`
	Description      string                 `json:"description"`
	Severity         string                 `json:"severity"`
	Impact           string                 `json:"impact"`
	Causes           []string               `json:"causes" gorm:"type:json"`
	Symptoms         []string               `json:"symptoms" gorm:"type:json"`
	Solutions        []string               `json:"solutions" gorm:"type:json"`
	Workarounds      []string               `json:"workarounds" gorm:"type:json"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	IdentifiedAt     time.Time              `json:"identified_at"`
}

// TranscendenceResult 超越结果
type TranscendenceResult struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	ThoughtID        string                 `json:"thought_id" gorm:"index"`
	TranscendenceType string                `json:"transcendence_type"`
	OriginalState    map[string]interface{} `json:"original_state" gorm:"type:json"`
	TranscendedState map[string]interface{} `json:"transcended_state" gorm:"type:json"`
	Aspects          []TranscendedAspect    `json:"aspects" gorm:"type:json"`
	Breakthrough     string                 `json:"breakthrough"`
	NewCapabilities  []string               `json:"new_capabilities" gorm:"type:json"`
	Insights         []string               `json:"insights" gorm:"type:json"`
	Implications     []string               `json:"implications" gorm:"type:json"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	AchievedAt       time.Time              `json:"achieved_at"`
}

// TranscendedAspect 超越方面
type TranscendedAspect struct {
	AspectName       string                 `json:"aspect_name"`
	OriginalLevel    float64                `json:"original_level"`
	TranscendedLevel float64                `json:"transcended_level"`
	Improvement      float64                `json:"improvement"`
	Description      string                 `json:"description"`
	Evidence         []string               `json:"evidence"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// Experience 体验
type Experience struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	EntityID         string                 `json:"entity_id" gorm:"index"`
	Type             string                 `json:"type"`
	Content          string                 `json:"content"`
	Intensity        float64                `json:"intensity"`
	Duration         time.Duration          `json:"duration"`
	Quality          string                 `json:"quality"`
	Context          map[string]interface{} `json:"context" gorm:"type:json"`
	Triggers         []string               `json:"triggers" gorm:"type:json"`
	Outcomes         []string               `json:"outcomes" gorm:"type:json"`
	Lessons          []string               `json:"lessons" gorm:"type:json"`
	Insights         []string               `json:"insights" gorm:"type:json"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	ExperiencedAt    time.Time              `json:"experienced_at"`
}

// WisdomCultivation 智慧培养
type WisdomCultivation struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	EntityID         string                 `json:"entity_id" gorm:"index"`
	CultivationType  string                 `json:"cultivation_type"`
	CurrentLevel     float64                `json:"current_level"`
	TargetLevel      float64                `json:"target_level"`
	Elements         []WisdomElement        `json:"elements" gorm:"type:json"`
	Practices        []string               `json:"practices" gorm:"type:json"`
	Progress         map[string]interface{} `json:"progress" gorm:"type:json"`
	Milestones       []string               `json:"milestones" gorm:"type:json"`
	Challenges       []string               `json:"challenges" gorm:"type:json"`
	Insights         []string               `json:"insights" gorm:"type:json"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	StartedAt        time.Time              `json:"started_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// WisdomElement 智慧元素
type WisdomElement struct {
	ElementName      string                 `json:"element_name"`
	Type             string                 `json:"type"`
	Level            float64                `json:"level"`
	MaxLevel         float64                `json:"max_level"`
	Description      string                 `json:"description"`
	Characteristics  []string               `json:"characteristics"`
	DevelopmentPath  []string               `json:"development_path"`
	Prerequisites    []string               `json:"prerequisites"`
	Indicators       []string               `json:"indicators"`
	Practices        []string               `json:"practices"`
	Metadata         map[string]interface{} `json:"metadata"`
}