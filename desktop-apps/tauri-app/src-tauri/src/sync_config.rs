use serde::{Deserialize, Serialize};
use std::time::Duration;

/// 多设备同步配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncConfig {
    /// WebSocket服务器端口
    pub websocket_port: u16,
    /// 心跳间隔（秒）
    pub heartbeat_interval: u64,
    /// 连接超时时间（秒）
    pub connection_timeout: u64,
    /// 重连最大尝试次数
    pub max_reconnect_attempts: u32,
    /// 重连间隔（秒）
    pub reconnect_interval: u64,
    /// 离线操作最大重试次数
    pub max_operation_retries: u32,
    /// 数据缓存过期时间（天）
    pub cache_expiry_days: i64,
    /// 同步批次大小
    pub sync_batch_size: usize,
    /// 是否启用实时同步
    pub enable_realtime_sync: bool,
    /// 是否启用离线模式
    pub enable_offline_mode: bool,
    /// 冲突解决策略
    pub default_conflict_resolution: ConflictResolutionStrategy,
}

/// 冲突解决策略
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ConflictResolutionStrategy {
    /// 使用最新时间戳的数据
    UseLatest,
    /// 使用本地数据
    UseLocal,
    /// 使用远程数据
    UseRemote,
    /// 手动解决
    Manual,
    /// 合并数据（如果可能）
    Merge,
}

impl Default for SyncConfig {
    fn default() -> Self {
        Self {
            websocket_port: 8080,
            heartbeat_interval: 30,
            connection_timeout: 60,
            max_reconnect_attempts: 5,
            reconnect_interval: 5,
            max_operation_retries: 3,
            cache_expiry_days: 30,
            sync_batch_size: 100,
            enable_realtime_sync: true,
            enable_offline_mode: true,
            default_conflict_resolution: ConflictResolutionStrategy::UseLatest,
        }
    }
}

impl SyncConfig {
    /// 创建新的同步配置
    pub fn new() -> Self {
        Self::default()
    }

    /// 从文件加载配置
    pub fn load_from_file(path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let content = std::fs::read_to_string(path)?;
        let config: SyncConfig = serde_json::from_str(&content)?;
        Ok(config)
    }

    /// 保存配置到文件
    pub fn save_to_file(&self, path: &str) -> Result<(), Box<dyn std::error::Error>> {
        let content = serde_json::to_string_pretty(self)?;
        std::fs::write(path, content)?;
        Ok(())
    }

    /// 验证配置有效性
    pub fn validate(&self) -> Result<(), String> {
        if self.websocket_port == 0 {
            return Err("WebSocket端口不能为0".to_string());
        }

        if self.heartbeat_interval == 0 {
            return Err("心跳间隔不能为0".to_string());
        }

        if self.connection_timeout == 0 {
            return Err("连接超时时间不能为0".to_string());
        }

        if self.sync_batch_size == 0 {
            return Err("同步批次大小不能为0".to_string());
        }

        Ok(())
    }

    /// 获取心跳间隔Duration
    pub fn heartbeat_duration(&self) -> Duration {
        Duration::from_secs(self.heartbeat_interval)
    }

    /// 获取连接超时Duration
    pub fn connection_timeout_duration(&self) -> Duration {
        Duration::from_secs(self.connection_timeout)
    }

    /// 获取重连间隔Duration
    pub fn reconnect_duration(&self) -> Duration {
        Duration::from_secs(self.reconnect_interval)
    }

    /// 更新WebSocket端口
    pub fn set_websocket_port(&mut self, port: u16) {
        self.websocket_port = port;
    }

    /// 更新心跳间隔
    pub fn set_heartbeat_interval(&mut self, interval: u64) {
        self.heartbeat_interval = interval;
    }

    /// 更新连接超时时间
    pub fn set_connection_timeout(&mut self, timeout: u64) {
        self.connection_timeout = timeout;
    }

    /// 更新最大重连尝试次数
    pub fn set_max_reconnect_attempts(&mut self, attempts: u32) {
        self.max_reconnect_attempts = attempts;
    }

    /// 更新重连间隔
    pub fn set_reconnect_interval(&mut self, interval: u64) {
        self.reconnect_interval = interval;
    }

    /// 更新离线操作最大重试次数
    pub fn set_max_operation_retries(&mut self, retries: u32) {
        self.max_operation_retries = retries;
    }

    /// 更新数据缓存过期时间
    pub fn set_cache_expiry_days(&mut self, days: i64) {
        self.cache_expiry_days = days;
    }

    /// 更新同步批次大小
    pub fn set_sync_batch_size(&mut self, size: usize) {
        self.sync_batch_size = size;
    }

    /// 启用/禁用实时同步
    pub fn set_realtime_sync(&mut self, enabled: bool) {
        self.enable_realtime_sync = enabled;
    }

    /// 启用/禁用离线模式
    pub fn set_offline_mode(&mut self, enabled: bool) {
        self.enable_offline_mode = enabled;
    }

    /// 更新默认冲突解决策略
    pub fn set_conflict_resolution(&mut self, strategy: ConflictResolutionStrategy) {
        self.default_conflict_resolution = strategy;
    }
}

/// 设备特定配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeviceConfig {
    /// 设备ID
    pub device_id: String,
    /// 设备名称
    pub device_name: String,
    /// 设备类型
    pub device_type: String,
    /// 是否启用自动同步
    pub auto_sync_enabled: bool,
    /// 同步频率（分钟）
    pub sync_frequency_minutes: u64,
    /// 最大存储空间（MB）
    pub max_storage_mb: u64,
    /// 是否启用数据压缩
    pub enable_compression: bool,
    /// 网络质量阈值
    pub network_quality_threshold: f32,
}

impl Default for DeviceConfig {
    fn default() -> Self {
        Self {
            device_id: uuid::Uuid::new_v4().to_string(),
            device_name: "Unknown Device".to_string(),
            device_type: "desktop".to_string(),
            auto_sync_enabled: true,
            sync_frequency_minutes: 5,
            max_storage_mb: 1024,
            enable_compression: true,
            network_quality_threshold: 0.7,
        }
    }
}

impl DeviceConfig {
    /// 创建新的设备配置
    pub fn new(device_name: String, device_type: String) -> Self {
        Self {
            device_name,
            device_type,
            ..Default::default()
        }
    }

    /// 获取同步频率Duration
    pub fn sync_frequency_duration(&self) -> Duration {
        Duration::from_secs(self.sync_frequency_minutes * 60)
    }

    /// 检查网络质量是否满足同步要求
    pub fn is_network_quality_sufficient(&self, quality: f32) -> bool {
        quality >= self.network_quality_threshold
    }
}

// Tauri命令函数
#[tauri::command]
pub async fn get_sync_config() -> Result<SyncConfig, String> {
    Ok(SyncConfig::default())
}

#[tauri::command]
pub async fn update_sync_config(config: SyncConfig) -> Result<(), String> {
    config.validate()
        .map_err(|e| format!("配置验证失败: {}", e))?;
    
    // 这里可以保存配置到文件或数据库
    // config.save_to_file("sync_config.json")?;
    
    Ok(())
}

#[tauri::command]
pub async fn get_device_config() -> Result<DeviceConfig, String> {
    Ok(DeviceConfig::default())
}

#[tauri::command]
pub async fn update_device_config(config: DeviceConfig) -> Result<(), String> {
    // 这里可以保存设备配置
    Ok(())
}

#[tauri::command]
pub async fn reset_sync_config() -> Result<SyncConfig, String> {
    Ok(SyncConfig::default())
}