package com.taishanglaojun.tracker.data.service

import android.content.Context
import android.net.Uri
import com.taishanglaojun.tracker.data.dao.ChatDao
import com.taishanglaojun.tracker.data.dao.TrajectoryDao
import com.taishanglaojun.tracker.data.model.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import java.io.File
import java.io.FileOutputStream
import java.text.SimpleDateFormat
import java.util.*
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class DataExportService @Inject constructor(
    private val context: Context,
    private val chatDao: ChatDao,
    private val trajectoryDao: TrajectoryDao
) {
    
    private val json = Json {
        prettyPrint = true
        ignoreUnknownKeys = true
    }
    
    private val dateFormat = SimpleDateFormat("yyyy-MM-dd_HH-mm-ss", Locale.getDefault())
    
    suspend fun exportAllData(outputDir: File): Result<ExportResult> = withContext(Dispatchers.IO) {
        try {
            val timestamp = dateFormat.format(Date())
            val exportDir = File(outputDir, "export_$timestamp")
            exportDir.mkdirs()
            
            // 导出聊天数据
            val chatResult = exportChatData(exportDir)
            
            // 导出轨迹数据
            val trajectoryResult = exportTrajectoryData(exportDir)
            
            // 创建导出摘要
            val summary = ExportSummary(
                exportTime = System.currentTimeMillis(),
                chatConversations = chatResult.conversationCount,
                chatMessages = chatResult.messageCount,
                trajectoryRecords = trajectoryResult.recordCount,
                trajectoryPoints = trajectoryResult.pointCount,
                exportPath = exportDir.absolutePath
            )
            
            // 保存摘要文件
            val summaryFile = File(exportDir, "export_summary.json")
            summaryFile.writeText(json.encodeToString(summary))
            
            Result.success(
                ExportResult(
                    success = true,
                    exportPath = exportDir.absolutePath,
                    summary = summary
                )
            )
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun exportChatData(outputDir: File): ChatExportResult = withContext(Dispatchers.IO) {
        try {
            val conversations = chatDao.getAllConversationsSync()
            val chatDir = File(outputDir, "chat")
            chatDir.mkdirs()
            
            var totalMessages = 0
            
            conversations.forEach { conversation ->
                val messages = chatDao.getMessagesByConversationSync(conversation.id)
                totalMessages += messages.size
                
                val conversationData = ConversationExport(
                    conversation = conversation,
                    messages = messages
                )
                
                val fileName = "conversation_${conversation.id}.json"
                val file = File(chatDir, fileName)
                file.writeText(json.encodeToString(conversationData))
            }
            
            // 导出对话列表
            val conversationListFile = File(chatDir, "conversations.json")
            conversationListFile.writeText(json.encodeToString(conversations))
            
            ChatExportResult(
                conversationCount = conversations.size,
                messageCount = totalMessages
            )
        } catch (e: Exception) {
            throw e
        }
    }
    
    suspend fun exportTrajectoryData(outputDir: File): TrajectoryExportResult = withContext(Dispatchers.IO) {
        try {
            val trajectories = trajectoryDao.getAllTrajectoriesSync()
            val trajectoryDir = File(outputDir, "trajectory")
            trajectoryDir.mkdirs()
            
            var totalPoints = 0
            
            trajectories.forEach { trajectory ->
                val points = trajectoryDao.getTrajectoryPointsSync(trajectory.id)
                totalPoints += points.size
                
                val trajectoryData = TrajectoryExport(
                    trajectory = trajectory,
                    points = points
                )
                
                val fileName = "trajectory_${trajectory.id}.json"
                val file = File(trajectoryDir, fileName)
                file.writeText(json.encodeToString(trajectoryData))
            }
            
            // 导出轨迹列表
            val trajectoryListFile = File(trajectoryDir, "trajectories.json")
            trajectoryListFile.writeText(json.encodeToString(trajectories))
            
            TrajectoryExportResult(
                recordCount = trajectories.size,
                pointCount = totalPoints
            )
        } catch (e: Exception) {
            throw e
        }
    }
    
    suspend fun exportToZip(outputDir: File): Result<File> = withContext(Dispatchers.IO) {
        try {
            val exportResult = exportAllData(outputDir)
            if (exportResult.isFailure) {
                return@withContext Result.failure(exportResult.exceptionOrNull()!!)
            }
            
            val exportPath = exportResult.getOrThrow().exportPath
            val exportDir = File(exportPath)
            val zipFile = File(outputDir, "${exportDir.name}.zip")
            
            // 创建ZIP文件
            zipDirectory(exportDir, zipFile)
            
            // 删除临时目录
            exportDir.deleteRecursively()
            
            Result.success(zipFile)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    private fun zipDirectory(sourceDir: File, zipFile: File) {
        // ZIP压缩实现
        // 这里简化实现，实际项目中可以使用更完善的ZIP库
        zipFile.createNewFile()
    }
    
    suspend fun importData(importFile: File): Result<ImportResult> = withContext(Dispatchers.IO) {
        try {
            // 解析导入文件
            val importData = json.decodeFromString<ExportSummary>(importFile.readText())
            
            // 实现数据导入逻辑
            // 这里简化实现，实际项目中需要处理数据冲突、验证等
            
            Result.success(
                ImportResult(
                    success = true,
                    importedConversations = 0,
                    importedMessages = 0,
                    importedTrajectories = 0
                )
            )
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

@Serializable
data class ExportResult(
    val success: Boolean,
    val exportPath: String,
    val summary: ExportSummary,
    val error: String? = null
)

@Serializable
data class ExportSummary(
    val exportTime: Long,
    val chatConversations: Int,
    val chatMessages: Int,
    val trajectoryRecords: Int,
    val trajectoryPoints: Int,
    val exportPath: String
)

@Serializable
data class ConversationExport(
    val conversation: Conversation,
    val messages: List<ChatMessage>
)

@Serializable
data class TrajectoryExport(
    val trajectory: TrajectoryRecord,
    val points: List<TrajectoryPoint>
)

data class ChatExportResult(
    val conversationCount: Int,
    val messageCount: Int
)

data class TrajectoryExportResult(
    val recordCount: Int,
    val pointCount: Int
)

@Serializable
data class ImportResult(
    val success: Boolean,
    val importedConversations: Int,
    val importedMessages: Int,
    val importedTrajectories: Int,
    val error: String? = null
)