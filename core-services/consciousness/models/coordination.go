package models

import (
	"time"
)

// CoordinationSession 协调会话
type CoordinationSession struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	EntityID     string                 `json:"entity_id" gorm:"index"`
	SessionType  string                 `json:"session_type"`
	Status       CoordinationStatus     `json:"status"`
	Participants []string               `json:"participants" gorm:"type:json"`
	Objectives   []string               `json:"objectives" gorm:"type:json"`
	Steps        []CoordinationStep     `json:"steps" gorm:"type:json"`
	Issues       []CoordinationIssue    `json:"issues" gorm:"type:json"`
	Records      []CoordinationRecord   `json:"records" gorm:"type:json"`
	Results      CoordinationResult     `json:"results" gorm:"type:json"`
	Metadata     map[string]interface{} `json:"metadata" gorm:"type:json"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
}

// CoordinationResult 协调结果
type CoordinationResult struct {
	Success            bool                   `json:"success"`
	AchievedObjectives []string               `json:"achieved_objectives"`
	FailedObjectives   []string               `json:"failed_objectives"`
	QualityScore       float64                `json:"quality_score"`
	EfficiencyScore    float64                `json:"efficiency_score"`
	SatisfactionScore  float64                `json:"satisfaction_score"`
	Outcomes           []string               `json:"outcomes"`
	Improvements       []string               `json:"improvements"`
	LessonsLearned     []string               `json:"lessons_learned"`
	NextSteps          []string               `json:"next_steps"`
	Interactions       []AxisInteraction      `json:"interactions"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// ResolutionMethod 解决方法
type ResolutionMethod struct {
	Strategy    string                 `json:"strategy"`
	Description string                 `json:"description"`
	Steps       []string               `json:"steps"`
	Priority    float64                `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// OptimizationGoal 优化目标
type OptimizationGoal struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Priority    float64 `json:"priority"`
	Target      float64 `json:"target"`
	Current     float64 `json:"current"`
}

// OptimizationRisk 优化风险
type OptimizationRisk struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Probability float64 `json:"probability"`
	Impact      float64 `json:"impact"`
}

// AxisInteraction 轴交互
type AxisInteraction struct {
	InteractionID   string                 `json:"interaction_id"`
	SourceAxis      string                 `json:"source_axis"`
	TargetAxis      string                 `json:"target_axis"`
	InteractionType string                 `json:"interaction_type"`
	Strength        float64                `json:"strength"`
	Quality         float64                `json:"quality"`
	Direction       string                 `json:"direction"`
	Frequency       float64                `json:"frequency"`
	Duration        time.Duration          `json:"duration"`
	Analysis        InteractionAnalysis    `json:"analysis"`
	Outcomes        []string               `json:"outcomes"`
	Issues          []string               `json:"issues"`
	Improvements    []string               `json:"improvements"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// InteractionAnalysis 交互分析
type InteractionAnalysis struct {
	EffectivenessScore float64                `json:"effectiveness_score"`
	CompatibilityScore float64                `json:"compatibility_score"`
	SynergyLevel       float64                `json:"synergy_level"`
	ConflictLevel      float64                `json:"conflict_level"`
	BalanceScore       float64                `json:"balance_score"`
	Patterns           []string               `json:"patterns"`
	Trends             []string               `json:"trends"`
	Anomalies          []string               `json:"anomalies"`
	Recommendations    []string               `json:"recommendations"`
	Insights           []string               `json:"insights"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// AxisConflict 轴冲突
type AxisConflict struct {
	ConflictID   string                 `json:"conflict_id"`
	SourceAxis   string                 `json:"source_axis"`
	TargetAxis   string                 `json:"target_axis"`
	ConflictType string                 `json:"conflict_type"`
	Severity     string                 `json:"severity"`
	Description  string                 `json:"description"`
	Causes       []string               `json:"causes"`
	Impact       float64                `json:"impact"`
	Resolution   ConflictResolution     `json:"resolution"`
	Status       string                 `json:"status"`
	DetectedAt   time.Time              `json:"detected_at"`
	ResolvedAt   *time.Time             `json:"resolved_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ConflictResolution 冲突解决
type ConflictResolution struct {
	ResolutionID    string                 `json:"resolution_id"`
	Strategy        string                 `json:"strategy"`
	Description     string                 `json:"description"`
	Steps           []string               `json:"steps"`
	ExpectedOutcome string                 `json:"expected_outcome"`
	Success         bool                   `json:"success"`
	ActualOutcome   string                 `json:"actual_outcome"`
	LessonsLearned  []string               `json:"lessons_learned"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// CoordinationState 协调状态
type CoordinationState struct {
	StateID       string                     `json:"state_id"`
	SessionID     string                     `json:"session_id"`
	Phase         string                     `json:"phase"`
	Progress      float64                    `json:"progress"`
	ActiveAxes    []string                   `json:"active_axes"`
	AxisStates    map[string]interface{}     `json:"axis_states"`
	Conflicts     []AxisConflict             `json:"conflicts"`
	Optimizations []CoordinationOptimization `json:"optimizations"`
	Metrics       map[string]float64         `json:"metrics"`
	Balance       float64                    `json:"balance"`
	Synergy       float64                    `json:"synergy"`
	Timestamp     time.Time                  `json:"timestamp"`
	Metadata      map[string]interface{}     `json:"metadata"`
}

// CoordinationOptimization 协调优化
type CoordinationOptimization struct {
	OptimizationID      string                 `json:"optimization_id"`
	Type                string                 `json:"type"`
	TargetAxes          []string               `json:"target_axes"`
	CurrentState        map[string]interface{} `json:"current_state"`
	TargetState         map[string]interface{} `json:"target_state"`
	Strategy            string                 `json:"strategy"`
	Steps               []string               `json:"steps"`
	ExpectedBenefit     float64                `json:"expected_benefit"`
	Progress            float64                `json:"progress"`
	Status              string                 `json:"status"`
	StartedAt           time.Time              `json:"started_at"`
	CompletedAt         *time.Time             `json:"completed_at"`
	OptimizationGoals   []OptimizationGoal     `json:"optimization_goals"`
	OptimizationPlan    []string               `json:"optimization_plan"`
	ExpectedImprovement float64                `json:"expected_improvement"`
	OptimizationRisks   []OptimizationRisk     `json:"optimization_risks"`
	Timestamp           time.Time              `json:"timestamp"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// AxisBalance 轴平衡
type AxisBalance struct {
	AxisName     string                 `json:"axis_name"`
	AxisType     string                 `json:"axis_type"`
	CurrentValue float64                `json:"current_value"`
	TargetValue  float64                `json:"target_value"`
	Balance      float64                `json:"balance"`
	Trend        string                 `json:"trend"`
	Stability    float64                `json:"stability"`
	Influences   []string               `json:"influences"`
	Constraints  []string               `json:"constraints"`
	Adjustments  []string               `json:"adjustments"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// BalanceMetricsAnalysis 平衡指标分析
type BalanceMetricsAnalysis struct {
	AnalysisID      string                 `json:"analysis_id"`
	SessionID       string                 `json:"session_id"`
	OverallBalance  float64                `json:"overall_balance"`
	AxisBalances    []AxisBalance          `json:"axis_balances"`
	Correlations    map[string]float64     `json:"correlations"`
	Trends          []string               `json:"trends"`
	Anomalies       []string               `json:"anomalies"`
	Recommendations []string               `json:"recommendations"`
	AnalyzedAt      time.Time              `json:"analyzed_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// BalanceEvolutionPrediction 平衡演化预测
type BalanceEvolutionPrediction struct {
	PredictionID     string                 `json:"prediction_id"`
	SessionID        string                 `json:"session_id"`
	TimeHorizon      time.Duration          `json:"time_horizon"`
	CurrentBalance   []AxisBalance          `json:"current_balance"`
	PredictedBalance []AxisBalance          `json:"predicted_balance"`
	Scenarios        []string               `json:"scenarios"`
	Confidence       float64                `json:"confidence"`
	Assumptions      []string               `json:"assumptions"`
	RiskFactors      []string               `json:"risk_factors"`
	PredictedAt      time.Time              `json:"predicted_at"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// BalancePerformance 平衡性能
type BalancePerformance struct {
	PerformanceID string                 `json:"performance_id"`
	SessionID     string                 `json:"session_id"`
	Period        time.Duration          `json:"period"`
	OverallScore  float64                `json:"overall_score"`
	AxisScores    map[string]float64     `json:"axis_scores"`
	Improvements  []string               `json:"improvements"`
	Degradations  []string               `json:"degradations"`
	Efficiency    float64                `json:"efficiency"`
	Stability     float64                `json:"stability"`
	Adaptability  float64                `json:"adaptability"`
	MeasuredAt    time.Time              `json:"measured_at"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// WeightAdjustment 权重调整
type WeightAdjustment struct {
	AdjustmentID        string                 `json:"adjustment_id"`
	CurrentWeights      map[string]float64     `json:"current_weights"`
	AdjustedWeights     map[string]float64     `json:"adjusted_weights"`
	AdjustmentRatio     map[string]float64     `json:"adjustment_ratio"`
	AdjustmentReason    map[string]string      `json:"adjustment_reason"`
	Performance         *BalancePerformance    `json:"performance"`
	ExpectedImprovement float64                `json:"expected_improvement"`
	RiskLevel           string                 `json:"risk_level"`
	Timestamp           time.Time              `json:"timestamp"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// CoordinationContext 协调上下?
type CoordinationContext struct {
	ContextID         string                 `json:"context_id"`
	SessionID         string                 `json:"session_id"`
	SequenceResult    interface{}            `json:"sequence_result"`
	CompositionResult interface{}            `json:"composition_result"`
	ThoughtResult     interface{}            `json:"thought_result"`
	Environment       map[string]interface{} `json:"environment"`
	Constraints       []string               `json:"constraints"`
	Objectives        []string               `json:"objectives"`
	Resources         []string               `json:"resources"`
	Participants      []string               `json:"participants"`
	Timestamp         time.Time              `json:"timestamp"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// AxisRelationship 轴关系
type AxisRelationship struct {
	RelationshipID   string                 `json:"relationship_id"`
	SourceAxis       string                 `json:"source_axis"`
	TargetAxis       string                 `json:"target_axis"`
	RelationshipType string                 `json:"relationship_type"`
	Strength         float64                `json:"strength"`
	Direction        string                 `json:"direction"`
	Quality          float64                `json:"quality"`
	Stability        float64                `json:"stability"`
	Influence        float64                `json:"influence"`
	Correlation      float64                `json:"correlation"`
	Dependencies     []string               `json:"dependencies"`
	Constraints      []string               `json:"constraints"`
	Opportunities    []string               `json:"opportunities"`
	Risks            []string               `json:"risks"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CoordinationStatus 协调状态
type CoordinationStatus struct {
	Phase          string                 `json:"phase"`
	Progress       float64                `json:"progress"`
	CurrentStep    string                 `json:"current_step"`
	CompletedSteps []string               `json:"completed_steps"`
	PendingSteps   []string               `json:"pending_steps"`
	BlockedSteps   []string               `json:"blocked_steps"`
	Issues         []string               `json:"issues"`
	Warnings       []string               `json:"warnings"`
	Metadata       map[string]interface{} `json:"metadata"`
	LastUpdated    time.Time              `json:"last_updated"`
}

// CoordinationStep 协调步骤
type CoordinationStep struct {
	StepID           string                 `json:"step_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Type             string                 `json:"type"`
	Status           string                 `json:"status"`
	Priority         int                    `json:"priority"`
	Dependencies     []string               `json:"dependencies"`
	Prerequisites    []string               `json:"prerequisites"`
	ExpectedDuration time.Duration          `json:"expected_duration"`
	ActualDuration   time.Duration          `json:"actual_duration"`
	Resources        []string               `json:"resources"`
	Participants     []string               `json:"participants"`
	Outcomes         []string               `json:"outcomes"`
	Issues           []string               `json:"issues"`
	Metadata         map[string]interface{} `json:"metadata"`
	StartedAt        *time.Time             `json:"started_at"`
	CompletedAt      *time.Time             `json:"completed_at"`
}

// CoordinationIssue 协调问题
type CoordinationIssue struct {
	IssueID       string                 `json:"issue_id"`
	Type          string                 `json:"type"`
	Severity      string                 `json:"severity"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	AffectedSteps []string               `json:"affected_steps"`
	AffectedAxes  []string               `json:"affected_axes"`
	Causes        []string               `json:"causes"`
	Impact        string                 `json:"impact"`
	Solutions     []string               `json:"solutions"`
	Workarounds   []string               `json:"workarounds"`
	Status        string                 `json:"status"`
	Priority      int                    `json:"priority"`
	AssignedTo    string                 `json:"assigned_to"`
	Metadata      map[string]interface{} `json:"metadata"`
	ReportedAt    time.Time              `json:"reported_at"`
	ResolvedAt    *time.Time             `json:"resolved_at"`
}

// CoordinationRecord 协调记录
type CoordinationRecord struct {
	RecordID  string                 `json:"record_id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	Action    string                 `json:"action"`
	Target    string                 `json:"target"`
	Details   map[string]interface{} `json:"details"`
	Context   map[string]interface{} `json:"context"`
	Results   []string               `json:"results"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// QualityEvaluation 质量评估
type QualityEvaluation struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	SessionID       string                 `json:"session_id" gorm:"index"`
	EvaluationType  string                 `json:"evaluation_type"`
	OverallScore    float64                `json:"overall_score"`
	Dimensions      []QualityDimension     `json:"dimensions" gorm:"type:json"`
	Strengths       []string               `json:"strengths" gorm:"type:json"`
	Weaknesses      []string               `json:"weaknesses" gorm:"type:json"`
	Improvements    []string               `json:"improvements" gorm:"type:json"`
	Recommendations []string               `json:"recommendations" gorm:"type:json"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:json"`
	EvaluatedAt     time.Time              `json:"evaluated_at"`
}

// QualityDimension 质量维度
type QualityDimension struct {
	Name         string                 `json:"name"`
	Score        float64                `json:"score"`
	MaxScore     float64                `json:"max_score"`
	Weight       float64                `json:"weight"`
	Description  string                 `json:"description"`
	Criteria     []string               `json:"criteria"`
	Evidence     []string               `json:"evidence"`
	Improvements []string               `json:"improvements"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// BalanceAnalysis 平衡分析
type BalanceAnalysis struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	SessionID       string                 `json:"session_id" gorm:"index"`
	AnalysisType    string                 `json:"analysis_type"`
	OverallBalance  float64                `json:"overall_balance"`
	AxisBalances    map[string]float64     `json:"axis_balances" gorm:"type:json"`
	Imbalances      []AxisImbalance        `json:"imbalances" gorm:"type:json"`
	Recommendations []string               `json:"recommendations" gorm:"type:json"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:json"`
	AnalyzedAt      time.Time              `json:"analyzed_at"`
}

// AxisImbalance 轴不平衡
type AxisImbalance struct {
	AxisName      string                 `json:"axis_name"`
	ImbalanceType string                 `json:"imbalance_type"`
	Severity      string                 `json:"severity"`
	Score         float64                `json:"score"`
	TargetScore   float64                `json:"target_score"`
	Deviation     float64                `json:"deviation"`
	Causes        []string               `json:"causes"`
	Effects       []string               `json:"effects"`
	Solutions     []string               `json:"solutions"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// BalanceOptimization 平衡优化
type BalanceOptimization struct {
	ID                  string                  `json:"id" gorm:"primaryKey"`
	SessionID           string                  `json:"session_id" gorm:"index"`
	OptimizationType    string                  `json:"optimization_type"`
	CurrentBalance      float64                 `json:"current_balance"`
	TargetBalance       float64                 `json:"target_balance"`
	Adjustments         []BalanceAdjustment     `json:"adjustments" gorm:"type:json"`
	ExpectedImprovement float64                 `json:"expected_improvement"`
	Recommendations     []BalanceRecommendation `json:"recommendations" gorm:"type:json"`
	RiskLevel           string                  `json:"risk_level"`
	ImplementationTime  time.Duration           `json:"implementation_time"`
	Metadata            map[string]interface{}  `json:"metadata" gorm:"type:json"`
	OptimizedAt         time.Time               `json:"optimized_at"`
}

// BalanceAdjustment 平衡调整
type BalanceAdjustment struct {
	AdjustmentID   string                 `json:"adjustment_id"`
	TargetAxis     string                 `json:"target_axis"`
	AdjustmentType string                 `json:"adjustment_type"`
	CurrentValue   float64                `json:"current_value"`
	TargetValue    float64                `json:"target_value"`
	Change         float64                `json:"change"`
	Priority       int                    `json:"priority"`
	Difficulty     string                 `json:"difficulty"`
	ExpectedImpact float64                `json:"expected_impact"`
	RiskLevel      string                 `json:"risk_level"`
	Dependencies   []string               `json:"dependencies"`
	Steps          []string               `json:"steps"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// BalanceRecommendation 平衡推荐
type BalanceRecommendation struct {
	RecommendationID    string                 `json:"recommendation_id"`
	Type                string                 `json:"type"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description"`
	TargetAxes          []string               `json:"target_axes"`
	ExpectedBenefit     float64                `json:"expected_benefit"`
	ImplementationSteps []string               `json:"implementation_steps"`
	Priority            string                 `json:"priority"`
	Confidence          float64                `json:"confidence"`
	RiskLevel           string                 `json:"risk_level"`
	Timeline            time.Duration          `json:"timeline"`
	Resources           []string               `json:"resources"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// StabilityValidation 稳定性验证
type StabilityValidation struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	SessionID       string                 `json:"session_id" gorm:"index"`
	ValidationType  string                 `json:"validation_type"`
	StabilityScore  float64                `json:"stability_score"`
	IsStable        bool                   `json:"is_stable"`
	Factors         []StabilityFactor      `json:"factors" gorm:"type:json"`
	Risks           []StabilityRisk        `json:"risks" gorm:"type:json"`
	Recommendations []string               `json:"recommendations" gorm:"type:json"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:json"`
	ValidatedAt     time.Time              `json:"validated_at"`
}

// StabilityFactor 稳定性因子
type StabilityFactor struct {
	FactorName  string                 `json:"factor_name"`
	Type        string                 `json:"type"`
	Impact      string                 `json:"impact"`
	Score       float64                `json:"score"`
	Weight      float64                `json:"weight"`
	Trend       string                 `json:"trend"`
	Confidence  float64                `json:"confidence"`
	Description string                 `json:"description"`
	Indicators  []string               `json:"indicators"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StabilityRisk 稳定性风险
type StabilityRisk struct {
	RiskID       string                 `json:"risk_id"`
	Type         string                 `json:"type"`
	Severity     string                 `json:"severity"`
	Probability  float64                `json:"probability"`
	Impact       float64                `json:"impact"`
	RiskScore    float64                `json:"risk_score"`
	Description  string                 `json:"description"`
	Triggers     []string               `json:"triggers"`
	Consequences []string               `json:"consequences"`
	Mitigation   []string               `json:"mitigation"`
	Contingency  []string               `json:"contingency"`
	Metadata     map[string]interface{} `json:"metadata"`
}
