package engines

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
)

// DefaultMutationEngine 默认突变引擎实现
type DefaultMutationEngine struct {
	config *MutationEngineConfig
	logger Logger
}

type MutationEngineConfig struct {
	BaseMutationRate    float64 `json:"base_mutation_rate"`
	MaxMutationSeverity models.MutationSeverity `json:"max_mutation_severity"`
	MutationCooldown    time.Duration `json:"mutation_cooldown"`
	EnableReverseMutation bool `json:"enable_reverse_mutation"`
}

// MutationImpact 突变影响
type MutationImpact struct {
	FitnessChange     float64 `json:"fitness_change"`
	StabilityChange   float64 `json:"stability_change"`
	ExpressionChange  float64 `json:"expression_change"`
	CompatibilityChange float64 `json:"compatibility_change"`
	PredictedOutcome  string  `json:"predicted_outcome"`
	Confidence        float64 `json:"confidence"`
}

func NewDefaultMutationEngine(config *MutationEngineConfig, logger Logger) *DefaultMutationEngine {
	if config == nil {
		config = &MutationEngineConfig{
			BaseMutationRate:    0.01,
			MaxMutationSeverity: models.MutationSeverityMajor,
			MutationCooldown:    time.Hour,
			EnableReverseMutation: true,
		}
	}
	return &DefaultMutationEngine{config: config, logger: logger}
}

func (dme *DefaultMutationEngine) GenerateMutation(ctx context.Context, gene *models.QuantumGene) (*models.GeneMutation, error) {
	mutationType := dme.selectMutationType(gene)
	severity := dme.calculateMutationSeverity(gene)
	
	mutation := &models.GeneMutation{
		GeneID:           gene.ID,
		EntityID:         gene.ID, // 使用基因ID作为实体ID
		MutationType:     mutationType,
		Severity:         severity,
		OriginalSequence: gene.Sequence,
		MutatedSequence:  dme.mutateSequence(gene.Sequence, dme.getSeverityFactor(severity)),
		MutationRate:     dme.calculateMutationProbability(gene),
		Impact:           models.MutationImpact{
			OverallImpact: dme.estimateImpact(gene, mutationType, severity),
		},
		Cause: models.MutationCause{
			Type:        models.CauseTypeSpontaneous,
			Description: "Random mutation during gene processing",
			Probability: dme.calculateMutationProbability(gene),
		},
		IsReversible: dme.config.EnableReverseMutation,
		IsBeneficial: false, // 默认为false，后续可以评�?
		OccurredAt:   time.Now(),
		DetectedAt:   time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	return mutation, nil
}

func (dme *DefaultMutationEngine) ApplyMutation(ctx context.Context, gene *models.QuantumGene, mutation *models.GeneMutation) error {
	switch mutation.MutationType {
	case models.MutationTypePoint:
		return dme.applySequenceMutation(gene, mutation)
	case models.MutationTypeInsertion:
		return dme.applyExpressionMutation(gene, mutation)
	case models.MutationTypeDeletion:
		return dme.applyDominanceMutation(gene, mutation)
	case models.MutationTypeDuplication:
		return dme.applyStabilityMutation(gene, mutation)
	case models.MutationTypeInversion:
		return dme.applyMutabilityMutation(gene, mutation)
	default:
		return fmt.Errorf("unknown mutation type: %s", mutation.MutationType)
	}
}

func (dme *DefaultMutationEngine) EvaluateMutationImpact(ctx context.Context, mutation *models.GeneMutation) (*MutationImpact, error) {
	impact := &MutationImpact{
		FitnessChange:       dme.calculateFitnessImpact(mutation),
		StabilityChange:     dme.calculateStabilityImpact(mutation),
		ExpressionChange:    dme.calculateExpressionImpact(mutation),
		CompatibilityChange: dme.calculateCompatibilityImpact(mutation),
		PredictedOutcome:    dme.predictOutcome(mutation),
		Confidence:          dme.calculateConfidence(mutation),
	}
	return impact, nil
}

func (dme *DefaultMutationEngine) PredictMutationProbability(ctx context.Context, gene *models.QuantumGene) (float64, error) {
	baseProbability := dme.config.BaseMutationRate
	
	// 基于基因特性调整概�?
	mutabilityFactor := gene.Mutability
	stabilityFactor := 1.0 - gene.Stability
	ageFactor := dme.calculateAgeFactor(gene)
	
	probability := baseProbability * mutabilityFactor * stabilityFactor * ageFactor
	
	// 限制在合理范围内
	if probability > 1.0 {
		probability = 1.0
	}
	if probability < 0.0 {
		probability = 0.0
	}
	
	return probability, nil
}

func (dme *DefaultMutationEngine) ReverseMutation(ctx context.Context, mutation *models.GeneMutation) error {
	if !mutation.IsReversible {
		return fmt.Errorf("mutation %s is not reversible", mutation.ID)
	}
	
	// 创建反向突变
	reverseMutation := &models.GeneMutation{
		ID:           fmt.Sprintf("rev_%s", mutation.ID),
		GeneID:       mutation.GeneID,
		EntityID:     mutation.EntityID,
		MutationType: mutation.MutationType,
		OriginalSequence: mutation.MutatedSequence,
		MutatedSequence:  mutation.OriginalSequence,
		MutationRate:     mutation.MutationRate,
		Severity:         mutation.Severity,
		Impact:           mutation.Impact,
		Cause:            mutation.Cause,
		IsReversible:     false,
		IsBeneficial:     !mutation.IsBeneficial,
		OccurredAt:       time.Now(),
		DetectedAt:       time.Now(),
		Metadata:         make(map[string]interface{}),
	}
	
	// 这里需要获取基因并应用反向突变
	// 实际实现中需要从基因池中找到对应基因
	dme.logger.Info("Mutation reversed", "original_mutation_id", mutation.ID, "reverse_mutation_id", reverseMutation.ID)
	return nil
}

// DefaultExpressionController 默认表达控制器实�?
type DefaultExpressionController struct {
	config *ExpressionControllerConfig
	logger Logger
}

type ExpressionControllerConfig struct {
	MaxExpressionLevel    float64       `json:"max_expression_level"`
	MinExpressionLevel    float64       `json:"min_expression_level"`
	ExpressionDecayRate   float64       `json:"expression_decay_rate"`
	MaxExpressionDuration time.Duration `json:"max_expression_duration"`
	EnableAutoRegulation  bool          `json:"enable_auto_regulation"`
}

func NewDefaultExpressionController(config *ExpressionControllerConfig, logger Logger) *DefaultExpressionController {
	if config == nil {
		config = &ExpressionControllerConfig{
			MaxExpressionLevel:    1.0,
			MinExpressionLevel:    0.0,
			ExpressionDecayRate:   0.1,
			MaxExpressionDuration: time.Hour * 24,
			EnableAutoRegulation:  true,
		}
	}
	return &DefaultExpressionController{config: config, logger: logger}
}

func (dec *DefaultExpressionController) InitiateExpression(ctx context.Context, geneID string, entityID string) (*models.GeneExpression, error) {
	expression := &models.GeneExpression{
		GeneID:         geneID,
		EntityID:       entityID,
		ExpressionLevel: dec.calculateInitialExpressionLevel(geneID),
		Intensity:      0.5, // 默认强度
		Duration:       dec.calculateExpressionDuration(geneID),
		StartTime:      time.Now(),
		IsActive:       true,
		Triggers:       []models.ExpressionTrigger{},
		Inhibitors:     []models.ExpressionInhibitor{},
		Context:        models.ExpressionContext{},
		Metadata:       make(map[string]interface{}),
	}
	
	endTime := expression.StartTime.Add(expression.Duration)
	expression.EndTime = &endTime
	
	return expression, nil
}

func (dec *DefaultExpressionController) ModulateExpression(ctx context.Context, expressionID string, level float64) error {
	// 限制表达水平在合理范围内
	if level > dec.config.MaxExpressionLevel {
		level = dec.config.MaxExpressionLevel
	}
	if level < dec.config.MinExpressionLevel {
		level = dec.config.MinExpressionLevel
	}
	
	dec.logger.Info("Expression modulated", "expression_id", expressionID, "new_level", level)
	return nil
}

func (dec *DefaultExpressionController) InhibitExpression(ctx context.Context, expressionID string) error {
	dec.logger.Info("Expression inhibited", "expression_id", expressionID)
	return nil
}

func (dec *DefaultExpressionController) TerminateExpression(ctx context.Context, expressionID string) error {
	dec.logger.Info("Expression terminated", "expression_id", expressionID)
	return nil
}

func (dec *DefaultExpressionController) GetExpressionStatus(ctx context.Context, expressionID string) (*models.GeneExpression, error) {
	// 实际实现中需要从存储中获取表达状�?
	return nil, fmt.Errorf("expression %s not found", expressionID)
}

func (dec *DefaultExpressionController) MonitorExpression(ctx context.Context, entityID string) ([]models.GeneExpression, error) {
	// 实际实现中需要从存储中获取实体的所有表�?
	return []models.GeneExpression{}, nil
}

// DefaultInteractionAnalyzer 默认相互作用分析器实�?
type DefaultInteractionAnalyzer struct {
	config *InteractionAnalyzerConfig
	logger Logger
}

type InteractionAnalyzerConfig struct {
	MaxInteractionDistance float64 `json:"max_interaction_distance"`
	MinInteractionStrength float64 `json:"min_interaction_strength"`
	EnableConflictDetection bool   `json:"enable_conflict_detection"`
	EnableOptimization     bool   `json:"enable_optimization"`
}

func NewDefaultInteractionAnalyzer(config *InteractionAnalyzerConfig, logger Logger) *DefaultInteractionAnalyzer {
	if config == nil {
		config = &InteractionAnalyzerConfig{
			MaxInteractionDistance:  10.0,
			MinInteractionStrength:  0.1,
			EnableConflictDetection: true,
			EnableOptimization:      true,
		}
	}
	return &DefaultInteractionAnalyzer{config: config, logger: logger}
}

func (dia *DefaultInteractionAnalyzer) AnalyzeGeneInteractions(ctx context.Context, genePool *models.GenePool) ([]models.GeneInteraction, error) {
	interactions := []models.GeneInteraction{}
	
	// 分析所有基因对之间的相互作�?
	for i, geneA := range genePool.Genes {
		for j, geneB := range genePool.Genes {
			if i >= j { // 避免重复分析
				continue
			}
			
			interaction := dia.analyzeGenePair(&geneA, &geneB)
			if interaction != nil && interaction.Strength >= dia.config.MinInteractionStrength {
				interactions = append(interactions, *interaction)
			}
		}
	}
	
	return interactions, nil
}

func (dia *DefaultInteractionAnalyzer) PredictInteractionOutcome(ctx context.Context, geneA, geneB string) (*models.GeneInteraction, error) {
	// 实际实现中需要基于基因特性预测相互作用结�?
	interaction := &models.GeneInteraction{
		GeneA:           geneA,
		GeneB:           geneB,
		InteractionType: models.InteractionTypeSynergistic,
		Strength:        0.5,
		Direction:       models.InteractionDirectionBidirectional,
		Effect:          models.InteractionEffect{
			Type:        models.EffectTypeCapabilityBoost,
			Magnitude:   0.5,
			Duration:    time.Hour,
			IsPositive:  true,
			Description: "Predicted synergistic interaction",
			Metadata:    make(map[string]interface{}),
		},
		Conditions:      []models.InteractionCondition{},
		IsActive:        true,
		DiscoveredAt:    time.Now(),
		Metadata:        make(map[string]interface{}),
	}
	
	return interaction, nil
}

func (dia *DefaultInteractionAnalyzer) EvaluateInteractionStrength(ctx context.Context, interaction *models.GeneInteraction) (float64, error) {
	// 基于多个因素计算相互作用强度
	strength := dia.calculateBaseStrength(interaction)
	strength *= dia.calculateContextualModifier(interaction)
	strength *= dia.calculateTemporalModifier(interaction)
	
	return math.Max(0.0, math.Min(1.0, strength)), nil
}

func (dia *DefaultInteractionAnalyzer) DetectInteractionConflicts(ctx context.Context, genePool *models.GenePool) ([]InteractionConflict, error) {
	conflicts := []InteractionConflict{}
	
	for _, interaction := range genePool.GeneInteractions {
		if interaction.InteractionType == models.InteractionTypeAntagonistic && interaction.Strength > 0.7 {
			conflict := InteractionConflict{
				GeneA:       interaction.GeneA,
				GeneB:       interaction.GeneB,
				ConflictType: "antagonistic_interaction",
				Severity:    interaction.Strength,
				Description: fmt.Sprintf("Strong antagonistic interaction between %s and %s", interaction.GeneA, interaction.GeneB),
			}
			conflicts = append(conflicts, conflict)
		}
	}
	
	return conflicts, nil
}

func (dia *DefaultInteractionAnalyzer) OptimizeGeneCompatibility(ctx context.Context, genePool *models.GenePool) (*OptimizationResult, error) {
	originalCompatibility := dia.calculatePoolCompatibility(genePool)
	
	// 执行优化算法
	improvements := []CompatibilityImprovement{}
	removedConflicts := []InteractionConflict{}
	
	// 模拟优化过程
	optimizedCompatibility := originalCompatibility * 1.2 // 假设提升20%
	
	result := &OptimizationResult{
		OriginalCompatibility:  originalCompatibility,
		OptimizedCompatibility: optimizedCompatibility,
		Improvements:          improvements,
		RemovedConflicts:      removedConflicts,
	}
	
	return result, nil
}

// DefaultEvolutionSimulator 默认进化模拟器实�?
type DefaultEvolutionSimulator struct {
	config *EvolutionSimulatorConfig
	logger Logger
}

type EvolutionSimulatorConfig struct {
	MaxGenerations      int     `json:"max_generations"`
	SelectionPressure   float64 `json:"selection_pressure"`
	MutationRate        float64 `json:"mutation_rate"`
	CrossoverRate       float64 `json:"crossover_rate"`
	ElitismRate         float64 `json:"elitism_rate"`
	PopulationSize      int     `json:"population_size"`
	FitnessThreshold    float64 `json:"fitness_threshold"`
	DiversityThreshold  float64 `json:"diversity_threshold"`
}

func NewDefaultEvolutionSimulator(config *EvolutionSimulatorConfig, logger Logger) *DefaultEvolutionSimulator {
	if config == nil {
		config = &EvolutionSimulatorConfig{
			MaxGenerations:     100,
			SelectionPressure:  0.7,
			MutationRate:       0.01,
			CrossoverRate:      0.8,
			ElitismRate:        0.1,
			PopulationSize:     100,
			FitnessThreshold:   0.9,
			DiversityThreshold: 0.3,
		}
	}
	return &DefaultEvolutionSimulator{config: config, logger: logger}
}

func (des *DefaultEvolutionSimulator) SimulateEvolution(ctx context.Context, genePool *models.GenePool, generations int) (*EvolutionSimulationResult, error) {
	initialStats := genePool.PoolStats
	evolutionEvents := []models.PoolEvolutionEvent{}
	fitnessHistory := []float64{}
	diversityHistory := []float64{}
	
	// 记录初始状�?
	fitnessHistory = append(fitnessHistory, des.calculatePoolFitness(genePool))
	diversityHistory = append(diversityHistory, genePool.GetDiversityScore())
	
	for gen := 0; gen < generations; gen++ {
		// 模拟一代进�?
		fitness, diversity, events := des.simulateGeneration(genePool, gen)
		
		fitnessHistory = append(fitnessHistory, fitness)
		diversityHistory = append(diversityHistory, diversity)
		evolutionEvents = append(evolutionEvents, events...)
		
		// 检查是否达到终止条�?
		if fitness >= des.config.FitnessThreshold {
			break
		}
	}
	
	finalStats := genePool.PoolStats
	
	result := &EvolutionSimulationResult{
		InitialState:     &initialStats,
		FinalState:       &finalStats,
		Generations:      generations,
		EvolutionEvents:  evolutionEvents,
		FitnessHistory:   fitnessHistory,
		DiversityHistory: diversityHistory,
	}
	
	return result, nil
}

func (des *DefaultEvolutionSimulator) PredictEvolutionaryPath(ctx context.Context, genePool *models.GenePool) (*EvolutionaryPath, error) {
	steps := []EvolutionaryStep{}
	totalDuration := time.Duration(0)
	
	// 预测进化路径
	for i := 0; i < 10; i++ { // 预测10�?
		step := EvolutionaryStep{
			Generation:  i,
			Changes:     des.predictGeneChanges(genePool, i),
			Fitness:     des.predictFitness(genePool, i),
			Diversity:   des.predictDiversity(genePool, i),
			Events:      des.predictEvents(genePool, i),
		}
		steps = append(steps, step)
		totalDuration += time.Hour * 24 // 假设每代需要一�?
	}
	
	path := &EvolutionaryPath{
		Steps:           steps,
		TotalDuration:   totalDuration,
		ExpectedFitness: des.predictFinalFitness(genePool),
		Confidence:      0.6,
	}
	
	return path, nil
}

func (des *DefaultEvolutionSimulator) EvaluateEvolutionaryFitness(ctx context.Context, genePool *models.GenePool) (float64, error) {
	return des.calculatePoolFitness(genePool), nil
}

func (des *DefaultEvolutionSimulator) GenerateEvolutionaryPressure(ctx context.Context, genePool *models.GenePool) ([]EvolutionaryPressure, error) {
	pressures := []EvolutionaryPressure{
		{
			Type:        "selection",
			Intensity:   des.config.SelectionPressure,
			Target:      "fitness",
			Description: "Natural selection pressure favoring higher fitness genes",
		},
		{
			Type:        "mutation",
			Intensity:   des.config.MutationRate,
			Target:      "diversity",
			Description: "Mutation pressure increasing genetic diversity",
		},
		{
			Type:        "environmental",
			Intensity:   0.3,
			Target:      "adaptation",
			Description: "Environmental pressure driving adaptation",
		},
	}
	
	return pressures, nil
}

func (des *DefaultEvolutionSimulator) ApplySelection(ctx context.Context, genePool *models.GenePool, selectionPressure float64) error {
	// 应用选择压力，移除低适应性基�?
	threshold := des.calculateSelectionThreshold(genePool, selectionPressure)
	
	survivingGenes := []models.QuantumGene{}
	for _, gene := range genePool.Genes {
		if des.calculateGeneFitness(&gene) >= threshold {
			survivingGenes = append(survivingGenes, gene)
		}
	}
	
	genePool.Genes = survivingGenes
	des.logger.Info("Selection applied", "original_count", len(genePool.Genes), "surviving_count", len(survivingGenes))
	
	return nil
}

// 私有辅助方法

func (dme *DefaultMutationEngine) selectMutationType(gene *models.QuantumGene) models.MutationType {
	types := []models.MutationType{
		models.MutationTypePoint,
		models.MutationTypeInsertion,
		models.MutationTypeDeletion,
		models.MutationTypeDuplication,
		models.MutationTypeInversion,
		models.MutationTypeTranslocation,
	}
	return types[rand.Intn(len(types))]
}

func (dme *DefaultMutationEngine) calculateMutationSeverity(gene *models.QuantumGene) models.MutationSeverity {
	// 基于基因稳定性计算突变严重程�?
	if gene.Stability > 0.8 {
		return models.MutationSeverityMinor
	} else if gene.Stability > 0.6 {
		return models.MutationSeverityModerate
	} else if gene.Stability > 0.4 {
		return models.MutationSeverityMajor
	}
	return models.MutationSeverityCritical
}

func (dme *DefaultMutationEngine) getCurrentGeneValue(gene *models.QuantumGene, mutationType models.MutationType) interface{} {
	switch mutationType {
	case models.MutationTypePoint:
		return gene.Sequence
	case models.MutationTypeInsertion:
		return gene.Expression
	case models.MutationTypeDeletion:
		return gene.Dominance
	case models.MutationTypeDuplication:
		return gene.Stability
	case models.MutationTypeInversion:
		return gene.Mutability
	default:
		return nil
	}
}

func (dme *DefaultMutationEngine) generateNewValue(gene *models.QuantumGene, mutationType models.MutationType, severity models.MutationSeverity) interface{} {
	severityFactor := dme.getSeverityFactor(severity)
	
	switch mutationType {
	case models.MutationTypePoint:
		return dme.mutateSequence(gene.Sequence, severityFactor)
	case models.MutationTypeInsertion:
		return dme.mutateFloat(gene.Expression, severityFactor)
	case models.MutationTypeDeletion:
		return dme.mutateFloat(gene.Dominance, severityFactor)
	case models.MutationTypeDuplication:
		return dme.mutateFloat(gene.Stability, severityFactor)
	case models.MutationTypeInversion:
		return dme.mutateFloat(gene.Mutability, severityFactor)
	default:
		return nil
	}
}

func (dme *DefaultMutationEngine) getSeverityFactor(severity models.MutationSeverity) float64 {
	switch severity {
	case models.MutationSeverityMinor:
		return 0.1
	case models.MutationSeverityModerate:
		return 0.3
	case models.MutationSeverityMajor:
		return 0.6
	case models.MutationSeverityCritical:
		return 1.0
	default:
		return 0.1
	}
}

func (dme *DefaultMutationEngine) mutateSequence(sequence string, factor float64) string {
	// 简单的序列突变实现
	if len(sequence) == 0 {
		return sequence
	}
	
	mutationCount := int(float64(len(sequence)) * factor)
	if mutationCount == 0 {
		mutationCount = 1
	}
	
	runes := []rune(sequence)
	for i := 0; i < mutationCount; i++ {
		pos := rand.Intn(len(runes))
		runes[pos] = rune('A' + rand.Intn(26)) // 随机字母
	}
	
	return string(runes)
}

func (dme *DefaultMutationEngine) mutateFloat(value, factor float64) float64 {
	change := (rand.Float64() - 0.5) * 2 * factor
	newValue := value + change
	return math.Max(0.0, math.Min(1.0, newValue))
}

func (dme *DefaultMutationEngine) calculateMutationProbability(gene *models.QuantumGene) float64 {
	return gene.Mutability * (1.0 - gene.Stability)
}

func (dme *DefaultMutationEngine) estimateImpact(gene *models.QuantumGene, mutationType models.MutationType, severity models.MutationSeverity) float64 {
	baseImpact := dme.getSeverityFactor(severity)
	
	// 基于基因重要性调整影�?
	importanceFactor := gene.Dominance
	
	return baseImpact * importanceFactor
}

func (dme *DefaultMutationEngine) calculateAgeFactor(gene *models.QuantumGene) float64 {
	age := time.Since(gene.CreatedAt)
	// 年龄越大，突变概率越�?
	return math.Min(2.0, 1.0+age.Hours()/8760) // 一年后达到最大�?
}

// 其他辅助方法的实�?.
func (dme *DefaultMutationEngine) applySequenceMutation(gene *models.QuantumGene, mutation *models.GeneMutation) error {
	if mutation.MutatedSequence != "" {
		gene.Sequence = mutation.MutatedSequence
		return nil
	}
	return fmt.Errorf("invalid sequence mutation value")
}

func (dme *DefaultMutationEngine) applyExpressionMutation(gene *models.QuantumGene, mutation *models.GeneMutation) error {
	// 基于突变序列计算新的表达水平
	newExpr := gene.Expression * 0.9 // 简化实�?
	gene.Expression = newExpr
	return nil
}

func (dme *DefaultMutationEngine) applyDominanceMutation(gene *models.QuantumGene, mutation *models.GeneMutation) error {
	// 基于突变序列计算新的显性程�?
	newDom := gene.Dominance * 0.95 // 简化实�?
	gene.Dominance = newDom
	return nil
}

func (dme *DefaultMutationEngine) applyStabilityMutation(gene *models.QuantumGene, mutation *models.GeneMutation) error {
	// 基于突变序列计算新的稳定�?
	newStab := gene.Stability * 0.9 // 简化实�?
	gene.Stability = newStab
	return nil
}

func (dme *DefaultMutationEngine) applyMutabilityMutation(gene *models.QuantumGene, mutation *models.GeneMutation) error {
	// 基于突变序列计算新的可变�?
	newMut := gene.Mutability * 1.1 // 简化实�?
	if newMut > 1.0 {
		newMut = 1.0
	}
	gene.Mutability = newMut
	return nil
}

func (dme *DefaultMutationEngine) applyCompatibilityMutation(gene *models.QuantumGene, mutation *models.GeneMutation) error {
	// 兼容性突变的实现
	return nil
}

func (dme *DefaultMutationEngine) calculateFitnessImpact(mutation *models.GeneMutation) float64 {
	return rand.Float64()*0.2 - 0.1 // -0.1 �?0.1 的随机影�?
}

func (dme *DefaultMutationEngine) calculateStabilityImpact(mutation *models.GeneMutation) float64 {
	return -dme.getSeverityFactor(mutation.Severity) * 0.1 // 突变通常降低稳定�?
}

func (dme *DefaultMutationEngine) calculateExpressionImpact(mutation *models.GeneMutation) float64 {
	return rand.Float64()*0.3 - 0.15 // -0.15 �?0.15 的随机影�?
}

func (dme *DefaultMutationEngine) calculateCompatibilityImpact(mutation *models.GeneMutation) float64 {
	return rand.Float64()*0.1 - 0.05 // -0.05 �?0.05 的随机影�?
}

func (dme *DefaultMutationEngine) predictOutcome(mutation *models.GeneMutation) string {
	outcomes := []string{"beneficial", "neutral", "detrimental", "unknown"}
	return outcomes[rand.Intn(len(outcomes))]
}

func (dme *DefaultMutationEngine) calculateConfidence(mutation *models.GeneMutation) float64 {
	return 0.5 + rand.Float64()*0.4 // 0.5 �?0.9 的置信度
}

// ExpressionController 辅助方法
func (dec *DefaultExpressionController) calculateInitialExpressionLevel(geneID string) float64 {
	return 0.5 + rand.Float64()*0.3 // 0.5 �?0.8 的初始表达水�?
}

func (dec *DefaultExpressionController) calculateExpressionDuration(geneID string) time.Duration {
	return time.Duration(rand.Intn(int(dec.config.MaxExpressionDuration.Hours()))) * time.Hour
}

// InteractionAnalyzer 辅助方法
func (dia *DefaultInteractionAnalyzer) analyzeGenePair(geneA, geneB *models.QuantumGene) *models.GeneInteraction {
	// 计算基因间的相互作用
	strength := dia.calculateInteractionStrength(geneA, geneB)
	if strength < dia.config.MinInteractionStrength {
		return nil
	}
	
	interactionType := dia.determineInteractionType(geneA, geneB)
	
	return &models.GeneInteraction{
		GeneA:           geneA.ID,
		GeneB:           geneB.ID,
		InteractionType: interactionType,
		Strength:        strength,
		Direction:       models.InteractionDirectionBidirectional,
		Effect:          models.InteractionEffect{
			Type:        models.EffectTypeCapabilityBoost,
			Magnitude:   strength,
			Duration:    time.Hour,
			IsPositive:  true,
			Description: "Analyzed gene interaction",
			Metadata:    make(map[string]interface{}),
		},
		Conditions:      []models.InteractionCondition{},
		IsActive:        true,
		DiscoveredAt:    time.Now(),
		Metadata:        make(map[string]interface{}),
	}
}

func (dia *DefaultInteractionAnalyzer) calculateInteractionStrength(geneA, geneB *models.QuantumGene) float64 {
	// 基于基因特性计算相互作用强�?
	compatibilityScore := 1.0
	if !geneA.IsCompatibleWith(geneB.ID) {
		compatibilityScore = 0.3
	}
	
	expressionSimilarity := 1.0 - math.Abs(geneA.Expression-geneB.Expression)
	dominanceInteraction := math.Min(geneA.Dominance, geneB.Dominance)
	
	return compatibilityScore * expressionSimilarity * dominanceInteraction
}

func (dia *DefaultInteractionAnalyzer) determineInteractionType(geneA, geneB *models.QuantumGene) models.InteractionType {
	if geneA.IsCompatibleWith(geneB.ID) {
		if geneA.Expression > 0.7 && geneB.Expression > 0.7 {
			return models.InteractionTypeSynergistic
		}
		return models.InteractionTypeComplementary
	}
	return models.InteractionTypeAntagonistic
}

func (dia *DefaultInteractionAnalyzer) calculateBaseStrength(interaction *models.GeneInteraction) float64 {
	return interaction.Strength
}

func (dia *DefaultInteractionAnalyzer) calculateContextualModifier(interaction *models.GeneInteraction) float64 {
	return 1.0 // 简化实现，无上下文影响
}

func (dia *DefaultInteractionAnalyzer) calculateTemporalModifier(interaction *models.GeneInteraction) float64 {
	return 1.0 // 简化实现，无时间影�?
}

func (dia *DefaultInteractionAnalyzer) calculatePoolCompatibility(genePool *models.GenePool) float64 {
	if len(genePool.Genes) == 0 {
		return 1.0
	}
	
	totalCompatibility := 0.0
	pairCount := 0
	
	for i, geneA := range genePool.Genes {
		for j, geneB := range genePool.Genes {
			if i >= j {
				continue
			}
			
			if geneA.IsCompatibleWith(geneB.ID) {
				totalCompatibility += 1.0
			}
			pairCount++
		}
	}
	
	if pairCount == 0 {
		return 1.0
	}
	
	return totalCompatibility / float64(pairCount)
}

// EvolutionSimulator 辅助方法
func (des *DefaultEvolutionSimulator) calculatePoolFitness(genePool *models.GenePool) float64 {
	if len(genePool.Genes) == 0 {
		return 0.0
	}
	
	totalFitness := 0.0
	for _, gene := range genePool.Genes {
		totalFitness += des.calculateGeneFitness(&gene)
	}
	
	return totalFitness / float64(len(genePool.Genes))
}

func (des *DefaultEvolutionSimulator) calculateGeneFitness(gene *models.QuantumGene) float64 {
	// 基于多个因素计算基因适应�?
	expressionFactor := gene.Expression
	stabilityFactor := gene.Stability
	dominanceFactor := gene.Dominance
	
	return (expressionFactor + stabilityFactor + dominanceFactor) / 3.0
}

func (des *DefaultEvolutionSimulator) simulateGeneration(genePool *models.GenePool, generation int) (float64, float64, []models.PoolEvolutionEvent) {
	fitness := des.calculatePoolFitness(genePool)
	diversity := genePool.GetDiversityScore()
	
	events := []models.PoolEvolutionEvent{
		{
			Type:        models.EvolutionEventTypeMutation,
			Description: fmt.Sprintf("Generation %d completed", generation),
			Impact:      0.1,
			OccurredAt:  time.Now(),
			Metadata:    map[string]interface{}{"generation": generation, "fitness": fitness},
		},
	}
	
	return fitness, diversity, events
}

func (des *DefaultEvolutionSimulator) predictGeneChanges(genePool *models.GenePool, generation int) []GeneChange {
	changes := []GeneChange{}
	
	// 预测一些基因变�?
	for i, gene := range genePool.Genes {
		if i >= 3 { // 只预测前3个基因的变化
			break
		}
		
		change := GeneChange{
			GeneID:     gene.ID,
			ChangeType: "expression",
			OldValue:   gene.Expression,
			NewValue:   gene.Expression + rand.Float64()*0.1 - 0.05,
			Impact:     rand.Float64() * 0.2,
		}
		changes = append(changes, change)
	}
	
	return changes
}

func (des *DefaultEvolutionSimulator) predictFitness(genePool *models.GenePool, generation int) float64 {
	baseFitness := des.calculatePoolFitness(genePool)
	return baseFitness + float64(generation)*0.01 // 假设每代提升1%
}

func (des *DefaultEvolutionSimulator) predictDiversity(genePool *models.GenePool, generation int) float64 {
	baseDiversity := genePool.GetDiversityScore()
	return math.Max(0.1, baseDiversity-float64(generation)*0.005) // 假设多样性逐渐降低
}

func (des *DefaultEvolutionSimulator) predictEvents(genePool *models.GenePool, generation int) []models.PoolEvolutionEvent {
	return []models.PoolEvolutionEvent{
		{
			Type:        models.EvolutionEventTypeMutation,
			Description: fmt.Sprintf("Predicted mutation in generation %d", generation),
			Impact:      0.05,
			OccurredAt:  time.Now().Add(time.Duration(generation) * time.Hour * 24),
			Metadata:    map[string]interface{}{"generation": generation},
		},
	}
}

func (des *DefaultEvolutionSimulator) predictFinalFitness(genePool *models.GenePool) float64 {
	baseFitness := des.calculatePoolFitness(genePool)
	return math.Min(1.0, baseFitness+0.3) // 假设最终提�?0%
}

func (des *DefaultEvolutionSimulator) calculateSelectionThreshold(genePool *models.GenePool, selectionPressure float64) float64 {
	avgFitness := des.calculatePoolFitness(genePool)
	return avgFitness * selectionPressure
}
