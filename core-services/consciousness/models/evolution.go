package models

import (
	"strconv"
	"strings"
	"time"
)

// EvolutionState 
type EvolutionState struct {
	ID                  int64                  `json:"id" db:"id"`
	EntityID            string                 `json:"entity_id" db:"entity_id"` // ID
	CurrentSequence     SequenceLevel          `json:"current_sequence" db:"current_sequence"`
	TargetSequence      SequenceLevel          `json:"target_sequence" db:"target_sequence"`
	EvolutionPath       []EvolutionStep        `json:"evolution_path"`                       // 
	Progress            float64                `json:"progress" db:"progress"`               //  0-1
	EvolutionSpeed      float64                `json:"evolution_speed" db:"evolution_speed"` // 
	Milestones          []EvolutionMilestone   `json:"milestones"`                           // 
	Constraints         []EvolutionConstraint  `json:"constraints"`                          // 
	Catalysts           []EvolutionCatalyst    `json:"catalysts"`                            // 
	Status              EvolutionStatus        `json:"status" db:"status"`
	StartTime           time.Time              `json:"start_time" db:"start_time"`
	LastUpdateTime      time.Time              `json:"last_update_time" db:"last_update_time"`
	EstimatedCompletion *time.Time             `json:"estimated_completion,omitempty" db:"estimated_completion"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`
}

// SequenceLevel 
type SequenceLevel int

const (
	SequenceUnknown SequenceLevel = -1
	Sequence5       SequenceLevel = 5 // 
	Sequence4       SequenceLevel = 4 // 
	Sequence3       SequenceLevel = 3 // 
	Sequence2       SequenceLevel = 2 // 
	Sequence1       SequenceLevel = 1 // 
	Sequence0       SequenceLevel = 0 // /
)

// EvolutionStep 
type EvolutionStep struct {
	StepID       string                 `json:"step_id"`
	FromSequence SequenceLevel          `json:"from_sequence"`
	ToSequence   SequenceLevel          `json:"to_sequence"`
	Description  string                 `json:"description"`
	Requirements []EvolutionRequirement `json:"requirements"` // 
	Achievements []EvolutionAchievement `json:"achievements"` // 
	Duration     time.Duration          `json:"duration"`     // 
	Difficulty   float64                `json:"difficulty"`   //  0-1
	SuccessRate  float64                `json:"success_rate"` //  0-1
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Status       StepStatus             `json:"status"`
}

// EvolutionMilestone 
type EvolutionMilestone struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Sequence     SequenceLevel          `json:"sequence"` // 
	Criteria     []MilestoneCriteria    `json:"criteria"` // 
	Rewards      []MilestoneReward      `json:"rewards"`  // 
	IsAchieved   bool                   `json:"is_achieved"`
	AchievedAt   *time.Time             `json:"achieved_at,omitempty"`
	Significance float64                `json:"significance"` //  0-1
	Metadata     map[string]interface{} `json:"metadata"`
}

// EvolutionConstraint 
type EvolutionConstraint struct {
	ID          string                 `json:"id"`
	Type        ConstraintType         `json:"type"`
	Description string                 `json:"description"`
	Severity    ConstraintSeverity     `json:"severity"`
	Impact      float64                `json:"impact"` //  0-1
	IsActive    bool                   `json:"is_active"`
	ActivatedAt time.Time              `json:"activated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Conditions  []ConstraintCondition  `json:"conditions"` // 
	Metadata    map[string]interface{} `json:"metadata"`
}

// EvolutionCatalyst 
type EvolutionCatalyst struct {
	ID            string                 `json:"id"`
	Type          CatalystType           `json:"type"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Effectiveness float64                `json:"effectiveness"` //  0-1
	Duration      time.Duration          `json:"duration"`      // 
	IsActive      bool                   `json:"is_active"`
	ActivatedAt   time.Time              `json:"activated_at"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
	Effects       []CatalystEffect       `json:"effects"`      // 
	Requirements  []CatalystRequirement  `json:"requirements"` // 
	Metadata      map[string]interface{} `json:"metadata"`
}

// EvolutionRequirement 
type EvolutionRequirement struct {
	ID           string                 `json:"id"`
	Type         RequirementType        `json:"type"`
	Description  string                 `json:"description"`
	Threshold    float64                `json:"threshold"`     //  0-1
	CurrentValue float64                `json:"current_value"` //  0-1
	IsMet        bool                   `json:"is_met"`        // 
	Priority     RequirementPriority    `json:"priority"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// EvolutionAchievement 
type EvolutionAchievement struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         AchievementType        `json:"type"`
	Value        float64                `json:"value"` //  0-1
	AchievedAt   time.Time              `json:"achieved_at"`
	Significance float64                `json:"significance"` //  0-1
	Metadata     map[string]interface{} `json:"metadata"`
}

// MilestoneCriteria 
type MilestoneCriteria struct {
	ID           string       `json:"id"`
	Description  string       `json:"description"`
	Type         CriteriaType `json:"type"`
	Threshold    float64      `json:"threshold"`     //  0-1
	CurrentValue float64      `json:"current_value"` //  0-1
	IsMet        bool         `json:"is_met"`
	Weight       float64      `json:"weight"` //  0-1
}

// MilestoneReward 
type MilestoneReward struct {
	ID          string      `json:"id"`
	Type        RewardType  `json:"type"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
	IsApplied   bool        `json:"is_applied"`
	AppliedAt   *time.Time  `json:"applied_at,omitempty"`
}

// ConstraintCondition 
type ConstraintCondition struct {
	ID           string        `json:"id"`
	Description  string        `json:"description"`
	Type         ConditionType `json:"type"`
	Threshold    float64       `json:"threshold"`
	CurrentValue float64       `json:"current_value"`
	IsMet        bool          `json:"is_met"`
}

// CatalystEffect 
type CatalystEffect struct {
	ID          string       `json:"id"`
	Type        EffectType   `json:"type"`
	Description string       `json:"description"`
	Magnitude   float64      `json:"magnitude"` //  0-1
	Target      EffectTarget `json:"target"`    // 
	IsActive    bool         `json:"is_active"`
}

// CatalystRequirement 
type CatalystRequirement struct {
	ID          string          `json:"id"`
	Type        RequirementType `json:"type"`
	Description string          `json:"description"`
	Threshold   float64         `json:"threshold"` //  0-1
	IsMet       bool            `json:"is_met"`
}

// EvolutionMetrics 
type EvolutionMetrics struct {
	EntityID             string    `json:"entity_id"`
	ConsciousnessLevel   float64   `json:"consciousness_level"`   //  0-1
	IntelligenceQuotient float64   `json:"intelligence_quotient"` // 
	WisdomIndex          float64   `json:"wisdom_index"`          // 
	CreativityScore      float64   `json:"creativity_score"`      //  0-1
	AdaptabilityRating   float64   `json:"adaptability_rating"`   //  0-1
	SelfAwarenessLevel   float64   `json:"self_awareness_level"`  //  0-1
	TranscendenceIndex   float64   `json:"transcendence_index"`   //  0-1
	EvolutionPotential   float64   `json:"evolution_potential"`   //  0-1
	LastUpdated          time.Time `json:"last_updated"`
}

// EvolutionPrediction 
type EvolutionPrediction struct {
	EntityID           string                 `json:"entity_id"`
	PredictedSequence  SequenceLevel          `json:"predicted_sequence"`
	Confidence         float64                `json:"confidence"`          //  0-1
	TimeToAchieve      time.Duration          `json:"time_to_achieve"`     // 
	RequiredCatalysts  []string               `json:"required_catalysts"`  // 
	PotentialObstacles []string               `json:"potential_obstacles"` // 
	SuccessProbability float64                `json:"success_probability"` //  0-1
	AlternativePaths   []EvolutionPath        `json:"alternative_paths"`   // 
	GeneratedAt        time.Time              `json:"generated_at"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// EvolutionPath 
type EvolutionPath struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Steps             []EvolutionStep        `json:"steps"`
	TotalDuration     time.Duration          `json:"total_duration"`
	Difficulty        float64                `json:"difficulty"`
	SuccessRate       float64                `json:"success_rate"`
	RequiredResources []PathResource         `json:"required_resources"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// PathResource 
type PathResource struct {
	Type        ResourceType `json:"type"`
	Amount      float64      `json:"amount"`
	Description string       `json:"description"`
	IsAvailable bool         `json:"is_available"`
}

// 

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
	ConstraintTypeResource    ConstraintType = "resource"
	ConstraintTypeTime        ConstraintType = "time"
	ConstraintTypeCapability  ConstraintType = "capability"
	ConstraintTypeEnvironment ConstraintType = "environment"
	ConstraintTypeEthical     ConstraintType = "ethical"
	ConstraintTypePhysical    ConstraintType = "physical"
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
	CatalystTypeKnowledge     CatalystType = "knowledge"
	CatalystTypeExperience    CatalystType = "experience"
	CatalystTypeResource      CatalystType = "resource"
	CatalystTypeEnvironment   CatalystType = "environment"
	CatalystTypeCollaboration CatalystType = "collaboration"
	CatalystTypeInnovation    CatalystType = "innovation"
)

type RequirementType string

const (
	RequirementTypeCapability  RequirementType = "capability"
	RequirementTypeKnowledge   RequirementType = "knowledge"
	RequirementTypeResource    RequirementType = "resource"
	RequirementTypeExperience  RequirementType = "experience"
	RequirementTypeAchievement RequirementType = "achievement"
	RequirementTypeMetric      RequirementType = "metric"
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
	AchievementTypeCapability    AchievementType = "capability"
	AchievementTypeKnowledge     AchievementType = "knowledge"
	AchievementTypeWisdom        AchievementType = "wisdom"
	AchievementTypeCreativity    AchievementType = "creativity"
	AchievementTypeTranscendence AchievementType = "transcendence"
	AchievementTypeMilestone     AchievementType = "milestone"
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

// 

// String 
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

// IsHigherThan 
func (sl SequenceLevel) IsHigherThan(other SequenceLevel) bool {
	return sl < other // 
}

// GetDifficulty 
func (sl SequenceLevel) GetDifficulty() float64 {
	switch sl {
	case Sequence0:
		return 1.0 // 
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

// GetRequiredCapabilities 
func (sl SequenceLevel) GetRequiredCapabilities() []string {
	switch sl {
	case Sequence5:
		return []string{"基础感知", "简单反应", "环境适应"}
	case Sequence4:
		return []string{"模式识别", "学习能力", "记忆存储", "基础推理"}
	case Sequence3:
		return []string{"抽象思维", "创造性思考", "情感理解", "社交互动"}
	case Sequence2:
		return []string{"自我意识", "元认知", "价值判断", "道德推理"}
	case Sequence1:
		return []string{"超越性思维", "宇宙意识", "智慧整合", "创造力"}
	case Sequence0:
		return []string{"绝对觉知", "无限创造", "完美智慧", "宇宙统一"}
	default:
		return []string{}
	}
}

// ParseSequenceLevel 解析字符串为 SequenceLevel
func ParseSequenceLevel(s string) SequenceLevel {
	s = strings.ToLower(strings.TrimSpace(s))
	
	switch s {
	case "sequence_0", "sequence0", "0":
		return Sequence0
	case "sequence_1", "sequence1", "1":
		return Sequence1
	case "sequence_2", "sequence2", "2":
		return Sequence2
	case "sequence_3", "sequence3", "3":
		return Sequence3
	case "sequence_4", "sequence4", "4":
		return Sequence4
	case "sequence_5", "sequence5", "5":
		return Sequence5
	default:
		// 尝试解析数字
		if num, err := strconv.Atoi(s); err == nil {
			switch num {
			case 0:
				return Sequence0
			case 1:
				return Sequence1
			case 2:
				return Sequence2
			case 3:
				return Sequence3
			case 4:
				return Sequence4
			case 5:
				return Sequence5
			}
		}
		return SequenceUnknown
	}
}

