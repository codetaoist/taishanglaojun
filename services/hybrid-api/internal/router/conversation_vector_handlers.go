package router

import (
	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/response"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/service"
)

// searchConversations 对话向量搜索
func searchConversations(conversationVectorService *service.ConversationVectorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Query string `json:"query" binding:"required"`
			TopK  int    `json:"topK"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		// 设置默认值
		if req.TopK <= 0 {
			req.TopK = 10
		}

		// 从上下文获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			return
		}
		
		// TODO: 实现用户权限检查，确保只返回属于当前用户的对话
		_ = userID

		// 执行搜索
		results, err := conversationVectorService.SearchConversations(c.Request.Context(), req.Query, req.TopK)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}

		// 过滤结果，只返回属于当前用户的对话
		filteredResults := make([]models.ConversationSearchResult, 0)
		for _, result := range results {
			// 这里需要检查对话是否属于当前用户
			// 由于SearchConversations方法没有返回用户ID，我们需要从数据库获取对话详情
			filteredResults = append(filteredResults, result)
		}

		response.Success(c, filteredResults)
	}
}

// searchInConversation 在特定对话中搜索消息
func searchInConversation(conversationVectorService *service.ConversationVectorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		if conversationID == "" {
			response.BadRequest(c, "Conversation ID is required")
			return
		}

		var req struct {
			Query string `json:"query" binding:"required"`
			TopK  int    `json:"topK"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		// 设置默认值
		if req.TopK <= 0 {
			req.TopK = 10
		}

		// 执行搜索
		results, err := conversationVectorService.SearchInConversation(c.Request.Context(), conversationID, req.Query, req.TopK)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}

		response.Success(c, results)
	}
}

// indexConversation 将对话添加到向量索引
func indexConversation(conversationVectorService *service.ConversationVectorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		if conversationID == "" {
			response.BadRequest(c, "Conversation ID is required")
			return
		}

		// 执行索引
		err := conversationVectorService.IndexConversation(c.Request.Context(), conversationID)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "Conversation indexed successfully"})
	}
}

// removeFromIndex 从向量索引中移除对话
func removeFromIndex(conversationVectorService *service.ConversationVectorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		if conversationID == "" {
			response.BadRequest(c, "Conversation ID is required")
			return
		}

		// 执行移除
		err := conversationVectorService.RemoveFromIndex(c.Request.Context(), conversationID)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "Conversation removed from index successfully"})
	}
}

// reindexConversation 重新索引对话
func reindexConversation(conversationVectorService *service.ConversationVectorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		if conversationID == "" {
			response.BadRequest(c, "Conversation ID is required")
			return
		}

		// 执行重新索引
		err := conversationVectorService.ReindexConversation(c.Request.Context(), conversationID)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "Conversation reindexed successfully"})
	}
}