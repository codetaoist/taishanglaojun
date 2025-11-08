package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/response"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/service"
)

// exportConversations 导出用户的对话
func exportConversations(importExportService *service.ConversationImportExportService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ExportRequest
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

		// 根据请求导出对话
		var export interface{}
		var err error

		if len(req.ConversationIDs) > 0 {
			// 导出指定的对话
			conversationExports := make([]models.ConversationExport, 0, len(req.ConversationIDs))
			for _, convID := range req.ConversationIDs {
				convExport, err := importExportService.ExportConversation(c.Request.Context(), convID)
				if err != nil {
					response.InternalServerError(c, fmt.Sprintf("Failed to export conversation %s: %s", convID, err.Error()))
					return
				}
				conversationExports = append(conversationExports, *convExport)
			}
			export = models.ConversationsExport{
				UserID:       userID.(string),
				Conversations: conversationExports,
				ExportedAt:   time.Now(),
			}
		} else {
			// 导出用户的所有对话
			export, err = importExportService.ExportConversations(c.Request.Context(), userID.(string))
			if err != nil {
				response.InternalServerError(c, err.Error())
				return
			}
		}

		// 转换为指定格式
		var data []byte
		var filename string

		switch req.Format {
		case models.ExportFormatJSON:
			data, err = importExportService.ExportToJSON(export)
			filename = fmt.Sprintf("conversations_%s.json", time.Now().Format("20060102_150405"))
		default:
			response.BadRequest(c, "Unsupported export format")
			return
		}

		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}

		// 设置响应头
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/octet-stream")
		c.Data(http.StatusOK, "application/octet-stream", data)
	}
}

// exportConversation 导出单个对话
func exportConversation(importExportService *service.ConversationImportExportService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		if conversationID == "" {
			response.BadRequest(c, "Conversation ID is required")
			return
		}

		// 获取格式参数，默认为JSON
		format := models.ExportFormat(c.DefaultQuery("format", "json"))
		includeMetadata := c.DefaultQuery("include_metadata", "true") == "true"
		
		// TODO: 使用includeMetadata选项
		_ = includeMetadata

		// 导出对话
		export, err := importExportService.ExportConversation(c.Request.Context(), conversationID)
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}

		// 如果不包含元数据，则清除元数据
		if !includeMetadata {
			// export.Conversation.Metadata = nil // 已移除Metadata字段
			for i := range export.Messages {
				export.Messages[i].Metadata = nil
			}
		}

		// 转换为指定格式
		var data []byte
		var filename string

		switch format {
		case models.ExportFormatJSON:
			data, err = importExportService.ExportToJSON(export)
			filename = fmt.Sprintf("conversation_%s.json", conversationID)
		default:
			response.BadRequest(c, "Unsupported export format")
			return
		}

		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}

		// 设置响应头
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/octet-stream")
		c.Data(http.StatusOK, "application/octet-stream", data)
	}
}

// importConversations 导入对话
func importConversations(importExportService *service.ConversationImportExportService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			return
		}

		// 获取导入选项
		conflictStrategy := models.ImportConflict(c.DefaultQuery("conflict_strategy", "skip"))
		preserveIDs := c.DefaultQuery("preserve_ids", "false") == "true"
		
		// TODO: 使用这些选项实现导入逻辑
		_ = conflictStrategy
		_ = preserveIDs

		// 解析上传的文件
		file, err := c.FormFile("file")
		if err != nil {
			response.BadRequest(c, "Failed to get file from form: "+err.Error())
			return
		}

		// 打开文件
		fileContent, err := file.Open()
		if err != nil {
			response.InternalServerError(c, "Failed to open file: "+err.Error())
			return
		}
		defer fileContent.Close()

		// 读取文件内容
		data := make([]byte, file.Size)
		_, err = fileContent.Read(data)
		if err != nil {
			response.InternalServerError(c, "Failed to read file: "+err.Error())
			return
		}

		// 根据文件扩展名确定格式
		// TODO: 实现多格式支持
		_ = models.ExportFormatJSON // 默认为JSON
		if len(file.Filename) > 5 && file.Filename[len(file.Filename)-5:] == ".json" {
			// format = models.ExportFormatJSON
		} else {
			response.BadRequest(c, "Unsupported file format")
			return
		}

		// 从JSON导入数据
		importedData, err := importExportService.ImportFromJSON(data, true) // 假设是多对话导出
		if err != nil {
			response.InternalServerError(c, "Failed to parse import data: "+err.Error())
			return
		}

		// 执行导入
		var importedConversations []*models.Conversation
		var importResponse models.ImportResponse
		importResponse.ImportedAt = time.Now()

		switch data := importedData.(type) {
		case *models.ConversationsExport:
			importedConversations, err = importExportService.ImportConversations(c.Request.Context(), data, userID.(string))
			importResponse.ImportedCount = len(importedConversations)
		case *models.ConversationExport:
			conversation, err := importExportService.ImportConversation(c.Request.Context(), data, userID.(string))
			if err == nil {
				importedConversations = []*models.Conversation{conversation}
				importResponse.ImportedCount = 1
			}
		default:
			response.BadRequest(c, "Invalid import data format")
			return
		}

		if err != nil {
			response.InternalServerError(c, "Failed to import conversations: "+err.Error())
			return
		}

		// 构建响应
		importResponse.ImportedIDs = make([]string, len(importedConversations))
		for i, conv := range importedConversations {
			importResponse.ImportedIDs[i] = conv.ID
		}

		response.Success(c, importResponse)
	}
}