package com.taishanglaojun.tracker.data.dao

import androidx.room.*
import com.taishanglaojun.tracker.data.model.ChatMessage
import com.taishanglaojun.tracker.data.model.Conversation
import kotlinx.coroutines.flow.Flow

@Dao
interface ChatDao {
    
    // 对话相关操作
    @Query("SELECT * FROM conversations WHERE isArchived = 0 ORDER BY updatedAt DESC")
    fun getAllConversations(): Flow<List<Conversation>>
    
    @Query("SELECT * FROM conversations WHERE id = :conversationId")
    suspend fun getConversationById(conversationId: String): Conversation?
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertConversation(conversation: Conversation)
    
    @Update
    suspend fun updateConversation(conversation: Conversation)
    
    @Query("DELETE FROM conversations WHERE id = :conversationId")
    suspend fun deleteConversation(conversationId: String)
    
    @Query("UPDATE conversations SET isArchived = 1 WHERE id = :conversationId")
    suspend fun archiveConversation(conversationId: String)
    
    // 消息相关操作
    @Query("SELECT * FROM chat_messages WHERE conversationId = :conversationId ORDER BY timestamp ASC")
    fun getMessagesByConversation(conversationId: String): Flow<List<ChatMessage>>
    
    @Query("SELECT * FROM chat_messages WHERE conversationId = :conversationId ORDER BY timestamp DESC LIMIT :limit OFFSET :offset")
    suspend fun getMessagesPaged(conversationId: String, limit: Int, offset: Int): List<ChatMessage>
    
    @Query("SELECT * FROM chat_messages WHERE id = :messageId")
    suspend fun getMessageById(messageId: String): ChatMessage?
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertMessage(message: ChatMessage)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertMessages(messages: List<ChatMessage>)
    
    @Update
    suspend fun updateMessage(message: ChatMessage)
    
    @Query("DELETE FROM chat_messages WHERE id = :messageId")
    suspend fun deleteMessage(messageId: String)
    
    @Query("DELETE FROM chat_messages WHERE conversationId = :conversationId")
    suspend fun deleteMessagesByConversation(conversationId: String)
    
    @Query("UPDATE chat_messages SET status = :status WHERE id = :messageId")
    suspend fun updateMessageStatus(messageId: String, status: String)
    
    // 统计相关
    @Query("SELECT COUNT(*) FROM chat_messages WHERE conversationId = :conversationId")
    suspend fun getMessageCount(conversationId: String): Int
    
    @Query("SELECT * FROM chat_messages WHERE conversationId = :conversationId ORDER BY timestamp DESC LIMIT 1")
    suspend fun getLastMessage(conversationId: String): ChatMessage?
    
    // 搜索功能
    @Query("SELECT * FROM chat_messages WHERE conversationId = :conversationId AND content LIKE '%' || :query || '%' ORDER BY timestamp DESC")
    suspend fun searchMessages(conversationId: String, query: String): List<ChatMessage>
    
    @Query("SELECT DISTINCT conversationId FROM chat_messages WHERE content LIKE '%' || :query || '%'")
    suspend fun searchConversationsByMessage(query: String): List<String>
    
    // 清理操作
    @Query("DELETE FROM chat_messages WHERE timestamp < :beforeTimestamp")
    suspend fun deleteOldMessages(beforeTimestamp: Long)
    
    @Query("DELETE FROM conversations WHERE isArchived = 1 AND updatedAt < :beforeTimestamp")
    suspend fun deleteOldArchivedConversations(beforeTimestamp: Long)
}