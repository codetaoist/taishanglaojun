package models

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultModelManager 
type DefaultModelManager struct {
	models    map[string]AIModel
	configs   map[string]*ModelConfig
	registry  ModelRegistry
	factory   ModelFactory
	logger    *zap.Logger
	mutex     sync.RWMutex
	metrics   map[string]*ModelMetrics
	metricsMu sync.RWMutex
}

// NewDefaultModelManager 
func NewDefaultModelManager(registry ModelRegistry, factory ModelFactory, logger *zap.Logger) *DefaultModelManager {
	return &DefaultModelManager{
		models:   make(map[string]AIModel),
		configs:  make(map[string]*ModelConfig),
		registry: registry,
		factory:  factory,
		logger:   logger,
		metrics:  make(map[string]*ModelMetrics),
	}
}

// LoadModel 
func (m *DefaultModelManager) LoadModel(ctx context.Context, config ModelConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid model config: %w", err)
	}

	// 
	if _, exists := m.models[config.ID]; exists {
		return fmt.Errorf("model %s already loaded", config.ID)
	}

	// 
	model, err := m.factory.CreateModel(config)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// 
	if err := model.Initialize(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize model: %w", err)
	}

	// 
	if err := model.Start(ctx); err != nil {
		return fmt.Errorf("failed to start model: %w", err)
	}

	// 
	if err := m.registry.Register(model); err != nil {
		model.Stop(ctx)
		return fmt.Errorf("failed to register model: %w", err)
	}

	// 
	m.models[config.ID] = model
	m.configs[config.ID] = &config

	// ?
	m.metricsMu.Lock()
	m.metrics[config.ID] = &ModelMetrics{
		TotalRequests:      0,
		SuccessfulRequests: 0,
		FailedRequests:     0,
		LastHealthCheck:    time.Now(),
	}
	m.metricsMu.Unlock()

	m.logger.Info("Model loaded successfully",
		zap.String("model_id", config.ID),
		zap.String("model_name", config.Name),
		zap.String("model_type", string(config.Type)))

	return nil
}

// UnloadModel 
func (m *DefaultModelManager) UnloadModel(ctx context.Context, modelID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	model, exists := m.models[modelID]
	if !exists {
		return fmt.Errorf("model %s not found", modelID)
	}

	// 
	if err := model.Stop(ctx); err != nil {
		m.logger.Warn("Failed to stop model gracefully",
			zap.String("model_id", modelID),
			zap.Error(err))
	}

	// 
	if err := m.registry.Unregister(modelID); err != nil {
		m.logger.Warn("Failed to unregister model",
			zap.String("model_id", modelID),
			zap.Error(err))
	}

	// 
	delete(m.models, modelID)
	delete(m.configs, modelID)

	m.metricsMu.Lock()
	delete(m.metrics, modelID)
	m.metricsMu.Unlock()

	m.logger.Info("Model unloaded successfully", zap.String("model_id", modelID))
	return nil
}

// ReloadModel 
func (m *DefaultModelManager) ReloadModel(ctx context.Context, modelID string) error {
	m.mutex.RLock()
	config, exists := m.configs[modelID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("model %s not found", modelID)
	}

	// 
	if err := m.UnloadModel(ctx, modelID); err != nil {
		return fmt.Errorf("failed to unload model: %w", err)
	}

	// 
	if err := m.LoadModel(ctx, *config); err != nil {
		return fmt.Errorf("failed to reload model: %w", err)
	}

	return nil
}

// ProcessRequest 
func (m *DefaultModelManager) ProcessRequest(ctx context.Context, modelID string, input ModelInput) (*ModelOutput, error) {
	startTime := time.Now()

	m.mutex.RLock()
	model, exists := m.models[modelID]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("model %s not found", modelID)
	}

	// 
	m.updateMetrics(modelID, func(metrics *ModelMetrics) {
		metrics.TotalRequests++
	})

	// 
	if err := model.Validate(ctx, input); err != nil {
		m.updateMetrics(modelID, func(metrics *ModelMetrics) {
			metrics.FailedRequests++
		})
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// 
	output, err := model.Process(ctx, input)
	processingTime := time.Since(startTime)

	if err != nil {
		m.updateMetrics(modelID, func(metrics *ModelMetrics) {
			metrics.FailedRequests++
		})
		return nil, fmt.Errorf("model processing failed: %w", err)
	}

	// 
	m.updateMetrics(modelID, func(metrics *ModelMetrics) {
		metrics.SuccessfulRequests++
		metrics.AverageLatency = time.Duration(
			(int64(metrics.AverageLatency)*metrics.SuccessfulRequests + int64(processingTime)) /
				(metrics.SuccessfulRequests + 1))
	})

	// 
	if output != nil {
		output.Metrics.ProcessingTime = processingTime
	}

	return output, nil
}

// BatchProcess 
func (m *DefaultModelManager) BatchProcess(ctx context.Context, modelID string, inputs []ModelInput) ([]*ModelOutput, error) {
	outputs := make([]*ModelOutput, len(inputs))
	errors := make([]error, len(inputs))

	// 
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) //  10

	for i, input := range inputs {
		wg.Add(1)
		go func(index int, inp ModelInput) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			output, err := m.ProcessRequest(ctx, modelID, inp)
			outputs[index] = output
			errors[index] = err
		}(i, input)
	}

	wg.Wait()

	// 
	var hasError bool
	for _, err := range errors {
		if err != nil {
			hasError = true
			break
		}
	}

	if hasError {
		return outputs, fmt.Errorf("batch processing completed with errors")
	}

	return outputs, nil
}

// StreamProcess 
func (m *DefaultModelManager) StreamProcess(ctx context.Context, modelID string, input ModelInput) (<-chan *ModelOutput, error) {
	m.mutex.RLock()
	model, exists := m.models[modelID]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("model %s not found", modelID)
	}

	// 
	capabilities := model.GetCapabilities()
	if !capabilities.SupportsStreaming {
		return nil, fmt.Errorf("model %s does not support streaming", modelID)
	}

	// 
	outputChan := make(chan *ModelOutput, 100)

	go func() {
		defer close(outputChan)

		// 
		// 
		output, err := model.Process(ctx, input)
		if err != nil {
			// 
			outputChan <- &ModelOutput{
				Error: &ModelError{
					Code:    "PROCESSING_ERROR",
					Message: err.Error(),
				},
			}
			return
		}

		outputChan <- output
	}()

	return outputChan, nil
}

// GetModelMetrics 
func (m *DefaultModelManager) GetModelMetrics(modelID string) (*ModelMetrics, error) {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()

	metrics, exists := m.metrics[modelID]
	if !exists {
		return nil, fmt.Errorf("metrics for model %s not found", modelID)
	}

	// 
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// GetAllMetrics 
func (m *DefaultModelManager) GetAllMetrics() (map[string]*ModelMetrics, error) {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()

	result := make(map[string]*ModelMetrics)
	for modelID, metrics := range m.metrics {
		metricsCopy := *metrics
		result[modelID] = &metricsCopy
	}

	return result, nil
}

// HealthCheck 
func (m *DefaultModelManager) HealthCheck(ctx context.Context, modelID string) error {
	m.mutex.RLock()
	model, exists := m.models[modelID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("model %s not found", modelID)
	}

	err := model.HealthCheck(ctx)
	if err != nil {
		// 
		m.updateMetrics(modelID, func(metrics *ModelMetrics) {
			metrics.HealthCheckFailures++
		})
		return fmt.Errorf("health check failed for model %s: %w", modelID, err)
	}

	// 
	m.updateMetrics(modelID, func(metrics *ModelMetrics) {
		metrics.HealthCheckSuccesses++
	})

	return nil
}

// UpdateModelConfig 
func (m *DefaultModelManager) UpdateModelConfig(modelID string, config ModelConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	model, exists := m.models[modelID]
	if !exists {
		return fmt.Errorf("model %s not found", modelID)
	}

	// 
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 
	ctx := context.Background()
	if err := model.Reload(ctx, config); err != nil {
		return fmt.Errorf("failed to reload model config: %w", err)
	}

	// 
	m.configs[modelID] = &config

	return nil
}

// GetModelConfig 
func (m *DefaultModelManager) GetModelConfig(modelID string) (*ModelConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config, exists := m.configs[modelID]
	if !exists {
		return nil, fmt.Errorf("config for model %s not found", modelID)
	}

	// 
	configCopy := *config
	return &configCopy, nil
}

// ListModels 
func (m *DefaultModelManager) ListModels() ([]*ModelConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	configs := make([]*ModelConfig, 0, len(m.configs))
	for _, config := range m.configs {
		configCopy := *config
		configs = append(configs, &configCopy)
	}

	return configs, nil
}

// updateMetrics 
func (m *DefaultModelManager) updateMetrics(modelID string, updateFunc func(*ModelMetrics)) {
	m.metricsMu.Lock()
	defer m.metricsMu.Unlock()

	if metrics, exists := m.metrics[modelID]; exists {
		updateFunc(metrics)
	}
}

// StartMetricsCollection 
func (m *DefaultModelManager) StartMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics(ctx)
		}
	}
}

// collectMetrics 
func (m *DefaultModelManager) collectMetrics(ctx context.Context) {
	m.mutex.RLock()
	models := make(map[string]AIModel)
	for id, model := range m.models {
		models[id] = model
	}
	m.mutex.RUnlock()

	for modelID, model := range models {
		// 
		if err := model.HealthCheck(ctx); err != nil {
			m.logger.Warn("Model health check failed",
				zap.String("model_id", modelID),
				zap.Error(err))
		}

		// 
		modelMetrics := model.GetMetrics()

		// 
		m.metricsMu.Lock()
		if metrics, exists := m.metrics[modelID]; exists {
			metrics.UpTime = modelMetrics.UpTime
			metrics.ResourceUsage = modelMetrics.ResourceUsage
			metrics.ThroughputRPS = modelMetrics.ThroughputRPS
			metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
		}
		m.metricsMu.Unlock()
	}
}

