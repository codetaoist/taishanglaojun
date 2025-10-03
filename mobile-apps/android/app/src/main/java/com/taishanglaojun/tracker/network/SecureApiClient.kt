package com.taishanglaojun.tracker.network

import android.content.Context
import com.google.gson.Gson
import com.google.gson.GsonBuilder
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.model.Trajectory
import com.taishanglaojun.tracker.security.DataEncryption
import com.taishanglaojun.tracker.security.TransportEncryptedData
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.logging.HttpLoggingInterceptor
import java.io.IOException
import java.security.cert.CertificateException
import java.util.concurrent.TimeUnit
import javax.net.ssl.*

/**
 * 安全API客户端
 * 负责与后端服务的安全通信
 */
class SecureApiClient private constructor(private val context: Context) {
    
    companion object {
        @Volatile
        private var INSTANCE: SecureApiClient? = null
        
        private const val BASE_URL = "https://api.taishanglaojun.com"
        private const val CONNECT_TIMEOUT = 30L
        private const val READ_TIMEOUT = 30L
        private const val WRITE_TIMEOUT = 30L
        
        fun getInstance(context: Context): SecureApiClient {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: SecureApiClient(context.applicationContext).also { INSTANCE = it }
            }
        }
    }
    
    private val gson: Gson = GsonBuilder()
        .setDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'")
        .create()
    
    private val dataEncryption = DataEncryption.getInstance(context)
    private val okHttpClient: OkHttpClient
    
    init {
        okHttpClient = createSecureOkHttpClient()
    }
    
    /**
     * 创建安全的OkHttp客户端
     */
    private fun createSecureOkHttpClient(): OkHttpClient {
        val builder = OkHttpClient.Builder()
            .connectTimeout(CONNECT_TIMEOUT, TimeUnit.SECONDS)
            .readTimeout(READ_TIMEOUT, TimeUnit.SECONDS)
            .writeTimeout(WRITE_TIMEOUT, TimeUnit.SECONDS)
            .retryOnConnectionFailure(true)
        
        // 添加日志拦截器（仅在调试模式下）
        if (BuildConfig.DEBUG) {
            val loggingInterceptor = HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BODY
            }
            builder.addInterceptor(loggingInterceptor)
        }
        
        // 添加认证拦截器
        builder.addInterceptor(AuthInterceptor())
        
        // 添加加密拦截器
        builder.addInterceptor(EncryptionInterceptor())
        
        // 配置SSL/TLS
        configureSsl(builder)
        
        return builder.build()
    }
    
    /**
     * 配置SSL/TLS安全连接
     */
    private fun configureSsl(builder: OkHttpClient.Builder) {
        try {
            // 创建信任所有证书的TrustManager（生产环境应使用证书固定）
            val trustAllCerts = arrayOf<TrustManager>(object : X509TrustManager {
                @Throws(CertificateException::class)
                override fun checkClientTrusted(chain: Array<java.security.cert.X509Certificate>, authType: String) {
                }
                
                @Throws(CertificateException::class)
                override fun checkServerTrusted(chain: Array<java.security.cert.X509Certificate>, authType: String) {
                    // 在生产环境中，这里应该验证服务器证书
                }
                
                override fun getAcceptedIssuers(): Array<java.security.cert.X509Certificate> {
                    return arrayOf()
                }
            })
            
            // 安装信任所有证书的TrustManager
            val sslContext = SSLContext.getInstance("SSL")
            sslContext.init(null, trustAllCerts, java.security.SecureRandom())
            
            // 创建SSL套接字工厂
            val sslSocketFactory = sslContext.socketFactory
            
            builder.sslSocketFactory(sslSocketFactory, trustAllCerts[0] as X509TrustManager)
            builder.hostnameVerifier { _, _ -> true }
            
        } catch (e: Exception) {
            throw RuntimeException("Failed to configure SSL", e)
        }
    }
    
    /**
     * 上传位置点
     */
    suspend fun uploadLocationPoints(locationPoints: List<LocationPoint>): ApiResponse<String> {
        return withContext(Dispatchers.IO) {
            try {
                val transportKey = dataEncryption.generateTransportKey()
                val encryptedPoints = locationPoints.map { point ->
                    val json = gson.toJson(point)
                    dataEncryption.encryptForTransport(json, transportKey)
                }
                
                val requestData = UploadLocationPointsRequest(
                    transportKey = transportKey,
                    encryptedPoints = encryptedPoints,
                    deviceId = getDeviceId(),
                    timestamp = System.currentTimeMillis()
                )
                
                val json = gson.toJson(requestData)
                val requestBody = json.toRequestBody("application/json".toMediaType())
                
                val request = Request.Builder()
                    .url("$BASE_URL/api/v1/location-points")
                    .post(requestBody)
                    .build()
                
                val response = okHttpClient.newCall(request).execute()
                handleResponse<String>(response)
                
            } catch (e: Exception) {
                ApiResponse.Error("Failed to upload location points: ${e.message}")
            }
        }
    }
    
    /**
     * 上传轨迹
     */
    suspend fun uploadTrajectory(trajectory: Trajectory): ApiResponse<String> {
        return withContext(Dispatchers.IO) {
            try {
                val transportKey = dataEncryption.generateTransportKey()
                val json = gson.toJson(trajectory)
                val encryptedTrajectory = dataEncryption.encryptForTransport(json, transportKey)
                
                val requestData = UploadTrajectoryRequest(
                    transportKey = transportKey,
                    encryptedTrajectory = encryptedTrajectory,
                    deviceId = getDeviceId(),
                    timestamp = System.currentTimeMillis()
                )
                
                val requestJson = gson.toJson(requestData)
                val requestBody = requestJson.toRequestBody("application/json".toMediaType())
                
                val request = Request.Builder()
                    .url("$BASE_URL/api/v1/trajectories")
                    .post(requestBody)
                    .build()
                
                val response = okHttpClient.newCall(request).execute()
                handleResponse<String>(response)
                
            } catch (e: Exception) {
                ApiResponse.Error("Failed to upload trajectory: ${e.message}")
            }
        }
    }
    
    /**
     * 下载轨迹列表
     */
    suspend fun downloadTrajectories(userId: String): ApiResponse<List<Trajectory>> {
        return withContext(Dispatchers.IO) {
            try {
                val request = Request.Builder()
                    .url("$BASE_URL/api/v1/trajectories?userId=$userId")
                    .get()
                    .build()
                
                val response = okHttpClient.newCall(request).execute()
                val apiResponse = handleResponse<DownloadTrajectoriesResponse>(response)
                
                when (apiResponse) {
                    is ApiResponse.Success -> {
                        val trajectories = apiResponse.data.encryptedTrajectories.map { encryptedData ->
                            val decryptedJson = dataEncryption.decryptFromTransport(
                                encryptedData, 
                                apiResponse.data.transportKey
                            )
                            gson.fromJson(decryptedJson, Trajectory::class.java)
                        }
                        ApiResponse.Success(trajectories)
                    }
                    is ApiResponse.Error -> ApiResponse.Error(apiResponse.message)
                }
                
            } catch (e: Exception) {
                ApiResponse.Error("Failed to download trajectories: ${e.message}")
            }
        }
    }
    
    /**
     * 下载轨迹的位置点
     */
    suspend fun downloadLocationPoints(trajectoryId: String): ApiResponse<List<LocationPoint>> {
        return withContext(Dispatchers.IO) {
            try {
                val request = Request.Builder()
                    .url("$BASE_URL/api/v1/trajectories/$trajectoryId/points")
                    .get()
                    .build()
                
                val response = okHttpClient.newCall(request).execute()
                val apiResponse = handleResponse<DownloadLocationPointsResponse>(response)
                
                when (apiResponse) {
                    is ApiResponse.Success -> {
                        val locationPoints = apiResponse.data.encryptedPoints.map { encryptedData ->
                            val decryptedJson = dataEncryption.decryptFromTransport(
                                encryptedData,
                                apiResponse.data.transportKey
                            )
                            gson.fromJson(decryptedJson, LocationPoint::class.java)
                        }
                        ApiResponse.Success(locationPoints)
                    }
                    is ApiResponse.Error -> ApiResponse.Error(apiResponse.message)
                }
                
            } catch (e: Exception) {
                ApiResponse.Error("Failed to download location points: ${e.message}")
            }
        }
    }
    
    /**
     * 删除轨迹
     */
    suspend fun deleteTrajectory(trajectoryId: String): ApiResponse<String> {
        return withContext(Dispatchers.IO) {
            try {
                val request = Request.Builder()
                    .url("$BASE_URL/api/v1/trajectories/$trajectoryId")
                    .delete()
                    .build()
                
                val response = okHttpClient.newCall(request).execute()
                handleResponse<String>(response)
                
            } catch (e: Exception) {
                ApiResponse.Error("Failed to delete trajectory: ${e.message}")
            }
        }
    }
    
    /**
     * 同步数据
     */
    suspend fun syncData(syncRequest: SyncRequest): ApiResponse<SyncResponse> {
        return withContext(Dispatchers.IO) {
            try {
                val json = gson.toJson(syncRequest)
                val requestBody = json.toRequestBody("application/json".toMediaType())
                
                val request = Request.Builder()
                    .url("$BASE_URL/api/v1/sync")
                    .post(requestBody)
                    .build()
                
                val response = okHttpClient.newCall(request).execute()
                handleResponse<SyncResponse>(response)
                
            } catch (e: Exception) {
                ApiResponse.Error("Failed to sync data: ${e.message}")
            }
        }
    }
    
    /**
     * 处理HTTP响应
     */
    private inline fun <reified T> handleResponse(response: Response): ApiResponse<T> {
        return try {
            if (response.isSuccessful) {
                val responseBody = response.body?.string()
                if (responseBody != null) {
                    val apiResponse = gson.fromJson(responseBody, ApiResponseWrapper::class.java)
                    if (apiResponse.success) {
                        val data = gson.fromJson(gson.toJson(apiResponse.data), T::class.java)
                        ApiResponse.Success(data)
                    } else {
                        ApiResponse.Error(apiResponse.message ?: "Unknown error")
                    }
                } else {
                    ApiResponse.Error("Empty response body")
                }
            } else {
                ApiResponse.Error("HTTP ${response.code}: ${response.message}")
            }
        } catch (e: Exception) {
            ApiResponse.Error("Failed to parse response: ${e.message}")
        } finally {
            response.close()
        }
    }
    
    /**
     * 获取设备ID
     */
    private fun getDeviceId(): String {
        return dataEncryption.getSecureConfig("device_id") ?: run {
            val deviceId = java.util.UUID.randomUUID().toString()
            dataEncryption.storeSecureConfig("device_id", deviceId)
            deviceId
        }
    }
    
    /**
     * 认证拦截器
     */
    private inner class AuthInterceptor : Interceptor {
        override fun intercept(chain: Interceptor.Chain): Response {
            val originalRequest = chain.request()
            
            val token = dataEncryption.getSecureConfig("auth_token")
            
            val newRequest = if (token != null) {
                originalRequest.newBuilder()
                    .addHeader("Authorization", "Bearer $token")
                    .addHeader("X-Device-ID", getDeviceId())
                    .build()
            } else {
                originalRequest.newBuilder()
                    .addHeader("X-Device-ID", getDeviceId())
                    .build()
            }
            
            return chain.proceed(newRequest)
        }
    }
    
    /**
     * 加密拦截器
     */
    private inner class EncryptionInterceptor : Interceptor {
        override fun intercept(chain: Interceptor.Chain): Response {
            val request = chain.request()
            
            // 添加加密相关的头部
            val newRequest = request.newBuilder()
                .addHeader("X-Encryption-Version", "1.0")
                .addHeader("X-Client-Version", BuildConfig.VERSION_NAME)
                .build()
            
            return chain.proceed(newRequest)
        }
    }
}

/**
 * API响应封装
 */
sealed class ApiResponse<out T> {
    data class Success<T>(val data: T) : ApiResponse<T>()
    data class Error(val message: String) : ApiResponse<Nothing>()
}

/**
 * API响应包装器
 */
data class ApiResponseWrapper(
    val success: Boolean,
    val message: String?,
    val data: Any?
)

/**
 * 上传位置点请求
 */
data class UploadLocationPointsRequest(
    val transportKey: String,
    val encryptedPoints: List<TransportEncryptedData>,
    val deviceId: String,
    val timestamp: Long
)

/**
 * 上传轨迹请求
 */
data class UploadTrajectoryRequest(
    val transportKey: String,
    val encryptedTrajectory: TransportEncryptedData,
    val deviceId: String,
    val timestamp: Long
)

/**
 * 下载轨迹响应
 */
data class DownloadTrajectoriesResponse(
    val transportKey: String,
    val encryptedTrajectories: List<TransportEncryptedData>
)

/**
 * 下载位置点响应
 */
data class DownloadLocationPointsResponse(
    val transportKey: String,
    val encryptedPoints: List<TransportEncryptedData>
)

/**
 * 同步请求
 */
data class SyncRequest(
    val lastSyncTime: Long,
    val deviceId: String,
    val trajectoryIds: List<String>
)

/**
 * 同步响应
 */
data class SyncResponse(
    val newTrajectories: List<Trajectory>,
    val updatedTrajectories: List<Trajectory>,
    val deletedTrajectoryIds: List<String>,
    val syncTime: Long
)