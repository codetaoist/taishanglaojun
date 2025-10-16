package models

import (
	"time"
)

// ConsciousnessRequest represents a general consciousness processing request
type ConsciousnessRequest struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "fusion", "evolution", "gene", "coordination"
	EntityID    string                 `json:"entity_id"`
	Priority    int                    `json:"priority"`
	Timeout     time.Duration          `json:"timeout"`
	Context     map[string]interface{} `json:"context"`
	Parameters  map[string]interface{} `json:"parameters"`
	Metadata    map[string]interface{} `json:"metadata"`
	RequestedAt time.Time              `json:"requested_at"`
}

// ConsciousnessResponse represents a general consciousness processing response
type ConsciousnessResponse struct {
	ID          string                 `json:"id"`
	RequestID   string                 `json:"request_id"`
	Type        string                 `json:"type"`
	Success     bool                   `json:"success"`
	Result      interface{}            `json:"result"`
	Error       string                 `json:"error,omitempty"`
	Metrics     map[string]interface{} `json:"metrics"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProcessedAt time.Time              `json:"processed_at"`
	Duration    time.Duration          `json:"duration"`
}

// FusionRequest represents a carbon-silicon fusion request
type FusionRequest struct {
	ID         string                 `json:"id"`
	EntityID   string                 `json:"entity_id"`
	CarbonData *CarbonInput           `json:"carbon_data"`
	SiliconData *SiliconInput         `json:"silicon_data"`
	Strategy   string                 `json:"strategy,omitempty"`
	Options    map[string]interface{} `json:"options"`
	Timeout    time.Duration          `json:"timeout"`
	Priority   int                    `json:"priority"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// EvolutionRequest represents an evolution tracking request
type EvolutionRequest struct {
	ID             string                 `json:"id"`
	EntityID       string                 `json:"entity_id"`
	Action         string                 `json:"action"` // "start", "stop", "update", "predict"
	TargetSequence SequenceLevel          `json:"target_sequence,omitempty"`
	Parameters     map[string]interface{} `json:"parameters"`
	Options        map[string]interface{} `json:"options"`
	Timeout        time.Duration          `json:"timeout"`
	Priority       int                    `json:"priority"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
}

// GeneRequest represents a quantum gene operation request
type GeneRequest struct {
	ID         string                 `json:"id"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"` // "create", "express", "mutate", "analyze"
	GeneData   *QuantumGene           `json:"gene_data,omitempty"`
	GeneID     string                 `json:"gene_id,omitempty"`
	Parameters map[string]interface{} `json:"parameters"`
	Options    map[string]interface{} `json:"options"`
	Timeout    time.Duration          `json:"timeout"`
	Priority   int                    `json:"priority"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// BalanceOptimizationRequest 平衡优化请求
type BalanceOptimizationRequest struct {
	Coordinate   Coordinate             `json:"coordinate" binding:"required"`   // 当前坐标
	Constraints  []Constraint           `json:"constraints"`                     // 约束条件
	Goals        []OptimizationGoal     `json:"goals"`                          // 优化目标
	Priority     string                 `json:"priority"`                       // 优先级：high, medium, low
	MaxDuration  int                    `json:"max_duration"`                   // 最大优化时间（秒）
	Metadata     map[string]interface{} `json:"metadata"`                       // 元数据
}

// SynergyCatalysisRequest 协同催化请求
type SynergyCatalysisRequest struct {
	Coordinate   Coordinate             `json:"coordinate" binding:"required"`   // 当前坐标
	TargetAxes   []string               `json:"target_axes"`                    // 目标轴：S, C, T
	Intensity    float64                `json:"intensity"`                      // 催化强度 0.0-1.0
	Duration     int                    `json:"duration"`                       // 催化持续时间（秒）
	Conditions   []string               `json:"conditions"`                     // 催化条件
	Metadata     map[string]interface{} `json:"metadata"`                       // 元数据
}

// EvolutionTrackingRequest 进化跟踪请求
type EvolutionTrackingRequest struct {
	EntityID     string                 `json:"entity_id" binding:"required"`   // 实体ID
	TrackingType string                 `json:"tracking_type"`                  // 跟踪类型：continuous, periodic, event-based
	Duration     int                    `json:"duration"`                       // 跟踪持续时间（秒）
	Metrics      []string               `json:"metrics"`                        // 跟踪指标
	Conditions   []string               `json:"conditions"`                     // 跟踪条件
	Metadata     map[string]interface{} `json:"metadata"`                       // 元数据
}

// GenePoolCreateRequest 基因池创建请求
type GenePoolCreateRequest struct {
	EntityID     string         `json:"entity_id" binding:"required"`
	Name         string         `json:"name" binding:"required"`
	Description  string         `json:"description"`
	InitialGenes []*QuantumGene `json:"initial_genes"`
}

// GeneExpressionRequest 基因表达请求
type GeneExpressionRequest struct {
	Intensity float64       `json:"intensity"`
	Duration  time.Duration `json:"duration"`
	Context   map[string]interface{} `json:"context"`
}

// GeneMutationRequest 基因突变请求
type GeneMutationRequest struct {
	MutationType string                 `json:"mutation_type"`
	Intensity    float64                `json:"intensity"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// EvolutionSimulationRequest 进化模拟请求
type EvolutionSimulationRequest struct {
	Generations       int     `json:"generations"`
	SelectionPressure float64 `json:"selection_pressure"`
	MutationRate      float64 `json:"mutation_rate"`
	PopulationSize    int     `json:"population_size"`
}