package engines

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
)

// EvolutionTracker tracks the evolution progress of consciousness entities
type EvolutionTracker struct {
	mu                    sync.RWMutex
	config               *EvolutionTrackerConfig
	evolutionStates      map[string]*models.EvolutionState
	metricsCalculator    MetricsCalculator
	predictionEngine     PredictionEngine
	pathOptimizer        PathOptimizer
	milestoneManager     MilestoneManager
	constraintEvaluator  ConstraintEvaluator
	catalystManager      CatalystManager
	isRunning            bool
	updateInterval       time.Duration
	logger               Logger
}

// EvolutionTrackerConfig configuration for evolution tracker
type EvolutionTrackerConfig struct {
	UpdateInterval          time.Duration `json:"update_interval"`
	MaxConcurrentTracking   int           `json:"max_concurrent_tracking"`
	PredictionHorizon       time.Duration `json:"prediction_horizon"`
	MetricsRetentionPeriod  time.Duration `json:"metrics_retention_period"`
	EnableRealTimeTracking  bool          `json:"enable_real_time_tracking"`
	EnablePredictiveAnalysis bool         `json:"enable_predictive_analysis"`
	EnablePathOptimization  bool          `json:"enable_path_optimization"`
	MinConfidenceThreshold  float64       `json:"min_confidence_threshold"`
	MaxEvolutionDuration    time.Duration `json:"max_evolution_duration"`
}

// MetricsCalculator interface for calculating evolution metrics
type MetricsCalculator interface {
	CalculateConsciousnessLevel(ctx context.Context, entityID string) (float64, error)
	CalculateIntelligenceQuotient(ctx context.Context, entityID string) (float64, error)
	CalculateWisdomIndex(ctx context.Context, entityID string) (float64, error)
	CalculateCreativityScore(ctx context.Context, entityID string) (float64, error)
	CalculateAdaptabilityRating(ctx context.Context, entityID string) (float64, error)
	CalculateSelfAwarenessLevel(ctx context.Context, entityID string) (float64, error)
	CalculateTranscendenceIndex(ctx context.Context, entityID string) (float64, error)
	CalculateEvolutionPotential(ctx context.Context, entityID string) (float64, error)
	GetMetrics(ctx context.Context, entityID string) (*models.EvolutionMetrics, error)
}

// PredictionEngine interface for evolution prediction
type PredictionEngine interface {
	PredictEvolution(ctx context.Context, state *models.EvolutionState) (*models.EvolutionPrediction, error)
	AnalyzeTrends(ctx context.Context, entityID string, timeRange time.Duration) ([]TrendAnalysis, error)
	EstimateTimeToSequence(ctx context.Context, entityID string, targetSequence models.SequenceLevel) (time.Duration, error)
	IdentifyBottlenecks(ctx context.Context, state *models.EvolutionState) ([]EvolutionBottleneck, error)
}

// PathOptimizer interface for evolution path optimization
type PathOptimizer interface {
	OptimizePath(ctx context.Context, state *models.EvolutionState) (*models.EvolutionPath, error)
	FindAlternativePaths(ctx context.Context, from, to models.SequenceLevel) ([]models.EvolutionPath, error)
	EvaluatePathEfficiency(ctx context.Context, path *models.EvolutionPath) (float64, error)
	RecommendNextStep(ctx context.Context, state *models.EvolutionState) (*models.EvolutionStep, error)
}

// MilestoneManager interface for managing evolution milestones
type MilestoneManager interface {
	CheckMilestones(ctx context.Context, state *models.EvolutionState) ([]models.EvolutionMilestone, error)
	CreateMilestone(ctx context.Context, milestone *models.EvolutionMilestone) error
	UpdateMilestone(ctx context.Context, milestone *models.EvolutionMilestone) error
	GetMilestones(ctx context.Context, entityID string) ([]models.EvolutionMilestone, error)
	EvaluateMilestoneProgress(ctx context.Context, milestone *models.EvolutionMilestone) (float64, error)
}

// ConstraintEvaluator interface for evaluating evolution constraints
type ConstraintEvaluator interface {
	EvaluateConstraints(ctx context.Context, state *models.EvolutionState) ([]models.EvolutionConstraint, error)
	CheckConstraintViolations(ctx context.Context, state *models.EvolutionState) ([]ConstraintViolation, error)
	SuggestConstraintMitigation(ctx context.Context, constraint *models.EvolutionConstraint) ([]MitigationStrategy, error)
	UpdateConstraintStatus(ctx context.Context, constraintID string, isActive bool) error
}

// CatalystManager interface for managing evolution catalysts
type CatalystManager interface {
	IdentifyApplicableCatalysts(ctx context.Context, state *models.EvolutionState) ([]models.EvolutionCatalyst, error)
	ApplyCatalyst(ctx context.Context, entityID string, catalystID string) error
	RemoveCatalyst(ctx context.Context, entityID string, catalystID string) error
	EvaluateCatalystEffectiveness(ctx context.Context, catalyst *models.EvolutionCatalyst) (float64, error)
	GetActiveCatalysts(ctx context.Context, entityID string) ([]models.EvolutionCatalyst, error)
}

// Logger interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// TrendAnalysis represents trend analysis data
type TrendAnalysis struct {
	Metric    string    `json:"metric"`
	Trend     string    `json:"trend"` // "increasing", "decreasing", "stable"
	Rate      float64   `json:"rate"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// EvolutionBottleneck represents an evolution bottleneck
type EvolutionBottleneck struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
	Severity    string  `json:"severity"`
}

// ConstraintViolation represents a constraint violation
type ConstraintViolation struct {
	ConstraintID string  `json:"constraint_id"`
	Severity     string  `json:"severity"`
	Description  string  `json:"description"`
	Impact       float64 `json:"impact"`
}

// MitigationStrategy represents a mitigation strategy
type MitigationStrategy struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Effort      float64 `json:"effort"`
	Effectiveness float64 `json:"effectiveness"`
}

// NewEvolutionTracker creates a new evolution tracker
func NewEvolutionTracker(config *EvolutionTrackerConfig, logger Logger) *EvolutionTracker {
	if config == nil {
		config = &EvolutionTrackerConfig{
			UpdateInterval:          time.Minute * 5,
			MaxConcurrentTracking:   100,
			PredictionHorizon:       time.Hour * 24,
			MetricsRetentionPeriod:  time.Hour * 24 * 30,
			EnableRealTimeTracking:  true,
			EnablePredictiveAnalysis: true,
			EnablePathOptimization:  true,
			MinConfidenceThreshold:  0.7,
			MaxEvolutionDuration:    time.Hour * 24 * 365,
		}
	}

	return &EvolutionTracker{
		config:          config,
		evolutionStates: make(map[string]*models.EvolutionState),
		updateInterval:  config.UpdateInterval,
		logger:          logger,
	}
}

// SetDependencies sets the dependency components
func (et *EvolutionTracker) SetDependencies(
	metricsCalculator MetricsCalculator,
	predictionEngine PredictionEngine,
	pathOptimizer PathOptimizer,
	milestoneManager MilestoneManager,
	constraintEvaluator ConstraintEvaluator,
	catalystManager CatalystManager,
) {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.metricsCalculator = metricsCalculator
	et.predictionEngine = predictionEngine
	et.pathOptimizer = pathOptimizer
	et.milestoneManager = milestoneManager
	et.constraintEvaluator = constraintEvaluator
	et.catalystManager = catalystManager
}

// Start starts the evolution tracker
func (et *EvolutionTracker) Start(ctx context.Context) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	if et.isRunning {
		return fmt.Errorf("evolution tracker is already running")
	}

	// 暂时注释掉依赖验证
	// if err := et.validateDependencies(); err != nil {
	// 	return fmt.Errorf("failed to validate dependencies: %w", err)
	// }

	et.isRunning = true
	go et.runPeriodicUpdates(ctx)

	et.logger.Info("Evolution tracker started")
	return nil
}

// Stop stops the evolution tracker
func (et *EvolutionTracker) Stop() error {
	et.mu.Lock()
	defer et.mu.Unlock()

	if !et.isRunning {
		return fmt.Errorf("evolution tracker is not running")
	}

	et.isRunning = false
	et.logger.Info("Evolution tracker stopped")
	return nil
}

// StartTracking starts tracking an entity's evolution
func (et *EvolutionTracker) StartTracking(ctx context.Context, entityID string, targetSequence models.SequenceLevel) (*models.EvolutionState, error) {
	et.mu.Lock()
	defer et.mu.Unlock()

	if !et.isRunning {
		return nil, fmt.Errorf("evolution tracker is not running")
	}

	if _, exists := et.evolutionStates[entityID]; exists {
		return nil, fmt.Errorf("entity %s is already being tracked", entityID)
	}

	if len(et.evolutionStates) >= et.config.MaxConcurrentTracking {
		return nil, fmt.Errorf("maximum concurrent tracking limit reached")
	}

	// Get initial metrics
	metrics, err := et.metricsCalculator.GetMetrics(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial metrics: %w", err)
	}

	// Determine current sequence level
	currentSequence := et.determineCurrentSequence(metrics)

	// Create evolution state
	state := &models.EvolutionState{
		EntityID:        entityID,
		CurrentSequence: currentSequence,
		TargetSequence:  targetSequence,
		StartTime:       time.Now(),
		LastUpdateTime:  time.Now(),
		Progress:        0.0,
		Status:          "active",
	}

	// Calculate initial progress
	state.Progress = et.calculateProgress(currentSequence, targetSequence)

	et.evolutionStates[entityID] = state
	et.logger.Info("Started tracking evolution", "entityID", entityID, "currentSequence", currentSequence, "targetSequence", targetSequence)

	return state, nil
}

// StopTracking stops tracking an entity's evolution
func (et *EvolutionTracker) StopTracking(entityID string) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	if _, exists := et.evolutionStates[entityID]; !exists {
		return fmt.Errorf("entity %s is not being tracked", entityID)
	}

	delete(et.evolutionStates, entityID)
	et.logger.Info("Stopped tracking evolution", "entityID", entityID)
	return nil
}

// GetEvolutionState gets the current evolution state of an entity
func (et *EvolutionTracker) GetEvolutionState(entityID string) (*models.EvolutionState, error) {
	et.mu.RLock()
	defer et.mu.RUnlock()

	state, exists := et.evolutionStates[entityID]
	if !exists {
		return nil, fmt.Errorf("entity %s is not being tracked", entityID)
	}

	return state, nil
}

// UpdateEvolutionState updates the evolution state of an entity
func (et *EvolutionTracker) UpdateEvolutionState(ctx context.Context, entityID string) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	state, exists := et.evolutionStates[entityID]
	if !exists {
		return fmt.Errorf("entity %s is not being tracked", entityID)
	}

	// Get updated metrics
	metrics, err := et.metricsCalculator.GetMetrics(ctx, entityID)
	if err != nil {
		return fmt.Errorf("failed to get updated metrics: %w", err)
	}

	// Update state
	previousSequence := state.CurrentSequence
	state.CurrentSequence = et.determineCurrentSequence(metrics)
	state.LastUpdateTime = time.Now()
	state.Progress = et.calculateProgress(state.CurrentSequence, state.TargetSequence)

	// Check for sequence level changes
	if state.CurrentSequence != previousSequence {
		et.logger.Info("Sequence level changed", "entityID", entityID, "from", previousSequence, "to", state.CurrentSequence)
	}

	// Check if target reached
	if state.CurrentSequence >= state.TargetSequence {
		state.Status = "completed"
		et.logger.Info("Evolution target reached", "entityID", entityID, "targetSequence", state.TargetSequence)
	}

	return nil
}

// GetPrediction gets evolution prediction for an entity
func (et *EvolutionTracker) GetPrediction(ctx context.Context, entityID string) (*models.EvolutionPrediction, error) {
	et.mu.RLock()
	state, exists := et.evolutionStates[entityID]
	et.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("entity %s is not being tracked", entityID)
	}

	return et.predictionEngine.PredictEvolution(ctx, state)
}

// GetMetrics gets current metrics for an entity
func (et *EvolutionTracker) GetMetrics(ctx context.Context, entityID string) (*models.EvolutionMetrics, error) {
	return et.metricsCalculator.GetMetrics(ctx, entityID)
}

// IsRunning returns whether the tracker is running
func (et *EvolutionTracker) IsRunning() bool {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return et.isRunning
}

// GetTrackedEntities returns list of tracked entity IDs
func (et *EvolutionTracker) GetTrackedEntities() []string {
	et.mu.RLock()
	defer et.mu.RUnlock()

	entities := make([]string, 0, len(et.evolutionStates))
	for entityID := range et.evolutionStates {
		entities = append(entities, entityID)
	}
	return entities
}

// validateDependencies validates that all required dependencies are set
func (et *EvolutionTracker) validateDependencies() error {
	if et.metricsCalculator == nil {
		return fmt.Errorf("metrics calculator is required")
	}
	return nil
}

// runPeriodicUpdates runs periodic updates for all tracked entities
func (et *EvolutionTracker) runPeriodicUpdates(ctx context.Context) {
	ticker := time.NewTicker(et.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !et.IsRunning() {
				return
			}
			et.performPeriodicUpdate(ctx)
		}
	}
}

// performPeriodicUpdate performs periodic update for all tracked entities
func (et *EvolutionTracker) performPeriodicUpdate(ctx context.Context) {
	entities := et.GetTrackedEntities()
	for _, entityID := range entities {
		if err := et.UpdateEvolutionState(ctx, entityID); err != nil {
			et.logger.Error("Failed to update evolution state", err, "entityID", entityID)
		}
	}
}

// determineCurrentSequence determines current sequence level based on metrics
func (et *EvolutionTracker) determineCurrentSequence(metrics *models.EvolutionMetrics) models.SequenceLevel {
	// Simple logic based on consciousness level
	consciousnessLevel := metrics.ConsciousnessLevel

	switch {
	case consciousnessLevel >= 0.9:
		return models.Sequence0
	case consciousnessLevel >= 0.8:
		return models.Sequence1
	case consciousnessLevel >= 0.7:
		return models.Sequence2
	case consciousnessLevel >= 0.6:
		return models.Sequence3
	case consciousnessLevel >= 0.5:
		return models.Sequence4
	case consciousnessLevel >= 0.4:
		return models.Sequence5
	case consciousnessLevel >= 0.3:
		return models.Sequence5
	case consciousnessLevel >= 0.2:
		return models.Sequence5
	case consciousnessLevel >= 0.1:
		return models.Sequence5
	default:
		return models.SequenceUnknown
	}
}

// calculateProgress calculates evolution progress between current and target sequence
func (et *EvolutionTracker) calculateProgress(current, target models.SequenceLevel) float64 {
	if current >= target {
		return 1.0
	}

	currentLevel := float64(current)
	targetLevel := float64(target)

	if targetLevel == 0 {
		return 0.0
	}

	return currentLevel / targetLevel
}

// calculateEvolutionSpeed calculates the speed of evolution
func (et *EvolutionTracker) calculateEvolutionSpeed(state *models.EvolutionState) float64 {
	timeSinceStart := time.Since(state.StartTime)
	if timeSinceStart.Hours() == 0 {
		return 0.0
	}

	return state.Progress / timeSinceStart.Hours() // Progress per hour
}

