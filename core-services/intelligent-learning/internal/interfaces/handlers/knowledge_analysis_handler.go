package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	knowledge "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/knowledge"
)

// KnowledgeAnalysisHandler ?
type KnowledgeAnalysisHandler struct {
	knowledgeAnalysisService *knowledge.KnowledgeAnalysisService
}

// NewKnowledgeAnalysisHandler ?
func NewKnowledgeAnalysisHandler(knowledgeAnalysisService *knowledge.KnowledgeAnalysisService) *KnowledgeAnalysisHandler {
	return &KnowledgeAnalysisHandler{
		knowledgeAnalysisService: knowledgeAnalysisService,
	}
}

// AnalyzeConceptRelationships 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param request body knowledge.ConceptRelationshipAnalysisRequest true ""
// @Success 200 {object} knowledge.ConceptRelationshipAnalysisResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/knowledge-analysis/concept-relationships [post]
func (h *KnowledgeAnalysisHandler) AnalyzeConceptRelationships(c *gin.Context) {
	var req knowledge.ConceptRelationshipAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// ?
	if req.AnalysisDepth == 0 {
		req.AnalysisDepth = 3
	}
	if req.MinStrength == 0 {
		req.MinStrength = 0.1
	}

	response, err := h.knowledgeAnalysisService.AnalyzeConceptRelationships(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to analyze concept relationships",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// BuildDependencyGraph ?
// @Summary ?
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body knowledge.DependencyGraphRequest true "?
// @Success 200 {object} knowledge.DependencyGraphResponse "?
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/knowledge-analysis/dependency-graph [post]
func (h *KnowledgeAnalysisHandler) BuildDependencyGraph(c *gin.Context) {
	var req knowledge.DependencyGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// ?
	if req.MaxDepth == 0 {
		req.MaxDepth = 10
	}

	response, err := h.knowledgeAnalysisService.BuildDependencyGraph(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to build dependency graph",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RecommendContent 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body knowledge.ContentRecommendationRequest true ""
// @Success 200 {object} knowledge.ContentRecommendationResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/knowledge-analysis/content-recommendations [post]
func (h *KnowledgeAnalysisHandler) RecommendContent(c *gin.Context) {
	var req knowledge.KnowledgeContentRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// ?
	if req.MaxRecommendations == 0 {
		req.MaxRecommendations = 10
	}
	if req.PersonalizationLevel == "" {
		req.PersonalizationLevel = "moderate"
	}

	response, err := h.knowledgeAnalysisService.RecommendContent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to recommend content",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetConceptClusters 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param graph_id path string true "ID"
// @Param include_metrics query bool false ""
// @Param min_cluster_size query int false "?
// @Success 200 {object} ConceptClustersResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/knowledge-analysis/graphs/{graph_id}/clusters [get]
func (h *KnowledgeAnalysisHandler) GetConceptClusters(c *gin.Context) {
	graphIDStr := c.Param("graph_id")
	graphID, err := uuid.Parse(graphIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid graph ID format",
			Message: "Graph ID must be a valid UUID",
		})
		return
	}

	includeMetrics := c.Query("include_metrics") == "true"
	minClusterSizeStr := c.DefaultQuery("min_cluster_size", "2")
	minClusterSize, err := strconv.Atoi(minClusterSizeStr)
	if err != nil {
		minClusterSize = 2
	}

	// 
	req := knowledge.ConceptRelationshipAnalysisRequest{
		GraphID:        graphID,
		AnalysisDepth:  3,
		IncludeMetrics: includeMetrics,
		MinStrength:    0.1,
	}

	response, err := h.knowledgeAnalysisService.AnalyzeConceptRelationships(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get concept clusters",
			Message: err.Error(),
		})
		return
	}

	// 
	filteredClusters := []knowledge.ConceptCluster{}
	for _, cluster := range response.ConceptClusters {
		if len(cluster.ConceptIDs) >= minClusterSize {
			filteredClusters = append(filteredClusters, cluster)
		}
	}

	clusterResponse := ConceptClustersResponse{
		GraphID:        graphID,
		Clusters:       filteredClusters,
		TotalClusters:  len(filteredClusters),
		AnalysisTime:   response.AnalysisTimestamp,
		IncludeMetrics: includeMetrics,
		MinClusterSize: minClusterSize,
	}

	c.JSON(http.StatusOK, clusterResponse)
}

// GetLearningPath 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param graph_id path string true "ID"
// @Param learner_id query string false "ID"
// @Param target_concept query string false "ID"
// @Param max_depth query int false "?
// @Param include_optional query bool false "?
// @Success 200 {object} LearningPathResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/knowledge-analysis/graphs/{graph_id}/learning-path [get]
func (h *KnowledgeAnalysisHandler) GetLearningPath(c *gin.Context) {
	graphIDStr := c.Param("graph_id")
	graphID, err := uuid.Parse(graphIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid graph ID format",
			Message: "Graph ID must be a valid UUID",
		})
		return
	}

	var targetConceptID *uuid.UUID
	if targetConceptStr := c.Query("target_concept"); targetConceptStr != "" {
		targetID, err := uuid.Parse(targetConceptStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid target concept ID format",
				Message: "Target concept ID must be a valid UUID",
			})
			return
		}
		targetConceptID = &targetID
	}

	maxDepthStr := c.DefaultQuery("max_depth", "10")
	maxDepth, err := strconv.Atoi(maxDepthStr)
	if err != nil {
		maxDepth = 10
	}

	includeOptional := c.Query("include_optional") == "true"

	// ?
	req := knowledge.DependencyGraphRequest{
		GraphID:         graphID,
		RootConceptID:   targetConceptID,
		MaxDepth:        maxDepth,
		IncludeOptional: includeOptional,
	}

	response, err := h.knowledgeAnalysisService.BuildDependencyGraph(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get learning path",
			Message: err.Error(),
		})
		return
	}

	pathResponse := LearningPathResponse{
		GraphID:               graphID,
		TargetConceptID:       targetConceptID,
		DependencyLayers:      response.DependencyLayers,
		CriticalPath:          response.CriticalPath,
		OptionalPaths:         response.OptionalPaths,
		LearningSequence:      response.LearningSequence,
		EstimatedDuration:     response.EstimatedDuration,
		DifficultyProgression: response.DifficultyProgression,
		GeneratedAt:           time.Now(),
	}

	c.JSON(http.StatusOK, pathResponse)
}

// GetPersonalizedRecommendations 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param learner_id path string true "ID"
// @Param concept_id query string false "ID"
// @Param max_recommendations query int false ""
// @Param personalization_level query string false ""
// @Param content_types query string false "?
// @Success 200 {object} PersonalizedRecommendationsResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/knowledge-analysis/learners/{learner_id}/recommendations [get]
func (h *KnowledgeAnalysisHandler) GetPersonalizedRecommendations(c *gin.Context) {
	learnerIDStr := c.Param("learner_id")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID format",
			Message: "Learner ID must be a valid UUID",
		})
		return
	}

	var conceptID *uuid.UUID
	if conceptIDStr := c.Query("concept_id"); conceptIDStr != "" {
		cID, err := uuid.Parse(conceptIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid concept ID format",
				Message: "Concept ID must be a valid UUID",
			})
			return
		}
		conceptID = &cID
	}

	maxRecommendationsStr := c.DefaultQuery("max_recommendations", "10")
	maxRecommendations, err := strconv.Atoi(maxRecommendationsStr)
	if err != nil {
		maxRecommendations = 10
	}

	personalizationLevel := c.DefaultQuery("personalization_level", "moderate")
	contentTypesStr := c.Query("content_types")
	var contentTypes []string
	if contentTypesStr != "" {
		contentTypes = parseCommaSeparated(contentTypesStr)
	}

	// 
	req := knowledge.KnowledgeContentRecommendationRequest{
		LearnerID:            learnerID,
		ConceptID:            conceptID,
		PreferredTypes:       contentTypes,
		MaxRecommendations:   maxRecommendations,
		IncludeReasoning:     true,
		PersonalizationLevel: personalizationLevel,
	}

	response, err := h.knowledgeAnalysisService.RecommendContent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get personalized recommendations",
			Message: err.Error(),
		})
		return
	}

	recommendationsResponse := PersonalizedRecommendationsResponse{
		LearnerID:              learnerID,
		ConceptID:              conceptID,
		Recommendations:        response.Recommendations,
		PersonalizationFactors: response.PersonalizationFactors,
		LearningPath:           response.LearningPath,
		TotalRecommendations:   len(response.Recommendations),
		PersonalizationLevel:   personalizationLevel,
		GeneratedAt:            response.GeneratedAt,
		ValidUntil:             response.ValidUntil,
	}

	c.JSON(http.StatusOK, recommendationsResponse)
}

// AnalyzeKnowledgeGaps 
// @Summary ?
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body KnowledgeGapAnalysisRequest true ""
// @Success 200 {object} KnowledgeGapAnalysisResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/knowledge-analysis/knowledge-gaps [post]
func (h *KnowledgeAnalysisHandler) AnalyzeKnowledgeGaps(c *gin.Context) {
	var req KnowledgeGapAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// 
	// ?
	response := KnowledgeGapAnalysisResponse{
		LearnerID:       req.LearnerID,
		GraphID:         req.GraphID,
		AnalysisID:      uuid.New(),
		Gaps:            []KnowledgeGap{},
		Recommendations: []GapRecommendation{},
		OverallScore:    0.75,
		AnalyzedAt:      time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// 嶨?
// KnowledgeGapAnalysisResponse 
type ConceptClustersResponse struct {
	GraphID        uuid.UUID                  `json:"graph_id"`
	Clusters       []knowledge.ConceptCluster `json:"clusters"`
	TotalClusters  int                        `json:"total_clusters"`
	AnalysisTime   time.Time                  `json:"analysis_time"`
	IncludeMetrics bool                       `json:"include_metrics"`
	MinClusterSize int                        `json:"min_cluster_size"`
}

// LearningPathResponse 
type LearningPathResponse struct {
	GraphID               uuid.UUID                   `json:"graph_id"`
	TargetConceptID       *uuid.UUID                  `json:"target_concept_id"`
	DependencyLayers      []knowledge.DependencyLayer `json:"dependency_layers"`
	CriticalPath          []uuid.UUID                 `json:"critical_path"`
	OptionalPaths         [][]uuid.UUID               `json:"optional_paths"`
	LearningSequence      []knowledge.LearningStep    `json:"learning_sequence"`
	EstimatedDuration     time.Duration               `json:"estimated_duration"`
	DifficultyProgression []knowledge.DifficultyPoint `json:"difficulty_progression"`
	GeneratedAt           time.Time                   `json:"generated_at"`
}

// PersonalizedRecommendationsResponse 
type PersonalizedRecommendationsResponse struct {
	LearnerID              uuid.UUID                                  `json:"learner_id"`
	ConceptID              *uuid.UUID                                 `json:"concept_id"`
	Recommendations        []knowledge.KnowledgeContentRecommendation `json:"recommendations"`
	PersonalizationFactors []knowledge.KnowledgePersonalizationFactor `json:"personalization_factors"`
	LearningPath           []uuid.UUID                                `json:"learning_path"`
	TotalRecommendations   int                                        `json:"total_recommendations"`
	PersonalizationLevel   string                                     `json:"personalization_level"`
	GeneratedAt            time.Time                                  `json:"generated_at"`
	ValidUntil             time.Time                                  `json:"valid_until"`
}

// KnowledgeGapAnalysisRequest 
type KnowledgeGapAnalysisRequest struct {
	LearnerID     uuid.UUID `json:"learner_id" binding:"required"`
	GraphID       uuid.UUID `json:"graph_id" binding:"required"`
	TargetSkills  []string  `json:"target_skills,omitempty"`
	CurrentLevel  string    `json:"current_level,omitempty"`
	AnalysisDepth int       `json:"analysis_depth"`
}

// KnowledgeGapAnalysisResponse 
type KnowledgeGapAnalysisResponse struct {
	LearnerID       uuid.UUID           `json:"learner_id"`
	GraphID         uuid.UUID           `json:"graph_id"`
	AnalysisID      uuid.UUID           `json:"analysis_id"`
	Gaps            []KnowledgeGap      `json:"gaps"`
	Recommendations []GapRecommendation `json:"recommendations"`
	OverallScore    float64             `json:"overall_score"`
	AnalyzedAt      time.Time           `json:"analyzed_at"`
}

// KnowledgeGap 
type KnowledgeGap struct {
	ConceptID     uuid.UUID `json:"concept_id"`
	ConceptName   string    `json:"concept_name"`
	CurrentLevel  float64   `json:"current_level"`
	RequiredLevel float64   `json:"required_level"`
	GapSize       float64   `json:"gap_size"`
	Priority      string    `json:"priority"`
	Category      string    `json:"category"`
}

// GapRecommendation 
type GapRecommendation struct {
	GapID               uuid.UUID     `json:"gap_id"`
	RecommendationType  string        `json:"recommendation_type"`
	ContentID           *uuid.UUID    `json:"content_id,omitempty"`
	Description         string        `json:"description"`
	EstimatedTime       time.Duration `json:"estimated_time"`
	Priority            string        `json:"priority"`
	ExpectedImprovement float64       `json:"expected_improvement"`
}

// ErrorResponse 
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// 

func parseCommaSeparated(str string) []string {
	if str == "" {
		return []string{}
	}

	var result []string
	for _, item := range strings.Split(str, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

