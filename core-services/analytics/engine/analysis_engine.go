package engine

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/analytics"
)

// AnalysisEngine
type AnalysisEngine struct {
	config     *AnalysisEngineConfig
	algorithms map[string]Algorithm
	workers    []*AnalysisWorker
	taskQueue  chan *AnalysisTask
	results    map[string]*analytics.AnalysisResult
	mutex      sync.RWMutex
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// AnalysisEngineConfig
type AnalysisEngineConfig struct {
	//
	MaxWorkers     int           `json:"max_workers"`
	TaskQueueSize  int           `json:"task_queue_size"`
	TaskTimeout    time.Duration `json:"task_timeout"`
	ResultCacheTTL time.Duration `json:"result_cache_ttl"`

	//
	MaxMemoryUsage int64   `json:"max_memory_usage"`
	MaxCPUUsage    float64 `json:"max_cpu_usage"`
	BatchSize      int     `json:"batch_size"`

	// 㷨
	DefaultAlgorithms map[string]AlgorithmConfig `json:"default_algorithms"`

	//
	MetricsEnabled  bool          `json:"metrics_enabled"`
	MetricsInterval time.Duration `json:"metrics_interval"`

	//
	EnableSandbox    bool          `json:"enable_sandbox"`
	MaxExecutionTime time.Duration `json:"max_execution_time"`
}

// Algorithm 㷨
type Algorithm interface {
	Name() string
	Type() analytics.AnalysisType
	Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error)
	Validate(params map[string]interface{}) error
	GetRequiredParams() []string
	GetOptionalParams() []string
}

// AlgorithmConfig 㷨
type AlgorithmConfig struct {
	Enabled    bool                   `json:"enabled"`
	Parameters map[string]interface{} `json:"parameters"`
	Timeout    time.Duration          `json:"timeout"`
	MaxMemory  int64                  `json:"max_memory"`
}

// AlgorithmResult 㷨
type AlgorithmResult struct {
	Data     map[string]interface{}    `json:"data"`
	Insights []analytics.Insight       `json:"insights"`
	Metrics  analytics.AnalysisMetrics `json:"metrics"`
}

// AnalysisTask
type AnalysisTask struct {
	ID         string
	Type       analytics.AnalysisType
	Algorithm  string
	Data       []*analytics.DataPoint
	Parameters map[string]interface{}
	Context    context.Context
	ResultChan chan *TaskResult
	CreatedAt  time.Time
}

// TaskResult
type TaskResult struct {
	Task   *AnalysisTask
	Result *analytics.AnalysisResult
	Error  error
}

// AnalysisWorker ?
type AnalysisWorker struct {
	id       int
	engine   *AnalysisEngine
	taskChan chan *AnalysisTask
	stopCh   chan struct{}
}

// NewAnalysisEngine
func NewAnalysisEngine(config *AnalysisEngineConfig) *AnalysisEngine {
	if config == nil {
		config = &AnalysisEngineConfig{
			MaxWorkers:       10,
			TaskQueueSize:    1000,
			TaskTimeout:      30 * time.Minute,
			ResultCacheTTL:   1 * time.Hour,
			MaxMemoryUsage:   1024 * 1024 * 1024, // 1GB
			MaxCPUUsage:      0.8,
			BatchSize:        1000,
			MetricsEnabled:   true,
			MetricsInterval:  1 * time.Minute,
			EnableSandbox:    true,
			MaxExecutionTime: 1 * time.Hour,
		}
	}

	engine := &AnalysisEngine{
		config:     config,
		algorithms: make(map[string]Algorithm),
		taskQueue:  make(chan *AnalysisTask, config.TaskQueueSize),
		results:    make(map[string]*analytics.AnalysisResult),
		stopCh:     make(chan struct{}),
	}

	// 㷨
	engine.registerDefaultAlgorithms()

	// ?
	engine.startWorkers()

	//
	if config.MetricsEnabled {
		engine.startMetrics()
	}

	return engine
}

// RegisterAlgorithm 㷨
func (e *AnalysisEngine) RegisterAlgorithm(algorithm Algorithm) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if algorithm == nil {
		return fmt.Errorf("algorithm cannot be nil")
	}

	name := algorithm.Name()
	if name == "" {
		return fmt.Errorf("algorithm name cannot be empty")
	}

	e.algorithms[name] = algorithm
	return nil
}

// UnregisterAlgorithm 㷨
func (e *AnalysisEngine) UnregisterAlgorithm(name string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	delete(e.algorithms, name)
}

// GetAlgorithm 㷨
func (e *AnalysisEngine) GetAlgorithm(name string) (Algorithm, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	algorithm, exists := e.algorithms[name]
	return algorithm, exists
}

// ListAlgorithms ?
func (e *AnalysisEngine) ListAlgorithms() []string {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	names := make([]string, 0, len(e.algorithms))
	for name := range e.algorithms {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ExecuteAnalysis
func (e *AnalysisEngine) ExecuteAnalysis(ctx context.Context, req *analytics.DataAnalysisRequest) (*analytics.AnalysisResult, error) {
	//
	if err := e.validateAnalysisRequest(req); err != nil {
		return nil, fmt.Errorf("invalid analysis request: %w", err)
	}

	// 㷨
	algorithm, exists := e.GetAlgorithm(req.Algorithm)
	if !exists {
		return nil, fmt.Errorf("algorithm not found: %s", req.Algorithm)
	}

	//
	if err := algorithm.Validate(req.Parameters); err != nil {
		return nil, fmt.Errorf("invalid algorithm parameters: %w", err)
	}

	//
	task := &AnalysisTask{
		ID:         analytics.GenerateID(),
		Type:       req.Type,
		Algorithm:  req.Algorithm,
		Parameters: req.Parameters,
		Context:    ctx,
		ResultChan: make(chan *TaskResult, 1),
		CreatedAt:  time.Now(),
	}

	//
	select {
	case e.taskQueue <- task:
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("task queue is full")
	}

	//
	select {
	case result := <-task.ResultChan:
		if result.Error != nil {
			return nil, result.Error
		}
		return result.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(e.config.TaskTimeout):
		return nil, fmt.Errorf("analysis timeout")
	}
}

// ExecuteBatchAnalysis
func (e *AnalysisEngine) ExecuteBatchAnalysis(ctx context.Context, req *analytics.BatchDataAnalysisRequest) (*analytics.BatchDataAnalysisResponse, error) {
	if len(req.Requests) == 0 {
		return &analytics.BatchDataAnalysisResponse{
			Success: true,
			Results: []*analytics.AnalysisResult{},
		}, nil
	}

	results := make([]*analytics.AnalysisResult, 0, len(req.Requests))
	errors := make([]string, 0)
	processedCount := 0
	failedCount := 0

	//
	var wg sync.WaitGroup
	resultChan := make(chan struct {
		index  int
		result *analytics.AnalysisResult
		err    error
	}, len(req.Requests))

	for i, analysisReq := range req.Requests {
		wg.Add(1)
		go func(index int, req *analytics.DataAnalysisRequest) {
			defer wg.Done()
			result, err := e.ExecuteAnalysis(ctx, req)
			resultChan <- struct {
				index  int
				result *analytics.AnalysisResult
				err    error
			}{index, result, err}
		}(i, analysisReq)
	}

	// ?
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	//
	resultMap := make(map[int]*analytics.AnalysisResult)
	for res := range resultChan {
		if res.err != nil {
			errors = append(errors, fmt.Sprintf("Request %d: %s", res.index, res.err.Error()))
			failedCount++
		} else {
			resultMap[res.index] = res.result
			processedCount++
		}
	}

	// ?
	for i := 0; i < len(req.Requests); i++ {
		if result, exists := resultMap[i]; exists {
			results = append(results, result)
		} else {
			results = append(results, nil)
		}
	}

	return &analytics.BatchDataAnalysisResponse{
		Success:        processedCount > 0,
		Results:        results,
		ProcessedCount: processedCount,
		FailedCount:    failedCount,
		Errors:         errors,
	}, nil
}

// GetAnalysisResult
func (e *AnalysisEngine) GetAnalysisResult(id string) (*analytics.AnalysisResult, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	result, exists := e.results[id]
	return result, exists
}

// Stop
func (e *AnalysisEngine) Stop() {
	close(e.stopCh)
	e.wg.Wait()
}

//

func (e *AnalysisEngine) registerDefaultAlgorithms() {
	// ?
	e.RegisterAlgorithm(&DescriptiveAnalysisAlgorithm{})
	e.RegisterAlgorithm(&StatisticalSummaryAlgorithm{})

	// 㷨
	e.RegisterAlgorithm(&TrendAnalysisAlgorithm{})
	e.RegisterAlgorithm(&MovingAverageAlgorithm{})

	// ?
	e.RegisterAlgorithm(&AnomalyDetectionAlgorithm{})
	e.RegisterAlgorithm(&OutlierDetectionAlgorithm{})

	// ?
	e.RegisterAlgorithm(&CorrelationAnalysisAlgorithm{})

	// 㷨
	e.RegisterAlgorithm(&KMeansClusteringAlgorithm{})

	// 㷨
	e.RegisterAlgorithm(&DecisionTreeAlgorithm{})

	// 㷨
	e.RegisterAlgorithm(&LinearRegressionAlgorithm{})
	e.RegisterAlgorithm(&TimeSeriesForecastAlgorithm{})
}

func (e *AnalysisEngine) startWorkers() {
	e.workers = make([]*AnalysisWorker, e.config.MaxWorkers)
	for i := 0; i < e.config.MaxWorkers; i++ {
		worker := &AnalysisWorker{
			id:       i,
			engine:   e,
			taskChan: e.taskQueue,
			stopCh:   e.stopCh,
		}
		e.workers[i] = worker
		e.wg.Add(1)
		go worker.run()
	}
}

func (e *AnalysisEngine) startMetrics() {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		ticker := time.NewTicker(e.config.MetricsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				e.collectMetrics()
			case <-e.stopCh:
				return
			}
		}
	}()
}

func (e *AnalysisEngine) collectMetrics() {
	//
	//
}

func (e *AnalysisEngine) validateAnalysisRequest(req *analytics.DataAnalysisRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Algorithm == "" {
		return fmt.Errorf("algorithm cannot be empty")
	}

	if !analytics.ValidateAnalysisType(req.Type) {
		return fmt.Errorf("invalid analysis type: %s", req.Type)
	}

	return nil
}

// AnalysisWorker

func (w *AnalysisWorker) run() {
	defer w.engine.wg.Done()

	for {
		select {
		case task := <-w.taskChan:
			w.processTask(task)
		case <-w.stopCh:
			return
		}
	}
}

func (w *AnalysisWorker) processTask(task *AnalysisTask) {
	startTime := time.Now()

	// 㷨
	algorithm, exists := w.engine.GetAlgorithm(task.Algorithm)
	if !exists {
		task.ResultChan <- &TaskResult{
			Task:  task,
			Error: fmt.Errorf("algorithm not found: %s", task.Algorithm),
		}
		return
	}

	// 㷨
	algorithmResult, err := algorithm.Execute(task.Context, task.Data, task.Parameters)
	if err != nil {
		task.ResultChan <- &TaskResult{
			Task:  task,
			Error: fmt.Errorf("algorithm execution failed: %w", err),
		}
		return
	}

	//
	result := &analytics.AnalysisResult{
		ID:          task.ID,
		Type:        task.Type,
		Status:      analytics.AnalysisStatusCompleted,
		Algorithm:   task.Algorithm,
		Parameters:  task.Parameters,
		Results:     algorithmResult.Data,
		Insights:    algorithmResult.Insights,
		Metrics:     algorithmResult.Metrics,
		StartedAt:   task.CreatedAt,
		CompletedAt: &startTime,
		CreatedAt:   task.CreatedAt,
	}

	//
	result.Metrics.ProcessingTime = time.Since(startTime)

	//
	w.engine.mutex.Lock()
	w.engine.results[task.ID] = result
	w.engine.mutex.Unlock()

	//
	task.ResultChan <- &TaskResult{
		Task:   task,
		Result: result,
	}
}

// 㷨

// DescriptiveAnalysisAlgorithm ?
type DescriptiveAnalysisAlgorithm struct{}

func (a *DescriptiveAnalysisAlgorithm) Name() string {
	return "descriptive_analysis"
}

func (a *DescriptiveAnalysisAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypeDescriptive
}

func (a *DescriptiveAnalysisAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	// ?
	values := make([]float64, 0)
	for _, point := range data {
		if val, ok := point.Value.(float64); ok {
			values = append(values, val)
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no numeric data found")
	}

	//
	sort.Float64s(values)

	count := float64(len(values))
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / count

	// 㷽
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= count
	stdDev := math.Sqrt(variance)

	// ?
	min := values[0]
	max := values[len(values)-1]
	median := calculatePercentile(values, 50)
	q1 := calculatePercentile(values, 25)
	q3 := calculatePercentile(values, 75)

	results := map[string]interface{}{
		"count":    count,
		"sum":      sum,
		"mean":     mean,
		"median":   median,
		"min":      min,
		"max":      max,
		"std_dev":  stdDev,
		"variance": variance,
		"q1":       q1,
		"q3":       q3,
		"range":    max - min,
		"iqr":      q3 - q1,
	}

	//
	insights := []analytics.Insight{
		{
			Type:        analytics.InsightTypePattern,
			Title:       "",
			Description: fmt.Sprintf("?%.0f  %.2f?%.2f", count, mean, stdDev),
			Confidence:  1.0,
			Impact:      analytics.ImpactLevelMedium,
			Category:    "statistical_summary",
			Data: map[string]interface{}{
				"mean":    mean,
				"std_dev": stdDev,
				"count":   count,
			},
			CreatedAt: time.Now(),
		},
	}

	//
	if stdDev > 0 {
		outliers := 0
		for _, v := range values {
			if math.Abs(v-mean) > 2*stdDev {
				outliers++
			}
		}
		if outliers > 0 {
			insights = append(insights, analytics.Insight{
				Type:        analytics.InsightTypeAnomaly,
				Title:       "Outliers Detected",
				Description: fmt.Sprintf("Found %d outliers beyond 2 standard deviations", outliers),
				Confidence:  0.8,
				Impact:      analytics.ImpactLevelMedium,
				Category:    "outlier_detection",
				Data: map[string]interface{}{
					"outlier_count": outliers,
					"threshold":     2 * stdDev,
				},
				CreatedAt: time.Now(),
			})
		}
	}

	return &AlgorithmResult{
		Data:     results,
		Insights: insights,
		Metrics: analytics.AnalysisMetrics{
			ProcessedRecords: int64(len(data)),
		},
	}, nil
}

func (a *DescriptiveAnalysisAlgorithm) Validate(params map[string]interface{}) error {
	// ?
	return nil
}

func (a *DescriptiveAnalysisAlgorithm) GetRequiredParams() []string {
	return []string{}
}

func (a *DescriptiveAnalysisAlgorithm) GetOptionalParams() []string {
	return []string{}
}

// StatisticalSummaryAlgorithm 㷨
type StatisticalSummaryAlgorithm struct{}

func (a *StatisticalSummaryAlgorithm) Name() string {
	return "statistical_summary"
}

func (a *StatisticalSummaryAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypeDescriptive
}

func (a *StatisticalSummaryAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	// 㷨
	return &AlgorithmResult{
		Data:     map[string]interface{}{},
		Insights: []analytics.Insight{},
		Metrics:  analytics.AnalysisMetrics{},
	}, nil
}

func (a *StatisticalSummaryAlgorithm) Validate(params map[string]interface{}) error {
	return nil
}

func (a *StatisticalSummaryAlgorithm) GetRequiredParams() []string {
	return []string{}
}

func (a *StatisticalSummaryAlgorithm) GetOptionalParams() []string {
	return []string{}
}

// TrendAnalysisAlgorithm 㷨
type TrendAnalysisAlgorithm struct{}

func (a *TrendAnalysisAlgorithm) Name() string {
	return "trend_analysis"
}

func (a *TrendAnalysisAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypeTrend
}

func (a *TrendAnalysisAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	// 㷨
	return &AlgorithmResult{
		Data:     map[string]interface{}{},
		Insights: []analytics.Insight{},
		Metrics:  analytics.AnalysisMetrics{},
	}, nil
}

func (a *TrendAnalysisAlgorithm) Validate(params map[string]interface{}) error {
	return nil
}

func (a *TrendAnalysisAlgorithm) GetRequiredParams() []string {
	return []string{}
}

func (a *TrendAnalysisAlgorithm) GetOptionalParams() []string {
	return []string{"window_size", "smoothing_factor"}
}

// 㷨
type MovingAverageAlgorithm struct{}

func (a *MovingAverageAlgorithm) Name() string                 { return "moving_average" }
func (a *MovingAverageAlgorithm) Type() analytics.AnalysisType { return analytics.AnalysisTypeTrend }
func (a *MovingAverageAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	return &AlgorithmResult{Data: map[string]interface{}{}, Insights: []analytics.Insight{}, Metrics: analytics.AnalysisMetrics{}}, nil
}
func (a *MovingAverageAlgorithm) Validate(params map[string]interface{}) error { return nil }
func (a *MovingAverageAlgorithm) GetRequiredParams() []string                  { return []string{"window_size"} }
func (a *MovingAverageAlgorithm) GetOptionalParams() []string                  { return []string{} }

type AnomalyDetectionAlgorithm struct{}

func (a *AnomalyDetectionAlgorithm) Name() string { return "anomaly_detection" }
func (a *AnomalyDetectionAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypeAnomaly
}
func (a *AnomalyDetectionAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	return &AlgorithmResult{Data: map[string]interface{}{}, Insights: []analytics.Insight{}, Metrics: analytics.AnalysisMetrics{}}, nil
}
func (a *AnomalyDetectionAlgorithm) Validate(params map[string]interface{}) error { return nil }
func (a *AnomalyDetectionAlgorithm) GetRequiredParams() []string                  { return []string{} }
func (a *AnomalyDetectionAlgorithm) GetOptionalParams() []string {
	return []string{"threshold", "method"}
}

type OutlierDetectionAlgorithm struct{}

func (a *OutlierDetectionAlgorithm) Name() string { return "outlier_detection" }
func (a *OutlierDetectionAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypeAnomaly
}
func (a *OutlierDetectionAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	return &AlgorithmResult{Data: map[string]interface{}{}, Insights: []analytics.Insight{}, Metrics: analytics.AnalysisMetrics{}}, nil
}
func (a *OutlierDetectionAlgorithm) Validate(params map[string]interface{}) error { return nil }
func (a *OutlierDetectionAlgorithm) GetRequiredParams() []string                  { return []string{} }
func (a *OutlierDetectionAlgorithm) GetOptionalParams() []string {
	return []string{"method", "threshold"}
}

type CorrelationAnalysisAlgorithm struct{}

func (a *CorrelationAnalysisAlgorithm) Name() string { return "correlation_analysis" }
func (a *CorrelationAnalysisAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypeCorrelation
}
func (a *CorrelationAnalysisAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	return &AlgorithmResult{Data: map[string]interface{}{}, Insights: []analytics.Insight{}, Metrics: analytics.AnalysisMetrics{}}, nil
}
func (a *CorrelationAnalysisAlgorithm) Validate(params map[string]interface{}) error { return nil }
func (a *CorrelationAnalysisAlgorithm) GetRequiredParams() []string                  { return []string{"variables"} }
func (a *CorrelationAnalysisAlgorithm) GetOptionalParams() []string                  { return []string{"method"} }

type KMeansClusteringAlgorithm struct{}

func (a *KMeansClusteringAlgorithm) Name() string { return "kmeans_clustering" }
func (a *KMeansClusteringAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypeClustering
}
func (a *KMeansClusteringAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	return &AlgorithmResult{Data: map[string]interface{}{}, Insights: []analytics.Insight{}, Metrics: analytics.AnalysisMetrics{}}, nil
}
func (a *KMeansClusteringAlgorithm) Validate(params map[string]interface{}) error { return nil }
func (a *KMeansClusteringAlgorithm) GetRequiredParams() []string                  { return []string{"k"} }
func (a *KMeansClusteringAlgorithm) GetOptionalParams() []string {
	return []string{"max_iterations", "tolerance"}
}

type DecisionTreeAlgorithm struct{}

func (a *DecisionTreeAlgorithm) Name() string { return "decision_tree" }
func (a *DecisionTreeAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypeClassification
}
func (a *DecisionTreeAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	return &AlgorithmResult{Data: map[string]interface{}{}, Insights: []analytics.Insight{}, Metrics: analytics.AnalysisMetrics{}}, nil
}
func (a *DecisionTreeAlgorithm) Validate(params map[string]interface{}) error { return nil }
func (a *DecisionTreeAlgorithm) GetRequiredParams() []string                  { return []string{"target_variable"} }
func (a *DecisionTreeAlgorithm) GetOptionalParams() []string {
	return []string{"max_depth", "min_samples_split"}
}

type LinearRegressionAlgorithm struct{}

func (a *LinearRegressionAlgorithm) Name() string { return "linear_regression" }
func (a *LinearRegressionAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypePredictive
}
func (a *LinearRegressionAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	return &AlgorithmResult{Data: map[string]interface{}{}, Insights: []analytics.Insight{}, Metrics: analytics.AnalysisMetrics{}}, nil
}
func (a *LinearRegressionAlgorithm) Validate(params map[string]interface{}) error { return nil }
func (a *LinearRegressionAlgorithm) GetRequiredParams() []string {
	return []string{"target_variable", "features"}
}
func (a *LinearRegressionAlgorithm) GetOptionalParams() []string { return []string{"regularization"} }

type TimeSeriesForecastAlgorithm struct{}

func (a *TimeSeriesForecastAlgorithm) Name() string { return "time_series_forecast" }
func (a *TimeSeriesForecastAlgorithm) Type() analytics.AnalysisType {
	return analytics.AnalysisTypePredictive
}
func (a *TimeSeriesForecastAlgorithm) Execute(ctx context.Context, data []*analytics.DataPoint, params map[string]interface{}) (*AlgorithmResult, error) {
	return &AlgorithmResult{Data: map[string]interface{}{}, Insights: []analytics.Insight{}, Metrics: analytics.AnalysisMetrics{}}, nil
}
func (a *TimeSeriesForecastAlgorithm) Validate(params map[string]interface{}) error { return nil }
func (a *TimeSeriesForecastAlgorithm) GetRequiredParams() []string {
	return []string{"forecast_periods"}
}
func (a *TimeSeriesForecastAlgorithm) GetOptionalParams() []string {
	return []string{"seasonality", "trend"}
}

//

func calculatePercentile(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := (percentile / 100.0) * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}
