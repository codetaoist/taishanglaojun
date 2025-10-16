﻿package recommendation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// RecommendationIntegrationService 
type RecommendationIntegrationService struct {
	personalizationEngine         *domainServices.PersonalizationEngine
	userBehaviorTracker           *domainServices.UserBehaviorTracker
	preferenceAnalyzer            *domainServices.PreferenceAnalyzer
	contextAnalyzer               *domainServices.ContextAnalyzer
	realtimeRecommendationService *RealtimeRecommendationService
	
	// ?
	cache                         map[string]*CachedRecommendation
	cacheMutex                    sync.RWMutex
	config                        *RecommendationIntegrationConfig
	
	// 
	metrics                       *IntegrationRecommendationMetrics
	metricsMutex                  sync.RWMutex
}

// IntegrationConfig 
type RecommendationIntegrationConfig struct {
	CacheEnabled         bool          `json:"cache_enabled"`
	CacheTTL            time.Duration `json:"cache_ttl"`
	RealtimeEnabled     bool          `json:"realtime_enabled"`
	BatchSize           int           `json:"batch_size"`
	MaxConcurrency      int           `json:"max_concurrency"`
	FallbackStrategy    string        `json:"fallback_strategy"`
	QualityThreshold    float64       `json:"quality_threshold"`
	DiversityWeight     float64       `json:"diversity_weight"`
	FreshnessWeight     float64       `json:"freshness_weight"`
	PersonalizationWeight float64     `json:"personalization_weight"`
}

// CachedRecommendation ?
type CachedRecommendation struct {
	UserID          string                      `json:"user_id"`
	Recommendations []*domainServices.PersonalizedRecommendation `json:"recommendations"`
	Strategy        string                      `json:"strategy"`
	Timestamp       time.Time                   `json:"timestamp"`
	TTL             time.Duration               `json:"ttl"`
	Context         map[string]interface{}      `json:"context"`
}

// RecommendationMetrics 
type IntegrationRecommendationMetrics struct {
	TotalRequests       int64     `json:"total_requests"`
	CacheHits          int64     `json:"cache_hits"`
	CacheMisses        int64     `json:"cache_misses"`
	AverageResponseTime float64   `json:"average_response_time_ms"`
	ErrorRate          float64   `json:"error_rate"`
	QualityScore       float64   `json:"quality_score"`
	DiversityScore     float64   `json:"diversity_score"`
	LastUpdated        time.Time `json:"last_updated"`
}

// IntegratedRecommendationRequest 
type IntegratedRecommendationRequest struct {
	UserID           string                 `json:"user_id" binding:"required"`
	ContentType      string                 `json:"content_type,omitempty"`
	Strategy         string                 `json:"strategy,omitempty"`
	Limit            int                    `json:"limit,omitempty"`
	Context          map[string]interface{} `json:"context,omitempty"`
	Preferences      map[string]interface{} `json:"preferences,omitempty"`
	UseCache         bool                   `json:"use_cache"`
	UseRealtime      bool                   `json:"use_realtime"`
	QualityThreshold float64                `json:"quality_threshold,omitempty"`
	DiversityEnabled bool                   `json:"diversity_enabled"`
}

// IntegratedRecommendationResponse 
type IntegratedRecommendationResponse struct {
	UserID          string                      `json:"user_id"`
	Recommendations []*domainServices.PersonalizedRecommendation `json:"recommendations"`
	Strategy        string                      `json:"strategy"`
	Source          string                      `json:"source"` // cache, realtime, engine
	Quality         *IntegrationQualityMetrics  `json:"quality"`
	Metadata        *IntegrationMetadata        `json:"metadata"`
}

// IntegrationQualityMetrics 
type IntegrationQualityMetrics struct {
	OverallScore     float64 `json:"overall_score"`
	RelevanceScore   float64 `json:"relevance_score"`
	DiversityScore   float64 `json:"diversity_score"`
	FreshnessScore   float64 `json:"freshness_score"`
	PersonalizationScore float64 `json:"personalization_score"`
	ConfidenceLevel  float64 `json:"confidence_level"`
}

// IntegrationMetadata ?
type IntegrationMetadata struct {
	ProcessingTime   int64                  `json:"processing_time_ms"`
	CacheUsed        bool                   `json:"cache_used"`
	RealtimeUsed     bool                   `json:"realtime_used"`
	FallbackUsed     bool                   `json:"fallback_used"`
	ComponentsUsed   []string               `json:"components_used"`
	Timestamp        time.Time              `json:"timestamp"`
	RequestID        string                 `json:"request_id"`
	Context          map[string]interface{} `json:"context"`
}

// NewRecommendationIntegrationService 
func NewRecommendationIntegrationService(
	personalizationEngine *domainServices.PersonalizationEngine,
	userBehaviorTracker *domainServices.UserBehaviorTracker,
	preferenceAnalyzer *domainServices.PreferenceAnalyzer,
	contextAnalyzer *domainServices.ContextAnalyzer,
	realtimeService *RealtimeRecommendationService,
	config *RecommendationIntegrationConfig,
) *RecommendationIntegrationService {
	if config == nil {
		config = &RecommendationIntegrationConfig{
			CacheEnabled:           true,
			CacheTTL:              15 * time.Minute,
			RealtimeEnabled:       true,
			BatchSize:             50,
			MaxConcurrency:        10,
			FallbackStrategy:      "popular",
			QualityThreshold:      0.7,
			DiversityWeight:       0.3,
			FreshnessWeight:       0.2,
			PersonalizationWeight: 0.5,
		}
	}

	return &RecommendationIntegrationService{
		personalizationEngine:         personalizationEngine,
		userBehaviorTracker:           userBehaviorTracker,
		preferenceAnalyzer:            preferenceAnalyzer,
		contextAnalyzer:               contextAnalyzer,
		realtimeRecommendationService: realtimeService,
		cache:                         make(map[string]*CachedRecommendation),
		config:                        config,
		metrics: &IntegrationRecommendationMetrics{
			LastUpdated: time.Now(),
		},
	}
}

// GetIntegratedRecommendations 
func (s *RecommendationIntegrationService) GetIntegratedRecommendations(
	ctx context.Context,
	req *IntegratedRecommendationRequest,
) (*IntegratedRecommendationResponse, error) {
	startTime := time.Now()
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	
	s.updateMetrics(func(m *IntegrationRecommendationMetrics) {
		m.TotalRequests++
	})

	// 
	response := &IntegratedRecommendationResponse{
		UserID:   req.UserID,
		Strategy: req.Strategy,
		Metadata: &IntegrationMetadata{
			RequestID:      requestID,
			Timestamp:      time.Now(),
			ComponentsUsed: []string{},
			Context:        req.Context,
		},
	}

	// 1. 黺?
	if req.UseCache && s.config.CacheEnabled {
		if cached := s.getCachedRecommendations(req.UserID, req.Strategy); cached != nil {
			response.Recommendations = cached.Recommendations
			response.Source = "cache"
			response.Metadata.CacheUsed = true
			response.Metadata.ProcessingTime = time.Since(startTime).Milliseconds()
			
			s.updateMetrics(func(m *IntegrationRecommendationMetrics) {
			m.CacheHits++
		})
		
		return response, nil
	}
	s.updateMetrics(func(m *IntegrationRecommendationMetrics) {
		m.CacheMisses++
	})
	}

	var recommendations []*domainServices.PersonalizedRecommendation
	var err error
	var source string
	componentsUsed := []string{}

	// 2. 
	if req.UseRealtime && s.config.RealtimeEnabled {
		recommendations, err = s.getRealtimeRecommendations(ctx, req)
		if err == nil && len(recommendations) > 0 {
			source = "realtime"
			componentsUsed = append(componentsUsed, "realtime")
			response.Metadata.RealtimeUsed = true
		}
	}

	// 3. 
	if len(recommendations) == 0 {
		recommendations, err = s.getPersonalizedRecommendations(ctx, req)
		if err == nil && len(recommendations) > 0 {
			source = "engine"
			componentsUsed = append(componentsUsed, "personalization_engine")
		}
	}

	// 4. 
	if len(recommendations) == 0 {
		recommendations, err = s.getFallbackRecommendations(ctx, req)
		if err == nil {
			source = "fallback"
			componentsUsed = append(componentsUsed, "fallback")
			response.Metadata.FallbackUsed = true
		}
	}

	if err != nil {
		s.updateMetrics(func(m *IntegrationRecommendationMetrics) {
		m.ErrorRate = (m.ErrorRate*float64(m.TotalRequests-1) + 1) / float64(m.TotalRequests)
	})
		return nil, fmt.Errorf(": %w", err)
	}

	// 5. ?
	if req.QualityThreshold > 0 || s.config.QualityThreshold > 0 {
		threshold := req.QualityThreshold
		if threshold == 0 {
			threshold = s.config.QualityThreshold
		}
		recommendations = s.filterByQuality(recommendations, threshold)
	}

	// 6. ?
	if req.DiversityEnabled {
		recommendations = s.enhanceDiversity(recommendations)
	}

	// 7. 
	if s.config.CacheEnabled && len(recommendations) > 0 {
		s.cacheRecommendations(req.UserID, req.Strategy, recommendations, req.Context)
	}

	// 8. 
	quality := s.calculateQualityMetrics(recommendations, req)

	// 9. 
	processingTime := time.Since(startTime).Milliseconds()
	s.updateMetrics(func(m *IntegrationRecommendationMetrics) {
		m.AverageResponseTime = (m.AverageResponseTime*float64(m.TotalRequests-1) + float64(processingTime)) / float64(m.TotalRequests)
		m.QualityScore = (m.QualityScore*float64(m.TotalRequests-1) + quality.OverallScore) / float64(m.TotalRequests)
		m.DiversityScore = (m.DiversityScore*float64(m.TotalRequests-1) + quality.DiversityScore) / float64(m.TotalRequests)
		m.LastUpdated = time.Now()
	})

	// 10. ?
	response.Recommendations = recommendations
	response.Source = source
	response.Quality = quality
	response.Metadata.ProcessingTime = processingTime
	response.Metadata.ComponentsUsed = componentsUsed

	return response, nil
}

// BatchGetRecommendations 
func (s *RecommendationIntegrationService) BatchGetRecommendations(
	ctx context.Context,
	requests []*IntegratedRecommendationRequest,
) ([]*IntegratedRecommendationResponse, error) {
	if len(requests) == 0 {
		return []*IntegratedRecommendationResponse{}, nil
	}

	if len(requests) > s.config.BatchSize {
		return nil, fmt.Errorf(": %d > %d", len(requests), s.config.BatchSize)
	}

	// 
	responses := make([]*IntegratedRecommendationResponse, len(requests))
	errors := make([]error, len(requests))
	
	// ?
	semaphore := make(chan struct{}, s.config.MaxConcurrency)
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request *IntegratedRecommendationRequest) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			response, err := s.GetIntegratedRecommendations(ctx, request)
			responses[index] = response
			errors[index] = err
		}(i, req)
	}

	wg.Wait()

	// ?
	var firstError error
	successCount := 0
	for _, err := range errors {
		if err != nil && firstError == nil {
			firstError = err
		}
		if err == nil {
			successCount++
		}
	}

	// ?
	if successCount == 0 {
		return nil, firstError
	}

	return responses, nil
}

// GetMetrics 
func (s *RecommendationIntegrationService) GetMetrics() *IntegrationRecommendationMetrics {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()
	
	// 
	metrics := *s.metrics
	return &metrics
}

// ClearCache 
func (s *RecommendationIntegrationService) ClearCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	s.cache = make(map[string]*CachedRecommendation)
}

// UpdateConfig 
func (s *RecommendationIntegrationService) UpdateConfig(config *RecommendationIntegrationConfig) {
	if config != nil {
		s.config = config
	}
}

// 

func (s *RecommendationIntegrationService) getCachedRecommendations(userID, strategy string) *CachedRecommendation {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	key := fmt.Sprintf("%s:%s", userID, strategy)
	cached, exists := s.cache[key]
	if !exists {
		return nil
	}
	
	// ?
	if time.Since(cached.Timestamp) > cached.TTL {
		// 
		go func() {
			s.cacheMutex.Lock()
			delete(s.cache, key)
			s.cacheMutex.Unlock()
		}()
		return nil
	}
	
	return cached
}

func (s *RecommendationIntegrationService) cacheRecommendations(
	userID, strategy string,
	recommendations []*domainServices.PersonalizedRecommendation,
	context map[string]interface{},
) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	key := fmt.Sprintf("%s:%s", userID, strategy)
	s.cache[key] = &CachedRecommendation{
		UserID:          userID,
		Recommendations: recommendations,
		Strategy:        strategy,
		Timestamp:       time.Now(),
		TTL:             s.config.CacheTTL,
		Context:         context,
	}
}

func (s *RecommendationIntegrationService) getRealtimeRecommendations(
	ctx context.Context,
	req *IntegratedRecommendationRequest,
) ([]*domainServices.PersonalizedRecommendation, error) {
	recommendations, err := s.realtimeRecommendationService.GetRealtimeRecommendations(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	
	// 
	result := make([]*domainServices.PersonalizedRecommendation, len(recommendations))
	for i, rec := range recommendations {
		result[i] = &domainServices.PersonalizedRecommendation{
			ContentID:     rec.ContentID,
			Title:         rec.Title,
			Description:   rec.Description,
			Type:          rec.Type,
			Score:         rec.Score,
			Confidence:    rec.Confidence,
			Reasoning:     rec.Reasoning,
			Tags:          rec.Tags,
			Metadata:      rec.Metadata,
			EstimatedTime: rec.EstimatedTime,
			Difficulty:    rec.Difficulty,
		}
	}
	
	return result, nil
}

func (s *RecommendationIntegrationService) getPersonalizedRecommendations(
	ctx context.Context,
	req *IntegratedRecommendationRequest,
) ([]*domainServices.PersonalizedRecommendation, error) {
	// 
	engineReq := &domainServices.PersonalizationRequest{
		LearnerID:           uuid.MustParse(req.UserID),
		RecommendationType:  req.ContentType,
		MaxRecommendations:  req.Limit,
		IncludeExplanations: true,
		PersonalizationLevel: req.Strategy,
	}
	
	response, err := s.personalizationEngine.GeneratePersonalizedRecommendations(ctx, engineReq)
	if err != nil {
		return nil, err
	}
	
	// ?
	result := make([]*domainServices.PersonalizedRecommendation, len(response.Recommendations))
	for i := range response.Recommendations {
		result[i] = &response.Recommendations[i]
	}
	
	return result, nil
}

func (s *RecommendationIntegrationService) getFallbackRecommendations(
	ctx context.Context,
	req *IntegratedRecommendationRequest,
) ([]*domainServices.PersonalizedRecommendation, error) {
	// 
	fallbackReq := &domainServices.PersonalizationRequest{
		LearnerID:           uuid.MustParse(req.UserID),
		RecommendationType:  req.ContentType,
		MaxRecommendations:  req.Limit,
		IncludeExplanations: true,
		PersonalizationLevel: s.config.FallbackStrategy,
	}
	
	response, err := s.personalizationEngine.GeneratePersonalizedRecommendations(ctx, fallbackReq)
	if err != nil {
		return nil, err
	}
	
	// ?
	result := make([]*domainServices.PersonalizedRecommendation, len(response.Recommendations))
	for i := range response.Recommendations {
		result[i] = &response.Recommendations[i]
	}
	
	return result, nil
}

func (s *RecommendationIntegrationService) filterByQuality(
	recommendations []*domainServices.PersonalizedRecommendation,
	threshold float64,
) []*domainServices.PersonalizedRecommendation {
	filtered := make([]*domainServices.PersonalizedRecommendation, 0, len(recommendations))
	for _, rec := range recommendations {
		if rec.Score >= threshold {
			filtered = append(filtered, rec)
		}
	}
	return filtered
}

func (s *RecommendationIntegrationService) enhanceDiversity(
	recommendations []*domainServices.PersonalizedRecommendation,
) []*domainServices.PersonalizedRecommendation {
	// ?
	typeMap := make(map[string][]*domainServices.PersonalizedRecommendation)
	for _, rec := range recommendations {
		typeMap[rec.Type] = append(typeMap[rec.Type], rec)
	}
	
	// ?
	result := make([]*domainServices.PersonalizedRecommendation, 0, len(recommendations))
	maxPerType := len(recommendations) / len(typeMap)
	if maxPerType == 0 {
		maxPerType = 1
	}
	
	for _, recs := range typeMap {
		count := 0
		for _, rec := range recs {
			if count < maxPerType {
				result = append(result, rec)
				count++
			}
		}
	}
	
	return result
}

func (s *RecommendationIntegrationService) calculateQualityMetrics(
	recommendations []*domainServices.PersonalizedRecommendation,
	req *IntegratedRecommendationRequest,
) *IntegrationQualityMetrics {
	if len(recommendations) == 0 {
		return &IntegrationQualityMetrics{}
	}
	
	// 
	var totalScore, totalConfidence float64
	typeCount := make(map[string]int)
	
	for _, rec := range recommendations {
		totalScore += rec.Score
		totalConfidence += rec.Confidence
		typeCount[rec.Type]++
	}
	
	avgScore := totalScore / float64(len(recommendations))
	avgConfidence := totalConfidence / float64(len(recommendations))
	
	// ?
	diversityScore := float64(len(typeCount)) / float64(len(recommendations))
	if diversityScore > 1.0 {
		diversityScore = 1.0
	}
	
	// 
	freshnessScore := 0.8 // ?
	
	// 
	personalizationScore := avgConfidence
	
	// ?
	overallScore := avgScore*0.4 + diversityScore*s.config.DiversityWeight + 
		freshnessScore*s.config.FreshnessWeight + 
		personalizationScore*s.config.PersonalizationWeight
	
	return &IntegrationQualityMetrics{
		OverallScore:         overallScore,
		RelevanceScore:       avgScore,
		DiversityScore:       diversityScore,
		FreshnessScore:       freshnessScore,
		PersonalizationScore: personalizationScore,
		ConfidenceLevel:      avgConfidence,
	}
}

func (s *RecommendationIntegrationService) updateMetrics(updateFunc func(*IntegrationRecommendationMetrics)) {
	s.metricsMutex.Lock()
	defer s.metricsMutex.Unlock()
	updateFunc(s.metrics)
}

