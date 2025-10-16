package logging

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LogPipeline 
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

// LogPipelineConfig 
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

// LogPipelineStats 
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

// WorkerStats ?
type WorkerStats struct {
	ID             int           `json:"id"`
	ProcessedLogs  int64         `json:"processed_logs"`
	ErrorLogs      int64         `json:"error_logs"`
	LastProcessed  time.Time     `json:"last_processed"`
	ProcessingTime time.Duration `json:"processing_time"`
	IsActive       bool          `json:"is_active"`
}

// NewLogPipeline 
func NewLogPipeline(config LogPipelineConfig) (*LogPipeline, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// ?
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
	
	// 
	for i := 0; i < config.WorkerCount; i++ {
		pipeline.stats.WorkerStats[i] = WorkerStats{
			ID: i,
		}
	}
	
	return pipeline, nil
}

// Start 
func (lp *LogPipeline) Start() error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	// ?
	for i := 0; i < lp.config.WorkerCount; i++ {
		lp.wg.Add(1)
		go lp.worker(i)
	}
	
	// 
	lp.wg.Add(1)
	go lp.updateStats()
	
	return nil
}

// Stop 
func (lp *LogPipeline) Stop() error {
	lp.cancel()
	
	// 
	close(lp.input)
	
	// 
	lp.wg.Wait()
	
	return nil
}

// Input 
func (lp *LogPipeline) Input(entry *LogEntry) error {
	select {
	case lp.input <- entry:
		return nil
	case <-lp.ctx.Done():
		return fmt.Errorf("pipeline is stopped")
	default:
		// ?
		lp.recordDropped(1)
		return fmt.Errorf("pipeline buffer is full")
	}
}

// AddProcessor ?
func (lp *LogPipeline) AddProcessor(processor LogProcessor) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.processors = append(lp.processors, processor)
}

// RemoveProcessor ?
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

// AddOutput 
func (lp *LogPipeline) AddOutput(output LogOutput) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.outputs = append(lp.outputs, output)
}

// RemoveOutput 
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

// GetStats 
func (lp *LogPipeline) GetStats() *LogPipelineStats {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	
	// 
	stats := *lp.stats
	stats.QueueSize = len(lp.input)
	
	// ?
	workerStats := make([]WorkerStats, len(lp.stats.WorkerStats))
	copy(workerStats, lp.stats.WorkerStats)
	stats.WorkerStats = workerStats
	
	return &stats
}

// HealthCheck ?
func (lp *LogPipeline) HealthCheck() error {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	
	// 
	activeWorkers := 0
	for _, worker := range lp.stats.WorkerStats {
		if worker.IsActive {
			activeWorkers++
		}
	}
	
	if activeWorkers == 0 {
		return fmt.Errorf("no active workers")
	}
	
	// ?
	queueSize := len(lp.input)
	if queueSize >= lp.config.BufferSize*9/10 {
		return fmt.Errorf("queue is nearly full: %d/%d", queueSize, lp.config.BufferSize)
	}
	
	return nil
}

// worker ?
func (lp *LogPipeline) worker(id int) {
	defer lp.wg.Done()
	
	batch := make([]*LogEntry, 0, lp.config.BatchSize)
	flushTimer := time.NewTimer(lp.config.FlushInterval)
	defer flushTimer.Stop()
	
	for {
		select {
		case <-lp.ctx.Done():
			// ?
			if len(batch) > 0 {
				lp.processBatch(id, batch)
			}
			return
			
		case entry, ok := <-lp.input:
			if !ok {
				// ?
				if len(batch) > 0 {
					lp.processBatch(id, batch)
				}
				return
			}
			
			// ?
			lp.setWorkerActive(id, true)
			
			// ?
			batch = append(batch, entry)
			
			// ?
			if len(batch) >= lp.config.BatchSize {
				lp.processBatch(id, batch)
				batch = batch[:0] // 
				
				// ?
				if !flushTimer.Stop() {
					<-flushTimer.C
				}
				flushTimer.Reset(lp.config.FlushInterval)
			}
			
		case <-flushTimer.C:
			// 
			if len(batch) > 0 {
				lp.processBatch(id, batch)
				batch = batch[:0] // 
			}
			
			// ?
			lp.setWorkerActive(id, false)
			
			// ?
			flushTimer.Reset(lp.config.FlushInterval)
		}
	}
}

// processBatch 
func (lp *LogPipeline) processBatch(workerID int, batch []*LogEntry) {
	start := time.Now()
	
	// 
	processedEntries := make([]*LogEntry, 0, len(batch))
	
	for _, entry := range batch {
		processedEntry := entry
		
		// ?
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
			
			// nil?
			if processedEntry == nil {
				break
			}
		}
		
		// 
		if processedEntry != nil {
			processedEntries = append(processedEntries, processedEntry)
		}
	}
	
	// 
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
	
	// 
	processingTime := time.Since(start)
	lp.recordWorkerProcessed(workerID, int64(len(batch)), processingTime)
	lp.recordProcessed(int64(len(batch)), processingTime)
}

// outputWithRetry 
func (lp *LogPipeline) outputWithRetry(output LogOutput, entries []*LogEntry) error {
	var lastErr error
	
	for i := 0; i <= lp.config.RetryCount; i++ {
		if err := output.Output(entries); err != nil {
			lastErr = err
			
			// ?
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

// setWorkerActive ?
func (lp *LogPipeline) setWorkerActive(workerID int, active bool) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	if workerID < len(lp.stats.WorkerStats) {
		lp.stats.WorkerStats[workerID].IsActive = active
	}
}

// recordWorkerProcessed ?
func (lp *LogPipeline) recordWorkerProcessed(workerID int, count int64, duration time.Duration) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	if workerID < len(lp.stats.WorkerStats) {
		lp.stats.WorkerStats[workerID].ProcessedLogs += count
		lp.stats.WorkerStats[workerID].LastProcessed = time.Now()
		lp.stats.WorkerStats[workerID].ProcessingTime = duration
	}
}

// recordWorkerError ?
func (lp *LogPipeline) recordWorkerError(workerID int, count int64) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	if workerID < len(lp.stats.WorkerStats) {
		lp.stats.WorkerStats[workerID].ErrorLogs += count
	}
}

// recordProcessed 
func (lp *LogPipeline) recordProcessed(count int64, duration time.Duration) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.stats.ProcessedLogs += count
	lp.stats.LastProcessed = time.Now()
	lp.stats.ProcessingTime = duration
}

// recordOutput 
func (lp *LogPipeline) recordOutput(count int64) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.stats.OutputLogs += count
}

// recordError 
func (lp *LogPipeline) recordError(count int64) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.stats.ErrorLogs += count
}

// recordDropped 
func (lp *LogPipeline) recordDropped(count int64) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	lp.stats.DroppedLogs += count
}

// updateStats 
func (lp *LogPipeline) updateStats() {
	defer lp.wg.Done()
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-lp.ctx.Done():
			return
		case <-ticker.C:
			// 
			lp.mutex.Lock()
			lp.stats.QueueSize = len(lp.input)
			lp.mutex.Unlock()
		}
	}
}

