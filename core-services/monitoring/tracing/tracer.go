package tracing

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// Tracer 分布式追踪器
type Tracer struct {
	config       *TracerConfig
	spans        map[string]*Span
	spansMutex   sync.RWMutex
	exporters    []SpanExporter
	sampler      Sampler
	idGenerator  IDGenerator
	processor    SpanProcessor
	running      bool
	stopCh       chan struct{}
}

// TracerConfig 追踪器配?
type TracerConfig struct {
	ServiceName     string        `json:"service_name" yaml:"service_name"`
	ServiceVersion  string        `json:"service_version" yaml:"service_version"`
	Environment     string        `json:"environment" yaml:"environment"`
	
	// 采样配置
	SamplingRate    float64       `json:"sampling_rate" yaml:"sampling_rate"`
	SamplingType    string        `json:"sampling_type" yaml:"sampling_type"` // always, never, probabilistic, rate_limiting
	
	// 批处理配?
	BatchTimeout    time.Duration `json:"batch_timeout" yaml:"batch_timeout"`
	BatchSize       int           `json:"batch_size" yaml:"batch_size"`
	MaxQueueSize    int           `json:"max_queue_size" yaml:"max_queue_size"`
	
	// 资源限制
	MaxSpansPerTrace int          `json:"max_spans_per_trace" yaml:"max_spans_per_trace"`
	MaxTraceAge      time.Duration `json:"max_trace_age" yaml:"max_trace_age"`
	
	// 导出器配?
	Exporters       []ExporterConfig `json:"exporters" yaml:"exporters"`
}

// ExporterConfig 导出器配?
type ExporterConfig struct {
	Type     string                 `json:"type" yaml:"type"` // jaeger, zipkin, otlp, console
	Endpoint string                 `json:"endpoint" yaml:"endpoint"`
	Headers  map[string]string      `json:"headers" yaml:"headers"`
	Timeout  time.Duration          `json:"timeout" yaml:"timeout"`
	Options  map[string]interface{} `json:"options" yaml:"options"`
}

// NewTracer 创建追踪?
func NewTracer(config *TracerConfig) *Tracer {
	if config == nil {
		config = &TracerConfig{
			ServiceName:      "unknown-service",
			ServiceVersion:   "1.0.0",
			Environment:      "development",
			SamplingRate:     1.0,
			SamplingType:     "probabilistic",
			BatchTimeout:     5 * time.Second,
			BatchSize:        100,
			MaxQueueSize:     1000,
			MaxSpansPerTrace: 1000,
			MaxTraceAge:      1 * time.Hour,
		}
	}
	
	tracer := &Tracer{
		config:      config,
		spans:       make(map[string]*Span),
		exporters:   make([]SpanExporter, 0),
		idGenerator: NewRandomIDGenerator(),
		stopCh:      make(chan struct{}),
	}
	
	// 初始化采样器
	tracer.sampler = tracer.createSampler()
	
	// 初始化处理器
	tracer.processor = NewBatchSpanProcessor(config.BatchSize, config.BatchTimeout, config.MaxQueueSize)
	
	// 初始化导出器
	tracer.initializeExporters()
	
	return tracer
}

// createSampler 创建采样?
func (t *Tracer) createSampler() Sampler {
	switch t.config.SamplingType {
	case "always":
		return NewAlwaysSampler()
	case "never":
		return NewNeverSampler()
	case "probabilistic":
		return NewProbabilisticSampler(t.config.SamplingRate)
	case "rate_limiting":
		return NewRateLimitingSampler(int(t.config.SamplingRate))
	default:
		return NewProbabilisticSampler(t.config.SamplingRate)
	}
}

// initializeExporters 初始化导出器
func (t *Tracer) initializeExporters() {
	for _, exporterConfig := range t.config.Exporters {
		exporter, err := t.createExporter(exporterConfig)
		if err != nil {
			fmt.Printf("Failed to create exporter %s: %v\n", exporterConfig.Type, err)
			continue
		}
		t.exporters = append(t.exporters, exporter)
	}
}

// createExporter 创建导出?
func (t *Tracer) createExporter(config ExporterConfig) (SpanExporter, error) {
	switch config.Type {
	case "jaeger":
		return NewJaegerExporter(config)
	case "zipkin":
		return NewZipkinExporter(config)
	case "otlp":
		return NewOTLPExporter(config)
	case "console":
		return NewConsoleExporter(config)
	default:
		return nil, fmt.Errorf("unknown exporter type: %s", config.Type)
	}
}

// Start 启动追踪?
func (t *Tracer) Start() error {
	if t.running {
		return fmt.Errorf("tracer is already running")
	}
	
	// 启动处理?
	if err := t.processor.Start(); err != nil {
		return fmt.Errorf("failed to start span processor: %w", err)
	}
	
	// 启动导出?
	for _, exporter := range t.exporters {
		if err := exporter.Start(); err != nil {
			fmt.Printf("Failed to start exporter: %v\n", err)
		}
	}
	
	t.running = true
	
	// 启动清理协程
	go t.cleanupLoop()
	
	return nil
}

// Stop 停止追踪?
func (t *Tracer) Stop() error {
	if !t.running {
		return nil
	}
	
	t.running = false
	close(t.stopCh)
	
	// 停止处理?
	if err := t.processor.Stop(); err != nil {
		fmt.Printf("Failed to stop span processor: %v\n", err)
	}
	
	// 停止导出?
	for _, exporter := range t.exporters {
		if err := exporter.Stop(); err != nil {
			fmt.Printf("Failed to stop exporter: %v\n", err)
		}
	}
	
	return nil
}

// StartSpan 开始一个新的span
func (t *Tracer) StartSpan(ctx context.Context, operationName string, opts ...SpanOption) (*Span, context.Context) {
	// 检查采?
	if !t.sampler.ShouldSample(ctx, operationName) {
		// 返回一个no-op span
		return NewNoOpSpan(), ctx
	}
	
	// 创建span
	span := t.createSpan(ctx, operationName, opts...)
	
	// 存储span
	t.spansMutex.Lock()
	t.spans[span.SpanID] = span
	t.spansMutex.Unlock()
	
	// 将span添加到context
	newCtx := ContextWithSpan(ctx, span)
	
	return span, newCtx
}

// createSpan 创建span
func (t *Tracer) createSpan(ctx context.Context, operationName string, opts ...SpanOption) *Span {
	// 获取父span
	parentSpan := SpanFromContext(ctx)
	
	// 生成ID
	var traceID, spanID string
	if parentSpan != nil && !parentSpan.IsNoOp() {
		traceID = parentSpan.TraceID
		spanID = t.idGenerator.GenerateSpanID()
	} else {
		traceID = t.idGenerator.GenerateTraceID()
		spanID = t.idGenerator.GenerateSpanID()
	}
	
	// 创建span
	span := &Span{
		TraceID:       traceID,
		SpanID:        spanID,
		ParentSpanID:  "",
		OperationName: operationName,
		StartTime:     time.Now(),
		Tags:          make(map[string]interface{}),
		Logs:          make([]*LogEntry, 0),
		Status:        SpanStatusOK,
		tracer:        t,
	}
	
	// 设置父span ID
	if parentSpan != nil && !parentSpan.IsNoOp() {
		span.ParentSpanID = parentSpan.SpanID
	}
	
	// 设置服务信息
	span.SetTag("service.name", t.config.ServiceName)
	span.SetTag("service.version", t.config.ServiceVersion)
	span.SetTag("environment", t.config.Environment)
	
	// 应用选项
	for _, opt := range opts {
		opt(span)
	}
	
	return span
}

// FinishSpan 完成span
func (t *Tracer) FinishSpan(span *Span) {
	if span == nil || span.IsNoOp() {
		return
	}
	
	// 设置结束时间
	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)
	
	// 从活跃spans中移?
	t.spansMutex.Lock()
	delete(t.spans, span.SpanID)
	t.spansMutex.Unlock()
	
	// 发送到处理?
	t.processor.OnEnd(span)
}

// GetActiveSpans 获取活跃的spans
func (t *Tracer) GetActiveSpans() []*Span {
	t.spansMutex.RLock()
	defer t.spansMutex.RUnlock()
	
	spans := make([]*Span, 0, len(t.spans))
	for _, span := range t.spans {
		spans = append(spans, span)
	}
	
	return spans
}

// GetStats 获取统计信息
func (t *Tracer) GetStats() *TracerStats {
	t.spansMutex.RLock()
	activeSpans := len(t.spans)
	t.spansMutex.RUnlock()
	
	return &TracerStats{
		ActiveSpans:   activeSpans,
		SamplingRate:  t.config.SamplingRate,
		ServiceName:   t.config.ServiceName,
		Environment:   t.config.Environment,
		ExporterCount: len(t.exporters),
	}
}

// cleanupLoop 清理循环
func (t *Tracer) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			t.cleanupOldSpans()
		case <-t.stopCh:
			return
		}
	}
}

// cleanupOldSpans 清理旧的spans
func (t *Tracer) cleanupOldSpans() {
	now := time.Now()
	maxAge := t.config.MaxTraceAge
	
	t.spansMutex.Lock()
	defer t.spansMutex.Unlock()
	
	for spanID, span := range t.spans {
		if now.Sub(span.StartTime) > maxAge {
			// 强制完成旧的span
			span.EndTime = now
			span.Duration = span.EndTime.Sub(span.StartTime)
			span.SetTag("timeout", true)
			
			// 发送到处理?
			t.processor.OnEnd(span)
			
			// 从活跃spans中移?
			delete(t.spans, spanID)
		}
	}
}

// TracerStats 追踪器统计信?
type TracerStats struct {
	ActiveSpans   int     `json:"active_spans"`
	SamplingRate  float64 `json:"sampling_rate"`
	ServiceName   string  `json:"service_name"`
	Environment   string  `json:"environment"`
	ExporterCount int     `json:"exporter_count"`
}

// Span span结构
type Span struct {
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	ParentSpanID  string                 `json:"parent_span_id,omitempty"`
	OperationName string                 `json:"operation_name"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Duration      time.Duration          `json:"duration"`
	Tags          map[string]interface{} `json:"tags"`
	Logs          []*LogEntry            `json:"logs"`
	Status        SpanStatus             `json:"status"`
	
	tracer        *Tracer
	mutex         sync.RWMutex
}

// SpanStatus span状?
type SpanStatus int

const (
	SpanStatusOK SpanStatus = iota
	SpanStatusError
	SpanStatusTimeout
	SpanStatusCancelled
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

// SetTag 设置标签
func (s *Span) SetTag(key string, value interface{}) {
	if s.IsNoOp() {
		return
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.Tags[key] = value
}

// SetStatus 设置状?
func (s *Span) SetStatus(status SpanStatus) {
	if s.IsNoOp() {
		return
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.Status = status
}

// LogFields 记录字段
func (s *Span) LogFields(fields map[string]interface{}) {
	if s.IsNoOp() {
		return
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	entry := &LogEntry{
		Timestamp: time.Now(),
		Fields:    fields,
	}
	
	s.Logs = append(s.Logs, entry)
}

// Log 记录日志
func (s *Span) Log(message string) {
	s.LogFields(map[string]interface{}{
		"message": message,
	})
}

// Finish 完成span
func (s *Span) Finish() {
	if s.tracer != nil {
		s.tracer.FinishSpan(s)
	}
}

// IsNoOp 检查是否为no-op span
func (s *Span) IsNoOp() bool {
	return s.tracer == nil
}

// SpanOption span选项
type SpanOption func(*Span)

// WithTag 设置标签选项
func WithTag(key string, value interface{}) SpanOption {
	return func(span *Span) {
		span.SetTag(key, value)
	}
}

// WithStartTime 设置开始时间选项
func WithStartTime(startTime time.Time) SpanOption {
	return func(span *Span) {
		span.StartTime = startTime
	}
}

// NoOpSpan no-op span
type NoOpSpan struct{}

// NewNoOpSpan 创建no-op span
func NewNoOpSpan() *Span {
	return &Span{}
}

// Context相关函数

type spanContextKey struct{}

// ContextWithSpan 将span添加到context
func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

// SpanFromContext 从context获取span
func SpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanContextKey{}).(*Span); ok {
		return span
	}
	return nil
}

// IDGenerator ID生成器接?
type IDGenerator interface {
	GenerateTraceID() string
	GenerateSpanID() string
}

// RandomIDGenerator 随机ID生成?
type RandomIDGenerator struct {
	rand *rand.Rand
	mutex sync.Mutex
}

// NewRandomIDGenerator 创建随机ID生成?
func NewRandomIDGenerator() *RandomIDGenerator {
	return &RandomIDGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateTraceID 生成trace ID
func (g *RandomIDGenerator) GenerateTraceID() string {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	// 生成128位ID
	high := g.rand.Uint64()
	low := g.rand.Uint64()
	
	return fmt.Sprintf("%016x%016x", high, low)
}

// GenerateSpanID 生成span ID
func (g *RandomIDGenerator) GenerateSpanID() string {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	// 生成64位ID
	id := g.rand.Uint64()
	
	return fmt.Sprintf("%016x", id)
}

// Sampler 采样器接?
type Sampler interface {
	ShouldSample(ctx context.Context, operationName string) bool
}

// AlwaysSampler 总是采样
type AlwaysSampler struct{}

// NewAlwaysSampler 创建总是采样?
func NewAlwaysSampler() *AlwaysSampler {
	return &AlwaysSampler{}
}

// ShouldSample 是否应该采样
func (s *AlwaysSampler) ShouldSample(ctx context.Context, operationName string) bool {
	return true
}

// NeverSampler 从不采样
type NeverSampler struct{}

// NewNeverSampler 创建从不采样?
func NewNeverSampler() *NeverSampler {
	return &NeverSampler{}
}

// ShouldSample 是否应该采样
func (s *NeverSampler) ShouldSample(ctx context.Context, operationName string) bool {
	return false
}

// ProbabilisticSampler 概率采样?
type ProbabilisticSampler struct {
	rate  float64
	rand  *rand.Rand
	mutex sync.Mutex
}

// NewProbabilisticSampler 创建概率采样?
func NewProbabilisticSampler(rate float64) *ProbabilisticSampler {
	if rate < 0 {
		rate = 0
	} else if rate > 1 {
		rate = 1
	}
	
	return &ProbabilisticSampler{
		rate: rate,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldSample 是否应该采样
func (s *ProbabilisticSampler) ShouldSample(ctx context.Context, operationName string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	return s.rand.Float64() < s.rate
}

// RateLimitingSampler 限流采样?
type RateLimitingSampler struct {
	maxTracesPerSecond int
	tokens             int
	lastRefill         time.Time
	mutex              sync.Mutex
}

// NewRateLimitingSampler 创建限流采样?
func NewRateLimitingSampler(maxTracesPerSecond int) *RateLimitingSampler {
	return &RateLimitingSampler{
		maxTracesPerSecond: maxTracesPerSecond,
		tokens:             maxTracesPerSecond,
		lastRefill:         time.Now(),
	}
}

// ShouldSample 是否应该采样
func (s *RateLimitingSampler) ShouldSample(ctx context.Context, operationName string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(s.lastRefill)
	
	// 补充令牌
	tokensToAdd := int(elapsed.Seconds()) * s.maxTracesPerSecond
	if tokensToAdd > 0 {
		s.tokens += tokensToAdd
		if s.tokens > s.maxTracesPerSecond {
			s.tokens = s.maxTracesPerSecond
		}
		s.lastRefill = now
	}
	
	// 检查是否有可用令牌
	if s.tokens > 0 {
		s.tokens--
		return true
	}
	
	return false
}

