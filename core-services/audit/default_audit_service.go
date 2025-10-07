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
	
	// 缓存和状态
	mu           sync.RWMutex
	eventBuffer  []*AuditEvent
	bufferSize   int
	lastFlush    time.Time
	
	// 统计信息
	stats        AuditServiceStats
	
	// 停止信号
	stopCh       chan struct{}
	flushTicker  *time.Ticker
}

// DefaultAuditServiceConfig 默认审计服务配置
type DefaultAuditServiceConfig struct {
	// 缓冲配置
	BufferSize     int           `json:"buffer_size"`
	FlushInterval  time.Duration `json:"flush_interval"`
	FlushThreshold int           `json:"flush_threshold"`
	
	// 异步处理
	AsyncLogging   bool `json:"async_logging"`
	WorkerCount    int  `json:"worker_count"`
	QueueSize      int  `json:"queue_size"`
	
	// 重试配置
	RetryEnabled   bool          `json:"retry_enabled"`
	MaxRetries     int           `json:"max_retries"`
	RetryInterval  time.Duration `json:"retry_interval"`
	
	// 过滤配置
	EnableFiltering bool     `json:"enable_filtering"`
	FilterRules     []string `json:"filter_rules"`
	
	// 安全配置
	EnableEncryption bool   `json:"enable_encryption"`
	EncryptionKey    string `json:"encryption_key"`
	
	// 合规配置
	EnableCompliance   bool     `json:"enable_compliance"`
	ComplianceRules    []string `json:"compliance_rules"`
	DataClassification bool     `json:"data_classification"`
	
	// 性能配置
	EnableCompression bool    `json:"enable_compression"`
	CompressionLevel  int     `json:"compression_level"`
	EnableBatching    bool    `json:"enable_batching"`
	BatchSize         int     `json:"batch_size"`
	
	// 监控配置
	EnableMetrics     bool          `json:"enable_metrics"`
	MetricsInterval   time.Duration `json:"metrics_interval"`
	EnableHealthCheck bool          `json:"enable_health_check"`
	
	// 存储配置
	RetentionPolicy   RetentionPolicy `json:"retention_policy"`
	ArchiveEnabled    bool            `json:"archive_enabled"`
	
	// 告警配置
	EnableAlerting    bool     `json:"enable_alerting"`
	AlertRules        []string `json:"alert_rules"`
	AlertThresholds   map[string]float64 `json:"alert_thresholds"`
}

// AuditServiceStats 审计服务统计
type AuditServiceStats struct {
	TotalEvents       int64     `json:"total_events"`
	SuccessfulEvents  int64     `json:"successful_events"`
	FailedEvents      int64     `json:"failed_events"`
	BufferedEvents    int       `json:"buffered_events"`
	LastFlushTime     time.Time `json:"last_flush_time"`
	AverageLatency    time.Duration `json:"average_latency"`
	EventsPerSecond   float64   `json:"events_per_second"`
	ErrorRate         float64   `json:"error_rate"`
	StartTime         time.Time `json:"start_time"`
	Uptime            time.Duration `json:"uptime"`
}

// NewDefaultAuditService 创建默认审计服务
func NewDefaultAuditService(
	repository AuditRepository,
	publisher AuditEventPublisher,
	config DefaultAuditServiceConfig,
	logger *zap.Logger,
) *DefaultAuditService {
	// 设置默认值
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

	// 启动后台任务
	service.startBackgroundTasks()

	return service
}

// LogEvent 记录审计事件
func (s *DefaultAuditService) LogEvent(ctx context.Context, event *AuditEvent) error {
	start := time.Now()
	defer func() {
		s.updateStats(time.Since(start), true)
	}()

	// 验证事件
	if err := s.validateEvent(event); err != nil {
		s.updateStats(time.Since(start), false)
		return fmt.Errorf("invalid audit event: %w", err)
	}

	// 应用过滤规则
	if s.config.EnableFiltering && s.shouldFilter(event) {
		s.logger.Debug("Event filtered", zap.String("event_id", event.ID))
		return nil
	}

	// 应用合规规则
	if s.config.EnableCompliance {
		s.applyComplianceRules(event)
	}

	// 数据分类
	if s.config.DataClassification {
		s.classifyData(event)
	}

	// 加密敏感数据
	if s.config.EnableEncryption {
		if err := s.encryptSensitiveData(event); err != nil {
			s.logger.Warn("Failed to encrypt sensitive data",
				zap.String("event_id", event.ID),
				zap.Error(err))
		}
	}

	// 异步处理
	if s.config.AsyncLogging {
		return s.logEventAsync(ctx, event)
	}

	// 同步处理
	return s.logEventSync(ctx, event)
}

// LogEvents 批量记录审计事件
func (s *DefaultAuditService) LogEvents(ctx context.Context, events []*AuditEvent) error {
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		s.updateBatchStats(time.Since(start), len(events), true)
	}()

	// 验证和处理事件
	validEvents := make([]*AuditEvent, 0, len(events))
	for _, event := range events {
		if err := s.validateEvent(event); err != nil {
			s.logger.Warn("Invalid event in batch",
				zap.String("event_id", event.ID),
				zap.Error(err))
			continue
		}

		// 应用过滤规则
		if s.config.EnableFiltering && s.shouldFilter(event) {
			continue
		}

		// 应用合规规则
		if s.config.EnableCompliance {
			s.applyComplianceRules(event)
		}

		// 数据分类
		if s.config.DataClassification {
			s.classifyData(event)
		}

		// 加密敏感数据
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

	// 保存到数据库
	if err := s.repository.SaveEvents(ctx, validEvents); err != nil {
		s.updateBatchStats(time.Since(start), len(events), false)
		return fmt.Errorf("failed to save events: %w", err)
	}

	// 发布事件
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

// QueryLogs 查询审计日志
func (s *DefaultAuditService) QueryLogs(ctx context.Context, query *AuditQuery) (*AuditLogResponse, error) {
	// 验证查询参数
	if err := s.validateQuery(query); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}
	if query.PageSize > 1000 {
		query.PageSize = 1000
	}

	// 查询事件
	events, total, err := s.repository.QueryEvents(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	// 解密敏感数据
	if s.config.EnableEncryption {
		for _, event := range events {
			if err := s.decryptSensitiveData(event); err != nil {
				s.logger.Warn("Failed to decrypt sensitive data",
					zap.String("event_id", event.ID),
					zap.Error(err))
			}
		}
	}

	// 计算分页信息
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

	// 计算聚合数据
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

// GetStatistics 获取审计统计
func (s *DefaultAuditService) GetStatistics(ctx context.Context, filter *StatisticsFilter) (*AuditStatistics, error) {
	// 验证过滤器
	if filter == nil {
		filter = &StatisticsFilter{}
	}

	// 设置默认时间范围
	if filter.StartTime == nil {
		start := time.Now().Add(-24 * time.Hour)
		filter.StartTime = &start
	}
	if filter.EndTime == nil {
		end := time.Now()
		filter.EndTime = &end
	}

	// 获取统计信息
	stats, err := s.repository.GetStatistics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}

// ExportLogs 导出审计日志
func (s *DefaultAuditService) ExportLogs(ctx context.Context, request *ExportRequest) (*ExportResponse, error) {
	// 验证导出请求
	if err := s.validateExportRequest(request); err != nil {
		return nil, fmt.Errorf("invalid export request: %w", err)
	}

	// 生成导出ID
	exportID := s.generateExportID()

	// 创建导出响应
	response := &ExportResponse{
		ExportID:  exportID,
		Status:    "pending",
		Format:    request.Format,
		CreatedAt: time.Now(),
	}

	// 异步处理导出
	go s.processExport(ctx, exportID, request, response)

	return response, nil
}

// CleanupLogs 清理过期日志
func (s *DefaultAuditService) CleanupLogs(ctx context.Context, retentionPolicy *RetentionPolicy) error {
	if retentionPolicy == nil {
		retentionPolicy = &s.config.RetentionPolicy
	}

	// 计算清理时间点
	cutoffTime := time.Now().Add(-retentionPolicy.DefaultRetention)

	// 归档事件（如果启用）
	if retentionPolicy.ArchiveEnabled {
		archiveTime := time.Now().Add(-retentionPolicy.ArchiveAfter)
		archived, err := s.repository.ArchiveEvents(ctx, archiveTime, retentionPolicy.BatchSize)
		if err != nil {
			s.logger.Error("Failed to archive events", zap.Error(err))
		} else {
			s.logger.Info("Events archived", zap.Int64("count", archived))
		}
	}

	// 删除过期事件
	deleted, err := s.repository.DeleteExpiredEvents(ctx, cutoffTime, retentionPolicy.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to delete expired events: %w", err)
	}

	s.logger.Info("Expired events deleted",
		zap.Int64("count", deleted),
		zap.Time("cutoff_time", cutoffTime))

	return nil
}

// HealthCheck 健康检查
func (s *DefaultAuditService) HealthCheck(ctx context.Context) error {
	// 检查数据库连接
	if err := s.repository.HealthCheck(ctx); err != nil {
		return fmt.Errorf("repository health check failed: %w", err)
	}

	// 检查事件发布器
	if s.publisher != nil {
		if err := s.publisher.HealthCheck(ctx); err != nil {
			return fmt.Errorf("publisher health check failed: %w", err)
		}
	}

	// 检查服务状态
	s.mu.RLock()
	bufferSize := len(s.eventBuffer)
	s.mu.RUnlock()

	if bufferSize > s.config.BufferSize*90/100 {
		return fmt.Errorf("event buffer nearly full: %d/%d", bufferSize, s.config.BufferSize)
	}

	return nil
}

// GetStats 获取服务统计信息
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

// Stop 停止服务
func (s *DefaultAuditService) Stop() error {
	close(s.stopCh)
	
	if s.flushTicker != nil {
		s.flushTicker.Stop()
	}

	// 刷新缓冲区
	if err := s.flushBuffer(context.Background()); err != nil {
		s.logger.Error("Failed to flush buffer on stop", zap.Error(err))
		return err
	}

	s.logger.Info("Audit service stopped")
	return nil
}

// 私有方法

// startBackgroundTasks 启动后台任务
func (s *DefaultAuditService) startBackgroundTasks() {
	// 启动定时刷新任务
	s.flushTicker = time.NewTicker(s.config.FlushInterval)
	go s.flushWorker()

	// 启动指标收集任务
	if s.config.EnableMetrics {
		go s.metricsWorker()
	}

	// 启动健康检查任务
	if s.config.EnableHealthCheck {
		go s.healthCheckWorker()
	}
}

// flushWorker 刷新工作器
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

// metricsWorker 指标工作器
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

// healthCheckWorker 健康检查工作器
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

// logEventSync 同步记录事件
func (s *DefaultAuditService) logEventSync(ctx context.Context, event *AuditEvent) error {
	// 保存到数据库
	if err := s.repository.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	// 发布事件
	if s.publisher != nil {
		if err := s.publisher.PublishEvent(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event",
				zap.String("event_id", event.ID),
				zap.Error(err))
		}
	}

	return nil
}

// logEventAsync 异步记录事件
func (s *DefaultAuditService) logEventAsync(ctx context.Context, event *AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查缓冲区是否已满
	if len(s.eventBuffer) >= s.config.BufferSize {
		// 强制刷新缓冲区
		if err := s.flushBufferUnsafe(ctx); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}
	}

	// 添加到缓冲区
	s.eventBuffer = append(s.eventBuffer, event)

	// 检查是否需要刷新
	if len(s.eventBuffer) >= s.config.FlushThreshold {
		if err := s.flushBufferUnsafe(ctx); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}
	}

	return nil
}

// flushBuffer 刷新缓冲区
func (s *DefaultAuditService) flushBuffer(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushBufferUnsafe(ctx)
}

// flushBufferUnsafe 刷新缓冲区（不加锁）
func (s *DefaultAuditService) flushBufferUnsafe(ctx context.Context) error {
	if len(s.eventBuffer) == 0 {
		return nil
	}

	events := make([]*AuditEvent, len(s.eventBuffer))
	copy(events, s.eventBuffer)
	s.eventBuffer = s.eventBuffer[:0]
	s.lastFlush = time.Now()

	// 批量保存
	if err := s.repository.SaveEvents(ctx, events); err != nil {
		// 如果保存失败，将事件重新加入缓冲区
		s.eventBuffer = append(events, s.eventBuffer...)
		return fmt.Errorf("failed to save events: %w", err)
	}

	// 发布事件
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

// validateEvent 验证事件
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

// validateQuery 验证查询
func (s *DefaultAuditService) validateQuery(query *AuditQuery) error {
	if query == nil {
		return &AuditError{
			Code:    "INVALID_QUERY",
			Message: "Query cannot be nil",
		}
	}

	// 验证时间范围
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

// validateExportRequest 验证导出请求
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

// shouldFilter 检查是否应该过滤事件
func (s *DefaultAuditService) shouldFilter(event *AuditEvent) bool {
	// 实现过滤逻辑
	// 这里简化实现
	return false
}

// applyComplianceRules 应用合规规则
func (s *DefaultAuditService) applyComplianceRules(event *AuditEvent) {
	// 实现合规规则逻辑
	// 这里简化实现
}

// classifyData 数据分类
func (s *DefaultAuditService) classifyData(event *AuditEvent) {
	// 实现数据分类逻辑
	// 这里简化实现
	if event.UserEmail != "" || event.IPAddress != "" {
		event.DataClassification = "PII"
	}
}

// encryptSensitiveData 加密敏感数据
func (s *DefaultAuditService) encryptSensitiveData(event *AuditEvent) error {
	// 实现加密逻辑
	// 这里简化实现
	return nil
}

// decryptSensitiveData 解密敏感数据
func (s *DefaultAuditService) decryptSensitiveData(event *AuditEvent) error {
	// 实现解密逻辑
	// 这里简化实现
	return nil
}

// calculateAggregates 计算聚合数据
func (s *DefaultAuditService) calculateAggregates(ctx context.Context, query *AuditQuery, events []*AuditEvent) (map[string]interface{}, error) {
	// 实现聚合计算逻辑
	// 这里简化实现
	aggregates := make(map[string]interface{})
	aggregates["count"] = len(events)
	return aggregates, nil
}

// generateExportID 生成导出ID
func (s *DefaultAuditService) generateExportID() string {
	// 实现导出ID生成逻辑
	return fmt.Sprintf("export_%d", time.Now().Unix())
}

// processExport 处理导出
func (s *DefaultAuditService) processExport(ctx context.Context, exportID string, request *ExportRequest, response *ExportResponse) {
	// 实现导出处理逻辑
	// 这里简化实现
}

// updateStats 更新统计信息
func (s *DefaultAuditService) updateStats(latency time.Duration, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalEvents++
	if success {
		s.stats.SuccessfulEvents++
	} else {
		s.stats.FailedEvents++
	}

	// 更新平均延迟
	if s.stats.TotalEvents == 1 {
		s.stats.AverageLatency = latency
	} else {
		s.stats.AverageLatency = (s.stats.AverageLatency*time.Duration(s.stats.TotalEvents-1) + latency) / time.Duration(s.stats.TotalEvents)
	}
}

// updateBatchStats 更新批量统计信息
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

// collectMetrics 收集指标
func (s *DefaultAuditService) collectMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 计算每秒事件数
	if s.stats.TotalEvents > 0 {
		duration := time.Since(s.stats.StartTime).Seconds()
		s.stats.EventsPerSecond = float64(s.stats.TotalEvents) / duration
	}

	s.stats.LastFlushTime = s.lastFlush
}