package com.taishanglaojun.tracker.data.service

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.media.MediaMetadataRetriever
import android.media.MediaRecorder
import android.net.Uri
import android.util.Base64
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.ByteArrayOutputStream
import java.io.File
import java.io.FileOutputStream
import java.io.IOException
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class MediaService @Inject constructor(
    @ApplicationContext private val context: Context
) {
    
    companion object {
        private const val MAX_IMAGE_SIZE = 1024 * 1024 // 1MB
        private const val IMAGE_QUALITY = 80
        private const val MAX_AUDIO_DURATION = 60 * 1000L // 60 seconds
    }
    
    /**
     * 压缩并编码图片为Base64
     */
    suspend fun compressAndEncodeImage(uri: Uri): Result<String> = withContext(Dispatchers.IO) {
        try {
            val inputStream = context.contentResolver.openInputStream(uri)
            val bitmap = BitmapFactory.decodeStream(inputStream)
            inputStream?.close()
            
            if (bitmap == null) {
                return@withContext Result.failure(Exception("无法解码图片"))
            }
            
            // 计算压缩比例
            val ratio = calculateCompressionRatio(bitmap)
            val scaledBitmap = Bitmap.createScaledBitmap(
                bitmap,
                (bitmap.width * ratio).toInt(),
                (bitmap.height * ratio).toInt(),
                true
            )
            
            // 压缩为JPEG
            val outputStream = ByteArrayOutputStream()
            scaledBitmap.compress(Bitmap.CompressFormat.JPEG, IMAGE_QUALITY, outputStream)
            val imageBytes = outputStream.toByteArray()
            
            // 编码为Base64
            val base64String = Base64.encodeToString(imageBytes, Base64.DEFAULT)
            
            bitmap.recycle()
            scaledBitmap.recycle()
            outputStream.close()
            
            Result.success(base64String)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    /**
     * 解码Base64图片并保存到本地
     */
    suspend fun decodeAndSaveImage(base64String: String, fileName: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            val imageBytes = Base64.decode(base64String, Base64.DEFAULT)
            val imagesDir = File(context.filesDir, "images")
            if (!imagesDir.exists()) {
                imagesDir.mkdirs()
            }
            
            val imageFile = File(imagesDir, fileName)
            val outputStream = FileOutputStream(imageFile)
            outputStream.write(imageBytes)
            outputStream.close()
            
            Result.success(imageFile.absolutePath)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    /**
     * 开始录音
     */
    fun startRecording(outputFile: File): MediaRecorder? {
        return try {
            MediaRecorder().apply {
                setAudioSource(MediaRecorder.AudioSource.MIC)
                setOutputFormat(MediaRecorder.OutputFormat.AAC_ADTS)
                setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
                setOutputFile(outputFile.absolutePath)
                setMaxDuration(MAX_AUDIO_DURATION.toInt())
                prepare()
                start()
            }
        } catch (e: Exception) {
            null
        }
    }
    
    /**
     * 停止录音
     */
    fun stopRecording(recorder: MediaRecorder?) {
        try {
            recorder?.apply {
                stop()
                release()
            }
        } catch (e: Exception) {
            // 忽略停止录音时的异常
        }
    }
    
    /**
     * 编码音频文件为Base64
     */
    suspend fun encodeAudioToBase64(audioFile: File): Result<String> = withContext(Dispatchers.IO) {
        try {
            val audioBytes = audioFile.readBytes()
            val base64String = Base64.encodeToString(audioBytes, Base64.DEFAULT)
            Result.success(base64String)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    /**
     * 解码Base64音频并保存到本地
     */
    suspend fun decodeAndSaveAudio(base64String: String, fileName: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            val audioBytes = Base64.decode(base64String, Base64.DEFAULT)
            val audioDir = File(context.filesDir, "audio")
            if (!audioDir.exists()) {
                audioDir.mkdirs()
            }
            
            val audioFile = File(audioDir, fileName)
            val outputStream = FileOutputStream(audioFile)
            outputStream.write(audioBytes)
            outputStream.close()
            
            Result.success(audioFile.absolutePath)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    /**
     * 获取音频时长
     */
    suspend fun getAudioDuration(filePath: String): Result<Long> = withContext(Dispatchers.IO) {
        try {
            val retriever = MediaMetadataRetriever()
            retriever.setDataSource(filePath)
            val duration = retriever.extractMetadata(MediaMetadataRetriever.METADATA_KEY_DURATION)?.toLong() ?: 0L
            retriever.release()
            Result.success(duration)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    /**
     * 计算图片压缩比例
     */
    private fun calculateCompressionRatio(bitmap: Bitmap): Float {
        val currentSize = bitmap.byteCount
        return if (currentSize > MAX_IMAGE_SIZE) {
            kotlin.math.sqrt(MAX_IMAGE_SIZE.toFloat() / currentSize)
        } else {
            1.0f
        }
    }
    
    /**
     * 创建临时音频文件
     */
    fun createTempAudioFile(): File {
        val audioDir = File(context.filesDir, "temp_audio")
        if (!audioDir.exists()) {
            audioDir.mkdirs()
        }
        return File(audioDir, "recording_${System.currentTimeMillis()}.aac")
    }
    
    /**
     * 清理临时文件
     */
    suspend fun cleanupTempFiles() = withContext(Dispatchers.IO) {
        try {
            val tempAudioDir = File(context.filesDir, "temp_audio")
            if (tempAudioDir.exists()) {
                tempAudioDir.listFiles()?.forEach { file ->
                    if (System.currentTimeMillis() - file.lastModified() > 24 * 60 * 60 * 1000) {
                        file.delete()
                    }
                }
            }
        } catch (e: Exception) {
            // 忽略清理异常
        }
    }
}