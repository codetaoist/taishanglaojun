package com.taishanglaojun.watch.service

import androidx.health.services.client.PassiveListenerService
import androidx.health.services.client.data.DataPointContainer
import androidx.health.services.client.data.DataType
import androidx.health.services.client.data.UserActivityInfo
import androidx.health.services.client.data.UserActivityState
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class HealthDataService : PassiveListenerService() {
    
    @Inject
    lateinit var healthService: HealthService
    
    @Inject
    lateinit var notificationService: NotificationService
    
    override fun onNewDataPointsReceived(dataPoints: DataPointContainer) {
        super.onNewDataPointsReceived(dataPoints)
        
        // Process different types of health data
        processStepsData(dataPoints)
        processCaloriesData(dataPoints)
        processDistanceData(dataPoints)
        processHeartRateData(dataPoints)
    }
    
    override fun onUserActivityInfoReceived(info: UserActivityInfo) {
        super.onUserActivityInfoReceived(info)
        
        // Handle user activity state changes
        when (info.userActivityState) {
            UserActivityState.USER_ACTIVITY_ASLEEP -> {
                // User is sleeping, pause active monitoring
                println("User is sleeping")
            }
            
            UserActivityState.USER_ACTIVITY_PASSIVE -> {
                // User is in passive state
                println("User activity: passive")
            }
            
            UserActivityState.USER_ACTIVITY_ACTIVE -> {
                // User is active, might want to suggest task completion
                println("User activity: active")
                checkForTaskOpportunities()
            }
            
            else -> {
                println("User activity state: ${info.userActivityState}")
            }
        }
    }
    
    private fun processStepsData(dataPoints: DataPointContainer) {
        val stepsData = dataPoints.getData(DataType.STEPS_DAILY)
        stepsData.forEach { dataPoint ->
            val steps = dataPoint.value.toInt()
            println("Steps updated: $steps")
            
            // Check for step milestones
            checkStepMilestones(steps)
        }
    }
    
    private fun processCaloriesData(dataPoints: DataPointContainer) {
        val caloriesData = dataPoints.getData(DataType.CALORIES_DAILY)
        caloriesData.forEach { dataPoint ->
            val calories = dataPoint.value.toInt()
            println("Calories updated: $calories")
            
            // Check for calorie milestones
            checkCalorieMilestones(calories)
        }
    }
    
    private fun processDistanceData(dataPoints: DataPointContainer) {
        val distanceData = dataPoints.getData(DataType.DISTANCE_DAILY)
        distanceData.forEach { dataPoint ->
            val distance = dataPoint.value
            println("Distance updated: ${distance}m")
        }
    }
    
    private fun processHeartRateData(dataPoints: DataPointContainer) {
        val heartRateData = dataPoints.getData(DataType.HEART_RATE_BPM)
        heartRateData.forEach { dataPoint ->
            val heartRate = dataPoint.value.toInt()
            println("Heart rate updated: $heartRate BPM")
            
            // Check for heart rate alerts
            checkHeartRateAlerts(heartRate)
        }
    }
    
    private fun checkStepMilestones(steps: Int) {
        val milestones = listOf(1000, 2500, 5000, 7500, 10000, 12500, 15000)
        
        milestones.forEach { milestone ->
            if (steps == milestone) {
                // Show milestone notification
                showMilestoneNotification("步数里程碑", "恭喜！您已完成 $milestone 步")
            }
        }
    }
    
    private fun checkCalorieMilestones(calories: Int) {
        val milestones = listOf(100, 250, 500, 750, 1000)
        
        milestones.forEach { milestone ->
            if (calories == milestone) {
                // Show milestone notification
                showMilestoneNotification("卡路里里程碑", "恭喜！您已燃烧 $milestone 卡路里")
            }
        }
    }
    
    private fun checkHeartRateAlerts(heartRate: Int) {
        when {
            heartRate > 180 -> {
                // High heart rate alert
                showHealthAlert("心率过高", "当前心率 $heartRate BPM，请注意休息")
            }
            
            heartRate < 50 -> {
                // Low heart rate alert (when not sleeping)
                showHealthAlert("心率过低", "当前心率 $heartRate BPM，如有不适请咨询医生")
            }
        }
    }
    
    private fun checkForTaskOpportunities() {
        // This could integrate with TaskRepository to suggest tasks
        // based on current activity level and location
        println("Checking for task opportunities based on activity")
    }
    
    private fun showMilestoneNotification(title: String, message: String) {
        // Create a simple milestone notification
        // This would integrate with NotificationService
        println("Milestone: $title - $message")
    }
    
    private fun showHealthAlert(title: String, message: String) {
        // Create a health alert notification
        // This would integrate with NotificationService
        println("Health Alert: $title - $message")
    }
}