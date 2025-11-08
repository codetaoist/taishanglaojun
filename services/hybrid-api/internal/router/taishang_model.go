package router

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/config"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/middleware"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/response"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupTaishangModelRoutes 设置太上域模型路由
func SetupTaishangModelRoutes(cfg *config.Config, r *gin.Engine, modelManager *service.ModelManager) {
	// Initialize DAOs
	gormDB, err := dao.NewGormDB(cfg.DB)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize GORM DB: %v", err))
	}
	
	modelDAO := dao.NewModelDAO(gormDB)
	
	// Initialize conversation service
	conversationService := service.NewConversationService(modelDAO)
	
	// Initialize vector DAO
	modelService, err := modelManager.GetDefaultService()
	if err != nil {
		log.Fatalf("Failed to get default model service: %v", err)
	}
	vectorDAO := dao.NewVectorDAO(gormDB, modelService)
	
	// Initialize conversation vector service
	conversationVectorService := service.NewConversationVectorService(
		modelDAO.GetConversationDAO(),
		modelDAO.GetMessageDAO(),
		vectorDAO,
		modelService,
	)
	
	// Initialize conversation import/export service
	conversationImportExportService := service.NewConversationImportExportService(
		modelDAO.GetConversationDAO(),
		modelDAO.GetMessageDAO(),
	)
	
	// Apply authentication middleware
	authMiddleware := middleware.Auth(*cfg)

	taishang := r.Group("/api/v1/taishang")
	taishang.Use(authMiddleware)
	
	// Model service management
	{
		// 列出所有可用的模型服务
		taishang.GET("/model/services", listModelServices(modelManager))
		
		// 创建模型服务配置
		taishang.POST("/model/services", createModelService(modelDAO, modelManager))
		
		// 获取模型服务配置
		taishang.GET("/model/services/:id", getModelService(modelDAO))
		
		// 更新模型服务配置
		taishang.PUT("/model/services/:id", updateModelService(modelDAO, modelManager))
		
		// 删除模型服务配置
		taishang.DELETE("/model/services/:id", deleteModelService(modelDAO, modelManager))
		
		// 健康检查所有模型服务
		taishang.GET("/model/services/health", healthCheckModelServices(modelManager))
		
		// 列出指定服务的所有模型
		taishang.GET("/model/service/:serviceName/models", listServiceModels(modelManager))
		
		// 获取指定服务的模型详情
		taishang.GET("/model/service/:serviceName/models/:modelId", getServiceModel(modelManager))
	}
	
	// Text generation
	{
		// 生成文本（非流式）
		taishang.POST("/model/generate", generateText(modelManager))
		
		// 生成文本（流式）
		taishang.POST("/model/generate/stream", generateTextStream(modelManager))
	}
	
	// Embedding generation
	{
		// 生成单个文本的嵌入向量
		taishang.POST("/model/embeddings", generateEmbedding(modelManager))
		
		// 批量生成文本的嵌入向量
		taishang.POST("/model/embeddings/batch", generateEmbeddings(modelManager))
	}
	
	// Conversation management
	{
		// 创建对话
		taishang.POST("/conversations", createConversation(conversationService))
		
		// 获取对话详情
		taishang.GET("/conversations/:id", getConversation(conversationService))
		
		// 列出用户的对话
		taishang.GET("/conversations", listConversations(conversationService))
		
		// 更新对话
		taishang.PUT("/conversations/:id", updateConversation(conversationService))
		
		// 删除对话
		taishang.DELETE("/conversations/:id", deleteConversation(conversationService))
		
		// 添加消息到对话
		taishang.POST("/conversations/:id/messages", addMessage(conversationService))
		
		// 获取对话中的消息
		taishang.GET("/conversations/:id/messages", getMessages(conversationService))
		
		// 删除对话中的消息
		taishang.DELETE("/conversations/:id/messages/:messageId", deleteMessage(conversationService))
		
		// 对话向量搜索
		taishang.POST("/conversations/search", searchConversations(conversationVectorService))
		
		// 在特定对话中搜索消息
		taishang.POST("/conversations/:id/messages/search", searchInConversation(conversationVectorService))
		
		// 将对话添加到向量索引
		taishang.POST("/conversations/:id/index", indexConversation(conversationVectorService))
		
		// 从向量索引中移除对话
		taishang.DELETE("/conversations/:id/index", removeFromIndex(conversationVectorService))
		
		// 重新索引对话
		taishang.POST("/conversations/:id/reindex", reindexConversation(conversationVectorService))
		
		// 导出对话
		taishang.POST("/conversations/export", exportConversations(conversationImportExportService))
		
		// 导出单个对话
		taishang.POST("/conversations/:id/export", exportConversation(conversationImportExportService))
		
		// 导入对话
		taishang.POST("/conversations/import", importConversations(conversationImportExportService))
	}
}

// Model service handlers

// listModelServices 列出所有模型服务
func listModelServices(modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		services := modelManager.ListServices()
		response.Success(c, gin.H{"services": services})
	}
}

// createModelService 创建模型服务配置
func createModelService(modelDAO *dao.ModelDAO, modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var config models.ModelConfig
		if err := c.ShouldBindJSON(&config); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		// 创建模型配置
		if err := modelDAO.CreateModelConfig(c.Request.Context(), &config); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		// 创建并注册模型服务
		service, err := modelManager.CreateService(&config)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		// 连接服务
		if err := service.Connect(c.Request.Context(), &config); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, config)
	}
}

// getModelService 获取模型服务配置
func getModelService(modelDAO *dao.ModelDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			response.BadRequest(c, "Service ID is required")
			return
		}
		
		// 转换ID为uint类型
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "Invalid ID format")
			return
		}
		
		config, err := modelDAO.GetModelConfig(c.Request.Context(), uint(id))
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}
		
		response.Success(c, config)
	}
}

// updateModelService 更新模型服务配置
func updateModelService(modelDAO *dao.ModelDAO, modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			response.BadRequest(c, "Service ID is required")
			return
		}
		
		// 转换ID为uint类型
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "Invalid ID format")
			return
		}
		
		var config models.ModelConfig
		if err := c.ShouldBindJSON(&config); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		// 确保ID匹配
		config.ID = uint(id)
		
		// 更新模型配置
		if err := modelDAO.UpdateModelConfig(c.Request.Context(), config.ID, &config); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		// 更新服务配置
		if err := modelManager.UpdateConfig(&config); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, config)
	}
}

// deleteModelService 删除模型服务配置
func deleteModelService(modelDAO *dao.ModelDAO, modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			response.BadRequest(c, "Service ID is required")
			return
		}
		
		// 转换ID为uint类型
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "Invalid ID format")
			return
		}
		
		// 获取配置以获取服务名称
		config, err := modelDAO.GetModelConfig(c.Request.Context(), uint(id))
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}
		
		// 删除服务
		if err := modelManager.RemoveService(config.Name); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		// 删除配置
		if err := modelDAO.DeleteModelConfig(c.Request.Context(), uint(id)); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, gin.H{"message": "Model service deleted successfully"})
	}
}

// healthCheckModelServices 健康检查所有模型服务
func healthCheckModelServices(modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		results := modelManager.HealthCheck(c.Request.Context())
		response.Success(c, gin.H{"results": results})
	}
}

// listServiceModels 列出指定服务的所有模型
func listServiceModels(modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("serviceName")
		if serviceName == "" {
			response.BadRequest(c, "Service name is required")
			return
		}
		
		service, err := modelManager.GetService(serviceName)
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}
		
		models, err := service.ListModels(c.Request.Context())
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, gin.H{"models": models})
	}
}

// getServiceModel 获取指定服务的模型详情
func getServiceModel(modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("serviceName")
		modelId := c.Param("modelId")
		
		if serviceName == "" || modelId == "" {
			response.BadRequest(c, "Service name and model ID are required")
			return
		}
		
		service, err := modelManager.GetService(serviceName)
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}
		
		model, err := service.GetModel(c.Request.Context(), modelId)
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}
		
		response.Success(c, model)
	}
}

// Text generation handlers

// generateText 生成文本（非流式）
func generateText(modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.TextGenerationRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		var service service.ModelService
		var err error
		
		// 获取服务
		if request.Model != "" {
			// 使用模型名称获取服务
			service, err = modelManager.GetServiceByModel(request.Model)
			if err != nil {
				// 如果找不到特定模型的服务，尝试使用默认服务
				service, err = modelManager.GetDefaultService()
				if err != nil {
					response.NotFound(c, "No service available for model and no default service")
					return
				}
			}
		} else {
			// 使用默认服务
			service, err = modelManager.GetDefaultService()
			if err != nil {
				response.NotFound(c, "No default service available")
				return
			}
		}
		
		// 生成文本
		resp, err := service.GenerateText(c.Request.Context(), &request)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, resp)
	}
}

// generateTextStream 生成文本（流式）
func generateTextStream(modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.TextGenerationRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		var service service.ModelService
		var err error
		
		// 获取服务
		if request.Model != "" {
			// 使用模型名称获取服务
			service, err = modelManager.GetDefaultService()
			if err != nil {
				response.NotFound(c, "No default service available")
				return
			}
		} else {
			// 使用默认服务
			service, err = modelManager.GetDefaultService()
			if err != nil {
				response.NotFound(c, "No default service available")
				return
			}
		}
		
		// 生成流式文本
		chunkChan, err := service.GenerateTextStream(c.Request.Context(), &request)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		// 设置流式响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")
		
		// 发送流式响应
		c.Stream(func(w io.Writer) bool {
			for chunk := range chunkChan {
				// 发送数据块
				c.SSEvent("", chunk)
				if chunk.Done {
					return false
				}
			}
			return false
		})
	}
}

// Embedding generation handlers

// generateEmbedding 生成嵌入向量（单个）
func generateEmbedding(modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.EmbeddingRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		var service service.ModelService
		var err error
		
		// 获取服务
		if request.Model != "" {
			// 使用模型名称获取服务
			service, err = modelManager.GetDefaultService()
			if err != nil {
				response.NotFound(c, "No default service available")
				return
			}
		} else {
			// 使用默认服务
			service, err = modelManager.GetDefaultService()
			if err != nil {
				response.NotFound(c, "No default service available")
				return
			}
		}
		
		// 生成嵌入向量
		resp, err := service.GenerateEmbedding(c.Request.Context(), &request)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, resp)
	}
}

// generateEmbeddings 生成嵌入向量（批量）
func generateEmbeddings(modelManager *service.ModelManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.EmbeddingsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		var service service.ModelService
		var err error
		
		// 获取服务
		if request.Model != "" {
			// 使用模型名称获取服务
			service, err = modelManager.GetDefaultService()
			if err != nil {
				response.NotFound(c, "No default service available")
				return
			}
		} else {
			// 使用默认服务
			service, err = modelManager.GetDefaultService()
			if err != nil {
				response.NotFound(c, "No default service available")
				return
			}
		}
		
		// 生成嵌入向量
		resp, err := service.GenerateEmbeddings(c.Request.Context(), &request)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, resp)
	}
}

// Conversation management handlers

// createConversation 创建对话
func createConversation(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateConversationRequest
		
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		// 从上下文获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			return
		}
		req.UserID = userID.(string)
		
		// 如果没有提供标题，使用默认标题
		if req.Title == "" {
			req.Title = "New Conversation"
		}
		
		// 使用对话服务创建对话
		conversation, err := conversationService.CreateConversation(c.Request.Context(), &req)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, conversation)
	}
}

// getConversation 获取对话详情
func getConversation(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		
		// 使用对话服务获取对话
		conversation, err := conversationService.GetConversation(c.Request.Context(), conversationID)
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}
		
		response.Success(c, conversation)
	}
}

// listConversations 列出用户的对话
func listConversations(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			return
		}
		
		// 获取分页参数
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		// 使用对话服务获取对话列表
		conversations, err := conversationService.ListConversations(c.Request.Context(), userID.(string), limit, offset)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, gin.H{
			"conversations": conversations,
			"limit":         limit,
			"offset":        offset,
		})
	}
}

// updateConversation 更新对话
func updateConversation(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		var req models.UpdateConversationRequest
		
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		// 使用对话服务更新对话
		conversation, err := conversationService.UpdateConversation(c.Request.Context(), conversationID, &req)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, conversation)
	}
}

// deleteConversation 删除对话
func deleteConversation(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		
		// 使用对话服务删除对话
		if err := conversationService.DeleteConversation(c.Request.Context(), conversationID); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, gin.H{"message": "Conversation deleted successfully"})
	}
}

// addMessage 添加消息到对话
func addMessage(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.AddMessageRequest
		
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		
		// 验证角色
		validRoles := []string{"user", "assistant", "system"}
		isValidRole := false
		for _, role := range validRoles {
			if req.Role == role {
				isValidRole = true
				break
			}
		}
		if !isValidRole {
			response.BadRequest(c, "Invalid role. Must be one of: user, assistant, system")
			return
		}
		
		// 设置对话ID
		conversationID := c.Param("id")
		req.ConversationID = conversationID
		
		// 使用对话服务添加消息
		message, err := conversationService.AddMessage(c.Request.Context(), &req)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, message)
	}
}

// getMessages 获取对话的消息列表
func getMessages(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		
		// 获取分页参数
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		// 使用对话服务获取消息列表
		messages, err := conversationService.GetMessages(c.Request.Context(), conversationID, limit, offset)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, gin.H{
			"messages": messages,
			"limit":    limit,
			"offset":   offset,
		})
	}
}

// deleteMessage 删除消息
func deleteMessage(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("messageId")
		
		// 使用对话服务删除消息
		if err := conversationService.DeleteMessage(c.Request.Context(), messageID); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		
		response.Success(c, gin.H{"message": "Message deleted successfully"})
	}
}

// Helper functions

// generateConversationID 生成对话ID
func generateConversationID() string {
	return fmt.Sprintf("conv_%d", time.Now().UnixNano())
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}