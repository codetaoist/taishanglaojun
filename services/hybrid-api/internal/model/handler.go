package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/services/api/internal/middleware"
)

// ModelHandler handles model-related HTTP requests
type ModelHandler struct {
	manager *ModelManager
	logger  *middleware.Logger
}

// NewModelHandler creates a new ModelHandler
func NewModelHandler(manager *ModelManager, logger *middleware.Logger) *ModelHandler {
	return &ModelHandler{
		manager: manager,
		logger:  logger,
	}
}

// RegisterRoutes registers the model routes
func (h *ModelHandler) RegisterRoutes(router gin.IRouter) {
	models := router.Group("/models")
	{
		models.GET("", h.ListModels)
		models.GET("/:id", h.GetModel)
		models.POST("", h.RegisterModel)
		models.PUT("/:id", h.UpdateModel)
		models.DELETE("/:id", h.UnregisterModel)
		
		// Model loading
		models.POST("/:id/load", h.LoadModel)
		models.POST("/:id/unload", h.UnloadModel)
		models.GET("/:id/loaded", h.IsModelLoaded)
		
		// Model operations
		models.POST("/:id/generate", h.GenerateText)
		models.POST("/:id/embed", h.GenerateEmbedding)
		models.POST("/:id/classify", h.ClassifyText)
		models.POST("/:id/summarize", h.SummarizeText)
		models.POST("/:id/translate", h.TranslateText)
		models.POST("/:id/multimodal", h.ProcessMultimodal)
	}
}

// ListModels handles the list models request
func (h *ModelHandler) ListModels(c *gin.Context) {
	ctx := c.Request.Context()
	
	models, err := h.manager.ListModels(ctx)
	if err != nil {
		h.logger.Errorf("Failed to list models: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list models",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"count":  len(models),
	})
}

// GetModel handles the get model request
func (h *ModelHandler) GetModel(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	model, err := h.manager.GetModel(ctx, modelID)
	if err != nil {
		h.logger.Errorf("Failed to get model %s: %v", modelID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Model not found",
			"details": err.Error(),
		})
		return
	}
	
	// Check if model is loaded
	loaded := h.manager.IsModelLoaded(ctx, modelID)
	
	c.JSON(http.StatusOK, gin.H{
		"model": model,
		"loaded": loaded,
	})
}

// RegisterModel handles the register model request
func (h *ModelHandler) RegisterModel(c *gin.Context) {
	ctx := c.Request.Context()
	
	var model ModelInfo
	if err := c.ShouldBindJSON(&model); err != nil {
		h.logger.Errorf("Failed to bind model: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	if err := h.manager.RegisterModel(ctx, &model); err != nil {
		h.logger.Errorf("Failed to register model %s: %v", model.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register model",
			"details": err.Error(),
		})
		return
	}
	
	h.logger.Infof("Registered model %s (%s)", model.ID, model.Name)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Model registered successfully",
		"model": model,
	})
}

// UpdateModel handles the update model request
func (h *ModelHandler) UpdateModel(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	// Check if model exists
	_, err := h.manager.GetModel(ctx, modelID)
	if err != nil {
		h.logger.Errorf("Model %s not found: %v", modelID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Model not found",
			"details": err.Error(),
		})
		return
	}
	
	var update ModelInfo
	if err := c.ShouldBindJSON(&update); err != nil {
		h.logger.Errorf("Failed to bind model update: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Ensure the ID matches the URL parameter
	update.ID = modelID
	
	// Unregister the old model and register the updated one
	if err := h.manager.UnregisterModel(ctx, modelID); err != nil {
		h.logger.Errorf("Failed to unregister model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update model",
			"details": err.Error(),
		})
		return
	}
	
	if err := h.manager.RegisterModel(ctx, &update); err != nil {
		h.logger.Errorf("Failed to register updated model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update model",
			"details": err.Error(),
		})
		return
	}
	
	h.logger.Infof("Updated model %s", modelID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Model updated successfully",
		"model": update,
	})
}

// UnregisterModel handles the unregister model request
func (h *ModelHandler) UnregisterModel(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	if err := h.manager.UnregisterModel(ctx, modelID); err != nil {
		h.logger.Errorf("Failed to unregister model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unregister model",
			"details": err.Error(),
		})
		return
	}
	
	h.logger.Infof("Unregistered model %s", modelID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Model unregistered successfully",
	})
}

// LoadModel handles the load model request
func (h *ModelHandler) LoadModel(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	if err := h.manager.LoadModel(ctx, modelID); err != nil {
		h.logger.Errorf("Failed to load model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load model",
			"details": err.Error(),
		})
		return
	}
	
	h.logger.Infof("Loaded model %s", modelID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Model loaded successfully",
	})
}

// UnloadModel handles the unload model request
func (h *ModelHandler) UnloadModel(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	if err := h.manager.UnloadModel(ctx, modelID); err != nil {
		h.logger.Errorf("Failed to unload model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unload model",
			"details": err.Error(),
		})
		return
	}
	
	h.logger.Infof("Unloaded model %s", modelID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Model unloaded successfully",
	})
}

// IsModelLoaded handles the is model loaded request
func (h *ModelHandler) IsModelLoaded(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	loaded := h.manager.IsModelLoaded(ctx, modelID)
	
	c.JSON(http.StatusOK, gin.H{
		"model_id": modelID,
		"loaded": loaded,
	})
}

// GenerateText handles the generate text request
func (h *ModelHandler) GenerateText(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	var request TextGenerationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("Failed to bind text generation request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Set the model ID from the URL parameter
	request.ModelID = modelID
	
	response, err := h.manager.GenerateText(ctx, &request)
	if err != nil {
		h.logger.Errorf("Failed to generate text with model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate text",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GenerateEmbedding handles the generate embedding request
func (h *ModelHandler) GenerateEmbedding(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	var request EmbeddingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("Failed to bind embedding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Set the model ID from the URL parameter
	request.ModelID = modelID
	
	response, err := h.manager.GenerateEmbedding(ctx, &request)
	if err != nil {
		h.logger.Errorf("Failed to generate embedding with model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate embedding",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// ClassifyText handles the classify text request
func (h *ModelHandler) ClassifyText(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	var request ClassificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("Failed to bind classification request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Set the model ID from the URL parameter
	request.ModelID = modelID
	
	response, err := h.manager.ClassifyText(ctx, &request)
	if err != nil {
		h.logger.Errorf("Failed to classify text with model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to classify text",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// SummarizeText handles the summarize text request
func (h *ModelHandler) SummarizeText(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	var request SummarizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("Failed to bind summarization request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Set the model ID from the URL parameter
	request.ModelID = modelID
	
	response, err := h.manager.SummarizeText(ctx, &request)
	if err != nil {
		h.logger.Errorf("Failed to summarize text with model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to summarize text",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// TranslateText handles the translate text request
func (h *ModelHandler) TranslateText(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	var request TranslationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("Failed to bind translation request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Set the model ID from the URL parameter
	request.ModelID = modelID
	
	response, err := h.manager.TranslateText(ctx, &request)
	if err != nil {
		h.logger.Errorf("Failed to translate text with model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to translate text",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// ProcessMultimodal handles the process multimodal request
func (h *ModelHandler) ProcessMultimodal(c *gin.Context) {
	ctx := c.Request.Context()
	modelID := c.Param("id")
	
	var request MultimodalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("Failed to bind multimodal request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Set the model ID from the URL parameter
	request.ModelID = modelID
	
	response, err := h.manager.ProcessMultimodal(ctx, &request)
	if err != nil {
		h.logger.Errorf("Failed to process multimodal data with model %s: %v", modelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process multimodal data",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// ModelRouter configures the model API routes
type ModelRouter struct {
	handler *ModelHandler
}

// NewModelRouter creates a new ModelRouter
func NewModelRouter(handler *ModelHandler) *ModelRouter {
	return &ModelRouter{
		handler: handler,
	}
}

// SetupRoutes configures the model API routes
func (r *ModelRouter) SetupRoutes(router gin.IRouter) {
	r.handler.RegisterRoutes(router)
}

// ModelAPIHandler combines all model-related handlers
type ModelAPIHandler struct {
	router *ModelRouter
	logger *middleware.Logger
}

// NewModelAPIHandler creates a new ModelAPIHandler
func NewModelAPIHandler(manager *ModelManager, logger *middleware.Logger) *ModelAPIHandler {
	handler := NewModelHandler(manager, logger)
	router := NewModelRouter(handler)
	
	return &ModelAPIHandler{
		router: router,
		logger: logger,
	}
}

// RegisterRoutes registers all model API routes
func (h *ModelAPIHandler) RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/api/v1")
	h.router.SetupRoutes(api)
}

// Helper functions for request validation

// ValidateTextGenerationRequest validates a text generation request
func ValidateTextGenerationRequest(request *TextGenerationRequest) error {
	if request.ModelID == "" {
		return fmt.Errorf("model_id is required")
	}
	if request.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if request.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}
	if request.Temperature < 0 || request.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	if request.TopP <= 0 || request.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1")
	}
	return nil
}

// ValidateEmbeddingRequest validates an embedding request
func ValidateEmbeddingRequest(request *EmbeddingRequest) error {
	if request.ModelID == "" {
		return fmt.Errorf("model_id is required")
	}
	if len(request.Texts) == 0 {
		return fmt.Errorf("texts is required and must not be empty")
	}
	return nil
}

// ValidateClassificationRequest validates a classification request
func ValidateClassificationRequest(request *ClassificationRequest) error {
	if request.ModelID == "" {
		return fmt.Errorf("model_id is required")
	}
	if request.Text == "" {
		return fmt.Errorf("text is required")
	}
	return nil
}

// ValidateSummarizationRequest validates a summarization request
func ValidateSummarizationRequest(request *SummarizationRequest) error {
	if request.ModelID == "" {
		return fmt.Errorf("model_id is required")
	}
	if request.Text == "" {
		return fmt.Errorf("text is required")
	}
	if request.MaxLength > 0 && request.MinLength > 0 && request.MaxLength < request.MinLength {
		return fmt.Errorf("max_length must be greater than or equal to min_length")
	}
	return nil
}

// ValidateTranslationRequest validates a translation request
func ValidateTranslationRequest(request *TranslationRequest) error {
	if request.ModelID == "" {
		return fmt.Errorf("model_id is required")
	}
	if request.Text == "" {
		return fmt.Errorf("text is required")
	}
	if request.SourceLang == "" {
		return fmt.Errorf("source_lang is required")
	}
	if request.TargetLang == "" {
		return fmt.Errorf("target_lang is required")
	}
	if request.SourceLang == request.TargetLang {
		return fmt.Errorf("source_lang and target_lang must be different")
	}
	return nil
}

// ValidateMultimodalRequest validates a multimodal request
func ValidateMultimodalRequest(request *MultimodalRequest) error {
	if request.ModelID == "" {
		return fmt.Errorf("model_id is required")
	}
	
	// At least one input type must be provided
	if request.Text == "" && request.ImageURL == "" && request.AudioURL == "" && request.VideoURL == "" {
		return fmt.Errorf("at least one of text, image_url, audio_url, or video_url must be provided")
	}
	
	return nil
}

// ParsePaginationParams parses pagination parameters from query string
func ParsePaginationParams(c *gin.Context) (int, int, error) {
	page := 1
	limit := 20
	
	// Parse page
	if pageStr := c.Query("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			return 0, 0, fmt.Errorf("invalid page parameter")
		}
		page = p
	}
	
	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 || l > 100 {
			return 0, 0, fmt.Errorf("invalid limit parameter (must be between 1 and 100)")
		}
		limit = l
	}
	
	return page, limit, nil
}

// ParseFilterParams parses filter parameters from query string
func ParseFilterParams(c *gin.Context) (map[string]string, error) {
	filters := make(map[string]string)
	
	// Parse model type filter
	if modelType := c.Query("type"); modelType != "" {
		switch ModelType(modelType) {
		case ModelTypeTextGeneration, ModelTypeEmbedding, ModelTypeClassification, 
		     ModelTypeSummarization, ModelTypeTranslation, ModelTypeMultimodal:
			filters["type"] = modelType
		default:
			return nil, fmt.Errorf("invalid model type: %s", modelType)
		}
	}
	
	// Parse tags filter
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		filters["tags"] = strings.Join(tags, ",")
	}
	
	return filters, nil
}

// FilterModels filters models based on the provided filters
func FilterModels(models []*ModelInfo, filters map[string]string) []*ModelInfo {
	var result []*ModelInfo
	
	for _, model := range models {
		include := true
		
		// Filter by type
		if modelType, ok := filters["type"]; ok {
			if string(model.Type) != modelType {
				include = false
			}
		}
		
		// Filter by tags
		if tagsStr, ok := filters["tags"]; ok && include {
			requiredTags := strings.Split(tagsStr, ",")
			for _, requiredTag := range requiredTags {
				requiredTag = strings.TrimSpace(requiredTag)
				found := false
				for _, tag := range model.Tags {
					if tag == requiredTag {
						found = true
						break
					}
				}
				if !found {
					include = false
					break
				}
			}
		}
		
		if include {
			result = append(result, model)
		}
	}
	
	return result
}

// PaginateModels paginates the models
func PaginateModels(models []*ModelInfo, page, limit int) ([]*ModelInfo, int) {
	total := len(models)
	start := (page - 1) * limit
	end := start + limit
	
	if start >= total {
		return []*ModelInfo{}, total
	}
	
	if end > total {
		end = total
	}
	
	return models[start:end], total
}