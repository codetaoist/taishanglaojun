package com.taishanglaojun.tracker.data.service

import com.taishanglaojun.tracker.data.model.*
import retrofit2.Response
import retrofit2.http.*
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.delay

interface AIApiService {
    
    @POST("api/ai/chat/send")
    suspend fun sendMessage(@Body request: SendMessageRequest): Response<AIResponse>
    
    @GET("api/ai/chat/conversations/{conversationId}/messages")
    suspend fun getMessages(
        @Path("conversationId") conversationId: String,
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 50,
        @Query("before") before: Long? = null
    ): Response<MessagesResponse>
    
    @POST("api/ai/chat/conversations")
    suspend fun createConversation(@Body request: CreateConversationRequest): Response<ConversationResponse>
    
    @GET("api/ai/chat/conversations")
    suspend fun getConversations(): Response<ConversationsResponse>
    
    @DELETE("api/ai/chat/conversations/{conversationId}")
    suspend fun deleteConversation(@Path("conversationId") conversationId: String): Response<Unit>
}

data class MessagesResponse(
    val success: Boolean,
    val data: MessagesData,
    val error: AIError? = null
)

data class MessagesData(
    val messages: List<ChatMessage>,
    val hasMore: Boolean,
    val total: Int
)

data class ConversationResponse(
    val success: Boolean,
    val data: ConversationData,
    val error: AIError? = null
)

data class ConversationData(
    val conversationId: String,
    val title: String,
    val createdAt: Long
)

data class ConversationsResponse(
    val success: Boolean,
    val data: List<Conversation>,
    val error: AIError? = null
)

@Singleton
class AIService @Inject constructor(
    private val apiService: AIApiService
) {
    
    suspend fun sendMessage(request: SendMessageRequest): Result<AIResponse> {
        return try {
            val response = apiService.sendMessage(request)
            if (response.isSuccessful) {
                response.body()?.let { aiResponse ->
                    Result.success(aiResponse)
                } ?: Result.failure(Exception("Empty response"))
            } else {
                Result.failure(Exception("HTTP ${response.code()}: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getMessages(
        conversationId: String,
        page: Int = 1,
        limit: Int = 50,
        before: Long? = null
    ): Result<MessagesData> {
        return try {
            val response = apiService.getMessages(conversationId, page, limit, before)
            if (response.isSuccessful) {
                response.body()?.let { messagesResponse ->
                    if (messagesResponse.success) {
                        Result.success(messagesResponse.data)
                    } else {
                        Result.failure(Exception(messagesResponse.error?.message ?: "Unknown error"))
                    }
                } ?: Result.failure(Exception("Empty response"))
            } else {
                Result.failure(Exception("HTTP ${response.code()}: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun createConversation(request: CreateConversationRequest): Result<ConversationData> {
        return try {
            val response = apiService.createConversation(request)
            if (response.isSuccessful) {
                response.body()?.let { conversationResponse ->
                    if (conversationResponse.success) {
                        Result.success(conversationResponse.data)
                    } else {
                        Result.failure(Exception(conversationResponse.error?.message ?: "Unknown error"))
                    }
                } ?: Result.failure(Exception("Empty response"))
            } else {
                Result.failure(Exception("HTTP ${response.code()}: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getConversations(): Result<List<Conversation>> {
        return try {
            val response = apiService.getConversations()
            if (response.isSuccessful) {
                response.body()?.let { conversationsResponse ->
                    if (conversationsResponse.success) {
                        Result.success(conversationsResponse.data)
                    } else {
                        Result.failure(Exception(conversationsResponse.error?.message ?: "Unknown error"))
                    }
                } ?: Result.failure(Exception("Empty response"))
            } else {
                Result.failure(Exception("HTTP ${response.code()}: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun deleteConversation(conversationId: String): Result<Unit> {
        return try {
            val response = apiService.deleteConversation(conversationId)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("HTTP ${response.code()}: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    // 流式响应模拟（实际项目中可能需要WebSocket或SSE）
    fun sendMessageStream(request: SendMessageRequest): Flow<String> = flow {
        try {
            val response = sendMessage(request)
            response.getOrNull()?.message?.content?.let { content ->
                // 模拟流式输出
                val words = content.split(" ")
                for (word in words) {
                    emit(word)
                    delay(100) // 模拟打字效果
                }
            }
        } catch (e: Exception) {
            throw e
        }
    }
}