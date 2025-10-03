package com.taishanglaojun.tracker.data.model

import androidx.room.Entity
import androidx.room.PrimaryKey
import java.util.*

@Entity(
    tableName = "chat_messages",
    foreignKeys = [
        androidx.room.ForeignKey(
            entity = Conversation::class,
            parentColumns = ["id"],
            childColumns = ["conversationId"],
            onDelete = androidx.room.ForeignKey.CASCADE
        )
    ],
    indices = [
        androidx.room.Index(value = ["conversationId"]),
        androidx.room.Index(value = ["timestamp"]),
        androidx.room.Index(value = ["status"])
    ]
)
data class ChatMessage(
    @PrimaryKey
    val id: String = UUID.randomUUID().toString(),
    val conversationId: String,
    val content: String,
    val messageType: MessageType,
    val sender: MessageSender,
    val timestamp: Long = System.currentTimeMillis(),
    val status: MessageStatus = MessageStatus.SENDING,
    val metadata: String = "{}" // JSON字符串
)

enum class MessageType {
    TEXT,
    IMAGE,
    AUDIO,
    SYSTEM
}

enum class MessageSender {
    USER,
    AI,
    SYSTEM
}

enum class MessageStatus {
    SENDING,
    SENT,
    DELIVERED,
    READ,
    FAILED
}

@Entity(
    tableName = "conversations",
    indices = [
        androidx.room.Index(value = ["updatedAt"]),
        androidx.room.Index(value = ["isArchived"])
    ]
)
data class Conversation(
    @PrimaryKey
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val createdAt: Long = System.currentTimeMillis(),
    val updatedAt: Long = System.currentTimeMillis(),
    val lastMessageId: String? = null,
    val messageCount: Int = 0,
    val isArchived: Boolean = false,
    val aiPersonality: AIPersonality = AIPersonality.DEFAULT
)

enum class AIPersonality {
    DEFAULT,
    WISE_SAGE,      // 智慧长者
    FRIENDLY_GUIDE, // 友善向导
    SCHOLARLY,      // 学者风格
    POETIC         // 诗意风格
}

// AI响应数据类
data class AIResponse(
    val success: Boolean,
    val message: AIMessage?,
    val suggestions: List<String> = emptyList(),
    val error: AIError? = null
)

data class AIMessage(
    val content: String,
    val messageType: MessageType,
    val timestamp: Long,
    val metadata: Map<String, Any> = emptyMap()
)

data class AIError(
    val code: String,
    val message: String,
    val details: Map<String, Any> = emptyMap()
)

// 发送消息请求
data class SendMessageRequest(
    val conversationId: String,
    val message: String,
    val messageType: MessageType = MessageType.TEXT,
    val aiPersonality: AIPersonality = AIPersonality.DEFAULT
)

// 创建对话请求
data class CreateConversationRequest(
    val title: String,
    val aiPersonality: AIPersonality = AIPersonality.DEFAULT
)