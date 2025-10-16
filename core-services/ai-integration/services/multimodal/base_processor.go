package multimodal

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
)

// BaseProcessor 处理器基类
type BaseProcessor struct {
	processorType string
	validator     InputValidator
	preprocessor  InputPreprocessor
	postprocessor OutputPostprocessor
	errorHandler  ErrorHandler
}

// NewBaseProcessor 创建基础处理器
func NewBaseProcessor(
	processorType string,
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) *BaseProcessor {
	return &BaseProcessor{
		processorType: processorType,
		validator:     validator,
		preprocessor:  preprocessor,
		postprocessor: postprocessor,
		errorHandler:  errorHandler,
	}
}

// GetType 获取处理器类型
func (bp *BaseProcessor) GetType() string {
	return bp.processorType
}

// Validate 验证输入
func (bp *BaseProcessor) Validate(inputs []models.MultimodalInput, config models.MultimodalConfig) error {
	if bp.validator == nil {
		return nil
	}
	
	// 验证每个输入
	for _, input := range inputs {
		if err := bp.validator.ValidateInput(input, string(input.Type)); err != nil {
			return fmt.Errorf("input validation failed: %w", err)
		}
	}
	
	return nil
}

// PreprocessInputs 预处理输入
func (bp *BaseProcessor) PreprocessInputs(ctx context.Context, inputs []models.MultimodalInput) ([]models.MultimodalInput, error) {
	if bp.preprocessor == nil {
		return inputs, nil
	}
	
	return bp.preprocessor.PreprocessInputs(ctx, inputs)
}

// PostprocessOutputs 后处理输出
func (bp *BaseProcessor) PostprocessOutputs(ctx context.Context, outputs []models.MultimodalOutput, expectedOutputs []models.OutputType) ([]models.MultimodalOutput, error) {
	if bp.postprocessor == nil {
		return outputs, nil
	}
	
	return bp.postprocessor.PostprocessOutputs(ctx, outputs, expectedOutputs)
}

// WrapError 包装错误
func (bp *BaseProcessor) WrapError(err error, operation string) error {
	if bp.errorHandler == nil {
		return err
	}
	
	return bp.errorHandler.WrapProviderError(err, operation)
}

// ShouldRetry 判断是否应该重试
func (bp *BaseProcessor) ShouldRetry(err error) bool {
	if bp.errorHandler == nil {
		return false
	}
	
	return bp.errorHandler.ShouldRetryError(err)
}

// ProcessWithRetry 带重试的处理
func (bp *BaseProcessor) ProcessWithRetry(ctx context.Context, operation func() error, maxRetries int) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := operation(); err != nil {
			lastErr = err
			
			if !bp.ShouldRetry(err) || attempt == maxRetries {
				return bp.WrapError(err, bp.processorType)
			}
			
			// 计算重试延迟
			if bp.errorHandler != nil {
				delay := bp.errorHandler.CalculateRetryDelay(attempt)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					continue
				}
			}
		} else {
			return nil
		}
	}
	
	return bp.WrapError(lastErr, bp.processorType)
}