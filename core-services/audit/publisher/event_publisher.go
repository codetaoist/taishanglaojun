package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"../audit"
)

// EventPublisher 事件发布器实现
type EventPublisher struct {
	providers map[string]PublisherProvider
	config    EventPublisherConfig
	logger    *zap.Logger
	
	// 状态管理
	mu       sync.RWMutex
	enabled  bool
	stats    PublisherStats
	
	// 异步处理
	eventChan chan *PublishTask
	workers   []*PublishWorker
	stopCh    chan struct{}
}

// EventPublisherConfig 事件发布器配置
type EventPublisherConfig struct {
	// 基础配置
	Enabled         bool                    `json:"enabled"`
	DefaultProvider string                  `json:"default_provider"`
	Providers       map[string]ProviderConfig `json:"providers"`
	
	// 异步配置
	AsyncEnabled    bool `json:"async_enabled"`
	WorkerCount     int  `json:"worker_count"`
	QueueSize       int  `json:"queue_size"`
	
	// 重试配置
	RetryEnabled    bool          `json:"retry_enabled"`
	MaxRetries      int           `json:"max_retries"`
	RetryInterval   time.Duration `json:"retry_interval"`
	RetryBackoff    string        `json:"retry_backoff"` // linear, exponential
	
	// 批处理配置
	BatchEnabled    bool          `json:"batch_enabled"`
	BatchSize       int           `json:"batch_size"`
	BatchTimeout    time.Duration `json:"batch_timeout"`
	
	// 过滤配置
	FilterEnabled   bool     `json:"filter_enabled"`
	FilterRules     []string `json:"filter_rules"`
	
	// 路由配置
	RoutingEnabled  bool                    `json:"routing_enabled"`
	RoutingRules    map[string]RoutingRule  `json:"routing_rules"`
	
	// 监控配置
	MetricsEnabled  bool          `json:"metrics_enabled"`
	MetricsInterval time.Duration `json:"metrics_interval"`
	
	// 安全配置
	EncryptionEnabled bool   `json:"encryption_enabled"`
	EncryptionKey     string `json:"encryption_key"`
	
	// 压缩配置
	CompressionEnabled bool   `json:"compression_enabled"`
	CompressionLevel   int    `json:"compression_level"`
	CompressionType    string `json:"compression_type"` // gzip, lz4, snappy
}

// ProviderConfig 提供者配置
type ProviderConfig struct {
	Type     string                 `json:"type"`     // kafka, rabbitmq, redis, webhook
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
	Config   map[string]interface{} `json:"config"`
}

// RoutingRule 路由规则
type RoutingRule struct {
	Condition string   `json:"condition"` // event_type=USER_LOGIN
	Providers []string `json:"providers"`
	Priority  int      `json:"priority"`
}

// PublisherStats 发布器统计
type PublisherStats struct {
	TotalEvents     int64     `json:"total_events"`
	SuccessfulEvents int64    `json:"successful_events"`
	FailedEvents    int64     `json:"failed_events"`
	QueuedEvents    int       `json:"queued_events"`
	AverageLatency  time.Duration `json:"average_latency"`
	EventsPerSecond float64   `json:"events_per_second"`
	ErrorRate       float64   `json:"error_rate"`
	StartTime       time.Time `json:"start_time"`
	LastEventTime   time.Time `json:"last_event_time"`
}

// PublishTask 发布任务
type PublishTask struct {
	Event     *audit.AuditEvent
	Providers []string
	Retries   int
	CreatedAt time.Time
}

// PublishWorker 发布工作器
type PublishWorker struct {
	id        int
	publisher *EventPublisher
	stopCh    chan struct{}
}

// PublisherProvider 发布器提供者接口
type PublisherProvider interface {
	PublishEvent(ctx context.Context, event *audit.AuditEvent) error
	PublishEvents(ctx context.Context, events []*audit.AuditEvent) error
	HealthCheck(ctx context.Context) error
	Close() error
}

// NewEventPublisher 创建事件发布器
func NewEventPublisher(config EventPublisherConfig, logger *zap.Logger) *EventPublisher {
	// 设置默认值
	if config.WorkerCount == 0 {
		config.WorkerCount = 4
	}
	if config.QueueSize == 0 {
		config.QueueSize = 10000
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryInterval == 0 {
		config.RetryInterval = 1 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.BatchTimeout == 0 {
		config.BatchTimeout = 5 * time.Second
	}
	if config.MetricsInterval == 0 {
		config.MetricsInterval = 1 * time.Minute
	}

	publisher := &EventPublisher{
		providers: make(map[string]PublisherProvider),
		config:    config,
		logger:    logger,
		enabled:   config.Enabled,
		stats: PublisherStats{
			StartTime: time.Now(),
		},
		stopCh: make(chan struct{}),
	}

	// 初始化异步处理
	if config.AsyncEnabled {
		publisher.eventChan = make(chan *PublishTask, config.QueueSize)
		publisher.initWorkers()
	}

	// 启动后台任务
	publisher.startBackgroundTasks()

	return publisher
}

// RegisterProvider 注册提供者
func (p *EventPublisher) RegisterProvider(name string, provider PublisherProvider) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	p.providers[name] = provider
	p.logger.Info("Provider registered", zap.String("name", name))
	return nil
}

// UnregisterProvider 注销提供者
func (p *EventPublisher) UnregisterProvider(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	provider, exists := p.providers[name]
	if !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	if err := provider.Close(); err != nil {
		p.logger.Warn("Failed to close provider", zap.String("name", name), zap.Error(err))
	}

	delete(p.providers, name)
	p.logger.Info("Provider unregistered", zap.String("name", name))
	return nil
}

// PublishEvent 发布单个事件
func (p *EventPublisher) PublishEvent(ctx context.Context, event *audit.AuditEvent) error {
	if !p.enabled {
		return nil
	}

	start := time.Now()
	defer func() {
		p.updateStats(time.Since(start), 1, true)
	}()

	// 应用过滤规则
	if p.config.FilterEnabled && p.shouldFilter(event) {
		p.logger.Debug("Event filtered", zap.String("event_id", event.ID))
		return nil
	}

	// 确定目标提供者
	providers := p.getTargetProviders(event)
	if len(providers) == 0 {
		p.logger.Warn("No providers available for event", zap.String("event_id", event.ID))
		return nil
	}

	// 异步处理
	if p.config.AsyncEnabled {
		return p.publishEventAsync(event, providers)
	}

	// 同步处理
	return p.publishEventSync(ctx, event, providers)
}

// PublishEvents 批量发布事件
func (p *EventPublisher) PublishEvents(ctx context.Context, events []*audit.AuditEvent) error {
	if !p.enabled || len(events) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		p.updateStats(time.Since(start), len(events), true)
	}()

	// 过滤事件
	validEvents := make([]*audit.AuditEvent, 0, len(events))
	for _, event := range events {
		if !p.config.FilterEnabled || !p.shouldFilter(event) {
			validEvents = append(validEvents, event)
		}
	}

	if len(validEvents) == 0 {
		return nil
	}

	// 按提供者分组事件
	providerEvents := p.groupEventsByProvider(validEvents)

	// 发布到各个提供者
	var lastErr error
	for providerName, providerEventList := range providerEvents {
		provider, exists := p.providers[providerName]
		if !exists {
			p.logger.Warn("Provider not found", zap.String("provider", providerName))
			continue
		}

		if err := provider.PublishEvents(ctx, providerEventList); err != nil {
			p.logger.Error("Failed to publish events to provider",
				zap.String("provider", providerName),
				zap.Int("count", len(providerEventList)),
				zap.Error(err))
			lastErr = err
		} else {
			p.logger.Debug("Events published to provider",
				zap.String("provider", providerName),
				zap.Int("count", len(providerEventList)))
		}
	}

	return lastErr
}

// HealthCheck 健康检查
func (p *EventPublisher) HealthCheck(ctx context.Context) error {
	if !p.enabled {
		return nil
	}

	p.mu.RLock()
	providers := make(map[string]PublisherProvider)
	for name, provider := range p.providers {
		providers[name] = provider
	}
	p.mu.RUnlock()

	var lastErr error
	for name, provider := range providers {
		if err := provider.HealthCheck(ctx); err != nil {
			p.logger.Warn("Provider health check failed",
				zap.String("provider", name),
				zap.Error(err))
			lastErr = err
		}
	}

	// 检查队列状态
	if p.config.AsyncEnabled {
		queueSize := len(p.eventChan)
		if queueSize > p.config.QueueSize*90/100 {
			return fmt.Errorf("event queue nearly full: %d/%d", queueSize, p.config.QueueSize)
		}
	}

	return lastErr
}

// GetStats 获取统计信息
func (p *EventPublisher) GetStats() PublisherStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.stats
	if p.config.AsyncEnabled {
		stats.QueuedEvents = len(p.eventChan)
	}

	if stats.TotalEvents > 0 {
		stats.ErrorRate = float64(stats.FailedEvents) / float64(stats.TotalEvents) * 100
		duration := time.Since(stats.StartTime).Seconds()
		stats.EventsPerSecond = float64(stats.TotalEvents) / duration
	}

	return stats
}

// Stop 停止发布器
func (p *EventPublisher) Stop() error {
	p.mu.Lock()
	p.enabled = false
	p.mu.Unlock()

	close(p.stopCh)

	// 停止工作器
	if p.config.AsyncEnabled {
		for _, worker := range p.workers {
			close(worker.stopCh)
		}
		close(p.eventChan)
	}

	// 关闭所有提供者
	for name, provider := range p.providers {
		if err := provider.Close(); err != nil {
			p.logger.Error("Failed to close provider",
				zap.String("provider", name),
				zap.Error(err))
		}
	}

	p.logger.Info("Event publisher stopped")
	return nil
}

// 私有方法

// initWorkers 初始化工作器
func (p *EventPublisher) initWorkers() {
	p.workers = make([]*PublishWorker, p.config.WorkerCount)
	for i := 0; i < p.config.WorkerCount; i++ {
		worker := &PublishWorker{
			id:        i,
			publisher: p,
			stopCh:    make(chan struct{}),
		}
		p.workers[i] = worker
		go worker.run()
	}
}

// startBackgroundTasks 启动后台任务
func (p *EventPublisher) startBackgroundTasks() {
	// 启动指标收集任务
	if p.config.MetricsEnabled {
		go p.metricsWorker()
	}

	// 启动健康检查任务
	go p.healthCheckWorker()
}

// metricsWorker 指标工作器
func (p *EventPublisher) metricsWorker() {
	ticker := time.NewTicker(p.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.collectMetrics()
		case <-p.stopCh:
			return
		}
	}
}

// healthCheckWorker 健康检查工作器
func (p *EventPublisher) healthCheckWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.HealthCheck(context.Background()); err != nil {
				p.logger.Warn("Publisher health check failed", zap.Error(err))
			}
		case <-p.stopCh:
			return
		}
	}
}

// publishEventAsync 异步发布事件
func (p *EventPublisher) publishEventAsync(event *audit.AuditEvent, providers []string) error {
	task := &PublishTask{
		Event:     event,
		Providers: providers,
		CreatedAt: time.Now(),
	}

	select {
	case p.eventChan <- task:
		return nil
	default:
		return fmt.Errorf("event queue is full")
	}
}

// publishEventSync 同步发布事件
func (p *EventPublisher) publishEventSync(ctx context.Context, event *audit.AuditEvent, providers []string) error {
	var lastErr error

	for _, providerName := range providers {
		provider, exists := p.providers[providerName]
		if !exists {
			p.logger.Warn("Provider not found", zap.String("provider", providerName))
			continue
		}

		if err := p.publishToProvider(ctx, provider, event); err != nil {
			p.logger.Error("Failed to publish event to provider",
				zap.String("provider", providerName),
				zap.String("event_id", event.ID),
				zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}

// publishToProvider 发布到提供者（带重试）
func (p *EventPublisher) publishToProvider(ctx context.Context, provider PublisherProvider, event *audit.AuditEvent) error {
	var err error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// 计算重试延迟
			delay := p.calculateRetryDelay(attempt)
			time.Sleep(delay)
		}

		err = provider.PublishEvent(ctx, event)
		if err == nil {
			return nil
		}

		p.logger.Warn("Publish attempt failed",
			zap.String("event_id", event.ID),
			zap.Int("attempt", attempt+1),
			zap.Error(err))
	}

	return fmt.Errorf("failed to publish after %d attempts: %w", p.config.MaxRetries+1, err)
}

// calculateRetryDelay 计算重试延迟
func (p *EventPublisher) calculateRetryDelay(attempt int) time.Duration {
	switch p.config.RetryBackoff {
	case "exponential":
		return p.config.RetryInterval * time.Duration(1<<uint(attempt-1))
	default: // linear
		return p.config.RetryInterval * time.Duration(attempt)
	}
}

// shouldFilter 检查是否应该过滤事件
func (p *EventPublisher) shouldFilter(event *audit.AuditEvent) bool {
	// 实现过滤逻辑
	// 这里简化实现
	return false
}

// getTargetProviders 获取目标提供者
func (p *EventPublisher) getTargetProviders(event *audit.AuditEvent) []string {
	if !p.config.RoutingEnabled {
		// 使用默认提供者
		if p.config.DefaultProvider != "" {
			return []string{p.config.DefaultProvider}
		}
		
		// 返回所有启用的提供者
		var providers []string
		for name, config := range p.config.Providers {
			if config.Enabled {
				providers = append(providers, name)
			}
		}
		return providers
	}

	// 应用路由规则
	return p.applyRoutingRules(event)
}

// applyRoutingRules 应用路由规则
func (p *EventPublisher) applyRoutingRules(event *audit.AuditEvent) []string {
	// 实现路由规则逻辑
	// 这里简化实现
	var providers []string
	
	for _, rule := range p.config.RoutingRules {
		if p.matchesCondition(event, rule.Condition) {
			providers = append(providers, rule.Providers...)
		}
	}

	if len(providers) == 0 && p.config.DefaultProvider != "" {
		providers = []string{p.config.DefaultProvider}
	}

	return providers
}

// matchesCondition 检查条件匹配
func (p *EventPublisher) matchesCondition(event *audit.AuditEvent, condition string) bool {
	// 实现条件匹配逻辑
	// 这里简化实现
	return true
}

// groupEventsByProvider 按提供者分组事件
func (p *EventPublisher) groupEventsByProvider(events []*audit.AuditEvent) map[string][]*audit.AuditEvent {
	providerEvents := make(map[string][]*audit.AuditEvent)

	for _, event := range events {
		providers := p.getTargetProviders(event)
		for _, provider := range providers {
			providerEvents[provider] = append(providerEvents[provider], event)
		}
	}

	return providerEvents
}

// updateStats 更新统计信息
func (p *EventPublisher) updateStats(latency time.Duration, count int, success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stats.TotalEvents += int64(count)
	p.stats.LastEventTime = time.Now()

	if success {
		p.stats.SuccessfulEvents += int64(count)
	} else {
		p.stats.FailedEvents += int64(count)
	}

	// 更新平均延迟
	if p.stats.TotalEvents == int64(count) {
		p.stats.AverageLatency = latency
	} else {
		totalLatency := p.stats.AverageLatency*time.Duration(p.stats.TotalEvents-int64(count)) + latency
		p.stats.AverageLatency = totalLatency / time.Duration(p.stats.TotalEvents)
	}
}

// collectMetrics 收集指标
func (p *EventPublisher) collectMetrics() {
	// 实现指标收集逻辑
	p.logger.Debug("Metrics collected", zap.Any("stats", p.GetStats()))
}

// PublishWorker 工作器运行
func (w *PublishWorker) run() {
	w.publisher.logger.Info("Publish worker started", zap.Int("worker_id", w.id))

	for {
		select {
		case task := <-w.publisher.eventChan:
			w.processTask(task)
		case <-w.stopCh:
			w.publisher.logger.Info("Publish worker stopped", zap.Int("worker_id", w.id))
			return
		}
	}
}

// processTask 处理任务
func (w *PublishWorker) processTask(task *PublishTask) {
	ctx := context.Background()
	
	// 检查任务是否过期
	if time.Since(task.CreatedAt) > 5*time.Minute {
		w.publisher.logger.Warn("Task expired",
			zap.String("event_id", task.Event.ID),
			zap.Duration("age", time.Since(task.CreatedAt)))
		return
	}

	// 发布事件
	err := w.publisher.publishEventSync(ctx, task.Event, task.Providers)
	if err != nil {
		// 重试逻辑
		if task.Retries < w.publisher.config.MaxRetries {
			task.Retries++
			
			// 重新加入队列
			select {
			case w.publisher.eventChan <- task:
				w.publisher.logger.Debug("Task requeued for retry",
					zap.String("event_id", task.Event.ID),
					zap.Int("retry", task.Retries))
			default:
				w.publisher.logger.Error("Failed to requeue task",
					zap.String("event_id", task.Event.ID))
			}
		} else {
			w.publisher.logger.Error("Task failed after max retries",
				zap.String("event_id", task.Event.ID),
				zap.Int("retries", task.Retries),
				zap.Error(err))
		}
	}
}