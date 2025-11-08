package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/service"
)

// ModelHandler 模型处理器
type ModelHandler struct {
	modelDAO      *dao.ModelDAO
	modelManager  *service.ModelManager
}

// NewModelHandler 创建模型处理器
func NewModelHandler(modelDAO *dao.ModelDAO, modelManager *service.ModelManager) *ModelHandler {
	return &ModelHandler{
		modelDAO:     modelDAO,
		modelManager: modelManager,
	}
}

// CreateModelConfig 创建模型配置
func (h *ModelHandler) CreateModelConfig(c *gin.Context) {
	var req models.ModelConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置创建时间
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	// 创建配置
	if err := h.modelDAO.CreateModelConfig(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建服务
	if _, err := h.modelManager.CreateService(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetModelConfig 获取模型配置
func (h *ModelHandler) GetModelConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	config, err := h.modelDAO.GetModelConfig(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// ListModelConfigs 列出模型配置
func (h *ModelHandler) ListModelConfigs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	configs, err := h.modelDAO.ListModelConfigs(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// UpdateModelConfig 更新模型配置
func (h *ModelHandler) UpdateModelConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.ModelConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置ID和更新时间
	req.ID = uint(id)
	req.UpdatedAt = time.Now()

	// 更新配置
	if err := h.modelDAO.UpdateModelConfig(c.Request.Context(), uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新服务
	if err := h.modelManager.UpdateConfig(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, req)
}

// DeleteModelConfig 删除模型配置
func (h *ModelHandler) DeleteModelConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// 获取配置
	config, err := h.modelDAO.GetModelConfig(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 删除服务
	if err := h.modelManager.RemoveService(config.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 删除配置
	if err := h.modelDAO.DeleteModelConfig(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model config deleted successfully"})
}

// ListServices 列出所有服务
func (h *ModelHandler) ListServices(c *gin.Context) {
	services := h.modelManager.ListServices()
	c.JSON(http.StatusOK, gin.H{"services": services})
}

// HealthCheck 健康检查
func (h *ModelHandler) HealthCheck(c *gin.Context) {
	results := h.modelManager.HealthCheck(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// GenerateText 生成文本
func (h *ModelHandler) GenerateText(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		// 使用默认服务
		_, err := h.modelManager.GetDefaultService()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		serviceName = "default"
	} else {
		// 使用指定的服务
		_, err := h.modelManager.GetService(serviceName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	var req models.TextGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc, _ := h.modelManager.GetService(serviceName)
	response, err := svc.GenerateText(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateEmbeddings 生成嵌入
func (h *ModelHandler) GenerateEmbeddings(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		// 使用默认服务
		_, err := h.modelManager.GetDefaultService()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		serviceName = "default"
	} else {
		// 使用指定的服务
		_, err := h.modelManager.GetService(serviceName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	var req models.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为EmbeddingsRequest类型
	embeddingsReq := &models.EmbeddingsRequest{
		Model: req.Model,
		Texts: req.Input,
		User:  req.User,
	}

	svc, _ := h.modelManager.GetService(serviceName)
	response, err := svc.GenerateEmbeddings(c.Request.Context(), embeddingsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateConversation 创建对话
func (h *ModelHandler) CreateConversation(c *gin.Context) {
	var req models.Conversation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置创建时间
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := h.modelDAO.CreateConversation(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetConversation 获取对话
func (h *ModelHandler) GetConversation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	conversation, err := h.modelDAO.GetConversation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// ListConversations 列出对话
func (h *ModelHandler) ListConversations(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	conversations, err := h.modelDAO.ListConversations(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// DeleteConversation 删除对话
func (h *ModelHandler) DeleteConversation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	if err := h.modelDAO.DeleteConversation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted successfully"})
}

// CreateMessage 创建消息
func (h *ModelHandler) CreateMessage(c *gin.Context) {
	var req models.Message
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置创建时间
	req.CreatedAt = time.Now()

	if err := h.modelDAO.CreateMessage(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetMessages 获取消息列表
func (h *ModelHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversation ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	messages, err := h.modelDAO.GetMessages(c.Request.Context(), conversationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// CreateFineTuningJob 创建微调作业
func (h *ModelHandler) CreateFineTuningJob(c *gin.Context) {
	var req models.FineTuningJob
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置创建时间
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := h.modelDAO.CreateFineTuningJob(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetFineTuningJob 获取微调作业
func (h *ModelHandler) GetFineTuningJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	job, err := h.modelDAO.GetFineTuningJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListFineTuningJobs 列出微调作业
func (h *ModelHandler) ListFineTuningJobs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	jobs, err := h.modelDAO.ListFineTuningJobs(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}