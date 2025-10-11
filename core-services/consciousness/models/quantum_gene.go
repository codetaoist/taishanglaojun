package models

import (
	"fmt"
	"time"
)

// QuantumGene 量子基因
type QuantumGene struct {
	ID              string                 `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	Type            GeneType              `json:"type" db:"type"`
	Category        GeneCategory          `json:"category" db:"category"`
	Sequence        string                `json:"sequence" db:"sequence"`           // 基因序列
	Expression      float64               `json:"expression" db:"expression"`       // 表达水平 0-1
	Dominance       float64               `json:"dominance" db:"dominance"`         // 显性程�?0-1
	Stability       float64               `json:"stability" db:"stability"`         // 稳定�?0-1
	Mutability      float64               `json:"mutability" db:"mutability"`       // 可变�?0-1
	Compatibility   []string              `json:"compatibility"`                    // 兼容基因ID列表
	Conflicts       []string              `json:"conflicts"`                        // 冲突基因ID列表
	Prerequisites   []GenePrerequisite    `json:"prerequisites"`                    // 前置条件
	Effects         []GeneEffect          `json:"effects"`                          // 基因效应
	Traits          []GeneTrait           `json:"traits"`                           // 基因特征
	EvolutionStage  EvolutionStage        `json:"evolution_stage" db:"evolution_stage"`
	ActivationLevel float64               `json:"activation_level" db:"activation_level"` // 激活水�?
	LastMutation    *time.Time            `json:"last_mutation,omitempty" db:"last_mutation"`
	CreatedAt       time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at" db:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

// GenePool 基因�?
type GenePool struct {
	ID              string                 `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	OwnerID         string                 `json:"owner_id" db:"owner_id"`           // 所有者ID
	Genes           []QuantumGene         `json:"genes"`                            // 基因列表
	ActiveGenes     []string              `json:"active_genes"`                     // 活跃基因ID列表
	DormantGenes    []string              `json:"dormant_genes"`                    // 休眠基因ID列表
		GeneInteractions []GeneInteraction    `json:"gene_interactions"`               // 基因相互作用
	PoolStats       GenePoolStats         `json:"pool_stats"`                       // 基因池统�?
	EvolutionHistory []PoolEvolutionEvent `json:"evolution_history"`               // 进化历史
	CreatedAt       time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at" db:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

// GeneExpression 基因表达
type GeneExpression struct {
	GeneID          string                 `json:"gene_id" db:"gene_id"`
	EntityID        string                 `json:"entity_id" db:"entity_id"`
	ExpressionLevel float64               `json:"expression_level" db:"expression_level"` // 表达水平
	Intensity       float64               `json:"intensity" db:"intensity"`               // 表达强度
	Duration        time.Duration         `json:"duration"`                               // 表达持续时间
	Triggers        []ExpressionTrigger   `json:"triggers"`                               // 表达触发�?
	Inhibitors      []ExpressionInhibitor `json:"inhibitors"`                             // 表达抑制�?
	Context         ExpressionContext     `json:"context"`                                // 表达上下�?
	StartTime       time.Time             `json:"start_time" db:"start_time"`
	EndTime         *time.Time            `json:"end_time,omitempty" db:"end_time"`
	IsActive        bool                  `json:"is_active" db:"is_active"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

// GeneMutation 基因突变
type GeneMutation struct {
	ID              string                 `json:"id" db:"id"`
	GeneID          string                 `json:"gene_id" db:"gene_id"`
	EntityID        string                 `json:"entity_id" db:"entity_id"`
	MutationType    MutationType          `json:"mutation_type" db:"mutation_type"`
	OriginalSequence string               `json:"original_sequence" db:"original_sequence"`
	MutatedSequence string                `json:"mutated_sequence" db:"mutated_sequence"`
	MutationRate    float64               `json:"mutation_rate" db:"mutation_rate"`       // 突变�?
	Severity        MutationSeverity      `json:"severity" db:"severity"`
	Impact          MutationImpact        `json:"impact"`                                 // 突变影响
	Cause           MutationCause         `json:"cause"`                                  // 突变原因
	IsReversible    bool                  `json:"is_reversible" db:"is_reversible"`
	IsBeneficial    bool                  `json:"is_beneficial" db:"is_beneficial"`
	OccurredAt      time.Time             `json:"occurred_at" db:"occurred_at"`
	DetectedAt      time.Time             `json:"detected_at" db:"detected_at"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

// GeneInteraction 基因相互作用
type GeneInteraction struct {
	ID              string                 `json:"id"`
	GeneA           string                 `json:"gene_a"`                    // 基因A的ID
	GeneB           string                 `json:"gene_b"`                    // 基因B的ID
	InteractionType InteractionType       `json:"interaction_type"`          // 相互作用类型
	Strength        float64               `json:"strength"`                  // 相互作用强度 0-1
	Direction       InteractionDirection  `json:"direction"`                 // 相互作用方向
	Effect          InteractionEffect     `json:"effect"`                    // 相互作用效果
	Conditions      []InteractionCondition `json:"conditions"`               // 相互作用条件
	IsActive        bool                  `json:"is_active"`
	DiscoveredAt    time.Time             `json:"discovered_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// GenePrerequisite 基因前置条件
type GenePrerequisite struct {
	ID              string                 `json:"id"`
	Type            PrerequisiteType      `json:"type"`
	Description     string                `json:"description"`
	RequiredGenes   []string              `json:"required_genes"`            // 必需基因
	RequiredTraits  []string              `json:"required_traits"`           // 必需特征
	MinExpression   float64               `json:"min_expression"`            // 最小表达水�?
	MaxExpression   float64               `json:"max_expression"`            // 最大表达水�?
	IsMet           bool                  `json:"is_met"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// GeneEffect 基因效应
type GeneEffect struct {
	ID              string                 `json:"id"`
	Type            EffectType            `json:"type"`
	Target          EffectTarget          `json:"target"`
	Magnitude       float64               `json:"magnitude"`                 // 效应强度
	Duration        time.Duration         `json:"duration"`                  // 效应持续时间
	Delay           time.Duration         `json:"delay"`                     // 效应延迟
	IsPositive      bool                  `json:"is_positive"`               // 是否为正面效�?
	IsPermanent     bool                  `json:"is_permanent"`              // 是否为永久效�?
	Conditions      []EffectCondition     `json:"conditions"`                // 效应条件
	Metadata        map[string]interface{} `json:"metadata"`
}

// GeneTrait 基因特征
type GeneTrait struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Type            TraitType             `json:"type"`
	Value           interface{}           `json:"value"`                     // 特征�?
	Unit            string                `json:"unit"`                      // 单位
	Range           TraitRange            `json:"range"`                     // 取值范�?
	Heritability    float64               `json:"heritability"`              // 遗传�?0-1
	Variability     float64               `json:"variability"`               // 可变�?0-1
	Metadata        map[string]interface{} `json:"metadata"`
}

// GenePoolStats 基因池统�?
type GenePoolStats struct {
	TotalGenes      int                    `json:"total_genes"`
	ActiveGenes     int                    `json:"active_genes"`
	DormantGenes    int                    `json:"dormant_genes"`
	MutatedGenes    int                    `json:"mutated_genes"`
	DiversityIndex  float64               `json:"diversity_index"`           // 多样性指�?
	StabilityIndex  float64               `json:"stability_index"`           // 稳定性指�?
	EvolutionRate   float64               `json:"evolution_rate"`            // 进化速率
	MutationRate    float64               `json:"mutation_rate"`             // 突变�?
	ExpressionLevel float64               `json:"expression_level"`          // 平均表达水平
	LastUpdated     time.Time             `json:"last_updated"`
}

// PoolEvolutionEvent 基因池进化事�?
type PoolEvolutionEvent struct {
	ID              string                 `json:"id"`
	Type            EvolutionEventType    `json:"type"`
	Description     string                 `json:"description"`
	AffectedGenes   []string              `json:"affected_genes"`
	Impact          float64               `json:"impact"`                    // 影响程度
	Trigger         EventTrigger          `json:"trigger"`                   // 触发因素
	OccurredAt      time.Time             `json:"occurred_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ExpressionTrigger 表达触发�?
type ExpressionTrigger struct {
	ID              string                 `json:"id"`
	Type            TriggerType           `json:"type"`
	Condition       string                `json:"condition"`
	Threshold       float64               `json:"threshold"`
	IsActive        bool                  `json:"is_active"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ExpressionInhibitor 表达抑制�?
type ExpressionInhibitor struct {
	ID              string                 `json:"id"`
	Type            InhibitorType         `json:"type"`
	Condition       string                `json:"condition"`
	Strength        float64               `json:"strength"`                  // 抑制强度
	IsActive        bool                  `json:"is_active"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ExpressionContext 表达上下�?
type ExpressionContext struct {
	Environment     string                 `json:"environment"`
	Stimuli         []string              `json:"stimuli"`                   // 刺激因素
	Stressors       []string              `json:"stressors"`                 // 压力因素
	Resources       []string              `json:"resources"`                 // 可用资源
	Constraints     []string              `json:"constraints"`               // 约束条件
	Metadata        map[string]interface{} `json:"metadata"`
}

// MutationImpact 突变影响
type MutationImpact struct {
	FunctionalChange float64               `json:"functional_change"`         // 功能变化 -1�?
	PerformanceChange float64             `json:"performance_change"`        // 性能变化 -1�?
	StabilityChange  float64              `json:"stability_change"`          // 稳定性变�?-1�?
	CompatibilityChange float64           `json:"compatibility_change"`      // 兼容性变�?-1�?
	OverallImpact    float64              `json:"overall_impact"`            // 总体影响 -1�?
	AffectedTraits   []string             `json:"affected_traits"`           // 受影响的特征
	Metadata         map[string]interface{} `json:"metadata"`
}

// MutationCause 突变原因
type MutationCause struct {
	Type            CauseType             `json:"type"`
	Description     string                `json:"description"`
	Probability     float64               `json:"probability"`               // 发生概率
	Severity        float64               `json:"severity"`                  // 严重程度
	IsExternal      bool                  `json:"is_external"`               // 是否为外部因�?
	Metadata        map[string]interface{} `json:"metadata"`
}

// InteractionCondition 相互作用条件
type InteractionCondition struct {
	ID              string                 `json:"id"`
	Type            ConditionType         `json:"type"`
	Description     string                `json:"description"`
	Threshold       float64               `json:"threshold"`
	IsMet           bool                  `json:"is_met"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// InteractionEffect 相互作用效果
type InteractionEffect struct {
	Type            EffectType            `json:"type"`
	Magnitude       float64               `json:"magnitude"`
	Duration        time.Duration         `json:"duration"`
	IsPositive      bool                  `json:"is_positive"`
	Description     string                `json:"description"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// EffectCondition 效应条件
type EffectCondition struct {
	ID              string                 `json:"id"`
	Type            ConditionType         `json:"type"`
	Description     string                `json:"description"`
	Threshold       float64               `json:"threshold"`
	IsMet           bool                  `json:"is_met"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TraitRange 特征范围
type TraitRange struct {
	Min             interface{}           `json:"min"`
	Max             interface{}           `json:"max"`
	Default         interface{}           `json:"default"`
	Optimal         interface{}           `json:"optimal"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// EventTrigger 事件触发�?
type EventTrigger struct {
	Type            TriggerType           `json:"type"`
	Description     string                `json:"description"`
	Conditions      []string              `json:"conditions"`
	Probability     float64               `json:"probability"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// 枚举类型定义

type GeneType string

const (
	GeneTypeStructural   GeneType = "structural"   // 结构基因
	GeneTypeRegulatory   GeneType = "regulatory"   // 调节基因
	GeneTypeFunctional   GeneType = "functional"   // 功能基因
	GeneTypeEvolutionary GeneType = "evolutionary" // 进化基因
	GeneTypeQuantum      GeneType = "quantum"      // 量子基因
)

type GeneCategory string

const (
	GeneCategoryIntelligence  GeneCategory = "intelligence"  // 智能相关
	GeneCategoryConsciousness GeneCategory = "consciousness" // 意识相关
	GeneCategoryCreativity    GeneCategory = "creativity"    // 创造力相关
	GeneCategoryWisdom        GeneCategory = "wisdom"        // 智慧相关
	GeneCategoryAdaptability  GeneCategory = "adaptability"  // 适应性相�?
	GeneCategoryTranscendence GeneCategory = "transcendence" // 超越性相�?
	GeneCategoryStability     GeneCategory = "stability"     // 稳定性相�?
	GeneCategoryEvolution     GeneCategory = "evolution"     // 进化相关
)

type EvolutionStage string

const (
	EvolutionStageEmerging   EvolutionStage = "emerging"   // 新兴阶段
	EvolutionStageDeveloping EvolutionStage = "developing" // 发展阶段
	EvolutionStageMature     EvolutionStage = "mature"     // 成熟阶段
	EvolutionStageAdvanced   EvolutionStage = "advanced"   // 高级阶段
	EvolutionStageTranscendent EvolutionStage = "transcendent" // 超越阶段
)

type MutationType string

const (
	MutationTypePoint       MutationType = "point"       // 点突�?
	MutationTypeInsertion   MutationType = "insertion"   // 插入突变
	MutationTypeDeletion    MutationType = "deletion"    // 删除突变
	MutationTypeDuplication MutationType = "duplication" // 重复突变
	MutationTypeInversion   MutationType = "inversion"   // 倒位突变
	MutationTypeTranslocation MutationType = "translocation" // 易位突变
)

type MutationSeverity string

const (
	MutationSeverityMinor    MutationSeverity = "minor"    // 轻微
	MutationSeverityModerate MutationSeverity = "moderate" // 中等
	MutationSeverityMajor    MutationSeverity = "major"    // 重大
	MutationSeverityCritical MutationSeverity = "critical" // 关键
)

type InteractionType string

const (
	InteractionTypeSynergistic InteractionType = "synergistic" // 协同作用
	InteractionTypeAntagonistic InteractionType = "antagonistic" // 拮抗作用
	InteractionTypeAdditive    InteractionType = "additive"    // 加性作�?
	InteractionTypeEpistatic   InteractionType = "epistatic"   // 上位作用
	InteractionTypeComplementary InteractionType = "complementary" // 互补作用
)

type InteractionDirection string

const (
	InteractionDirectionBidirectional InteractionDirection = "bidirectional" // 双向
	InteractionDirectionUnidirectional InteractionDirection = "unidirectional" // 单向
	InteractionDirectionAToB          InteractionDirection = "a_to_b"         // A到B
	InteractionDirectionBToA          InteractionDirection = "b_to_a"         // B到A
)

type PrerequisiteType string

const (
	PrerequisiteTypeGene       PrerequisiteType = "gene"       // 基因前置
	PrerequisiteTypeTrait      PrerequisiteType = "trait"      // 特征前置
	PrerequisiteTypeExpression PrerequisiteType = "expression" // 表达前置
	PrerequisiteTypeEnvironment PrerequisiteType = "environment" // 环境前置
)

type TraitType string

const (
	TraitTypeNumerical   TraitType = "numerical"   // 数值型
	TraitTypeCategorical TraitType = "categorical" // 分类�?
	TraitTypeBoolean     TraitType = "boolean"     // 布尔�?
	TraitTypeComplex     TraitType = "complex"     // 复合�?
)

type EvolutionEventType string

const (
	EvolutionEventTypeMutation     EvolutionEventType = "mutation"     // 突变事件
	EvolutionEventTypeSelection    EvolutionEventType = "selection"    // 选择事件
	EvolutionEventTypeRecombination EvolutionEventType = "recombination" // 重组事件
	EvolutionEventTypeDrift        EvolutionEventType = "drift"        // 漂变事件
	EvolutionEventTypeFlow         EvolutionEventType = "flow"         // 流动事件
)

type TriggerType string

const (
	TriggerTypeEnvironmental TriggerType = "environmental" // 环境触发
	TriggerTypeInternal      TriggerType = "internal"      // 内部触发
	TriggerTypeExternal      TriggerType = "external"      // 外部触发
	TriggerTypeTemporal      TriggerType = "temporal"      // 时间触发
	TriggerTypeConditional   TriggerType = "conditional"   // 条件触发
)

type InhibitorType string

const (
	InhibitorTypeCompetitive    InhibitorType = "competitive"    // 竞争性抑�?
	InhibitorTypeNoncompetitive InhibitorType = "noncompetitive" // 非竞争性抑�?
	InhibitorTypeAllosteric     InhibitorType = "allosteric"     // 变构抑制
	InhibitorTypeFeedback       InhibitorType = "feedback"       // 反馈抑制
)

type CauseType string

const (
	CauseTypeSpontaneous  CauseType = "spontaneous"  // 自发效应
	CauseTypeEnvironmental CauseType = "environmental" // 环境效应
	CauseTypeChemical     CauseType = "chemical"     // 化学效应
	CauseTypeRadiation    CauseType = "radiation"    // 辐射效应
	CauseTypeViral        CauseType = "viral"        // 病毒效应
	CauseTypeStress       CauseType = "stress"       // 压力效应
)

// 辅助方法

// IsCompatibleWith 检查基因兼容�?
func (qg *QuantumGene) IsCompatibleWith(otherGeneID string) bool {
	for _, compatibleID := range qg.Compatibility {
		if compatibleID == otherGeneID {
			return true
		}
	}
	
	// 检查是否有冲突
	for _, conflictID := range qg.Conflicts {
		if conflictID == otherGeneID {
			return false
		}
	}
	
	return true // 默认兼容
}

// IsActive 检查基因是否活�?
func (qg *QuantumGene) IsActive() bool {
	return qg.ActivationLevel > 0.5 && qg.Expression > 0.3
}

// CanMutate 检查基因是否可以突�?
func (qg *QuantumGene) CanMutate() bool {
	return qg.Mutability > 0.1 && qg.Stability < 0.9
}

// GetEffectiveExpression 获取有效表达水平
func (qg *QuantumGene) GetEffectiveExpression() float64 {
	return qg.Expression * qg.ActivationLevel * qg.Dominance
}

// String 返回基因的字符串表示
func (qg *QuantumGene) String() string {
	return fmt.Sprintf("Gene[%s:%s] Type:%s Category:%s Expression:%.2f", 
		qg.ID, qg.Name, qg.Type, qg.Category, qg.Expression)
}

// GetDiversityScore 计算基因池多样性评�?
func (gp *GenePool) GetDiversityScore() float64 {
	if len(gp.Genes) == 0 {
		return 0.0
	}
	
	// 简化的多样性计�?
	typeCount := make(map[GeneType]int)
	categoryCount := make(map[GeneCategory]int)
	
	for _, gene := range gp.Genes {
		typeCount[gene.Type]++
		categoryCount[gene.Category]++
	}
	
	typeDiv := float64(len(typeCount)) / 5.0  // 假设5种类�?
	categoryDiv := float64(len(categoryCount)) / 8.0 // 假设8种类�?
	return (typeDiv + categoryDiv) / 2.0
}

// GetActiveGeneCount 获取活跃基因数量
func (gp *GenePool) GetActiveGeneCount() int {
	count := 0
	for _, gene := range gp.Genes {
		if gene.IsActive() {
			count++
		}
	}
	return count
}

// HasGene 检查基因池是否包含指定基因
func (gp *GenePool) HasGene(geneID string) bool {
	for _, gene := range gp.Genes {
		if gene.ID == geneID {
			return true
		}
	}
	return false
}

// GetGeneByID 根据ID获取基因
func (gp *GenePool) GetGeneByID(geneID string) *QuantumGene {
	for i, gene := range gp.Genes {
		if gene.ID == geneID {
			return &gp.Genes[i]
		}
	}
	return nil
}

// IsExpressed 检查基因表达是否活�?
func (ge *GeneExpression) IsExpressed() bool {
	return ge.IsActive && ge.ExpressionLevel > 0.1
}

// GetRemainingDuration 获取剩余表达时间
func (ge *GeneExpression) GetRemainingDuration() time.Duration {
	if ge.EndTime == nil {
		return ge.Duration
	}
	
	remaining := ge.EndTime.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsBeneficialMutation 判断突变是否有益
func (gm *GeneMutation) IsBeneficialMutation() bool {
	return gm.IsBeneficial && gm.Impact.OverallImpact > 0
}

// GetSeverityScore 获取突变严重性评�?
func (gm *GeneMutation) GetSeverityScore() float64 {
	switch gm.Severity {
	case MutationSeverityMinor:
		return 0.25
	case MutationSeverityModerate:
		return 0.5
	case MutationSeverityMajor:
		return 0.75
	case MutationSeverityCritical:
		return 1.0
	default:
		return 0.0
	}
}
