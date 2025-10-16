package audit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultAuditService 默认审计服务实现
type DefaultAuditService struct {
	repository AuditRepository
	publisher  AuditEventPublisher
	logger     *zap.Logger
	config     DefaultAuditServiceConfig

	//
	mu          sync.RWMutex
	eventBuffer []*AuditEvent
	bufferSize  int
	lastFlush   time.Time

	//
	stats AuditServiceStats

	//
	stopCh      chan struct{}
	flushTicker *time.Ticker
}

// DefaultAuditServiceConfig 默认审计服务配置
type DefaultAuditServiceConfig struct {
	// 緩衝區大小
	BufferSize int `json:"buffer_size"`
	// 刷新間隔
	FlushInterval time.Duration `json:"flush_interval"`
	// 刷新閾值
	FlushThreshold int `json:"flush_threshold"`

	//
	AsyncLogging bool `json:"async_logging"`
	WorkerCount  int  `json:"worker_count"`
	QueueSize    int  `json:"queue_size"`

	//
	RetryEnabled  bool          `json:"retry_enabled"`
	MaxRetries    int           `json:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval"`

	//
	EnableFiltering bool     `json:"enable_filtering"`
	FilterRules     []string `json:"filter_rules"`

	//
	EnableEncryption bool   `json:"enable_encryption"`
	EncryptionKey    string `json:"encryption_key"`

	//
	EnableCompliance   bool     `json:"enable_compliance"`
	ComplianceRules    []string `json:"compliance_rules"`
	DataClassification bool     `json:"data_classification"`

	//
	EnableCompression bool `json:"enable_compression"`
	CompressionLevel  int  `json:"compression_level"`
	EnableBatching    bool `json:"enable_batching"`
	BatchSize         int  `json:"batch_size"`

	//
	EnableMetrics     bool          `json:"enable_metrics"`
	MetricsInterval   time.Duration `json:"metrics_interval"`
	EnableHealthCheck bool          `json:"enable_health_check"`

	// 洢
	RetentionPolicy RetentionPolicy `json:"retention_policy"`
	ArchiveEnabled  bool            `json:"archive_enabled"`

	// 澯
	EnableAlerting  bool               `json:"enable_alerting"`
	AlertRules      []string           `json:"alert_rules"`
	AlertThresholds map[string]float64 `json:"alert_thresholds"`
}

// AuditServiceStats
type AuditServiceStats struct {
	TotalEvents      int64         `json:"total_events"`
	SuccessfulEvents int64         `json:"successful_events"`
	FailedEvents     int64         `json:"failed_events"`
	BufferedEvents   int           `json:"buffered_events"`
	LastFlushTime    time.Time     `json:"last_flush_time"`
	AverageLatency   time.Duration `json:"average_latency"`
	EventsPerSecond  float64       `json:"events_per_second"`
	ErrorRate        float64       `json:"error_rate"`
	StartTime        time.Time     `json:"start_time"`
	Uptime           time.Duration `json:"uptime"`
}

// NewDefaultAuditService
func NewDefaultAuditService(
	repository AuditRepository,
	publisher AuditEventPublisher,
	config DefaultAuditServiceConfig,
	logger *zap.Logger,
) *DefaultAuditService {
	//
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.FlushThreshold == 0 {
		config.FlushThreshold = 100
	}
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
		config.BatchSize = 50
	}
	if config.MetricsInterval == 0 {
		config.MetricsInterval = 1 * time.Minute
	}

	service := &DefaultAuditService{
		repository:  repository,
		publisher:   publisher,
		logger:      logger,
		config:      config,
		eventBuffer: make([]*AuditEvent, 0, config.BufferSize),
		stopCh:      make(chan struct{}),
		stats: AuditServiceStats{
			StartTime: time.Now(),
		},
	}

	//
	service.startBackgroundTasks()

	return service
}

// LogEvent
func (s *DefaultAuditService) LogEvent(ctx context.Context, event *AuditEvent) error {
	start := time.Now()
	defer func() {
		s.updateStats(time.Since(start), true)
	}()

	//
	if err := s.validateEvent(event); err != nil {
		s.updateStats(time.Since(start), false)
		return fmt.Errorf("invalid audit event: %w", err)
	}

	//
	if s.config.EnableFiltering && s.shouldFilter(event) {
		s.logger.Debug("Event filtered", zap.String("event_id", event.ID))
		return nil
	}

	//
	if s.config.EnableCompliance {
		s.applyComplianceRules(event)
	}

	//
	if s.config.DataClassification {
		s.classifyData(event)
	}

	//
	if s.config.EnableEncryption {
		if err := s.encryptSensitiveData(event); err != nil {
			s.logger.Warn("Failed to encrypt sensitive data",
				zap.String("event_id", event.ID),
				zap.Error(err))
		}
	}

	//
	if s.config.AsyncLogging {
		return s.logEventAsync(ctx, event)
	}

	//
	return s.logEventSync(ctx, event)
}

// LogEvents
func (s *DefaultAuditService) LogEvents(ctx context.Context, events []*AuditEvent) error {
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		s.updateBatchStats(time.Since(start), len(events), true)
	}()

	//
	validEvents := make([]*AuditEvent, 0, len(events))
	for _, event := range events {
		if err := s.validateEvent(event); err != nil {
			s.logger.Warn("Invalid event in batch",
				zap.String("event_id", event.ID),
				zap.Error(err))
			continue
		}

		//
		if s.config.EnableFiltering && s.shouldFilter(event) {
			continue
		}

		//
		if s.config.EnableCompliance {
			s.applyComplianceRules(event)
		}

		//
		if s.config.DataClassification {
			s.classifyData(event)
		}

		//
		if s.config.EnableEncryption {
			if err := s.encryptSensitiveData(event); err != nil {
				s.logger.Warn("Failed to encrypt sensitive data",
					zap.String("event_id", event.ID),
					zap.Error(err))
			}
		}

		validEvents = append(validEvents, event)
	}

	if len(validEvents) == 0 {
		return nil
	}

	// 浽
	if err := s.repository.SaveEvents(ctx, validEvents); err != nil {
		s.updateBatchStats(time.Since(start), len(events), false)
		return fmt.Errorf("failed to save events: %w", err)
	}

	//
	if s.publisher != nil {
		if err := s.publisher.PublishEvents(ctx, validEvents); err != nil {
			s.logger.Warn("Failed to publish events",
				zap.Int("count", len(validEvents)),
				zap.Error(err))
		}
	}

	s.logger.Debug("Events logged",
		zap.Int("total", len(events)),
		zap.Int("valid", len(validEvents)))

	return nil
}

// QueryLogs
func (s *DefaultAuditService) QueryLogs(ctx context.Context, query *AuditQuery) (*AuditLogResponse, error) {
	//
	if err := s.validateQuery(query); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	//
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}
	if query.PageSize > 1000 {
		query.PageSize = 1000
	}

	//
	events, total, err := s.repository.QueryEvents(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	//
	if s.config.EnableEncryption {
		for _, event := range events {
			if err := s.decryptSensitiveData(event); err != nil {
				s.logger.Warn("Failed to decrypt sensitive data",
					zap.String("event_id", event.ID),
					zap.Error(err))
			}
		}
	}

	//
	totalPages := (total + int64(query.PageSize) - 1) / int64(query.PageSize)
	hasNext := query.Page < int(totalPages)
	hasPrev := query.Page > 1

	response := &AuditLogResponse{
		Events: events,
		Pagination: PaginationResponse{
			Page:       query.Page,
			PageSize:   query.PageSize,
			Total:      total,
			TotalPages: int(totalPages),
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}

	//
	if len(query.Aggregates) > 0 {
		aggregates, err := s.calculateAggregates(ctx, query, events)
		if err != nil {
			s.logger.Warn("Failed to calculate aggregates", zap.Error(err))
		} else {
			response.Aggregates = aggregates
		}
	}

	return response, nil
}

// GetStatistics
func (s *DefaultAuditService) GetStatistics(ctx context.Context, filter *StatisticsFilter) (*AuditStatistics, error) {
	//
	if filter == nil {
		filter = &StatisticsFilter{}
	}

	//
	if filter.StartTime == nil {
		start := time.Now().Add(-24 * time.Hour)
		filter.StartTime = &start
	}
	if filter.EndTime == nil {
		end := time.Now()
		filter.EndTime = &end
	}

	//
	stats, err := s.repository.GetStatistics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}

// ExportLogs
func (s *DefaultAuditService) ExportLogs(ctx context.Context, request *ExportRequest) (*ExportResponse, error) {
	//
	if err := s.validateExportRequest(request); err != nil {
		return nil, fmt.Errorf("invalid export request: %w", err)
	}

	// ID
	exportID := s.generateExportID()

	//
	response := &ExportResponse{
		ExportID:  exportID,
		Status:    "pending",
		Format:    request.Format,
		CreatedAt: time.Now(),
	}

	//
	go s.processExport(ctx, exportID, request, response)

	return response, nil
}

// CleanupLogs
func (s *DefaultAuditService) CleanupLogs(ctx context.Context, retentionPolicy *RetentionPolicy) (int64, error) {
	if retentionPolicy == nil {
		retentionPolicy = &s.config.RetentionPolicy
	}

	//
	cutoffTime := time.Now().Add(-retentionPolicy.DefaultRetention)

	// 鵵
	if retentionPolicy.ArchiveEnabled {
		archiveTime := time.Now().Add(-retentionPolicy.ArchiveAfter)
		archived, err := s.repository.ArchiveEvents(ctx, archiveTime, retentionPolicy.BatchSize)
		if err != nil {
			s.logger.Error("Failed to archive events", zap.Error(err))
		} else {
			s.logger.Info("Events archived", zap.Int64("count", archived))
		}
	}

	//
	deleted, err := s.repository.DeleteExpiredEvents(ctx, cutoffTime, retentionPolicy.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired events: %w", err)
	}

	s.logger.Info("Expired events deleted",
		zap.Int64("count", deleted),
		zap.Time("cutoff_time", cutoffTime))

	return deleted, nil
}

// HealthCheck
func (s *DefaultAuditService) HealthCheck(ctx context.Context) error {
	//
	if err := s.repository.HealthCheck(ctx); err != nil {
		return fmt.Errorf("repository health check failed: %w", err)
	}

	//
	if s.publisher != nil {
		if err := s.publisher.HealthCheck(ctx); err != nil {
			return fmt.Errorf("publisher health check failed: %w", err)
		}
	}

	//
	s.mu.RLock()
	bufferSize := len(s.eventBuffer)
	s.mu.RUnlock()

	if bufferSize > s.config.BufferSize*90/100 {
		return fmt.Errorf("event buffer nearly full: %d/%d", bufferSize, s.config.BufferSize)
	}

	return nil
}

// GetStats
func (s *DefaultAuditService) GetStats() AuditServiceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := s.stats
	stats.BufferedEvents = len(s.eventBuffer)
	stats.Uptime = time.Since(stats.StartTime)

	if stats.TotalEvents > 0 {
		stats.ErrorRate = float64(stats.FailedEvents) / float64(stats.TotalEvents) * 100
	}

	return stats
}

// Stop
func (s *DefaultAuditService) Stop() error {
	close(s.stopCh)

	if s.flushTicker != nil {
		s.flushTicker.Stop()
	}

	//
	if err := s.flushBuffer(context.Background()); err != nil {
		s.logger.Error("Failed to flush buffer on stop", zap.Error(err))
		return err
	}

	s.logger.Info("Audit service stopped")
	return nil
}

//

// startBackgroundTasks
func (s *DefaultAuditService) startBackgroundTasks() {
	//
	s.flushTicker = time.NewTicker(s.config.FlushInterval)
	go s.flushWorker()

	//
	if s.config.EnableMetrics {
		go s.metricsWorker()
	}

	//
	if s.config.EnableHealthCheck {
		go s.healthCheckWorker()
	}
}

// flushWorker
func (s *DefaultAuditService) flushWorker() {
	for {
		select {
		case <-s.flushTicker.C:
			if err := s.flushBuffer(context.Background()); err != nil {
				s.logger.Error("Failed to flush buffer", zap.Error(err))
			}
		case <-s.stopCh:
			return
		}
	}
}

// metricsWorker
func (s *DefaultAuditService) metricsWorker() {
	ticker := time.NewTicker(s.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.collectMetrics()
		case <-s.stopCh:
			return
		}
	}
}

// healthCheckWorker 鹤
func (s *DefaultAuditService) healthCheckWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.HealthCheck(context.Background()); err != nil {
				s.logger.Warn("Health check failed", zap.Error(err))
			}
		case <-s.stopCh:
			return
		}
	}
}

// logEventSync
func (s *DefaultAuditService) logEventSync(ctx context.Context, event *AuditEvent) error {
	// 浽
	if err := s.repository.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	//
	if s.publisher != nil {
		if err := s.publisher.PublishEvent(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event",
				zap.String("event_id", event.ID),
				zap.Error(err))
		}
	}

	return nil
}

// logEventAsync
func (s *DefaultAuditService) logEventAsync(ctx context.Context, event *AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 黺
	if len(s.eventBuffer) >= s.config.BufferSize {
		//
		if err := s.flushBufferUnsafe(ctx); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}
	}

	//
	s.eventBuffer = append(s.eventBuffer, event)

	//
	if len(s.eventBuffer) >= s.config.FlushThreshold {
		if err := s.flushBufferUnsafe(ctx); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}
	}

	return nil
}

// flushBuffer
func (s *DefaultAuditService) flushBuffer(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushBufferUnsafe(ctx)
}

// flushBufferUnsafe
func (s *DefaultAuditService) flushBufferUnsafe(ctx context.Context) error {
	if len(s.eventBuffer) == 0 {
		return nil
	}

	events := make([]*AuditEvent, len(s.eventBuffer))
	copy(events, s.eventBuffer)
	s.eventBuffer = s.eventBuffer[:0]
	s.lastFlush = time.Now()

	//
	if err := s.repository.SaveEvents(ctx, events); err != nil {
		//
		s.eventBuffer = append(events, s.eventBuffer...)
		return fmt.Errorf("failed to save events: %w", err)
	}

	//
	if s.publisher != nil {
		if err := s.publisher.PublishEvents(ctx, events); err != nil {
			s.logger.Warn("Failed to publish events",
				zap.Int("count", len(events)),
				zap.Error(err))
		}
	}

	s.logger.Debug("Buffer flushed", zap.Int("count", len(events)))
	return nil
}

// validateEvent
func (s *DefaultAuditService) validateEvent(event *AuditEvent) error {
	if event == nil {
		return &AuditError{
			Code:    "INVALID_EVENT",
			Message: "Event cannot be nil",
		}
	}

	if event.ID == "" {
		event.ID = generateEventID()
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if event.EventType == "" {
		return &AuditError{
			Code:    "MISSING_EVENT_TYPE",
			Message: "Event type is required",
		}
	}

	if event.EventAction == "" {
		return &AuditError{
			Code:    "MISSING_EVENT_ACTION",
			Message: "Event action is required",
		}
	}

	return nil
}

// validateQuery
func (s *DefaultAuditService) validateQuery(query *AuditQuery) error {
	if query == nil {
		return &AuditError{
			Code:    "INVALID_QUERY",
			Message: "Query cannot be nil",
		}
	}

	//
	if query.StartTime != nil && query.EndTime != nil {
		if query.StartTime.After(*query.EndTime) {
			return &AuditError{
				Code:    "INVALID_TIME_RANGE",
				Message: "Start time cannot be after end time",
			}
		}
	}

	return nil
}

// validateExportRequest
func (s *DefaultAuditService) validateExportRequest(request *ExportRequest) error {
	if request == nil {
		return &AuditError{
			Code:    "INVALID_EXPORT_REQUEST",
			Message: "Export request cannot be nil",
		}
	}

	validFormats := map[string]bool{
		"json": true,
		"csv":  true,
		"xlsx": true,
		"pdf":  true,
	}

	if !validFormats[request.Format] {
		return &AuditError{
			Code:    "INVALID_FORMAT",
			Message: "Invalid export format",
			Details: map[string]interface{}{"format": request.Format},
		}
	}

	return nil
}

// shouldFilter
func (s *DefaultAuditService) shouldFilter(event *AuditEvent) bool {
	//
	//
	return true
}

// applyComplianceRules
func (s *DefaultAuditService) applyComplianceRules(event *AuditEvent) {
	//
	//
}

// classifyData
func (s *DefaultAuditService) classifyData(event *AuditEvent) {
	//
	// IPPII
	if event.UserEmail != "" || event.IPAddress != "" {
		event.DataClassification = "PII"
	}
}

// encryptSensitiveData
func (s *DefaultAuditService) encryptSensitiveData(event *AuditEvent) error {
	//
	//
	return nil
}

// decryptSensitiveData
func (s *DefaultAuditService) decryptSensitiveData(event *AuditEvent) error {
	//
	//
	return nil
}

// calculateAggregates
func (s *DefaultAuditService) calculateAggregates(ctx context.Context, query *AuditQuery, events []*AuditEvent) (map[string]interface{}, error) {
	//
	//
	aggregates := make(map[string]interface{})
	aggregates["count"] = len(events)
	return aggregates, nil
}

// generateExportID ID
func (s *DefaultAuditService) generateExportID() string {
	// ID
	return fmt.Sprintf("export_%d", time.Now().Unix())
}

// processExport
func (s *DefaultAuditService) processExport(ctx context.Context, exportID string, request *ExportRequest, response *ExportResponse) {
	//
	//
	response.Status = "success"
	response.ExportID = exportID
}

// updateStats
func (s *DefaultAuditService) updateStats(latency time.Duration, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalEvents++
	if success {
		s.stats.SuccessfulEvents++
	} else {
		s.stats.FailedEvents++
	}

	//
	if s.stats.TotalEvents == 1 {
		s.stats.AverageLatency = latency
	} else {
		s.stats.AverageLatency = (s.stats.AverageLatency*time.Duration(s.stats.TotalEvents-1) + latency) / time.Duration(s.stats.TotalEvents)
	}
}

// updateBatchStats
func (s *DefaultAuditService) updateBatchStats(latency time.Duration, count int, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalEvents += int64(count)
	if success {
		s.stats.SuccessfulEvents += int64(count)
	} else {
		s.stats.FailedEvents += int64(count)
	}
}

// collectMetrics
func (s *DefaultAuditService) collectMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	//
	if s.stats.TotalEvents > 0 {
		duration := time.Since(s.stats.StartTime).Seconds()
		s.stats.EventsPerSecond = float64(s.stats.TotalEvents) / duration
	}

	s.stats.LastFlushTime = s.lastFlush
}
