package com.taishanglaojun.watch.service

import android.Manifest
import android.app.Service
import android.content.Intent
import android.content.pm.PackageManager
import android.location.Location
import android.os.IBinder
import androidx.core.app.ActivityCompat
import com.google.android.gms.location.*
import com.google.android.gms.tasks.Task
import com.taishanglaojun.watch.data.repository.TaskRepository
import com.taishanglaojun.watch.data.models.TaskCoordinate
import com.taishanglaojun.watch.data.models.WatchTask
import com.taishanglaojun.watch.services.ConnectivityService
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject
import kotlin.math.*

@AndroidEntryPoint
class LocationService : Service() {

    @Inject
    lateinit var taskRepository: TaskRepository

    @Inject
    lateinit var connectivityService: ConnectivityService

    private lateinit var fusedLocationClient: FusedLocationProviderClient
    private lateinit var geofencingClient: GeofencingClient
    
    private val serviceScope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    
    // Location state
    private val _currentLocation = MutableStateFlow<Location?>(null)
    val currentLocation: StateFlow<Location?> = _currentLocation.asStateFlow()
    
    private val _isTracking = MutableStateFlow(false)
    val isTracking: StateFlow<Boolean> = _isTracking.asStateFlow()
    
    private val _nearbyTasks = MutableStateFlow<List<WatchTask>>(emptyList())
    val nearbyTasks: StateFlow<List<WatchTask>> = _nearbyTasks.asStateFlow()
    
    // Location request configuration
    private val locationRequest = LocationRequest.Builder(
        Priority.PRIORITY_HIGH_ACCURACY,
        10000L // 10 seconds
    ).apply {
        setMinUpdateIntervalMillis(5000L) // 5 seconds
        setMaxUpdateDelayMillis(30000L) // 30 seconds
    }.build()
    
    private val locationCallback = object : LocationCallback() {
        override fun onLocationResult(locationResult: LocationResult) {
            locationResult.lastLocation?.let { location ->
                _currentLocation.value = location
                checkNearbyTasks(location)
                
                // Send location update to phone if connected
                serviceScope.launch {
                    if (connectivityService.connectionState.value.isConnected) {
                        // Send location update (implement as needed)
                    }
                }
            }
        }
    }
    
    override fun onCreate() {
        super.onCreate()
        fusedLocationClient = LocationServices.getFusedLocationProviderClient(this)
        geofencingClient = LocationServices.getGeofencingClient(this)
    }
    
    override fun onBind(intent: Intent?): IBinder? = null
    
    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START_TRACKING -> startLocationTracking()
            ACTION_STOP_TRACKING -> stopLocationTracking()
            ACTION_UPDATE_GEOFENCES -> updateGeofences()
        }
        return START_STICKY
    }
    
    override fun onDestroy() {
        super.onDestroy()
        stopLocationTracking()
        serviceScope.cancel()
    }
    
    private fun startLocationTracking() {
        if (!hasLocationPermission()) {
            return
        }
        
        try {
            fusedLocationClient.requestLocationUpdates(
                locationRequest,
                locationCallback,
                null
            )
            _isTracking.value = true
            
            // Get last known location
            fusedLocationClient.lastLocation.addOnSuccessListener { location ->
                location?.let {
                    _currentLocation.value = it
                    checkNearbyTasks(it)
                }
            }
            
        } catch (e: SecurityException) {
            // Handle permission error
        }
    }
    
    private fun stopLocationTracking() {
        fusedLocationClient.removeLocationUpdates(locationCallback)
        _isTracking.value = false
    }
    
    private fun hasLocationPermission(): Boolean {
        return ActivityCompat.checkSelfPermission(
            this,
            Manifest.permission.ACCESS_FINE_LOCATION
        ) == PackageManager.PERMISSION_GRANTED
    }
    
    private fun checkNearbyTasks(location: Location) {
        serviceScope.launch {
            try {
                val allTasks = taskRepository.getAllTasks().value
                val nearby = allTasks.filter { task ->
                    task.coordinate?.let { coord ->
                        val distance = calculateDistance(
                            location.latitude,
                            location.longitude,
                            coord.latitude,
                            coord.longitude
                        )
                        distance <= PROXIMITY_THRESHOLD_METERS
                    } ?: false
                }
                
                _nearbyTasks.value = nearby
                
                // Check for task arrival notifications
                nearby.forEach { task ->
                    if (shouldNotifyTaskArrival(task, location)) {
                        notifyTaskArrival(task)
                    }
                }
                
            } catch (e: Exception) {
                // Handle error
            }
        }
    }
    
    private fun calculateDistance(
        lat1: Double, lon1: Double,
        lat2: Double, lon2: Double
    ): Double {
        val earthRadius = 6371000.0 // Earth radius in meters
        
        val dLat = Math.toRadians(lat2 - lat1)
        val dLon = Math.toRadians(lon2 - lon1)
        
        val a = sin(dLat / 2).pow(2) +
                cos(Math.toRadians(lat1)) * cos(Math.toRadians(lat2)) *
                sin(dLon / 2).pow(2)
        
        val c = 2 * atan2(sqrt(a), sqrt(1 - a))
        
        return earthRadius * c
    }
    
    private fun shouldNotifyTaskArrival(task: WatchTask, location: Location): Boolean {
        // Check if we should notify about arriving at task location
        // This could be based on task status, user preferences, etc.
        return task.status == com.taishanglaojun.watch.model.TaskStatus.ACCEPTED ||
               task.status == com.taishanglaojun.watch.model.TaskStatus.IN_PROGRESS
    }
    
    private fun notifyTaskArrival(task: WatchTask) {
        // Send notification about arriving at task location
        // This would integrate with the notification system
    }
    
    private fun updateGeofences() {
        if (!hasLocationPermission()) {
            return
        }
        
        serviceScope.launch {
            try {
                val tasks = taskRepository.getAllTasks().value
                val geofenceRequests = tasks.mapNotNull { task ->
                    task.coordinate?.let { coord ->
                        createGeofenceRequest(task.id, coord)
                    }
                }
                
                if (geofenceRequests.isNotEmpty()) {
                    // Remove existing geofences first
                    geofencingClient.removeGeofences(geofenceRequests.map { it.requestId })
                    
                    // Add new geofences
                    val geofencingRequest = GeofencingRequest.Builder().apply {
                        setInitialTrigger(GeofencingRequest.INITIAL_TRIGGER_ENTER)
                        addGeofences(geofenceRequests)
                    }.build()
                    
                    // Note: This would need a PendingIntent for geofence transitions
                    // geofencingClient.addGeofences(geofencingRequest, geofencePendingIntent)
                }
                
            } catch (e: Exception) {
                // Handle geofencing error
            }
        }
    }
    
    private fun createGeofenceRequest(taskId: String, coordinate: TaskCoordinate): Geofence {
        return Geofence.Builder()
            .setRequestId(taskId)
            .setCircularRegion(
                coordinate.latitude,
                coordinate.longitude,
                GEOFENCE_RADIUS_METERS
            )
            .setExpirationDuration(Geofence.NEVER_EXPIRE)
            .setTransitionTypes(
                Geofence.GEOFENCE_TRANSITION_ENTER or 
                Geofence.GEOFENCE_TRANSITION_EXIT
            )
            .build()
    }
    
    // Public API methods
    fun getCurrentLocation(): Location? = _currentLocation.value
    
    fun getNearbyTasks(): List<WatchTask> = _nearbyTasks.value
    
    fun getDistanceToTask(task: WatchTask): Double? {
        val currentLoc = _currentLocation.value
        val taskCoord = task.coordinate
        
        return if (currentLoc != null && taskCoord != null) {
            calculateDistance(
                currentLoc.latitude,
                currentLoc.longitude,
                taskCoord.latitude,
                taskCoord.longitude
            )
        } else null
    }
    
    companion object {
        const val ACTION_START_TRACKING = "start_tracking"
        const val ACTION_STOP_TRACKING = "stop_tracking"
        const val ACTION_UPDATE_GEOFENCES = "update_geofences"
        
        private const val PROXIMITY_THRESHOLD_METERS = 100.0
        private const val GEOFENCE_RADIUS_METERS = 50.0f
    }
}