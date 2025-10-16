package multimodal

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	maxRetries   int
	retryDelay   time.Duration
	maxRetryDelay time.Duration
}

// NewDefaultErrorHandler 创建默认错误处理器
func NewDefaultErrorHandler(maxRetries int, retryDelay, maxRetryDelay time.Duration) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		maxRetries:    maxRetries,
		retryDelay:    retryDelay,
		maxRetryDelay: maxRetryDelay,
	}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	// 包装错误，添加操作上下文
	wrappedErr := &Error{
		Type:      ErrorTypeProcessing,
		Message:   fmt.Sprintf("operation '%s' failed: %s", operation, err.Error()),
		Cause:     err,
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"operation": operation,
		},
	}

	// 根据错误类型进行分类
	wrappedErr.Type = h.classifyError(err)

	// 记录错误（这里可以集成日志系统）
	h.logError(ctx, wrappedErr)

	return wrappedErr
}

// WrapError 包装错误
func (h *DefaultErrorHandler) WrapError(err error, errorType ErrorType, message string) error {
	if err == nil {
		return nil
	}

	return &Error{
		Type:      errorType,
		Message:   message,
		Cause:     err,
		Timestamp: time.Now(),
	}
}

// WrapProviderError 包装提供者错误
func (h *DefaultErrorHandler) WrapProviderError(err error, operation string) error {
	if err == nil {
		return nil
	}

	return &Error{
		Type:      h.classifyError(err),
		Message:   fmt.Sprintf("provider operation '%s' failed: %s", operation, err.Error()),
		Cause:     err,
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"operation": operation,
		},
	}
}

// ShouldRetryError 判断是否应该重试错误
func (h *DefaultErrorHandler) ShouldRetryError(err error) bool {
	return h.IsRetryable(err)
}

// CalculateRetryDelay 计算重试延迟
func (h *DefaultErrorHandler) CalculateRetryDelay(attempt int) time.Duration {
	return h.GetRetryDelay(attempt)
}

// IsRetryable 判断错误是否可重试
func (h *DefaultErrorHandler) IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是我们的自定义错误类型
	if multimodalErr, ok := err.(*Error); ok {
		return h.isRetryableByType(multimodalErr.Type)
	}

	// 检查错误消息中的关键词
	errMsg := strings.ToLower(err.Error())
	
	// 网络相关错误通常可重试
	networkErrors := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"network unreachable",
		"temporary failure",
		"service unavailable",
		"too many requests",
		"rate limit",
	}

	for _, keyword := range networkErrors {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	// 服务器错误（5xx）通常可重试
	serverErrors := []string{
		"internal server error",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
	}

	for _, keyword := range serverErrors {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

// GetRetryDelay 获取重试延迟
func (h *DefaultErrorHandler) GetRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return h.retryDelay
	}

	// 指数退避策略
	delay := h.retryDelay * time.Duration(1<<uint(attempt-1))
	
	// 限制最大延迟
	if delay > h.maxRetryDelay {
		delay = h.maxRetryDelay
	}

	return delay
}

// classifyError 分类错误
func (h *DefaultErrorHandler) classifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errMsg := strings.ToLower(err.Error())

	// 验证错误
	validationKeywords := []string{
		"validation",
		"invalid",
		"required",
		"missing",
		"empty",
		"format",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(errMsg, keyword) {
			return ErrorTypeValidation
		}
	}

	// 网络错误
	networkKeywords := []string{
		"network",
		"connection",
		"timeout",
		"unreachable",
		"dns",
		"socket",
	}

	for _, keyword := range networkKeywords {
		if strings.Contains(errMsg, keyword) {
			return ErrorTypeNetwork
		}
	}

	// 认证错误
	authKeywords := []string{
		"unauthorized",
		"authentication",
		"permission",
		"access denied",
		"forbidden",
		"api key",
		"token",
	}

	for _, keyword := range authKeywords {
		if strings.Contains(errMsg, keyword) {
			return ErrorTypeAuth
		}
	}

	// 配置错误
	configKeywords := []string{
		"configuration",
		"config",
		"setting",
		"parameter",
		"option",
	}

	for _, keyword := range configKeywords {
		if strings.Contains(errMsg, keyword) {
			return ErrorTypeConfig
		}
	}

	// 资源错误
	resourceKeywords := []string{
		"not found",
		"file not found",
		"resource",
		"quota",
		"limit",
		"capacity",
	}

	for _, keyword := range resourceKeywords {
		if strings.Contains(errMsg, keyword) {
			return ErrorTypeResource
		}
	}

	// 默认为处理错误
	return ErrorTypeProcessing
}

// isRetryableByType 根据错误类型判断是否可重试
func (h *DefaultErrorHandler) isRetryableByType(errorType ErrorType) bool {
	switch errorType {
	case ErrorTypeNetwork:
		return true // 网络错误通常可重试
	case ErrorTypeProcessing:
		return true // 处理错误可能是临时的
	case ErrorTypeResource:
		return false // 资源错误通常不可重试
	case ErrorTypeValidation:
		return false // 验证错误不可重试
	case ErrorTypeAuth:
		return false // 认证错误不可重试
	case ErrorTypeConfig:
		return false // 配置错误不可重试
	default:
		return false
	}
}

// logError 记录错误
func (h *DefaultErrorHandler) logError(ctx context.Context, err *Error) {
	// 这里可以集成实际的日志系统
	// 例如：logrus, zap, 或者自定义的日志系统
	
	// 简单的控制台输出（在实际应用中应该替换为真正的日志记录）
	fmt.Printf("[ERROR] %s - Type: %s, Message: %s\n", 
		err.Timestamp.Format(time.RFC3339), 
		err.Type, 
		err.Message)
	
	if err.Cause != nil {
		fmt.Printf("[ERROR] Caused by: %s\n", err.Cause.Error())
	}
}

// RecoverFromPanic 从panic中恢复
func (h *DefaultErrorHandler) RecoverFromPanic(ctx context.Context, operation string) error {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic in operation '%s': %v", operation, r)
		return h.HandleError(ctx, err, operation)
	}
	return nil
}

// CreateTimeoutError 创建超时错误
func (h *DefaultErrorHandler) CreateTimeoutError(operation string, timeout time.Duration) error {
	return &Error{
		Type:      ErrorTypeNetwork,
		Message:   fmt.Sprintf("operation '%s' timed out after %v", operation, timeout),
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"operation": operation,
			"timeout":   timeout.String(),
		},
	}
}

// CreateValidationError 创建验证错误
func (h *DefaultErrorHandler) CreateValidationError(field, message string) error {
	return &Error{
		Type:      ErrorTypeValidation,
		Message:   fmt.Sprintf("validation failed for field '%s': %s", field, message),
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"field":   field,
			"message": message,
		},
	}
}

// CreateResourceError 创建资源错误
func (h *DefaultErrorHandler) CreateResourceError(resource, message string) error {
	return &Error{
		Type:      ErrorTypeResource,
		Message:   fmt.Sprintf("resource error for '%s': %s", resource, message),
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"resource": resource,
			"message":  message,
		},
	}
}

// CreateAuthError 创建认证错误
func (h *DefaultErrorHandler) CreateAuthError(message string) error {
	return &Error{
		Type:      ErrorTypeAuth,
		Message:   fmt.Sprintf("authentication error: %s", message),
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"message": message,
		},
	}
}

// CreateConfigError 创建配置错误
func (h *DefaultErrorHandler) CreateConfigError(config, message string) error {
	return &Error{
		Type:      ErrorTypeConfig,
		Message:   fmt.Sprintf("configuration error for '%s': %s", config, message),
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"config":  config,
			"message": message,
		},
	}
}

// CreateProcessingError 创建处理错误
func (h *DefaultErrorHandler) CreateProcessingError(operation, message string) error {
	return &Error{
		Type:      ErrorTypeProcessing,
		Message:   fmt.Sprintf("processing error in '%s': %s", operation, message),
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"operation": operation,
			"message":   message,
		},
	}
}

// AggregateErrors 聚合多个错误
func (h *DefaultErrorHandler) AggregateErrors(errors []error, operation string) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return h.HandleError(context.Background(), errors[0], operation)
	}

	var messages []string
	var causes []error
	errorTypes := make(map[ErrorType]int)

	for _, err := range errors {
		if err != nil {
			messages = append(messages, err.Error())
			causes = append(causes, err)
			
			if multimodalErr, ok := err.(*Error); ok {
				errorTypes[multimodalErr.Type]++
			} else {
				errorTypes[h.classifyError(err)]++
			}
		}
	}

	// 确定主要错误类型
	var primaryType ErrorType = ErrorTypeUnknown
	maxCount := 0
	for errType, count := range errorTypes {
		if count > maxCount {
			maxCount = count
			primaryType = errType
		}
	}

	return &Error{
		Type:      primaryType,
		Message:   fmt.Sprintf("multiple errors in operation '%s': %s", operation, strings.Join(messages, "; ")),
		Timestamp: time.Now(),
		Context: map[string]interface{}{
			"operation":    operation,
			"error_count":  len(errors),
			"error_types":  errorTypes,
		},
	}
}

// ShouldRetry 判断是否应该重试
func (h *DefaultErrorHandler) ShouldRetry(err error, attempt int) bool {
	if err == nil {
		return false
	}

	if attempt >= h.maxRetries {
		return false
	}

	return h.IsRetryable(err)
}