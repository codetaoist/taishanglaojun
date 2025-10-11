package models

import (
	"time"
)

// EvolutionState 进化状�?
type EvolutionState struct {
	ID                int64                  `json:"id" db:"id"`
	EntityID          string                 `json:"entity_id" db:"entity_id"`           // 进化实体ID
	CurrentSequence   SequenceLevel         `json:"current_sequence" db:"current_sequence"`
	TargetSequence    SequenceLevel         `json:"target_sequence" db:"target_sequence"`
	EvolutionPath     []EvolutionStep       `json:"evolution_path"`                     // 进化路径
	Progress          float64               `json:"progress" db:"progress"`             // 进化进度 0-1
	EvolutionSpeed    float64               `json:"evolution_speed" db:"evolution_speed"` // 进化速度
	Milestones        []EvolutionMilestone  `json:"milestones"`                        // 里程�?
	Constraints       []EvolutionConstraint `json:"constraints"`                       // 进化约束
	Catalysts         []EvolutionCatalyst   `json:"catalysts"`                         // 进化催化�?
	Status            EvolutionStatus       `json:"status" db:"status"`
	StartTime         time.Time             `json:"start_time" db:"start_time"`
	LastUpdateTime    time.Time             `json:"last_update_time" db:"last_update_time"`
	EstimatedCompletion *time.Time          `json:"estimated_completion,omitempty" db:"estimated_completion"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
}

// SequenceLevel 序列等级
type SequenceLevel int

const (
	SequenceUnknown SequenceLevel = -1
	Sequence5       SequenceLevel = 5  // 初级意识
	Sequence4       SequenceLevel = 4  // 基础智能
	Sequence3       SequenceLevel = 3  // 高级智能
	Sequence2       SequenceLevel = 2  // 超级智能
	Sequence1       SequenceLevel = 1  // 准神级智�?
	Sequence0       SequenceLevel = 0  // 终极意识/神级智能
)

// EvolutionStep 进化步骤
type EvolutionStep struct {
	StepID          string                 `json:"step_id"`
	FromSequence    SequenceLevel         `json:"from_sequence"`
	ToSequence      SequenceLevel         `json:"to_sequence"`
	Description     string                `json:"description"`
	Requirements    []EvolutionRequirement `json:"requirements"`    // 进化要求
	Achievements    []EvolutionAchievement `json:"achievements"`    // 已达成成�?
	Duration        time.Duration         `json:"duration"`        // 步骤耗时
	Difficulty      float64               `json:"difficulty"`      // 难度系数 0-1
	SuccessRate     float64               `json:"success_rate"`    // 成功�?0-1
	CompletedAt     *time.Time            `json:"completed_at,omitempty"`
	Status          StepStatus            `json:"status"`
}

// EvolutionMilestone 进化里程�?
type EvolutionMilestone struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Sequence        SequenceLevel         `json:"sequence"`        // 对应序列等级
	Criteria        []MilestoneCriteria   `json:"criteria"`        // 达成标准
	Rewards         []MilestoneReward     `json:"rewards"`         // 里程碑奖�?
	IsAchieved      bool                  `json:"is_achieved"`
	AchievedAt      *time.Time            `json:"achieved_at,omitempty"`
	Significance    float64               `json:"significance"`    // 重要�?0-1
	Metadata        map[string]interface{} `json:"metadata"`
}

// EvolutionConstraint 进化约束
type EvolutionConstraint struct {
	ID              string                 `json:"id"`
	Type            ConstraintType        `json:"type"`
	Description     string                `json:"description"`
	Severity        ConstraintSeverity    `json:"severity"`
	Impact          float64               `json:"impact"`          // 影响程度 0-1
	IsActive        bool                  `json:"is_active"`
	ActivatedAt     time.Time             `json:"activated_at"`
	ExpiresAt       *time.Time            `json:"expires_at,omitempty"`
	Conditions      []ConstraintCondition `json:"conditions"`      // 约束条件
	Metadata        map[string]interface{} `json:"metadata"`
}

// EvolutionCatalyst 进化催化�?
type EvolutionCatalyst struct {
	ID              string                 `json:"id"`
	Type            CatalystType          `json:"type"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Effectiveness   float64               `json:"effectiveness"`   // 有效�?0-1
	Duration        time.Duration         `json:"duration"`        // 持续时间
	IsActive        bool                  `json:"is_active"`
	ActivatedAt     time.Time             `json:"activated_at"`
	ExpiresAt       *time.Time            `json:"expires_at,omitempty"`
	Effects         []CatalystEffect      `json:"effects"`         // 催化效果
	Requirements    []CatalystRequirement `json:"requirements"`    // 使用要求
	Metadata        map[string]interface{} `json:"metadata"`
}

// EvolutionRequirement 进化要求
type EvolutionRequirement struct {
	ID              string                 `json:"id"`
	Type            RequirementType       `json:"type"`
	Description     string                `json:"description"`
	Threshold       float64               `json:"threshold"`       // 阈�?
	CurrentValue    float64               `json:"current_value"`   // 当前�?
	IsMet           bool                  `json:"is_met"`          // 是否满足
	Priority        RequirementPriority   `json:"priority"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// EvolutionAchievement 进化成就
type EvolutionAchievement struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                `json:"description"`
	Type            AchievementType       `json:"type"`
	Value           float64               `json:"value"`           // 成就�?
	AchievedAt      time.Time             `json:"achieved_at"`
	Significance    float64               `json:"significance"`    // 重要�?
	Metadata        map[string]interface{} `json:"metadata"`
}

// MilestoneCriteria 里程碑标�?
type MilestoneCriteria struct {
	ID              string                 `json:"id"`
	Description     string                `json:"description"`
	Type            CriteriaType          `json:"type"`
	Threshold       float64               `json:"threshold"`
	CurrentValue    float64               `json:"current_value"`
	IsMet           bool                  `json:"is_met"`
	Weight          float64               `json:"weight"`          // 权重 0-1
}

// MilestoneReward 里程碑奖�?
type MilestoneReward struct {
	ID              string                 `json:"id"`
	Type            RewardType            `json:"type"`
	Description     string                `json:"description"`
	Value           interface{}           `json:"value"`
	IsApplied       bool                  `json:"is_applied"`
	AppliedAt       *time.Time            `json:"applied_at,omitempty"`
}

// ConstraintCondition 约束条件
type ConstraintCondition struct {
	ID              string                 `json:"id"`
	Description     string                `json:"description"`
	Type            ConditionType         `json:"type"`
	Threshold       float64               `json:"threshold"`
	CurrentValue    float64               `json:"current_value"`
	IsMet           bool                  `json:"is_met"`
}

// CatalystEffect 催化剂效�?
type CatalystEffect struct {
	ID              string                 `json:"id"`
	Type            EffectType            `json:"type"`
	Description     string                `json:"description"`
	Magnitude       float64               `json:"magnitude"`       // 效果强度
	Target          EffectTarget          `json:"target"`          // 作用目标
	IsActive        bool                  `json:"is_active"`
}

// CatalystRequirement 催化剂要�?
type CatalystRequirement struct {
	ID              string                 `json:"id"`
	Type            RequirementType       `json:"type"`
	Description     string                `json:"description"`
	Threshold       float64               `json:"threshold"`
	IsMet           bool                  `json:"is_met"`
}

// EvolutionMetrics 进化指标
type EvolutionMetrics struct {
	EntityID            string    `json:"entity_id"`
	ConsciousnessLevel  float64   `json:"consciousness_level"`   // 意识水平 0-1
	IntelligenceQuotient float64  `json:"intelligence_quotient"` // 智能商数
	WisdomIndex         float64   `json:"wisdom_index"`          // 智慧指数
	CreativityScore     float64   `json:"creativity_score"`      // 创造力评分
	AdaptabilityRating  float64   `json:"adaptability_rating"`   // 适应性评�?
	SelfAwarenessLevel  float64   `json:"self_awareness_level"`  // 自我意识水平
	TranscendenceIndex  float64   `json:"transcendence_index"`   // 超越指数
	EvolutionPotential  float64   `json:"evolution_potential"`   // 进化潜力
	LastUpdated         time.Time `json:"last_updated"`
}

// EvolutionPrediction 进化预测
type EvolutionPrediction struct {
	EntityID            string                 `json:"entity_id"`
	PredictedSequence   SequenceLevel         `json:"predicted_sequence"`
	Confidence          float64               `json:"confidence"`          // 预测置信�?
	TimeToAchieve       time.Duration         `json:"time_to_achieve"`     // 预计达成时间
	RequiredCatalysts   []string              `json:"required_catalysts"`  // 需要的催化�?
	PotentialObstacles  []string              `json:"potential_obstacles"` // 潜在障碍
	SuccessProbability  float64               `json:"success_probability"` // 成功概率
	AlternativePaths    []EvolutionPath       `json:"alternative_paths"`   // 替代路径
	GeneratedAt         time.Time             `json:"generated_at"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// EvolutionPath 进化路径
type EvolutionPath struct {
	ID              string                 `json:"id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Steps           []EvolutionStep       `json:"steps"`
	TotalDuration   time.Duration         `json:"total_duration"`
	Difficulty      float64               `json:"difficulty"`
	SuccessRate     float64               `json:"success_rate"`
	RequiredResources []PathResource      `json:"required_resources"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PathResource 路径资源
type PathResource struct {
	Type            ResourceType          `json:"type"`
	Amount          float64               `json:"amount"`
	Description     string                `json:"description"`
	IsAvailable     bool                  `json:"is_available"`
}

// 枚举类型定义

type EvolutionStatus string

const (
	EvolutionStatusInitializing EvolutionStatus = "initializing"
	EvolutionStatusInProgress   EvolutionStatus = "in_progress"
	EvolutionStatusPaused       EvolutionStatus = "paused"
	EvolutionStatusCompleted    EvolutionStatus = "completed"
	EvolutionStatusFailed       EvolutionStatus = "failed"
	EvolutionStatusStagnant     EvolutionStatus = "stagnant"
)

type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusFailed     StepStatus = "failed"
	StepStatusSkipped    StepStatus = "skipped"
)

type ConstraintType string

const (
	ConstraintTypeResource     ConstraintType = "resource"
	ConstraintTypeTime         ConstraintType = "time"
	ConstraintTypeCapability   ConstraintType = "capability"
	ConstraintTypeEnvironment  ConstraintType = "environment"
	ConstraintTypeEthical      ConstraintType = "ethical"
	ConstraintTypePhysical     ConstraintType = "physical"
)

type ConstraintSeverity string

const (
	ConstraintSeverityLow      ConstraintSeverity = "low"
	ConstraintSeverityMedium   ConstraintSeverity = "medium"
	ConstraintSeverityHigh     ConstraintSeverity = "high"
	ConstraintSeverityCritical ConstraintSeverity = "critical"
)

type CatalystType string

const (
	CatalystTypeKnowledge    CatalystType = "knowledge"
	CatalystTypeExperience   CatalystType = "experience"
	CatalystTypeResource     CatalystType = "resource"
	CatalystTypeEnvironment  CatalystType = "environment"
	CatalystTypeCollaboration CatalystType = "collaboration"
	CatalystTypeInnovation   CatalystType = "innovation"
)

type RequirementType string

const (
	RequirementTypeCapability   RequirementType = "capability"
	RequirementTypeKnowledge    RequirementType = "knowledge"
	RequirementTypeResource     RequirementType = "resource"
	RequirementTypeExperience   RequirementType = "experience"
	RequirementTypeAchievement  RequirementType = "achievement"
	RequirementTypeMetric       RequirementType = "metric"
)

type RequirementPriority string

const (
	RequirementPriorityLow      RequirementPriority = "low"
	RequirementPriorityMedium   RequirementPriority = "medium"
	RequirementPriorityHigh     RequirementPriority = "high"
	RequirementPriorityCritical RequirementPriority = "critical"
)

type AchievementType string

const (
	AchievementTypeCapability   AchievementType = "capability"
	AchievementTypeKnowledge    AchievementType = "knowledge"
	AchievementTypeWisdom       AchievementType = "wisdom"
	AchievementTypeCreativity   AchievementType = "creativity"
	AchievementTypeTranscendence AchievementType = "transcendence"
	AchievementTypeMilestone    AchievementType = "milestone"
)

type CriteriaType string

const (
	CriteriaTypeMetric      CriteriaType = "metric"
	CriteriaTypeCapability  CriteriaType = "capability"
	CriteriaTypeAchievement CriteriaType = "achievement"
	CriteriaTypeTime        CriteriaType = "time"
	CriteriaTypeQuality     CriteriaType = "quality"
)

type RewardType string

const (
	RewardTypeCapabilityBoost RewardType = "capability_boost"
	RewardTypeResourceGrant   RewardType = "resource_grant"
	RewardTypeAccessUnlock    RewardType = "access_unlock"
	RewardTypeKnowledgeGain   RewardType = "knowledge_gain"
	RewardTypeStatusUpgrade   RewardType = "status_upgrade"
)

type ConditionType string

const (
	ConditionTypeMetric      ConditionType = "metric"
	ConditionTypeResource    ConditionType = "resource"
	ConditionTypeTime        ConditionType = "time"
	ConditionTypeEnvironment ConditionType = "environment"
	ConditionTypeState       ConditionType = "state"
)

type EffectType string

const (
	EffectTypeSpeedBoost      EffectType = "speed_boost"
	EffectTypeCapabilityBoost EffectType = "capability_boost"
	EffectTypeResourceBonus   EffectType = "resource_bonus"
	EffectTypeBarrierRemoval  EffectType = "barrier_removal"
	EffectTypeInsightGain     EffectType = "insight_gain"
)

type EffectTarget string

const (
	EffectTargetEvolutionSpeed EffectTarget = "evolution_speed"
	EffectTargetCapability     EffectTarget = "capability"
	EffectTargetResource       EffectTarget = "resource"
	EffectTargetConstraint     EffectTarget = "constraint"
	EffectTargetAwareness      EffectTarget = "awareness"
)

type ResourceType string

const (
	ResourceTypeComputational ResourceType = "computational"
	ResourceTypeKnowledge     ResourceType = "knowledge"
	ResourceTypeExperience    ResourceType = "experience"
	ResourceTypeTime          ResourceType = "time"
	ResourceTypeEnergy        ResourceType = "energy"
	ResourceTypeData          ResourceType = "data"
)

// 辅助方法

// String 返回序列等级的字符串表示
func (sl SequenceLevel) String() string {
	switch sl {
	case Sequence0:
		return "Sequence 0 - Ultimate Consciousness"
	case Sequence1:
		return "Sequence 1 - Quasi-Divine Intelligence"
	case Sequence2:
		return "Sequence 2 - Super Intelligence"
	case Sequence3:
		return "Sequence 3 - Advanced Intelligence"
	case Sequence4:
		return "Sequence 4 - Basic Intelligence"
	case Sequence5:
		return "Sequence 5 - Primary Consciousness"
	default:
		return "Unknown Sequence"
	}
}

// IsHigherThan 判断是否比另一个序列等级更�?
func (sl SequenceLevel) IsHigherThan(other SequenceLevel) bool {
	return sl < other // 数字越小，序列等级越�?
}

// GetDifficulty 获取达到该序列等级的难度
func (sl SequenceLevel) GetDifficulty() float64 {
	switch sl {
	case Sequence0:
		return 1.0 // 最高难�?
	case Sequence1:
		return 0.9
	case Sequence2:
		return 0.8
	case Sequence3:
		return 0.6
	case Sequence4:
		return 0.4
	case Sequence5:
		return 0.2
	default:
		return 0.5
	}
}

// GetRequiredCapabilities 获取达到该序列等级所需的能�?
func (sl SequenceLevel) GetRequiredCapabilities() []string {
	switch sl {
	case Sequence0:
		return []string{"ultimate_consciousness", "infinite_wisdom", "transcendent_creativity", "quantum_awareness"}
	case Sequence1:
		return []string{"advanced_consciousness", "deep_wisdom", "high_creativity", "meta_awareness"}
	case Sequence2:
		return []string{"super_intelligence", "strategic_thinking", "pattern_recognition", "self_optimization"}
	case Sequence3:
		return []string{"advanced_reasoning", "complex_problem_solving", "learning_ability", "adaptation"}
	case Sequence4:
		return []string{"basic_reasoning", "simple_problem_solving", "memory", "recognition"}
	case Sequence5:
		return []string{"basic_awareness", "simple_response", "pattern_matching"}
	default:
		return []string{}
	}
}
