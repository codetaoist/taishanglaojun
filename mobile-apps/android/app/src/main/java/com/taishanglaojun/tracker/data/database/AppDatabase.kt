package com.taishanglaojun.tracker.data.database

import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.TypeConverters
import android.content.Context
import com.taishanglaojun.tracker.data.dao.ChatDao
import com.taishanglaojun.tracker.data.dao.LocationPointDao
import com.taishanglaojun.tracker.data.dao.TrajectoryDao
import com.taishanglaojun.tracker.data.model.ChatMessage
import com.taishanglaojun.tracker.data.model.Conversation
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.model.Trajectory

/**
 * 太上老君追踪器应用数据库
 * 包含位置追踪和AI对话功能的数据存储
 */
@Database(
    entities = [
        LocationPoint::class,
        Trajectory::class,
        ChatMessage::class,
        Conversation::class
    ],
    version = 2,
    exportSchema = false
)
@TypeConverters(DatabaseConverters::class)
abstract class AppDatabase : RoomDatabase() {
    
    abstract fun locationPointDao(): LocationPointDao
    abstract fun trajectoryDao(): TrajectoryDao
    abstract fun chatDao(): ChatDao
    
    companion object {
        @Volatile
        private var INSTANCE: AppDatabase? = null
        
        private const val DATABASE_NAME = "taishanglaojun_tracker.db"
        
        fun getDatabase(context: Context): AppDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    AppDatabase::class.java,
                    DATABASE_NAME
                )
                    .addMigrations(MIGRATION_1_2)
                    .fallbackToDestructiveMigration() // 开发阶段使用，生产环境需要移除
                    .build()
                INSTANCE = instance
                instance
            }
        }
        
        /**
         * 数据库迁移：从版本1到版本2
         * 添加AI对话功能相关的表
         */
        private val MIGRATION_1_2 = androidx.room.migration.Migration(1, 2) { database ->
            // 创建conversations表
            database.execSQL("""
                CREATE TABLE IF NOT EXISTS conversations (
                    id TEXT PRIMARY KEY NOT NULL,
                    title TEXT NOT NULL,
                    createdAt INTEGER NOT NULL,
                    updatedAt INTEGER NOT NULL,
                    lastMessageId TEXT,
                    messageCount INTEGER NOT NULL DEFAULT 0,
                    isArchived INTEGER NOT NULL DEFAULT 0,
                    aiPersonality TEXT NOT NULL DEFAULT 'WISE_SAGE'
                )
            """)
            
            // 创建chat_messages表
            database.execSQL("""
                CREATE TABLE IF NOT EXISTS chat_messages (
                    id TEXT PRIMARY KEY NOT NULL,
                    conversationId TEXT NOT NULL,
                    content TEXT NOT NULL,
                    messageType TEXT NOT NULL,
                    sender TEXT NOT NULL,
                    timestamp INTEGER NOT NULL,
                    status TEXT NOT NULL DEFAULT 'SENDING',
                    metadata TEXT NOT NULL DEFAULT '{}',
                    FOREIGN KEY(conversationId) REFERENCES conversations(id) ON DELETE CASCADE
                )
            """)
            
            // 创建索引以优化查询性能
            database.execSQL("CREATE INDEX IF NOT EXISTS index_conversations_updatedAt ON conversations(updatedAt)")
            database.execSQL("CREATE INDEX IF NOT EXISTS index_conversations_isArchived ON conversations(isArchived)")
            database.execSQL("CREATE INDEX IF NOT EXISTS index_chat_messages_conversationId ON chat_messages(conversationId)")
            database.execSQL("CREATE INDEX IF NOT EXISTS index_chat_messages_timestamp ON chat_messages(timestamp)")
            database.execSQL("CREATE INDEX IF NOT EXISTS index_chat_messages_status ON chat_messages(status)")
        }
    }
}