-- 创建轨迹表
CREATE TABLE IF NOT EXISTS trajectories (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_time BIGINT NOT NULL,
    end_time BIGINT,
    distance DOUBLE DEFAULT 0,
    duration BIGINT DEFAULT 0,
    max_speed DOUBLE DEFAULT 0,
    avg_speed DOUBLE DEFAULT 0,
    point_count INT DEFAULT 0,
    min_latitude DOUBLE DEFAULT 0,
    max_latitude DOUBLE DEFAULT 0,
    min_longitude DOUBLE DEFAULT 0,
    max_longitude DOUBLE DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_trajectories_user_id (user_id),
    INDEX idx_trajectories_start_time (start_time),
    INDEX idx_trajectories_end_time (end_time),
    INDEX idx_trajectories_is_active (is_active),
    INDEX idx_trajectories_deleted_at (deleted_at)
);

-- 创建位置点表
CREATE TABLE IF NOT EXISTS location_points (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    trajectory_id VARCHAR(36) NOT NULL,
    latitude DOUBLE NOT NULL,
    longitude DOUBLE NOT NULL,
    altitude DOUBLE,
    accuracy DOUBLE,
    speed DOUBLE,
    bearing DOUBLE,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_location_points_user_id (user_id),
    INDEX idx_location_points_trajectory_id (trajectory_id),
    INDEX idx_location_points_timestamp (timestamp),
    INDEX idx_location_points_lat_lng (latitude, longitude),
    INDEX idx_location_points_deleted_at (deleted_at),
    
    FOREIGN KEY (trajectory_id) REFERENCES trajectories(id) ON DELETE CASCADE
);