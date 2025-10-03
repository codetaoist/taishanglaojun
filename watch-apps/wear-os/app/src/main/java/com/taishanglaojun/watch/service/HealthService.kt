package com.taishanglaojun.watch.service

import android.content.Context
import androidx.health.services.client.HealthServices
import androidx.health.services.client.HealthServicesClient
import androidx.health.services.client.data.DataType
import androidx.health.services.client.data.ExerciseCapabilities
import androidx.health.services.client.data.ExerciseConfig
import androidx.health.services.client.data.ExerciseGoal
import androidx.health.services.client.data.ExerciseType
import androidx.health.services.client.data.WarmUpConfig
import androidx.health.services.client.ExerciseClient
import androidx.health.services.client.PassiveMonitoringClient
import androidx.health.services.client.data.PassiveMonitoringConfig
import androidx.health.services.client.data.PassiveGoal
import androidx.health.services.client.data.ComparisonType
import androidx.health.services.client.data.ExerciseUpdate
import androidx.health.services.client.data.LocationAvailability
import androidx.health.services.client.data.ExerciseState
import androidx.health.services.client.data.DataPointContainer
import androidx.health.services.client.data.AggregateDataType
import androidx.health.services.client.data.TimeRangeFilter
import androidx.health.services.client.MeasureClient
import androidx.health.services.client.data.MeasureCapabilities
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import java.time.Instant
import java.time.Duration
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class HealthService @Inject constructor(
    @ApplicationContext private val context: Context
) {
    
    private val healthServicesClient: HealthServicesClient = HealthServices.getClient(context)
    private val exerciseClient: ExerciseClient = healthServicesClient.exerciseClient
    private val passiveMonitoringClient: PassiveMonitoringClient = healthServicesClient.passiveMonitoringClient
    private val measureClient: MeasureClient = healthServicesClient.measureClient
    
    private val scope = CoroutineScope(Dispatchers.IO)
    
    // Health metrics state
    private val _healthMetrics = MutableStateFlow(HealthMetrics())
    val healthMetrics: StateFlow<HealthMetrics> = _healthMetrics.asStateFlow()
    
    private val _exerciseState = MutableStateFlow<ExerciseState?>(null)
    val exerciseState: StateFlow<ExerciseState?> = _exerciseState.asStateFlow()
    
    private val _isTrackingExercise = MutableStateFlow(false)
    val isTrackingExercise: StateFlow<Boolean> = _isTrackingExercise.asStateFlow()
    
    private val _heartRateAvailable = MutableStateFlow(false)
    val heartRateAvailable: StateFlow<Boolean> = _heartRateAvailable.asStateFlow()
    
    init {
        checkCapabilities()
        setupPassiveMonitoring()
    }
    
    // MARK: - Capabilities Check
    
    private fun checkCapabilities() {
        scope.launch {
            try {
                // Check exercise capabilities
                val exerciseCapabilities = exerciseClient.getCapabilitiesAsync().await()
                
                // Check measure capabilities (for heart rate)
                val measureCapabilities = measureClient.getCapabilitiesAsync().await()
                _heartRateAvailable.value = DataType.HEART_RATE_BPM in measureCapabilities.supportedDataTypesMeasure
                
                // Log available capabilities
                println("Exercise types supported: ${exerciseCapabilities.supportedExerciseTypes}")
                println("Heart rate available: ${_heartRateAvailable.value}")
                
            } catch (e: Exception) {
                println("Error checking capabilities: ${e.message}")
            }
        }
    }
    
    // MARK: - Passive Monitoring
    
    private fun setupPassiveMonitoring() {
        scope.launch {
            try {
                val passiveConfig = PassiveMonitoringConfig.builder()
                    .setDataTypes(
                        setOf(
                            DataType.STEPS_DAILY,
                            DataType.CALORIES_DAILY,
                            DataType.DISTANCE_DAILY,
                            DataType.HEART_RATE_BPM
                        )
                    )
                    .setShouldUserActivityInfoBeRequested(true)
                    .build()
                
                passiveMonitoringClient.setPassiveListenerServiceAsync(
                    HealthDataService::class.java,
                    passiveConfig
                ).await()
                
                println("Passive monitoring setup completed")
                
            } catch (e: Exception) {
                println("Error setting up passive monitoring: ${e.message}")
            }
        }
    }
    
    // MARK: - Exercise Tracking
    
    fun startExerciseTracking(exerciseType: ExerciseType = ExerciseType.WALKING) {
        scope.launch {
            try {
                val exerciseConfig = ExerciseConfig.builder(exerciseType)
                    .setDataTypes(
                        setOf(
                            DataType.HEART_RATE_BPM,
                            DataType.STEPS,
                            DataType.CALORIES_TOTAL,
                            DataType.DISTANCE_TOTAL,
                            DataType.PACE,
                            DataType.SPEED
                        )
                    )
                    .setIsAutoPauseAndResumeEnabled(true)
                    .setIsGpsEnabled(true)
                    .build()
                
                exerciseClient.startExerciseAsync(exerciseConfig).await()
                _isTrackingExercise.value = true
                
                println("Exercise tracking started: $exerciseType")
                
            } catch (e: Exception) {
                println("Error starting exercise tracking: ${e.message}")
            }
        }
    }
    
    fun pauseExerciseTracking() {
        scope.launch {
            try {
                exerciseClient.pauseExerciseAsync().await()
                println("Exercise tracking paused")
            } catch (e: Exception) {
                println("Error pausing exercise tracking: ${e.message}")
            }
        }
    }
    
    fun resumeExerciseTracking() {
        scope.launch {
            try {
                exerciseClient.resumeExerciseAsync().await()
                println("Exercise tracking resumed")
            } catch (e: Exception) {
                println("Error resuming exercise tracking: ${e.message}")
            }
        }
    }
    
    fun stopExerciseTracking() {
        scope.launch {
            try {
                exerciseClient.endExerciseAsync().await()
                _isTrackingExercise.value = false
                _exerciseState.value = null
                
                println("Exercise tracking stopped")
                
            } catch (e: Exception) {
                println("Error stopping exercise tracking: ${e.message}")
            }
        }
    }
    
    // MARK: - Heart Rate Monitoring
    
    fun startHeartRateMonitoring() {
        if (!_heartRateAvailable.value) return
        
        scope.launch {
            try {
                measureClient.registerMeasureCallback(
                    DataType.HEART_RATE_BPM,
                    HeartRateCallback()
                )
                
                println("Heart rate monitoring started")
                
            } catch (e: Exception) {
                println("Error starting heart rate monitoring: ${e.message}")
            }
        }
    }
    
    fun stopHeartRateMonitoring() {
        scope.launch {
            try {
                measureClient.unregisterMeasureCallbackAsync(
                    DataType.HEART_RATE_BPM,
                    HeartRateCallback()
                ).await()
                
                println("Heart rate monitoring stopped")
                
            } catch (e: Exception) {
                println("Error stopping heart rate monitoring: ${e.message}")
            }
        }
    }
    
    // MARK: - Health Data Retrieval
    
    fun getTodayHealthData() {
        scope.launch {
            try {
                val now = Instant.now()
                val startOfDay = now.truncatedTo(java.time.temporal.ChronoUnit.DAYS)
                
                val timeRangeFilter = TimeRangeFilter.between(startOfDay, now)
                
                // Get steps
                val stepsResponse = passiveMonitoringClient.getDataPointsAsync(
                    setOf(AggregateDataType.STEPS_DAILY),
                    timeRangeFilter
                ).await()
                
                // Get calories
                val caloriesResponse = passiveMonitoringClient.getDataPointsAsync(
                    setOf(AggregateDataType.CALORIES_DAILY),
                    timeRangeFilter
                ).await()
                
                // Get distance
                val distanceResponse = passiveMonitoringClient.getDataPointsAsync(
                    setOf(AggregateDataType.DISTANCE_DAILY),
                    timeRangeFilter
                ).await()
                
                // Update health metrics
                val currentMetrics = _healthMetrics.value
                _healthMetrics.value = currentMetrics.copy(
                    steps = stepsResponse.firstOrNull()?.total?.toInt() ?: currentMetrics.steps,
                    calories = caloriesResponse.firstOrNull()?.total?.toInt() ?: currentMetrics.calories,
                    distance = distanceResponse.firstOrNull()?.total ?: currentMetrics.distance,
                    lastUpdated = now
                )
                
            } catch (e: Exception) {
                println("Error getting today's health data: ${e.message}")
            }
        }
    }
    
    // MARK: - Task Integration
    
    fun getTaskHealthBonus(steps: Int, calories: Int, heartRate: Int?): Int {
        var bonus = 0
        
        // Steps bonus
        when {
            steps >= 10000 -> bonus += 50
            steps >= 5000 -> bonus += 25
            steps >= 2000 -> bonus += 10
        }
        
        // Calories bonus
        when {
            calories >= 500 -> bonus += 30
            calories >= 300 -> bonus += 20
            calories >= 150 -> bonus += 10
        }
        
        // Heart rate bonus (if in healthy range during activity)
        heartRate?.let { hr ->
            when {
                hr in 120..160 -> bonus += 20 // Active range
                hr in 100..119 -> bonus += 10 // Moderate range
            }
        }
        
        return bonus
    }
    
    fun isHealthGoalMet(): Boolean {
        val metrics = _healthMetrics.value
        return metrics.steps >= 8000 || metrics.calories >= 300
    }
    
    // MARK: - Data Classes
    
    data class HealthMetrics(
        val steps: Int = 0,
        val calories: Int = 0,
        val distance: Double = 0.0, // in meters
        val heartRate: Int? = null,
        val activeMinutes: Int = 0,
        val lastUpdated: Instant = Instant.now()
    )
    
    // MARK: - Callbacks
    
    private inner class HeartRateCallback : MeasureClient.MeasureCallback {
        override fun onAvailabilityChanged(
            dataType: DataType<*, *>,
            availability: androidx.health.services.client.data.Availability
        ) {
            println("Heart rate availability changed: $availability")
        }
        
        override fun onDataReceived(data: DataPointContainer) {
            val heartRateData = data.getData(DataType.HEART_RATE_BPM)
            heartRateData.lastOrNull()?.let { dataPoint ->
                val heartRate = dataPoint.value.toInt()
                val currentMetrics = _healthMetrics.value
                _healthMetrics.value = currentMetrics.copy(
                    heartRate = heartRate,
                    lastUpdated = Instant.now()
                )
                
                println("Heart rate updated: $heartRate BPM")
            }
        }
    }
    
    private inner class ExerciseUpdateCallback : ExerciseClient.ExerciseUpdateCallback {
        override fun onExerciseUpdateReceived(update: ExerciseUpdate) {
            _exerciseState.value = update.exerciseStateInfo.state
            
            // Update metrics from exercise data
            val latestMetrics = update.latestMetrics
            val currentMetrics = _healthMetrics.value
            
            val steps = latestMetrics.getData(DataType.STEPS).lastOrNull()?.value?.toInt()
            val calories = latestMetrics.getData(DataType.CALORIES_TOTAL).lastOrNull()?.value?.toInt()
            val distance = latestMetrics.getData(DataType.DISTANCE_TOTAL).lastOrNull()?.value
            val heartRate = latestMetrics.getData(DataType.HEART_RATE_BPM).lastOrNull()?.value?.toInt()
            
            _healthMetrics.value = currentMetrics.copy(
                steps = steps ?: currentMetrics.steps,
                calories = calories ?: currentMetrics.calories,
                distance = distance ?: currentMetrics.distance,
                heartRate = heartRate ?: currentMetrics.heartRate,
                lastUpdated = Instant.now()
            )
        }
        
        override fun onLapSummaryReceived(lapSummary: androidx.health.services.client.data.ExerciseLapSummary) {
            println("Lap summary received: $lapSummary")
        }
        
        override fun onRegistered() {
            println("Exercise update callback registered")
        }
        
        override fun onRegistrationFailed(throwable: Throwable) {
            println("Exercise update callback registration failed: ${throwable.message}")
        }
        
        override fun onAvailabilityChanged(
            dataType: DataType<*, *>,
            availability: androidx.health.services.client.data.Availability
        ) {
            println("Exercise data availability changed: $dataType -> $availability")
        }
    }
}