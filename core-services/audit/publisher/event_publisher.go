package publisher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/audit"
)

// EventPublisher 
type EventPublisher struct {
	providers map[string]PublisherProvider
	config    EventPublisherConfig
	logger    *zap.Logger

	// 
	mu      sync.RWMutex
	enabled bool
	stats   PublisherStats

	// 
	eventChan chan *PublishTask
	workers   []*PublishWorker
	stopCh    chan struct{}
}

// EventPublisherConfig 
type EventPublisherConfig struct {
	// 
	Enabled         bool                      `json:"enabled"`
	DefaultProvider string                    `json:"default_provider"`
	Providers       map[string]ProviderConfig `json:"providers"`

	// 
	AsyncEnabled bool `json:"async_enabled"`
	WorkerCount  int  `json:"worker_count"`
	QueueSize    int  `json:"queue_size"`

	// 
	RetryEnabled  bool          `json:"retry_enabled"`
	MaxRetries    int           `json:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval"`
	RetryBackoff  string        `json:"retry_backoff"` // linear, exponential

	// 
	BatchEnabled bool          `json:"batch_enabled"`
	BatchSize    int           `json:"batch_size"`
	BatchTimeout time.Duration `json:"batch_timeout"`

	// 
	FilterEnabled bool     `json:"filter_enabled"`
	FilterRules   []string `json:"filter_rules"`

	// 
	RoutingEnabled bool                   `json:"routing_enabled"`
	RoutingRules   map[string]RoutingRule `json:"routing_rules"`

	// 
	MetricsEnabled  bool          `json:"metrics_enabled"`
	MetricsInterval time.Duration `json:"metrics_interval"`

	// 
	EncryptionEnabled bool   `json:"encryption_enabled"`
	EncryptionKey     string `json:"encryption_key"`

	// 
	CompressionEnabled bool   `json:"compression_enabled"`
	CompressionLevel   int    `json:"compression_level"`
	CompressionType    string `json:"compression_type"` // gzip, lz4, snappy
}

// ProviderConfig 
type ProviderConfig struct {
	Type     string                 `json:"type"` // kafka, rabbitmq, redis, webhook
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
	Config   map[string]interface{} `json:"config"`
}

// RoutingRule 
type RoutingRule struct {
	Condition string   `json:"condition"` // event_type=USER_LOGIN
	Providers []string `json:"providers"`
	Priority  int      `json:"priority"`
}

// PublisherStats 
type PublisherStats struct {
	TotalEvents      int64         `json:"total_events"`
	SuccessfulEvents int64         `json:"successful_events"`
	FailedEvents     int64         `json:"failed_events"`
	QueuedEvents     int           `json:"queued_events"`
	AverageLatency   time.Duration `json:"average_latency"`
	EventsPerSecond  float64       `json:"events_per_second"`
	ErrorRate        float64       `json:"error_rate"`
	StartTime        time.Time     `json:"start_time"`
	LastEventTime    time.Time     `json:"last_event_time"`
}

// PublishTask 
type PublishTask struct {
	Event     *audit.AuditEvent
	Providers []string
	Retries   int
	CreatedAt time.Time
}

// PublishWorker 
type PublishWorker struct {
	id        int
	publisher *EventPublisher
	stopCh    chan struct{}
}

// PublisherProvider 
type PublisherProvider interface {
	PublishEvent(ctx context.Context, event *audit.AuditEvent) error
	PublishEvents(ctx context.Context, events []*audit.AuditEvent) error
	HealthCheck(ctx context.Context) error
	Close() error
}

// NewEventPublisher 
func NewEventPublisher(config EventPublisherConfig, logger *zap.Logger) *EventPublisher {
	// 
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

	// 
	if config.AsyncEnabled {
		publisher.eventChan = make(chan *PublishTask, config.QueueSize)
		publisher.initWorkers()
	}

	// 
	publisher.startBackgroundTasks()

	return publisher
}

// RegisterProvider 
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

// UnregisterProvider 
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

// PublishEvent 
func (p *EventPublisher) PublishEvent(ctx context.Context, event *audit.AuditEvent) error {
	if !p.enabled {
		return nil
	}

	start := time.Now()
	defer func() {
		p.updateStats(time.Since(start), 1, true)
	}()

	// 
	if p.config.FilterEnabled && p.shouldFilter(event) {
		p.logger.Debug("Event filtered", zap.String("event_id", event.ID))
		return nil
	}

	// 
	providers := p.getTargetProviders(event)
	if len(providers) == 0 {
		p.logger.Warn("No providers available for event", zap.String("event_id", event.ID))
		return nil
	}

	// 
	if p.config.AsyncEnabled {
		return p.publishEventAsync(event, providers)
	}

	// 
	return p.publishEventSync(ctx, event, providers)
}

// PublishEvents 
func (p *EventPublisher) PublishEvents(ctx context.Context, events []*audit.AuditEvent) error {
	if !p.enabled || len(events) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		p.updateStats(time.Since(start), len(events), true)
	}()

	// 
	validEvents := make([]*audit.AuditEvent, 0, len(events))
	for _, event := range events {
		if !p.config.FilterEnabled || !p.shouldFilter(event) {
			validEvents = append(validEvents, event)
		}
	}

	if len(validEvents) == 0 {
		return nil
	}

	// 
	providerEvents := p.groupEventsByProvider(validEvents)

	// 
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

// HealthCheck 鷢
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

	// 
	if p.config.AsyncEnabled {
		queueSize := len(p.eventChan)
		if queueSize > p.config.QueueSize*90/100 {
			return fmt.Errorf("event queue nearly full: %d/%d", queueSize, p.config.QueueSize)
		}
	}

	return lastErr
}

// GetStats 
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

// Stop 
func (p *EventPublisher) Stop() error {
	p.mu.Lock()
	p.enabled = false
	p.mu.Unlock()

	close(p.stopCh)

	// 
	if p.config.AsyncEnabled {
		for _, worker := range p.workers {
			close(worker.stopCh)
		}
		close(p.eventChan)
	}

	// 
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

// 

// initWorkers 
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

// startBackgroundTasks 
func (p *EventPublisher) startBackgroundTasks() {
	// 
	if p.config.MetricsEnabled {
		go p.metricsWorker()
	}

	// 
	go p.healthCheckWorker()
}

// metricsWorker 
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

// healthCheckWorker 鹤
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

// publishEventAsync 
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

// publishEventSync 
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

// publishToProvider 
func (p *EventPublisher) publishToProvider(ctx context.Context, provider PublisherProvider, event *audit.AuditEvent) error {
	var err error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// 
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

// calculateRetryDelay 
func (p *EventPublisher) calculateRetryDelay(attempt int) time.Duration {
	switch p.config.RetryBackoff {
	case "exponential":
		return p.config.RetryInterval * time.Duration(1<<uint(attempt-1))
	default: // linear
		return p.config.RetryInterval * time.Duration(attempt)
	}
}

// shouldFilter 
func (p *EventPublisher) shouldFilter(event *audit.AuditEvent) bool {
	// 
	// 
	return false
}

// getTargetProviders 
func (p *EventPublisher) getTargetProviders(event *audit.AuditEvent) []string {
	if !p.config.RoutingEnabled {
		// 
		if p.config.DefaultProvider != "" {
			return []string{p.config.DefaultProvider}
		}

		// 
		var providers []string
		for name, config := range p.config.Providers {
			if config.Enabled {
				providers = append(providers, name)
			}
		}
		return providers
	}

	// 
	return p.applyRoutingRules(event)
}

// applyRoutingRules 
func (p *EventPublisher) applyRoutingRules(event *audit.AuditEvent) []string {
	// 
	// 
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

// matchesCondition 
func (p *EventPublisher) matchesCondition(event *audit.AuditEvent, condition string) bool {
	// 
	// 
	return true
}

// groupEventsByProvider 
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

// updateStats 
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

	// 
	if p.stats.TotalEvents == int64(count) {
		p.stats.AverageLatency = latency
	} else {
		totalLatency := p.stats.AverageLatency*time.Duration(p.stats.TotalEvents-int64(count)) + latency
		p.stats.AverageLatency = totalLatency / time.Duration(p.stats.TotalEvents)
	}
}

// collectMetrics 
func (p *EventPublisher) collectMetrics() {
	// 
	p.logger.Debug("Metrics collected", zap.Any("stats", p.GetStats()))
}

// PublishWorker 
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

// processTask 
func (w *PublishWorker) processTask(task *PublishTask) {
	ctx := context.Background()

	// 
	if time.Since(task.CreatedAt) > 5*time.Minute {
		w.publisher.logger.Warn("Task expired",
			zap.String("event_id", task.Event.ID),
			zap.Duration("age", time.Since(task.CreatedAt)))
		return
	}

	// 
	err := w.publisher.publishEventSync(ctx, task.Event, task.Providers)
	if err != nil {
		// 
		if task.Retries < w.publisher.config.MaxRetries {
			task.Retries++

			// 
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

