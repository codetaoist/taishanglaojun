﻿package vision

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultImageService 
type DefaultImageService struct {
	config    ImageConfig
	providers map[string]ImageProvider
	cache     ImageCache
	logger    *zap.Logger
	mutex     sync.RWMutex
}

// ImageCache 
type ImageCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, expiry time.Duration)
	Delete(key string)
	Clear()
}

// NewDefaultImageService 
func NewDefaultImageService(config ImageConfig, cache ImageCache, logger *zap.Logger) *DefaultImageService {
	return &DefaultImageService{
		config:    config,
		providers: make(map[string]ImageProvider),
		cache:     cache,
		logger:    logger,
	}
}

// RegisterProvider 
func (s *DefaultImageService) RegisterProvider(name string, provider ImageProvider) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.providers[name] = provider
}

// RecognizeObjects 
func (s *DefaultImageService) RecognizeObjects(ctx context.Context, input ImageInput) (*ObjectRecognitionResult, error) {
	// 黺
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("objects:%s", s.generateCacheKey(input))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*ObjectRecognitionResult); ok {
				s.logger.Debug("Object recognition result found in cache",
					zap.String("image_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getObjectDetectionProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.RecognizeObjects(ctx, input)
	if err != nil {
		s.logger.Error("Object recognition failed",
			zap.String("image_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("objects:%s", s.generateCacheKey(input))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Object recognition completed",
		zap.String("image_id", input.ID),
		zap.Int("objects_count", result.TotalObjects),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// RecognizeFaces 
func (s *DefaultImageService) RecognizeFaces(ctx context.Context, input ImageInput) (*FaceRecognitionResult, error) {
	// 黺
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("faces:%s", s.generateCacheKey(input))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*FaceRecognitionResult); ok {
				s.logger.Debug("Face recognition result found in cache",
					zap.String("image_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getFaceRecognitionProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.RecognizeFaces(ctx, input)
	if err != nil {
		s.logger.Error("Face recognition failed",
			zap.String("image_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("faces:%s", s.generateCacheKey(input))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Face recognition completed",
		zap.String("image_id", input.ID),
		zap.Int("faces_count", result.TotalFaces),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// RecognizeText 
func (s *DefaultImageService) RecognizeText(ctx context.Context, input ImageInput) (*TextRecognitionResult, error) {
	// 黺
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("text:%s", s.generateCacheKey(input))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*TextRecognitionResult); ok {
				s.logger.Debug("Text recognition result found in cache",
					zap.String("image_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getTextRecognitionProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.RecognizeText(ctx, input)
	if err != nil {
		s.logger.Error("Text recognition failed",
			zap.String("image_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("text:%s", s.generateCacheKey(input))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Text recognition completed",
		zap.String("image_id", input.ID),
		zap.String("text_length", fmt.Sprintf("%d", len(result.Text))),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// RecognizeScene 
func (s *DefaultImageService) RecognizeScene(ctx context.Context, input ImageInput) (*SceneRecognitionResult, error) {
	// 黺
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("scene:%s", s.generateCacheKey(input))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*SceneRecognitionResult); ok {
				s.logger.Debug("Scene recognition result found in cache",
					zap.String("image_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getSceneRecognitionProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.RecognizeScene(ctx, input)
	if err != nil {
		s.logger.Error("Scene recognition failed",
			zap.String("image_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("scene:%s", s.generateCacheKey(input))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Scene recognition completed",
		zap.String("image_id", input.ID),
		zap.String("scene", result.Scene),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// AnalyzeImage 
func (s *DefaultImageService) AnalyzeImage(ctx context.Context, input ImageInput) (*ImageAnalysisResult, error) {
	// 黺
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("analysis:%s", s.generateCacheKey(input))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*ImageAnalysisResult); ok {
				s.logger.Debug("Image analysis result found in cache",
					zap.String("image_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getImageAnalysisProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.AnalyzeImage(ctx, input)
	if err != nil {
		s.logger.Error("Image analysis failed",
			zap.String("image_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("analysis:%s", s.generateCacheKey(input))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Image analysis completed",
		zap.String("image_id", input.ID),
		zap.Float64("quality_overall", result.Quality.Overall),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// DetectAnomalies 
func (s *DefaultImageService) DetectAnomalies(ctx context.Context, input ImageInput) (*AnomalyDetectionResult, error) {
	startTime := time.Now()

	// 
	result := &AnomalyDetectionResult{
		ID:        uuid.New().String(),
		RequestID: input.ID,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// 
	anomalies := make([]DetectedAnomaly, 0)

	// 1. 
	qualityAnomalies, err := s.detectQualityAnomalies(ctx, input)
	if err != nil {
		s.logger.Warn("Quality anomaly detection failed", zap.Error(err))
	} else {
		anomalies = append(anomalies, qualityAnomalies...)
	}

	// 2. 
	contentAnomalies, err := s.detectContentAnomalies(ctx, input)
	if err != nil {
		s.logger.Warn("Content anomaly detection failed", zap.Error(err))
	} else {
		anomalies = append(anomalies, contentAnomalies...)
	}

	// 3. 
	technicalAnomalies, err := s.detectTechnicalAnomalies(ctx, input)
	if err != nil {
		s.logger.Warn("Technical anomaly detection failed", zap.Error(err))
	} else {
		anomalies = append(anomalies, technicalAnomalies...)
	}

	result.Anomalies = anomalies
	result.HasAnomalies = len(anomalies) > 0
	result.AnomalyScore = s.calculateAnomalyScore(anomalies)
	result.ProcessingTime = time.Since(startTime)

	s.logger.Info("Anomaly detection completed",
		zap.String("image_id", input.ID),
		zap.Bool("has_anomalies", result.HasAnomalies),
		zap.Float64("anomaly_score", result.AnomalyScore),
		zap.Int("anomalies_count", len(anomalies)),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// CompareImages 
func (s *DefaultImageService) CompareImages(ctx context.Context, image1, image2 ImageInput) (*ImageComparisonResult, error) {
	startTime := time.Now()

	result := &ImageComparisonResult{
		ID:        uuid.New().String(),
		RequestID: fmt.Sprintf("%s-%s", image1.ID, image2.ID),
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// 
	similarity, err := s.calculateImageSimilarity(ctx, image1, image2)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate similarity: %w", err)
	}

	differences, err := s.findImageDifferences(ctx, image1, image2)
	if err != nil {
		s.logger.Warn("Failed to find differences", zap.Error(err))
		differences = make([]ImageDifference, 0)
	}

	matchedRegions, err := s.findMatchedRegions(ctx, image1, image2)
	if err != nil {
		s.logger.Warn("Failed to find matched regions", zap.Error(err))
		matchedRegions = make([]MatchedRegion, 0)
	}

	result.Similarity = similarity
	result.Differences = differences
	result.MatchedRegions = matchedRegions
	result.ProcessingTime = time.Since(startTime)

	s.logger.Info("Image comparison completed",
		zap.String("image1_id", image1.ID),
		zap.String("image2_id", image2.ID),
		zap.Float64("similarity", similarity),
		zap.Int("differences_count", len(differences)),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// ProcessImage 
func (s *DefaultImageService) ProcessImage(ctx context.Context, input ImageInput, operations []ImageOperation) (*ImageProcessingResult, error) {
	provider, err := s.getImageProcessingProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.ProcessImage(ctx, input, operations)
	if err != nil {
		s.logger.Error("Image processing failed",
			zap.String("image_id", input.ID),
			zap.Int("operations_count", len(operations)),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	s.logger.Info("Image processing completed",
		zap.String("image_id", input.ID),
		zap.Int("operations_count", len(operations)),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// EnhanceImage 
func (s *DefaultImageService) EnhanceImage(ctx context.Context, input ImageInput, options EnhancementOptions) (*ImageProcessingResult, error) {
	// 
	operations := s.generateEnhancementOperations(options)

	return s.ProcessImage(ctx, input, operations)
}

// BatchProcess 
func (s *DefaultImageService) BatchProcess(ctx context.Context, inputs []ImageInput, operations []ImageOperation) (*BatchProcessingResult, error) {
	if len(inputs) > s.config.MaxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum %d", len(inputs), s.config.MaxBatchSize)
	}

	startTime := time.Now()
	result := &BatchProcessingResult{
		ID:          uuid.New().String(),
		TotalImages: len(inputs),
		Results:     make([]ImageProcessingResult, 0),
		Errors:      make([]BatchProcessingError, 0),
		Timestamp:   time.Now(),
	}

	// 
	type processResult struct {
		index  int
		result *ImageProcessingResult
		err    error
	}

	resultChan := make(chan processResult, len(inputs))
	semaphore := make(chan struct{}, 10) //  10

	for i, input := range inputs {
		go func(index int, img ImageInput) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			res, err := s.ProcessImage(ctx, img, operations)
			resultChan <- processResult{
				index:  index,
				result: res,
				err:    err,
			}
		}(i, input)
	}

	// 
	for i := 0; i < len(inputs); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-resultChan:
			if res.err != nil {
				result.FailedImages++
				result.Errors = append(result.Errors, BatchProcessingError{
					ImageID: inputs[res.index].ID,
					Error:   res.err.Error(),
				})
			} else {
				result.ProcessedImages++
				result.Results = append(result.Results, *res.result)
			}
		}
	}

	result.ProcessingTime = time.Since(startTime)

	s.logger.Info("Batch processing completed",
		zap.Int("total_images", result.TotalImages),
		zap.Int("processed_images", result.ProcessedImages),
		zap.Int("failed_images", result.FailedImages),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// UpdateConfig 
func (s *DefaultImageService) UpdateConfig(config ImageConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.config = config
	return nil
}

// GetSupportedFormats 
func (s *DefaultImageService) GetSupportedFormats() []ImageFormat {
	formats := make(map[ImageFormat]bool)

	for _, provider := range s.providers {
		for _, format := range provider.GetSupportedFormats() {
			formats[format] = true
		}
	}

	result := make([]ImageFormat, 0, len(formats))
	for format := range formats {
		result = append(result, format)
	}

	return result
}

// GetSupportedOperations 
func (s *DefaultImageService) GetSupportedOperations() []OperationType {
	operations := make(map[OperationType]bool)

	for _, provider := range s.providers {
		for _, op := range provider.GetSupportedOperations() {
			operations[op] = true
		}
	}

	result := make([]OperationType, 0, len(operations))
	for op := range operations {
		result = append(result, op)
	}

	return result
}

// 

// generateCacheKey 
func (s *DefaultImageService) generateCacheKey(input ImageInput) string {
	return fmt.Sprintf("%s-%d-%d-%d", input.ID, input.Width, input.Height, input.Size)
}

// getObjectDetectionProvider 
func (s *DefaultImageService) getObjectDetectionProvider() (ImageProvider, error) {
	providerName := s.config.ObjectDetection.Provider
	if providerName == "" {
		return nil, fmt.Errorf("no object detection provider configured")
	}

	provider, exists := s.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("object detection provider %s not found", providerName)
	}

	return provider, nil
}

// getFaceRecognitionProvider 
func (s *DefaultImageService) getFaceRecognitionProvider() (ImageProvider, error) {
	providerName := s.config.FaceRecognition.Provider
	if providerName == "" {
		return nil, fmt.Errorf("no face recognition provider configured")
	}

	provider, exists := s.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("face recognition provider %s not found", providerName)
	}

	return provider, nil
}

// getTextRecognitionProvider 
func (s *DefaultImageService) getTextRecognitionProvider() (ImageProvider, error) {
	providerName := s.config.TextRecognition.Provider
	if providerName == "" {
		return nil, fmt.Errorf("no text recognition provider configured")
	}

	provider, exists := s.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("text recognition provider %s not found", providerName)
	}

	return provider, nil
}

// getSceneRecognitionProvider 
func (s *DefaultImageService) getSceneRecognitionProvider() (ImageProvider, error) {
	// 
	return s.getObjectDetectionProvider()
}

// getImageAnalysisProvider 
func (s *DefaultImageService) getImageAnalysisProvider() (ImageProvider, error) {
	// 
	return s.getObjectDetectionProvider()
}

// getImageProcessingProvider 
func (s *DefaultImageService) getImageProcessingProvider() (ImageProvider, error) {
	providerName := s.config.ImageProcessing.Provider
	if providerName == "" {
		return nil, fmt.Errorf("no image processing provider configured")
	}

	provider, exists := s.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("image processing provider %s not found", providerName)
	}

	return provider, nil
}

// 

// detectQualityAnomalies 
func (s *DefaultImageService) detectQualityAnomalies(ctx context.Context, input ImageInput) ([]DetectedAnomaly, error) {
	anomalies := make([]DetectedAnomaly, 0)

	// 
	if input.Width < 100 || input.Height < 100 {
		anomalies = append(anomalies, DetectedAnomaly{
			ID:          uuid.New().String(),
			Type:        "low_resolution",
			Severity:    "medium",
			Confidence:  0.9,
			Description: "Image resolution is too low",
		})
	}

	// 
	if input.Size > s.config.MaxImageSize {
		anomalies = append(anomalies, DetectedAnomaly{
			ID:          uuid.New().String(),
			Type:        "oversized",
			Severity:    "high",
			Confidence:  1.0,
			Description: "Image file size exceeds maximum limit",
		})
	}

	return anomalies, nil
}

// detectContentAnomalies 
func (s *DefaultImageService) detectContentAnomalies(ctx context.Context, input ImageInput) ([]DetectedAnomaly, error) {
	// 
	return make([]DetectedAnomaly, 0), nil
}

// detectTechnicalAnomalies 
func (s *DefaultImageService) detectTechnicalAnomalies(ctx context.Context, input ImageInput) ([]DetectedAnomaly, error) {
	anomalies := make([]DetectedAnomaly, 0)

	// 
	supportedFormats := s.GetSupportedFormats()
	isSupported := false
	for _, format := range supportedFormats {
		if format == input.Format {
			isSupported = true
			break
		}
	}

	if !isSupported {
		anomalies = append(anomalies, DetectedAnomaly{
			ID:          uuid.New().String(),
			Type:        "unsupported_format",
			Severity:    "high",
			Confidence:  1.0,
			Description: fmt.Sprintf("Image format %s is not supported", input.Format),
		})
	}

	return anomalies, nil
}

// calculateAnomalyScore 
func (s *DefaultImageService) calculateAnomalyScore(anomalies []DetectedAnomaly) float64 {
	if len(anomalies) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, anomaly := range anomalies {
		switch anomaly.Severity {
		case "low":
			totalScore += 0.3 * anomaly.Confidence
		case "medium":
			totalScore += 0.6 * anomaly.Confidence
		case "high":
			totalScore += 1.0 * anomaly.Confidence
		}
	}

	return totalScore / float64(len(anomalies))
}

// 

// calculateImageSimilarity 
func (s *DefaultImageService) calculateImageSimilarity(ctx context.Context, image1, image2 ImageInput) (float64, error) {
	// 㷨
	if image1.Width == image2.Width && image1.Height == image2.Height && image1.Size == image2.Size {
		return 0.95, nil
	}

	sizeDiff := float64(abs(int(image1.Size-image2.Size))) / float64(max(image1.Size, image2.Size))
	return 1.0 - sizeDiff, nil
}

// findImageDifferences 
func (s *DefaultImageService) findImageDifferences(ctx context.Context, image1, image2 ImageInput) ([]ImageDifference, error) {
	differences := make([]ImageDifference, 0)

	// 
	if image1.Width != image2.Width || image1.Height != image2.Height {
		differences = append(differences, ImageDifference{
			Type:        "size_difference",
			Severity:    0.8,
			Description: "Images have different dimensions",
		})
	}

	// 
	if image1.Format != image2.Format {
		differences = append(differences, ImageDifference{
			Type:        "format_difference",
			Severity:    0.5,
			Description: "Images have different formats",
		})
	}

	return differences, nil
}

// findMatchedRegions 
func (s *DefaultImageService) findMatchedRegions(ctx context.Context, image1, image2 ImageInput) ([]MatchedRegion, error) {
	// 㷨
	regions := make([]MatchedRegion, 0)

	if image1.Width == image2.Width && image1.Height == image2.Height {
		regions = append(regions, MatchedRegion{
			Region1: BoundingBox{
				X:      0,
				Y:      0,
				Width:  float64(image1.Width),
				Height: float64(image1.Height),
			},
			Region2: BoundingBox{
				X:      0,
				Y:      0,
				Width:  float64(image2.Width),
				Height: float64(image2.Height),
			},
			Similarity:  0.9,
			Description: "Full image match",
		})
	}

	return regions, nil
}

// generateEnhancementOperations 
func (s *DefaultImageService) generateEnhancementOperations(options EnhancementOptions) []ImageOperation {
	operations := make([]ImageOperation, 0)
	order := 1

	if options.Denoise {
		operations = append(operations, CreateImageOperation(OpFilter, map[string]interface{}{
			"type": "denoise",
		}, order))
		order++
	}

	if options.Sharpen {
		operations = append(operations, CreateImageOperation(OpSharpen, map[string]interface{}{
			"strength": 0.5,
		}, order))
		order++
	}

	if options.ColorCorrect {
		operations = append(operations, CreateImageOperation(OpFilter, map[string]interface{}{
			"type": "color_correct",
		}, order))
		order++
	}

	if options.Upscale && options.UpscaleFactor > 1.0 {
		operations = append(operations, CreateImageOperation(OpResize, map[string]interface{}{
			"scale": options.UpscaleFactor,
		}, order))
		order++
	}

	if options.Quality > 0 && options.Quality <= 1.0 {
		operations = append(operations, CreateImageOperation(OpCompress, map[string]interface{}{
			"quality": options.Quality,
		}, order))
		order++
	}

	return operations
}

// 

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

