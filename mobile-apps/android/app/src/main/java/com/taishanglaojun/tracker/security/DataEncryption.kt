package com.taishanglaojun.tracker.security

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import com.google.gson.Gson
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.model.Trajectory
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import kotlin.random.Random

/**
 * 数据加密管理器
 * 负责位置数据的加密存储和安全传输
 */
class DataEncryption private constructor(private val context: Context) {
    
    companion object {
        @Volatile
        private var INSTANCE: DataEncryption? = null
        
        private const val ANDROID_KEYSTORE = "AndroidKeyStore"
        private const val KEY_ALIAS = "TaishanglaojunTrackerKey"
        private const val TRANSFORMATION = "AES/GCM/NoPadding"
        private const val GCM_IV_LENGTH = 12
        private const val GCM_TAG_LENGTH = 16
        
        fun getInstance(context: Context): DataEncryption {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: DataEncryption(context.applicationContext).also { INSTANCE = it }
            }
        }
    }
    
    private val gson = Gson()
    private val keyStore: KeyStore = KeyStore.getInstance(ANDROID_KEYSTORE)
    private val masterKey: MasterKey
    private val encryptedSharedPreferences: android.content.SharedPreferences
    
    init {
        keyStore.load(null)
        
        // 创建主密钥
        masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
        
        // 创建加密的SharedPreferences
        encryptedSharedPreferences = EncryptedSharedPreferences.create(
            context,
            "encrypted_prefs",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )
        
        // 生成或获取加密密钥
        generateOrGetSecretKey()
    }
    
    /**
     * 生成或获取密钥
     */
    private fun generateOrGetSecretKey(): SecretKey {
        return if (keyStore.containsAlias(KEY_ALIAS)) {
            keyStore.getKey(KEY_ALIAS, null) as SecretKey
        } else {
            val keyGenerator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, ANDROID_KEYSTORE)
            val keyGenParameterSpec = KeyGenParameterSpec.Builder(
                KEY_ALIAS,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
            )
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setRandomizedEncryptionRequired(true)
                .build()
            
            keyGenerator.init(keyGenParameterSpec)
            keyGenerator.generateKey()
        }
    }
    
    /**
     * 加密数据
     */
    fun encryptData(data: String): EncryptedData {
        try {
            val secretKey = keyStore.getKey(KEY_ALIAS, null) as SecretKey
            val cipher = Cipher.getInstance(TRANSFORMATION)
            cipher.init(Cipher.ENCRYPT_MODE, secretKey)
            
            val iv = cipher.iv
            val encryptedBytes = cipher.doFinal(data.toByteArray(Charsets.UTF_8))
            
            return EncryptedData(
                encryptedData = Base64.encodeToString(encryptedBytes, Base64.DEFAULT),
                iv = Base64.encodeToString(iv, Base64.DEFAULT)
            )
        } catch (e: Exception) {
            throw SecurityException("Failed to encrypt data", e)
        }
    }
    
    /**
     * 解密数据
     */
    fun decryptData(encryptedData: EncryptedData): String {
        try {
            val secretKey = keyStore.getKey(KEY_ALIAS, null) as SecretKey
            val cipher = Cipher.getInstance(TRANSFORMATION)
            
            val iv = Base64.decode(encryptedData.iv, Base64.DEFAULT)
            val gcmParameterSpec = GCMParameterSpec(GCM_TAG_LENGTH * 8, iv)
            cipher.init(Cipher.DECRYPT_MODE, secretKey, gcmParameterSpec)
            
            val encryptedBytes = Base64.decode(encryptedData.encryptedData, Base64.DEFAULT)
            val decryptedBytes = cipher.doFinal(encryptedBytes)
            
            return String(decryptedBytes, Charsets.UTF_8)
        } catch (e: Exception) {
            throw SecurityException("Failed to decrypt data", e)
        }
    }
    
    /**
     * 加密位置点
     */
    fun encryptLocationPoint(locationPoint: LocationPoint): EncryptedLocationPoint {
        val json = gson.toJson(locationPoint)
        val encryptedData = encryptData(json)
        
        return EncryptedLocationPoint(
            id = locationPoint.id,
            trajectoryId = locationPoint.trajectoryId,
            timestamp = locationPoint.timestamp,
            encryptedData = encryptedData.encryptedData,
            iv = encryptedData.iv,
            checksum = calculateChecksum(json)
        )
    }
    
    /**
     * 解密位置点
     */
    fun decryptLocationPoint(encryptedLocationPoint: EncryptedLocationPoint): LocationPoint {
        val encryptedData = EncryptedData(
            encryptedData = encryptedLocationPoint.encryptedData,
            iv = encryptedLocationPoint.iv
        )
        
        val json = decryptData(encryptedData)
        
        // 验证校验和
        val calculatedChecksum = calculateChecksum(json)
        if (calculatedChecksum != encryptedLocationPoint.checksum) {
            throw SecurityException("Data integrity check failed")
        }
        
        return gson.fromJson(json, LocationPoint::class.java)
    }
    
    /**
     * 加密轨迹
     */
    fun encryptTrajectory(trajectory: Trajectory): EncryptedTrajectory {
        val json = gson.toJson(trajectory)
        val encryptedData = encryptData(json)
        
        return EncryptedTrajectory(
            id = trajectory.id,
            name = trajectory.name,
            startTime = trajectory.startTime,
            endTime = trajectory.endTime,
            encryptedData = encryptedData.encryptedData,
            iv = encryptedData.iv,
            checksum = calculateChecksum(json)
        )
    }
    
    /**
     * 解密轨迹
     */
    fun decryptTrajectory(encryptedTrajectory: EncryptedTrajectory): Trajectory {
        val encryptedData = EncryptedData(
            encryptedData = encryptedTrajectory.encryptedData,
            iv = encryptedTrajectory.iv
        )
        
        val json = decryptData(encryptedData)
        
        // 验证校验和
        val calculatedChecksum = calculateChecksum(json)
        if (calculatedChecksum != encryptedTrajectory.checksum) {
            throw SecurityException("Data integrity check failed")
        }
        
        return gson.fromJson(json, Trajectory::class.java)
    }
    
    /**
     * 安全存储敏感配置
     */
    fun storeSecureConfig(key: String, value: String) {
        encryptedSharedPreferences.edit()
            .putString(key, value)
            .apply()
    }
    
    /**
     * 获取安全存储的配置
     */
    fun getSecureConfig(key: String, defaultValue: String? = null): String? {
        return encryptedSharedPreferences.getString(key, defaultValue)
    }
    
    /**
     * 删除安全存储的配置
     */
    fun removeSecureConfig(key: String) {
        encryptedSharedPreferences.edit()
            .remove(key)
            .apply()
    }
    
    /**
     * 生成传输用的临时密钥
     */
    fun generateTransportKey(): String {
        val keyBytes = ByteArray(32) // 256位密钥
        Random.nextBytes(keyBytes)
        return Base64.encodeToString(keyBytes, Base64.NO_WRAP)
    }
    
    /**
     * 使用传输密钥加密数据
     */
    fun encryptForTransport(data: String, transportKey: String): TransportEncryptedData {
        try {
            val keyBytes = Base64.decode(transportKey, Base64.NO_WRAP)
            val secretKey = javax.crypto.spec.SecretKeySpec(keyBytes, "AES")
            
            val cipher = Cipher.getInstance(TRANSFORMATION)
            cipher.init(Cipher.ENCRYPT_MODE, secretKey)
            
            val iv = cipher.iv
            val encryptedBytes = cipher.doFinal(data.toByteArray(Charsets.UTF_8))
            
            return TransportEncryptedData(
                encryptedData = Base64.encodeToString(encryptedBytes, Base64.NO_WRAP),
                iv = Base64.encodeToString(iv, Base64.NO_WRAP),
                timestamp = System.currentTimeMillis(),
                checksum = calculateChecksum(data)
            )
        } catch (e: Exception) {
            throw SecurityException("Failed to encrypt data for transport", e)
        }
    }
    
    /**
     * 使用传输密钥解密数据
     */
    fun decryptFromTransport(transportData: TransportEncryptedData, transportKey: String): String {
        try {
            val keyBytes = Base64.decode(transportKey, Base64.NO_WRAP)
            val secretKey = javax.crypto.spec.SecretKeySpec(keyBytes, "AES")
            
            val cipher = Cipher.getInstance(TRANSFORMATION)
            val iv = Base64.decode(transportData.iv, Base64.NO_WRAP)
            val gcmParameterSpec = GCMParameterSpec(GCM_TAG_LENGTH * 8, iv)
            cipher.init(Cipher.DECRYPT_MODE, secretKey, gcmParameterSpec)
            
            val encryptedBytes = Base64.decode(transportData.encryptedData, Base64.NO_WRAP)
            val decryptedBytes = cipher.doFinal(encryptedBytes)
            val decryptedData = String(decryptedBytes, Charsets.UTF_8)
            
            // 验证校验和
            val calculatedChecksum = calculateChecksum(decryptedData)
            if (calculatedChecksum != transportData.checksum) {
                throw SecurityException("Transport data integrity check failed")
            }
            
            return decryptedData
        } catch (e: Exception) {
            throw SecurityException("Failed to decrypt data from transport", e)
        }
    }
    
    /**
     * 计算数据校验和
     */
    private fun calculateChecksum(data: String): String {
        val digest = java.security.MessageDigest.getInstance("SHA-256")
        val hashBytes = digest.digest(data.toByteArray(Charsets.UTF_8))
        return Base64.encodeToString(hashBytes, Base64.NO_WRAP).take(16)
    }
    
    /**
     * 清除所有密钥和加密数据
     */
    fun clearAllEncryptedData() {
        try {
            // 删除密钥
            if (keyStore.containsAlias(KEY_ALIAS)) {
                keyStore.deleteEntry(KEY_ALIAS)
            }
            
            // 清除加密的SharedPreferences
            encryptedSharedPreferences.edit().clear().apply()
            
            // 重新生成密钥
            generateOrGetSecretKey()
        } catch (e: Exception) {
            throw SecurityException("Failed to clear encrypted data", e)
        }
    }
    
    /**
     * 验证数据完整性
     */
    fun verifyDataIntegrity(originalData: String, checksum: String): Boolean {
        val calculatedChecksum = calculateChecksum(originalData)
        return calculatedChecksum == checksum
    }
}

/**
 * 加密数据结构
 */
data class EncryptedData(
    val encryptedData: String,
    val iv: String
)

/**
 * 加密的位置点
 */
data class EncryptedLocationPoint(
    val id: String,
    val trajectoryId: String,
    val timestamp: Long,
    val encryptedData: String,
    val iv: String,
    val checksum: String
)

/**
 * 加密的轨迹
 */
data class EncryptedTrajectory(
    val id: String,
    val name: String,
    val startTime: Long?,
    val endTime: Long?,
    val encryptedData: String,
    val iv: String,
    val checksum: String
)

/**
 * 传输加密数据
 */
data class TransportEncryptedData(
    val encryptedData: String,
    val iv: String,
    val timestamp: Long,
    val checksum: String
)