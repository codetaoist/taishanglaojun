package multimodal

import (
	"fmt"
	"time"
)

// Config 多模态服务配置
type Config struct {
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	MaxRetryDelay time.Duration `json:"max_retry_delay"`
	Timeout       time.Duration `json:"timeout"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxRetries:    3,
		RetryDelay:    time.Second,
		MaxRetryDelay: 30 * time.Second,
		Timeout:       60 * time.Second,
	}
}

// Error 多模态服务错误
type Error struct {
	Type      ErrorType              `json:"type"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"-"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// ErrorType 错误类型
type ErrorType string

// 错误类型常量
const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeProvider   ErrorType = "provider"
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeProcessing ErrorType = "processing"
	ErrorTypeUnknown    ErrorType = "unknown"
	ErrorTypeAuth       ErrorType = "auth"
	ErrorTypeConfig     ErrorType = "config"
	ErrorTypeResource   ErrorType = "resource"
)

// ProcessorType 处理器类型
type ProcessorType string

// 处理器类型常量
const (
	ProcessorTypeChat        ProcessorType = "chat"
	ProcessorTypeAnalysis    ProcessorType = "analysis"
	ProcessorTypeGeneration  ProcessorType = "generation"
	ProcessorTypeTranslation ProcessorType = "translation"
	ProcessorTypeSearch      ProcessorType = "search"
)