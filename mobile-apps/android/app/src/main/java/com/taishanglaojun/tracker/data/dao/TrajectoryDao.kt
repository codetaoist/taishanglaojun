package com.taishanglaojun.tracker.data.dao

import androidx.room.*
import com.taishanglaojun.tracker.data.model.Trajectory
import kotlinx.coroutines.flow.Flow

@Dao
interface TrajectoryDao {
    
    @Query("SELECT * FROM trajectories ORDER BY startTime DESC")
    fun getAllTrajectories(): Flow<List<Trajectory>>
    
    @Query("SELECT * FROM trajectories ORDER BY startTime DESC LIMIT :limit OFFSET :offset")
    suspend fun getTrajectoriesPaged(limit: Int, offset: Int): List<Trajectory>
    
    @Query("SELECT * FROM trajectories WHERE id = :trajectoryId")
    suspend fun getTrajectoryById(trajectoryId: String): Trajectory?
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertTrajectory(trajectory: Trajectory)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertTrajectories(trajectories: List<Trajectory>)
    
    @Update
    suspend fun updateTrajectory(trajectory: Trajectory)
    
    @Query("DELETE FROM trajectories WHERE id = :trajectoryId")
    suspend fun deleteTrajectory(trajectoryId: String)
    
    @Query("SELECT COUNT(*) FROM trajectories")
    suspend fun getTrajectoryCount(): Int
    
    @Query("SELECT * FROM trajectories WHERE name LIKE '%' || :query || '%' ORDER BY startTime DESC")
    suspend fun searchTrajectories(query: String): List<Trajectory>
    
    @Query("SELECT * FROM trajectories WHERE startTime BETWEEN :startTime AND :endTime ORDER BY startTime DESC")
    suspend fun getTrajectoriesByTimeRange(startTime: Long, endTime: Long): List<Trajectory>
    
    @Query("DELETE FROM trajectories WHERE startTime < :beforeTimestamp")
    suspend fun deleteOldTrajectories(beforeTimestamp: Long)
    
    @Query("SELECT SUM(distance) FROM trajectories")
    suspend fun getTotalDistance(): Double?
    
    @Query("SELECT SUM(duration) FROM trajectories")
    suspend fun getTotalDuration(): Long?
    
    @Query("SELECT * FROM trajectories WHERE endTime IS NULL ORDER BY startTime DESC LIMIT 1")
    suspend fun getActiveTrajectory(): Trajectory?
    
    @Query("UPDATE trajectories SET endTime = :endTime, distance = :distance, duration = :duration WHERE id = :trajectoryId")
    suspend fun finishTrajectory(trajectoryId: String, endTime: Long, distance: Double, duration: Long)
}