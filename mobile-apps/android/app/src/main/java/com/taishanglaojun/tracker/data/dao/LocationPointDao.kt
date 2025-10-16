package com.taishanglaojun.tracker.data.dao

import androidx.room.*
import com.taishanglaojun.tracker.data.model.LocationPoint
import kotlinx.coroutines.flow.Flow

@Dao
interface LocationPointDao {
    
    @Query("SELECT * FROM location_points WHERE trajectoryId = :trajectoryId ORDER BY timestamp ASC")
    fun getLocationPointsByTrajectory(trajectoryId: String): Flow<List<LocationPoint>>
    
    @Query("SELECT * FROM location_points WHERE trajectoryId = :trajectoryId ORDER BY timestamp ASC LIMIT :limit OFFSET :offset")
    suspend fun getLocationPointsPaged(trajectoryId: String, limit: Int, offset: Int): List<LocationPoint>
    
    @Query("SELECT * FROM location_points WHERE id = :pointId")
    suspend fun getLocationPointById(pointId: String): LocationPoint?
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertLocationPoint(locationPoint: LocationPoint)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertLocationPoints(locationPoints: List<LocationPoint>)
    
    @Update
    suspend fun updateLocationPoint(locationPoint: LocationPoint)
    
    @Query("DELETE FROM location_points WHERE id = :pointId")
    suspend fun deleteLocationPoint(pointId: String)
    
    @Query("DELETE FROM location_points WHERE trajectoryId = :trajectoryId")
    suspend fun deleteLocationPointsByTrajectory(trajectoryId: String)
    
    @Query("SELECT COUNT(*) FROM location_points WHERE trajectoryId = :trajectoryId")
    suspend fun getLocationPointCount(trajectoryId: String): Int
    
    @Query("SELECT * FROM location_points WHERE trajectoryId = :trajectoryId ORDER BY timestamp ASC LIMIT 1")
    suspend fun getFirstLocationPoint(trajectoryId: String): LocationPoint?
    
    @Query("SELECT * FROM location_points WHERE trajectoryId = :trajectoryId ORDER BY timestamp DESC LIMIT 1")
    suspend fun getLastLocationPoint(trajectoryId: String): LocationPoint?
    
    @Query("DELETE FROM location_points WHERE timestamp < :beforeTimestamp")
    suspend fun deleteOldLocationPoints(beforeTimestamp: Long)
    
    @Query("SELECT * FROM location_points WHERE timestamp BETWEEN :startTime AND :endTime ORDER BY timestamp ASC")
    suspend fun getLocationPointsByTimeRange(startTime: Long, endTime: Long): List<LocationPoint>
}