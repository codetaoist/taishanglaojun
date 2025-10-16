package evolution

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// EvolutionStrategy 
type EvolutionStrategy string

const (
	StrategyGenetic        EvolutionStrategy = "genetic"
	StrategyNeuroEvolution EvolutionStrategy = "neuro_evolution"
	StrategyGradientFree   EvolutionStrategy = "gradient_free"
	StrategyHybrid         EvolutionStrategy = "hybrid"
	StrategyReinforcement  EvolutionStrategy = "reinforcement"
	StrategySwarmIntel     EvolutionStrategy = "swarm_intelligence"
)

// Individual 
type Individual struct {
	ID         string                 `json:"id"`
	Genome     map[string]interface{} `json:"genome"`
	Fitness    float64                `json:"fitness"`
	Age        int                    `json:"age"`
	Generation int                    `json:"generation"`
	Parents    []string               `json:"parents"`
	Mutations  []string               `json:"mutations"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Population 
type Population struct {
	ID          string       `json:"id"`
	Individuals []Individual `json:"individuals"`
	Generation  int          `json:"generation"`
	Size        int          `json:"size"`
	BestFitness float64      `json:"best_fitness"`
	AvgFitness  float64      `json:"avg_fitness"`
	Diversity   float64      `json:"diversity"`
	CreatedAt   time.Time    `json:"created_at"`
}

// EvolutionConfig 
type EvolutionConfig struct {
	Strategy         EvolutionStrategy `json:"strategy"`
	PopulationSize   int               `json:"population_size"`
	Generations      int               `json:"generations"`
	MutationRate     float64           `json:"mutation_rate"`
	CrossoverRate    float64           `json:"crossover_rate"`
	SelectionMethod  string            `json:"selection_method"`
	ElitismRate      float64           `json:"elitism_rate"`
	DiversityWeight  float64           `json:"diversity_weight"`
	FitnessThreshold float64           `json:"fitness_threshold"`
	MaxAge           int               `json:"max_age"`
}

// EvolutionResult 
type EvolutionResult struct {
	BestIndividual  Individual             `json:"best_individual"`
	FinalPopulation Population             `json:"final_population"`
	Generations     int                    `json:"generations"`
	Convergence     bool                   `json:"convergence"`
	EvolutionTime   time.Duration          `json:"evolution_time"`
	Improvements    []float64              `json:"improvements"`
	Statistics      map[string]interface{} `json:"statistics"`
	CreatedAt       time.Time              `json:"created_at"`
}

// PerformanceMetrics 
type PerformanceMetrics struct {
	Accuracy      float64            `json:"accuracy"`
	Efficiency    float64            `json:"efficiency"`
	Robustness    float64            `json:"robustness"`
	Adaptability  float64            `json:"adaptability"`
	ResourceUsage float64            `json:"resource_usage"`
	Latency       time.Duration      `json:"latency"`
	Throughput    float64            `json:"throughput"`
	ErrorRate     float64            `json:"error_rate"`
	CustomMetrics map[string]float64 `json:"custom_metrics"`
	Timestamp     time.Time          `json:"timestamp"`
}

// OptimizationTarget 
type OptimizationTarget struct {
	Name      string  `json:"name"`
	Weight    float64 `json:"weight"`
	Target    float64 `json:"target"`
	Tolerance float64 `json:"tolerance"`
	Maximize  bool    `json:"maximize"`
	Priority  int     `json:"priority"`
}

// SelfEvolutionSystem 
type SelfEvolutionSystem struct {
	mu                  sync.RWMutex
	config              *EvolutionConfig
	currentPopulation   *Population
	bestIndividual      *Individual
	performanceHist     []PerformanceMetrics
	optimizationTargets []OptimizationTarget
	evolutionHistory    []EvolutionResult
	isRunning           bool
	stopChan            chan struct{}
	strategies          map[EvolutionStrategy]EvolutionStrategyImpl
}

// EvolutionStrategyImpl 
type EvolutionStrategyImpl interface {
	Initialize(config *EvolutionConfig) error
	Evolve(ctx context.Context, population *Population, metrics *PerformanceMetrics) (*Population, error)
	Mutate(individual *Individual, rate float64) *Individual
	Crossover(parent1, parent2 *Individual) (*Individual, *Individual)
	Select(population *Population, count int) []Individual
	Evaluate(individual *Individual, metrics *PerformanceMetrics) float64
}

// NewSelfEvolutionSystem 
func NewSelfEvolutionSystem(config *EvolutionConfig) *SelfEvolutionSystem {
	system := &SelfEvolutionSystem{
		config:              config,
		performanceHist:     make([]PerformanceMetrics, 0),
		optimizationTargets: make([]OptimizationTarget, 0),
		evolutionHistory:    make([]EvolutionResult, 0),
		stopChan:            make(chan struct{}),
		strategies:          make(map[EvolutionStrategy]EvolutionStrategyImpl),
	}

	// 
	system.registerStrategies()

	return system
}

// StartEvolution 
func (ses *SelfEvolutionSystem) StartEvolution(ctx context.Context) error {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	if ses.isRunning {
		return fmt.Errorf("evolution system is already running")
	}

	ses.isRunning = true

	// 
	if ses.currentPopulation == nil {
		ses.currentPopulation = ses.initializePopulation()
	}

	// 
	go ses.evolutionLoop(ctx)

	return nil
}

// StopEvolution 
func (ses *SelfEvolutionSystem) StopEvolution() error {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	if !ses.isRunning {
		return fmt.Errorf("evolution system is not running")
	}

	close(ses.stopChan)
	ses.isRunning = false

	return nil
}

// UpdatePerformanceMetrics 
func (ses *SelfEvolutionSystem) UpdatePerformanceMetrics(metrics *PerformanceMetrics) {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	metrics.Timestamp = time.Now()
	ses.performanceHist = append(ses.performanceHist, *metrics)

	// 
	if len(ses.performanceHist) > 1000 {
		ses.performanceHist = ses.performanceHist[len(ses.performanceHist)-1000:]
	}
}

// SetOptimizationTargets 
func (ses *SelfEvolutionSystem) SetOptimizationTargets(targets []OptimizationTarget) {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	ses.optimizationTargets = targets
}

// GetBestIndividual 
func (ses *SelfEvolutionSystem) GetBestIndividual() *Individual {
	ses.mu.RLock()
	defer ses.mu.RUnlock()

	if ses.bestIndividual == nil {
		return nil
	}

	// 
	best := *ses.bestIndividual
	return &best
}

// GetEvolutionStatus 
func (ses *SelfEvolutionSystem) GetEvolutionStatus() map[string]interface{} {
	ses.mu.RLock()
	defer ses.mu.RUnlock()

	status := map[string]interface{}{
		"is_running":         ses.isRunning,
		"current_generation": 0,
		"population_size":    0,
		"best_fitness":       0.0,
		"avg_fitness":        0.0,
		"diversity":          0.0,
	}

	if ses.currentPopulation != nil {
		status["current_generation"] = ses.currentPopulation.Generation
		status["population_size"] = len(ses.currentPopulation.Individuals)
		status["best_fitness"] = ses.currentPopulation.BestFitness
		status["avg_fitness"] = ses.currentPopulation.AvgFitness
		status["diversity"] = ses.currentPopulation.Diversity
	}

	return status
}

// GetPerformanceHistory 
func (ses *SelfEvolutionSystem) GetPerformanceHistory(limit int) []PerformanceMetrics {
	ses.mu.RLock()
	defer ses.mu.RUnlock()

	if limit <= 0 || limit > len(ses.performanceHist) {
		limit = len(ses.performanceHist)
	}

	start := len(ses.performanceHist) - limit
	history := make([]PerformanceMetrics, limit)
	copy(history, ses.performanceHist[start:])

	return history
}

// OptimizeConfiguration 
func (ses *SelfEvolutionSystem) OptimizeConfiguration(ctx context.Context) (*EvolutionConfig, error) {
	ses.mu.RLock()
	currentConfig := *ses.config
	ses.mu.RUnlock()

	// 
	optimizedConfig := ses.optimizeConfigBasedOnHistory(&currentConfig)

	return optimizedConfig, nil
}

// 

func (ses *SelfEvolutionSystem) registerStrategies() {
	ses.strategies[StrategyGenetic] = NewGeneticStrategy()
	ses.strategies[StrategyNeuroEvolution] = NewNeuroEvolutionStrategy()
	ses.strategies[StrategyGradientFree] = NewGradientFreeStrategy()
	ses.strategies[StrategyHybrid] = NewHybridStrategy()
	ses.strategies[StrategyReinforcement] = NewReinforcementStrategy()
	ses.strategies[StrategySwarmIntel] = NewSwarmIntelligenceStrategy()
}

func (ses *SelfEvolutionSystem) evolutionLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5) // 5
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ses.stopChan:
			return
		case <-ticker.C:
			ses.performEvolutionStep(ctx)
		}
	}
}

func (ses *SelfEvolutionSystem) performEvolutionStep(ctx context.Context) {
	ses.mu.Lock()
	defer ses.mu.Unlock()

	if len(ses.performanceHist) == 0 {
		return
	}

	// 
	latestMetrics := ses.performanceHist[len(ses.performanceHist)-1]

	// 
	strategy := ses.strategies[ses.config.Strategy]
	if strategy == nil {
		return
	}

	// 
	newPopulation, err := strategy.Evolve(ctx, ses.currentPopulation, &latestMetrics)
	if err != nil {
		return
	}

	// 
	ses.currentPopulation = newPopulation
	ses.updateBestIndividual()
	ses.updatePopulationStatistics()
}

func (ses *SelfEvolutionSystem) initializePopulation() *Population {
	individuals := make([]Individual, ses.config.PopulationSize)

	for i := 0; i < ses.config.PopulationSize; i++ {
		individuals[i] = Individual{
			ID:         fmt.Sprintf("ind_%d_%d", 0, i),
			Genome:     ses.generateRandomGenome(),
			Fitness:    0.0,
			Age:        0,
			Generation: 0,
			Parents:    []string{},
			Mutations:  []string{},
			CreatedAt:  time.Now(),
		}
	}

	population := &Population{
		ID:          fmt.Sprintf("pop_%d", 0),
		Individuals: individuals,
		Generation:  0,
		Size:        ses.config.PopulationSize,
		CreatedAt:   time.Now(),
	}

	return population
}

func (ses *SelfEvolutionSystem) generateRandomGenome() map[string]interface{} {
	genome := map[string]interface{}{
		"learning_rate":  rand.Float64() * 0.1,
		"batch_size":     rand.Intn(64) + 16,
		"hidden_layers":  rand.Intn(5) + 1,
		"dropout_rate":   rand.Float64() * 0.5,
		"activation":     []string{"relu", "tanh", "sigmoid"}[rand.Intn(3)],
		"optimizer":      []string{"adam", "sgd", "rmsprop"}[rand.Intn(3)],
		"regularization": rand.Float64() * 0.01,
		"architecture":   ses.generateRandomArchitecture(),
	}

	return genome
}

func (ses *SelfEvolutionSystem) generateRandomArchitecture() []int {
	layers := rand.Intn(5) + 2 // 2-6
	architecture := make([]int, layers)

	architecture[0] = 128 // 128
	for i := 1; i < layers-1; i++ {
		architecture[i] = rand.Intn(256) + 32 // 32-288
	}
	architecture[layers-1] = 10 // 10
	return architecture
}

func (ses *SelfEvolutionSystem) updateBestIndividual() {
	if len(ses.currentPopulation.Individuals) == 0 {
		return
	}

	bestFitness := -math.Inf(1)
	var bestIdx int

	for i, individual := range ses.currentPopulation.Individuals {
		if individual.Fitness > bestFitness {
			bestFitness = individual.Fitness
			bestIdx = i
		}
	}

	ses.bestIndividual = &ses.currentPopulation.Individuals[bestIdx]
}

func (ses *SelfEvolutionSystem) updatePopulationStatistics() {
	if len(ses.currentPopulation.Individuals) == 0 {
		return
	}

	// 
	bestFitness := -math.Inf(1)
	totalFitness := 0.0

	for _, individual := range ses.currentPopulation.Individuals {
		if individual.Fitness > bestFitness {
			bestFitness = individual.Fitness
		}
		totalFitness += individual.Fitness
	}

	ses.currentPopulation.BestFitness = bestFitness
	ses.currentPopulation.AvgFitness = totalFitness / float64(len(ses.currentPopulation.Individuals))
	ses.currentPopulation.Diversity = ses.calculateDiversity()
}

func (ses *SelfEvolutionSystem) calculateDiversity() float64 {
	if len(ses.currentPopulation.Individuals) < 2 {
		return 0.0
	}

	totalDistance := 0.0
	comparisons := 0

	for i := 0; i < len(ses.currentPopulation.Individuals); i++ {
		for j := i + 1; j < len(ses.currentPopulation.Individuals); j++ {
			distance := ses.calculateGenomeDistance(
				ses.currentPopulation.Individuals[i].Genome,
				ses.currentPopulation.Individuals[j].Genome,
			)
			totalDistance += distance
			comparisons++
		}
	}

	if comparisons == 0 {
		return 0.0
	}

	return totalDistance / float64(comparisons)
}

func (ses *SelfEvolutionSystem) calculateGenomeDistance(genome1, genome2 map[string]interface{}) float64 {
	distance := 0.0
	count := 0

	for key, value1 := range genome1 {
		if value2, exists := genome2[key]; exists {
			switch v1 := value1.(type) {
			case float64:
				if v2, ok := value2.(float64); ok {
					distance += math.Abs(v1 - v2)
					count++
				}
			case int:
				if v2, ok := value2.(int); ok {
					distance += math.Abs(float64(v1 - v2))
					count++
				}
			case string:
				if v2, ok := value2.(string); ok {
					if v1 != v2 {
						distance += 1.0
					}
					count++
				}
			}
		}
	}

	if count == 0 {
		return 0.0
	}

	return distance / float64(count)
}

func (ses *SelfEvolutionSystem) optimizeConfigBasedOnHistory(config *EvolutionConfig) *EvolutionConfig {
	optimized := *config

	if len(ses.performanceHist) < 10 {
		return &optimized
	}

	// 
	recentPerf := ses.performanceHist[len(ses.performanceHist)-10:]
	avgAccuracy := 0.0
	avgEfficiency := 0.0

	for _, perf := range recentPerf {
		avgAccuracy += perf.Accuracy
		avgEfficiency += perf.Efficiency
	}

	avgAccuracy /= float64(len(recentPerf))
	avgEfficiency /= float64(len(recentPerf))

	// 
	if avgAccuracy < 0.7 {
		optimized.MutationRate *= 1.2 // 
		optimized.PopulationSize = int(float64(optimized.PopulationSize) * 1.1)
	} else if avgAccuracy > 0.9 {
		optimized.MutationRate *= 0.8 // 
		optimized.PopulationSize = int(float64(optimized.PopulationSize) * 0.9)
	}

	if avgEfficiency < 0.6 {
		optimized.ElitismRate *= 1.1 // 
	}

	return &optimized
}

// CalculateFitness 
func (ses *SelfEvolutionSystem) CalculateFitness(individual *Individual, metrics *PerformanceMetrics) float64 {
	if len(ses.optimizationTargets) == 0 {
		// 40%30%20%10%
		return metrics.Accuracy*0.4 + metrics.Efficiency*0.3 + metrics.Robustness*0.2 + metrics.Adaptability*0.1
	}

	totalFitness := 0.0
	totalWeight := 0.0

	for _, target := range ses.optimizationTargets {
		value := ses.getMetricValue(metrics, target.Name)

		// 
		achievement := 0.0
		if target.Maximize {
			achievement = math.Min(value/target.Target, 1.0)
		} else {
			if value <= target.Target {
				achievement = 1.0
			} else {
				achievement = math.Max(0.0, 1.0-(value-target.Target)/target.Target)
			}
		}

		totalFitness += achievement * target.Weight
		totalWeight += target.Weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalFitness / totalWeight
}

func (ses *SelfEvolutionSystem) getMetricValue(metrics *PerformanceMetrics, name string) float64 {
	switch name {
	case "accuracy":
		return metrics.Accuracy
	case "efficiency":
		return metrics.Efficiency
	case "robustness":
		return metrics.Robustness
	case "adaptability":
		return metrics.Adaptability
	case "resource_usage":
		return metrics.ResourceUsage
	case "latency":
		return metrics.Latency.Seconds()
	case "throughput":
		return metrics.Throughput
	case "error_rate":
		return metrics.ErrorRate
	default:
		if value, exists := metrics.CustomMetrics[name]; exists {
			return value
		}
		return 0.0
	}
}

