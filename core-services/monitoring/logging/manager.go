package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
)

// LogManager 日志管理�?
type LogManager struct {
	config     LogManagerConfig
	collectors map[string]LogCollector
	processors map[string]LogProcessor
	outputs    map[string]LogOutput
	pipeline   *LogPipeline
	stats      *LogManagerStats
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// LogManagerConfig 日志管理器配�?
type LogManagerConfig struct {
	// 缓冲区配�?
	BufferSize     int           `json:"buffer_size" yaml:"buffer_size"`
	FlushInterval  time.Duration `json:"flush_interval" yaml:"flush_interval"`
	FlushThreshold int           `json:"flush_threshold" yaml:"flush_threshold"`
	
	// 处理配置
	WorkerCount    int           `json:"worker_count" yaml:"worker_count"`
	BatchSize      int           `json:"batch_size" yaml:"batch_size"`
	ProcessTimeout time.Duration `json:"process_timeout" yaml:"process_timeout"`
	
	// 存储配置
	RetentionDays  int    `json:"retention_days" yaml:"retention_days"`
	MaxLogSize     int64  `json:"max_log_size" yaml:"max_log_size"`
	CompressionEnabled bool `json:"compression_enabled" yaml:"compression_enabled"`
	
	// 监控配置
	MetricsEnabled bool          `json:"metrics_enabled" yaml:"metrics_enabled"`
	HealthCheck    time.Duration `json:"health_check" yaml:"health_check"`
	
	// 收集器配�?
	Collectors map[string]CollectorConfig `json:"collectors" yaml:"collectors"`
	
	// 处理器配�?
	Processors map[string]ProcessorConfig `json:"processors" yaml:"processors"`
	
	// 输出配置
	Outputs map[string]OutputConfig `json:"outputs" yaml:"outputs"`
}

// LogManagerStats 日志管理器统计信�?
type LogManagerStats struct {
	CollectedLogs  int64         `json:"collected_logs"`
	ProcessedLogs  int64         `json:"processed_logs"`
	OutputLogs     int64         `json:"output_logs"`
	DroppedLogs    int64         `json:"dropped_logs"`
	ErrorLogs      int64         `json:"error_logs"`
	LastProcessed  time.Time     `json:"last_processed"`
	ProcessingTime time.Duration `json:"processing_time"`
	QueueSize      int           `json:"queue_size"`
	
	// 收集器统�?
	CollectorStats map[string]*CollectorStats `json:"collector_stats"`
	
	// 处理器统�?
	ProcessorStats map[string]*ProcessorStats `json:"processor_stats"`
	
	// 输出统计
	OutputStats map[string]*OutputStats `json:"output_stats"`
}

// NewLogManager 创建日志管理�?
func NewLogManager(config LogManagerConfig) (*LogManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 设置默认�?
	if config.BufferSize == 0 {
		config.BufferSize = 10000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.WorkerCount == 0 {
		config.WorkerCount = 4
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.ProcessTimeout == 0 {
		config.ProcessTimeout = 30 * time.Second
	}
	if config.HealthCheck == 0 {
		config.HealthCheck = 30 * time.Second
	}
	
	manager := &LogManager{
		config:     config,
		collectors: make(map[string]LogCollector),
		processors: make(map[string]LogProcessor),
		outputs:    make(map[string]LogOutput),
		stats: &LogManagerStats{
			CollectorStats: make(map[string]*CollectorStats),
			ProcessorStats: make(map[string]*ProcessorStats),
			OutputStats:    make(map[string]*OutputStats),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// 创建日志处理管道
	pipeline, err := NewLogPipeline(LogPipelineConfig{
		BufferSize:     config.BufferSize,
		WorkerCount:    config.WorkerCount,
		BatchSize:      config.BatchSize,
		FlushInterval:  config.FlushInterval,
		FlushThreshold: config.FlushThreshold,
		ProcessTimeout: config.ProcessTimeout,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create log pipeline: %w", err)
	}
	manager.pipeline = pipeline
	
	// 初始化收集器
	if err := manager.initCollectors(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize collectors: %w", err)
	}
	
	// 初始化处理器
	if err := manager.initProcessors(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize processors: %w", err)
	}
	
	// 初始化输�?
	if err := manager.initOutputs(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize outputs: %w", err)
	}
	
	return manager, nil
}

// Start 启动日志管理�?
func (lm *LogManager) Start() error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	// 启动管道
	if err := lm.pipeline.Start(); err != nil {
		return fmt.Errorf("failed to start pipeline: %w", err)
	}
	
	// 启动收集�?
	for name, collector := range lm.collectors {
		if err := collector.Start(); err != nil {
			return fmt.Errorf("failed to start collector %s: %w", name, err)
		}
	}
	
	// 启动处理�?
	for name, processor := range lm.processors {
		if err := processor.Start(); err != nil {
			return fmt.Errorf("failed to start processor %s: %w", name, err)
		}
	}
	
	// 启动输出
	for name, output := range lm.outputs {
		if err := output.Start(); err != nil {
			return fmt.Errorf("failed to start output %s: %w", name, err)
		}
	}
	
	// 启动统计更新
	lm.wg.Add(1)
	go lm.updateStats()
	
	// 启动健康检�?
	if lm.config.HealthCheck > 0 {
		lm.wg.Add(1)
		go lm.healthCheck()
	}
	
	return nil
}

// Stop 停止日志管理�?
func (lm *LogManager) Stop() error {
	lm.cancel()
	lm.wg.Wait()
	
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	// 停止收集�?
	for _, collector := range lm.collectors {
		collector.Stop()
	}
	
	// 停止处理�?
	for _, processor := range lm.processors {
		processor.Stop()
	}
	
	// 停止输出
	for _, output := range lm.outputs {
		output.Stop()
	}
	
	// 停止管道
	return lm.pipeline.Stop()
}

// AddCollector 添加收集�?
func (lm *LogManager) AddCollector(name string, collector LogCollector) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	if _, exists := lm.collectors[name]; exists {
		return fmt.Errorf("collector %s already exists", name)
	}
	
	lm.collectors[name] = collector
	lm.stats.CollectorStats[name] = &CollectorStats{}
	
	// 设置日志处理回调
	collector.SetLogHandler(lm.pipeline.Input)
	
	return nil
}

// RemoveCollector 移除收集�?
func (lm *LogManager) RemoveCollector(name string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	collector, exists := lm.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	collector.Stop()
	delete(lm.collectors, name)
	delete(lm.stats.CollectorStats, name)
	
	return nil
}

// AddProcessor 添加处理�?
func (lm *LogManager) AddProcessor(name string, processor LogProcessor) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	if _, exists := lm.processors[name]; exists {
		return fmt.Errorf("processor %s already exists", name)
	}
	
	lm.processors[name] = processor
	lm.stats.ProcessorStats[name] = &ProcessorStats{}
	
	// 添加到管�?
	lm.pipeline.AddProcessor(processor)
	
	return nil
}

// RemoveProcessor 移除处理�?
func (lm *LogManager) RemoveProcessor(name string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	processor, exists := lm.processors[name]
	if !exists {
		return fmt.Errorf("processor %s not found", name)
	}
	
	processor.Stop()
	delete(lm.processors, name)
	delete(lm.stats.ProcessorStats, name)
	
	// 从管道移�?
	lm.pipeline.RemoveProcessor(processor)
	
	return nil
}

// AddOutput 添加输出
func (lm *LogManager) AddOutput(name string, output LogOutput) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	if _, exists := lm.outputs[name]; exists {
		return fmt.Errorf("output %s already exists", name)
	}
	
	lm.outputs[name] = output
	lm.stats.OutputStats[name] = &OutputStats{}
	
	// 添加到管�?
	lm.pipeline.AddOutput(output)
	
	return nil
}

// RemoveOutput 移除输出
func (lm *LogManager) RemoveOutput(name string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	output, exists := lm.outputs[name]
	if !exists {
		return fmt.Errorf("output %s not found", name)
	}
	
	output.Stop()
	delete(lm.outputs, name)
	delete(lm.stats.OutputStats, name)
	
	// 从管道移�?
	lm.pipeline.RemoveOutput(output)
	
	return nil
}

// GetCollector 获取收集�?
func (lm *LogManager) GetCollector(name string) (LogCollector, bool) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	collector, exists := lm.collectors[name]
	return collector, exists
}

// GetProcessor 获取处理�?
func (lm *LogManager) GetProcessor(name string) (LogProcessor, bool) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	processor, exists := lm.processors[name]
	return processor, exists
}

// GetOutput 获取输出
func (lm *LogManager) GetOutput(name string) (LogOutput, bool) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	output, exists := lm.outputs[name]
	return output, exists
}

// ListCollectors 列出所有收集器
func (lm *LogManager) ListCollectors() []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	names := make([]string, 0, len(lm.collectors))
	for name := range lm.collectors {
		names = append(names, name)
	}
	return names
}

// ListProcessors 列出所有处理器
func (lm *LogManager) ListProcessors() []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	names := make([]string, 0, len(lm.processors))
	for name := range lm.processors {
		names = append(names, name)
	}
	return names
}

// ListOutputs 列出所有输�?
func (lm *LogManager) ListOutputs() []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	names := make([]string, 0, len(lm.outputs))
	for name := range lm.outputs {
		names = append(names, name)
	}
	return names
}

// GetStats 获取统计信息
func (lm *LogManager) GetStats() *LogManagerStats {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	// 复制统计信息
	stats := *lm.stats
	
	// 复制收集器统�?
	stats.CollectorStats = make(map[string]*CollectorStats)
	for name, stat := range lm.stats.CollectorStats {
		statCopy := *stat
		stats.CollectorStats[name] = &statCopy
	}
	
	// 复制处理器统�?
	stats.ProcessorStats = make(map[string]*ProcessorStats)
	for name, stat := range lm.stats.ProcessorStats {
		statCopy := *stat
		stats.ProcessorStats[name] = &statCopy
	}
	
	// 复制输出统计
	stats.OutputStats = make(map[string]*OutputStats)
	for name, stat := range lm.stats.OutputStats {
		statCopy := *stat
		stats.OutputStats[name] = &statCopy
	}
	
	return &stats
}

// HealthCheck 健康检�?
func (lm *LogManager) HealthCheck() error {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	// 检查管�?
	if err := lm.pipeline.HealthCheck(); err != nil {
		return fmt.Errorf("pipeline health check failed: %w", err)
	}
	
	// 检查收集器
	for name, collector := range lm.collectors {
		if err := collector.HealthCheck(); err != nil {
			return fmt.Errorf("collector %s health check failed: %w", name, err)
		}
	}
	
	// 检查处理器
	for name, processor := range lm.processors {
		if err := processor.HealthCheck(); err != nil {
			return fmt.Errorf("processor %s health check failed: %w", name, err)
		}
	}
	
	// 检查输�?
	for name, output := range lm.outputs {
		if err := output.HealthCheck(); err != nil {
			return fmt.Errorf("output %s health check failed: %w", name, err)
		}
	}
	
	return nil
}

// initCollectors 初始化收集器
func (lm *LogManager) initCollectors() error {
	for name, config := range lm.config.Collectors {
		collector, err := CreateCollector(config)
		if err != nil {
			return fmt.Errorf("failed to create collector %s: %w", name, err)
		}
		
		if err := lm.AddCollector(name, collector); err != nil {
			return fmt.Errorf("failed to add collector %s: %w", name, err)
		}
	}
	
	return nil
}

// initProcessors 初始化处理器
func (lm *LogManager) initProcessors() error {
	for name, config := range lm.config.Processors {
		processor, err := CreateProcessor(config)
		if err != nil {
			return fmt.Errorf("failed to create processor %s: %w", name, err)
		}
		
		if err := lm.AddProcessor(name, processor); err != nil {
			return fmt.Errorf("failed to add processor %s: %w", name, err)
		}
	}
	
	return nil
}

// initOutputs 初始化输�?
func (lm *LogManager) initOutputs() error {
	for name, config := range lm.config.Outputs {
		output, err := CreateOutput(config)
		if err != nil {
			return fmt.Errorf("failed to create output %s: %w", name, err)
		}
		
		if err := lm.AddOutput(name, output); err != nil {
			return fmt.Errorf("failed to add output %s: %w", name, err)
		}
	}
	
	return nil
}

// updateStats 更新统计信息
func (lm *LogManager) updateStats() {
	defer lm.wg.Done()
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-lm.ctx.Done():
			return
		case <-ticker.C:
			lm.doUpdateStats()
		}
	}
}

// doUpdateStats 执行统计更新
func (lm *LogManager) doUpdateStats() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	// 更新管道统计
	pipelineStats := lm.pipeline.GetStats()
	lm.stats.ProcessedLogs = pipelineStats.ProcessedLogs
	lm.stats.OutputLogs = pipelineStats.OutputLogs
	lm.stats.DroppedLogs = pipelineStats.DroppedLogs
	lm.stats.ErrorLogs = pipelineStats.ErrorLogs
	lm.stats.LastProcessed = pipelineStats.LastProcessed
	lm.stats.ProcessingTime = pipelineStats.ProcessingTime
	lm.stats.QueueSize = pipelineStats.QueueSize
	
	// 更新收集器统�?
	var totalCollected int64
	for name, collector := range lm.collectors {
		stats := collector.GetStats()
		lm.stats.CollectorStats[name] = stats
		totalCollected += stats.CollectedLogs
	}
	lm.stats.CollectedLogs = totalCollected
	
	// 更新处理器统�?
	for name, processor := range lm.processors {
		stats := processor.GetStats()
		lm.stats.ProcessorStats[name] = stats
	}
	
	// 更新输出统计
	for name, output := range lm.outputs {
		stats := output.GetStats()
		lm.stats.OutputStats[name] = stats
	}
}

// healthCheck 健康检查循�?
func (lm *LogManager) healthCheck() {
	defer lm.wg.Done()
	
	ticker := time.NewTicker(lm.config.HealthCheck)
	defer ticker.Stop()
	
	for {
		select {
		case <-lm.ctx.Done():
			return
		case <-ticker.C:
			if err := lm.HealthCheck(); err != nil {
				// 记录健康检查错�?
				fmt.Printf("Health check failed: %v\n", err)
			}
		}
	}
}

// 确保LogManager实现了interfaces.LogManager接口
var _ interfaces.LogManager = (*LogManager)(nil)
