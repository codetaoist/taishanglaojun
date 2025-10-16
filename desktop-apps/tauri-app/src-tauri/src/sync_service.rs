use serde::{Deserialize, Serialize};
use sqlx::{Pool, Sqlite};
use std::collections::HashMap;
use tokio::sync::RwLock;
use uuid::Uuid;
use chrono::{DateTime, Utc};
use anyhow::Result;

/// 设备类型枚举
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum DeviceType {
    Desktop,
    Mobile,
    Watch,
    Tablet,
}

/// 设备信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeviceInfo {
    pub device_id: String,
    pub device_type: DeviceType,
    pub device_name: String,
    pub platform: String,
    pub app_version: String,
    pub last_sync: DateTime<Utc>,
    pub is_online: bool,
}

/// 同步数据类型
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SyncDataType {
    UserProfile,
    ChatMessage,
    ChatSession,
    Friend,
    Project,
    File,
    Settings,
}

/// 同步操作类型
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SyncOperation {
    Create,
    Update,
    Delete,
    Read,
}

/// 同步记录
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncRecord {
    pub id: String,
    pub user_id: String,
    pub device_id: String,
    pub data_type: SyncDataType,
    pub operation: SyncOperation,
    pub data_id: String,
    pub data_hash: String,
    pub timestamp: DateTime<Utc>,
    pub version: i64,
    pub conflict_resolution: Option<String>,
}

/// 冲突解决策略
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ConflictResolution {
    LastWriteWins,
    FirstWriteWins,
    MergeChanges,
    UserChoice,
    DevicePriority(DeviceType),
}

/// 多设备同步服务
pub struct MultiDeviceSyncService {
    main_db: Pool<Sqlite>,
    chat_db: Pool<Sqlite>,
    storage_db: Pool<Sqlite>,
    devices: RwLock<HashMap<String, DeviceInfo>>,
    sync_queue: RwLock<Vec<SyncRecord>>,
    conflict_resolver: ConflictResolution,
}

impl MultiDeviceSyncService {
    pub fn new(
        main_db: Pool<Sqlite>,
        chat_db: Pool<Sqlite>,
        storage_db: Pool<Sqlite>,
    ) -> Self {
        Self {
            main_db,
            chat_db,
            storage_db,
            devices: RwLock::new(HashMap::new()),
            sync_queue: RwLock::new(Vec::new()),
            conflict_resolver: ConflictResolution::LastWriteWins,
        }
    }

    /// 注册设备
    pub async fn register_device(&self, device_info: DeviceInfo) -> Result<()> {
        let mut devices = self.devices.write().await;
        devices.insert(device_info.device_id.clone(), device_info.clone());

        // 在数据库中保存设备信息
        sqlx::query!(
            r#"
            INSERT OR REPLACE INTO devices 
            (device_id, device_type, device_name, platform, app_version, last_sync, is_online)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            "#,
            device_info.device_id,
            serde_json::to_string(&device_info.device_type)?,
            device_info.device_name,
            device_info.platform,
            device_info.app_version,
            device_info.last_sync,
            device_info.is_online
        )
        .execute(&self.main_db)
        .await?;

        Ok(())
    }

    /// 获取用户的所有设备
    pub async fn get_user_devices(&self, user_id: &str) -> Result<Vec<DeviceInfo>> {
        let rows = sqlx::query!(
            "SELECT * FROM devices WHERE user_id = ? ORDER BY last_sync DESC",
            user_id
        )
        .fetch_all(&self.main_db)
        .await?;

        let mut devices = Vec::new();
        for row in rows {
            let device_type: DeviceType = serde_json::from_str(&row.device_type)?;
            devices.push(DeviceInfo {
                device_id: row.device_id,
                device_type,
                device_name: row.device_name,
                platform: row.platform,
                app_version: row.app_version,
                last_sync: row.last_sync,
                is_online: row.is_online,
            });
        }

        Ok(devices)
    }

    /// 创建同步记录
    pub async fn create_sync_record(
        &self,
        user_id: &str,
        device_id: &str,
        data_type: SyncDataType,
        operation: SyncOperation,
        data_id: &str,
        data: &str,
    ) -> Result<SyncRecord> {
        let record = SyncRecord {
            id: Uuid::new_v4().to_string(),
            user_id: user_id.to_string(),
            device_id: device_id.to_string(),
            data_type,
            operation,
            data_id: data_id.to_string(),
            data_hash: self.calculate_hash(data),
            timestamp: Utc::now(),
            version: self.get_next_version(user_id, data_id).await?,
            conflict_resolution: None,
        };

        // 保存到同步队列
        let mut queue = self.sync_queue.write().await;
        queue.push(record.clone());

        // 保存到数据库
        sqlx::query!(
            r#"
            INSERT INTO sync_records 
            (id, user_id, device_id, data_type, operation, data_id, data_hash, timestamp, version)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            "#,
            record.id,
            record.user_id,
            record.device_id,
            serde_json::to_string(&record.data_type)?,
            serde_json::to_string(&record.operation)?,
            record.data_id,
            record.data_hash,
            record.timestamp,
            record.version
        )
        .execute(&self.main_db)
        .await?;

        Ok(record)
    }

    /// 获取设备的增量同步数据
    pub async fn get_incremental_sync(
        &self,
        user_id: &str,
        device_id: &str,
        last_sync_time: DateTime<Utc>,
    ) -> Result<Vec<SyncRecord>> {
        let rows = sqlx::query!(
            r#"
            SELECT * FROM sync_records 
            WHERE user_id = ? AND device_id != ? AND timestamp > ?
            ORDER BY timestamp ASC
            "#,
            user_id,
            device_id,
            last_sync_time
        )
        .fetch_all(&self.main_db)
        .await?;

        let mut records = Vec::new();
        for row in rows {
            let data_type: SyncDataType = serde_json::from_str(&row.data_type)?;
            let operation: SyncOperation = serde_json::from_str(&row.operation)?;
            
            records.push(SyncRecord {
                id: row.id,
                user_id: row.user_id,
                device_id: row.device_id,
                data_type,
                operation,
                data_id: row.data_id,
                data_hash: row.data_hash,
                timestamp: row.timestamp,
                version: row.version,
                conflict_resolution: row.conflict_resolution,
            });
        }

        Ok(records)
    }

    /// 处理数据冲突
    pub async fn resolve_conflict(
        &self,
        local_record: &SyncRecord,
        remote_record: &SyncRecord,
    ) -> Result<SyncRecord> {
        match self.conflict_resolver {
            ConflictResolution::LastWriteWins => {
                if remote_record.timestamp > local_record.timestamp {
                    Ok(remote_record.clone())
                } else {
                    Ok(local_record.clone())
                }
            }
            ConflictResolution::FirstWriteWins => {
                if local_record.timestamp < remote_record.timestamp {
                    Ok(local_record.clone())
                } else {
                    Ok(remote_record.clone())
                }
            }
            ConflictResolution::DevicePriority(priority_device) => {
                let local_device = self.get_device_info(&local_record.device_id).await?;
                let remote_device = self.get_device_info(&remote_record.device_id).await?;

                if local_device.device_type == priority_device {
                    Ok(local_record.clone())
                } else if remote_device.device_type == priority_device {
                    Ok(remote_record.clone())
                } else {
                    // 回退到时间戳比较
                    if remote_record.timestamp > local_record.timestamp {
                        Ok(remote_record.clone())
                    } else {
                        Ok(local_record.clone())
                    }
                }
            }
            _ => {
                // 默认使用最后写入获胜
                if remote_record.timestamp > local_record.timestamp {
                    Ok(remote_record.clone())
                } else {
                    Ok(local_record.clone())
                }
            }
        }
    }

    /// 同步聊天消息
    pub async fn sync_chat_messages(
        &self,
        user_id: &str,
        device_id: &str,
        last_sync: DateTime<Utc>,
    ) -> Result<Vec<ChatMessage>> {
        let sync_records = self.get_incremental_sync(user_id, device_id, last_sync).await?;
        let mut messages = Vec::new();

        for record in sync_records {
            if matches!(record.data_type, SyncDataType::ChatMessage) {
                // 从chat_db获取消息详情
                let message = self.get_chat_message(&record.data_id).await?;
                if let Some(msg) = message {
                    messages.push(msg);
                }
            }
        }

        Ok(messages)
    }

    /// 同步好友数据
    pub async fn sync_friends(
        &self,
        user_id: &str,
        device_id: &str,
        last_sync: DateTime<Utc>,
    ) -> Result<Vec<Friend>> {
        let sync_records = self.get_incremental_sync(user_id, device_id, last_sync).await?;
        let mut friends = Vec::new();

        for record in sync_records {
            if matches!(record.data_type, SyncDataType::Friend) {
                let friend = self.get_friend(&record.data_id).await?;
                if let Some(f) = friend {
                    friends.push(f);
                }
            }
        }

        Ok(friends)
    }

    // 辅助方法
    async fn get_device_info(&self, device_id: &str) -> Result<DeviceInfo> {
        let devices = self.devices.read().await;
        devices.get(device_id)
            .cloned()
            .ok_or_else(|| anyhow::anyhow!("Device not found"))
    }

    async fn get_next_version(&self, user_id: &str, data_id: &str) -> Result<i64> {
        let row = sqlx::query!(
            "SELECT MAX(version) as max_version FROM sync_records WHERE user_id = ? AND data_id = ?",
            user_id,
            data_id
        )
        .fetch_one(&self.main_db)
        .await?;

        Ok(row.max_version.unwrap_or(0) + 1)
    }

    fn calculate_hash(&self, data: &str) -> String {
        use sha2::{Sha256, Digest};
        let mut hasher = Sha256::new();
        hasher.update(data.as_bytes());
        format!("{:x}", hasher.finalize())
    }

    async fn get_chat_message(&self, message_id: &str) -> Result<Option<ChatMessage>> {
        // 从chat_db获取消息
        let row = sqlx::query!(
            "SELECT * FROM chat_messages WHERE id = ?",
            message_id
        )
        .fetch_optional(&self.chat_db)
        .await?;

        if let Some(row) = row {
            Ok(Some(ChatMessage {
                id: row.id,
                session_id: row.session_id,
                sender_id: row.sender_id,
                content: row.content,
                message_type: row.message_type,
                timestamp: row.timestamp,
                is_read: row.is_read,
            }))
        } else {
            Ok(None)
        }
    }

    async fn get_friend(&self, friend_id: &str) -> Result<Option<Friend>> {
        // 从main_db获取好友信息
        let row = sqlx::query!(
            "SELECT * FROM friends WHERE id = ?",
            friend_id
        )
        .fetch_optional(&self.main_db)
        .await?;

        if let Some(row) = row {
            Ok(Some(Friend {
                id: row.id,
                user_id: row.user_id,
                friend_user_id: row.friend_user_id,
                status: row.status,
                created_at: row.created_at,
                updated_at: row.updated_at,
            }))
        } else {
            Ok(None)
        }
    }
}

// 数据结构定义
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatMessage {
    pub id: String,
    pub session_id: String,
    pub sender_id: String,
    pub content: String,
    pub message_type: String,
    pub timestamp: DateTime<Utc>,
    pub is_read: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Friend {
    pub id: String,
    pub user_id: String,
    pub friend_user_id: String,
    pub status: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl MultiDeviceSyncService {
    /// 增量同步
    pub async fn sync_incremental(
        &self,
        user_id: &str,
        device_id: &str,
        last_sync: DateTime<Utc>,
    ) -> Result<Vec<SyncRecord>> {
        self.get_incremental_sync(user_id, device_id, last_sync).await
    }

    /// 同步聊天消息（批量）
    pub async fn sync_chat_messages(
        &self,
        user_id: &str,
        device_id: &str,
        messages: Vec<ChatMessage>,
    ) -> Result<()> {
        for message in messages {
            self.create_sync_record(
                user_id,
                device_id,
                SyncDataType::ChatMessage,
                SyncOperation::Create,
                &message.id,
                &serde_json::to_string(&message)?,
            ).await?;
        }
        Ok(())
    }

    /// 同步好友数据（批量）
    pub async fn sync_friend_data(
        &self,
        user_id: &str,
        device_id: &str,
        friends: Vec<Friend>,
    ) -> Result<()> {
        for friend in friends {
            self.create_sync_record(
                user_id,
                device_id,
                SyncDataType::Friend,
                SyncOperation::Create,
                &friend.id,
                &serde_json::to_string(&friend)?,
            ).await?;
        }
        Ok(())
    }
}

// Tauri命令函数
#[tauri::command]
pub async fn register_device(
    device_info: serde_json::Value,
    sync_service: tauri::State<'_, std::sync::Arc<MultiDeviceSyncService>>,
) -> Result<String, String> {
    let device_info: DeviceInfo = serde_json::from_value(device_info)
        .map_err(|e| format!("Invalid device info: {}", e))?;
    
    sync_service.register_device(device_info).await
        .map_err(|e| format!("Failed to register device: {}", e))?;
    
    Ok("Device registered successfully".to_string())
}

#[tauri::command]
pub async fn get_user_devices(
    user_id: String,
    sync_service: tauri::State<'_, std::sync::Arc<MultiDeviceSyncService>>,
) -> Result<Vec<DeviceInfo>, String> {
    sync_service.get_user_devices(&user_id).await
        .map_err(|e| format!("Failed to get user devices: {}", e))
}

#[tauri::command]
pub async fn sync_incremental(
    user_id: String,
    device_id: String,
    last_sync_time: String,
    sync_service: tauri::State<'_, std::sync::Arc<MultiDeviceSyncService>>,
) -> Result<Vec<SyncRecord>, String> {
    let last_sync = chrono::DateTime::parse_from_rfc3339(&last_sync_time)
        .map_err(|e| format!("Invalid timestamp: {}", e))?;
    
    sync_service.sync_incremental(&user_id, &device_id, last_sync.into()).await
        .map_err(|e| format!("Failed to sync incremental: {}", e))
}

#[tauri::command]
pub async fn sync_chat_messages(
    user_id: String,
    device_id: String,
    messages: Vec<serde_json::Value>,
    sync_service: tauri::State<'_, std::sync::Arc<MultiDeviceSyncService>>,
) -> Result<(), String> {
    let chat_messages: Vec<ChatMessage> = messages.into_iter()
        .map(|v| serde_json::from_value(v))
        .collect::<Result<Vec<_>, _>>()
        .map_err(|e| format!("Invalid chat messages: {}", e))?;
    
    sync_service.sync_chat_messages(&user_id, &device_id, chat_messages).await
        .map_err(|e| format!("Failed to sync chat messages: {}", e))
}

#[tauri::command]
pub async fn sync_friend_data(
    user_id: String,
    device_id: String,
    friends: Vec<serde_json::Value>,
    sync_service: tauri::State<'_, std::sync::Arc<MultiDeviceSyncService>>,
) -> Result<(), String> {
    let friend_list: Vec<Friend> = friends.into_iter()
        .map(|v| serde_json::from_value(v))
        .collect::<Result<Vec<_>, _>>()
        .map_err(|e| format!("Invalid friend data: {}", e))?;
    
    sync_service.sync_friend_data(&user_id, &device_id, friend_list).await
        .map_err(|e| format!("Failed to sync friend data: {}", e))
}