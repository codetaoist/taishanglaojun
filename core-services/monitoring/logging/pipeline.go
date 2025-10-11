package logging

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LogPipeline 日志处理管道
type LogPipeline struct {
	config     LogPipelineConfig
	input      chan *LogEntry
	processors []LogProcessor
	outputs    []LogOutput
	stats      *LogPipelineStats
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// LogPipelineConfig 日志管道配置
type LogPipelineConfig struct {
	BufferSize     int           `json:"buffer_size" yaml:"buffer_size"`
	WorkerCount    int           `json:"worker_count" yaml:"worker_count"`
	BatchSize      int           `json:"batch_size" yaml:"batch_size"`
	FlushInterval  time.Duration `json:"flush_interval" yaml:"flush_interval"`
	FlushThreshold int           `json:"flush_threshold" yaml:"flush_threshold"`
	ProcessTimeout time.Duration `json:"process_timeout" yaml:"process_timeout"`
	RetryCount     int           `json:"retry_count" yaml:"retry_count"`
	RetryDelay     time.Duration `json:"retry_delay" yaml:"retry_delay"`
}

// LogPipelineStats 日志管道统计信息
type LogPipelineStats struct {
	ProcessedLogs  int64         `json:"processed_logs"`
	OutputLogs     int64         `json:"output_logs"`
	DroppedLogs    int64         `json:"dropped_logs"`
	ErrorLogs      int64         `json:"error_logs"`
	LastProcessed  time.Time     `json:"last_processed"`
	ProcessingTime time.Duration `json:"processing_time"`
	QueueSize      int           `json:"queue_size"`
	WorkerStats    []WorkerStats `json:"worker_stats"`
}

// WorkerStats 工作器统计信�?
type WorkerStats struct {
	ID             int           `json:"id"`
	ProcessedLogs  int64         `json:"processed_logs"`
	ErrorLogs      int64         `json:"error_logs"`
	LastProcessed  time.Time     `json:"last_processed"`
	ProcessingTime time.Duration `json:"processing_time"`
	IsActive       bool          `json:"is_active"`
}

// NewLogPipeline 创建日志管道
func NewLogPipeline(config LogPipelineConfig) (*LogPipeline, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 设置默认�?
	if config.BufferSize == 0 {
		config.BufferSize = 10000
	}
	if config.WorkerCount == 0 {
		config.WorkerCount = 4
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.FlushThreshold == 0 {
		config.FlushThreshold = 1000
	}
	if config.ProcessTimeout == 0 {
		config.ProcessTimeout = 30 * time.Second
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}
	
	pipeline := &LogPipeline{
		config:     config,
		input:      make(chan *LogEntry, config.BufferSize),
		processors: make([]LogProcessor, 0),
		outputs:    make([]LogOutput, 0),
		stats: &LogPipelineStats{
			WorkerStats: make([]WorkerStats, config.WorkerCount),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// 初始化工作器统计
	for i := 0; i < config.WorkerCount; i++ {
		pipeline.stats.WorkerStats[i] = WorkerStats{
			ID: i,
		}
	}
	
	return pipeline, nil
}

// Start 启动管道
func (lp *LogPipeline) Start() error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	// 启动工作�?
	for i := 0; i < lp.config.WorkerCount; i++ {
		lp.wg.Add(1)
		go lp.worker(i)
	}
	
	// 启动统计更新
	lp.wg.Add(1)
	go lp.updateStats()
	
	return nil
}

// Stop 停止管道
func (lp *LogPipeline) Stop() error {
	lp.cancel()
	
	// 关闭输入通道
	close(lp.input)
	
	// 等待所有工作器完成
	lp.wg.Wait()
	
	return nil
}

// Input 输入日志
func (lp *LogPipeline) Input(entry *LogEntry) error {
	select {
	case lp.input <- entry:
		return nil
	case <-lp.ctx.Done():
		return fmt.Errorf("pipeline is stopped")
	default:
		// 缓冲区满，丢弃日�?
		lp.recordDropped(1)
		return fmt.Errorf("pipeline buffer is full")
	}
}

// AddProcessor 添加处理�?
func (lp *LogPipeline) AddProcessor(processor LogProcessor) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.processors = append(lp.processors, processor)
}

// RemoveProcessor 移除处理�?
func (lp *LogPipeline) RemoveProcessor(processor LogProcessor) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	for i, p := range lp.processors {
		if p == processor {
			lp.processors = append(lp.processors[:i], lp.processors[i+1:]...)
			break
		}
	}
}

// AddOutput 添加输出
func (lp *LogPipeline) AddOutput(output LogOutput) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.outputs = append(lp.outputs, output)
}

// RemoveOutput 移除输出
func (lp *LogPipeline) RemoveOutput(output LogOutput) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	for i, o := range lp.outputs {
		if o == output {
			lp.outputs = append(lp.outputs[:i], lp.outputs[i+1:]...)
			break
		}
	}
}

// GetStats 获取统计信息
func (lp *LogPipeline) GetStats() *LogPipelineStats {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	
	// 复制统计信息
	stats := *lp.stats
	stats.QueueSize = len(lp.input)
	
	// 复制工作器统�?
	workerStats := make([]WorkerStats, len(lp.stats.WorkerStats))
	copy(workerStats, lp.stats.WorkerStats)
	stats.WorkerStats = workerStats
	
	return &stats
}

// HealthCheck 健康检�?
func (lp *LogPipeline) HealthCheck() error {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	
	// 检查是否有活跃的工作器
	activeWorkers := 0
	for _, worker := range lp.stats.WorkerStats {
		if worker.IsActive {
			activeWorkers++
		}
	}
	
	if activeWorkers == 0 {
		return fmt.Errorf("no active workers")
	}
	
	// 检查队列大�?
	queueSize := len(lp.input)
	if queueSize >= lp.config.BufferSize*9/10 {
		return fmt.Errorf("queue is nearly full: %d/%d", queueSize, lp.config.BufferSize)
	}
	
	return nil
}

// worker 工作�?
func (lp *LogPipeline) worker(id int) {
	defer lp.wg.Done()
	
	batch := make([]*LogEntry, 0, lp.config.BatchSize)
	flushTimer := time.NewTimer(lp.config.FlushInterval)
	defer flushTimer.Stop()
	
	for {
		select {
		case <-lp.ctx.Done():
			// 处理剩余的批�?
			if len(batch) > 0 {
				lp.processBatch(id, batch)
			}
			return
			
		case entry, ok := <-lp.input:
			if !ok {
				// 通道已关闭，处理剩余的批�?
				if len(batch) > 0 {
					lp.processBatch(id, batch)
				}
				return
			}
			
			// 标记工作器为活跃状�?
			lp.setWorkerActive(id, true)
			
			// 添加到批�?
			batch = append(batch, entry)
			
			// 检查是否需要刷�?
			if len(batch) >= lp.config.BatchSize {
				lp.processBatch(id, batch)
				batch = batch[:0] // 重置批次
				
				// 重置定时�?
				if !flushTimer.Stop() {
					<-flushTimer.C
				}
				flushTimer.Reset(lp.config.FlushInterval)
			}
			
		case <-flushTimer.C:
			// 定时刷新
			if len(batch) > 0 {
				lp.processBatch(id, batch)
				batch = batch[:0] // 重置批次
			}
			
			// 标记工作器为非活跃状�?
			lp.setWorkerActive(id, false)
			
			// 重置定时�?
			flushTimer.Reset(lp.config.FlushInterval)
		}
	}
}

// processBatch 处理批次
func (lp *LogPipeline) processBatch(workerID int, batch []*LogEntry) {
	start := time.Now()
	
	// 处理每个日志条目
	processedEntries := make([]*LogEntry, 0, len(batch))
	
	for _, entry := range batch {
		processedEntry := entry
		
		// 应用处理�?
		lp.mutex.RLock()
		processors := make([]LogProcessor, len(lp.processors))
		copy(processors, lp.processors)
		lp.mutex.RUnlock()
		
		for _, processor := range processors {
			var err error
			processedEntry, err = processor.Process(processedEntry)
			if err != nil {
				lp.recordWorkerError(workerID, 1)
				lp.recordError(1)
				continue
			}
			
			// 如果处理器返回nil，表示过滤掉该日�?
			if processedEntry == nil {
				break
			}
		}
		
		// 如果日志没有被过滤掉，添加到处理后的列表
		if processedEntry != nil {
			processedEntries = append(processedEntries, processedEntry)
		}
	}
	
	// 输出到所有输出器
	if len(processedEntries) > 0 {
		lp.mutex.RLock()
		outputs := make([]LogOutput, len(lp.outputs))
		copy(outputs, lp.outputs)
		lp.mutex.RUnlock()
		
		for _, output := range outputs {
			if err := lp.outputWithRetry(output, processedEntries); err != nil {
				lp.recordWorkerError(workerID, int64(len(processedEntries)))
				lp.recordError(int64(len(processedEntries)))
			} else {
				lp.recordOutput(int64(len(processedEntries)))
			}
		}
	}
	
	// 更新统计信息
	processingTime := time.Since(start)
	lp.recordWorkerProcessed(workerID, int64(len(batch)), processingTime)
	lp.recordProcessed(int64(len(batch)), processingTime)
}

// outputWithRetry 带重试的输出
func (lp *LogPipeline) outputWithRetry(output LogOutput, entries []*LogEntry) error {
	var lastErr error
	
	for i := 0; i <= lp.config.RetryCount; i++ {
		if err := output.Output(entries); err != nil {
			lastErr = err
			
			// 如果不是最后一次重试，等待后重�?
			if i < lp.config.RetryCount {
				time.Sleep(lp.config.RetryDelay * time.Duration(i+1))
				continue
			}
		} else {
			return nil
		}
	}
	
	return lastErr
}

// setWorkerActive 设置工作器活跃状�?
func (lp *LogPipeline) setWorkerActive(workerID int, active bool) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	if workerID < len(lp.stats.WorkerStats) {
		lp.stats.WorkerStats[workerID].IsActive = active
	}
}

// recordWorkerProcessed 记录工作器处理统�?
func (lp *LogPipeline) recordWorkerProcessed(workerID int, count int64, duration time.Duration) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	if workerID < len(lp.stats.WorkerStats) {
		lp.stats.WorkerStats[workerID].ProcessedLogs += count
		lp.stats.WorkerStats[workerID].LastProcessed = time.Now()
		lp.stats.WorkerStats[workerID].ProcessingTime = duration
	}
}

// recordWorkerError 记录工作器错误统�?
func (lp *LogPipeline) recordWorkerError(workerID int, count int64) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	if workerID < len(lp.stats.WorkerStats) {
		lp.stats.WorkerStats[workerID].ErrorLogs += count
	}
}

// recordProcessed 记录处理统计
func (lp *LogPipeline) recordProcessed(count int64, duration time.Duration) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.stats.ProcessedLogs += count
	lp.stats.LastProcessed = time.Now()
	lp.stats.ProcessingTime = duration
}

// recordOutput 记录输出统计
func (lp *LogPipeline) recordOutput(count int64) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.stats.OutputLogs += count
}

// recordError 记录错误统计
func (lp *LogPipeline) recordError(count int64) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.stats.ErrorLogs += count
}

// recordDropped 记录丢弃统计
func (lp *LogPipeline) recordDropped(count int64) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.stats.DroppedLogs += count
}

// updateStats 更新统计信息
func (lp *LogPipeline) updateStats() {
	defer lp.wg.Done()
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-lp.ctx.Done():
			return
		case <-ticker.C:
			// 更新队列大小
			lp.mutex.Lock()
			lp.stats.QueueSize = len(lp.input)
			lp.mutex.Unlock()
		}
	}
}
