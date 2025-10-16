package com.taishanglaojun.tracker.data.repository

import com.taishanglaojun.tracker.data.dao.ChatDao
import com.taishanglaojun.tracker.data.model.*
import com.taishanglaojun.tracker.data.service.AIService
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatRepository @Inject constructor(
    private val chatDao: ChatDao,
    private val aiService: AIService
) {
    
    // 对话相关操作
    fun getAllConversations(): Flow<List<Conversation>> = chatDao.getAllConversations()
    
    suspend fun getConversationById(conversationId: String): Conversation? {
        return chatDao.getConversationById(conversationId)
    }
    
    suspend fun createConversation(title: String, aiPersonality: AIPersonality = AIPersonality.DEFAULT): Result<Conversation> {
        return try {
            // 先尝试在服务器创建
            val request = CreateConversationRequest(title, aiPersonality)
            val serverResult = aiService.createConversation(request)
            
            val conversation = if (serverResult.isSuccess) {
                val serverData = serverResult.getOrThrow()
                Conversation(
                    id = serverData.conversationId,
                    title = title,
                    createdAt = serverData.createdAt,
                    updatedAt = serverData.createdAt,
                    aiPersonality = aiPersonality
                )
            } else {
                // 服务器创建失败，创建本地对话
                Conversation(
                    title = title,
                    aiPersonality = aiPersonality
                )
            }
            
            // 保存到本地数据库
            chatDao.insertConversation(conversation)
            Result.success(conversation)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun updateConversation(conversation: Conversation) {
        chatDao.updateConversation(conversation.copy(updatedAt = System.currentTimeMillis()))
    }
    
    suspend fun deleteConversation(conversationId: String): Result<Unit> {
        return try {
            // 先删除服务器数据
            aiService.deleteConversation(conversationId)
            
            // 删除本地数据
            chatDao.deleteMessagesByConversation(conversationId)
            chatDao.deleteConversation(conversationId)
            
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun archiveConversation(conversationId: String) {
        chatDao.archiveConversation(conversationId)
    }
    
    // 消息相关操作
    fun getMessagesByConversation(conversationId: String): Flow<List<ChatMessage>> {
        return chatDao.getMessagesByConversation(conversationId)
    }
    
    suspend fun sendMessage(
        conversationId: String,
        content: String,
        messageType: MessageType = MessageType.TEXT
    ): Result<ChatMessage> {
        return try {
            // 创建用户消息
            val userMessage = ChatMessage(
                conversationId = conversationId,
                content = content,
                messageType = messageType,
                sender = MessageSender.USER,
                status = MessageStatus.SENDING
            )
            
            // 保存用户消息到本地
            chatDao.insertMessage(userMessage)
            
            // 获取对话信息
            val conversation = chatDao.getConversationById(conversationId)
            
            // 发送到AI服务
            val request = SendMessageRequest(
                conversationId = conversationId,
                message = content,
                messageType = messageType,
                aiPersonality = conversation?.aiPersonality ?: AIPersonality.DEFAULT
            )
            
            val aiResponse = aiService.sendMessage(request)
            
            if (aiResponse.isSuccess) {
                val response = aiResponse.getOrThrow()
                
                // 更新用户消息状态
                chatDao.updateMessageStatus(userMessage.id, MessageStatus.SENT.name)
                
                // 创建AI回复消息
                response.message?.let { aiMessage ->
                    val aiReplyMessage = ChatMessage(
                        conversationId = conversationId,
                        content = aiMessage.content,
                        messageType = aiMessage.messageType,
                        sender = MessageSender.AI,
                        timestamp = aiMessage.timestamp,
                        status = MessageStatus.DELIVERED
                    )
                    
                    // 保存AI回复到本地
                    chatDao.insertMessage(aiReplyMessage)
                    
                    // 更新对话信息
                    conversation?.let { conv ->
                        val updatedConv = conv.copy(
                            updatedAt = System.currentTimeMillis(),
                            lastMessageId = aiReplyMessage.id,
                            messageCount = conv.messageCount + 2 // 用户消息 + AI回复
                        )
                        chatDao.updateConversation(updatedConv)
                    }
                }
                
                Result.success(userMessage.copy(status = MessageStatus.SENT))
            } else {
                // AI服务失败，更新消息状态
                chatDao.updateMessageStatus(userMessage.id, MessageStatus.FAILED.name)
                Result.failure(aiResponse.exceptionOrNull() ?: Exception("AI service failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun sendImageMessage(
        conversationId: String,
        imageUri: String,
        caption: String = ""
    ): Result<ChatMessage> {
        return try {
            // 创建图片消息
            val imageMessage = ChatMessage(
                conversationId = conversationId,
                content = imageUri,
                messageType = MessageType.IMAGE,
                sender = MessageSender.USER,
                status = MessageStatus.SENDING,
                metadata = mapOf(
                    "caption" to caption,
                    "uri" to imageUri
                )
            )
            
            // 保存图片消息到本地
            chatDao.insertMessage(imageMessage)
            
            // 发送到服务器
            try {
                val request = SendMessageRequest(
                    conversationId = conversationId,
                    content = imageUri,
                    messageType = MessageType.IMAGE,
                    metadata = imageMessage.metadata
                )
                val response = aiService.sendMessage(request)
                
                if (response.isSuccess) {
                    // 更新消息状态为已发送
                    val sentMessage = imageMessage.copy(status = MessageStatus.SENT)
                    chatDao.updateMessage(sentMessage)
                    
                    // 处理AI回复
                    response.getOrNull()?.aiResponse?.let { aiResponse ->
                        val aiMessage = ChatMessage(
                            conversationId = conversationId,
                            content = aiResponse.content,
                            messageType = aiResponse.messageType,
                            sender = MessageSender.AI,
                            status = MessageStatus.DELIVERED
                        )
                        chatDao.insertMessage(aiMessage)
                    }
                    
                    Result.success(sentMessage)
                } else {
                    // 标记为发送失败
                    val failedMessage = imageMessage.copy(status = MessageStatus.FAILED)
                    chatDao.updateMessage(failedMessage)
                    Result.failure(response.exceptionOrNull() ?: Exception("发送图片失败"))
                }
            } catch (e: Exception) {
                // 网络错误，标记为发送失败
                val failedMessage = imageMessage.copy(status = MessageStatus.FAILED)
                chatDao.updateMessage(failedMessage)
                Result.failure(e)
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun sendAudioMessage(
        conversationId: String,
        audioUri: String,
        duration: Long
    ): Result<ChatMessage> {
        return try {
            // 创建音频消息
            val audioMessage = ChatMessage(
                conversationId = conversationId,
                content = audioUri,
                messageType = MessageType.AUDIO,
                sender = MessageSender.USER,
                status = MessageStatus.SENDING,
                metadata = mapOf(
                    "duration" to duration.toString(),
                    "uri" to audioUri
                )
            )
            
            // 保存音频消息到本地
            chatDao.insertMessage(audioMessage)
            
            // 发送到服务器
            try {
                val request = SendMessageRequest(
                    conversationId = conversationId,
                    content = audioUri,
                    messageType = MessageType.AUDIO,
                    metadata = audioMessage.metadata
                )
                val response = aiService.sendMessage(request)
                
                if (response.isSuccess) {
                    // 更新消息状态为已发送
                    val sentMessage = audioMessage.copy(status = MessageStatus.SENT)
                    chatDao.updateMessage(sentMessage)
                    
                    // 处理AI回复
                    response.getOrNull()?.aiResponse?.let { aiResponse ->
                        val aiMessage = ChatMessage(
                            conversationId = conversationId,
                            content = aiResponse.content,
                            messageType = aiResponse.messageType,
                            sender = MessageSender.AI,
                            status = MessageStatus.DELIVERED
                        )
                        chatDao.insertMessage(aiMessage)
                    }
                    
                    Result.success(sentMessage)
                } else {
                    // 标记为发送失败
                    val failedMessage = audioMessage.copy(status = MessageStatus.FAILED)
                    chatDao.updateMessage(failedMessage)
                    Result.failure(response.exceptionOrNull() ?: Exception("发送语音失败"))
                }
            } catch (e: Exception) {
                // 网络错误，标记为发送失败
                val failedMessage = audioMessage.copy(status = MessageStatus.FAILED)
                chatDao.updateMessage(failedMessage)
                Result.failure(e)
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun retryMessage(messageId: String): Result<Unit> {
        return try {
            val message = chatDao.getMessageById(messageId)
            if (message != null && message.sender == MessageSender.USER && message.status == MessageStatus.FAILED) {
                sendMessage(message.conversationId, message.content, message.messageType)
                Result.success(Unit)
            } else {
                Result.failure(Exception("Message not found or cannot be retried"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun deleteMessage(messageId: String) {
        chatDao.deleteMessage(messageId)
    }
    
    suspend fun searchMessages(conversationId: String, query: String): List<ChatMessage> {
        return chatDao.searchMessages(conversationId, query)
    }
    
    suspend fun searchConversations(query: String): List<String> {
        return chatDao.searchConversationsByMessage(query)
    }
    
    // 同步相关操作
    suspend fun syncConversations(): Result<Unit> {
        return try {
            val serverConversations = aiService.getConversations()
            if (serverConversations.isSuccess) {
                val conversations = serverConversations.getOrThrow()
                // 这里可以实现更复杂的同步逻辑
                // 比如合并本地和服务器的对话，处理冲突等
                conversations.forEach { conversation ->
                    chatDao.insertConversation(conversation)
                }
            }
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun syncMessages(conversationId: String): Result<Unit> {
        return try {
            val serverMessages = aiService.getMessages(conversationId)
            if (serverMessages.isSuccess) {
                val messagesData = serverMessages.getOrThrow()
                chatDao.insertMessages(messagesData.messages)
            }
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // 清理操作
    suspend fun cleanupOldData(daysToKeep: Int = 30) {
        val cutoffTime = System.currentTimeMillis() - (daysToKeep * 24 * 60 * 60 * 1000L)
        chatDao.deleteOldMessages(cutoffTime)
        chatDao.deleteOldArchivedConversations(cutoffTime)
    }
}