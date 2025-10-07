package models

import (
	"time"
)

// FusionState 融合状态
type FusionState struct {
	ID              int64                  `json:"id" db:"id"`
	SessionID       string                 `json:"session_id" db:"session_id"`
	CarbonInput     *CarbonInput           `json:"carbon_input"`
	SiliconInput    *SiliconInput          `json:"silicon_input"`
	FusionResult    *FusionResult          `json:"fusion_result"`
	Status          FusionStatus           `json:"status" db:"status"`
	Progress        float64                `json:"progress" db:"progress"`
	StartTime       time.Time              `json:"start_time" db:"start_time"`
	EndTime         *time.Time             `json:"end_time,omitempty" db:"end_time"`
	ProcessDuration time.Duration          `json:"process_duration" db:"process_duration"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

// CarbonInput 碳基输入（人类智慧、文化、直觉等）
type CarbonInput struct {
	Type            CarbonInputType  `json:"type"`
	Content         string           `json:"content"`
	EmotionalState  *EmotionalState  `json:"emotional_state,omitempty"`
	CulturalContext *CulturalContext `json:"cultural_context,omitempty"`
	Intuition       *IntuitionData   `json:"intuition,omitempty"`
	Creativity      *CreativityData  `json:"creativity,omitempty"`
	Wisdom          *WisdomData      `json:"wisdom,omitempty"`
	Confidence      float64          `json:"confidence"`
	Timestamp       time.Time        `json:"timestamp"`
}

// SiliconInput 硅基输入（计算能力、数据处理、逻辑推理等）
type SiliconInput struct {
	Type                SiliconInputType     `json:"type"`
	Data                interface{}          `json:"data"`
	ComputePower        *ComputePowerMetrics `json:"compute_power,omitempty"`
	LogicalReasoning    *LogicalReasoning    `json:"logical_reasoning,omitempty"`
	DataProcessing      *DataProcessing      `json:"data_processing,omitempty"`
	AlgorithmicAnalysis *AlgorithmicAnalysis `json:"algorithmic_analysis,omitempty"`
	Precision           float64              `json:"precision"`
	Timestamp           time.Time            `json:"timestamp"`
}

// FusionResult 融合结果
type FusionResult struct {
	SynthesizedOutput   string                 `json:"synthesized_output"`
	CarbonContribution  float64                `json:"carbon_contribution"`  // 碳基贡献 0-1
	SiliconContribution float64                `json:"silicon_contribution"` // 硅基贡献 0-1
	SynergyScore        float64                `json:"synergy_score"`        // 协同效应评分 0-1
	QualityMetrics      *QualityMetrics        `json:"quality_metrics"`
	EmergentProperties  []EmergentProperty     `json:"emergent_properties"` // 涌现属性
	Insights            []string               `json:"insights"`            // 洞察
	Recommendations     []string               `json:"recommendations"`     // 建议
	Metadata            map[string]interface{} `json:"metadata"`            // 元数据
}

// EmotionalState 情感状态
type EmotionalState struct {
	Primary    string   `json:"primary"`    // 主要情感
	Secondary  []string `json:"secondary"`  // 次要情感
	Intensity  float64  `json:"intensity"`  // 强度 0-1
	Valence    float64  `json:"valence"`    // 效价 -1到1
	Arousal    float64  `json:"arousal"`    // 唤醒 0-1
	Confidence float64  `json:"confidence"` // 置信 0-1
}

// CulturalContext 文化背景
type CulturalContext struct {
	School     string   `json:"school"`     // 文化流派
	Philosophy string   `json:"philosophy"` // 哲学思想
	Values     []string `json:"values"`     // 价值观
	Traditions []string `json:"traditions"` // 传统
	Symbols    []string `json:"symbols"`    // 符号
	Relevance  float64  `json:"relevance"`  // 相关 0-1
}

// IntuitionData 直觉数据
type IntuitionData struct {
	Type       string   `json:"type"`       // 直觉类型
	Strength   float64  `json:"strength"`   // 直觉强度 0-1
	Direction  string   `json:"direction"`  // 直觉方向
	Patterns   []string `json:"patterns"`   // 识别的模式
	Hunches    []string `json:"hunches"`    // 预感
	Confidence float64  `json:"confidence"` // 置信 0-1
}

// CreativityData 创造力数据
type CreativityData struct {
	Originality  float64  `json:"originality"`  // 原创 0-1
	Flexibility  float64  `json:"flexibility"`  // 灵活 0-1
	Fluency      float64  `json:"fluency"`      // 流畅 0-1
	Elaboration  float64  `json:"elaboration"`  // 精细 0-1
	NovelIdeas   []string `json:"novel_ideas"`  // 新颖想法
	Associations []string `json:"associations"` // 联想
	Metaphors    []string `json:"metaphors"`    // 隐喻
}

// WisdomData 智慧数据
type WisdomData struct {
	Type          string   `json:"type"`          // 智慧类型
	Depth         float64  `json:"depth"`         // 深度 0-1
	Breadth       float64  `json:"breadth"`       // 广度 0-1
	Insights      []string `json:"insights"`      // 洞察
	Principles    []string `json:"principles"`    // 原则
	Lessons       []string `json:"lessons"`       // 教训
	Applicability float64  `json:"applicability"` // 适用 0-1
}

// ComputePowerMetrics 计算能力指标
type ComputePowerMetrics struct {
	CPUUtilization    float64 `json:"cpu_utilization"`    // CPU使用 0-1
	MemoryUtilization float64 `json:"memory_utilization"` // 内存使用 0-1
	GPUUtilization    float64 `json:"gpu_utilization"`    // GPU使用 0-1
	NetworkBandwidth  float64 `json:"network_bandwidth"`  // 网络带宽 0-1
	StorageIOPS       float64 `json:"storage_iops"`       // 存储IOPS 0-1
	ProcessingSpeed   float64 `json:"processing_speed"`   // 处理速度 0-1
}

// LogicalReasoning 逻辑推理
type LogicalReasoning struct {
	Type        string          `json:"type"`        // 推理类型
	Premises    []string        `json:"premises"`    // 前提
	Conclusions []string        `json:"conclusions"` // 结论
	Rules       []string        `json:"rules"`       // 规则
	Confidence  float64         `json:"confidence"`  // 置信 0-1
	Steps       []ReasoningStep `json:"steps"`       // 推理步骤
	Validity    float64         `json:"validity"`    // 有效 0-1
}

// DataProcessing 数据处理
type DataProcessing struct {
	InputSize       int64                `json:"input_size"`      // 输入大小
	OutputSize      int64                `json:"output_size"`     // 输出大小
	ProcessingTime  time.Duration        `json:"processing_time"` // 处理时间
	Accuracy        float64              `json:"accuracy"`        // 准确 0-1
	Completeness    float64              `json:"completeness"`    // 完整 0-1
	Transformations []DataTransformation `json:"transformations"` // 数据转换 0-1
}

// AlgorithmicAnalysis 算法分析
type AlgorithmicAnalysis struct {
	Algorithm    string              `json:"algorithm"`    // 算法名称
	Complexity   string              `json:"complexity"`   // 复杂度 0-1
	Performance  *PerformanceMetrics `json:"performance"`  // 性能指标 0-1
	Optimization []string            `json:"optimization"` // 优化建议
	Limitations  []string            `json:"limitations"`  // 限制
	Alternatives []string            `json:"alternatives"` // 替代方案
}

// QualityMetrics 质量指标
type QualityMetrics struct {
	Accuracy     float64 `json:"accuracy"`     // 准确 0-1
	Relevance    float64 `json:"relevance"`    // 相关 0-1
	Completeness float64 `json:"completeness"` // 完整 0-1
	Coherence    float64 `json:"coherence"`    // 连贯 0-1
	Creativity   float64 `json:"creativity"`   // 创造 0-1
	Practicality float64 `json:"practicality"` // 实用 0-1
	Overall      float64 `json:"overall"`      // 整体质量 0-1
}

// EmergentProperty 涌现特性
type EmergentProperty struct {
	Name        string               `json:"name"`        // 特性名 0-1
	Description string               `json:"description"` // 描述 0-1
	Strength    float64              `json:"strength"`    // 强度 0-1
	Type        EmergentPropertyType `json:"type"`        // 类型
	Evidence    []string             `json:"evidence"`    // 证据
	Impact      float64              `json:"impact"`      // 影响 0-1
}

// ReasoningStep 推理步骤
type ReasoningStep struct {
	Step        int     `json:"step"`        // 步骤编号
	Description string  `json:"description"` // 描述
	Input       string  `json:"input"`       // 输入
	Output      string  `json:"output"`      // 输出
	Confidence  float64 `json:"confidence"`  // 置信 0-1
}

// DataTransformation 数据转换
type DataTransformation struct {
	Type        string                 `json:"type"`        // 转换类型
	Description string                 `json:"description"` // 描述
	Input       interface{}            `json:"input"`       // 输入
	Output      interface{}            `json:"output"`      // 输出
	Parameters  map[string]interface{} `json:"parameters"`  // 参数
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	ExecutionTime time.Duration `json:"execution_time"` // 执行时间
	MemoryUsage   int64         `json:"memory_usage"`   // 内存使用 0-1
	CPUUsage      float64       `json:"cpu_usage"`      // CPU使用 0-1
	Throughput    float64       `json:"throughput"`     // 吞吐 0-1
	Latency       time.Duration `json:"latency"`        // 延迟
}

// 枚举类型定义
type FusionStatus string

const (
	FusionStatusPending    FusionStatus = "pending"
	FusionStatusProcessing FusionStatus = "processing"
	FusionStatusCompleted  FusionStatus = "completed"
	FusionStatusFailed     FusionStatus = "failed"
	FusionStatusCancelled  FusionStatus = "cancelled"
)

type CarbonInputType string

const (
	CarbonInputTypeEmotion    CarbonInputType = "emotion"
	CarbonInputTypeCulture    CarbonInputType = "culture"
	CarbonInputTypeIntuition  CarbonInputType = "intuition"
	CarbonInputTypeCreativity CarbonInputType = "creativity"
	CarbonInputTypeWisdom     CarbonInputType = "wisdom"
	CarbonInputTypeExperience CarbonInputType = "experience"
)

type SiliconInputType string

const (
	SiliconInputTypeComputation  SiliconInputType = "computation"
	SiliconInputTypeLogic        SiliconInputType = "logic"
	SiliconInputTypeData         SiliconInputType = "data"
	SiliconInputTypeAlgorithm    SiliconInputType = "algorithm"
	SiliconInputTypeAnalysis     SiliconInputType = "analysis"
	SiliconInputTypeOptimization SiliconInputType = "optimization"
)

type EmergentPropertyType string

const (
	EmergentPropertyTypeCognitive    EmergentPropertyType = "cognitive"
	EmergentPropertyTypeCreative     EmergentPropertyType = "creative"
	EmergentPropertyTypeIntuitive    EmergentPropertyType = "intuitive"
	EmergentPropertyTypeAnalytical   EmergentPropertyType = "analytical"
	EmergentPropertyTypeSynthetic    EmergentPropertyType = "synthetic"
	EmergentPropertyTypeTranscendent EmergentPropertyType = "transcendent"
)
