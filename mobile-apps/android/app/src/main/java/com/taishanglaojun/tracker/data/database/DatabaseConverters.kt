package com.taishanglaojun.tracker.data.database

import androidx.room.TypeConverter
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import com.taishanglaojun.tracker.data.model.AIPersonality
import com.taishanglaojun.tracker.data.model.MessageSender
import com.taishanglaojun.tracker.data.model.MessageStatus
import com.taishanglaojun.tracker.data.model.MessageType

/**
 * Room数据库类型转换器
 * 用于处理复杂数据类型的序列化和反序列化
 */
class DatabaseConverters {
    
    private val gson = Gson()
    
    // MessageType转换
    @TypeConverter
    fun fromMessageType(messageType: MessageType): String {
        return messageType.name
    }
    
    @TypeConverter
    fun toMessageType(messageType: String): MessageType {
        return MessageType.valueOf(messageType)
    }
    
    // MessageSender转换
    @TypeConverter
    fun fromMessageSender(sender: MessageSender): String {
        return sender.name
    }
    
    @TypeConverter
    fun toMessageSender(sender: String): MessageSender {
        return MessageSender.valueOf(sender)
    }
    
    // MessageStatus转换
    @TypeConverter
    fun fromMessageStatus(status: MessageStatus): String {
        return status.name
    }
    
    @TypeConverter
    fun toMessageStatus(status: String): MessageStatus {
        return MessageStatus.valueOf(status)
    }
    
    // AIPersonality转换
    @TypeConverter
    fun fromAIPersonality(personality: AIPersonality): String {
        return personality.name
    }
    
    @TypeConverter
    fun toAIPersonality(personality: String): AIPersonality {
        return AIPersonality.valueOf(personality)
    }
    
    // Map<String, Any>转换（用于metadata）
    @TypeConverter
    fun fromStringMap(value: Map<String, Any>?): String {
        return gson.toJson(value)
    }
    
    @TypeConverter
    fun toStringMap(value: String): Map<String, Any>? {
        val mapType = object : TypeToken<Map<String, Any>>() {}.type
        return try {
            gson.fromJson(value, mapType)
        } catch (e: Exception) {
            emptyMap()
        }
    }
    
    // List<String>转换
    @TypeConverter
    fun fromStringList(value: List<String>?): String {
        return gson.toJson(value)
    }
    
    @TypeConverter
    fun toStringList(value: String): List<String>? {
        val listType = object : TypeToken<List<String>>() {}.type
        return try {
            gson.fromJson(value, listType)
        } catch (e: Exception) {
            emptyList()
        }
    }
}