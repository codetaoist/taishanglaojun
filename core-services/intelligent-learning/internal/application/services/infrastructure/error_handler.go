﻿package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// ErrorType 
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeService        ErrorType = "service"
	ErrorTypeIntegration    ErrorType = "integration"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeDatabase       ErrorType = "database"
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypePermission     ErrorType = "permission"
	ErrorTypeResource       ErrorType = "resource"
	ErrorTypeUnknown        ErrorType = "unknown"
)

// ErrorSeverity 
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// ServiceError 
type ServiceError struct {
	ID          string                 `json:"id"`
	Type        ErrorType              `json:"type"`
	Severity    ErrorSeverity          `json:"severity"`
	Service     string                 `json:"service"`
	Operation   string                 `json:"operation"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Cause       error                  `json:"cause,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Recoverable bool                   `json:"recoverable"`
	RetryCount  int                    `json:"retry_count"`
}

// Error error
func (e *ServiceError) Error() string {
	return fmt.Sprintf("[%s:%s] %s - %s", e.Service, e.Type, e.Operation, e.Message)
}

// ErrorHandler ?
type ErrorHandler struct {
	config       *ErrorHandlerConfig
	logger       *Logger
	errorHistory []ServiceError
	mu           sync.RWMutex
}

// ErrorHandlerConfig ?
type ErrorHandlerConfig struct {
	MaxHistorySize    int           `json:"max_history_size"`
	LogLevel          LogLevel      `json:"log_level"`
	EnableStackTrace  bool          `json:"enable_stack_trace"`
	EnableRetry       bool          `json:"enable_retry"`
	MaxRetryAttempts  int           `json:"max_retry_attempts"`
	RetryDelay        time.Duration `json:"retry_delay"`
	AlertThresholds   map[ErrorSeverity]int `json:"alert_thresholds"`
	NotificationHooks []string      `json:"notification_hooks"`
}

// NewErrorHandler ?
func NewErrorHandler(config *ErrorHandlerConfig, logger *Logger) *ErrorHandler {
	if config == nil {
		config = &ErrorHandlerConfig{
			MaxHistorySize:   1000,
			LogLevel:         LogLevelInfo,
			EnableStackTrace: true,
			EnableRetry:      true,
			MaxRetryAttempts: 3,
			RetryDelay:       time.Second * 2,
			AlertThresholds: map[ErrorSeverity]int{
				SeverityLow:      10,
				SeverityMedium:   5,
				SeverityHigh:     3,
				SeverityCritical: 1,
			},
		}
	}
	
	return &ErrorHandler{
		config:       config,
		logger:       logger,
		errorHistory: make([]ServiceError, 0),
	}
}

// HandleError 
func (eh *ErrorHandler) HandleError(ctx context.Context, err error, service, operation string, errorType ErrorType, severity ErrorSeverity) *ServiceError {
	serviceError := &ServiceError{
		ID:          eh.generateErrorID(),
		Type:        errorType,
		Severity:    severity,
		Service:     service,
		Operation:   operation,
		Message:     err.Error(),
		Cause:       err,
		Timestamp:   time.Now(),
		Recoverable: eh.isRecoverable(errorType, severity),
		RetryCount:  0,
	}
	
	// ?
	if ctx != nil {
		serviceError.Context = eh.extractContext(ctx)
	}
	
	// 
	if eh.config.EnableStackTrace {
		serviceError.StackTrace = eh.getStackTrace()
	}
	
	// 
	eh.recordError(serviceError)
	
	// 
	eh.logError(serviceError)
	
	// ?
	eh.checkAlertThresholds(serviceError)
	
	return serviceError
}

// CreateError 
func (eh *ErrorHandler) CreateError(service, operation, message string, errorType ErrorType, severity ErrorSeverity) *ServiceError {
	return &ServiceError{
		ID:          eh.generateErrorID(),
		Type:        errorType,
		Severity:    severity,
		Service:     service,
		Operation:   operation,
		Message:     message,
		Timestamp:   time.Now(),
		Recoverable: eh.isRecoverable(errorType, severity),
		RetryCount:  0,
	}
}

// WrapError 
func (eh *ErrorHandler) WrapError(err error, service, operation string, errorType ErrorType, severity ErrorSeverity, details map[string]interface{}) *ServiceError {
	serviceError := &ServiceError{
		ID:          eh.generateErrorID(),
		Type:        errorType,
		Severity:    severity,
		Service:     service,
		Operation:   operation,
		Message:     err.Error(),
		Details:     details,
		Cause:       err,
		Timestamp:   time.Now(),
		Recoverable: eh.isRecoverable(errorType, severity),
		RetryCount:  0,
	}
	
	if eh.config.EnableStackTrace {
		serviceError.StackTrace = eh.getStackTrace()
	}
	
	eh.recordError(serviceError)
	eh.logError(serviceError)
	eh.checkAlertThresholds(serviceError)
	
	return serviceError
}

// RetryOperation 
func (eh *ErrorHandler) RetryOperation(ctx context.Context, operation func() error, service, operationName string, maxRetries int) error {
	if !eh.config.EnableRetry {
		return operation()
	}
	
	if maxRetries <= 0 {
		maxRetries = eh.config.MaxRetryAttempts
	}
	
	var lastError error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(eh.config.RetryDelay * time.Duration(attempt)):
			}
			
			eh.logger.Info(fmt.Sprintf("Retrying operation %s for service %s (attempt %d/%d)", operationName, service, attempt, maxRetries))
		}
		
		err := operation()
		if err == nil {
			if attempt > 0 {
				eh.logger.Info(fmt.Sprintf("Operation %s for service %s succeeded after %d retries", operationName, service, attempt))
			}
			return nil
		}
		
		lastError = err
		
		// ?
		if serviceErr, ok := err.(*ServiceError); ok {
			if !serviceErr.Recoverable {
				eh.logger.Error(fmt.Sprintf("Non-recoverable error in operation %s for service %s: %v", operationName, service, err))
				return err
			}
			serviceErr.RetryCount = attempt
		}
		
		eh.logger.Warn(fmt.Sprintf("Operation %s for service %s failed (attempt %d/%d): %v", operationName, service, attempt+1, maxRetries+1, err))
	}
	
	// ?
	finalError := eh.HandleError(ctx, lastError, service, operationName, ErrorTypeService, SeverityHigh)
	finalError.RetryCount = maxRetries
	
	return finalError
}

// GetErrorHistory 
func (eh *ErrorHandler) GetErrorHistory(limit int) []ServiceError {
	eh.mu.RLock()
	defer eh.mu.RUnlock()
	
	if limit <= 0 || limit > len(eh.errorHistory) {
		limit = len(eh.errorHistory)
	}
	
	// 
	start := len(eh.errorHistory) - limit
	if start < 0 {
		start = 0
	}
	
	result := make([]ServiceError, limit)
	copy(result, eh.errorHistory[start:])
	
	return result
}

// GetErrorStatistics 
func (eh *ErrorHandler) GetErrorStatistics(timeRange time.Duration) map[string]interface{} {
	eh.mu.RLock()
	defer eh.mu.RUnlock()
	
	cutoff := time.Now().Add(-timeRange)
	
	stats := map[string]interface{}{
		"total_errors":    0,
		"by_type":         make(map[ErrorType]int),
		"by_severity":     make(map[ErrorSeverity]int),
		"by_service":      make(map[string]int),
		"recoverable":     0,
		"non_recoverable": 0,
	}
	
	for _, err := range eh.errorHistory {
		if err.Timestamp.After(cutoff) {
			stats["total_errors"] = stats["total_errors"].(int) + 1
			stats["by_type"].(map[ErrorType]int)[err.Type]++
			stats["by_severity"].(map[ErrorSeverity]int)[err.Severity]++
			stats["by_service"].(map[string]int)[err.Service]++
			
			if err.Recoverable {
				stats["recoverable"] = stats["recoverable"].(int) + 1
			} else {
				stats["non_recoverable"] = stats["non_recoverable"].(int) + 1
			}
		}
	}
	
	return stats
}

// recordError 
func (eh *ErrorHandler) recordError(err *ServiceError) {
	eh.mu.Lock()
	defer eh.mu.Unlock()
	
	eh.errorHistory = append(eh.errorHistory, *err)
	
	// 
	if len(eh.errorHistory) > eh.config.MaxHistorySize {
		eh.errorHistory = eh.errorHistory[len(eh.errorHistory)-eh.config.MaxHistorySize:]
	}
}

// logError 
func (eh *ErrorHandler) logError(err *ServiceError) {
	logData := map[string]interface{}{
		"error_id":   err.ID,
		"type":       err.Type,
		"severity":   err.Severity,
		"service":    err.Service,
		"operation":  err.Operation,
		"message":    err.Message,
		"timestamp":  err.Timestamp,
		"recoverable": err.Recoverable,
	}
	
	if err.Details != nil {
		logData["details"] = err.Details
	}
	
	if err.Context != nil {
		logData["context"] = err.Context
	}
	
	switch err.Severity {
	case SeverityCritical:
		eh.logger.Error(fmt.Sprintf("CRITICAL ERROR: %s", err.Error()), logData)
	case SeverityHigh:
		eh.logger.Error(fmt.Sprintf("HIGH SEVERITY ERROR: %s", err.Error()), logData)
	case SeverityMedium:
		eh.logger.Warn(fmt.Sprintf("MEDIUM SEVERITY ERROR: %s", err.Error()), logData)
	case SeverityLow:
		eh.logger.Info(fmt.Sprintf("LOW SEVERITY ERROR: %s", err.Error()), logData)
	}
}

// checkAlertThresholds 龯?
func (eh *ErrorHandler) checkAlertThresholds(err *ServiceError) {
	threshold, exists := eh.config.AlertThresholds[err.Severity]
	if !exists {
		return
	}
	
	// 
	cutoff := time.Now().Add(-time.Hour)
	count := 0
	
	eh.mu.RLock()
	for _, historyErr := range eh.errorHistory {
		if historyErr.Timestamp.After(cutoff) && historyErr.Severity == err.Severity {
			count++
		}
	}
	eh.mu.RUnlock()
	
	if count >= threshold {
		eh.sendAlert(err, count, threshold)
	}
}

// sendAlert ?
func (eh *ErrorHandler) sendAlert(err *ServiceError, count, threshold int) {
	alertData := map[string]interface{}{
		"error":     err,
		"count":     count,
		"threshold": threshold,
		"timestamp": time.Now(),
	}
	
	eh.logger.Error(fmt.Sprintf("ALERT: Error threshold exceeded for severity %s (count: %d, threshold: %d)", err.Severity, count, threshold), alertData)
	
	// 緢Slack?
	for _, hook := range eh.config.NotificationHooks {
		eh.executeNotificationHook(hook, alertData)
	}
}

// executeNotificationHook 
func (eh *ErrorHandler) executeNotificationHook(hook string, data map[string]interface{}) {
	// 
	eh.logger.Info(fmt.Sprintf("Executing notification hook: %s", hook), data)
}

// generateErrorID ID
func (eh *ErrorHandler) generateErrorID() string {
	return fmt.Sprintf("err_%d_%d", time.Now().UnixNano(), runtime.NumGoroutine())
}

// getStackTrace 
func (eh *ErrorHandler) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// isRecoverable ?
func (eh *ErrorHandler) isRecoverable(errorType ErrorType, severity ErrorSeverity) bool {
	// 
	if severity == SeverityCritical {
		return false
	}
	
	// 
	nonRecoverableTypes := []ErrorType{
		ErrorTypeConfiguration,
		ErrorTypePermission,
	}
	
	for _, nonRecoverable := range nonRecoverableTypes {
		if errorType == nonRecoverable {
			return false
		}
	}
	
	return true
}

// extractContext ?
func (eh *ErrorHandler) extractContext(ctx context.Context) map[string]interface{} {
	contextData := make(map[string]interface{})
	
	// ?
	if userID := ctx.Value("user_id"); userID != nil {
		contextData["user_id"] = userID
	}
	
	if requestID := ctx.Value("request_id"); requestID != nil {
		contextData["request_id"] = requestID
	}
	
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		contextData["session_id"] = sessionID
	}
	
	if traceID := ctx.Value("trace_id"); traceID != nil {
		contextData["trace_id"] = traceID
	}
	
	return contextData
}

// LogLevel 
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Logger ?
type Logger struct {
	level  LogLevel
	output func(level LogLevel, message string, data map[string]interface{})
	mu     sync.RWMutex
}

// NewLogger ?
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level: level,
		output: func(level LogLevel, message string, data map[string]interface{}) {
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			
			logEntry := map[string]interface{}{
				"timestamp": timestamp,
				"level":     level,
				"message":   message,
			}
			
			if data != nil {
				logEntry["data"] = data
			}
			
			jsonData, _ := json.Marshal(logEntry)
			log.Printf("[%s] %s", level, string(jsonData))
		},
	}
}

// SetOutput 
func (l *Logger) SetOutput(output func(level LogLevel, message string, data map[string]interface{})) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// Debug 
func (l *Logger) Debug(message string, data ...map[string]interface{}) {
	if l.shouldLog(LogLevelDebug) {
		l.log(LogLevelDebug, message, data...)
	}
}

// Info 
func (l *Logger) Info(message string, data ...map[string]interface{}) {
	if l.shouldLog(LogLevelInfo) {
		l.log(LogLevelInfo, message, data...)
	}
}

// Warn 
func (l *Logger) Warn(message string, data ...map[string]interface{}) {
	if l.shouldLog(LogLevelWarn) {
		l.log(LogLevelWarn, message, data...)
	}
}

// Error 
func (l *Logger) Error(message string, data ...map[string]interface{}) {
	if l.shouldLog(LogLevelError) {
		l.log(LogLevelError, message, data...)
	}
}

// log 
func (l *Logger) log(level LogLevel, message string, data ...map[string]interface{}) {
	l.mu.RLock()
	output := l.output
	l.mu.RUnlock()
	
	var logData map[string]interface{}
	if len(data) > 0 {
		logData = data[0]
	}
	
	output(level, message, logData)
}

// shouldLog ?
func (l *Logger) shouldLog(level LogLevel) bool {
	levelOrder := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}
	
	return levelOrder[level] >= levelOrder[l.level]
}

// SetLevel 
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

