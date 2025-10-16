use serde::{Deserialize, Serialize};
use sqlx::{Pool, Sqlite};
use std::collections::{HashMap, VecDeque};
use tokio::sync::RwLock;
use chrono::{DateTime, Utc};
use anyhow::Result;
use uuid::Uuid;

use crate::sync_service::{SyncRecord, SyncDataType, SyncOperation};
use crate::realtime_sync::RealtimeSyncMessage;

/// 离线操作队列项
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OfflineOperation {
    pub id: String,
    pub user_id: String,
    pub device_id: String,
    pub operation_type: SyncOperation,
    pub data_type: SyncDataType,
    pub data_id: String,
    pub data_payload: String,
    pub created_at: DateTime<Utc>,
    pub retry_count: i32,
    pub max_retries: i32,
    pub priority: OperationPriority,
}

/// 操作优先级
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, PartialOrd, Ord)]
pub enum OperationPriority {
    Low = 1,
    Normal = 2,
    High = 3,
    Critical = 4,
}

/// 离线数据缓存项
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OfflineCache {
    pub key: String,
    pub data: String,
    pub data_type: SyncDataType,
    pub created_at: DateTime<Utc>,
    pub expires_at: Option<DateTime<Utc>>,
    pub size_bytes: usize,
}

/// 同步冲突项
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncConflict {
    pub id: String,
    pub user_id: String,
    pub data_id: String,
    pub local_version: SyncRecord,
    pub remote_version: SyncRecord,
    pub conflict_type: ConflictType,
    pub created_at: DateTime<Utc>,
    pub resolved: bool,
}

/// 冲突类型
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ConflictType {
    DataModified,
    DataDeleted,
    VersionMismatch,
    SchemaChange,
}

/// 离线数据管理器
pub struct OfflineDataManager {
    main_db: Pool<Sqlite>,
    operation_queue: RwLock<VecDeque<OfflineOperation>>,
    cache_storage: RwLock<HashMap<String, OfflineCache>>,
    conflict_queue: RwLock<Vec<SyncConflict>>,
    max_cache_size: usize,
    max_queue_size: usize,
}

impl OfflineDataManager {
    pub fn new(main_db: Pool<Sqlite>) -> Self {
        Self {
            main_db,
            operation_queue: RwLock::new(VecDeque::new()),
            cache_storage: RwLock::new(HashMap::new()),
            conflict_queue: RwLock::new(Vec::new()),
            max_cache_size: 100 * 1024 * 1024, // 100MB
            max_queue_size: 10000,
        }
    }

    /// 初始化离线管理器
    pub async fn initialize(&self) -> Result<()> {
        // 创建离线操作表
        sqlx::query!(
            r#"
            CREATE TABLE IF NOT EXISTS offline_operations (
                id TEXT PRIMARY KEY,
                user_id TEXT NOT NULL,
                device_id TEXT NOT NULL,
                operation_type TEXT NOT NULL,
                data_type TEXT NOT NULL,
                data_id TEXT NOT NULL,
                data_payload TEXT NOT NULL,
                created_at DATETIME NOT NULL,
                retry_count INTEGER DEFAULT 0,
                max_retries INTEGER DEFAULT 3,
                priority INTEGER DEFAULT 2
            )
            "#
        )
        .execute(&self.main_db)
        .await?;

        // 创建离线缓存表
        sqlx::query!(
            r#"
            CREATE TABLE IF NOT EXISTS offline_cache (
                key TEXT PRIMARY KEY,
                data TEXT NOT NULL,
                data_type TEXT NOT NULL,
                created_at DATETIME NOT NULL,
                expires_at DATETIME,
                size_bytes INTEGER NOT NULL
            )
            "#
        )
        .execute(&self.main_db)
        .await?;

        // 创建同步冲突表
        sqlx::query!(
            r#"
            CREATE TABLE IF NOT EXISTS sync_conflicts (
                id TEXT PRIMARY KEY,
                user_id TEXT NOT NULL,
                data_id TEXT NOT NULL,
                local_version TEXT NOT NULL,
                remote_version TEXT NOT NULL,
                conflict_type TEXT NOT NULL,
                created_at DATETIME NOT NULL,
                resolved BOOLEAN DEFAULT FALSE
            )
            "#
        )
        .execute(&self.main_db)
        .await?;

        // 加载未完成的操作
        self.load_pending_operations().await?;
        
        // 加载缓存数据
        self.load_cache_data().await?;

        Ok(())
    }

    /// 添加离线操作
    pub async fn add_offline_operation(
        &self,
        user_id: &str,
        device_id: &str,
        operation_type: SyncOperation,
        data_type: SyncDataType,
        data_id: &str,
        data_payload: &str,
        priority: OperationPriority,
    ) -> Result<String> {
        let operation = OfflineOperation {
            id: Uuid::new_v4().to_string(),
            user_id: user_id.to_string(),
            device_id: device_id.to_string(),
            operation_type: operation_type.clone(),
            data_type: data_type.clone(),
            data_id: data_id.to_string(),
            data_payload: data_payload.to_string(),
            created_at: Utc::now(),
            retry_count: 0,
            max_retries: 3,
            priority,
        };

        // 保存到数据库
        sqlx::query!(
            r#"
            INSERT INTO offline_operations 
            (id, user_id, device_id, operation_type, data_type, data_id, data_payload, created_at, priority)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            "#,
            operation.id,
            operation.user_id,
            operation.device_id,
            serde_json::to_string(&operation.operation_type)?,
            serde_json::to_string(&operation.data_type)?,
            operation.data_id,
            operation.data_payload,
            operation.created_at,
            operation.priority as i32
        )
        .execute(&self.main_db)
        .await?;

        // 添加到内存队列
        let mut queue = self.operation_queue.write().await;
        
        // 按优先级插入
        let insert_pos = queue.iter().position(|op| op.priority < operation.priority)
            .unwrap_or(queue.len());
        queue.insert(insert_pos, operation.clone());

        // 限制队列大小
        if queue.len() > self.max_queue_size {
            queue.pop_back();
        }

        Ok(operation.id)
    }

    /// 处理离线操作队列
    pub async fn process_offline_queue(&self) -> Result<Vec<String>> {
        let mut processed_ids = Vec::new();
        let mut queue = self.operation_queue.write().await;

        while let Some(operation) = queue.pop_front() {
            match self.execute_operation(&operation).await {
                Ok(_) => {
                    // 操作成功，从数据库删除
                    sqlx::query!(
                        "DELETE FROM offline_operations WHERE id = ?",
                        operation.id
                    )
                    .execute(&self.main_db)
                    .await?;
                    
                    processed_ids.push(operation.id);
                }
                Err(e) => {
                    eprintln!("执行离线操作失败: {}, 错误: {}", operation.id, e);
                    
                    // 增加重试次数
                    let mut updated_op = operation.clone();
                    updated_op.retry_count += 1;

                    if updated_op.retry_count < updated_op.max_retries {
                        // 重新加入队列（降低优先级）
                        let insert_pos = queue.len();
                        queue.insert(insert_pos, updated_op.clone());

                        // 更新数据库
                        sqlx::query!(
                            "UPDATE offline_operations SET retry_count = ? WHERE id = ?",
                            updated_op.retry_count,
                            updated_op.id
                        )
                        .execute(&self.main_db)
                        .await?;
                    } else {
                        // 超过最大重试次数，记录错误并删除
                        eprintln!("操作 {} 超过最大重试次数，已放弃", operation.id);
                        sqlx::query!(
                            "DELETE FROM offline_operations WHERE id = ?",
                            operation.id
                        )
                        .execute(&self.main_db)
                        .await?;
                    }
                }
            }
        }

        Ok(processed_ids)
    }

    /// 缓存数据
    pub async fn cache_data(
        &self,
        key: &str,
        data: &str,
        data_type: SyncDataType,
        ttl_seconds: Option<i64>,
    ) -> Result<()> {
        let expires_at = ttl_seconds.map(|ttl| Utc::now() + chrono::Duration::seconds(ttl));
        let size_bytes = data.len();

        let cache_item = OfflineCache {
            key: key.to_string(),
            data: data.to_string(),
            data_type: data_type.clone(),
            created_at: Utc::now(),
            expires_at,
            size_bytes,
        };

        // 检查缓存大小限制
        self.ensure_cache_size_limit(size_bytes).await?;

        // 保存到数据库
        sqlx::query!(
            r#"
            INSERT OR REPLACE INTO offline_cache 
            (key, data, data_type, created_at, expires_at, size_bytes)
            VALUES (?, ?, ?, ?, ?, ?)
            "#,
            cache_item.key,
            cache_item.data,
            serde_json::to_string(&cache_item.data_type)?,
            cache_item.created_at,
            cache_item.expires_at,
            cache_item.size_bytes as i64
        )
        .execute(&self.main_db)
        .await?;

        // 添加到内存缓存
        let mut cache = self.cache_storage.write().await;
        cache.insert(key.to_string(), cache_item);

        Ok(())
    }

    /// 获取缓存数据
    pub async fn get_cached_data(&self, key: &str) -> Result<Option<String>> {
        let cache = self.cache_storage.read().await;
        
        if let Some(cache_item) = cache.get(key) {
            // 检查是否过期
            if let Some(expires_at) = cache_item.expires_at {
                if Utc::now() > expires_at {
                    drop(cache);
                    self.remove_cached_data(key).await?;
                    return Ok(None);
                }
            }
            
            Ok(Some(cache_item.data.clone()))
        } else {
            Ok(None)
        }
    }

    /// 删除缓存数据
    pub async fn remove_cached_data(&self, key: &str) -> Result<()> {
        // 从数据库删除
        sqlx::query!("DELETE FROM offline_cache WHERE key = ?", key)
            .execute(&self.main_db)
            .await?;

        // 从内存删除
        let mut cache = self.cache_storage.write().await;
        cache.remove(key);

        Ok(())
    }

    /// 记录同步冲突
    pub async fn record_conflict(
        &self,
        user_id: &str,
        data_id: &str,
        local_version: SyncRecord,
        remote_version: SyncRecord,
        conflict_type: ConflictType,
    ) -> Result<String> {
        let conflict = SyncConflict {
            id: Uuid::new_v4().to_string(),
            user_id: user_id.to_string(),
            data_id: data_id.to_string(),
            local_version: local_version.clone(),
            remote_version: remote_version.clone(),
            conflict_type: conflict_type.clone(),
            created_at: Utc::now(),
            resolved: false,
        };

        // 保存到数据库
        sqlx::query!(
            r#"
            INSERT INTO sync_conflicts 
            (id, user_id, data_id, local_version, remote_version, conflict_type, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            "#,
            conflict.id,
            conflict.user_id,
            conflict.data_id,
            serde_json::to_string(&conflict.local_version)?,
            serde_json::to_string(&conflict.remote_version)?,
            serde_json::to_string(&conflict.conflict_type)?,
            conflict.created_at
        )
        .execute(&self.main_db)
        .await?;

        // 添加到内存队列
        let mut conflicts = self.conflict_queue.write().await;
        conflicts.push(conflict.clone());

        Ok(conflict.id)
    }

    /// 获取未解决的冲突
    pub async fn get_unresolved_conflicts(&self, user_id: &str) -> Result<Vec<SyncConflict>> {
        let conflicts = self.conflict_queue.read().await;
        Ok(conflicts.iter()
            .filter(|c| c.user_id == user_id && !c.resolved)
            .cloned()
            .collect())
    }

    /// 解决冲突
    pub async fn resolve_conflict(&self, conflict_id: &str, resolution: SyncRecord) -> Result<()> {
        // 更新数据库
        sqlx::query!(
            "UPDATE sync_conflicts SET resolved = TRUE WHERE id = ?",
            conflict_id
        )
        .execute(&self.main_db)
        .await?;

        // 更新内存队列
        let mut conflicts = self.conflict_queue.write().await;
        if let Some(conflict) = conflicts.iter_mut().find(|c| c.id == conflict_id) {
            conflict.resolved = true;
        }

        Ok(())
    }

    /// 清理过期数据
    pub async fn cleanup_expired_data(&self) -> Result<()> {
        let now = Utc::now();

        // 清理过期缓存
        sqlx::query!(
            "DELETE FROM offline_cache WHERE expires_at IS NOT NULL AND expires_at < ?",
            now
        )
        .execute(&self.main_db)
        .await?;

        // 清理内存缓存
        let mut cache = self.cache_storage.write().await;
        cache.retain(|_, item| {
            item.expires_at.map_or(true, |expires| expires > now)
        });

        // 清理旧的已解决冲突（保留30天）
        let thirty_days_ago = now - chrono::Duration::days(30);
        sqlx::query!(
            "DELETE FROM sync_conflicts WHERE resolved = TRUE AND created_at < ?",
            thirty_days_ago
        )
        .execute(&self.main_db)
        .await?;

        Ok(())
    }

    /// 获取离线操作队列
    pub async fn get_offline_queue(&self, user_id: &str) -> Result<Vec<OfflineOperation>> {
        let rows = sqlx::query!(
            "SELECT * FROM offline_operations WHERE user_id = ? ORDER BY priority DESC, created_at ASC",
            user_id
        )
        .fetch_all(&self.main_db)
        .await?;

        let mut operations = Vec::new();
        for row in rows {
            let operation_type: SyncOperation = serde_json::from_str(&row.operation_type)?;
            let data_type: SyncDataType = serde_json::from_str(&row.data_type)?;
            let priority = match row.priority {
                1 => OperationPriority::Low,
                2 => OperationPriority::Normal,
                3 => OperationPriority::High,
                4 => OperationPriority::Critical,
                _ => OperationPriority::Normal,
            };

            let operation = OfflineOperation {
                id: row.id,
                user_id: row.user_id,
                device_id: row.device_id,
                operation_type,
                data_type,
                data_id: row.data_id,
                data_payload: row.data_payload,
                created_at: row.created_at,
                retry_count: row.retry_count,
                max_retries: row.max_retries,
                priority,
            };

            operations.push(operation);
        }

        Ok(operations)
    }

    // 私有辅助方法
    async fn load_pending_operations(&self) -> Result<()> {
        let rows = sqlx::query!(
            "SELECT * FROM offline_operations ORDER BY priority DESC, created_at ASC"
        )
        .fetch_all(&self.main_db)
        .await?;

        let mut queue = self.operation_queue.write().await;
        for row in rows {
            let operation_type: SyncOperation = serde_json::from_str(&row.operation_type)?;
            let data_type: SyncDataType = serde_json::from_str(&row.data_type)?;
            let priority = match row.priority {
                1 => OperationPriority::Low,
                2 => OperationPriority::Normal,
                3 => OperationPriority::High,
                4 => OperationPriority::Critical,
                _ => OperationPriority::Normal,
            };

            let operation = OfflineOperation {
                id: row.id,
                user_id: row.user_id,
                device_id: row.device_id,
                operation_type,
                data_type,
                data_id: row.data_id,
                data_payload: row.data_payload,
                created_at: row.created_at,
                retry_count: row.retry_count,
                max_retries: row.max_retries,
                priority,
            };

            queue.push_back(operation);
        }

        Ok(())
    }

    async fn load_cache_data(&self) -> Result<()> {
        let rows = sqlx::query!("SELECT * FROM offline_cache")
            .fetch_all(&self.main_db)
            .await?;

        let mut cache = self.cache_storage.write().await;
        for row in rows {
            let data_type: SyncDataType = serde_json::from_str(&row.data_type)?;
            
            let cache_item = OfflineCache {
                key: row.key.clone(),
                data: row.data,
                data_type,
                created_at: row.created_at,
                expires_at: row.expires_at,
                size_bytes: row.size_bytes as usize,
            };

            cache.insert(row.key, cache_item);
        }

        Ok(())
    }

    async fn execute_operation(&self, operation: &OfflineOperation) -> Result<()> {
        // 这里实现具体的操作执行逻辑
        // 根据operation_type和data_type执行相应的数据库操作
        match (&operation.operation_type, &operation.data_type) {
            (SyncOperation::Create, SyncDataType::ChatMessage) => {
                // 创建聊天消息
                self.create_chat_message(&operation.data_payload).await
            }
            (SyncOperation::Update, SyncDataType::UserProfile) => {
                // 更新用户资料
                self.update_user_profile(&operation.data_payload).await
            }
            (SyncOperation::Delete, SyncDataType::Friend) => {
                // 删除好友
                self.delete_friend(&operation.data_id).await
            }
            _ => {
                // 其他操作类型
                Ok(())
            }
        }
    }

    async fn ensure_cache_size_limit(&self, new_item_size: usize) -> Result<()> {
        let cache = self.cache_storage.read().await;
        let current_size: usize = cache.values().map(|item| item.size_bytes).sum();
        
        if current_size + new_item_size > self.max_cache_size {
            drop(cache);
            // 清理最旧的缓存项
            self.cleanup_old_cache_items().await?;
        }

        Ok(())
    }

    async fn cleanup_old_cache_items(&self) -> Result<()> {
        // 删除最旧的25%缓存项
        let mut cache = self.cache_storage.write().await;
        let mut items: Vec<_> = cache.iter().collect();
        items.sort_by_key(|(_, item)| item.created_at);
        
        let remove_count = items.len() / 4;
        for (key, _) in items.iter().take(remove_count) {
            cache.remove(*key);
            sqlx::query!("DELETE FROM offline_cache WHERE key = ?", key)
                .execute(&self.main_db)
                .await?;
        }

        Ok(())
    }

    async fn create_chat_message(&self, _payload: &str) -> Result<()> {
        // 实现聊天消息创建逻辑
        Ok(())
    }

    async fn update_user_profile(&self, _payload: &str) -> Result<()> {
        // 实现用户资料更新逻辑
        Ok(())
    }

    async fn delete_friend(&self, _friend_id: &str) -> Result<()> {
        // 实现好友删除逻辑
        Ok(())
    }
}

// Tauri命令函数
#[tauri::command]
pub async fn add_offline_operation(
    user_id: String,
    device_id: String,
    operation_type: String,
    data_type: String,
    data_id: String,
    data_payload: String,
    priority: Option<i32>,
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<String, String> {
    let operation_type: SyncOperation = serde_json::from_str(&operation_type)
        .map_err(|e| format!("Invalid operation type: {}", e))?;
    let data_type: SyncDataType = serde_json::from_str(&data_type)
        .map_err(|e| format!("Invalid data type: {}", e))?;
    let priority = match priority.unwrap_or(2) {
        1 => OperationPriority::Low,
        2 => OperationPriority::Normal,
        3 => OperationPriority::High,
        4 => OperationPriority::Critical,
        _ => OperationPriority::Normal,
    };

    offline_manager.add_offline_operation(
        &user_id,
        &device_id,
        operation_type,
        data_type,
        &data_id,
        &data_payload,
        priority,
    ).await
    .map_err(|e| format!("Failed to add offline operation: {}", e))
}

#[tauri::command]
pub async fn process_offline_queue(
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<Vec<String>, String> {
    offline_manager.process_offline_queue().await
        .map_err(|e| format!("Failed to process offline queue: {}", e))
}

#[tauri::command]
pub async fn get_offline_queue(
    user_id: String,
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<Vec<serde_json::Value>, String> {
    let operations = offline_manager.get_offline_queue(&user_id).await
        .map_err(|e| format!("Failed to get offline queue: {}", e))?;
    
    let json_operations: Vec<serde_json::Value> = operations.into_iter()
        .map(|op| serde_json::to_value(op).unwrap_or(serde_json::Value::Null))
        .collect();
    
    Ok(json_operations)
}

#[tauri::command]
pub async fn cache_data(
    key: String,
    data: String,
    data_type: String,
    ttl_seconds: Option<i64>,
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<(), String> {
    let data_type: SyncDataType = serde_json::from_str(&data_type)
        .map_err(|e| format!("Invalid data type: {}", e))?;
    
    offline_manager.cache_data(&key, &data, data_type, ttl_seconds).await
        .map_err(|e| format!("Failed to cache data: {}", e))
}

#[tauri::command]
pub async fn get_cached_data(
    key: String,
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<Option<String>, String> {
    offline_manager.get_cached_data(&key).await
        .map_err(|e| format!("Failed to get cached data: {}", e))
}

#[tauri::command]
pub async fn remove_cached_data(
    key: String,
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<(), String> {
    offline_manager.remove_cached_data(&key).await
        .map_err(|e| format!("Failed to remove cached data: {}", e))
}

#[tauri::command]
pub async fn get_unresolved_conflicts(
    user_id: String,
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<Vec<serde_json::Value>, String> {
    let conflicts = offline_manager.get_unresolved_conflicts(&user_id).await
        .map_err(|e| format!("Failed to get unresolved conflicts: {}", e))?;
    
    let json_conflicts: Vec<serde_json::Value> = conflicts.into_iter()
        .map(|conflict| serde_json::to_value(conflict).unwrap_or(serde_json::Value::Null))
        .collect();
    
    Ok(json_conflicts)
}

#[tauri::command]
pub async fn resolve_conflict(
    conflict_id: String,
    resolution: String,
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<(), String> {
    let resolution: SyncRecord = serde_json::from_str(&resolution)
        .map_err(|e| format!("Invalid resolution data: {}", e))?;
    
    offline_manager.resolve_conflict(&conflict_id, resolution).await
        .map_err(|e| format!("Failed to resolve conflict: {}", e))
}

#[tauri::command]
pub async fn cleanup_expired_data(
    offline_manager: tauri::State<'_, std::sync::Arc<OfflineDataManager>>,
) -> Result<(), String> {
    offline_manager.cleanup_expired_data().await
        .map_err(|e| format!("Failed to cleanup expired data: {}", e))
}