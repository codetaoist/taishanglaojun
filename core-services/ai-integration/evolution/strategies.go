package evolution

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// GeneticStrategy 遗传算法策略
type GeneticStrategy struct {
	config *EvolutionConfig
}

// NewGeneticStrategy 创建新的遗传算法策略
func NewGeneticStrategy() *GeneticStrategy {
	return &GeneticStrategy{}
}

// Initialize 初始化遗传算法策略
func (gs *GeneticStrategy) Initialize(config *EvolutionConfig) error {
	gs.config = config
	return nil
}

// Evolve 进化种群
func (gs *GeneticStrategy) Evolve(ctx context.Context, population *Population, metrics *PerformanceMetrics) (*Population, error) {
	//
	for i := range population.Individuals {
		population.Individuals[i].Fitness = gs.Evaluate(&population.Individuals[i], metrics)
	}

	//
	parents := gs.Select(population, int(float64(population.Size)*gs.config.CrossoverRate))

	//
	newIndividuals := make([]Individual, 0, population.Size)

	//
	eliteCount := int(float64(population.Size) * gs.config.ElitismRate)
	elite := gs.selectElite(population, eliteCount)
	newIndividuals = append(newIndividuals, elite...)

	//
	for len(newIndividuals) < population.Size {
		if len(parents) >= 2 {
			parent1 := parents[rand.Intn(len(parents))]
			parent2 := parents[rand.Intn(len(parents))]

			child1, child2 := gs.Crossover(&parent1, &parent2)

			if rand.Float64() < gs.config.MutationRate {
				child1 = gs.Mutate(child1, gs.config.MutationRate)
			}
			if rand.Float64() < gs.config.MutationRate {
				child2 = gs.Mutate(child2, gs.config.MutationRate)
			}

			child1.Generation = population.Generation + 1
			child2.Generation = population.Generation + 1
			child1.ID = fmt.Sprintf("ind_%d_%d", child1.Generation, len(newIndividuals))
			child2.ID = fmt.Sprintf("ind_%d_%d", child2.Generation, len(newIndividuals)+1)

			newIndividuals = append(newIndividuals, *child1)
			if len(newIndividuals) < population.Size {
				newIndividuals = append(newIndividuals, *child2)
			}
		} else {
			break
		}
	}

	newPopulation := &Population{
		ID:          fmt.Sprintf("pop_%d", population.Generation+1),
		Individuals: newIndividuals,
		Generation:  population.Generation + 1,
		Size:        len(newIndividuals),
		CreatedAt:   time.Now(),
	}

	return newPopulation, nil
}

// Mutate 变异个体
func (gs *GeneticStrategy) Mutate(individual *Individual, rate float64) *Individual {
	mutated := *individual
	mutated.Genome = make(map[string]interface{})

	// 复制基因组
	for key, value := range individual.Genome {
		mutated.Genome[key] = value
	}

	// 变异基因组
	for key, value := range mutated.Genome {
		if rand.Float64() < rate {
			switch v := value.(type) {
			case float64:
				// 浮点数变异
				noise := rand.NormFloat64() * 0.1
				mutated.Genome[key] = math.Max(0, v+noise)
			case int:
				// 整数变异
				delta := rand.Intn(21) - 10 // [-10, 10]
				mutated.Genome[key] = int(math.Max(1, float64(v+delta)))
			case string:
				// 字符串变异
				options := gs.getStringOptions(key)
				if len(options) > 0 {
					mutated.Genome[key] = options[rand.Intn(len(options))]
				}
			case []int:
				// 数组变异
				arr := make([]int, len(v))
				copy(arr, v)
				if len(arr) > 0 {
					idx := rand.Intn(len(arr))
					arr[idx] = rand.Intn(512) + 16
				}
				mutated.Genome[key] = arr
			}

			mutated.Mutations = append(mutated.Mutations, fmt.Sprintf("mutated_%s", key))
		}
	}

	return &mutated
}

// Crossover 交叉个体
func (gs *GeneticStrategy) Crossover(parent1, parent2 *Individual) (*Individual, *Individual) {
	child1 := Individual{
		Genome:    make(map[string]interface{}),
		Age:       0,
		Parents:   []string{parent1.ID, parent2.ID},
		Mutations: []string{},
		CreatedAt: time.Now(),
	}

	child2 := Individual{
		Genome:    make(map[string]interface{}),
		Age:       0,
		Parents:   []string{parent1.ID, parent2.ID},
		Mutations: []string{},
		CreatedAt: time.Now(),
	}

	// 基因组交叉
	for key := range parent1.Genome {
		if rand.Float64() < 0.5 {
			child1.Genome[key] = parent1.Genome[key]
			child2.Genome[key] = parent2.Genome[key]
		} else {
			child1.Genome[key] = parent2.Genome[key]
			child2.Genome[key] = parent1.Genome[key]
		}
	}

	return &child1, &child2
}

// Select 选择个体
func (gs *GeneticStrategy) Select(population *Population, count int) []Individual {
	// 锦标赛选择
	selected := make([]Individual, 0, count)
	tournamentSize := 3

	for i := 0; i < count; i++ {
		tournament := make([]Individual, 0, tournamentSize)

		for j := 0; j < tournamentSize && j < len(population.Individuals); j++ {
			idx := rand.Intn(len(population.Individuals))
			tournament = append(tournament, population.Individuals[idx])
		}

		// 选择锦标赛中的最优个体
		best := tournament[0]
		for _, individual := range tournament[1:] {
			if individual.Fitness > best.Fitness {
				best = individual
			}
		}

		selected = append(selected, best)
	}

	return selected
}

// Evaluate 评估个体
func (gs *GeneticStrategy) Evaluate(individual *Individual, metrics *PerformanceMetrics) float64 {
	// 基础适应度
	baseFitness := metrics.Accuracy*0.4 + metrics.Efficiency*0.3 + metrics.Robustness*0.2 + metrics.Adaptability*0.1

	// 基因组质量
	genomeQuality := gs.evaluateGenomeQuality(individual.Genome)

	return baseFitness*0.7 + genomeQuality*0.3
}

// selectElite 选择精英个体
func (gs *GeneticStrategy) selectElite(population *Population, count int) []Individual {
	if count <= 0 {
		return []Individual{}
	}

	// 复制个体
	individuals := make([]Individual, len(population.Individuals))
	copy(individuals, population.Individuals)

	// 按适应度排序
	sort.Slice(individuals, func(i, j int) bool {
		return individuals[i].Fitness > individuals[j].Fitness
	})

	if count > len(individuals) {
		count = len(individuals)
	}

	return individuals[:count]
}

// getStringOptions 获取字符串选项
func (gs *GeneticStrategy) getStringOptions(key string) []string {
	switch key {
	case "activation":
		return []string{"relu", "tanh", "sigmoid", "leaky_relu", "elu"}
	case "optimizer":
		return []string{"adam", "sgd", "rmsprop", "adagrad", "adamw"}
	default:
		return []string{}
	}
}

// evaluateGenomeQuality 评估基因组质量
func (gs *GeneticStrategy) evaluateGenomeQuality(genome map[string]interface{}) float64 {
	quality := 0.0

	// 学习率
	if lr, ok := genome["learning_rate"].(float64); ok {
		if lr > 0.001 && lr < 0.1 {
			quality += 0.2
		}
	}

	// 批量大小
	if bs, ok := genome["batch_size"].(int); ok {
		if bs >= 16 && bs <= 128 {
			quality += 0.2
		}
	}

	// 隐藏层数量
	if hl, ok := genome["hidden_layers"].(int); ok {
		if hl >= 2 && hl <= 5 {
			quality += 0.2
		}
	}

	// Dropout 率
	if dr, ok := genome["dropout_rate"].(float64); ok {
		if dr >= 0.1 && dr <= 0.5 {
			quality += 0.2
		}
	}

	// 架构
	if arch, ok := genome["architecture"].([]int); ok {
		if len(arch) >= 3 && len(arch) <= 6 {
			quality += 0.2
		}
	}

	return quality
}

// NeuroEvolutionStrategy 神经进化策略
type NeuroEvolutionStrategy struct {
	config *EvolutionConfig
}

// NewNeuroEvolutionStrategy 创建神经进化策略
func NewNeuroEvolutionStrategy() *NeuroEvolutionStrategy {
	return &NeuroEvolutionStrategy{}
}

// Initialize 初始化神经进化策略
func (nes *NeuroEvolutionStrategy) Initialize(config *EvolutionConfig) error {
	nes.config = config
	return nil
}

// Evolve 进化种群
func (nes *NeuroEvolutionStrategy) Evolve(ctx context.Context, population *Population, metrics *PerformanceMetrics) (*Population, error) {
	// 评估个体适应度
	for i := range population.Individuals {
		population.Individuals[i].Fitness = nes.Evaluate(&population.Individuals[i], metrics)
	}

	// NEAT㷨
	newIndividuals := nes.evolveWithNEAT(population)

	newPopulation := &Population{
		ID:          fmt.Sprintf("pop_%d", population.Generation+1),
		Individuals: newIndividuals,
		Generation:  population.Generation + 1,
		Size:        len(newIndividuals),
		CreatedAt:   time.Now(),
	}

	return newPopulation, nil
}

// Mutate 变异个体
func (nes *NeuroEvolutionStrategy) Mutate(individual *Individual, rate float64) *Individual {
	mutated := *individual
	mutated.Genome = make(map[string]interface{})

	//
	for key, value := range individual.Genome {
		mutated.Genome[key] = value
	}

	//
	if rand.Float64() < rate {
		//
		if weights, ok := mutated.Genome["weights"].([]float64); ok {
			for i := range weights {
				if rand.Float64() < 0.1 {
					weights[i] += rand.NormFloat64() * 0.1
				}
			}
			mutated.Genome["weights"] = weights
		}

		//
		if rand.Float64() < 0.05 {
			nes.mutateStructure(&mutated)
		}
	}

	return &mutated
}

// Crossover 交叉个体
func (nes *NeuroEvolutionStrategy) Crossover(parent1, parent2 *Individual) (*Individual, *Individual) {
	// 基因交叉
	child1 := *parent1
	child2 := *parent2

	child1.Genome = make(map[string]interface{})
	child2.Genome = make(map[string]interface{})

	// 基因交叉
	for key, value := range parent1.Genome {
		child1.Genome[key] = value
	}
	for key, value := range parent2.Genome {
		child2.Genome[key] = value
	}

	// 权重交叉
	if weights1, ok1 := parent1.Genome["weights"].([]float64); ok1 {
		if weights2, ok2 := parent2.Genome["weights"].([]float64); ok2 {
			newWeights1, newWeights2 := nes.crossoverWeights(weights1, weights2)
			child1.Genome["weights"] = newWeights1
			child2.Genome["weights"] = newWeights2
		}
	}

	child1.Parents = []string{parent1.ID, parent2.ID}
	child2.Parents = []string{parent1.ID, parent2.ID}
	child1.CreatedAt = time.Now()
	child2.CreatedAt = time.Now()

	return &child1, &child2
}

// Select 选择个体
func (nes *NeuroEvolutionStrategy) Select(population *Population, count int) []Individual {
	// 按适应度排序
	individuals := make([]Individual, len(population.Individuals))
	copy(individuals, population.Individuals)

	sort.Slice(individuals, func(i, j int) bool {
		return individuals[i].Fitness > individuals[j].Fitness
	})

	if count > len(individuals) {
		count = len(individuals)
	}

	return individuals[:count]
}

// Evaluate 评估个体适应度
func (nes *NeuroEvolutionStrategy) Evaluate(individual *Individual, metrics *PerformanceMetrics) float64 {
	// 基础适应度
	baseFitness := metrics.Accuracy*0.5 + metrics.Efficiency*0.3 + metrics.Robustness*0.2

	// 复杂度惩罚
	complexity := nes.calculateComplexity(individual.Genome)
	complexityPenalty := complexity * 0.1

	return math.Max(0, baseFitness-complexityPenalty)
}

// getBestIndividuals 获取最佳个体
func (nes *NeuroEvolutionStrategy) getBestIndividuals(population *Population, count int) []Individual {
	// 按适应度排序
	individuals := make([]Individual, len(population.Individuals))
	copy(individuals, population.Individuals)

	sort.Slice(individuals, func(i, j int) bool {
		return individuals[i].Fitness > individuals[j].Fitness
	})

	if count > len(individuals) {
		count = len(individuals)
	}

	return individuals[:count]
}

// evolveWithNEAT NEAT㷨
func (nes *NeuroEvolutionStrategy) evolveWithNEAT(population *Population) []Individual {
	// NEAT
	newIndividuals := make([]Individual, 0, population.Size)

	// 选择最佳个体
	best := nes.getBestIndividuals(population, population.Size/4)
	newIndividuals = append(newIndividuals, best...)

	// 交叉变异
	for len(newIndividuals) < population.Size {
		if len(best) >= 2 {
			parent1 := best[rand.Intn(len(best))]
			parent2 := best[rand.Intn(len(best))]

			child1, child2 := nes.Crossover(&parent1, &parent2)

			if rand.Float64() < nes.config.MutationRate {
				child1 = nes.Mutate(child1, nes.config.MutationRate)
			}
			if rand.Float64() < nes.config.MutationRate {
				child2 = nes.Mutate(child2, nes.config.MutationRate)
			}

			child1.Generation = population.Generation + 1
			child2.Generation = population.Generation + 1
			child1.ID = fmt.Sprintf("neuro_ind_%d_%d", child1.Generation, len(newIndividuals))
			child2.ID = fmt.Sprintf("neuro_ind_%d_%d", child2.Generation, len(newIndividuals)+1)

			newIndividuals = append(newIndividuals, *child1)
			if len(newIndividuals) < population.Size {
				newIndividuals = append(newIndividuals, *child2)
			}
		} else {
			break
		}
	}

	return newIndividuals
}

// mutateStructure 结构变异
func (nes *NeuroEvolutionStrategy) mutateStructure(individual *Individual) {
	// 结构变异
	if arch, ok := individual.Genome["architecture"].([]int); ok {
		newArch := make([]int, len(arch))
		copy(newArch, arch)

		if rand.Float64() < 0.5 && len(newArch) > 2 {
			//
			idx := rand.Intn(len(newArch)-1) + 1
			newSize := rand.Intn(256) + 32
			newArch = append(newArch[:idx], append([]int{newSize}, newArch[idx:]...)...)
		} else if len(newArch) > 3 {
			//
			idx := rand.Intn(len(newArch)-2) + 1
			newArch = append(newArch[:idx], newArch[idx+1:]...)
		}

		individual.Genome["architecture"] = newArch
		individual.Mutations = append(individual.Mutations, "structure_mutation")
	}
}

// crossoverWeights 权重交叉
func (nes *NeuroEvolutionStrategy) crossoverWeights(weights1, weights2 []float64) ([]float64, []float64) {
	minLen := len(weights1)
	if len(weights2) < minLen {
		minLen = len(weights2)
	}

	newWeights1 := make([]float64, len(weights1))
	newWeights2 := make([]float64, len(weights2))

	copy(newWeights1, weights1)
	copy(newWeights2, weights2)

	// 权重交叉
	crossoverPoint := rand.Intn(minLen)

	for i := crossoverPoint; i < minLen; i++ {
		newWeights1[i] = weights2[i]
		newWeights2[i] = weights1[i]
	}

	return newWeights1, newWeights2
}

// calculateComplexity 计算复杂度
func (nes *NeuroEvolutionStrategy) calculateComplexity(genome map[string]interface{}) float64 {
	complexity := 0.0

	if arch, ok := genome["architecture"].([]int); ok {
		// 层数
		complexity += float64(len(arch)) * 0.1

		// 总参数数量
		totalParams := 0
		for i := 1; i < len(arch); i++ {
			totalParams += arch[i-1] * arch[i]
		}
		complexity += float64(totalParams) / 10000.0
	}

	return complexity
}



// GradientFreeStrategy 无梯度策略
type GradientFreeStrategy struct {
	config *EvolutionConfig
}

// NewGradientFreeStrategy 创建无梯度策略
func NewGradientFreeStrategy() *GradientFreeStrategy {
	return &GradientFreeStrategy{}
}

// Initialize 初始化无梯度策略
func (gfs *GradientFreeStrategy) Initialize(config *EvolutionConfig) error {
	gfs.config = config
	return nil
}

// Evolve 无梯度策略进化
func (gfs *GradientFreeStrategy) Evolve(ctx context.Context, population *Population, metrics *PerformanceMetrics) (*Population, error) {
	// (ES)
	return gfs.evolveWithES(population, metrics)
}

// Mutate 无梯度策略变异
func (gfs *GradientFreeStrategy) Mutate(individual *Individual, rate float64) *Individual {
	// ES
	mutated := *individual
	mutated.Genome = make(map[string]interface{})

	for key, value := range individual.Genome {
		if v, ok := value.(float64); ok {
			// 权重变异
			sigma := 0.1 // 变异标准差
			mutated.Genome[key] = v + rand.NormFloat64()*sigma
		} else {
			mutated.Genome[key] = value
		}
	}

	return &mutated
}

// Crossover 无梯度策略交叉
func (gfs *GradientFreeStrategy) Crossover(parent1, parent2 *Individual) (*Individual, *Individual) {
	// ES
	child1 := *parent1
	child2 := *parent2

	child1.Genome = make(map[string]interface{})
	child2.Genome = make(map[string]interface{})

	// 权重交叉
	for key := range parent1.Genome {
		if v1, ok1 := parent1.Genome[key].(float64); ok1 {
			if v2, ok2 := parent2.Genome[key].(float64); ok2 {
				child1.Genome[key] = (v1 + v2) / 2.0
				child2.Genome[key] = (v1 + v2) / 2.0
			}
		}
	}

	return &child1, &child2
}

// Select 无梯度策略选择
func (gfs *GradientFreeStrategy) Select(population *Population, count int) []Individual {
	// (+)
	return gfs.selectMuPlusLambda(population, count)
}

// Evaluate 无梯度策略评估
func (gfs *GradientFreeStrategy) Evaluate(individual *Individual, metrics *PerformanceMetrics) float64 {
	return metrics.Accuracy*0.6 + metrics.Efficiency*0.4
}

// evolveWithES 无梯度策略进化（ES）
func (gfs *GradientFreeStrategy) evolveWithES(population *Population, metrics *PerformanceMetrics) (*Population, error) {
	// 评估个体适应度
	for i := range population.Individuals {
		population.Individuals[i].Fitness = gfs.Evaluate(&population.Individuals[i], metrics)
	}

	// 选择父代
	parents := gfs.selectMuPlusLambda(population, population.Size/2)

	// 生成子代
	offspring := make([]Individual, 0, population.Size)

	for len(offspring) < population.Size {
		parent := parents[rand.Intn(len(parents))]
		child := gfs.Mutate(&parent, gfs.config.MutationRate)
		child.Generation = population.Generation + 1
		child.ID = fmt.Sprintf("es_ind_%d_%d", child.Generation, len(offspring))
		offspring = append(offspring, *child)
	}

	newPopulation := &Population{
		ID:          fmt.Sprintf("pop_%d", population.Generation+1),
		Individuals: offspring,
		Generation:  population.Generation + 1,
		Size:        len(offspring),
		CreatedAt:   time.Now(),
	}

	return newPopulation, nil
}

// selectMuPlusLambda 无梯度策略选择（+）
func (gfs *GradientFreeStrategy) selectMuPlusLambda(population *Population, mu int) []Individual {
	individuals := make([]Individual, len(population.Individuals))
	copy(individuals, population.Individuals)

	sort.Slice(individuals, func(i, j int) bool {
		return individuals[i].Fitness > individuals[j].Fitness
	})

	if mu > len(individuals) {
		mu = len(individuals)
	}

	return individuals[:mu]
}

type HybridStrategy struct{ *GeneticStrategy }
type ReinforcementStrategy struct{ *GeneticStrategy }
type SwarmIntelligenceStrategy struct{ *GeneticStrategy }

func NewHybridStrategy() *HybridStrategy {
	return &HybridStrategy{NewGeneticStrategy()}
}

func NewReinforcementStrategy() *ReinforcementStrategy {
	return &ReinforcementStrategy{NewGeneticStrategy()}
}

func NewSwarmIntelligenceStrategy() *SwarmIntelligenceStrategy {
	return &SwarmIntelligenceStrategy{NewGeneticStrategy()}
}
