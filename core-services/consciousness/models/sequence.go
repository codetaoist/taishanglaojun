package models

import (
	"time"
)

// SequenceRequest S轴能力序列请?
type SequenceRequest struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	EntityID        string                 `json:"entity_id" gorm:"index"`
	RequestType     string                 `json:"request_type"`
	CurrentSequence int                    `json:"current_sequence"`
	TargetSequence  int                    `json:"target_sequence"`
	Capabilities    map[string]float64     `json:"capabilities" gorm:"type:json"`
	Requirements    []string               `json:"requirements" gorm:"type:json"`
	Constraints     []string               `json:"constraints" gorm:"type:json"`
	Priority        string                 `json:"priority"`
	Context         map[string]interface{} `json:"context" gorm:"type:json"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// CapabilityEvaluation 能力评估
type CapabilityEvaluation struct {
	ID              string                     `json:"id" gorm:"primaryKey"`
	EntityID        string                     `json:"entity_id" gorm:"index"`
	Capability      string                     `json:"capability"`
	CurrentLevel    float64                    `json:"current_level"`
	MaxPotential    float64                    `json:"max_potential"`
	GrowthRate      float64                    `json:"growth_rate"`
	Bottlenecks     []string                   `json:"bottlenecks" gorm:"type:json"`
	Strengths       []string                   `json:"strengths" gorm:"type:json"`
	Weaknesses      []string                   `json:"weaknesses" gorm:"type:json"`
	Recommendations []CapabilityRecommendation `json:"recommendations" gorm:"type:json"`
	EvaluationScore float64                    `json:"evaluation_score"`
	Confidence      float64                    `json:"confidence"`
	Metadata        map[string]interface{}     `json:"metadata" gorm:"type:json"`
	EvaluatedAt     time.Time                  `json:"evaluated_at"`
}

// CapabilityRecommendation 能力推荐
type CapabilityRecommendation struct {
	Capability         string        `json:"capability"`
	CurrentLevel       float64       `json:"current_level"`
	TargetLevel        float64       `json:"target_level"`
	Priority           string        `json:"priority"`
	ImprovementMethods []string      `json:"improvement_methods"`
	EstimatedTime      time.Duration `json:"estimated_time"`
	Difficulty         string        `json:"difficulty"`
	Prerequisites      []string      `json:"prerequisites"`
}

// SequenceOptimization 序列优化
type SequenceOptimization struct {
	ID                 string             `json:"id" gorm:"primaryKey"`
	EntityID           string             `json:"entity_id" gorm:"index"`
	CurrentSequence    int                `json:"current_sequence"`
	TargetSequence     int                `json:"target_sequence"`
	OptimizationSteps  []OptimizationStep `json:"optimization_steps" gorm:"type:json"`
	NextMilestone      SequenceMilestone  `json:"next_milestone" gorm:"type:json"`
	RequiredEfforts    []RequiredEffort   `json:"required_efforts" gorm:"type:json"`
	EstimatedDuration  time.Duration      `json:"estimated_duration"`
	SuccessProbability float64            `json:"success_probability"`
	RiskFactors        []string           `json:"risk_factors" gorm:"type:json"`
	Recommendations    []string           `json:"recommendations" gorm:"type:json"`
	OptimizedAt        time.Time          `json:"optimized_at"`
}

// OptimizationStep 优化步骤
type OptimizationStep struct {
	StepID          string                 `json:"step_id"`
	Description     string                 `json:"description"`
	Action          string                 `json:"action"`
	ExpectedOutcome string                 `json:"expected_outcome"`
	Duration        time.Duration          `json:"duration"`
	Priority        int                    `json:"priority"`
	Dependencies    []string               `json:"dependencies"`
	Resources       map[string]interface{} `json:"resources"`
}

// SequenceMilestone 序列里程?
type SequenceMilestone struct {
	MilestoneID   string                 `json:"milestone_id"`
	SequenceLevel int                    `json:"sequence_level"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Criteria      []string               `json:"criteria"`
	Rewards       []string               `json:"rewards"`
	Difficulty    string                 `json:"difficulty"`
	EstimatedTime time.Duration          `json:"estimated_time"`
	Prerequisites []string               `json:"prerequisites"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// RequiredEffort 所需努力
type RequiredEffort struct {
	EffortID    string                 `json:"effort_id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Intensity   string                 `json:"intensity"`
	Duration    time.Duration          `json:"duration"`
	Frequency   string                 `json:"frequency"`
	Resources   []string               `json:"resources"`
	Skills      []string               `json:"skills"`
	Difficulty  float64                `json:"difficulty"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SequencePrediction 序列预测
type SequencePrediction struct {
	ID                string               `json:"id" gorm:"primaryKey"`
	EntityID          string               `json:"entity_id" gorm:"index"`
	CurrentSequence   int                  `json:"current_sequence"`
	PredictedSequence int                  `json:"predicted_sequence"`
	TimeHorizon       time.Duration        `json:"time_horizon"`
	Scenarios         []PredictionScenario `json:"scenarios" gorm:"type:json"`
	ConfidenceScore   float64              `json:"confidence_score"`
	Factors           []string             `json:"factors" gorm:"type:json"`
	Assumptions       []string             `json:"assumptions" gorm:"type:json"`
	Limitations       []string             `json:"limitations" gorm:"type:json"`
	PredictedAt       time.Time            `json:"predicted_at"`
}

// PredictionScenario 预测场景
type PredictionScenario struct {
	ScenarioID    string                 `json:"scenario_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Probability   float64                `json:"probability"`
	Outcome       int                    `json:"outcome"`
	Timeline      time.Duration          `json:"timeline"`
	Conditions    []string               `json:"conditions"`
	Risks         []string               `json:"risks"`
	Opportunities []string               `json:"opportunities"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// SequenceRequirements 序列要求
type SequenceRequirements struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	SequenceLevel   int                    `json:"sequence_level" gorm:"index"`
	MinCapabilities map[string]float64     `json:"min_capabilities" gorm:"type:json"`
	RequiredSkills  []string               `json:"required_skills" gorm:"type:json"`
	Prerequisites   []string               `json:"prerequisites" gorm:"type:json"`
	Challenges      []string               `json:"challenges" gorm:"type:json"`
	Opportunities   []string               `json:"opportunities" gorm:"type:json"`
	EstimatedTime   time.Duration          `json:"estimated_time"`
	DifficultyLevel string                 `json:"difficulty_level"`
	SuccessRate     float64                `json:"success_rate"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

