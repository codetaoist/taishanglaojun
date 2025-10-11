package models

import (
	"time"
)

// SynergyOpportunity 协同机会
type SynergyOpportunity struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Coordinate  *Coordinate            `json:"coordinate"`
	Potential   float64                `json:"potential"`
	Feasibility float64                `json:"feasibility"`
	Priority    int                    `json:"priority"`
	Conditions  []SynergyCondition     `json:"conditions"`
	Benefits    []string               `json:"benefits"`
	Risks       []string               `json:"risks"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// SynergyCondition 协同条件
type SynergyCondition struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Value       interface{} `json:"value"`
	Status      string      `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SynergyResult 协同结果
type SynergyResult struct {
	OpportunityID   string                 `json:"opportunity_id"`
	Success         bool                   `json:"success"`
	EffectivenessScore float64             `json:"effectiveness_score"`
	Outcomes        []SynergyOutcome       `json:"outcomes"`
	Improvements    []SynergyImprovement   `json:"improvements"`
	SideEffects     []SynergySideEffect    `json:"side_effects"`
	Lessons         []string               `json:"lessons"`
	Metadata        map[string]interface{} `json:"metadata"`
	CompletedAt     time.Time              `json:"completed_at"`
}

// SynergyOutcome 协同结果
type SynergyOutcome struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
	Measurement string  `json:"measurement"`
	Evidence    []string `json:"evidence"`
}

// SynergyImprovement 协同改进
type SynergyImprovement struct {
	Area        string  `json:"area"`
	Description string  `json:"description"`
	Magnitude   float64 `json:"magnitude"`
	Confidence  float64 `json:"confidence"`
	Evidence    []string `json:"evidence"`
}

// SynergySideEffect 协同副作�?
type SynergySideEffect struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    float64 `json:"severity"`
	Mitigation  []string `json:"mitigation"`
}

// SynergyMeasurement 协同测量
type SynergyMeasurement struct {
	ResultID        string                 `json:"result_id"`
	Metrics         []SynergyMetric        `json:"metrics"`
	OverallScore    float64                `json:"overall_score"`
	Benchmarks      []SynergyBenchmark     `json:"benchmarks"`
	Trends          []SynergyTrend         `json:"trends"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
	MeasuredAt      time.Time              `json:"measured_at"`
}

// SynergyMetric 协同指标
type SynergyMetric struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Target      float64 `json:"target"`
	Threshold   float64 `json:"threshold"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}

// SynergyBenchmark 协同基准
type SynergyBenchmark struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Comparison  string  `json:"comparison"`
	Difference  float64 `json:"difference"`
	Significance string `json:"significance"`
}

// SynergyTrend 协同趋势
type SynergyTrend struct {
	Metric      string    `json:"metric"`
	Direction   string    `json:"direction"`
	Rate        float64   `json:"rate"`
	Confidence  float64   `json:"confidence"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
}

// SynergyOptimization 协同优化
type SynergyOptimization struct {
	Conditions          []SynergyCondition     `json:"conditions"`
	OptimizedConditions []SynergyCondition     `json:"optimized_conditions"`
	Improvements        []OptimizationChange   `json:"improvements"`
	ExpectedBenefit     float64                `json:"expected_benefit"`
	ImplementationPlan  []SynergyOptimizationStep     `json:"implementation_plan"`
	Metadata            map[string]interface{} `json:"metadata"`
	OptimizedAt         time.Time              `json:"optimized_at"`
}

// OptimizationChange 优化变更
type OptimizationChange struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
	Effort      float64 `json:"effort"`
	Priority    int     `json:"priority"`
}

// SynergyOptimizationStep 协同优化步骤
type SynergyOptimizationStep struct {
	Order       int       `json:"order"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Duration    time.Duration `json:"duration"`
	Dependencies []string  `json:"dependencies"`
	Resources   []string  `json:"resources"`
}

// SynergyScenario 协同场景
type SynergyScenario struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Conditions  []SynergyCondition     `json:"conditions"`
	Variables   []ScenarioVariable     `json:"variables"`
	Constraints []Constraint           `json:"constraints"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ScenarioVariable 场景变量
type ScenarioVariable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	Range       []interface{} `json:"range"`
	Description string      `json:"description"`
}

// SynergyPrediction 协同预测
type SynergyPrediction struct {
	ScenarioID      string                 `json:"scenario_id"`
	PredictedResult *SynergyResult         `json:"predicted_result"`
	Confidence      float64                `json:"confidence"`
	Assumptions     []string               `json:"assumptions"`
	Uncertainties   []PredictionUncertainty `json:"uncertainties"`
	Alternatives    []AlternativePrediction `json:"alternatives"`
	Metadata        map[string]interface{} `json:"metadata"`
	PredictedAt     time.Time              `json:"predicted_at"`
}

// PredictionUncertainty 预测不确定�?
type PredictionUncertainty struct {
	Factor      string  `json:"factor"`
	Impact      float64 `json:"impact"`
	Probability float64 `json:"probability"`
	Description string  `json:"description"`
	Mitigation  []string `json:"mitigation"`
}

// SynergyCatalysis 协同催化
type SynergyCatalysis struct {
	CatalysisID      string                 `json:"catalysis_id"`
	OpportunityID    string                 `json:"opportunity_id"`
	CatalystTypes    []string               `json:"catalyst_types"`
	ActivationLevel  float64                `json:"activation_level"`
	CatalysisResult  *CatalysisResult       `json:"catalysis_result"`
	Effectiveness    float64                `json:"effectiveness"`
	Duration         time.Duration          `json:"duration"`
	SideEffects      []SynergySideEffect    `json:"side_effects"`
	Improvements     []SynergyImprovement   `json:"improvements"`
	Metadata         map[string]interface{} `json:"metadata"`
	CatalyzedAt      time.Time              `json:"catalyzed_at"`
}

// SynergyPotentialAnalysis 协同潜力分析
type SynergyPotentialAnalysis struct {
	AnalysisID       string                 `json:"analysis_id"`
	AxisResults      []interface{}          `json:"axis_results"`
	PotentialScore   float64                `json:"potential_score"`
	Opportunities    []SynergyOpportunity   `json:"opportunities"`
	Constraints      []string               `json:"constraints"`
	Recommendations  []string               `json:"recommendations"`
	RiskFactors      []string               `json:"risk_factors"`
	SuccessFactors   []string               `json:"success_factors"`
	Timeline         time.Duration          `json:"timeline"`
	Resources        []string               `json:"resources"`
	Metadata         map[string]interface{} `json:"metadata"`
	AnalyzedAt       time.Time              `json:"analyzed_at"`
}

// SynergyContext 协同上下�?
type SynergyContext struct {
	ContextID        string                 `json:"context_id"`
	SessionID        string                 `json:"session_id"`
	Environment      map[string]interface{} `json:"environment"`
	Participants     []string               `json:"participants"`
	Resources        []string               `json:"resources"`
	Constraints      []string               `json:"constraints"`
	Objectives       []string               `json:"objectives"`
	CurrentState     map[string]interface{} `json:"current_state"`
	TargetState      map[string]interface{} `json:"target_state"`
	Opportunities    []SynergyOpportunity   `json:"opportunities"`
	Catalysts        []Catalyst             `json:"catalysts"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
}

// CatalystOptimization 催化剂优�?
type CatalystOptimization struct {
	OptimizationID   string                 `json:"optimization_id"`
	SynergyContextID string                 `json:"synergy_context_id"`
	CurrentCatalysts []Catalyst             `json:"current_catalysts"`
	OptimalCatalysts []Catalyst             `json:"optimal_catalysts"`
	OptimizationStrategy string             `json:"optimization_strategy"`
	ExpectedBenefit  float64                `json:"expected_benefit"`
	ImplementationSteps []string            `json:"implementation_steps"`
	RiskLevel        string                 `json:"risk_level"`
	Timeline         time.Duration          `json:"timeline"`
	Resources        []string               `json:"resources"`
	Metadata         map[string]interface{} `json:"metadata"`
	OptimizedAt      time.Time              `json:"optimized_at"`
}

// Catalyst 催化�?
type Catalyst struct {
	CatalystID       string                 `json:"catalyst_id"`
	Type             string                 `json:"type"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Properties       map[string]interface{} `json:"properties"`
	ActivationLevel  float64                `json:"activation_level"`
	Effectiveness    float64                `json:"effectiveness"`
	Efficiency       float64                `json:"efficiency"`
	Stability        float64                `json:"stability"`
	Lifetime         time.Duration          `json:"lifetime"`
	Interactions     []string               `json:"interactions"`
	Requirements     []string               `json:"requirements"`
	SideEffects      []string               `json:"side_effects"`
	Status           string                 `json:"status"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	ActivatedAt      *time.Time             `json:"activated_at"`
	DeactivatedAt    *time.Time             `json:"deactivated_at"`
}

// CatalystEffectivenessReport 催化剂效果报�?
type CatalystEffectivenessReport struct {
	ReportID         string                 `json:"report_id"`
	SessionID        string                 `json:"session_id"`
	ActiveCatalysts  []Catalyst             `json:"active_catalysts"`
	OverallEffectiveness float64            `json:"overall_effectiveness"`
	IndividualScores map[string]float64     `json:"individual_scores"`
	Interactions     []CatalystInteraction  `json:"interactions"`
	Improvements     []string               `json:"improvements"`
	Issues           []string               `json:"issues"`
	Recommendations  []string               `json:"recommendations"`
	Trends           []string               `json:"trends"`
	Metadata         map[string]interface{} `json:"metadata"`
	GeneratedAt      time.Time              `json:"generated_at"`
}

// CatalystInteraction 催化剂交�?
type CatalystInteraction struct {
	InteractionID    string                 `json:"interaction_id"`
	CatalystA        string                 `json:"catalyst_a"`
	CatalystB        string                 `json:"catalyst_b"`
	InteractionType  string                 `json:"interaction_type"`
	Strength         float64                `json:"strength"`
	Effect           string                 `json:"effect"`
	Outcome          string                 `json:"outcome"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CatalysisResult 催化结果
type CatalysisResult struct {
	ResultID         string                 `json:"result_id"`
	Success          bool                   `json:"success"`
	AmplificationFactor float64             `json:"amplification_factor"`
	EfficiencyGain   float64                `json:"efficiency_gain"`
	QualityImprovement float64              `json:"quality_improvement"`
	Outcomes         []SynergyOutcome       `json:"outcomes"`
	SideEffects      []SynergySideEffect    `json:"side_effects"`
	Measurements     *SynergyMeasurement    `json:"measurements"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// AlternativePrediction 替代预测
type AlternativePrediction struct {
	Name        string         `json:"name"`
	Result      *SynergyResult `json:"result"`
	Probability float64        `json:"probability"`
	Conditions  []string       `json:"conditions"`
	Description string         `json:"description"`
}

// CatalystRecommendation 催化剂推�?
type CatalystRecommendation struct {
	RecommendationID   string                 `json:"recommendation_id"`
	CatalystID         string                 `json:"catalyst_id"`
	RecommendationType string                 `json:"recommendation_type"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	Priority           string                 `json:"priority"`
	ExpectedImpact     float64                `json:"expected_impact"`
	ImplementationSteps []string              `json:"implementation_steps"`
	Confidence         float64                `json:"confidence"`
	RiskLevel          string                 `json:"risk_level"`
	Timeline           time.Duration          `json:"timeline"`
	Resources          []string               `json:"resources"`
	Metadata           map[string]interface{} `json:"metadata"`
	CreatedAt          time.Time              `json:"created_at"`
}
