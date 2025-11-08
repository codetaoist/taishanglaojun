package service

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// ConversationVectorService 对话向量搜索服务
type ConversationVectorService struct {
	conversationDAO *dao.ConversationDAO
	messageDAO      *dao.MessageDAO
	vectorDAO       *dao.VectorDAO
	modelService    ModelService
}

// NewConversationVectorService 创建对话向量搜索服务
func NewConversationVectorService(conversationDAO *dao.ConversationDAO, messageDAO *dao.MessageDAO, vectorDAO *dao.VectorDAO, modelService ModelService) *ConversationVectorService {
	return &ConversationVectorService{
		conversationDAO: conversationDAO,
		messageDAO:      messageDAO,
		vectorDAO:       vectorDAO,
		modelService:    modelService,
	}
}

// IndexConversation 将对话内容索引到向量数据库
func (s *ConversationVectorService) IndexConversation(ctx context.Context, conversationID string) error {
	// 获取对话及其消息
	conversation, err := s.conversationDAO.GetConversation(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	if conversation == nil {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	// 获取对话的所有消息
	messages, err := s.messageDAO.GetMessagesByConversationID(ctx, conversationID, 0, 0) // 获取所有消息
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		return nil // 没有消息需要索引
	}

	// 为每条消息创建向量
		for _, message := range messages {
			if err := s.indexMessage(ctx, conversationID, *message); err != nil {
			// 记录错误但继续处理其他消息
			fmt.Printf("Failed to index message %s: %v\n", message.ID, err)
		}
	}

	return nil
}

// indexMessage 索引单个消息到向量数据库
func (s *ConversationVectorService) indexMessage(ctx context.Context, conversationID string, message models.Message) error {
	// 生成消息内容的向量
	embeddingReq := &models.EmbeddingRequest{
		Model: "text-embedding-ada-002", // 使用默认的嵌入模型
		Input: []string{message.Content},
	}

	embeddingResp, err := s.modelService.GenerateEmbedding(ctx, embeddingReq)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(embeddingResp.Embedding) == 0 {
		return fmt.Errorf("empty embedding generated")
	}

	// 准备向量数据
	vectorData := models.VectorData{
		ID:     fmt.Sprintf("%s:%s", conversationID, message.ID),
		Vector: embeddingResp.Embedding,
		Metadata: map[string]interface{}{
			"conversation_id": conversationID,
			"message_id":      message.ID,
			"role":            message.Role,
			"content":         message.Content,
		},
	}

	// 插入或更新向量到向量数据库
	upsertReq := &models.UpsertVectorsRequest{
		CollectionName: "conversations",
		Vectors:        []models.VectorData{vectorData},
	}

	_, err = s.vectorDAO.UpsertVectors(ctx, upsertReq)
	if err != nil {
		return fmt.Errorf("failed to upsert vector: %w", err)
	}

	return nil
}

// SearchConversations 基于向量搜索相关对话
func (s *ConversationVectorService) SearchConversations(ctx context.Context, query string, topK int) ([]models.ConversationSearchResult, error) {
	// 生成查询向量
	embeddingReq := &models.EmbeddingRequest{
		Model: "text-embedding-ada-002", // 使用默认的嵌入模型
		Input: []string{query},
	}

	embeddingResp, err := s.modelService.GenerateEmbedding(ctx, embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddingResp.Embedding) == 0 {
		return nil, fmt.Errorf("empty query embedding generated")
	}

	// 在向量数据库中搜索
	searchReq := &models.SearchRequest{
		CollectionName: "conversations",
		QueryVector:    embeddingResp.Embedding,
		TopK:           topK,
	}

	searchResp, err := s.vectorDAO.SearchVectorData(ctx, "conversations", searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// 转换搜索结果
	results := make([]models.ConversationSearchResult, 0, len(searchResp.Results))
	conversationMap := make(map[string]models.ConversationSearchResult)

	for _, result := range searchResp.Results {
		// 从元数据中提取对话ID
		conversationID, ok := result.Metadata["conversation_id"].(string)
		if !ok {
			continue
		}

		// 如果已经处理过该对话，则合并结果
		if searchResult, exists := conversationMap[conversationID]; exists {
			// 更新分数（取最高分）
			if result.Score > searchResult.Score {
				searchResult.Score = result.Score
			}
			// 添加匹配的消息片段
			messageID, _ := result.Metadata["message_id"].(string)
			role, _ := result.Metadata["role"].(string)
			content, _ := result.Metadata["content"].(string)
			
			searchResult.MatchedMessages = append(searchResult.MatchedMessages, models.MatchedMessage{
				ID:      messageID,
				Role:    role,
				Content: content,
				Score:   result.Score,
			})
			conversationMap[conversationID] = searchResult
		} else {
			// 获取对话详情
			conversation, err := s.conversationDAO.GetConversation(ctx, conversationID)
			if err != nil {
				fmt.Printf("Failed to get conversation %s: %v\n", conversationID, err)
				continue
			}

			if conversation == nil {
				continue
			}

			// 创建新的搜索结果
			messageID, _ := result.Metadata["message_id"].(string)
			role, _ := result.Metadata["role"].(string)
			content, _ := result.Metadata["content"].(string)
			
			searchResult := models.ConversationSearchResult{
				ConversationID:   conversationID,
				Title:            conversation.Title,
				UpdatedAt:        conversation.UpdatedAt,
				Score:            result.Score,
				MatchedMessages: []models.MatchedMessage{
					{
						ID:      messageID,
						Role:    role,
						Content: content,
						Score:   result.Score,
					},
				},
			}
			conversationMap[conversationID] = searchResult
		}
	}

	// 转换map为slice
	for _, result := range conversationMap {
		results = append(results, result)
	}

	return results, nil
}

// SearchInConversation 在特定对话中搜索相关消息
func (s *ConversationVectorService) SearchInConversation(ctx context.Context, conversationID, query string, topK int) ([]models.MessageSearchResult, error) {
	// 生成查询向量
	embeddingReq := &models.EmbeddingRequest{
		Model: "text-embedding-ada-002", // 使用默认的嵌入模型
		Input: []string{query},
	}

	embeddingResp, err := s.modelService.GenerateEmbedding(ctx, embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddingResp.Embedding) == 0 {
		return nil, fmt.Errorf("empty query embedding generated")
	}

	// 在向量数据库中搜索
	searchReq := &models.SearchRequest{
		CollectionName: "conversations",
		QueryVector:    embeddingResp.Embedding,
		TopK:           topK,
		Filter: map[string]interface{}{
			"conversation_id": conversationID,
		},
	}

	searchResp, err := s.vectorDAO.SearchVectorData(ctx, "conversations", searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// 转换搜索结果
	results := make([]models.MessageSearchResult, 0, len(searchResp.Results))
	for _, result := range searchResp.Results {
		messageID, _ := result.Metadata["message_id"].(string)
		role, _ := result.Metadata["role"].(string)
		content, _ := result.Metadata["content"].(string)
		
		messageResult := models.MessageSearchResult{
			MessageID: messageID,
			Role:      role,
			Content:   content,
			Score:     result.Score,
		}
		results = append(results, messageResult)
	}

	return results, nil
}

// RemoveFromIndex 从向量数据库中移除对话
func (s *ConversationVectorService) RemoveFromIndex(ctx context.Context, conversationID string) error {
	// 在向量数据库中搜索
	searchReq := &models.SearchRequest{
		CollectionName: "conversations",
		QueryVector:    make([]float64, 1536), // 使用零向量进行搜索，仅用于过滤
		TopK:           1000, // 获取所有匹配的结果
		Filter: map[string]interface{}{
			"conversation_id": conversationID,
		},
	}

	searchResp, err := s.vectorDAO.SearchVectorData(ctx, "conversations", searchReq)
	if err != nil {
		return fmt.Errorf("failed to search vectors for removal: %w", err)
	}

	// 提取所有向量ID
	vectorIDs := make([]string, 0, len(searchResp.Results))
	for _, result := range searchResp.Results {
		vectorIDs = append(vectorIDs, result.ID)
	}

	if len(vectorIDs) == 0 {
		return nil // 没有需要删除的向量
	}

	// 从向量数据库中删除向量
	deleteReq := &models.DeleteVectorsRequest{
		CollectionName: "conversations",
		Ids:            vectorIDs,
	}

	_, err = s.vectorDAO.BatchDeleteVectorData(ctx, "conversations", deleteReq)
	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	return nil
}

// ReindexConversation 重新索引对话（先删除再添加）
func (s *ConversationVectorService) ReindexConversation(ctx context.Context, conversationID string) error {
	// 先从索引中移除
	if err := s.RemoveFromIndex(ctx, conversationID); err != nil {
		return fmt.Errorf("failed to remove from index: %w", err)
	}

	// 重新索引
	if err := s.IndexConversation(ctx, conversationID); err != nil {
		return fmt.Errorf("failed to reindex: %w", err)
	}

	return nil
}

// GetSimilarMessages 获取与指定消息相似的消息
func (s *ConversationVectorService) GetSimilarMessages(ctx context.Context, conversationID, messageID string, topK int) ([]models.MessageSearchResult, error) {
	// 获取指定消息
	messages, err := s.messageDAO.GetMessagesByConversationID(ctx, conversationID, 0, 0) // 获取所有消息
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	var targetMessage models.Message
	for _, msg := range messages {
		if msg.ID == messageID {
			targetMessage = *msg
			break
		}
	}

	if targetMessage.ID == "" {
		return nil, fmt.Errorf("message not found: %s", messageID)
	}

	// 使用消息内容进行搜索
	return s.SearchInConversation(ctx, conversationID, targetMessage.Content, topK+1)
}