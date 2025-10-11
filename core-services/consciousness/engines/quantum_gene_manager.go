package engines

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
)

// QuantumGeneManager 量子基因管理�?
type QuantumGeneManager struct {
	mu                   sync.RWMutex
	config               *QuantumGeneManagerConfig
	genePools            map[string]*models.GenePool
	geneExpressions      map[string][]models.GeneExpression
	mutationEngine       MutationEngine
	expressionController ExpressionController
	interactionAnalyzer  InteractionAnalyzer
	evolutionSimulator   EvolutionSimulator
	geneRepository       GeneRepository
	isRunning            bool
	stopChan             chan struct{}
	logger               Logger
}

// QuantumGeneManagerConfig 量子基因管理器配�?
type QuantumGeneManagerConfig struct {
	MaxGenePools                int                     `json:"max_gene_pools"`
	MaxGenesPerPool             int                     `json:"max_genes_per_pool"`
	MutationRate                float64                 `json:"mutation_rate"`
	ExpressionUpdateInterval    time.Duration           `json:"expression_update_interval"`
	EvolutionSimulationInterval time.Duration           `json:"evolution_simulation_interval"`
	EnableAutoMutation          bool                    `json:"enable_auto_mutation"`
	EnableExpressionControl     bool                    `json:"enable_expression_control"`
	EnableInteractionAnalysis   bool                    `json:"enable_interaction_analysis"`
	EnableEvolutionSimulation   bool                    `json:"enable_evolution_simulation"`
	GeneStabilityThreshold      float64                 `json:"gene_stability_threshold"`
	ExpressionThreshold         float64                 `json:"expression_threshold"`
	MutationSeverityLimit       models.MutationSeverity `json:"mutation_severity_limit"`
}

// MutationEngine 突变引擎接口
type MutationEngine interface {
	GenerateMutation(ctx context.Context, gene *models.QuantumGene) (*models.GeneMutation, error)
	ApplyMutation(ctx context.Context, gene *models.QuantumGene, mutation *models.GeneMutation) error
	EvaluateMutationImpact(ctx context.Context, mutation *models.GeneMutation) (*models.MutationImpact, error)
	PredictMutationProbability(ctx context.Context, gene *models.QuantumGene) (float64, error)
	ReverseMutation(ctx context.Context, mutation *models.GeneMutation) error
}

// ExpressionController 表达控制器接�?
type ExpressionController interface {
	InitiateExpression(ctx context.Context, geneID string, entityID string) (*models.GeneExpression, error)
	ModulateExpression(ctx context.Context, expressionID string, level float64) error
	InhibitExpression(ctx context.Context, expressionID string) error
	TerminateExpression(ctx context.Context, expressionID string) error
	GetExpressionStatus(ctx context.Context, expressionID string) (*models.GeneExpression, error)
	MonitorExpression(ctx context.Context, entityID string) ([]models.GeneExpression, error)
}

// InteractionAnalyzer 相互作用分析器接�?
type InteractionAnalyzer interface {
	AnalyzeGeneInteractions(ctx context.Context, genePool *models.GenePool) ([]models.GeneInteraction, error)
	PredictInteractionOutcome(ctx context.Context, geneA, geneB string) (*models.GeneInteraction, error)
	EvaluateInteractionStrength(ctx context.Context, interaction *models.GeneInteraction) (float64, error)
	DetectInteractionConflicts(ctx context.Context, genePool *models.GenePool) ([]InteractionConflict, error)
	OptimizeGeneCompatibility(ctx context.Context, genePool *models.GenePool) (*OptimizationResult, error)
}

// EvolutionSimulator 进化模拟器接�?
type EvolutionSimulator interface {
	SimulateEvolution(ctx context.Context, genePool *models.GenePool, generations int) (*EvolutionSimulationResult, error)
	PredictEvolutionaryPath(ctx context.Context, genePool *models.GenePool) (*EvolutionaryPath, error)
	EvaluateEvolutionaryFitness(ctx context.Context, genePool *models.GenePool) (float64, error)
	GenerateEvolutionaryPressure(ctx context.Context, genePool *models.GenePool) ([]EvolutionaryPressure, error)
	ApplySelection(ctx context.Context, genePool *models.GenePool, selectionPressure float64) error
}

// GeneRepository 基因仓库接口
type GeneRepository interface {
	SaveGene(ctx context.Context, gene *models.QuantumGene) error
	GetGene(ctx context.Context, geneID string) (*models.QuantumGene, error)
	UpdateGene(ctx context.Context, gene *models.QuantumGene) error
	DeleteGene(ctx context.Context, geneID string) error
	SaveGenePool(ctx context.Context, pool *models.GenePool) error
	GetGenePool(ctx context.Context, poolID string) (*models.GenePool, error)
	UpdateGenePool(ctx context.Context, pool *models.GenePool) error
	DeleteGenePool(ctx context.Context, poolID string) error
	SaveMutation(ctx context.Context, mutation *models.GeneMutation) error
	GetMutationHistory(ctx context.Context, geneID string) ([]models.GeneMutation, error)
	SaveExpression(ctx context.Context, expression *models.GeneExpression) error
	GetActiveExpressions(ctx context.Context, entityID string) ([]models.GeneExpression, error)
}

// 辅助结构 InteractionConflict 表示相互作用冲突
type InteractionConflict struct {
	GeneA        string  `json:"gene_a"`
	GeneB        string  `json:"gene_b"`
	ConflictType string  `json:"conflict_type"`
	Severity     float64 `json:"severity"`
	Description  string  `json:"description"`
}

type OptimizationResult struct {
	OriginalCompatibility  float64                    `json:"original_compatibility"`
	OptimizedCompatibility float64                    `json:"optimized_compatibility"`
	Improvements           []CompatibilityImprovement `json:"improvements"`
	RemovedConflicts       []InteractionConflict      `json:"removed_conflicts"`
}

type CompatibilityImprovement struct {
	GeneID      string  `json:"gene_id"`
	Improvement float64 `json:"improvement"`
	Method      string  `json:"method"`
}

type EvolutionSimulationResult struct {
	InitialState     *models.GenePoolStats       `json:"initial_state"`
	FinalState       *models.GenePoolStats       `json:"final_state"`
	Generations      int                         `json:"generations"`
	EvolutionEvents  []models.PoolEvolutionEvent `json:"evolution_events"`
	FitnessHistory   []float64                   `json:"fitness_history"`
	DiversityHistory []float64                   `json:"diversity_history"`
}

type EvolutionaryPath struct {
	Steps           []EvolutionaryStep `json:"steps"`
	TotalDuration   time.Duration      `json:"total_duration"`
	ExpectedFitness float64            `json:"expected_fitness"`
	Confidence      float64            `json:"confidence"`
}

type EvolutionaryStep struct {
	Generation int                         `json:"generation"`
	Changes    []GeneChange                `json:"changes"`
	Fitness    float64                     `json:"fitness"`
	Diversity  float64                     `json:"diversity"`
	Events     []models.PoolEvolutionEvent `json:"events"`
}

type GeneChange struct {
	GeneID     string      `json:"gene_id"`
	ChangeType string      `json:"change_type"`
	OldValue   interface{} `json:"old_value"`
	NewValue   interface{} `json:"new_value"`
	Impact     float64     `json:"impact"`
}

type EvolutionaryPressure struct {
	Type        string  `json:"type"`
	Intensity   float64 `json:"intensity"`
	Target      string  `json:"target"`
	Description string  `json:"description"`
}

// NewQuantumGeneManager 创建新的量子基因管理�?
func NewQuantumGeneManager(config *QuantumGeneManagerConfig, logger Logger) *QuantumGeneManager {
	if config == nil {
		config = &QuantumGeneManagerConfig{
			MaxGenePools:                100,
			MaxGenesPerPool:             1000,
			MutationRate:                0.01,
			ExpressionUpdateInterval:    time.Minute * 5,
			EvolutionSimulationInterval: time.Hour,
			EnableAutoMutation:          true,
			EnableExpressionControl:     true,
			EnableInteractionAnalysis:   true,
			EnableEvolutionSimulation:   true,
			GeneStabilityThreshold:      0.7,
			ExpressionThreshold:         0.3,
			MutationSeverityLimit:       models.MutationSeverityMajor,
		}
	}

	return &QuantumGeneManager{
		config:          config,
		genePools:       make(map[string]*models.GenePool),
		geneExpressions: make(map[string][]models.GeneExpression),
		stopChan:        make(chan struct{}),
		logger:          logger,
	}
}

// SetDependencies 设置依赖组件
func (qgm *QuantumGeneManager) SetDependencies(
	mutationEngine MutationEngine,
	expressionController ExpressionController,
	interactionAnalyzer InteractionAnalyzer,
	evolutionSimulator EvolutionSimulator,
	geneRepository GeneRepository,
) {
	qgm.mu.Lock()
	defer qgm.mu.Unlock()

	qgm.mutationEngine = mutationEngine
	qgm.expressionController = expressionController
	qgm.interactionAnalyzer = interactionAnalyzer
	qgm.evolutionSimulator = evolutionSimulator
	qgm.geneRepository = geneRepository
}

// Start 启动量子基因管理�?
func (qgm *QuantumGeneManager) Start(ctx context.Context) error {
	qgm.mu.Lock()
	defer qgm.mu.Unlock()

	if qgm.isRunning {
		return fmt.Errorf("quantum gene manager is already running")
	}

	if err := qgm.validateDependencies(); err != nil {
		return fmt.Errorf("failed to validate dependencies: %w", err)
	}

	qgm.isRunning = true
	qgm.stopChan = make(chan struct{})

	// 启动后台处理协程
	go qgm.runBackgroundProcesses(ctx)

	qgm.logger.Info("Quantum gene manager started successfully")
	return nil
}

// Stop 停止量子基因管理�?
func (qgm *QuantumGeneManager) Stop() error {
	qgm.mu.Lock()
	defer qgm.mu.Unlock()

	if !qgm.isRunning {
		return fmt.Errorf("quantum gene manager is not running")
	}

	close(qgm.stopChan)
	qgm.isRunning = false

	qgm.logger.Info("Quantum gene manager stopped successfully")
	return nil
}

// CreateGenePool 创建基因�?
func (qgm *QuantumGeneManager) CreateGenePool(ctx context.Context, ownerID, name, description string) (*models.GenePool, error) {
	qgm.mu.Lock()
	defer qgm.mu.Unlock()

	if len(qgm.genePools) >= qgm.config.MaxGenePools {
		return nil, fmt.Errorf("maximum number of gene pools reached: %d", qgm.config.MaxGenePools)
	}

	poolID := qgm.generatePoolID()
	pool := &models.GenePool{
		ID:               poolID,
		Name:             name,
		Description:      description,
		OwnerID:          ownerID,
		Genes:            []models.QuantumGene{},
		ActiveGenes:      []string{},
		DormantGenes:     []string{},
		GeneInteractions: []models.GeneInteraction{},
		PoolStats: models.GenePoolStats{
			TotalGenes:      0,
			ActiveGenes:     0,
			DormantGenes:    0,
			MutatedGenes:    0,
			DiversityIndex:  0.0,
			StabilityIndex:  1.0,
			EvolutionRate:   0.0,
			MutationRate:    qgm.config.MutationRate,
			ExpressionLevel: 0.0,
			LastUpdated:     time.Now(),
		},
		EvolutionHistory: []models.PoolEvolutionEvent{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}

	qgm.genePools[poolID] = pool

	// 保存到仓�?
	if qgm.geneRepository != nil {
		if err := qgm.geneRepository.SaveGenePool(ctx, pool); err != nil {
			qgm.logger.Error("Failed to save gene pool to repository", err, "pool_id", poolID)
		}
	}

	qgm.logger.Info("Created gene pool", "pool_id", poolID, "owner_id", ownerID, "name", name)
	return pool, nil
}

// AddGeneToPool 向基因池添加基因
func (qgm *QuantumGeneManager) AddGeneToPool(ctx context.Context, poolID string, gene *models.QuantumGene) error {
	qgm.mu.Lock()
	defer qgm.mu.Unlock()

	pool, exists := qgm.genePools[poolID]
	if !exists {
		return fmt.Errorf("gene pool %s not found", poolID)
	}

	if len(pool.Genes) >= qgm.config.MaxGenesPerPool {
		return fmt.Errorf("maximum number of genes per pool reached: %d", qgm.config.MaxGenesPerPool)
	}

	// 检查基因兼容�?
	if err := qgm.checkGeneCompatibility(pool, gene); err != nil {
		return fmt.Errorf("gene compatibility check failed: %w", err)
	}

	// 设置基因创建时间
	gene.CreatedAt = time.Now()
	gene.UpdatedAt = time.Now()

	// 添加基因到池
	pool.Genes = append(pool.Genes, *gene)

	// 更新基因池统计信�?
	qgm.updatePoolStats(pool)

	// 如果基因是活跃的，添加到活跃基因列表
	if gene.IsActive() {
		pool.ActiveGenes = append(pool.ActiveGenes, gene.ID)
	} else {
		pool.DormantGenes = append(pool.DormantGenes, gene.ID)
	}

	pool.UpdatedAt = time.Now()

	// 保存基因和基因池
	if qgm.geneRepository != nil {
		if err := qgm.geneRepository.SaveGene(ctx, gene); err != nil {
			qgm.logger.Error("Failed to save gene to repository", err, "gene_id", gene.ID)
		}
		if err := qgm.geneRepository.UpdateGenePool(ctx, pool); err != nil {
			qgm.logger.Error("Failed to update gene pool in repository", err, "pool_id", poolID)
		}
	}

	qgm.logger.Info("Added gene to pool", "gene_id", gene.ID, "pool_id", poolID, "gene_type", gene.Type)
	return nil
}

// ExpressGene 表达基因
func (qgm *QuantumGeneManager) ExpressGene(ctx context.Context, geneID, entityID string, duration time.Duration) (*models.GeneExpression, error) {
	if qgm.expressionController == nil {
		return nil, fmt.Errorf("expression controller is not available")
	}

	// 创建基因表达
	expression, err := qgm.expressionController.InitiateExpression(ctx, geneID, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate gene expression: %w", err)
	}

	// 设置表达持续时间
	expression.Duration = duration
	endTime := time.Now().Add(duration)
	expression.EndTime = &endTime

	// 添加到表达列�?
	qgm.mu.Lock()
	qgm.geneExpressions[entityID] = append(qgm.geneExpressions[entityID], *expression)
	qgm.mu.Unlock()

	// 保存表达记录
	if qgm.geneRepository != nil {
		if err := qgm.geneRepository.SaveExpression(ctx, expression); err != nil {
			qgm.logger.Error("Failed to save gene expression to repository", err, "expression_id", expression.GeneID)
		}
	}

	qgm.logger.Info("Gene expression initiated", "gene_id", geneID, "entity_id", entityID, "duration", duration)
	return expression, nil
}

// MutateGene 突变基因
func (qgm *QuantumGeneManager) MutateGene(ctx context.Context, geneID string) (*models.GeneMutation, error) {
	if qgm.mutationEngine == nil {
		return nil, fmt.Errorf("mutation engine is not available")
	}

	// 查找基因
	gene := qgm.findGeneByID(geneID)
	if gene == nil {
		return nil, fmt.Errorf("gene %s not found", geneID)
	}

	// 检查基因是否可以突�?
	if !gene.CanMutate() {
		return nil, fmt.Errorf("gene %s cannot mutate (mutability: %.2f, stability: %.2f)",
			geneID, gene.Mutability, gene.Stability)
	}

	// 生成突变
	mutation, err := qgm.mutationEngine.GenerateMutation(ctx, gene)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mutation: %w", err)
	}

	// 检查突变严重性是否超过限�?
	if qgm.exceedsSeverityLimit(mutation.Severity) {
		return nil, fmt.Errorf("mutation severity %s exceeds limit %s",
			mutation.Severity, qgm.config.MutationSeverityLimit)
	}

	// 应用突变
	if err := qgm.mutationEngine.ApplyMutation(ctx, gene, mutation); err != nil {
		return nil, fmt.Errorf("failed to apply mutation: %w", err)
	}

	// 更新基因的最后突变时�?
	now := time.Now()
	gene.LastMutation = &now
	gene.UpdatedAt = now

	// 保存突变记录
	if qgm.geneRepository != nil {
		if err := qgm.geneRepository.SaveMutation(ctx, mutation); err != nil {
			qgm.logger.Error("Failed to save mutation to repository", err, "mutation_id", mutation.ID)
		}
		if err := qgm.geneRepository.UpdateGene(ctx, gene); err != nil {
			qgm.logger.Error("Failed to update gene in repository", err, "gene_id", geneID)
		}
	}

	qgm.logger.Info("Gene mutation applied", "gene_id", geneID, "mutation_type", mutation.MutationType, "severity", mutation.Severity)
	return mutation, nil
}

// AnalyzeInteractions 分析基因相互作用
func (qgm *QuantumGeneManager) AnalyzeInteractions(ctx context.Context, poolID string) ([]models.GeneInteraction, error) {
	if qgm.interactionAnalyzer == nil {
		return nil, fmt.Errorf("interaction analyzer is not available")
	}

	qgm.mu.RLock()
	pool, exists := qgm.genePools[poolID]
	qgm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("gene pool %s not found", poolID)
	}

	interactions, err := qgm.interactionAnalyzer.AnalyzeGeneInteractions(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze gene interactions: %w", err)
	}

	// 更新基因池的相互作用信息
	qgm.mu.Lock()
	pool.GeneInteractions = interactions
	pool.UpdatedAt = time.Now()
	qgm.mu.Unlock()

	// 保存更新的基因池
	if qgm.geneRepository != nil {
		if err := qgm.geneRepository.UpdateGenePool(ctx, pool); err != nil {
			qgm.logger.Error("Failed to update gene pool with interactions", err, "pool_id", poolID)
		}
	}

	qgm.logger.Info("Gene interactions analyzed", "pool_id", poolID, "interaction_count", len(interactions))
	return interactions, nil
}

// SimulateEvolution 模拟进化
func (qgm *QuantumGeneManager) SimulateEvolution(ctx context.Context, poolID string, generations int) (*EvolutionSimulationResult, error) {
	if qgm.evolutionSimulator == nil {
		return nil, fmt.Errorf("evolution simulator is not available")
	}

	qgm.mu.RLock()
	pool, exists := qgm.genePools[poolID]
	qgm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("gene pool %s not found", poolID)
	}

	result, err := qgm.evolutionSimulator.SimulateEvolution(ctx, pool, generations)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate evolution: %w", err)
	}

	// 记录进化事件
	qgm.mu.Lock()
	pool.EvolutionHistory = append(pool.EvolutionHistory, result.EvolutionEvents...)
	pool.UpdatedAt = time.Now()
	qgm.mu.Unlock()

	qgm.logger.Info("Evolution simulation completed", "pool_id", poolID, "generations", generations, "events", len(result.EvolutionEvents))
	return result, nil
}

// GetGenePool 获取基因�?
func (qgm *QuantumGeneManager) GetGenePool(poolID string) (*models.GenePool, error) {
	qgm.mu.RLock()
	defer qgm.mu.RUnlock()

	pool, exists := qgm.genePools[poolID]
	if !exists {
		return nil, fmt.Errorf("gene pool %s not found", poolID)
	}

	// 返回基因池副本，防止外部修改
	poolCopy := *pool
	return &poolCopy, nil
}

// GetActiveExpressions 获取活跃的基因表�?
func (qgm *QuantumGeneManager) GetActiveExpressions(entityID string) ([]models.GeneExpression, error) {
	qgm.mu.RLock()
	defer qgm.mu.RUnlock()

	expressions, exists := qgm.geneExpressions[entityID]
	if !exists {
		return []models.GeneExpression{}, nil
	}

	// 过滤活跃的基因表�?
	activeExpressions := []models.GeneExpression{}
	for _, expr := range expressions {
		if expr.IsExpressed() {
			activeExpressions = append(activeExpressions, expr)
		}
	}

	return activeExpressions, nil
}

// IsRunning 检查是否正在运�?
func (qgm *QuantumGeneManager) IsRunning() bool {
	qgm.mu.RLock()
	defer qgm.mu.RUnlock()
	return qgm.isRunning
}

// GetStats 获取统计信息
func (qgm *QuantumGeneManager) GetStats() map[string]interface{} {
	qgm.mu.RLock()
	defer qgm.mu.RUnlock()

	totalGenes := 0
	totalActiveGenes := 0
	totalExpressions := 0

	for _, pool := range qgm.genePools {
		totalGenes += len(pool.Genes)
		totalActiveGenes += pool.GetActiveGeneCount()
	}

	for _, expressions := range qgm.geneExpressions {
		totalExpressions += len(expressions)
	}

	return map[string]interface{}{
		"total_gene_pools":   len(qgm.genePools),
		"total_genes":        totalGenes,
		"total_active_genes": totalActiveGenes,
		"total_expressions":  totalExpressions,
		"is_running":         qgm.isRunning,
	}
}

// 私有方法

func (qgm *QuantumGeneManager) validateDependencies() error {
	// 基本依赖检查，某些组件可以为空
	return nil
}

func (qgm *QuantumGeneManager) runBackgroundProcesses(ctx context.Context) {
	expressionTicker := time.NewTicker(qgm.config.ExpressionUpdateInterval)
	evolutionTicker := time.NewTicker(qgm.config.EvolutionSimulationInterval)
	defer expressionTicker.Stop()
	defer evolutionTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-qgm.stopChan:
			return
		case <-expressionTicker.C:
			qgm.updateExpressions(ctx)
		case <-evolutionTicker.C:
			if qgm.config.EnableEvolutionSimulation {
				qgm.performEvolutionSimulation(ctx)
			}
		}
	}
}

func (qgm *QuantumGeneManager) updateExpressions(ctx context.Context) {
	qgm.mu.Lock()
	defer qgm.mu.Unlock()

	for entityID, expressions := range qgm.geneExpressions {
		activeExpressions := []models.GeneExpression{}
		for _, expr := range expressions {
			if expr.IsExpressed() && expr.GetRemainingDuration() > 0 {
				activeExpressions = append(activeExpressions, expr)
			}
		}
		qgm.geneExpressions[entityID] = activeExpressions
	}
}

func (qgm *QuantumGeneManager) performEvolutionSimulation(ctx context.Context) {
	qgm.mu.RLock()
	poolIDs := make([]string, 0, len(qgm.genePools))
	for poolID := range qgm.genePools {
		poolIDs = append(poolIDs, poolID)
	}
	qgm.mu.RUnlock()

	for _, poolID := range poolIDs {
		if _, err := qgm.SimulateEvolution(ctx, poolID, 1); err != nil {
			qgm.logger.Error("Failed to perform evolution simulation", err, "pool_id", poolID)
		}
	}
}

func (qgm *QuantumGeneManager) generatePoolID() string {
	return fmt.Sprintf("pool_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

func (qgm *QuantumGeneManager) checkGeneCompatibility(pool *models.GenePool, newGene *models.QuantumGene) error {
	for _, existingGene := range pool.Genes {
		if !newGene.IsCompatibleWith(existingGene.ID) {
			return fmt.Errorf("gene %s is not compatible with existing gene %s", newGene.ID, existingGene.ID)
		}
	}
	return nil
}

func (qgm *QuantumGeneManager) updatePoolStats(pool *models.GenePool) {
	pool.PoolStats.TotalGenes = len(pool.Genes)
	pool.PoolStats.ActiveGenes = pool.GetActiveGeneCount()
	pool.PoolStats.DormantGenes = pool.PoolStats.TotalGenes - pool.PoolStats.ActiveGenes
	pool.PoolStats.DiversityIndex = pool.GetDiversityScore()
	pool.PoolStats.LastUpdated = time.Now()

	// 计算平均表达水平
	totalExpression := 0.0
	for _, gene := range pool.Genes {
		totalExpression += gene.Expression
	}
	if len(pool.Genes) > 0 {
		pool.PoolStats.ExpressionLevel = totalExpression / float64(len(pool.Genes))
	}

	// 计算稳定性指�?
	totalStability := 0.0
	for _, gene := range pool.Genes {
		totalStability += gene.Stability
	}
	if len(pool.Genes) > 0 {
		pool.PoolStats.StabilityIndex = totalStability / float64(len(pool.Genes))
	}
}

func (qgm *QuantumGeneManager) findGeneByID(geneID string) *models.QuantumGene {
	for _, pool := range qgm.genePools {
		for i, gene := range pool.Genes {
			if gene.ID == geneID {
				return &pool.Genes[i]
			}
		}
	}
	return nil
}

func (qgm *QuantumGeneManager) exceedsSeverityLimit(severity models.MutationSeverity) bool {
	severityLevels := map[models.MutationSeverity]int{
		models.MutationSeverityMinor:    1,
		models.MutationSeverityModerate: 2,
		models.MutationSeverityMajor:    3,
		models.MutationSeverityCritical: 4,
	}

	currentLevel := severityLevels[severity]
	limitLevel := severityLevels[qgm.config.MutationSeverityLimit]

	return currentLevel > limitLevel
}
