package tracing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SpanExporter span导出器接口
type SpanExporter interface {
	Start() error
	Stop() error
	Export(spans []*Span) error
	GetStats() *ExporterStats
}

// ExporterStats 导出器统计信息
type ExporterStats struct {
	ExportedSpans int64         `json:"exported_spans"`
	FailedSpans   int64         `json:"failed_spans"`
	LastExport    time.Time     `json:"last_export"`
	Errors        []string      `json:"errors"`
	Latency       time.Duration `json:"latency"`
}

// BaseExporter 基础导出器
type BaseExporter struct {
	config ExporterConfig
	stats  *ExporterStats
	mutex  sync.RWMutex
	client *http.Client
}

// NewBaseExporter 创建基础导出器
func NewBaseExporter(config ExporterConfig) *BaseExporter {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	
	return &BaseExporter{
		config: config,
		stats: &ExporterStats{
			Errors: make([]string, 0),
		},
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetStats 获取统计信息
func (be *BaseExporter) GetStats() *ExporterStats {
	be.mutex.RLock()
	defer be.mutex.RUnlock()
	
	// 复制统计信息
	stats := *be.stats
	errors := make([]string, len(be.stats.Errors))
	copy(errors, be.stats.Errors)
	stats.Errors = errors
	
	return &stats
}

// recordSuccess 记录成功
func (be *BaseExporter) recordSuccess(spanCount int, latency time.Duration) {
	be.mutex.Lock()
	defer be.mutex.Unlock()
	
	be.stats.ExportedSpans += int64(spanCount)
	be.stats.LastExport = time.Now()
	be.stats.Latency = latency
}

// recordError 记录错误
func (be *BaseExporter) recordError(spanCount int, err error) {
	be.mutex.Lock()
	defer be.mutex.Unlock()
	
	be.stats.FailedSpans += int64(spanCount)
	
	// 保留最近的10个错误
	if len(be.stats.Errors) >= 10 {
		be.stats.Errors = be.stats.Errors[1:]
	}
	be.stats.Errors = append(be.stats.Errors, fmt.Sprintf("%s: %v", time.Now().Format("15:04:05"), err))
}

// JaegerExporter Jaeger导出器
type JaegerExporter struct {
	*BaseExporter
	endpoint string
}

// NewJaegerExporter 创建Jaeger导出器
func NewJaegerExporter(config ExporterConfig) (*JaegerExporter, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("jaeger endpoint is required")
	}
	
	return &JaegerExporter{
		BaseExporter: NewBaseExporter(config),
		endpoint:     config.Endpoint,
	}, nil
}

// Start 启动导出器
func (je *JaegerExporter) Start() error {
	return nil
}

// Stop 停止导出器
func (je *JaegerExporter) Stop() error {
	return nil
}

// Export 导出spans
func (je *JaegerExporter) Export(spans []*Span) error {
	if len(spans) == 0 {
		return nil
	}
	
	start := time.Now()
	
	// 转换为Jaeger格式
	jaegerSpans := je.convertToJaegerFormat(spans)
	
	// 构建请求
	payload := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"traceID": spans[0].TraceID,
				"spans":   jaegerSpans,
			},
		},
	}
	
	// 发送请求
	if err := je.sendRequest(payload); err != nil {
		je.recordError(len(spans), err)
		return err
	}
	
	je.recordSuccess(len(spans), time.Since(start))
	return nil
}

// convertToJaegerFormat 转换为Jaeger格式
func (je *JaegerExporter) convertToJaegerFormat(spans []*Span) []map[string]interface{} {
	jaegerSpans := make([]map[string]interface{}, 0, len(spans))
	
	for _, span := range spans {
		jaegerSpan := map[string]interface{}{
			"traceID":       span.TraceID,
			"spanID":        span.SpanID,
			"operationName": span.OperationName,
			"startTime":     span.StartTime.UnixNano() / 1000, // 微秒
			"duration":      span.Duration.Nanoseconds() / 1000, // 微秒
			"tags":          je.convertTags(span.Tags),
			"logs":          je.convertLogs(span.Logs),
		}
		
		if span.ParentSpanID != "" {
			jaegerSpan["parentSpanID"] = span.ParentSpanID
		}
		
		jaegerSpans = append(jaegerSpans, jaegerSpan)
	}
	
	return jaegerSpans
}

// convertTags 转换标签
func (je *JaegerExporter) convertTags(tags map[string]interface{}) []map[string]interface{} {
	jaegerTags := make([]map[string]interface{}, 0, len(tags))
	
	for key, value := range tags {
		tag := map[string]interface{}{
			"key":   key,
			"value": fmt.Sprintf("%v", value),
		}
		
		// 确定类型
		switch value.(type) {
		case string:
			tag["type"] = "string"
		case bool:
			tag["type"] = "bool"
		case int, int32, int64, float32, float64:
			tag["type"] = "number"
		default:
			tag["type"] = "string"
		}
		
		jaegerTags = append(jaegerTags, tag)
	}
	
	return jaegerTags
}

// convertLogs 转换日志
func (je *JaegerExporter) convertLogs(logs []*LogEntry) []map[string]interface{} {
	jaegerLogs := make([]map[string]interface{}, 0, len(logs))
	
	for _, log := range logs {
		fields := make([]map[string]interface{}, 0, len(log.Fields))
		
		for key, value := range log.Fields {
			fields = append(fields, map[string]interface{}{
				"key":   key,
				"value": fmt.Sprintf("%v", value),
			})
		}
		
		jaegerLog := map[string]interface{}{
			"timestamp": log.Timestamp.UnixNano() / 1000, // 微秒
			"fields":    fields,
		}
		
		jaegerLogs = append(jaegerLogs, jaegerLog)
	}
	
	return jaegerLogs
}

// sendRequest 发送请求
func (je *JaegerExporter) sendRequest(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequest("POST", je.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 添加自定义头
	for key, value := range je.config.Headers {
		req.Header.Set(key, value)
	}
	
	resp, err := je.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	
	return nil
}

// ZipkinExporter Zipkin导出器
type ZipkinExporter struct {
	*BaseExporter
	endpoint string
}

// NewZipkinExporter 创建Zipkin导出器
func NewZipkinExporter(config ExporterConfig) (*ZipkinExporter, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("zipkin endpoint is required")
	}
	
	return &ZipkinExporter{
		BaseExporter: NewBaseExporter(config),
		endpoint:     config.Endpoint,
	}, nil
}

// Start 启动导出器
func (ze *ZipkinExporter) Start() error {
	return nil
}

// Stop 停止导出器
func (ze *ZipkinExporter) Stop() error {
	return nil
}

// Export 导出spans
func (ze *ZipkinExporter) Export(spans []*Span) error {
	if len(spans) == 0 {
		return nil
	}
	
	start := time.Now()
	
	// 转换为Zipkin格式
	zipkinSpans := ze.convertToZipkinFormat(spans)
	
	// 发送请求
	if err := ze.sendRequest(zipkinSpans); err != nil {
		ze.recordError(len(spans), err)
		return err
	}
	
	ze.recordSuccess(len(spans), time.Since(start))
	return nil
}

// convertToZipkinFormat 转换为Zipkin格式
func (ze *ZipkinExporter) convertToZipkinFormat(spans []*Span) []map[string]interface{} {
	zipkinSpans := make([]map[string]interface{}, 0, len(spans))
	
	for _, span := range spans {
		zipkinSpan := map[string]interface{}{
			"traceId":      span.TraceID,
			"id":           span.SpanID,
			"name":         span.OperationName,
			"timestamp":    span.StartTime.UnixNano() / 1000, // 微秒
			"duration":     span.Duration.Nanoseconds() / 1000, // 微秒
			"tags":         span.Tags,
			"annotations":  ze.convertLogs(span.Logs),
		}
		
		if span.ParentSpanID != "" {
			zipkinSpan["parentId"] = span.ParentSpanID
		}
		
		// 添加本地端点信息
		if serviceName, ok := span.Tags["service.name"]; ok {
			zipkinSpan["localEndpoint"] = map[string]interface{}{
				"serviceName": serviceName,
			}
		}
		
		zipkinSpans = append(zipkinSpans, zipkinSpan)
	}
	
	return zipkinSpans
}

// convertLogs 转换日志为Zipkin注释
func (ze *ZipkinExporter) convertLogs(logs []*LogEntry) []map[string]interface{} {
	annotations := make([]map[string]interface{}, 0, len(logs))
	
	for _, log := range logs {
		// 将字段合并为单个值
		var value string
		if message, ok := log.Fields["message"]; ok {
			value = fmt.Sprintf("%v", message)
		} else {
			// 合并所有字段
			fields := make([]string, 0, len(log.Fields))
			for k, v := range log.Fields {
				fields = append(fields, fmt.Sprintf("%s=%v", k, v))
			}
			value = fmt.Sprintf("{%s}", fmt.Sprintf("%v", fields))
		}
		
		annotation := map[string]interface{}{
			"timestamp": log.Timestamp.UnixNano() / 1000, // 微秒
			"value":     value,
		}
		
		annotations = append(annotations, annotation)
	}
	
	return annotations
}

// sendRequest 发送请求
func (ze *ZipkinExporter) sendRequest(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequest("POST", ze.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 添加自定义头
	for key, value := range ze.config.Headers {
		req.Header.Set(key, value)
	}
	
	resp, err := ze.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	
	return nil
}

// OTLPExporter OTLP导出器
type OTLPExporter struct {
	*BaseExporter
	endpoint string
}

// NewOTLPExporter 创建OTLP导出器
func NewOTLPExporter(config ExporterConfig) (*OTLPExporter, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("otlp endpoint is required")
	}
	
	return &OTLPExporter{
		BaseExporter: NewBaseExporter(config),
		endpoint:     config.Endpoint,
	}, nil
}

// Start 启动导出器
func (oe *OTLPExporter) Start() error {
	return nil
}

// Stop 停止导出器
func (oe *OTLPExporter) Stop() error {
	return nil
}

// Export 导出spans
func (oe *OTLPExporter) Export(spans []*Span) error {
	if len(spans) == 0 {
		return nil
	}
	
	start := time.Now()
	
	// 转换为OTLP格式
	otlpData := oe.convertToOTLPFormat(spans)
	
	// 发送请求
	if err := oe.sendRequest(otlpData); err != nil {
		oe.recordError(len(spans), err)
		return err
	}
	
	oe.recordSuccess(len(spans), time.Since(start))
	return nil
}

// convertToOTLPFormat 转换为OTLP格式
func (oe *OTLPExporter) convertToOTLPFormat(spans []*Span) map[string]interface{} {
	// 按服务分组spans
	serviceSpans := make(map[string][]*Span)
	for _, span := range spans {
		serviceName := "unknown-service"
		if name, ok := span.Tags["service.name"]; ok {
			serviceName = fmt.Sprintf("%v", name)
		}
		serviceSpans[serviceName] = append(serviceSpans[serviceName], span)
	}
	
	// 构建OTLP格式
	resourceSpans := make([]map[string]interface{}, 0, len(serviceSpans))
	
	for serviceName, spans := range serviceSpans {
		// 资源信息
		resource := map[string]interface{}{
			"attributes": []map[string]interface{}{
				{
					"key": "service.name",
					"value": map[string]interface{}{
						"stringValue": serviceName,
					},
				},
			},
		}
		
		// 转换spans
		otlpSpans := make([]map[string]interface{}, 0, len(spans))
		for _, span := range spans {
			otlpSpan := map[string]interface{}{
				"traceId":           oe.hexToBytes(span.TraceID),
				"spanId":            oe.hexToBytes(span.SpanID),
				"name":              span.OperationName,
				"startTimeUnixNano": span.StartTime.UnixNano(),
				"endTimeUnixNano":   span.EndTime.UnixNano(),
				"attributes":        oe.convertAttributes(span.Tags),
				"events":            oe.convertEvents(span.Logs),
				"status":            oe.convertStatus(span.Status),
			}
			
			if span.ParentSpanID != "" {
				otlpSpan["parentSpanId"] = oe.hexToBytes(span.ParentSpanID)
			}
			
			otlpSpans = append(otlpSpans, otlpSpan)
		}
		
		resourceSpan := map[string]interface{}{
			"resource": resource,
			"scopeSpans": []map[string]interface{}{
				{
					"scope": map[string]interface{}{
						"name":    "monitoring-tracer",
						"version": "1.0.0",
					},
					"spans": otlpSpans,
				},
			},
		}
		
		resourceSpans = append(resourceSpans, resourceSpan)
	}
	
	return map[string]interface{}{
		"resourceSpans": resourceSpans,
	}
}

// hexToBytes 将十六进制字符串转换为字节数组
func (oe *OTLPExporter) hexToBytes(hexStr string) []byte {
	// 这里应该实现十六进制到字节的转换
	// 为了简化，直接返回字符串的字节表示
	return []byte(hexStr)
}

// convertAttributes 转换属性
func (oe *OTLPExporter) convertAttributes(tags map[string]interface{}) []map[string]interface{} {
	attributes := make([]map[string]interface{}, 0, len(tags))
	
	for key, value := range tags {
		attr := map[string]interface{}{
			"key": key,
		}
		
		// 根据类型设置值
		switch v := value.(type) {
		case string:
			attr["value"] = map[string]interface{}{
				"stringValue": v,
			}
		case bool:
			attr["value"] = map[string]interface{}{
				"boolValue": v,
			}
		case int, int32, int64:
			attr["value"] = map[string]interface{}{
				"intValue": v,
			}
		case float32, float64:
			attr["value"] = map[string]interface{}{
				"doubleValue": v,
			}
		default:
			attr["value"] = map[string]interface{}{
				"stringValue": fmt.Sprintf("%v", v),
			}
		}
		
		attributes = append(attributes, attr)
	}
	
	return attributes
}

// convertEvents 转换事件
func (oe *OTLPExporter) convertEvents(logs []*LogEntry) []map[string]interface{} {
	events := make([]map[string]interface{}, 0, len(logs))
	
	for _, log := range logs {
		event := map[string]interface{}{
			"timeUnixNano": log.Timestamp.UnixNano(),
			"name":         "log",
			"attributes":   oe.convertAttributes(log.Fields),
		}
		
		events = append(events, event)
	}
	
	return events
}

// convertStatus 转换状态
func (oe *OTLPExporter) convertStatus(status SpanStatus) map[string]interface{} {
	var code int
	var message string
	
	switch status {
	case SpanStatusOK:
		code = 1 // STATUS_CODE_OK
		message = "OK"
	case SpanStatusError:
		code = 2 // STATUS_CODE_ERROR
		message = "Error"
	case SpanStatusTimeout:
		code = 2 // STATUS_CODE_ERROR
		message = "Timeout"
	case SpanStatusCancelled:
		code = 2 // STATUS_CODE_ERROR
		message = "Cancelled"
	default:
		code = 0 // STATUS_CODE_UNSET
		message = "Unset"
	}
	
	return map[string]interface{}{
		"code":    code,
		"message": message,
	}
}

// sendRequest 发送请求
func (oe *OTLPExporter) sendRequest(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequest("POST", oe.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 添加自定义头
	for key, value := range oe.config.Headers {
		req.Header.Set(key, value)
	}
	
	resp, err := oe.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	
	return nil
}

// ConsoleExporter 控制台导出器
type ConsoleExporter struct {
	*BaseExporter
}

// NewConsoleExporter 创建控制台导出器
func NewConsoleExporter(config ExporterConfig) (*ConsoleExporter, error) {
	return &ConsoleExporter{
		BaseExporter: NewBaseExporter(config),
	}, nil
}

// Start 启动导出器
func (ce *ConsoleExporter) Start() error {
	return nil
}

// Stop 停止导出器
func (ce *ConsoleExporter) Stop() error {
	return nil
}

// Export 导出spans
func (ce *ConsoleExporter) Export(spans []*Span) error {
	if len(spans) == 0 {
		return nil
	}
	
	start := time.Now()
	
	for _, span := range spans {
		ce.printSpan(span)
	}
	
	ce.recordSuccess(len(spans), time.Since(start))
	return nil
}

// printSpan 打印span
func (ce *ConsoleExporter) printSpan(span *Span) {
	fmt.Printf("Span: %s\n", span.OperationName)
	fmt.Printf("  TraceID: %s\n", span.TraceID)
	fmt.Printf("  SpanID: %s\n", span.SpanID)
	if span.ParentSpanID != "" {
		fmt.Printf("  ParentSpanID: %s\n", span.ParentSpanID)
	}
	fmt.Printf("  StartTime: %s\n", span.StartTime.Format(time.RFC3339Nano))
	fmt.Printf("  Duration: %v\n", span.Duration)
	fmt.Printf("  Status: %v\n", span.Status)
	
	if len(span.Tags) > 0 {
		fmt.Printf("  Tags:\n")
		for key, value := range span.Tags {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}
	
	if len(span.Logs) > 0 {
		fmt.Printf("  Logs:\n")
		for _, log := range span.Logs {
			fmt.Printf("    %s: %v\n", log.Timestamp.Format(time.RFC3339Nano), log.Fields)
		}
	}
	
	fmt.Println()
}