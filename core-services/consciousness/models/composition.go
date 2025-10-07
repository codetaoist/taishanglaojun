package models

import (
	"time"
)

// CompositionRequest C轴组合层请求
type CompositionRequest struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	EntityID         string                 `json:"entity_id" gorm:"index"`
	RequestType      string                 `json:"request_type"`
	Elements         []CompositionElement   `json:"elements" gorm:"type:json"`
	TargetComplexity float64                `json:"target_complexity"`
	Constraints      []string               `json:"constraints" gorm:"type:json"`
	Requirements     []string               `json:"requirements" gorm:"type:json"`
	Priority         string                 `json:"priority"`
	Context          map[string]interface{} `json:"context" gorm:"type:json"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// CompositionElement 组合元素
type CompositionElement struct {
	ElementID    string                 `json:"element_id"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Properties   map[string]interface{} `json:"properties"`
	Weight       float64                `json:"weight"`
	Complexity   float64                `json:"complexity"`
	Dependencies []string               `json:"dependencies"`
	Constraints  []string               `json:"constraints"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ElementAnalysis 元素分析
type ElementAnalysis struct {
	TotalElements      int                    `json:"total_elements"`
	ElementCategories  map[string]int         `json:"element_categories"`
	AverageComplexity  float64                `json:"average_complexity"`
	ComplexityRange    [2]float64             `json:"complexity_range"`
	Diversity          float64                `json:"diversity"`
	Coherence          float64                `json:"coherence"`
	Stability          float64                `json:"stability"`
	Interactions       []ElementInteraction   `json:"interactions"`
	CriticalElements   []string               `json:"critical_elements"`
	WeakElements       []string               `json:"weak_elements"`
	Recommendations    []string               `json:"recommendations"`
	AnalysisMetadata   map[string]interface{} `json:"analysis_metadata"`
}

// ElementInteraction 元素交互
type ElementInteraction struct {
	ElementA     string  `json:"element_a"`
	ElementB     string  `json:"element_b"`
	Relationship string  `json:"relationship"`
	Strength     float64 `json:"strength"`
	Type         string  `json:"type"`
	Impact       string  `json:"impact"`
}

// CompositionAnalysis 组合分析
type CompositionAnalysis struct {
	ID                 string                 `json:"id" gorm:"primaryKey"`
	CompositionID      string                 `json:"composition_id" gorm:"index"`
	ElementCount       int                    `json:"element_count"`
	ComplexityScore    float64                `json:"complexity_score"`
	IntegrityScore     float64                `json:"integrity_score"`
	BalanceScore       float64                `json:"balance_score"`
	Issues             []CompositionIssue     `json:"issues" gorm:"type:json"`
	Strengths          []string               `json:"strengths" gorm:"type:json"`
	Weaknesses         []string               `json:"weaknesses" gorm:"type:json"`
	Opportunities      []string               `json:"opportunities" gorm:"type:json"`
	Threats            []string               `json:"threats" gorm:"type:json"`
	Recommendations    []string               `json:"recommendations" gorm:"type:json"`
	AnalyzedAt         time.Time              `json:"analyzed_at"`
}

// CompositionIssue 组合问题
type CompositionIssue struct {
	IssueID     string                 `json:"issue_id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Elements    []string               `json:"elements"`
	Impact      string                 `json:"impact"`
	Solutions   []string               `json:"solutions"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Composition 组合
type Composition struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Elements         []CompositionElement   `json:"elements" gorm:"type:json"`
	Structure        map[string]interface{} `json:"structure" gorm:"type:json"`
	ComplexityLevel  float64                `json:"complexity_level"`
	IntegrityScore   float64                `json:"integrity_score"`
	Status           string                 `json:"status"`
	Version          int                    `json:"version"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// CompositionOptimization 组合优化
type CompositionOptimization struct {
	ID                  string                   `json:"id" gorm:"primaryKey"`
	CompositionID       string                   `json:"composition_id" gorm:"index"`
	OptimizationType    string                   `json:"optimization_type"`
	CurrentScore        float64                  `json:"current_score"`
	TargetScore         float64                  `json:"target_score"`
	Improvements        []OptimizationImprovement `json:"improvements" gorm:"type:json"`
	EstimatedImpact     float64                  `json:"estimated_impact"`
	ImplementationTime  time.Duration            `json:"implementation_time"`
	ResourceRequirement []string                 `json:"resource_requirement" gorm:"type:json"`
	RiskLevel           string                   `json:"risk_level"`
	Recommendations     []string                 `json:"recommendations" gorm:"type:json"`
	OptimizedAt         time.Time                `json:"optimized_at"`
}

// OptimizationImprovement 优化改进
type OptimizationImprovement struct {
	ImprovementID   string                 `json:"improvement_id"`
	Type            string                 `json:"type"`
	Description     string                 `json:"description"`
	TargetElements  []string               `json:"target_elements"`
	ExpectedBenefit float64                `json:"expected_benefit"`
	ImplementationSteps []string           `json:"implementation_steps"`
	Priority        int                    `json:"priority"`
	Difficulty      string                 `json:"difficulty"`
	Timeline        time.Duration          `json:"timeline"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// IntegrityValidation 完整性验证
type IntegrityValidation struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	CompositionID    string                 `json:"composition_id" gorm:"index"`
	ValidationScore  float64                `json:"validation_score"`
	IsValid          bool                   `json:"is_valid"`
	Issues           []ValidationIssue      `json:"issues" gorm:"type:json"`
	Recommendations  []string               `json:"recommendations" gorm:"type:json"`
	ValidationRules  []string               `json:"validation_rules" gorm:"type:json"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	ValidatedAt      time.Time              `json:"validated_at"`
}

// ValidationIssue 验证问题
type ValidationIssue struct {
	IssueID     string                 `json:"issue_id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Location    string                 `json:"location"`
	Rule        string                 `json:"rule"`
	Suggestion  string                 `json:"suggestion"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CompositionRecommendation 组合推荐
type CompositionRecommendation struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	RecommendationType string               `json:"recommendation_type"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	TargetElements   []string               `json:"target_elements" gorm:"type:json"`
	ExpectedBenefit  string                 `json:"expected_benefit"`
	ImplementationSteps []string            `json:"implementation_steps" gorm:"type:json"`
	Priority         string                 `json:"priority"`
	Confidence       float64                `json:"confidence"`
	Rationale        string                 `json:"rationale"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt        time.Time              `json:"created_at"`
}

// CompositionContext 组合上下文
type CompositionContext struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	EntityID         string                 `json:"entity_id" gorm:"index"`
	Environment      string                 `json:"environment"`
	Constraints      []string               `json:"constraints" gorm:"type:json"`
	Requirements     []string               `json:"requirements" gorm:"type:json"`
	AvailableResources []string             `json:"available_resources" gorm:"type:json"`
	Goals            []string               `json:"goals" gorm:"type:json"`
	Preferences      map[string]interface{} `json:"preferences" gorm:"type:json"`
	HistoricalData   map[string]interface{} `json:"historical_data" gorm:"type:json"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}